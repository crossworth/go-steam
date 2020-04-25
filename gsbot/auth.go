package gsbot

import (
	"io/ioutil"

	"github.com/13k/go-steam"
)

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
			a.bot.Log.Fatalf("error logging in: %v", err)
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
