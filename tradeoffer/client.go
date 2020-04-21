package tradeoffer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/13k/go-steam/community"
	"github.com/13k/go-steam/economy/inventory"
	"github.com/13k/go-steam/netutil"
	"github.com/13k/go-steam/steamid"
)

type APIKey string

const apiURL = "https://api.steampowered.com/IEconService/%s/v%d"

type Client struct {
	client    *http.Client
	key       APIKey
	sessionID string
}

func NewClient(key APIKey, sessionID, steamLogin, steamLoginSecure string) (*Client, error) {
	c := &Client{
		client:    &http.Client{},
		key:       key,
		sessionID: sessionID,
	}

	if err := community.SetCookies(c.client, sessionID, steamLogin, steamLoginSecure); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) GetOffer(offerID uint64) (*Result, error) {
	resp, err := c.client.Get(fmt.Sprintf(apiURL, "GetTradeOffer", 1) + "?" + netutil.ToURLValues(map[string]string{
		"key":          string(c.key),
		"tradeofferid": strconv.FormatUint(offerID, 10),
		"language":     "en_us",
	}).Encode())

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	t := &struct {
		Response *Result
	}{}

	if err = json.NewDecoder(resp.Body).Decode(t); err != nil {
		return nil, err
	}

	if t.Response == nil || t.Response.Offer == nil {
		return nil, newSteamErrorf("steam returned empty offer result")
	}

	return t.Response, nil
}

func (c *Client) GetOffers(
	getSent bool,
	getReceived bool,
	getDescriptions bool,
	activeOnly bool,
	historicalOnly bool,
	timeHistoricalCutoff *uint32,
) (*MultiResult, error) {
	if !getSent && !getReceived {
		return nil, errors.New("getSent and getReceived can't be both false")
	}

	params := map[string]string{
		"key": string(c.key),
	}

	if getSent {
		params["get_sent_offers"] = "1"
	}

	if getReceived {
		params["get_received_offers"] = "1"
	}

	if getDescriptions {
		params["get_descriptions"] = "1"
		params["language"] = "en_us"
	}

	if activeOnly {
		params["active_only"] = "1"
	}

	if historicalOnly {
		params["historical_only"] = "1"
	}

	if timeHistoricalCutoff != nil {
		params["time_historical_cutoff"] = strconv.FormatUint(uint64(*timeHistoricalCutoff), 10)
	}

	resp, err := c.client.Get(fmt.Sprintf(apiURL, "GetTradeOffers", 1) + "?" + netutil.ToURLValues(params).Encode())

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	t := &struct {
		Response *MultiResult
	}{}

	if err = json.NewDecoder(resp.Body).Decode(t); err != nil {
		return nil, err
	}

	if t.Response == nil {
		return nil, newSteamErrorf("steam returned empty offers result\n")
	}

	return t.Response, nil
}

// action is used by Decline and Cancel.
//
// Steam only returns success and error fields for malformed requests, hence client shall use
// GetOffer to check action result.
//
// It is also possible to implement Decline/Cancel using steamcommunity, which have more predictable
// responses.
func (c *Client) action(method string, version uint, offerID uint64) error {
	data := netutil.ToURLValues(map[string]string{
		"key":          string(c.key),
		"tradeofferid": strconv.FormatUint(offerID, 10),
	})

	req, err := netutil.NewPostForm(fmt.Sprintf(apiURL, method, version), data)

	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("%s error: status code %d", method, resp.StatusCode)
	}

	return nil
}

func (c *Client) Decline(offerID uint64) error {
	return c.action("DeclineTradeOffer", 1, offerID)
}

func (c *Client) Cancel(offerID uint64) error {
	return c.action("CancelTradeOffer", 1, offerID)
}

// Accept accepts received trade offer.
//
// It is best to confirm that offer was actually accepted by calling GetOffer after Accept and
// checking offer state.
func (c *Client) Accept(offerID uint64) error {
	baseurl := fmt.Sprintf("https://steamcommunity.com/tradeoffer/%d/", offerID)

	data := netutil.ToURLValues(map[string]string{
		"sessionid":    c.sessionID,
		"serverid":     "1",
		"tradeofferid": strconv.FormatUint(offerID, 10),
	})

	req, err := netutil.NewPostForm(baseurl+"accept", data)

	if err != nil {
		return err
	}

	req.Header.Add("Referer", baseurl)

	resp, err := c.client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	t := &struct {
		StrError string `json:"strError"`
	}{}

	if err = json.NewDecoder(resp.Body).Decode(t); err != nil {
		return err
	}

	if t.StrError != "" {
		return newSteamErrorf("accept error: %v\n", t.StrError)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("accept error: status code %d", resp.StatusCode)
	}

	return nil
}

type TradeItem struct {
	AppID      uint32 `json:"appid"`
	ContextID  uint64 `json:"contextid,string"`
	Amount     uint64 `json:"amount"`
	AssetID    uint64 `json:"assetid,string,omitempty"`
	CurrencyID uint64 `json:"currencyid,string,omitempty"`
}

