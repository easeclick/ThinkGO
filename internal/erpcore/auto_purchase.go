package erpcore

import (
	"log/slog"

	"gorm.io/gorm"
)

// CheckLowStock checks for products with stock < 10 and triggers auto-purchase.
func CheckLowStock(db *gorm.DB) {
	slog.Info("checking low stock products...")

	var products []struct {
		ID           uint
		AliProductID string
		Title        string
		Stock        int
	}

	if err := db.Table("products").Where("stock < ?", 10).Find(&products).Error; err != nil {
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

		// TODO: trigger actual 1688 purchase when credentials configured
		// orderID, err := alibabaClient.CreateDropShippingOrder(p.AliProductID, 50, "warehouse")
		// if err != nil { slog.Error("auto purchase failed", "error", err); continue }
		// slog.Info("auto purchase created", "order_id", orderID, "sku", p.AliProductID)
	}
}
