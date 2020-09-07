package steam

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"net"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/13k/go-steam-resources/protobuf/steam"
	"github.com/13k/go-steam-resources/steamlang"

	"github.com/13k/go-steam/cryptoutil"
	"github.com/13k/go-steam/netutil"
	"github.com/13k/go-steam/protocol"
	"github.com/13k/go-steam/steamid"
)

// Client implements a client to the Steam network.
//
// Always poll events from the channel returned by Events() or receiving messages will stop. All
// access, unless otherwise noted, should be threadsafe.
//
// When a FatalErrorEvent is emitted, the connection is automatically closed. The same client can be
// used to reconnect. Other errors don't have any effect.
type Client struct {
	// these need to be 64 bit aligned for sync/atomic on 32bit
	sessionID    int32
	_            uint32
	steamID      uint64
	currentJobID uint64

	Auth          *Auth
	Social        *Social
	Web           *Web
	Notifications *Notifications
	Trading       *Trading
	GC            *GameCoordinator

	events      chan interface{}
	handlers    []protocol.PacketHandler
	handlersMtx sync.RWMutex

	tempSessionKey []byte

	mtx       sync.RWMutex // guarding conn and writeChan
	conn      connection
	writeChan chan protocol.Message
	writeBuf  *bytes.Buffer
	heartbeat *time.Ticker
}

func NewClient() *Client {
	client := &Client{
		events:   make(chan interface{}, 30),
		writeBuf: &bytes.Buffer{},
	}

	client.Auth = NewAuth(client)
	client.Social = NewSocial(client)
	client.Web = NewWeb(client)
	client.Notifications = NewNotifications(client)
	client.Trading = NewTrading(client)
	client.GC = NewGC(client)

	client.RegisterPacketHandler(client.Auth)
	client.RegisterPacketHandler(client.Social)
	client.RegisterPacketHandler(client.Web)
	client.RegisterPacketHandler(client.Notifications)
	client.RegisterPacketHandler(client.Trading)
	client.RegisterPacketHandler(client.GC)

	return client
}

// Get the event channel. By convention all events are pointers, except for errors.
// It is never closed.
func (c *Client) Events() <-chan interface{} {
	return c.events
}

func (c *Client) CloseEventsChannel() {
	close(c.events)
}

func (c *Client) Emit(event interface{}) {
	c.events <- event
}

// Emits a FatalErrorEvent formatted with fmt.Errorf and disconnects.
func (c *Client) Fatalf(format string, a ...interface{}) {
	c.Emit(FatalErrorEvent(fmt.Errorf(format, a...)))
	c.Disconnect()
}

// Emits an error formatted with fmt.Errorf.
func (c *Client) Errorf(format string, a ...interface{}) {
	c.Emit(fmt.Errorf(format, a...))
}

// Registers a protocol.PacketHandler that receives all incoming packets.
func (c *Client) RegisterPacketHandler(handler protocol.PacketHandler) {
	c.handlersMtx.Lock()
	defer c.handlersMtx.Unlock()
	c.handlers = append(c.handlers, handler)
}

// NextJobID returns the next job ID to use.
func (c *Client) NextJobID() protocol.JobID {
	return protocol.JobID(atomic.AddUint64(&c.currentJobID, 1))
}

// SteamID returns the client's steam ID.
func (c *Client) SteamID() steamid.SteamID {
	return steamid.SteamID(atomic.LoadUint64(&c.steamID))
}

func (c *Client) setSteamID(steamID steamid.SteamID) {
	atomic.StoreUint64(&c.steamID, steamID.Uint64())
}

// SessionID returns the session id.
func (c *Client) SessionID() int32 {
	return atomic.LoadInt32(&c.sessionID)
}

func (c *Client) setSessionID(sessionID int32) {
	atomic.StoreInt32(&c.sessionID, sessionID)
}

func (c *Client) Connected() bool {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.conn != nil
}

// Connect connects to a random Steam server and returns its address.
//
// If this client is already connected, it is disconnected first.
//
// This method tries to use an address from the Steam Directory and falls back to the built-in
// server list if the Steam Directory can't be reached.
//
// If you want to connect to a specific server, use `ConnectTo`.
func (c *Client) Connect() (*netutil.PortAddr, error) {
	var (
		server *netutil.PortAddr
		err    error
	)

	if steamDirectoryCache.IsInitialized() {
		server, err = steamDirectoryCache.GetRandomCM()
	} else {
		server, err = GetRandomCM()
	}

	if err != nil {
		return nil, err
	}

	if err = c.ConnectTo(server); err != nil {
		return nil, err
	}

	return server, nil
}

