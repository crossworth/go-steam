package trade

import (
	"github.com/13k/go-steam/trade/tradeapi"
)

type EndReason uint

const (
	EndReasonComplete EndReason = iota + 1
	EndReasonCanceled
	EndReasonTimeout
	EndReasonFailed
)

type EndEvent struct {
	Reason EndReason
}

func newItem(event *tradeapi.Event) *Item {
	return &Item{
		event.AppID,
		event.ContextID,
		event.AssetID,
	}
}

type Item struct {
	AppID     uint32
	ContextID uint64
	AssetID   uint64
}

type ItemAddedEvent struct {
	Item *Item
}

type ItemRemovedEvent struct {
	Item *Item
}

type ReadyEvent struct{}
type UnreadyEvent struct{}

func newCurrency(event *tradeapi.Event) *Currency {
	return &Currency{
		event.AppID,
		event.ContextID,
		event.CurrencyID,
	}
}

type Currency struct {
	AppID      uint32
	ContextID  uint64
	CurrencyID uint64
}

type SetCurrencyEvent struct {
	Currency  *Currency
	OldAmount uint64
	NewAmount uint64
}

type ChatEvent struct {
	Message string
}
