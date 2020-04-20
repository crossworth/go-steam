package steam

import (
	"errors"
	"math/rand"
	"time"

	"github.com/13k/go-steam/netutil"
)

//go:generate go run ./cmd/gen/cmservers steam servers_cmservers.go

func GetRandomCM() (*netutil.PortAddr, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := rng.Int31n(int32(len(CMServers)))
	addr := netutil.ParsePortAddr(CMServers[i])

	if addr == nil {
		return nil, errors.New("invalid address in CMServers slice")
	}

	return addr, nil
}
