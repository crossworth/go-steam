package steam

import (
	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
)

type TradeProposedEvent struct {
	RequestId TradeRequestId
	Other     steamid.SteamID `json:",string"`
}

type TradeResultEvent struct {
	RequestId TradeRequestId
	Response  steamlang.EEconTradeResponse
	Other     steamid.SteamID `json:",string"`
	// Number of days Steam Guard is required to have been active
	NumDaysSteamGuardRequired uint32
	// Number of days a new device cannot trade for.
	NumDaysNewDeviceCooldown uint32
	// Default number of days one cannot trade after a password reset.
	DefaultNumDaysPasswordResetProbation uint32
	// See above.
	NumDaysPasswordResetProbation uint32
}

type TradeSessionStartEvent struct {
	Other steamid.SteamID `json:",string"`
}
