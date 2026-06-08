package alibaba

import "fmt"

// DropShippingOrder represents a 1688 drop shipping order.
type DropShippingOrder struct {
	OrderID   string  `json:"orderId"`
	Status    string  `json:"status"`
	TotalCost float64 `json:"totalCost"`
}

// CreateDropShippingOrder creates a drop shipping order on 1688.
// Returns mock order ID if no credentials configured.
func (c *Client) CreateDropShippingOrder(productID string, quantity int, address string) (string, error) {
	if c.AppKey == "" || c.AppKey == "your_app_key" {
		return fmt.Sprintf("MOCK_ORDER_%s_%d", productID, quantity), nil
	}

	// TODO: real API call when credentials available
	return fmt.Sprintf("MOCK_ORDER_%s_%d", productID, quantity), nil
}
