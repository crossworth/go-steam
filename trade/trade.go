package trade

import (
	"errors"
	"time"

	"github.com/13k/go-steam/steamid"
	"github.com/13k/go-steam/trade/tradeapi"
)

const pollTimeout = time.Second

type Trade struct {
	ThemID    steamid.SteamID
	MeReady   bool
	ThemReady bool

	lastPoll     time.Time
	queuedEvents []interface{}
	api          *tradeapi.Trade
}

func New(sessionID, steamLogin, steamLoginSecure string, other steamid.SteamID) (*Trade, error) {
	api, err := tradeapi.New(sessionID, steamLogin, steamLoginSecure, other)

	if err != nil {
		return nil, err
	}

	t := &Trade{
		ThemID:       other,
		MeReady:      false,
		ThemReady:    false,
		lastPoll:     time.Unix(0, 0),
		queuedEvents: nil,
		api:          api,
	}

	return t, nil
}

func (t *Trade) Version() uint {
	return t.api.Version
}

// Returns all queued events and removes them from the queue without performing a HTTP request, like Poll() would.
func (t *Trade) Events() []interface{} {
	qe := t.queuedEvents
	t.queuedEvents = nil
	return qe
}

func (t *Trade) onStatus(status *tradeapi.Result) error {
	if !status.Success {
		return errors.New("trade: returned status not successful! error message: " + status.Error)
	}

	if status.NewVersion {
		t.api.Version = status.Version
		t.MeReady = bool(status.Me.Ready)
		t.ThemReady = bool(status.Them.Ready)
	}

	switch status.TradeStatus {
	case tradeapi.StatusComplete:
		t.addEvent(&EndEvent{EndReasonComplete})
	case tradeapi.StatusCanceled:
		t.addEvent(&EndEvent{EndReasonCanceled})
	case tradeapi.StatusTimeout:
		t.addEvent(&EndEvent{EndReasonTimeout})
	case tradeapi.StatusFailed:
		t.addEvent(&EndEvent{EndReasonFailed})
	case tradeapi.StatusOpen:
		// nothing
	default:
		// ignore too
	}

	t.updateEvents(status.Events)
	return nil
}

func (t *Trade) updateEvents(events tradeapi.EventList) {
	if len(events) == 0 {
		return
	}

	var lastLogPos uint

	for i, event := range events {
		if i < t.api.LogPos {
			continue
		}

		if event.SteamID != t.ThemID {
			continue
		}

		if lastLogPos < i {
			lastLogPos = i
		}

		switch event.Action {
		case tradeapi.ActionAddItem:
			t.addEvent(&ItemAddedEvent{newItem(event)})
		case tradeapi.ActionRemoveItem:
			t.addEvent(&ItemRemovedEvent{newItem(event)})
		case tradeapi.ActionReady:
			t.ThemReady = true
			t.addEvent(&ReadyEvent{})
		case tradeapi.ActionUnready:
			t.ThemReady = false
			t.addEvent(&UnreadyEvent{})
		case tradeapi.ActionSetCurrency:
			t.addEvent(&SetCurrencyEvent{
				newCurrency(event),
				event.OldAmount,
				event.NewAmount,
			})
		case tradeapi.ActionChatMessage:
			t.addEvent(&ChatEvent{
				event.Text,
			})
		}
	}

	t.api.LogPos = lastLogPos + 1
}

func (t *Trade) addEvent(event interface{}) {
	t.queuedEvents = append(t.queuedEvents, event)
}