// ConnectTo connects to a specific server.
//
// You may want to use one of the `GetRandom*CM()` functions in this package.
//
// If this client is already connected, it is disconnected first.
func (c *Client) ConnectTo(addr *netutil.PortAddr) error {
	return c.ConnectToBind(addr, nil)
}

// ConnectToBind connects to a specific server, and binds to a specified local IP.
//
// If this client is already connected, it is disconnected first.
func (c *Client) ConnectToBind(addr *netutil.PortAddr, local *net.TCPAddr) error {
	c.Disconnect()

	conn, err := dialTCP(local, addr.ToTCPAddr())

	if err != nil {
		return err
	}

	c.conn = conn
	c.writeChan = make(chan protocol.Message, 5)

	go c.readLoop()
	go c.writeLoop()

	return nil
}

func (c *Client) Disconnect() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.conn == nil {
		return
	}

	c.conn.Close()
	c.conn = nil

	if c.heartbeat != nil {
		c.heartbeat.Stop()
	}

	close(c.writeChan)

	c.Emit(&DisconnectedEvent{})
}

// Write adds a message to the send queue.
//
// Modifications to the given message after writing are not allowed (possible race conditions).
//
// Writes to this client when not connected are ignored.
func (c *Client) Write(msg protocol.Message) {
	if cm, ok := msg.(protocol.ClientMessage); ok {
		cm.SetSessionID(c.SessionID())
		cm.SetSteamID(c.SteamID())
	}

	c.mtx.RLock()
	defer c.mtx.RUnlock()

	if c.conn == nil {
		return
	}

	c.writeChan <- msg
}

func (c *Client) readLoop() {
	for {
		// This *should* be atomic on most platforms, but the Go spec doesn't guarantee it
		c.mtx.RLock()
		conn := c.conn
		c.mtx.RUnlock()

		if conn == nil {
			return
		}

		packet, err := conn.Read()

		if err != nil {
			c.Fatalf("client/read: error reading from the connection: %v", err)
			return
		}

		c.handlePacket(packet)
	}
}

func (c *Client) writeLoop() {
	defer func() {
		c.writeBuf.Reset()
	}()

	for {
		c.mtx.RLock()
		conn := c.conn
		c.mtx.RUnlock()

		if conn == nil {
			return
		}

		msg, ok := <-c.writeChan

		if !ok {
			return
		}

		c.writeBuf.Reset()

		if err := msg.Serialize(c.writeBuf); err != nil {
			c.Fatalf("client/write: error serializing message %s: %v", msg.Type(), err)
			return
		}

		if _, err := c.conn.Write(c.writeBuf.Bytes()); err != nil {
			c.Fatalf("client/write: error writing message %s: %v", msg.Type(), err)
			return
		}

		c.writeBuf.Reset()
	}
}

func (c *Client) heartbeatLoop(seconds time.Duration) {
	if c.heartbeat != nil {
		c.heartbeat.Stop()
	}

	c.heartbeat = time.NewTicker(seconds * time.Second)

	for {
		_, ok := <-c.heartbeat.C

		if !ok {
			break
		}

		c.Write(protocol.NewProtoMessage(steamlang.EMsg_ClientHeartBeat, &pb.CMsgClientHeartBeat{}))
	}

	c.heartbeat = nil
}

func (c *Client) handlePacket(packet *protocol.Packet) {
	switch packet.EMsg() {
	case steamlang.EMsg_ChannelEncryptRequest:
		c.handleChannelEncryptRequest(packet)
	case steamlang.EMsg_ChannelEncryptResult:
		c.handleChannelEncryptResult(packet)
	case steamlang.EMsg_Multi:
		c.handleMulti(packet)
	case steamlang.EMsg_ClientCMList:
		c.handleClientCMList(packet)
	}

	c.handlersMtx.RLock()
	defer c.handlersMtx.RUnlock()

	for _, handler := range c.handlers {
		handler.HandlePacket(packet)
	}
}

