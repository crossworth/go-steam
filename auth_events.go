package steam

import (
	pb "github.com/13k/go-steam-resources/protobuf/steam"
	"github.com/13k/go-steam-resources/steamlang"
	"github.com/13k/go-steam/steamid"
)

type LoggedOnEvent struct {
	Result         steamlang.EResult
	ExtendedResult steamlang.EResult
	AccountFlags   steamlang.EAccountFlags
	ClientSteamID  steamid.SteamID `json:",string"`
	Body           *pb.CMsgClientLogonResponse
}

type LogOnFailedEvent struct {
	Result steamlang.EResult
}

type LoginKeyEvent struct {
	UniqueID uint32
	LoginKey string
}

type LoggedOffEvent struct {
	Result steamlang.EResult
}

type MachineAuthUpdateEvent struct {
	Hash []byte
}

type AccountInfoEvent struct {
	PersonaName          string
	Country              string
	CountAuthedComputers int32
	AccountFlags         steamlang.EAccountFlags
	FacebookID           uint64 `json:",string"`
	FacebookName         string
}

// FailureEvent is emitted when Steam is down for some reason.
type FailureEvent struct {
	Result steamlang.EResult
}
