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
	"log"
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

	gsbot.NewAuth(bot, credentials, "sentry.bin")
	_, err := gsbot.NewDebug(bot, "debug")

	if err != nil {
		log.Fatal(err)
	}

	serverList := gsbot.NewServerList(bot, "serverlist.json")

	if _, err := serverList.Connect(); err != nil {
		log.Fatal(err)
	}

	for event := range client.Events() {
		bot.HandleEvent(event)

		switch e := event.(type) {
		case *steam.LoggedOnEvent:
			client.Social.SetPersonaState(steamlang.EPersonaState_Online)
		case error:
			log.Printf("Error: %v", e)
		}
	}
}
