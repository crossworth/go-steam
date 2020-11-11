// +build linux

package sys

import (
	"github.com/google/uuid"
)

func RootDiskUUID() (uuid.UUID, error) {
	return DefaultDiskUUID, nil
}
