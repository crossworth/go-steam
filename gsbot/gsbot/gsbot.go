// A simple example that uses the modules from the gsbot package and go-steam to log on
// to the Steam network.
//
// The command expects log on data, optionally with an auth code:
//
//     gsbot [username] [password]
//     gsbot [username] [password] [authcode]
package main

import (
	"fmt"
	"os"

	"github.com/13k/go-steam"
	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/gsbot"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("gsbot example\nusage: \n\tgsbot [username] [password] [authcode]")
		return
	}

	authcode := ""

	if len(os.Args) > 3 {
		authcode = os.Args[3]
	}

	bot := gsbot.Default()
	client := bot.Client
	credentials := &gsbot.LogOnDetails{
		Username: os.Args[1],
		Password: os.Args[2],
		AuthCode: authcode,
	}

	auth := gsbot.NewAuth(bot, credentials, "sentry.bin")
	debug, err := gsbot.NewDebug(bot, "debug")

	if err != nil {
		panic(err)
	}

	client.RegisterPacketHandler(debug)
	serverList := gsbot.NewServerList(bot, "serverlist.json")

	serverList.Connect()

	for event := range client.Events() {
		auth.HandleEvent(event)
		debug.HandleEvent(event)
		serverList.HandleEvent(event)

		switch e := event.(type) {
		case error:
			fmt.Printf("Error: %v", e)
		case *steam.LoggedOnEvent:
			client.Social.SetPersonaState(steamlang.EPersonaState_Online)
		}
	}
}
