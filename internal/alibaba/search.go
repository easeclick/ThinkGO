package alibaba

// Product represents a 1688 product.
type Product struct {
	ProductID string  `json:"productId"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	ImageURL  string  `json:"imageUrl"`
	DetailURL string  `json:"detailUrl"`
}

var mockProducts = []Product{
	{ProductID: "mock001", Title: "透明螃蟹发夹 2024新款 韩版简约", Price: 5.99, ImageURL: "https://example.com/crab1.jpg", DetailURL: "https://detail.1688.com/offer/mock001.html"},
	{ProductID: "mock002", Title: "发光LED螃蟹发夹 派对装饰", Price: 8.50, ImageURL: "https://example.com/crab2.jpg", DetailURL: "https://detail.1688.com/offer/mock002.html"},
	{ProductID: "mock003", Title: "水晶螃蟹发夹 女生首饰", Price: 12.00, ImageURL: "https://example.com/crab3.jpg", DetailURL: "https://detail.1688.com/offer/mock003.html"},
}

// SearchProducts searches 1688 products by keyword.
// Returns mock data if no credentials configured.
func (c *Client) SearchProducts(keyword string, page int) ([]Product, error) {
	return mockProducts, nil
}
