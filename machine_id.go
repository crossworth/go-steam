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
	machineUUID, err := sys.MachineUUID()

	if err != nil {
		return nil, err
	}

	if len(machineUUID) == 0 {
		machineUUID = sys.DefaultMachineUUID
	}

	iface, err := sys.FirstPublicInterface()

	if err != nil {
		return nil, err
	}

	macAddr := sys.DefaultMACAddress

	if iface != nil {
		macAddr = iface.HardwareAddr
	}

	diskUUID, err := sys.RootDiskUUID()

	if err != nil {
		return nil, err
	}

	if len(diskUUID) == 0 {
		diskUUID = sys.DefaultDiskUUID
	}

	mid := &MachineID{
		MachineUUID: machineUUID,
		MacAddress:  macAddr,
		DiskUUID:    diskUUID,
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
