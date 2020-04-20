package inventory

import (
	"fmt"
	"net/http"
	"strconv"
)

func GetPartialOwnInventory(
	client *http.Client,
	contextID uint64,
	appID uint32,
	start *uint,
) (*PartialInventory, error) {
	// TODO: the "trading" parameter can be left off to return non-tradable items too
	url := fmt.Sprintf("http://steamcommunity.com/my/inventory/json/%d/%d?trading=1", appID, contextID)

	if start != nil {
		url += "&start=" + strconv.FormatUint(uint64(*start), 10)
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	return DoInventoryRequest(client, req)
}

func GetOwnInventory(client *http.Client, contextID uint64, appID uint32) (*Inventory, error) {
	return GetFullInventory(func() (*PartialInventory, error) {
		return GetPartialOwnInventory(client, contextID, appID, nil)
	}, func(start uint) (*PartialInventory, error) {
		return GetPartialOwnInventory(client, contextID, appID, &start)
	})
}
