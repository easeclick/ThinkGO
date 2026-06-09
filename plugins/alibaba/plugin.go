package alibaba

import (
	"github.com/easeclick/ThinkGO/internal/alibaba"
	"github.com/easeclick/ThinkGO/internal/framework"
	"github.com/easeclick/ThinkGO/plugin"
)

func init() {
	plugin.Register(&AlibabaPlugin{})
}

// AlibabaPlugin wraps the internal 1688 API client as a plugin.
type AlibabaPlugin struct {
	plugin.BasePlugin
	client *alibaba.Client
}

func (p *AlibabaPlugin) ID() string { return "alibaba" }

func (p *AlibabaPlugin) Description() string {
	return "1688 OpenAPI integration — product search, drop shipping"
}

func (p *AlibabaPlugin) Routes() []plugin.RouteInfo {
	return []plugin.RouteInfo{
		{Method: "GET", Path: "/-/alibaba/search", Summary: "Search 1688 products by keyword"},
	}
}

func (p *AlibabaPlugin) Boot(app *thinkgo.App) error {
	cfg := app.Config()
	appKey := cfg.GetString("alibaba.app_key")
	appSecret := cfg.GetString("alibaba.app_secret")

	if appKey == "" {
		p.client = nil // unconfigured — returns mock data
		return nil
	}
	p.client = alibaba.NewClient(appKey, appSecret)
	return nil
}

func (p *AlibabaPlugin) RegisterRoutes(r *thinkgo.Router) {
	r.Get("/-/alibaba/search", func(c *thinkgo.Context) error {
		keyword := c.DefaultQuery("keyword", "螃蟹")
		page := c.DefaultQuery("page", "1")
		pageInt := 1
		if n := c.QueryInt("page"); n > 0 {
			pageInt = n
		}
		_ = page // page unused in mock

		if p.client == nil {
			mock := alibaba.NewClient("", "")
			products, err := mock.SearchProducts(keyword, pageInt)
			if err != nil {
				return c.Error("alibaba: " + err.Error())
			}
			return c.Success("ok", thinkgo.Map{"keyword": keyword, "products": products})
		}
		products, err := p.client.SearchProducts(keyword, pageInt)
		if err != nil {
			return c.Error("alibaba: " + err.Error())
		}
		return c.Success("ok", thinkgo.Map{"keyword": keyword, "products": products})
	})
}
