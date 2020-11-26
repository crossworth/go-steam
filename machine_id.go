package steam

import (
	"net"

	"github.com/13k/go-steam/cryptoutil"
	"github.com/13k/go-steam/internal/sys"
	"github.com/google/uuid"
)

const (
	MachineIDKeyMachineUUID = "BB3"
	MachineIDKeyMacAddress  = "FF2"
	MachineIDKeyDiskUUID    = "3B3"
	MachineIDKeyUnknownData = "333"
)

// MachineID identifies the machine running the Steam client.
type MachineID struct {
	MachineUUID uuid.UUID
	DiskUUID    uuid.UUID
	MacAddress  net.HardwareAddr

	mo *MessageObject
}

func NewMachineID() (*MachineID, error) {

	mid := &MachineID{
		MachineUUID: uuid.Nil,
		MacAddress:  sys.DefaultMACAddress,
		DiskUUID:    uuid.Nil,
	}

	return mid, nil
}

func (id *MachineID) MessageObject() *MessageObject {
	if id.mo == nil {
		id.mo = NewMessageObject().
			AddString(MachineIDKeyMachineUUID, cryptoutil.SHA1String(id.MachineUUID[:])).
			AddString(MachineIDKeyMacAddress, cryptoutil.SHA1String(id.MacAddress)).
			AddString(MachineIDKeyDiskUUID, cryptoutil.SHA1String(id.DiskUUID[:]))
	}

	return id.mo
}

func (id *MachineID) Auth() ([]byte, error) {
	return id.MessageObject().MarshalBinary()
}