// Create sends a new trade offer to the given Steam user.
//
// You can optionally specify an access token if you've got one. In addition, `counteredOfferID` can
// be non-nil, indicating the trade offer this is a counter for.
//
// On success returns trade offer id.
func (c *Client) Create(
	other steamid.SteamID,
	accessToken *string,
	myItems, theirItems []TradeItem,
	counteredOfferID *uint64,
	message string,
) (uint64, error) {
	// Create new trade offer status
	to := map[string]interface{}{
		"newversion": true,
		"version":    3,
		"me": map[string]interface{}{
			"assets":   myItems,
			"currency": make([]struct{}, 0),
			"ready":    false,
		},
		"them": map[string]interface{}{
			"assets":   theirItems,
			"currency": make([]struct{}, 0),
			"ready":    false,
		},
	}

	jto, err := json.Marshal(to)

	if err != nil {
		return 0, err
	}

	// Create url parameters for request
	data := map[string]string{
		"sessionid":         c.sessionID,
		"serverid":          "1",
		"partner":           other.FormatString(),
		"tradeoffermessage": message,
		"json_tradeoffer":   string(jto),
	}

	var referer string

	if counteredOfferID != nil {
		referer = fmt.Sprintf("https://steamcommunity.com/tradeoffer/%d/", *counteredOfferID)
		data["tradeofferid_countered"] = strconv.FormatUint(*counteredOfferID, 10)
	} else {
		// Add token for non-friend offers
		if accessToken != nil {
			params := map[string]string{
				"trade_offer_access_token": *accessToken,
			}

			var paramsJSON []byte

			paramsJSON, err = json.Marshal(params)

			if err != nil {
				return 0, err
			}

			data["trade_offer_create_params"] = string(paramsJSON)

			referer = "https://steamcommunity.com/tradeoffer/new/?partner=" +
				other.AccountID().FormatString() +
				"&token=" +
				*accessToken
		} else {
			referer = "https://steamcommunity.com/tradeoffer/new/?partner=" +
				other.AccountID().FormatString()
		}
	}

	// Create request
	req, err := netutil.NewPostForm("https://steamcommunity.com/tradeoffer/new/send", netutil.ToURLValues(data))

	if err != nil {
		return 0, err
	}

	req.Header.Add("Referer", referer)

	// Send request
	resp, err := c.client.Do(req)

	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	t := &struct {
		StrError     string `json:"strError"`
		TradeOfferID uint64 `json:"tradeofferid,string"`
	}{}

	if err = json.NewDecoder(resp.Body).Decode(t); err != nil {
		return 0, err
	}

	// strError code descriptions:
	// 15	invalide trade access token
	// 16	timeout
	// 20	wrong contextid
	// 25	can't send more offers until some is accepted/canceled...
	// 26	object is not in our inventory
	// error code names are in internal/steamlang/enums.go EResult_name
	if t.StrError != "" {
		return 0, newSteamErrorf("create error: %v\n", t.StrError)
	}

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("create error: status code %d", resp.StatusCode)
	}

	if t.TradeOfferID == 0 {
		return 0, newSteamErrorf("create error: steam returned 0 for trade offer id")
	}

	return t.TradeOfferID, nil
}

func (c *Client) GetOwnInventory(contextID uint64, appID uint32) (*inventory.Inventory, error) {
	return inventory.GetOwnInventory(c.client, contextID, appID)
}

func (c *Client) GetPartnerInventory(
	other steamid.SteamID,
	contextID uint64,
	appID uint32,
	offerID *uint64,
) (*inventory.Inventory, error) {
	return inventory.GetFullInventory(func() (*inventory.PartialInventory, error) {
		return c.getPartialPartnerInventory(other, contextID, appID, offerID, nil)
	}, func(start uint) (*inventory.PartialInventory, error) {
		return c.getPartialPartnerInventory(other, contextID, appID, offerID, &start)
	})
}

