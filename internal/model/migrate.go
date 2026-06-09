package model

import (
	"time"

	"gorm.io/gorm"
)

// ShopOrder represents a Shopee order
type ShopOrder struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   string    `gorm:"uniqueIndex;size:64;not null" json:"order_id"`
	Amount    float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status    string    `gorm:"size:20" json:"status"`
	Sku       string    `gorm:"size:100" json:"sku"`
	CreatedAt time.Time `json:"created_at"`
}

func (ShopOrder) TableName() string { return "shop_orders" }

// AliPurchase represents a 1688 purchase order
type AliPurchase struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PurchaseID string    `gorm:"uniqueIndex;size:64;not null" json:"purchase_id"`
	Cost       float64   `gorm:"type:decimal(10,2);not null" json:"cost"`
	Sku        string    `gorm:"size:100" json:"sku"`
	OrderRef   string    `gorm:"size:64" json:"order_ref"`
	CreatedAt  time.Time `json:"created_at"`
}

func (AliPurchase) TableName() string { return "ali_purchases" }

// Product represents a cross-listed product
type Product struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	ShopeeItemID int64   `gorm:"index" json:"shopee_item_id"`
	AliProductID string  `gorm:"size:64" json:"ali_product_id"`
	Title        string  `gorm:"type:text" json:"title"`
	Price        float64 `gorm:"type:decimal(10,2)" json:"price"`
	Stock        int     `gorm:"default:0" json:"stock"`
}

func (Product) TableName() string { return "products" }

// Profit represents a daily profit record
type Profit struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Date        time.Time `gorm:"type:date;not null;uniqueIndex" json:"date"`
	TotalProfit float64   `gorm:"type:decimal(10,2)" json:"total_profit"`
	TotalSales  float64   `gorm:"type:decimal(10,2);default:0" json:"total_sales"`
	TotalCost   float64   `gorm:"type:decimal(10,2);default:0" json:"total_cost"`
	OrderCount  int       `gorm:"default:0" json:"order_count"`
	BestSku     string    `gorm:"size:100" json:"best_sku"`
	WorstSku    string    `gorm:"size:100" json:"worst_sku"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Profit) TableName() string { return "profits" }

// MigrateDB runs auto-migration for all models
func MigrateDB(db *gorm.DB) error {
	return db.AutoMigrate(&ShopOrder{}, &AliPurchase{}, &Product{}, &Profit{})
}
