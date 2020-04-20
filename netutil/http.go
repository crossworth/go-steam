package netutil

import (
	"net/http"
	"net/url"
	"strings"
)

// NewPostForm is like http.Client.PostForm, but returns a new request instead of executing it
// directly.
func NewPostForm(url string, data url.Values) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}
