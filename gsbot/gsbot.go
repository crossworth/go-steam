// The GsBot package contains some useful utilites for working with the
// steam package. It implements authentication with sentries, server lists and
// logging messages and events.
//
// Every module is optional and requires an instance of the GsBot struct.
// Should a module have a `HandlePacket` method, you must register it with the
// steam.Client with `RegisterPacketHandler`. Any module with a `HandleEvent`
// method must be integrated into your event loop and should be called for each
// event you receive.
package gsbot

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/13k/go-steam"
	"github.com/13k/go-steam/netutil"
	"github.com/13k/go-steam/protocol"
	"github.com/davecgh/go-spew/spew"
)

type EventHandler interface {
	HandleEvent(event interface{})
}

// GsBot is the base struct holding common data among GsBot modules.
type GsBot struct {
	Client *steam.Client
	Log    *log.Logger

	handlers []EventHandler
}

var _ EventHandler = (*GsBot)(nil)

// Default creates a new GsBot with a new steam.Client where logs are written to stdout.
func Default() *GsBot {
	return &GsBot{
		Client: steam.NewClient(),
		Log:    log.New(os.Stdout, "", 0),
	}
}

func (bot *GsBot) HandleEvent(event interface{}) {
	for _, handler := range bot.handlers {
		handler.HandleEvent(event)
	}
}

func (bot *GsBot) RegisterEventHandler(handler EventHandler) {
	bot.handlers = append(bot.handlers, handler)
}

// Auth module handles authentication.
//
// It logs on automatically after a ConnectedEvent and saves the sentry data to a file which is also
// used for logon if available.
//
// If you're logging on for the first time Steam may require an authcode. You can then connect again
// with the new logon details.
type Auth struct {
	bot             *GsBot
	credentials     *steam.LogOnDetails
	sentryPath      string
	machineAuthHash []byte
}

var _ EventHandler = (*Auth)(nil)

func NewAuth(bot *GsBot, credentials *steam.LogOnDetails, sentryPath string) *Auth {
	auth := &Auth{
		bot:         bot,
		credentials: credentials,
		sentryPath:  sentryPath,
	}

	bot.RegisterEventHandler(auth)

	return auth
}

const fmtErrSentryRead = "Error loading sentry file from path %v - " +
	"This is normal if you're logging in for the first time."

// LogOn is called automatically after every ConnectedEvent, but must be called once again manually
// with an authcode if Steam requires it when logging on for the first time.
func (a *Auth) LogOn(credentials *steam.LogOnDetails) error {
	a.credentials = credentials
	sentry, err := ioutil.ReadFile(a.sentryPath)

	if err != nil {
		a.bot.Log.Printf(fmtErrSentryRead, a.sentryPath)
		sentry = nil
	}

	return a.bot.Client.Auth.LogOn(&steam.LogOnDetails{
		Username:       credentials.Username,
		Password:       credentials.Password,
		SentryFileHash: sentry,
		AuthCode:       credentials.AuthCode,
		TwoFactorCode:  credentials.TwoFactorCode,
	})
}

func (a *Auth) HandleEvent(event interface{}) {
	switch e := event.(type) {
	case *steam.ConnectedEvent:
		if err := a.LogOn(a.credentials); err != nil {
			a.bot.Log.Fatalf("error writing sentry file: %v", err)
		}
	case *steam.LoggedOnEvent:
		a.bot.Log.Printf("Logged on (%v) with SteamID %v and account flags %v", e.Result, e.ClientSteamID, e.AccountFlags)
	case *steam.MachineAuthUpdateEvent:
		a.machineAuthHash = e.Hash

		if err := ioutil.WriteFile(a.sentryPath, e.Hash, 0666); err != nil {
			a.bot.Log.Fatalf("error writing sentry file: %v", err)
		}
	}
}

// ServerList module saves the server list from ClientCMListEvent and uses it when you call
// `client.Connect`.
type ServerList struct {
	bot      *GsBot
	listPath string
}

