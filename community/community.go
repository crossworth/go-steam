package community

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const cookiePath = "https://steamcommunity.com/"

func SetCookies(client *http.Client, sessionID, steamLogin, steamLoginSecure string) {
	if client.Jar == nil {
		client.Jar, _ = cookiejar.New(nil)
	}

	base, err := url.Parse(cookiePath)

	if err != nil {
		panic(err)
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
}
