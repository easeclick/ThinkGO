package worker

import (
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	gormlogger "gorm.io/gorm/logger"

	"github.com/easeclick/ThinkGO/internal/framework"
	"github.com/easeclick/ThinkGO/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Run() {
	app := thinkgo.NewApp()

	if err := app.Config().Load("config/app.yaml"); err != nil {
		slog.Warn("config not loaded, using defaults", "error", err)
	}

	app.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	// Database
	dsn := app.Config().GetString("database.dsn")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	thinkgo.DB = db
	app.SetDB(db)

	// Auto-migrate
	if app.Config().GetBool("database.auto_migrate") {
		if err := model.MigrateDB(db); err != nil {
			slog.Error("migration failed", "error", err)
			os.Exit(1)
		}
		slog.Info("database migrated")
	}

	slog.Info("worker starting")

	// Order sync ticker (every 30 minutes)
	orderSyncTicker := time.NewTicker(30 * time.Minute)
	defer orderSyncTicker.Stop()

	// Auto-purchase check ticker (every 5 minutes)
	purchaseCheckTicker := time.NewTicker(5 * time.Minute)
	defer purchaseCheckTicker.Stop()

	// Immediate first run
	go syncOrders(db)

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-orderSyncTicker.C:
			slog.Info("tick: syncing orders")
			syncOrders(db)
		case <-purchaseCheckTicker.C:
			slog.Info("tick: checking low stock")
			checkLowStock(db)
		case sig := <-sigCh:
			slog.Info("worker stopped by signal", "signal", sig)
			return
		}
	}
}

// syncOrders pulls recent orders from Shopee and stores them
func syncOrders(db *gorm.DB) {
	slog.Info("syncing orders from Shopee...")

	// TODO: replace with real Shopee API call when credentials configured
	// client := shopee.NewClient(partnerID, partnerKey, shopID)
	// orders, err := client.GetOrders(time.Now().Add(-24*time.Hour), time.Now())

	// Mock: generate a few random orders for development
	skus := []string{"ali-mock001", "ali-mock002", "ali-mock003", "ali-mock004", "ali-mock005"}
	statuses := []string{"COMPLETED", "COMPLETED", "PROCESSING", "PENDING"}
	created := 0

	for i := 0; i < 3; i++ {
		order := model.ShopOrder{
			OrderID:   "SYNC-" + time.Now().Format("150405") + "-" + randString(6),
			Amount:    float64(rand.Intn(3000)+500) / 100,
			Status:    statuses[rand.Intn(len(statuses))],
			Sku:       skus[rand.Intn(len(skus))],
			CreatedAt: time.Now(),
		}
		if err := db.Create(&order).Error; err != nil {
			slog.Warn("failed to save synced order", "error", err)
			continue
		}
		created++
	}

	slog.Info("order sync completed", "created", created)
}

// checkLowStock checks for products with stock < 10 and triggers purchase
func checkLowStock(db *gorm.DB) {
	slog.Info("checking low stock products...")

	var lowStock []model.Product
	if err := db.Where("stock < ?", 10).Find(&lowStock).Error; err != nil {
		slog.Error("low stock query failed", "error", err)
		return
	}

	if len(lowStock) == 0 {
		slog.Info("no low stock products found")
		return
	}

	for _, p := range lowStock {
		slog.Info("low stock product detected", "product_id", p.ID, "title", p.Title, "stock", p.Stock)

		// Mock purchase: create an AliPurchase record and restock
		purchase := model.AliPurchase{
			PurchaseID: "AUTO-" + time.Now().Format("150405") + "-" + randString(6),
			Cost:       float64(rand.Intn(1000)+100) / 100,
			Sku:        p.AliProductID,
			OrderRef:   "",
			CreatedAt:  time.Now(),
		}
		if err := db.Create(&purchase).Error; err != nil {
			slog.Error("failed to create purchase", "error", err)
			continue
		}

		// Restock
		db.Model(&p).Update("stock", 100)
		slog.Info("auto purchase created and restocked", "purchase_id", purchase.PurchaseID, "sku", p.AliProductID)
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
