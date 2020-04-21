package tradeapi

import (
	"encoding/json"
	"strconv"

	"github.com/13k/go-steam/jsont"
	"github.com/13k/go-steam/steamid"
)

type Result struct {
	Success     bool
	Error       string
	NewVersion  bool   `json:"newversion"`
	TradeStatus Status `json:"trade_status"`
	Version     uint
	LogPos      int
	Me          User
	Them        User
	Events      EventList
}

type Status uint

const (
	StatusOpen Status = iota
	StatusComplete
	StatusEmpty // when both parties trade no items
	StatusCanceled
	StatusTimeout // the partner timed out
	StatusFailed
)

type EventList map[uint]*Event

// The EventList can either be an array or an object of id -> event
func (e *EventList) UnmarshalJSON(data []byte) error {
	// initialize the map if it's nil
	if *e == nil {
		*e = make(EventList)
	}

	o := make(map[string]*Event)

	// it's an object
	if err := json.Unmarshal(data, &o); err == nil {
		for is, event := range o {
			var i uint64

			i, err = strconv.ParseUint(is, 10, 32)

			if err != nil {
				return err
			}

			(*e)[uint(i)] = event
		}

		return nil
	}

	// it should be an array
	var a []*Event

	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	for i, event := range a {
		(*e)[uint(i)] = event
	}

	return nil
}

type Event struct {
	SteamID   steamid.SteamID `json:",string"`
	Action    Action          `json:",string"`
	Timestamp uint64

	AppID     uint32
	ContextID uint64 `json:",string"`
	AssetID   uint64 `json:",string"`

	Text string // only used for chat messages

	// The following is used for SetCurrency
	CurrencyID uint64 `json:",string"`
	OldAmount  uint64 `json:"old_amount,string"`
	NewAmount  uint64 `json:"amount,string"`
}

type Action uint

const (
	ActionAddItem Action = iota
	ActionRemoveItem
	ActionReady
	ActionUnready
	ActionAccept
	_ // skip
	ActionSetCurrency
	ActionChatMessage
)

type User struct {
	Ready             jsont.UintBool
	Confirmed         jsont.UintBool
	SecSinceTouch     int  `json:"sec_since_touch"`
	ConnectionPending bool `json:"connection_pending"`
	Assets            interface{}
	Currency          interface{} // either []*Currency or empty string
}

type Currency struct {
	AppID      uint64 `json:",string"`
	ContextID  uint64 `json:",string"`
	CurrencyID uint64 `json:",string"`
	Amount     uint64 `json:",string"`
}
