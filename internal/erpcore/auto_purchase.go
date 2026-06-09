package erpcore

import (
	"log/slog"
	"math/rand"
	"time"

	"github.com/easeclick/ThinkGO/internal/model"
	"gorm.io/gorm"
)

// CheckLowStock checks for products with stock < 10 and triggers auto-purchase.
func CheckLowStock(db *gorm.DB) {
	slog.Info("checking low stock products...")

	var products []model.Product

	if err := db.Where("stock < ?", 10).Find(&products).Error; err != nil {
		slog.Error("low stock query failed", "error", err)
		return
	}

	if len(products) == 0 {
		slog.Info("no low stock products found")
		return
	}

	for _, p := range products {
		slog.Info("low stock product detected",
			"product_id", p.ID,
			"title", p.Title,
			"stock", p.Stock,
			"ali_product_id", p.AliProductID,
		)

		purchase := model.AliPurchase{
			PurchaseID: "ERP-" + time.Now().Format("20060102") + "-" + randString(8),
			Cost:       float64(rand.Intn(2000)+200) / 100,
			Sku:        p.AliProductID,
			OrderRef:   "",
			CreatedAt:  time.Now(),
		}
		if err := db.Create(&purchase).Error; err != nil {
			slog.Error("failed to create auto purchase", "error", err)
			continue
		}

		db.Model(&p).Update("stock", 100)
		slog.Info("auto purchase completed", "purchase_id", purchase.PurchaseID, "sku", p.AliProductID)
	}
}

func randString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
