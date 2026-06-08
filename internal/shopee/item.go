package shopee

import "strconv"

// ItemResponse represents a Shopee item's public details.
type ItemResponse struct {
	ItemID      int64    `json:"item_id"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	Description string   `json:"description"`
	ImageURLs   []string `json:"images"`
	ShopID      int64    `json:"shop_id"`
	Status      string   `json:"status"`
}

type itemGetResponse struct {
	Item ItemResponse `json:"item"`
}

// GetItem fetches item details from Shopee.
func (c *Client) GetItem(itemID, shopID int64) (*ItemResponse, error) {
	params := map[string]string{
		"item_id": strconv.FormatInt(itemID, 10),
		"shop_id": strconv.FormatInt(shopID, 10),
	}

	resp, err := c.Get("/api/v2/item/get", params)
	if err != nil {
		return nil, err
	}

	var data itemGetResponse
	if err := resp.UnmarshalData(&data); err != nil {
		return nil, err
	}

	return &data.Item, nil
}
