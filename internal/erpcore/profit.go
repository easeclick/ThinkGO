package erpcore

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DailyProfit holds profit calculation result for a single day.
type DailyProfit struct {
	Date        time.Time `json:"date"`
	TotalProfit float64   `json:"total_profit"`
	TotalSales  float64   `json:"total_sales"`
	TotalCost   float64   `json:"total_cost"`
	OrderCount  int       `json:"order_count"`
	BestSku     string    `json:"best_sku"`
	WorstSku    string    `json:"worst_sku"`
}

// CalculateDailyProfit computes profit for a given day.
// Logic: reads shop_orders.amount (sales) and ali_purchases.cost (expenses) for that day,
// then profit = total_sales - total_cost.
func CalculateDailyProfit(db *gorm.DB, date time.Time) (*DailyProfit, error) {
	type orderRow struct {
		Amount float64
		Sku    string
	}
	type purchaseRow struct {
		Cost float64
		Sku  string
	}

	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var orders []orderRow
	if err := db.Table("shop_orders").
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}

	var purchases []purchaseRow
	if err := db.Table("ali_purchases").
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Find(&purchases).Error; err != nil {
		return nil, fmt.Errorf("query purchases: %w", err)
	}

	var totalSales, totalCost float64
	skuSales := make(map[string]float64)

	for _, o := range orders {
		totalSales += o.Amount
		skuSales[o.Sku] += o.Amount
	}

	for _, p := range purchases {
		totalCost += p.Cost
	}

	var bestSku, worstSku string
	var bestSales float64 = -1
	var worstSales float64 = -1
	for sku, sales := range skuSales {
		if bestSales < 0 || sales > bestSales {
			bestSales = sales
			bestSku = sku
		}
		if worstSales < 0 || sales < worstSales {
			worstSales = sales
			worstSku = sku
		}
	}

	return &DailyProfit{
		Date:        startOfDay,
		TotalProfit: totalSales - totalCost,
		TotalSales:  totalSales,
		TotalCost:   totalCost,
		OrderCount:  len(orders),
		BestSku:     bestSku,
		WorstSku:    worstSku,
	}, nil
}

// SaveDailyProfit persists calculated profit to the profits table.
func SaveDailyProfit(db *gorm.DB, profit *DailyProfit) error {
	type profitRecord struct {
		Date        time.Time `gorm:"type:date"`
		TotalProfit float64   `gorm:"type:decimal(10,2)"`
		TotalSales  float64   `gorm:"type:decimal(10,2)"`
		TotalCost   float64   `gorm:"type:decimal(10,2)"`
		OrderCount  int
		BestSku     string `gorm:"size:100"`
		WorstSku    string `gorm:"size:100"`
	}

	record := profitRecord{
		Date:        profit.Date,
		TotalProfit: profit.TotalProfit,
		TotalSales:  profit.TotalSales,
		TotalCost:   profit.TotalCost,
		OrderCount:  profit.OrderCount,
		BestSku:     profit.BestSku,
		WorstSku:    profit.WorstSku,
	}

	// Use Clauses to handle upsert on duplicate date (unique index)
	return db.Table("profits").Where("date = ?", record.Date).Assign(record).FirstOrCreate(&record).Error
}
