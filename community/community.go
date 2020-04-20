package community

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const cookiePath = "https://steamcommunity.com/"

func SetCookies(client *http.Client, sessionID, steamLogin, steamLoginSecure string) error {
	var err error

	if client.Jar == nil {
		if client.Jar, err = cookiejar.New(nil); err != nil {
			return err
		}
	}

	base, err := url.Parse(cookiePath)

	if err != nil {
		return err
	}

	client.Jar.SetCookies(base, []*http.Cookie{
		// It seems that, for some reason, Steam tries to URL-decode the cookie.
		{
			Name:  "sessionid",
			Value: url.QueryEscape(sessionID),
		},
		// steamLogin is already URL-encoded.
		{
			Name:  "steamLogin",
			Value: steamLogin,
		},
		{
			Name:  "steamLoginSecure",
			Value: steamLoginSecure,
		},
	})

	return nil
}
