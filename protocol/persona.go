package protocol

import (
	"github.com/13k/go-steam-resources/steamlang"
)

// The default details to request in most situations
const DefaultPersonaStateFlagInfoRequest = steamlang.EClientPersonaStateFlag_PlayerName |
	steamlang.EClientPersonaStateFlag_Presence |
	steamlang.EClientPersonaStateFlag_SourceID |
	steamlang.EClientPersonaStateFlag_GameExtraInfo

const DefaultAvatar = "fef49e7fa7e1997310d705b2a6158ff8dc1cdfeb"

func ValidAvatar(avatar string) bool {
	return !(avatar == "0000000000000000000000000000000000000000" || len(avatar) != 40)
}
