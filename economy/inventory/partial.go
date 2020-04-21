package inventory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// PartialInventory is a partial inventory as sent by the Steam API.
type PartialInventory struct {
	Inventory

	Success   bool
	Error     string
	More      bool
	MoreStart MoreStart `json:"more_start"`
}

type MoreStart uint

func (m *MoreStart) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("false")) {
		return nil
	}

	return json.Unmarshal(data, (*uint)(m))
}

func PerformRequest(client *http.Client, req *http.Request) (*PartialInventory, error) {
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	inv := &PartialInventory{}

	if err = json.NewDecoder(resp.Body).Decode(inv); err != nil {
		return nil, err
	}

	return inv, nil
}

type PartialInventoryFetcher func(start uint) (*PartialInventory, error)

func GetFullInventory(fetch PartialInventoryFetcher) (*Inventory, error) {
	result := NewInventory()
	start := uint(0)

	for {
		partial, err := fetch(start)

		if err != nil {
			return nil, err
		}

		if !partial.Success {
			return nil, fmt.Errorf("GetFullInventory API call failed: %s", partial.Error)
		}

		result = Merge(result, &partial.Inventory)
		start = uint(partial.MoreStart)

		if !partial.More {
			break
		}
	}

	return result, nil
}

// Merge merges the given srcs into dst.
func Merge(dst *Inventory, srcs ...*Inventory) *Inventory {
	for _, i := range srcs {
		for key, value := range i.Items {
			dst.Items[key] = value
		}

		for key, value := range i.Descriptions {
			dst.Descriptions[key] = value
		}

		for key, value := range i.Currencies {
			dst.Currencies[key] = value
		}
	}

	return dst
}
