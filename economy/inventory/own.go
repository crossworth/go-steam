package inventory

import (
	"fmt"
	"net/http"
)

func GetOwnPartialInventory(
	client *http.Client,
	contextID uint64,
	appID uint32,
	start uint,
) (*PartialInventory, error) {
	// TODO: the "trading" parameter can be left off to return non-tradable items too
	url := fmt.Sprintf("http://steamcommunity.com/my/inventory/json/%d/%d?trading=1", appID, contextID)

	if start != 0 {
		url += fmt.Sprintf("&start=%d", start)
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	return PerformRequest(client, req)
}

func GetOwnInventory(client *http.Client, contextID uint64, appID uint32) (*Inventory, error) {
	return GetFullInventory(func(start uint) (*PartialInventory, error) {
		return GetOwnPartialInventory(client, contextID, appID, start)
	})
}
