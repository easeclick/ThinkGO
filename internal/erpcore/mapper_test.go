package erpcore

import (
	"strings"
	"testing"

	"github.com/easeclick/ThinkGO/internal/alibaba"
)

func TestMapAlibabaToShopeeItem(t *testing.T) {
	aliProd := &alibaba.Product{
		ProductID: "test001",
		Title:     "韩版简约透明螃蟹发夹 2024新款 批发1688",
		Price:     10.00,
		ImageURL:  "https://example.com/img.jpg",
		DetailURL: "https://detail.1688.com/offer/test001.html",
	}

	item := MapAlibabaToShopeeItem(aliProd)

	if item.Title != "韩版简约透明螃蟹发夹 2024新款" {
		t.Errorf("expected cleaned title, got: %s", item.Title)
	}
	if strings.Contains(item.Title, "1688") || strings.Contains(item.Title, "批发") {
		t.Errorf("title should not contain keywords: %s", item.Title)
	}
	if item.Price != 25.00 {
		t.Errorf("expected price 25.00 (2.5x markup), got: %.2f", item.Price)
	}
	if item.Stock != 100 {
		t.Errorf("expected default stock 100, got: %d", item.Stock)
	}
	if item.OriginalSKU != "test001" {
		t.Errorf("expected original SKU test001, got: %s", item.OriginalSKU)
	}
	if len(item.ImageURLs) != 1 || item.ImageURLs[0] != "https://example.com/img.jpg" {
		t.Errorf("image URLs not mapped correctly: %v", item.ImageURLs)
	}
	if item.CategoryID != 100001 {
		t.Errorf("expected default category 100001, got: %d", item.CategoryID)
	}
}

func TestMapAlibabaToShopeeItemLongTitle(t *testing.T) {
	longTitle := strings.Repeat("商品", 100)
	aliProd := &alibaba.Product{
		ProductID: "test002",
		Title:     longTitle,
		Price:     5.00,
	}

	item := MapAlibabaToShopeeItem(aliProd)

	if len([]rune(item.Title)) > 120 {
		t.Errorf("title exceeds 120 chars: %d", len([]rune(item.Title)))
	}
	if !strings.HasSuffix(item.Title, "...") {
		t.Errorf("long title should end with ...: %s", item.Title)
	}
}

func TestRandomPrice(t *testing.T) {
	min, max := 10.0, 20.0
	for i := 0; i < 100; i++ {
		price := RandomPrice(min, max)
		if price < min || price > max {
			t.Errorf("price %.2f out of range [%.2f, %.2f]", price, min, max)
		}
	}
}
