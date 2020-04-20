package tradeoffer

import (
	"encoding/json"
	"errors"
	"regexp"

	"github.com/13k/go-steam/economy/inventory"
)

type TradeReceiptItem struct {
	AssetId   uint64 `json:"id,string"`
	AppId     uint32
	ContextId uint64
	Owner     uint64 `json:",string"`
	Pos       uint32
	inventory.Description
}

var receiptItemRE = regexp.MustCompile(`oItem =\s+(.+?});`)

func parseTradeReceipt(data []byte) ([]*TradeReceiptItem, error) {
	itemMatches := receiptItemRE.FindAllSubmatch(data, -1)

	if itemMatches == nil {
		return nil, errors.New("no items found")
	}

	items := make([]*TradeReceiptItem, 0, len(itemMatches))

	for _, m := range itemMatches {
		item := &TradeReceiptItem{}
		err := json.Unmarshal(m[1], &item)

		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}
