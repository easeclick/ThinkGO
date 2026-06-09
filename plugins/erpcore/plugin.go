package erpcore

import (
	"strconv"
	"time"

	"github.com/easeclick/ThinkGO/internal/alibaba"
	"github.com/easeclick/ThinkGO/internal/erpcore"
	"github.com/easeclick/ThinkGO/internal/framework"
	"github.com/easeclick/ThinkGO/internal/model"
	"github.com/easeclick/ThinkGO/plugin"
	"gorm.io/gorm"
)

func init() {
	plugin.Register(&ERPCorePlugin{})
}

// ERPCorePlugin provides the core ERP business logic and API routes.
type ERPCorePlugin struct {
	plugin.BasePlugin
	db *gorm.DB
}

func (p *ERPCorePlugin) ID() string { return "erpcore" }

func (p *ERPCorePlugin) Description() string {
	return "ERP core — product listing, order management, profit reports, 1688 product search"
}

func (p *ERPCorePlugin) Routes() []plugin.RouteInfo {
	return []plugin.RouteInfo{
		{Method: "GET", Path: "/api/v1/products", Summary: "List all products"},
		{Method: "GET", Path: "/api/v1/products/:id", Summary: "Get product by ID"},
		{Method: "GET", Path: "/api/v1/orders", Summary: "List shop orders"},
		{Method: "GET", Path: "/api/v1/orders/:id", Summary: "Get order by ID or order_id"},
		{Method: "GET", Path: "/api/v1/purchases", Summary: "List 1688 purchase orders"},
		{Method: "GET", Path: "/api/v1/search", Summary: "Search 1688 products (mock)"},
		{Method: "GET", Path: "/api/v1/report/daily", Summary: "Daily profit report"},
		{Method: "GET", Path: "/api/v1/report/monthly", Summary: "Monthly profit report"},
	}
}

func (p *ERPCorePlugin) Boot(app *thinkgo.App) error {
	if db, ok := app.GetDB().(*gorm.DB); ok && db != nil {
		p.db = db
		return nil
	}
	if thinkgo.DB != nil {
		p.db = thinkgo.DB
	}
	return nil
}

func (p *ERPCorePlugin) RegisterRoutes(r *thinkgo.Router) {
	v1 := r.Group("/api/v1")

	v1.Get("/products", p.handleListProducts)
	v1.Get("/products/:id", p.handleGetProduct)
	v1.Get("/orders", p.handleListOrders)
	v1.Get("/orders/:id", p.handleGetOrder)
	v1.Get("/purchases", p.handleListPurchases)
	v1.Get("/search", p.handleSearch)
	v1.Get("/report/daily", p.handleDailyReport)
	v1.Get("/report/monthly", p.handleMonthlyReport)
}

// --- Handlers ---

func (p *ERPCorePlugin) handleListProducts(ctx *thinkgo.Context) error {
	var products []model.Product
	if err := p.db.Find(&products).Error; err != nil {
		return ctx.Error("query products failed")
	}
	return ctx.Success("ok", products)
}

func (p *ERPCorePlugin) handleGetProduct(ctx *thinkgo.Context) error {
	id := ctx.Param("id")
	var product model.Product
	if err := p.db.First(&product, id).Error; err != nil {
		return ctx.Error("product not found")
	}
	return ctx.Success("ok", product)
}

func (p *ERPCorePlugin) handleListOrders(ctx *thinkgo.Context) error {
	var orders []model.ShopOrder
	limit := 50
	if l := ctx.QueryInt("limit"); l > 0 && l <= 200 {
		limit = l
	}
	if err := p.db.Order("created_at DESC").Limit(limit).Find(&orders).Error; err != nil {
		return ctx.Error("query orders failed")
	}
	return ctx.Success("ok", orders)
}

func (p *ERPCorePlugin) handleGetOrder(ctx *thinkgo.Context) error {
	id := ctx.Param("id")
	var order model.ShopOrder

	if nid, err := strconv.ParseUint(id, 10, 64); err == nil {
		if err := p.db.First(&order, nid).Error; err == nil {
			return ctx.Success("ok", order)
		}
	}

	if err := p.db.Where("order_id = ?", id).First(&order).Error; err != nil {
		return ctx.Error("order not found")
	}
	return ctx.Success("ok", order)
}

func (p *ERPCorePlugin) handleListPurchases(ctx *thinkgo.Context) error {
	var purchases []model.AliPurchase
	limit := 50
	if l := ctx.QueryInt("limit"); l > 0 && l <= 200 {
		limit = l
	}
	if err := p.db.Order("created_at DESC").Limit(limit).Find(&purchases).Error; err != nil {
		return ctx.Error("query purchases failed")
	}
	return ctx.Success("ok", purchases)
}

func (p *ERPCorePlugin) handleSearch(ctx *thinkgo.Context) error {
	keyword := ctx.DefaultQuery("keyword", "螃蟹")
	page := ctx.QueryInt("page")
	if page < 1 {
		page = 1
	}

	mockClient := alibaba.NewClient("", "")
	products, err := mockClient.SearchProducts(keyword, page)
	if err != nil {
		return ctx.Error("search failed")
	}
	return ctx.Success("ok", thinkgo.Map{
		"keyword":  keyword,
		"page":     page,
		"products": products,
	})
}

func (p *ERPCorePlugin) handleDailyReport(ctx *thinkgo.Context) error {
	dateStr := ctx.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return ctx.Error("invalid date format, use YYYY-MM-DD")
	}

	profit, err := erpcore.CalculateDailyProfit(p.db, date)
	if err != nil {
		return ctx.Error("profit calculation failed: " + err.Error())
	}
	return ctx.Success("ok", profit)
}

func (p *ERPCorePlugin) handleMonthlyReport(ctx *thinkgo.Context) error {
	year := ctx.QueryInt("year")
	if year <= 0 {
		year = time.Now().Year()
	}
	month := ctx.QueryInt("month")
	if month < 1 || month > 12 {
		month = int(time.Now().Month())
	}

	location := time.UTC
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, location)
	endDate := startDate.AddDate(0, 1, 0)

	var dailyProfits []*erpcore.DailyProfit
	totalProfit := 0.0
	totalSales := 0.0
	totalCost := 0.0

	current := startDate
	for current.Before(endDate) {
		profit, err := erpcore.CalculateDailyProfit(p.db, current)
		if err != nil {
			return ctx.Error("profit calc failed: " + err.Error())
		}
		dailyProfits = append(dailyProfits, profit)
		totalProfit += profit.TotalProfit
		totalSales += profit.TotalSales
		totalCost += profit.TotalCost
		current = current.AddDate(0, 0, 1)
	}

	type SkuRow struct {
		Sku    string
		Amount float64
	}
	var rows []SkuRow
	p.db.Table("shop_orders").
		Select("sku, SUM(amount) as amount").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("sku").
		Find(&rows)

	var bestSku, worstSku string
	var bestSales float64 = -1
	var worstSales float64 = -1
	for _, r := range rows {
		if bestSales < 0 || r.Amount > bestSales {
			bestSales = r.Amount
			bestSku = r.Sku
		}
		if worstSales < 0 || r.Amount < worstSales {
			worstSales = r.Amount
			worstSku = r.Sku
		}
	}

	return ctx.Success("ok", thinkgo.Map{
		"year":             year,
		"month":            month,
		"total_profit":     totalProfit,
		"total_sales":      totalSales,
		"total_cost":       totalCost,
		"total_orders":     len(rows),
		"best_seller_sku":  bestSku,
		"worst_seller_sku": worstSku,
		"daily_breakdown":  dailyProfits,
	})
}
