package trade

import (
	"errors"
	"time"

	"github.com/13k/go-steam/economy/trade/api"
	"github.com/13k/go-steam/steamid"
)

const pollTimeout = time.Second

type Trade struct {
	ThemID    steamid.SteamID
	MeReady   bool
	ThemReady bool

	lastPoll     time.Time
	queuedEvents []interface{}
	api          *api.Client
}

func New(sessionID, steamLogin, steamLoginSecure string, other steamid.SteamID) (*Trade, error) {
	client, err := api.New(sessionID, steamLogin, steamLoginSecure, other)

	if err != nil {
		return nil, err
	}

	t := &Trade{
		ThemID:   other,
		lastPoll: time.Unix(0, 0),
		api:      client,
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

func (t *Trade) onStatus(status *api.Result) error {
	if !status.Success {
		return errors.New("trade: returned status not successful! error message: " + status.Error)
	}

	if status.NewVersion {
		t.api.Version = status.Version
		t.MeReady = bool(status.Me.Ready)
		t.ThemReady = bool(status.Them.Ready)
	}

	switch status.TradeStatus {
	case api.StatusComplete:
		t.addEvent(&EndEvent{EndReasonComplete})
	case api.StatusCanceled:
		t.addEvent(&EndEvent{EndReasonCanceled})
	case api.StatusTimeout:
		t.addEvent(&EndEvent{EndReasonTimeout})
	case api.StatusFailed:
		t.addEvent(&EndEvent{EndReasonFailed})
	case api.StatusOpen:
		// nothing
	default:
		// ignore too
	}

	t.updateEvents(status.Events)
	return nil
}

func (t *Trade) updateEvents(events api.EventList) {
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
		case api.ActionAddItem:
			t.addEvent(&ItemAddedEvent{newItem(event)})
		case api.ActionRemoveItem:
			t.addEvent(&ItemRemovedEvent{newItem(event)})
		case api.ActionReady:
			t.ThemReady = true
			t.addEvent(&ReadyEvent{})
		case api.ActionUnready:
			t.ThemReady = false
			t.addEvent(&UnreadyEvent{})
		case api.ActionSetCurrency:
			t.addEvent(&SetCurrencyEvent{
				newCurrency(event),
				event.OldAmount,
				event.NewAmount,
			})
		case api.ActionChatMessage:
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
