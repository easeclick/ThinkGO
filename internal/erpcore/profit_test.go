package erpcore

import (
	"testing"
	"time"

	"github.com/easeclick/ThinkGO/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.ShopOrder{}, &model.AliPurchase{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func seedTestData(t *testing.T, db *gorm.DB) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	orders := []model.ShopOrder{
		{OrderID: "T1", Amount: 100.00, Status: "COMPLETED", Sku: "SKU-A", CreatedAt: today.Add(2 * time.Hour)},
		{OrderID: "T2", Amount: 200.00, Status: "COMPLETED", Sku: "SKU-B", CreatedAt: today.Add(4 * time.Hour)},
		{OrderID: "T3", Amount: 50.00, Status: "COMPLETED", Sku: "SKU-A", CreatedAt: today.Add(6 * time.Hour)},
	}
	purchases := []model.AliPurchase{
		{PurchaseID: "P1", Cost: 60.00, Sku: "SKU-A", CreatedAt: today.Add(1 * time.Hour)},
		{PurchaseID: "P2", Cost: 80.00, Sku: "SKU-B", CreatedAt: today.Add(3 * time.Hour)},
	}

	if err := db.Create(&orders).Error; err != nil {
		t.Fatalf("failed to seed orders: %v", err)
	}
	if err := db.Create(&purchases).Error; err != nil {
		t.Fatalf("failed to seed purchases: %v", err)
	}
}

func TestCalculateDailyProfit(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	profit, err := CalculateDailyProfit(db, date)
	if err != nil {
		t.Fatalf("CalculateDailyProfit failed: %v", err)
	}

	if profit.TotalSales != 350.00 {
		t.Errorf("expected total sales 350.00, got: %.2f", profit.TotalSales)
	}
	if profit.TotalCost != 140.00 {
		t.Errorf("expected total cost 140.00, got: %.2f", profit.TotalCost)
	}
	if profit.TotalProfit != 210.00 {
		t.Errorf("expected total profit 210.00, got: %.2f", profit.TotalProfit)
	}
	if profit.OrderCount != 3 {
		t.Errorf("expected 3 orders, got: %d", profit.OrderCount)
	}
	if profit.BestSku != "SKU-B" {
		t.Errorf("expected best SKU SKU-B (200.00), got: %s", profit.BestSku)
	}
}

func TestCalculateDailyProfitNoOrders(t *testing.T) {
	db := setupTestDB(t)

	date := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	profit, err := CalculateDailyProfit(db, date)
	if err != nil {
		t.Fatalf("CalculateDailyProfit failed: %v", err)
	}

	if profit.TotalSales != 0 || profit.TotalCost != 0 || profit.TotalProfit != 0 {
		t.Errorf("expected zero profit for empty day, got: sales=%.2f cost=%.2f profit=%.2f",
			profit.TotalSales, profit.TotalCost, profit.TotalProfit)
	}
	if profit.OrderCount != 0 {
		t.Errorf("expected 0 orders, got: %d", profit.OrderCount)
	}
}

func TestSaveDailyProfit(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.Profit{}); err != nil {
		t.Fatalf("failed to migrate profits: %v", err)
	}

	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	profit := &DailyProfit{
		Date:        date,
		TotalProfit: 500.00,
		TotalSales:  1000.00,
		TotalCost:   500.00,
		OrderCount:  10,
		BestSku:     "BEST",
		WorstSku:    "WORST",
	}

	if err := SaveDailyProfit(db, profit); err != nil {
		t.Fatalf("SaveDailyProfit failed: %v", err)
	}

	var saved model.Profit
	if err := db.First(&saved).Error; err != nil {
		t.Fatalf("failed to read saved profit: %v", err)
	}

	if saved.TotalProfit != 500.00 {
		t.Errorf("expected TotalProfit 500.00, got: %.2f", saved.TotalProfit)
	}
	if saved.TotalSales != 1000.00 {
		t.Errorf("expected TotalSales 1000.00, got: %.2f", saved.TotalSales)
	}
	if saved.TotalCost != 500.00 {
		t.Errorf("expected TotalCost 500.00, got: %.2f", saved.TotalCost)
	}
}
