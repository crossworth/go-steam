// Package steamid provides types and functions to represent and manipulate a SteamID.
//
// https://developer.valvesoftware.com/wiki/SteamID
package steamid

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/13k/go-steam-resources/steamlang"
)

var (
	// STEAM_X:Y:Z
	steam2RE = regexp.MustCompile(`^STEAM_(?P<universe>[0-4]):(?P<auth_server>[0-1]):(?P<account_id>\d+)$`) //nolint:lll
	// [X:Y:Z:W] -- `:W` optional
	steam3RE = regexp.MustCompile(`^\[(?P<type>[AGMPCgcLTIUai]):(?P<universe>[0-4]):(?P<account_id>\d+)(?::(?P<instance>\d+))?\]$`) //nolint:lll
)

// SteamID is a steam identifier.
type SteamID uint64

// New creates a SteamID with explicit parameters.
func New(
	accountType steamlang.EAccountType,
	universe steamlang.EUniverse,
	accountID AccountID,
	instance AccountInstance,
) SteamID {
	return SteamID(0).
		SetAccountType(accountType).
		SetAccountUniverse(universe).
		SetAccountID(accountID).
		SetAccountInstance(instance)
}

// Parse attempts to Parse a SteamID from the given string.
//
// It tries to Parse the string with `ParseSteam2` and if it does not match, tries to Parse the
// value with `ParseSteam3`.
//
// It returns zero and a nil error if the string does not match any of the formats.
//
// It returns a non-nil error if any of the `ParseSteam*` functions returned a non-nil error.
func Parse(s string) (SteamID, error) {
	id, err := ParseSteam2(s)

	if err != nil {
		return 0, err
	}

	if id != 0 {
		return id, nil
	}

	id, err = ParseSteam3(s)

	if err != nil {
		return 0, err
	}

	if id != 0 {
		return id, nil
	}

	return 0, nil
}

// ParseSteam2 attempts to parse a SteamID from the given string in the steam2 ID format
// (`STEAM_X:Y:Z`).
//
// It returns zero and a nil error if the string does not match the steam2 ID format.
//
// It returns a non-nil error if the string matches the steam2 ID format but could not be parsed.
func ParseSteam2(s string) (SteamID, error) {
	match := steam2RE.FindStringSubmatch(s)

	if match == nil {
		return 0, nil
	}

	universeStr, authserverStr, accountIDStr := match[1], match[2], match[3]

	universe64, err := strconv.ParseInt(universeStr, 10, 32)

	if err != nil {
		return 0, err
	}

	universe := steamlang.EUniverse(universe64)

	authserver64, err := strconv.ParseUint(authserverStr, 10, 32)

	if err != nil {
		return 0, err
	}

	authserver := uint32(authserver64)

	accountID64, err := strconv.ParseUint(accountIDStr, 10, 32)

	if err != nil {
		return 0, err
	}

	accountID := AccountID((uint32(accountID64) << 1) | authserver)

	id := New(
		steamlang.EAccountType_Individual,
		universe,
		accountID,
		DesktopInstance,
	)

	return id, nil
}

// ParseSteam3 attempts to parse a SteamID from the given string in the Steam3 ID format
// (`[X:Y:Z:W]`).
//
// In the Steam3 format, the last `:W` part is optional.
//
// It returns zero and a nil error if the string does not match the Steam3 ID format.
//
// It returns a non-nil error if the string matches the Steam3 ID format but could not be parsed.
func ParseSteam3(s string) (SteamID, error) {
	match := steam3RE.FindStringSubmatch(s)

	if match == nil {
		return 0, nil
	}

	typeStr, universeStr, accountIDStr, instanceStr := match[1], match[2], match[3], match[4]
	typeRune := rune(typeStr[0])

	universe64, err := strconv.ParseInt(universeStr, 10, 32)

	if err != nil {
		return 0, err
	}

	universe := steamlang.EUniverse(universe64)

	accountID64, err := strconv.ParseUint(accountIDStr, 10, 32)

	if err != nil {
		return 0, err
	}

	accountID := AccountID(accountID64)

	var instance AccountInstance

	if instanceStr != "" {
		instance, err = AccountInstanceFromString(instanceStr)

		if err != nil {
			return 0, err
		}
	} else {
		switch typeRune {
		case 'g', 'T', 'c', 'L':
			instance = UnknownInstance
		default:
			instance = DesktopInstance
		}
	}

	switch typeRune {
	case 'c':
		instance = instance.SetChatFlags(ChatInstanceFlagClan)
	case 'L':
		instance = instance.SetChatFlags(ChatInstanceFlagLobby)
	}

	accountType := AccountTypeFromRune(typeRune)

	id := New(
		accountType.Enum(),
		universe,
		accountID,
		instance,
	)

	return id, nil
}

