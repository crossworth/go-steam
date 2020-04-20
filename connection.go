package steam

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/13k/go-steam/cryptoutil"
	"github.com/13k/go-steam/protocol"
)

type connection interface {
	Read() (*protocol.Packet, error)
	Write([]byte) error
	Close() error
	SetEncryptionKey([]byte) error
	IsEncrypted() bool
}

const tcpConnectionMagic uint32 = 0x31305456 // "VT01"

var _ connection = (*tcpConnection)(nil)

type tcpConnection struct {
	conn        *net.TCPConn
	ciph        cipher.Block
	cipherMutex sync.RWMutex
}

func dialTCP(laddr, raddr *net.TCPAddr) (*tcpConnection, error) {
	conn, err := net.DialTCP("tcp", laddr, raddr)
	if err != nil {
		return nil, err
	}

	return &tcpConnection{
		conn: conn,
	}, nil
}

func (c *tcpConnection) Read() (*protocol.Packet, error) {
	// All packets begin with a packet length
	var packetLen uint32
	err := binary.Read(c.conn, binary.LittleEndian, &packetLen)
	if err != nil {
		return nil, err
	}

	// A magic value follows for validation
	var packetMagic uint32

	if err = binary.Read(c.conn, binary.LittleEndian, &packetMagic); err != nil {
		return nil, err
	}

	if packetMagic != tcpConnectionMagic {
		return nil, fmt.Errorf(
			"steam/connection: invalid connection magic. expected %d, got %d",
			tcpConnectionMagic,
			packetMagic,
		)
	}

	buf := make([]byte, packetLen)

	if _, err = io.ReadFull(c.conn, buf); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, io.EOF
		}

		return nil, err
	}

	// Packets after ChannelEncryptResult are encrypted
	buf = c.decrypt(buf)

	return protocol.NewPacket(buf)
}

// Write sends a message.
//
// This may only be used by one goroutine at a time.
func (c *tcpConnection) Write(message []byte) error {
	var err error

	message, err = c.encrypt(message)

	if err != nil {
		return err
	}

	if err = binary.Write(c.conn, binary.LittleEndian, uint32(len(message))); err != nil {
		return err
	}

	if err = binary.Write(c.conn, binary.LittleEndian, tcpConnectionMagic); err != nil {
		return err
	}

	_, err = c.conn.Write(message)

	return err
}

func (c *tcpConnection) Close() error {
	return c.conn.Close()
}

func (c *tcpConnection) SetEncryptionKey(key []byte) error {
	c.cipherMutex.Lock()
	defer c.cipherMutex.Unlock()

	if key == nil {
		c.ciph = nil
		return nil
	}

	if len(key) != 32 {
		return errors.New("connection AES key is not 32 bytes long")
	}

	var err error

	c.ciph, err = aes.NewCipher(key)

	return err
}

func (c *tcpConnection) IsEncrypted() bool {
	c.cipherMutex.RLock()
	defer c.cipherMutex.RUnlock()
	return c.ciph != nil
}

func (c *tcpConnection) encrypt(message []byte) ([]byte, error) {
	c.cipherMutex.RLock()
	defer c.cipherMutex.RUnlock()

	if c.ciph != nil {
		return cryptoutil.SymmetricEncrypt(c.ciph, message)
	}

	return message, nil
}

func (c *tcpConnection) decrypt(message []byte) []byte {
	c.cipherMutex.RLock()
	defer c.cipherMutex.RUnlock()

	if c.ciph != nil {
		return cryptoutil.SymmetricDecrypt(c.ciph, message)
	}

	return message
}
