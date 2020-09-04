// +build windows

package sys

import (
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
)

// RootDiskUUID returns the volume guid for the System Drive
func RootDiskUUID() (uuid.UUID, error) {
	systemDrive := os.Getenv("SystemDrive")

	if systemDrive == "" {
		return uuid.Nil, nil
	}

	cmd := exec.Command("mountvol", systemDrive, "/L")
	output, err := cmd.Output()
	if err != nil {
		return uuid.Nil, err
	}

	uuidStr := strings.TrimSpace(string(output))
	uuidStr = strings.TrimPrefix(uuidStr, `\\?\Volume{`)
	uuidStr = strings.TrimSuffix(uuidStr, `}\`)
	return uuid.Parse(uuidStr)
}
