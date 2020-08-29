// +build darwin

package sys

import (
	"os/exec"
	"strings"

	"github.com/google/uuid"
)

func MachineUUID() (uuid.UUID, error) {
	// we parse the output of `ioreg -rd1 -c IOPlatformExpertDevice` (IOPlatformUUID)
	cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return uuid.Nil, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if !strings.Contains(line, "IOPlatformUUID") {
			continue
		}

		l := strings.TrimSpace(line)
		l = strings.TrimPrefix(l, `"IOPlatformUUID" = "`)
		l = strings.TrimSuffix(l, `"`)

		if len(l) > 0 {
			return uuid.Parse(l)
		}
	}

	return DefaultMachineUUID, nil
}
