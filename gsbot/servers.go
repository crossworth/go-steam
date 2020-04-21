package gsbot

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net"

	"github.com/13k/go-steam"
	"github.com/13k/go-steam/netutil"
)

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
