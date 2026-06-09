package model

import (
	"log/slog"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

// SeedData populates the database with mock data for development/testing.
func SeedData(db *gorm.DB) error {
	slog.Info("seeding database with mock data...")

	if err := seedProducts(db); err != nil {
		return err
	}
	if err := seedOrders(db); err != nil {
		return err
	}
	if err := seedPurchases(db); err != nil {
		return err
	}
	if err := seedProfits(db); err != nil {
		return err
	}

	slog.Info("seed data inserted successfully")
	return nil
}

func seedProducts(db *gorm.DB) error {
	products := []Product{
		{ShopeeItemID: 1001, AliProductID: "ali-mock001", Title: "韩版简约透明发夹", Price: 12.80, Stock: 50},
		{ShopeeItemID: 1002, AliProductID: "ali-mock002", Title: "LED发光螃蟹发夹派对装饰", Price: 19.90, Stock: 5},
		{ShopeeItemID: 1003, AliProductID: "ali-mock003", Title: "水晶螃蟹发夹女生首饰", Price: 28.00, Stock: 120},
		{ShopeeItemID: 1004, AliProductID: "ali-mock004", Title: "简约金属发夹套装", Price: 15.50, Stock: 3},
		{ShopeeItemID: 1005, AliProductID: "ali-mock005", Title: "蝴蝶结珍珠发夹", Price: 22.00, Stock: 8},
	}

	for _, p := range products {
		var count int64
		db.Model(&Product{}).Where("shopee_item_id = ?", p.ShopeeItemID).Count(&count)
		if count == 0 {
			if err := db.Create(&p).Error; err != nil {
				return err
			}
		}
	}

	slog.Info("seeded products", "count", len(products))
	return nil
}

func seedOrders(db *gorm.DB) error {
	now := time.Now()
	skus := []string{"ali-mock001", "ali-mock002", "ali-mock003", "ali-mock004", "ali-mock005"}
	statuses := []string{"COMPLETED", "COMPLETED", "COMPLETED", "PENDING", "SHIPPED"}

	// Seed orders for the last 30 days
	var count int64
	db.Model(&ShopOrder{}).Count(&count)
	if count > 0 {
		slog.Info("orders already seeded, skipping", "count", count)
		return nil
	}

	orders := make([]ShopOrder, 0, 50)
	for i := 0; i < 50; i++ {
		daysAgo := rand.Intn(30)
		hoursAgo := rand.Intn(24)
		createdAt := now.Add(-time.Duration(daysAgo)*24*time.Hour - time.Duration(hoursAgo)*time.Hour)
		sku := skus[rand.Intn(len(skus))]
		status := statuses[rand.Intn(len(statuses))]

		orders = append(orders, ShopOrder{
			OrderID:   orderID(i),
			Amount:    float64(rand.Intn(5000)+500) / 100,
			Status:    status,
			Sku:       sku,
			CreatedAt: createdAt,
		})
	}

	if err := db.Create(&orders).Error; err != nil {
		return err
	}

	slog.Info("seeded orders", "count", len(orders))
	return nil
}

func seedPurchases(db *gorm.DB) error {
	var count int64
	db.Model(&AliPurchase{}).Count(&count)
	if count > 0 {
		slog.Info("purchases already seeded, skipping", "count", count)
		return nil
	}

	now := time.Now()
	skus := []string{"ali-mock001", "ali-mock002", "ali-mock003"}

	purchases := make([]AliPurchase, 0, 20)
	for i := 0; i < 20; i++ {
		daysAgo := rand.Intn(30)
		createdAt := now.Add(-time.Duration(daysAgo) * 24 * time.Hour)
		sku := skus[rand.Intn(len(skus))]

		purchases = append(purchases, AliPurchase{
			PurchaseID: "PUR-" + time.Now().Format("20060102") + "-" + randString(8),
			Cost:       float64(rand.Intn(2000)+200) / 100,
			Sku:        sku,
			OrderRef:   orderID(rand.Intn(50)),
			CreatedAt:  createdAt,
		})
	}

	if err := db.Create(&purchases).Error; err != nil {
		return err
	}

	slog.Info("seeded purchases", "count", len(purchases))
	return nil
}

func seedProfits(db *gorm.DB) error {
	var count int64
	db.Model(&Profit{}).Count(&count)
	if count > 0 {
		slog.Info("profits already seeded, skipping", "count", count)
		return nil
	}

	profits := make([]Profit, 0, 14)
	for i := 0; i < 14; i++ {
		date := time.Now().Add(-time.Duration(i) * 24 * time.Hour)
		// Use date only (truncate time)
		date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

		sales := float64(rand.Intn(10000)+1000) / 100
		cost := sales * (0.3 + rand.Float64()*0.2) // 30-50% cost ratio
		profit := sales - cost

		profits = append(profits, Profit{
			Date:        date,
			TotalProfit: profit,
			TotalSales:  sales,
			TotalCost:   cost,
			OrderCount:  rand.Intn(30) + 5,
			BestSku:     "ali-mock001",
			WorstSku:    "ali-mock004",
			CreatedAt:   time.Now(),
		})
	}

	if err := db.Create(&profits).Error; err != nil {
		return err
	}

	slog.Info("seeded profits", "count", len(profits))
	return nil
}

var orderCounter int

func orderID(n int) string {
	orderCounter++
	return "MOCK-ORDER-" + time.Now().Format("20060102") + "-" + randString(6)
}

func randString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
