// +build darwin

package sys

import (
	"github.com/fsnotify/fsevents"
	"github.com/google/uuid"
)

func RootDiskUUID() (uuid.UUID, error) {
	devno, err := fsevents.DeviceForPath("/")

	if err != nil {
		return "", err
	}

	uuidStr := fsevents.GetDeviceUUID(devno)

	return uuid.Parse(uuidStr)
}
