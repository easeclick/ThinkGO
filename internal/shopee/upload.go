package shopee

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// UploadItem uploads a new item to Shopee.
func (c *Client) UploadItem(item *ItemResponse) (int64, error) {
	params := map[string]string{
		"shop_id": strconv.FormatInt(c.ShopID, 10),
	}

	bodyData := map[string]interface{}{
		"name":         item.Name,
		"description":  item.Description,
		"price":        item.Price,
		"stock":        item.Stock,
		"category_id":  100001,
		"images":       item.ImageURLs,
	}

	bodyBytes, _ := json.Marshal(bodyData)

	resp, err := c.Post("/api/v2/item/add", params, bytes.NewReader(bodyBytes))
	if err != nil {
		return 0, err
	}

	var result struct {
		ItemID int64 `json:"item_id"`
	}
	if err := resp.UnmarshalData(&result); err != nil {
		return 0, err
	}

	return result.ItemID, nil
}
