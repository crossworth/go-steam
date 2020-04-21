package tradeoffer

import (
	"encoding/json"
	"errors"
	"regexp"

	"github.com/13k/go-steam/economy/inventory"
)

type ReceiptItem struct {
	AssetID   uint64 `json:"id,string"`
	AppID     uint32
	ContextID uint64
	Owner     uint64 `json:",string"`
	Pos       uint32
	inventory.Description
}

var receiptItemRE = regexp.MustCompile(`oItem =\s+(.+?});`)

func parseTradeReceipt(data []byte) ([]*ReceiptItem, error) {
	matches := receiptItemRE.FindAllSubmatch(data, -1)

	if matches == nil {
		return nil, errors.New("no items found")
	}

	items := make([]*ReceiptItem, len(matches))

	for i, m := range matches {
		item := &ReceiptItem{}

		if err := json.Unmarshal(m[1], item); err != nil {
			return nil, err
		}

		items[i] = item
	}

	return items, nil
}