func (id SteamID) get(offset uint, mask uint64) uint64 {
	return (uint64(id) >> offset) & mask
}

func (id SteamID) set(offset uint, mask, value uint64) SteamID {
	return SteamID((uint64(id) & ^(mask << offset)) | (value&mask)<<offset)
}

func (id SteamID) AccountID() AccountID {
	return AccountID(id.get(AccountIDOffset, AccountIDMask))
}

func (id SteamID) SetAccountID(aid AccountID) SteamID {
	return id.set(AccountIDOffset, AccountIDMask, uint64(aid))
}

func (id SteamID) AccountInstance() AccountInstance {
	return AccountInstance(id.get(AccountInstanceOffset, AccountInstanceMask))
}

func (id SteamID) SetAccountInstance(i AccountInstance) SteamID {
	return id.set(AccountInstanceOffset, AccountInstanceMask, uint64(i))
}

func (id SteamID) AccountType() AccountType {
	eType := steamlang.EAccountType(id.get(AccountTypeOffset, AccountTypeMask))

	if eType <= steamlang.EAccountType_Invalid || eType >= steamlang.EAccountType_Max {
		eType = steamlang.EAccountType_Invalid
	}

	return AccountType(eType)
}

func (id SteamID) SetAccountType(t steamlang.EAccountType) SteamID {
	return id.set(AccountTypeOffset, AccountTypeMask, uint64(t))
}

func (id SteamID) AccountUniverse() steamlang.EUniverse {
	eUniv := steamlang.EUniverse(id.get(AccountUniverseOffset, AccountUniverseMask))

	if eUniv <= steamlang.EUniverse_Invalid || eUniv >= steamlang.EUniverse_Max {
		eUniv = steamlang.EUniverse_Invalid
	}

	return eUniv
}

func (id SteamID) SetAccountUniverse(u steamlang.EUniverse) SteamID {
	return id.set(AccountUniverseOffset, AccountUniverseMask, uint64(u))
}

// ClanToChat returns a copy of this SteamID with type changed from Clan to Chat and the instance
// flagged with `ChatInstanceFlagClan`.
//
// If the SteamID's type is not Clan, it returns the receiver.
func (id SteamID) ClanToChat() SteamID {
	if !id.AccountType().IsClan() {
		return id
	}

	inst := id.
		AccountInstance().
		ClearChatFlags().
		SetChatFlags(ChatInstanceFlagClan)

	return id.
		SetAccountInstance(inst).
		SetAccountType(steamlang.EAccountType_Chat)
}

// ChatToClan returns a copy of this SteamID with type changed from Chat to Clan and the instance
// with Chat flags reset.
//
// If the SteamID's type is not Chat, it returns the receiver.
func (id SteamID) ChatToClan() SteamID {
	if !id.AccountType().IsChat() {
		return id
	}

	inst := id.
		AccountInstance().
		ClearChatFlags()

	return id.
		SetAccountInstance(inst).
		SetAccountType(steamlang.EAccountType_Clan)
}

func (id SteamID) Uint64() uint64 {
	return uint64(id)
}

func (id SteamID) FormatString() string {
	return strconv.FormatUint(uint64(id), 10)
}

func (id SteamID) String() string {
	return id.Steam3()
}

// Steam2 generates a string in the steam2 ID format.
func (id SteamID) Steam2() string {
	accountID := id.AccountID()

	return fmt.Sprintf(
		"STEAM_%d:%d:%d",
		id.AccountUniverse(),
		accountID.AuthServer(),
		accountID.ID(),
	)
}

// Steam3 generates a string in the steam3 ID format.
//
// It always renders the account instance, even if it's zero.
func (id SteamID) Steam3() string {
	inst := id.AccountInstance()
	typeRune := id.AccountType().Rune(inst)
	inst = inst.ClearChatFlags()

	return fmt.Sprintf(
		"[%s:%d:%d:%d]",
		string(typeRune),
		id.AccountUniverse(),
		id.AccountID(),
		inst,
	)
}
