package inventory

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/13k/go-steam/steamid"
)

type Apps map[string]*App

func (i *Apps) Get(appID uint32) (*App, error) {
	iMap := (map[string]*App)(*i)

	if inventoryApp, ok := iMap[strconv.FormatUint(uint64(appID), 10)]; ok {
		return inventoryApp, nil
	}

	return nil, fmt.Errorf("inventory app not found")
}

func (i *Apps) ToMap() map[string]*App {
	return (map[string]*App)(*i)
}

type App struct {
	AppID            uint32
	Name             string
	Icon             string
	Link             string
	AssetCount       uint32   `json:"asset_count"`
	InventoryLogo    string   `json:"inventory_logo"`
	TradePermissions string   `json:"trade_permissions"`
	Contexts         Contexts `json:"rgContexts"`
}

type Contexts map[string]*Context

func (c *Contexts) Get(contextID uint64) (*Context, error) {
	cMap := (map[string]*Context)(*c)

	if context, ok := cMap[strconv.FormatUint(contextID, 10)]; ok {
		return context, nil
	}

	return nil, fmt.Errorf("context not found")
}

func (c *Contexts) ToMap() map[string]*Context {
	return (map[string]*Context)(*c)
}

type Context struct {
	ContextID  uint64 `json:"id,string"`
	AssetCount uint32 `json:"asset_count"`
	Name       string
}

func GetApps(client *http.Client, steamID steamid.SteamID) (Apps, error) {
	resp, err := http.Get("http://steamcommunity.com/profiles/" + steamID.FormatString() + "/inventory/")

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	// TODO: investigate a better heuristic than this
	reg := regexp.MustCompile("var g_rgAppContextData = (.*?);")
	matches := reg.FindSubmatch(respBody)

	if matches == nil {
		return nil, fmt.Errorf("profile inventory not found in steam response")
	}

	var apps Apps

	if err = json.Unmarshal(matches[1], &apps); err != nil {
		return nil, err
	}

	return apps, nil
}
