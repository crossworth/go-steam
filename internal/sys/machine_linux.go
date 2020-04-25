// +build linux

package sys

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"

	"github.com/google/uuid"
)

var (
	machineIDFiles = []string{
		"/etc/machine-id",
		"/var/lib/dbus/machine-id",
	}
)

func MachineUUID() (uuid.UUID, error) {
	for _, f := range machineIDFiles {
		bUUID, err := ioutil.ReadFile(f)

		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			return uuid.Nil, err
		}

		return uuid.ParseBytes(bytes.TrimSpace(bUUID))
	}

	return DefaultMachineUUID, nil
}
