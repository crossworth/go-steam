package steamid

import (
	"strconv"

	"github.com/13k/go-steam-resources/steamlang"
)

const (
	AccountIDOffset uint   = 0
	AccountIDMask   uint64 = 0xFFFFFFFF

	accountIDOnlyOffset  uint   = 1
	authServerOnlyOffset uint   = 0
	accountIDOnlyMask    uint32 = 0xFFFFFFFF
	authServerOnlyMask   uint32 = 0x1
)

// AccountID represents the "account number" part of a SteamID.
//
// It's a 32-bits field with two parts: the least significant bit indicates wether it account uses
// authentication or not, the highest 31 bits represent the account number.
type AccountID uint32

func NewAccountID(id uint32, authServer uint32) AccountID {
	return AccountID(0).
		SetID(id).
		SetAuthServer(authServer)
}

func (id AccountID) get(offset uint, mask uint32) uint32 {
	return (uint32(id) >> offset) & mask
}

func (id AccountID) set(offset uint, mask, value uint32) AccountID {
	return AccountID((uint32(id) & ^(mask << offset)) | (value&mask)<<offset)
}

func (id AccountID) ID() uint32 {
	return id.get(accountIDOnlyOffset, accountIDOnlyMask)
}

func (id AccountID) SetID(value uint32) AccountID {
	return id.set(accountIDOnlyOffset, accountIDOnlyMask, value)
}

func (id AccountID) AuthServer() uint32 {
	return id.get(authServerOnlyOffset, authServerOnlyMask)
}

func (id AccountID) SetAuthServer(value uint32) AccountID {
	return id.set(authServerOnlyOffset, authServerOnlyMask, value)
}

func (id AccountID) UsesAuthServer() bool {
	return id.AuthServer() != 0
}

// SteamID creates a SteamID with this AccountID and default type, universe and instance.
func (id AccountID) SteamID() SteamID {
	return New(steamlang.EAccountType_Individual, steamlang.EUniverse_Public, id, DesktopInstance)
}

func (id AccountID) Uint32() uint32 {
	return uint32(id)
}

func (id AccountID) FormatString() string {
	return strconv.FormatUint(uint64(id), 10)
}
