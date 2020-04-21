package steamid

import (
	"strconv"
)

const (
	AccountInstanceOffset uint   = 32
	AccountInstanceMask   uint64 = 0xFFFFF

	instanceOnlyMask AccountInstance = 0x1FFFF
)

// ChatInstanceFlag is a flag a chat SteamID may have.
type ChatInstanceFlag uint32

const (
	// ChatInstanceFlagClan is set for clan based chat steam IDs.
	ChatInstanceFlagClan = ChatInstanceFlag((AccountInstanceMask + 1) >> (iota + 1))
	// ChatInstanceFlagLobby is set for lobby based chat steam IDs.
	ChatInstanceFlagLobby
	// ChatInstanceFlagMMSLobby is set for matchmaking lobby based chat steam IDs.
	ChatInstanceFlagMMSLobby
)

// AccountInstance is an instance of an account.
//
// It's a 20-bit bitfield where the lowest 17 bits store the `*Instance` flags and the highest 3
// bits store the `ChatInstanceFlag*` flags.
type AccountInstance uint32

const (
	UnknownInstance AccountInstance = 0
)

const (
	// DesktopInstance is the account instance value for a desktop.
	DesktopInstance AccountInstance = 1 << iota
	// ConsoleInstance is the account instance value for a console.
	ConsoleInstance
	// WebInstance is the account instance value for mobile or web-based.
	WebInstance
)

func AccountInstanceFromString(s string) (AccountInstance, error) {
	instance64, err := strconv.ParseUint(s, 10, 32)

	if err != nil {
		return UnknownInstance, err
	}

	return AccountInstance(instance64), nil
}

func (i AccountInstance) IsDesktop() bool {
	return i&DesktopInstance != 0
}

func (i AccountInstance) IsConsole() bool {
	return i&ConsoleInstance != 0
}

func (i AccountInstance) IsWeb() bool {
	return i&WebInstance != 0
}

func (i AccountInstance) HasChatFlag(flag ChatInstanceFlag) bool {
	return i&AccountInstance(flag) != 0
}

func (i AccountInstance) SetChatFlags(flags ...ChatInstanceFlag) AccountInstance {
	for _, flag := range flags {
		i = i | AccountInstance(flag)
	}

	return i
}

func (i AccountInstance) ClearChatFlags() AccountInstance {
	return i & instanceOnlyMask
}
