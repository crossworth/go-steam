// Package api implements a trading API client.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/13k/go-steam/community"
	"github.com/13k/go-steam/economy/inventory"
	"github.com/13k/go-steam/netutil"
	"github.com/13k/go-steam/steamid"
)

const tradeURL = "https://steamcommunity.com/trade/%d/"

type Client struct {
	LogPos  uint // not automatically updated
	Version uint // Incremented for each item change by Steam; not automatically updated.

	client    *http.Client
	other     steamid.SteamID
	sessionID string // the `sessionid` cookie is sent as a parameter/POST data for CSRF protection.
	baseURL   string
}

// New creates a new Trade based on the given cookies `sessionid`, `steamLogin`, `steamLoginSecure` and
// the trade partner's Steam ID.
func New(sessionID, steamLogin, steamLoginSecure string, other steamid.SteamID) (*Client, error) {
	client := &http.Client{}
	client.Timeout = 10 * time.Second

	t := &Client{
		client:    client,
		other:     other,
		sessionID: sessionID,
		baseURL:   fmt.Sprintf(tradeURL, other),
		Version:   1,
	}

	if err := community.SetCookies(t.client, sessionID, steamLogin, steamLoginSecure); err != nil {
		return nil, err
	}

	return t, nil
}

type Main struct {
	PartnerOnProbation bool
}

var onProbationRegex = regexp.MustCompile(`var g_bTradePartnerProbation = (\w+);`)

// GetMain fetches the main HTML page and parses it. Thread-safe.
func (t *Client) GetMain() (*Main, error) {
	resp, err := t.client.Get(t.baseURL)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	match := onProbationRegex.FindSubmatch(body)

	if len(match) == 0 {
		return nil, errors.New("tradeapi.GetMain: Could not find probation info")
	}

	return &Main{
		PartnerOnProbation: string(match[1]) == "true",
	}, nil
}

// Ajax POSTs to an API endpoint that should return a status
func (t *Client) postWithStatus(url string, data map[string]string) (*Result, error) {
	status := &Result{}
	req, err := netutil.NewPostForm(url, netutil.ToURLValues(data))

	if err != nil {
		return nil, err
	}

	// Tales of Madness and Pain, Episode 1: If you forget this, Steam will return an error
	// saying "missing required parameter", even though they are all there. IT WAS JUST THE HEADER, ARGH!
	req.Header.Add("Referer", t.baseURL)

	resp, err := t.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(status); err != nil {
		return nil, err
	}

	return status, nil
}

func (t *Client) GetStatus() (*Result, error) {
	return t.postWithStatus(t.baseURL+"tradestatus/", map[string]string{
		"sessionid": t.sessionID,
		"logpos":    strconv.FormatUint(uint64(t.LogPos), 10),
		"version":   strconv.FormatUint(uint64(t.Version), 10),
	})
}

// Thread-safe.
func (t *Client) GetForeignInventory(contextID uint64, appID uint32, start uint) (*inventory.PartialInventory, error) {
	data := map[string]string{
		"sessionid": t.sessionID,
		"steamid":   fmt.Sprintf("%d", t.other),
		"contextid": strconv.FormatUint(contextID, 10),
		"appid":     strconv.FormatUint(uint64(appID), 10),
	}

	if start != 0 {
		data["start"] = strconv.FormatUint(uint64(start), 10)
	}

	req, err := http.NewRequest("GET", t.baseURL+"foreigninventory?"+netutil.ToURLValues(data).Encode(), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Referer", t.baseURL)

	return inventory.PerformRequest(t.client, req)
}

// Thread-safe.
func (t *Client) GetOwnInventory(contextID uint64, appID uint32) (*inventory.Inventory, error) {
	return inventory.GetOwnInventory(t.client, contextID, appID)
}

func (t *Client) Chat(message string) (*Result, error) {
	return t.postWithStatus(t.baseURL+"chat", map[string]string{
		"sessionid": t.sessionID,
		"logpos":    strconv.FormatUint(uint64(t.LogPos), 10),
		"version":   strconv.FormatUint(uint64(t.Version), 10),
		"message":   message,
	})
}

func (t *Client) AddItem(slot uint, itemID, contextID uint64, appID uint32) (*Result, error) {
	return t.postWithStatus(t.baseURL+"additem", map[string]string{
		"sessionid": t.sessionID,
		"slot":      strconv.FormatUint(uint64(slot), 10),
		"itemid":    strconv.FormatUint(itemID, 10),
		"contextid": strconv.FormatUint(contextID, 10),
		"appid":     strconv.FormatUint(uint64(appID), 10),
	})
}

func (t *Client) RemoveItem(slot uint, itemID, contextID uint64, appID uint32) (*Result, error) {
	return t.postWithStatus(t.baseURL+"removeitem", map[string]string{
		"sessionid": t.sessionID,
		"slot":      strconv.FormatUint(uint64(slot), 10),
		"itemid":    strconv.FormatUint(itemID, 10),
		"contextid": strconv.FormatUint(contextID, 10),
		"appid":     strconv.FormatUint(uint64(appID), 10),
	})
}

func (t *Client) SetCurrency(amount uint, currencyID, contextID uint64, appID uint32) (*Result, error) {
	return t.postWithStatus(t.baseURL+"setcurrency", map[string]string{
		"sessionid":  t.sessionID,
		"amount":     strconv.FormatUint(uint64(amount), 10),
		"currencyid": strconv.FormatUint(currencyID, 10),
		"contextid":  strconv.FormatUint(contextID, 10),
		"appid":      strconv.FormatUint(uint64(appID), 10),
	})
}

func (t *Client) SetReady(ready bool) (*Result, error) {
	return t.postWithStatus(t.baseURL+"toggleready", map[string]string{
		"sessionid": t.sessionID,
		"version":   strconv.FormatUint(uint64(t.Version), 10),
		"ready":     fmt.Sprint(ready),
	})
}

func (t *Client) Confirm() (*Result, error) {
	return t.postWithStatus(t.baseURL+"confirm", map[string]string{
		"sessionid": t.sessionID,
		"version":   strconv.FormatUint(uint64(t.Version), 10),
	})
}

func (t *Client) Cancel() (*Result, error) {
	return t.postWithStatus(t.baseURL+"cancel", map[string]string{
		"sessionid": t.sessionID,
	})
}
