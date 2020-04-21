/*
Implements methods to interact with the official Trade Offer API.

See: https://developer.valvesoftware.com/wiki/Steam_Web_API/IEconService
*/
package tradeoffer

import (
	"encoding/json"

	"github.com/13k/go-steam/economy/inventory"
	"github.com/13k/go-steam/steamid"
)

type State uint

const (
	// Invalid
	StateInvalid State = iota + 1
	// This trade offer has been sent, neither party has acted on it yet.
	StateActive
	// The trade offer was accepted by the recipient and items were exchanged.
	StateAccepted
	// The recipient made a counter offer
	StateCountered
	// The trade offer was not accepted before the expiration date
	StateExpired
	// The sender canceled the offer
	StateCanceled
	// The recipient declined the offer
	StateDeclined
	// Some of the items in the offer are no longer available (indicated by the missing flag in the
	// output)
	StateInvalidItems
	// The offer hasn't been sent yet and is awaiting email/mobile confirmation. The offer is only
	// visible to the sender.
	StateCreatedNeedsConfirmation
	// Either party canceled the offer via email/mobile. The offer is visible to both parties, even if
	// the sender canceled it before it was sent.
	StateCanceledBySecondFactor
	// The trade has been placed on hold. The items involved in the trade have all been removed from
	// both parties' inventories and will be automatically delivered in the future.
	StateInEscrow
)

type ConfirmationMethod uint

const (
	ConfirmationMethodInvalid ConfirmationMethod = iota
	ConfirmationMethodEmail
	ConfirmationMethodMobileApp
)

type Asset struct {
	AppID      uint32 `json:"-"`
	ContextID  uint64 `json:",string"`
	AssetID    uint64 `json:",string"`
	CurrencyID uint64 `json:",string"`
	ClassID    uint64 `json:",string"`
	InstanceID uint64 `json:",string"`
	Amount     uint64 `json:",string"`
	Missing    bool
}

type Offer struct {
	TradeOfferID       uint64             `json:",string"`
	TradeID            uint64             `json:",string"`
	OtherAccountID     uint32             `json:"accountid_other"`
	OtherSteamID       steamid.SteamID    `json:"-"`
	Message            string             `json:"message"`
	ExpirationTime     uint32             `json:"expiraton_time"`
	State              State              `json:"trade_offer_state"`
	ToGive             []*Asset           `json:"items_to_give"`
	ToReceive          []*Asset           `json:"items_to_receive"`
	IsOurOffer         bool               `json:"is_our_offer"`
	TimeCreated        uint32             `json:"time_created"`
	TimeUpdated        uint32             `json:"time_updated"`
	EscrowEndDate      uint32             `json:"escrow_end_date"`
	ConfirmationMethod ConfirmationMethod `json:"confirmation_method"`
}

func (t *Offer) UnmarshalJSON(data []byte) error {
	type Alias Offer
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if t.OtherAccountID == 0 {
		t.OtherSteamID = steamid.SteamID(0)
		return nil
	}
	t.OtherSteamID = steamid.SteamID(uint64(t.OtherAccountID) + 76561197960265728)
	return nil
}

type MultiResult struct {
	Sent         []*Offer `json:"trade_offers_sent"`
	Received     []*Offer `json:"trade_offers_received"`
	Descriptions []*Description
}

type Result struct {
	Offer        *Offer
	Descriptions []*Description
}

type Description struct {
	AppID      uint32 `json:"appid"`
	ClassID    uint64 `json:"classid,string"`
	InstanceID uint64 `json:"instanceid,string"`

	IconURL      string `json:"icon_url"`
	IconLargeURL string `json:"icon_url_large"`

	Name           string
	MarketName     string `json:"market_name"`
	MarketHashName string `json:"market_hash_name"`

	// Colors in hex, for example `B2B2B2`
	NameColor       string `json:"name_color"`
	BackgroundColor string `json:"background_color"`

	Type string

	Tradable                  bool   `json:"tradable"`
	Commodity                 bool   `json:"commodity"`
	MarketTradableRestriction uint32 `json:"market_tradable_restriction"`

	Descriptions inventory.DescriptionLines `json:"descriptions"`
	Actions      []*inventory.Action        `json:"actions"`
}