func (c *Client) getPartialPartnerInventory(
	other steamid.SteamID,
	contextID uint64,
	appID uint32,
	offerID *uint64,
	start *uint,
) (*inventory.PartialInventory, error) {
	data := map[string]string{
		"sessionid": c.sessionID,
		"partner":   other.FormatString(),
		"contextid": strconv.FormatUint(contextID, 10),
		"appid":     strconv.FormatUint(uint64(appID), 10),
	}
	if start != nil {
		data["start"] = strconv.FormatUint(uint64(*start), 10)
	}

	baseURL := "https://steamcommunity.com/tradeoffer/%v/"

	if offerID != nil {
		baseURL = fmt.Sprintf(baseURL, *offerID)
	} else {
		baseURL = fmt.Sprintf(baseURL, "new")
	}

	req, err := http.NewRequest("GET", baseURL+"partnerinventory/?"+netutil.ToURLValues(data).Encode(), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Referer", baseURL+"?partner="+other.AccountID().FormatString())

	return inventory.PerformRequest(c.client, req)
}

// Can be used to verify accepted tradeoffer and find out received asset ids
func (c *Client) GetTradeReceipt(tradeID uint64) ([]*TradeReceiptItem, error) {
	url := fmt.Sprintf("https://steamcommunity.com/trade/%d/receipt", tradeID)
	resp, err := c.client.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	items, err := parseTradeReceipt(respBody)

	if err != nil {
		return nil, newSteamErrorf("failed to parse trade receipt: %v", err)
	}

	return items, nil
}

// Get duration of escrow in days. Call this before sending a trade offer
func (c *Client) GetPartnerEscrowDuration(other steamid.SteamID, accessToken *string) (*EscrowDuration, error) {
	data := map[string]string{
		"partner": other.AccountID().FormatString(),
	}
	if accessToken != nil {
		data["token"] = *accessToken
	}
	return c.getEscrowDuration("https://steamcommunity.com/tradeoffer/new/?" + netutil.ToURLValues(data).Encode())
}

// Get duration of escrow in days. Call this after receiving a trade offer
func (c *Client) GetOfferEscrowDuration(offerID uint64) (*EscrowDuration, error) {
	return c.getEscrowDuration("http://steamcommunity.com/tradeoffer/" + strconv.FormatUint(offerID, 10))
}

func (c *Client) getEscrowDuration(queryURL string) (*EscrowDuration, error) {
	resp, err := c.client.Get(queryURL)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve escrow duration: %v", err)
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	escrowDuration, err := parseEscrowDuration(respBody)

	if err != nil {
		return nil, newSteamErrorf("failed to parse escrow duration: %v", err)
	}
	return escrowDuration, nil
}

func (c *Client) GetOfferWithRetry(
	offerID uint64,
	retryCount int,
	retryDelay time.Duration,
) (*Result, error) {
	var res *Result

	return res, withRetry(
		func() (err error) {
			res, err = c.GetOffer(offerID)
			return err
		}, retryCount, retryDelay)
}

func (c *Client) GetOffersWithRetry(
	getSent bool,
	getReceived bool,
	getDescriptions bool,
	activeOnly bool,
	historicalOnly bool,
	timeHistoricalCutoff *uint32,
	retryCount int,
	retryDelay time.Duration,
) (*MultiResult, error) {
	var res *MultiResult
	return res, withRetry(
		func() (err error) {
			res, err = c.GetOffers(getSent, getReceived, getDescriptions, activeOnly, historicalOnly, timeHistoricalCutoff)
			return err
		}, retryCount, retryDelay)
}

func (c *Client) DeclineWithRetry(offerID uint64, retryCount int, retryDelay time.Duration) error {
	return withRetry(
		func() error {
			return c.Decline(offerID)
		}, retryCount, retryDelay)
}

func (c *Client) CancelWithRetry(offerID uint64, retryCount int, retryDelay time.Duration) error {
	return withRetry(
		func() error {
			return c.Cancel(offerID)
		}, retryCount, retryDelay)
}

func (c *Client) AcceptWithRetry(offerID uint64, retryCount int, retryDelay time.Duration) error {
	return withRetry(
		func() error {
			return c.Accept(offerID)
		}, retryCount, retryDelay)
}

func (c *Client) CreateWithRetry(
	other steamid.SteamID,
	accessToken *string,
	myItems, theirItems []TradeItem,
	counteredOfferID *uint64,
	message string,
	retryCount int,
	retryDelay time.Duration,
) (uint64, error) {
	var res uint64

	return res, withRetry(
		func() (err error) {
			res, err = c.Create(other, accessToken, myItems, theirItems, counteredOfferID, message)
			return err
		}, retryCount, retryDelay)
}

func (c *Client) GetOwnInventoryWithRetry(
	contextID uint64,
	appID uint32,
	retryCount int,
	retryDelay time.Duration,
) (*inventory.Inventory, error) {
	var res *inventory.Inventory

	return res, withRetry(
		func() (err error) {
			res, err = c.GetOwnInventory(contextID, appID)
			return err
		}, retryCount, retryDelay)
}

func (c *Client) GetPartnerInventoryWithRetry(
	other steamid.SteamID,
	contextID uint64,
	appID uint32,
	offerID *uint64,
	retryCount int,
	retryDelay time.Duration,
) (*inventory.Inventory, error) {
	var res *inventory.Inventory

	return res, withRetry(
		func() (err error) {
			res, err = c.GetPartnerInventory(other, contextID, appID, offerID)
			return err
		}, retryCount, retryDelay,
	)
}

func (c *Client) GetTradeReceiptWithRetry(
	tradeID uint64,
	retryCount int,
	retryDelay time.Duration,
) ([]*TradeReceiptItem, error) {
	var res []*TradeReceiptItem
	return res, withRetry(
		func() (err error) {
			res, err = c.GetTradeReceipt(tradeID)
			return err
		}, retryCount, retryDelay)
}

// TODO: rewrite this
func withRetry(f func() error, retryCount int, retryDelay time.Duration) error {
	if retryCount <= 0 {
		return errors.New("retry count must be more than 0")
	}

	i := 0

	for {
		i++

		if err := f(); err != nil {
			// If we got steam error do not retry
			if _, ok := err.(*SteamError); ok {
				return err
			}

			if i == retryCount {
				return err
			}

			time.Sleep(retryDelay)

			continue
		}

		break
	}

	return nil
}