var _ EventHandler = (*ServerList)(nil)

func NewServerList(bot *GsBot, listPath string) *ServerList {
	sl := &ServerList{
		bot:      bot,
		listPath: listPath,
	}

	bot.RegisterEventHandler(sl)

	return sl
}

func (s *ServerList) HandleEvent(event interface{}) {
	switch e := event.(type) {
	case *steam.ClientCMListEvent:
		d, err := json.Marshal(e.Addresses)

		if err != nil {
			s.bot.Log.Fatalf("error encoding servers JSON: %v", err)
		}

		if err = ioutil.WriteFile(s.listPath, d, 0666); err != nil {
			s.bot.Log.Fatalf("error writing servers file: %v", err)
		}
	}
}

func (s *ServerList) Connect() (bool, error) {
	return s.ConnectBind(nil)
}

func (s *ServerList) ConnectBind(laddr *net.TCPAddr) (bool, error) {
	d, err := ioutil.ReadFile(s.listPath)

	if err != nil {
		s.bot.Log.Println("Connecting to random server.")

		if _, err = s.bot.Client.Connect(); err != nil {
			return false, err
		}

		return false, nil
	}

	var addrs []*netutil.PortAddr

	if err = json.Unmarshal(d, &addrs); err != nil {
		return false, err
	}

	raddr := addrs[rand.Intn(len(addrs))]

	s.bot.Log.Printf("Connecting to %v from server list", raddr)

	if err := s.bot.Client.ConnectToBind(raddr, laddr); err != nil {
		return true, err
	}

	return true, nil
}

// Debug module logs incoming packets and events to a directory.
type Debug struct {
	packetID uint64
	eventID  uint64
	bot      *GsBot
	dir      string
}

var _ EventHandler = (*Debug)(nil)

func NewDebug(bot *GsBot, basedir string) (*Debug, error) {
	basedir = filepath.Join(basedir, fmt.Sprint(time.Now().Unix()))

	if err := os.MkdirAll(filepath.Join(basedir, "events"), 0700); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Join(basedir, "packets"), 0700); err != nil {
		return nil, err
	}

	d := &Debug{
		bot: bot,
		dir: basedir,
	}

	bot.Client.RegisterPacketHandler(d)
	bot.RegisterEventHandler(d)

	return d, nil
}

func (d *Debug) HandlePacket(packet *protocol.Packet) {
	d.packetID++

	name := filepath.Join(d.dir, "packets", fmt.Sprintf("%d_%d_%s", time.Now().Unix(), d.packetID, packet.EMsg))
	text := packet.String() + "\n\n" + hex.Dump(packet.Data)
	fname := name + ".txt"

	if err := ioutil.WriteFile(fname, []byte(text), 0666); err != nil {
		d.bot.Log.Fatalf("error writing debug file %s: %v", fname, err)
	}

	fname = name + ".bin"

	if err := ioutil.WriteFile(fname, packet.Data, 0666); err != nil {
		d.bot.Log.Fatalf("error writing debug file %s: %v", fname, err)
	}

	d.bot.Log.Printf("received packet %s", packet.EMsg)
}

func (d *Debug) HandleEvent(event interface{}) {
	d.eventID++

	eventName := reflectName(event)
	name := fmt.Sprintf("%d_%d_%s.txt", time.Now().Unix(), d.eventID, eventName)
	fname := filepath.Join(d.dir, "events", name)
	data := []byte(spew.Sdump(event))

	if err := ioutil.WriteFile(fname, data, 0666); err != nil {
		d.bot.Log.Fatalf("error writing debug file %s: %v", fname, err)
	}

	d.bot.Log.Printf("received event %s", eventName)
}

func reflectName(obj interface{}) string {
	val := reflect.ValueOf(obj)
	ind := reflect.Indirect(val)

	if ind.IsValid() {
		return ind.Type().Name()
	}

	return val.Type().Name()
}
