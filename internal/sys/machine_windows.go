// +build windows

package sys

import (
	"os/exec"
	"strings"

	"github.com/google/uuid"
)

func MachineUUID() (uuid.UUID, error) {
	// we use the output from the command `wmic csproduct get UUID`
	cmd := exec.Command("wmic", "csproduct", "get", "UUID")
	output, err := cmd.Output()
	if err != nil {
		return uuid.Nil, err
	}

	outputStr := strings.TrimSpace(strings.ReplaceAll(string(output), "UUID", ""))
	return uuid.Parse(outputStr)
}
