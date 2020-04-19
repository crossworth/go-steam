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
	ClientSteamId  steamid.SteamId `json:",string"`
	Body           *pb.CMsgClientLogonResponse
}

type LogOnFailedEvent struct {
	Result steamlang.EResult
}

type LoginKeyEvent struct {
	UniqueId uint32
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
	FacebookId           uint64 `json:",string"`
	FacebookName         string
}

// Returned when Steam is down for some reason.
// A disconnect will follow, probably.
type SteamFailureEvent struct {
	Result steamlang.EResult
}
