package steam

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/13k/go-steam/netutil"
)

// Load initial server list from Steam Directory Web API.
// Call InitializeSteamDirectory() before Connect() to use
// steam directory server list instead of static one.
func InitializeSteamDirectory() error {
	return steamDirectoryCache.Initialize(0)
}

// InitializeSteamDirectoryWithCellID
// loads the initial server list from Steam Directory Web API for the provided Cell ID
func InitializeSteamDirectoryWithCellID(cellID uint32) error {
	return steamDirectoryCache.Initialize(cellID)
}

var steamDirectoryCache *steamDirectory = &steamDirectory{}

type steamDirectory struct {
	sync.RWMutex
	servers       []string
	isInitialized bool
}

// Get server list from steam directory and save it for later
func (sd *steamDirectory) Initialize(cellID uint32) error {
	sd.Lock()
	defer sd.Unlock()

	client := &http.Client{}
	resp, err := client.Get(fmt.Sprintf("https://api.steampowered.com/ISteamDirectory/GetCMList/v1/?cellid=%d", cellID))

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	r := struct {
		Response struct {
			ServerList []string
			Result     uint32
			Message    string
		}
	}{}

	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}

	if r.Response.Result != 1 {
		return fmt.Errorf("Failed to get steam directory, result: %v, message: %v", r.Response.Result, r.Response.Message)
	}

	if len(r.Response.ServerList) == 0 {
		return fmt.Errorf("Steam returned zero servers for steam directory request")
	}

	sd.servers = r.Response.ServerList
	sd.isInitialized = true

	return nil
}

func (sd *steamDirectory) GetRandomCM() (*netutil.PortAddr, error) {
	sd.RLock()
	defer sd.RUnlock()

	if !sd.isInitialized {
		return nil, errors.New("steam directory is not initialized")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	addr := netutil.ParsePortAddr(sd.servers[rng.Int31n(int32(len(sd.servers)))])

	return addr, nil
}

func (sd *steamDirectory) IsInitialized() bool {
	sd.RLock()
	defer sd.RUnlock()
	return sd.isInitialized
}

func SetSteamDirectoryServers(servers []string) {
	steamDirectoryCache.SetServers(servers)
}

// SetServers allows to update the server list.
func (sd *steamDirectory) SetServers(servers []string) {
	sd.Lock()
	defer sd.Unlock()

	sd.servers = servers
}
