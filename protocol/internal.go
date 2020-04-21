package protocol

import (
	"io"
	"math"
	"strconv"

	"github.com/13k/go-steam-resources/steamlang"
)

type JobID uint64

func (j JobID) String() string {
	if j == math.MaxUint64 {
		return "(none)"
	}
	return strconv.FormatUint(uint64(j), 10)
}

type Serializer interface {
	Serialize(w io.Writer) error
}

type Deserializer interface {
	Deserialize(r io.Reader) error
}

type Serializable interface {
	Serializer
	Deserializer
}

type MessageBody interface {
	Serializable
	GetEMsg() steamlang.EMsg
}

// The default details to request in most situations
const DefaultPersonaStateFlagInfoRequest = steamlang.EClientPersonaStateFlag_PlayerName |
	steamlang.EClientPersonaStateFlag_Presence |
	steamlang.EClientPersonaStateFlag_SourceID |
	steamlang.EClientPersonaStateFlag_GameExtraInfo

const DefaultAvatar = "fef49e7fa7e1997310d705b2a6158ff8dc1cdfeb"

func ValidAvatar(avatar string) bool {
	return !(avatar == "0000000000000000000000000000000000000000" || len(avatar) != 40)
}
