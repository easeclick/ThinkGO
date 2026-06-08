package erpcore

import (
	"math/rand"
	"strings"

	"github.com/user/thinkgo/internal/alibaba"
)

// MappedItem represents a Shopee-ready item converted from a 1688 product.
type MappedItem struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	ImageURLs   []string `json:"images"`
	CategoryID  int64    `json:"category_id"`
	OriginalSKU string   `json:"original_sku"`
}

// MapAlibabaToShopeeItem converts an Alibaba product to a Shopee-ready item.
// Applies title cleanup, markup calculation, and field mapping.
func MapAlibabaToShopeeItem(aliProd *alibaba.Product) *MappedItem {
	title := aliProd.Title
	if len([]rune(title)) > 120 {
		title = string([]rune(title)[:117]) + "..."
	}

	title = strings.ReplaceAll(title, "1688", "")
	title = strings.ReplaceAll(title, "批发", "")
	title = strings.TrimSpace(title)

	price := aliProd.Price * 2.5

	return &MappedItem{
		Title:       title,
		Description: "商品来源于1688\n\n" + title + "\n\n高品质产品，欢迎下单",
		Price:       price,
		Stock:       100,
		ImageURLs:   []string{aliProd.ImageURL},
		CategoryID:  100001,
		OriginalSKU: aliProd.ProductID,
	}
}

// RandomPrice generates a random price between min and max (for testing).
func RandomPrice(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
