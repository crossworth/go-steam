package sys

import (
	"net"
)

var DefaultMACAddress = MustParseMAC("00:00:00:00:00:00")

func MustParseMAC(s string) net.HardwareAddr {
	addr, err := net.ParseMAC(s)

	if err != nil {
		panic(err)
	}

	return addr
}
