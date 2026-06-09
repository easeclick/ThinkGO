package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/easeclick/ThinkGO/internal/framework"
	"github.com/easeclick/ThinkGO/internal/model"
	"github.com/easeclick/ThinkGO/plugin"
	_ "github.com/easeclick/ThinkGO/plugins/alibaba"
	_ "github.com/easeclick/ThinkGO/plugins/erpcore"
	_ "github.com/easeclick/ThinkGO/plugins/shopee"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
}

func setupPluginRouter(t *testing.T) (*thinkgo.Router, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.ShopOrder{}, &model.AliPurchase{}, &model.Product{}, &model.Profit{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	db.Create(&model.Product{ShopeeItemID: 1001, AliProductID: "ali-001", Title: "Test Product", Price: 25.00, Stock: 10})
	db.Create(&model.ShopOrder{OrderID: "ORD-TEST-1", Amount: 150.00, Status: "COMPLETED", Sku: "ali-001", CreatedAt: today})
	db.Create(&model.AliPurchase{PurchaseID: "PUR-TEST-1", Cost: 60.00, Sku: "ali-001", CreatedAt: today})

	app := thinkgo.NewApp()
	app.SetDB(db)
	thinkgo.DB = db

	router := thinkgo.NewRouter()
	router.Use(thinkgo.Recovery())
	router.Use(thinkgo.LoggerMW())
	router.Use(thinkgo.CORSMiddleware())

	router.Get("/ping", func(ctx *thinkgo.Context) error {
		return ctx.JSON(thinkgo.Map{"message": "pong"})
	})

	pm := plugin.NewManager(app, router)
	for _, p := range plugin.Registered() {
		pm.Register(p)
	}
	if err := pm.Boot(); err != nil {
		t.Fatalf("plugin boot failed: %v", err)
	}

	return router, db
}

func executeRequest(router *thinkgo.Router, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestPing(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "GET", "/ping")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
	}
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "pong" {
		t.Errorf("expected pong, got: %s", resp["message"])
	}
}

func TestListProducts(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "GET", "/api/v1/products")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
	}
	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "ok" {
		t.Errorf("expected msg=ok, got: %s", resp.Message)
	}
}

func TestGetProduct(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "GET", "/api/v1/products/1")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
	}
	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "ok" {
		t.Errorf("expected msg=ok, got: %s", resp.Message)
	}
}

func TestGetOrderByOrderID(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "GET", "/api/v1/orders/ORD-TEST-1")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for order_id string lookup, got: %d", w.Code)
	}
	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "ok" {
		t.Errorf("expected msg=ok for order_id, got: %s", resp.Message)
	}
}

func TestGetOrderByNumericID(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "GET", "/api/v1/orders/1")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for numeric id lookup, got: %d", w.Code)
	}
	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "ok" {
		t.Errorf("expected msg=ok for numeric id, got: %s", resp.Message)
	}
}

func TestListPurchases(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "GET", "/api/v1/purchases")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
	}
	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "ok" {
		t.Errorf("expected msg=ok, got: %s", resp.Message)
	}
}

func TestDailyReport(t *testing.T) {
	router, _ := setupPluginRouter(t)
	date := time.Now().Format("2006-01-02")
	w := executeRequest(router, "GET", "/api/v1/report/daily?date="+date)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
	}
	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "ok" {
		t.Errorf("expected msg=ok, got: %s", resp.Message)
	}
}

func TestMonthlyReport(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "GET", "/api/v1/report/monthly")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
	}
	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "ok" {
		t.Errorf("expected msg=ok, got: %s", resp.Message)
	}
}

func TestSearchProducts(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "GET", "/api/v1/search?keyword=发夹")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
	}
	var resp testResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Message != "ok" {
		t.Errorf("expected msg=ok, got: %s", resp.Message)
	}
}

func TestDiscoveryPlugins(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "GET", "/-/plugins")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
	}
	var plugins []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &plugins); err != nil {
		t.Fatalf("failed to parse plugins list: %v", err)
	}
	if len(plugins) < 3 {
		t.Errorf("expected at least 3 plugins, got: %d", len(plugins))
	}
	ids := make(map[string]bool)
	for _, p := range plugins {
		ids[p["id"].(string)] = true
	}
	for _, id := range []string{"shopee", "alibaba", "erpcore"} {
		if !ids[id] {
			t.Errorf("missing plugin: %s", id)
		}
	}
}

func TestDiscoveryAPISpec(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "GET", "/-/api.json")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
	}
	var spec map[string]any
	json.Unmarshal(w.Body.Bytes(), &spec)
	if spec["framework"] != "ThinkGo" {
		t.Errorf("expected framework ThinkGo, got: %s", spec["framework"])
	}
}

func TestShopeePluginRoute(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "GET", "/-/shopee/item")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
	}
}

func TestAlibabaPluginRoute(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "GET", "/-/alibaba/search")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
	}
}

func TestNotFound(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "GET", "/api/v1/nonexistent")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got: %d", w.Code)
	}
}

func Test405MethodNotAllowed(t *testing.T) {
	router, _ := setupPluginRouter(t)
	w := executeRequest(router, "POST", "/api/v1/products")
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got: %d", w.Code)
	}
}