func (c *Client) handleChannelEncryptRequest(packet *protocol.Packet) {
	body := steamlang.NewMsgChannelEncryptRequest()

	if _, err := packet.ReadMsg(body); err != nil {
		c.Fatalf("client/ChannelEncryptRequest: error reading message: %v", err)
		return
	}

	if body.Universe != steamlang.EUniverse_Public {
		c.Fatalf("client/ChannelEncryptRequest: invalid universe %v", body.Universe)
		return
	}

	c.tempSessionKey = make([]byte, 32)

	if _, err := rand.Read(c.tempSessionKey); err != nil {
		c.Fatalf("client/ChannelEncryptRequest: error generating session key: %v", err)
		return
	}

	encryptedKey, err := cryptoutil.RSAEncrypt(GetPublicKey(steamlang.EUniverse_Public), c.tempSessionKey)

	if err != nil {
		c.Fatalf("client/ChannelEncryptRequest: error encrypting session key: %v", err)
		return
	}

	payload := &bytes.Buffer{}
	payload.Write(encryptedKey)

	if err := binary.Write(payload, binary.LittleEndian, crc32.ChecksumIEEE(encryptedKey)); err != nil {
		c.Fatalf("client/ChannelEncryptRequest: error creating encrypted response payload: %v", err)
		return
	}

	payload.WriteByte(0)
	payload.WriteByte(0)
	payload.WriteByte(0)
	payload.WriteByte(0)

	c.Write(protocol.NewStructMessage(steamlang.NewMsgChannelEncryptResponse(), payload.Bytes()))
}

func (c *Client) handleChannelEncryptResult(packet *protocol.Packet) {
	body := steamlang.NewMsgChannelEncryptResult()

	if _, err := packet.ReadMsg(body); err != nil {
		c.Fatalf("client/ChannelEncryptResult: error reading message: %v", err)
		return
	}

	if body.Result != steamlang.EResult_OK {
		c.Fatalf("client/ChannelEncryptResult: encryption failed: %v", body.Result)
		return
	}

	if err := c.conn.SetEncryptionKey(c.tempSessionKey); err != nil {
		c.Fatalf("client/ChannelEncryptResult: error setting encryption key: %v", err)
		return
	}

	c.tempSessionKey = nil

	c.Emit(&ConnectedEvent{})
}

func (c *Client) handleMulti(packet *protocol.Packet) {
	body := &pb.CMsgMulti{}

	if _, err := packet.ReadProtoMsg(body); err != nil {
		c.Errorf("client/Multi: error reading message: %v", err)
		return
	}

	payload := body.GetMessageBody()

	if body.GetSizeUnzipped() > 0 {
		r, err := gzip.NewReader(bytes.NewReader(payload))

		if err != nil {
			c.Errorf("client/Multi: error decompressing: %v", err)
			return
		}

		payload, err = ioutil.ReadAll(r)

		if err != nil {
			c.Errorf("client/Multi: error decompressing: %v", err)
			return
		}
	}

	pr := bytes.NewReader(payload)

	for pr.Len() > 0 {
		var length uint32

		if err := binary.Read(pr, binary.LittleEndian, &length); err != nil {
			c.Errorf("client/Multi: error reading packet %v: %v", packet, err)
			return
		}

		packetData := make([]byte, length)

		if _, err := pr.Read(packetData); err != nil {
			c.Errorf("client/Multi: error reading packet %v: %v", packet, err)
			return
		}

		p, err := protocol.NewPacket(packetData)

		if err != nil {
			c.Errorf("client/Multi: error reading packet %v: %v", packet, err)
			continue
		}

		c.handlePacket(p)
	}
}

func (c *Client) handleClientCMList(packet *protocol.Packet) {
	body := &pb.CMsgClientCMList{}

	if _, err := packet.ReadProtoMsg(body); err != nil {
		c.Errorf("client/ClientCMList: error reading message: %v", err)
		return
	}

	l := make([]*netutil.PortAddr, len(body.GetCmAddresses()))

	for i, ip := range body.GetCmAddresses() {
		l[i] = &netutil.PortAddr{
			IP:   readIPv4(ip),
			Port: uint16(body.GetCmPorts()[i]),
		}
	}

	c.Emit(&ClientCMListEvent{Addresses: l})
}

func readIPv4(ip uint32) net.IP {
	r := make(net.IP, 4)
	r[3] = byte(ip)
	r[2] = byte(ip >> 8)
	r[1] = byte(ip >> 16)
	r[0] = byte(ip >> 24)
	return r
}
