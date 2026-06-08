package shopee

import (
	"strconv"
	"time"
)

// Order represents a Shopee order.
type Order struct {
	OrderID   string    `json:"order_sn"`
	Status    string    `json:"order_status"`
	Amount    float64   `json:"total_amount"`
	Sku       string    `json:"item_sku"`
	CreatedAt time.Time `json:"create_time"`
}

type orderListResponse struct {
	Orders []Order `json:"order_list"`
	More   bool    `json:"more"`
}

// GetOrders fetches completed orders within a time range.
func (c *Client) GetOrders(startTime, endTime time.Time) ([]Order, error) {
	params := map[string]string{
		"time_range_field": "create_time",
		"time_from":        strconv.FormatInt(startTime.Unix(), 10),
		"time_to":          strconv.FormatInt(endTime.Unix(), 10),
		"order_status":     "COMPLETED",
		"page_size":        "100",
	}

	resp, err := c.Get("/api/v2/order/get_order_list", params)
	if err != nil {
		return nil, err
	}

	var data orderListResponse
	if err := resp.UnmarshalData(&data); err != nil {
		return nil, err
	}

	return data.Orders, nil
}
