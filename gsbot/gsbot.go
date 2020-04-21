// Package gsbot contains some useful utilites for working with the steam package. It implements
// authentication with sentries, server lists and logging messages and events.
//
// Every module is optional and requires an instance of the GsBot struct.
//
// Each a module auto-registers as `steam.PacketHandler` with the `steam.Client` and `EventHandler`
// with GsBot.
package gsbot

import (
	"log"
	"os"

	"github.com/13k/go-steam"
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
