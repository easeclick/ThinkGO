package shopee

import (
	"strconv"

	"github.com/easeclick/ThinkGO/internal/framework"
	"github.com/easeclick/ThinkGO/internal/shopee"
	"github.com/easeclick/ThinkGO/plugin"
)

func init() {
	plugin.Register(&ShopeePlugin{})
}

// ShopeePlugin wraps the internal Shopee API client as a plugin.
type ShopeePlugin struct {
	plugin.BasePlugin
	client *shopee.Client
}

func (p *ShopeePlugin) ID() string { return "shopee" }

func (p *ShopeePlugin) Description() string {
	return "Shopee OpenAPI v2 integration — items, orders, upload"
}

func (p *ShopeePlugin) Routes() []plugin.RouteInfo {
	return []plugin.RouteInfo{
		{Method: "GET", Path: "/-/shopee/item", Summary: "Get item by ID from Shopee"},
		{Method: "GET", Path: "/-/shopee/orders", Summary: "List recent Shopee orders"},
	}
}

func (p *ShopeePlugin) Boot(app *thinkgo.App) error {
	cfg := app.Config()
	partnerID := int64(cfg.GetInt("shopee.partner_id"))
	partnerKey := cfg.GetString("shopee.partner_key")
	shopID := int64(cfg.GetInt("shopee.shop_id"))

	if partnerID == 0 || partnerKey == "" {
		p.client = nil // unconfigured — routes return mock data
		return nil
	}
	p.client = shopee.NewClient(partnerID, partnerKey, shopID)
	return nil
}

func (p *ShopeePlugin) RegisterRoutes(r *thinkgo.Router) {
	r.Get("/-/shopee/item", func(c *thinkgo.Context) error {
		itemIDStr := c.DefaultQuery("item_id", "1001")
		itemID, _ := strconv.ParseInt(itemIDStr, 10, 64)
		if p.client == nil {
			return c.Success("mock", thinkgo.Map{"item_id": itemID, "name": "Mock Shopee Item", "price": 29.90})
		}
		item, err := p.client.GetItem(itemID, 0)
		if err != nil {
			return c.Error("shopee: " + err.Error())
		}
		return c.Success("ok", item)
	})

	r.Get("/-/shopee/orders", func(c *thinkgo.Context) error {
		if p.client == nil {
			return c.Success("mock", thinkgo.Map{"message": "Shopee not configured, configure shopee.partner_id in config"})
		}
		// Would call p.client.GetOrders(...) in production
		return c.Success("ok", thinkgo.Map{"message": "realtime sync not implemented"})
	})
}
