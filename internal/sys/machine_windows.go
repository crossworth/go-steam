// +build windows

package sys

import (
	"os/exec"
	"strings"

	"github.com/google/uuid"
)

func MachineUUID() (uuid.UUID, error) {
	// we use the output from the command `reg query HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography /v MachineGuid`
	cmd := exec.Command("reg", "query", `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography`, "/v", "MachineGuid")
	output, err := cmd.Output()
	if err != nil {
		return uuid.Nil, err
	}

	uuidStr := strings.ReplaceAll(string(output), `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography`, "")
	uuidStr = strings.ReplaceAll(uuidStr, "MachineGuid", "")
	uuidStr = strings.ReplaceAll(uuidStr, "REG_SZ", "")
	return uuid.Parse(strings.TrimSpace(uuidStr))
}
