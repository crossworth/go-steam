package gsbot

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/13k/go-steam/protocol"
	"github.com/davecgh/go-spew/spew"
)

// Debug module logs incoming packets and events to a directory.
type Debug struct {
	packetID uint64
	eventID  uint64
	bot      *GsBot
	dir      string
}

var _ EventHandler = (*Debug)(nil)
var _ protocol.PacketHandler = (*Debug)(nil)

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

	basename := filepath.Join(d.dir, "packets", fmt.Sprintf("%08d_%d_%s", d.packetID, time.Now().Unix(), packet.EMsg()))
	content := fmt.Sprintf("%s\n\n%s", packet, hex.Dump(packet.Data))
	fname := basename + ".txt"

	if err := ioutil.WriteFile(fname, []byte(content), 0666); err != nil {
		d.bot.Log.Fatalf("error writing debug file %s: %v", fname, err)
	}

	fname = basename + ".bin"

	if err := ioutil.WriteFile(fname, packet.Data, 0666); err != nil {
		d.bot.Log.Fatalf("error writing debug file %s: %v", fname, err)
	}

	d.bot.Log.Printf("received packet %s", packet.EMsg())
}

func (d *Debug) HandleEvent(event interface{}) {
	d.eventID++

	eventName := reflectName(event)
	basename := fmt.Sprintf("%08d_%d_%s.txt", d.eventID, time.Now().Unix(), eventName)
	fname := filepath.Join(d.dir, "events", basename)
	content := spew.Sdump(event)

	if err := ioutil.WriteFile(fname, []byte(content), 0666); err != nil {
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
