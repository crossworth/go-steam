package steamid

import (
	"github.com/13k/go-steam-resources/steamlang"
)

const (
	AccountTypeOffset      uint   = 52
	AccountTypeMask        uint64 = 0xF
	AccountTypeRuneUnknown rune   = 'i'
)

type AccountType steamlang.EAccountType

func AccountTypeFromRune(r rune) AccountType {
	switch r {
	case 'I':
		return AccountType(steamlang.EAccountType_Invalid)
	case 'U':
		return AccountType(steamlang.EAccountType_Individual)
	case 'M':
		return AccountType(steamlang.EAccountType_Multiseat)
	case 'G':
		return AccountType(steamlang.EAccountType_GameServer)
	case 'A':
		return AccountType(steamlang.EAccountType_AnonGameServer)
	case 'P':
		return AccountType(steamlang.EAccountType_Pending)
	case 'C':
		return AccountType(steamlang.EAccountType_ContentServer)
	case 'g':
		return AccountType(steamlang.EAccountType_Clan)
	case 'T':
		return AccountType(steamlang.EAccountType_Chat)
	case 'L': // Lobby chat
		return AccountType(steamlang.EAccountType_Chat)
	case 'c': // Clan chat
		return AccountType(steamlang.EAccountType_Chat)
	case 'a':
		return AccountType(steamlang.EAccountType_AnonUser)
	default:
		return AccountType(steamlang.EAccountType_Invalid)
	}
}

func (t AccountType) Enum() steamlang.EAccountType {
	eType := steamlang.EAccountType(t)

	if eType <= steamlang.EAccountType_Invalid || eType >= steamlang.EAccountType_Max {
		eType = steamlang.EAccountType_Invalid
	}

	return eType
}

func (t AccountType) Rune(instance AccountInstance) rune {
	switch t.Enum() {
	case steamlang.EAccountType_AnonGameServer:
		return 'A'
	case steamlang.EAccountType_GameServer:
		return 'G'
	case steamlang.EAccountType_Multiseat:
		return 'M'
	case steamlang.EAccountType_Pending:
		return 'P'
	case steamlang.EAccountType_ContentServer:
		return 'C'
	case steamlang.EAccountType_Clan:
		return 'g'
	case steamlang.EAccountType_Chat:
		switch {
		case instance.HasChatFlag(ChatInstanceFlagClan):
			return 'c'
		case instance.HasChatFlag(ChatInstanceFlagLobby),
			instance.HasChatFlag(ChatInstanceFlagMMSLobby):
			return 'L'
		default:
			return 'T'
		}
	case steamlang.EAccountType_Invalid:
		return 'I'
	case steamlang.EAccountType_Individual:
		return 'U'
	case steamlang.EAccountType_AnonUser:
		return 'a'
	default:
		return AccountTypeRuneUnknown
	}
}

func (t AccountType) IsInvalid() bool {
	return t.Enum() == steamlang.EAccountType_Invalid
}

func (t AccountType) IsIndividual() bool {
	return t.Enum() == steamlang.EAccountType_Individual
}

func (t AccountType) IsMultiseat() bool {
	return t.Enum() == steamlang.EAccountType_Multiseat
}

func (t AccountType) IsGameServer() bool {
	return t.Enum() == steamlang.EAccountType_GameServer
}

func (t AccountType) IsAnonGameServer() bool {
	return t.Enum() == steamlang.EAccountType_AnonGameServer
}

func (t AccountType) IsPending() bool {
	return t.Enum() == steamlang.EAccountType_Pending
}

func (t AccountType) IsContentServer() bool {
	return t.Enum() == steamlang.EAccountType_ContentServer
}

func (t AccountType) IsClan() bool {
	return t.Enum() == steamlang.EAccountType_Clan
}

func (t AccountType) IsChat() bool {
	return t.Enum() == steamlang.EAccountType_Chat
}

func (t AccountType) IsConsoleUser() bool {
	return t.Enum() == steamlang.EAccountType_ConsoleUser
}

func (t AccountType) IsAnonUser() bool {
	return t.Enum() == steamlang.EAccountType_AnonUser
}
