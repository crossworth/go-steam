package sys

import (
	"fmt"
	"net"
)

type Interface struct {
	net.Interface

	IPNetworks []*net.IPNet
}

func Interfaces() ([]*Interface, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		return nil, err
	}

	infos := make([]*Interface, len(ifaces))

	for i, iface := range ifaces {
		addrs, err := iface.Addrs()

		if err != nil {
			return nil, err
		}

		info := &Interface{
			Interface:  iface,
			IPNetworks: make([]*net.IPNet, 0, len(addrs)),
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				info.IPNetworks = append(info.IPNetworks, ipnet)
			}
		}

		infos[i] = info
	}

	return infos, nil
}

func FirstPublicInterface() (*Interface, error) {
	conn, err := net.Dial("tcp", "1.1.1.1:80")

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	localAddr := conn.LocalAddr()
	tcpAddr, ok := localAddr.(*net.TCPAddr)

	if !ok {
		return nil, fmt.Errorf("local address is not a TCPAddr, got %T", localAddr)
	}

	ifaces, err := Interfaces()

	if err != nil {
		return nil, err
	}

	var pubIface *Interface

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagPointToPoint != 0 {
			continue
		}

		for _, ipnet := range iface.IPNetworks {
			if ipnet.Contains(tcpAddr.IP) {
				pubIface = iface
				break
			}
		}
	}

	return pubIface, nil
}
