package worker

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	gormlogger "gorm.io/gorm/logger"

	"github.com/user/thinkgo/internal/framework"
	"github.com/user/thinkgo/internal/model"
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

	// TODO: implement actual Shopee order sync when credentials configured
	// client := shopee.NewClient(partnerID, partnerKey, shopID)
	// orders, err := client.GetOrders(time.Now().Add(-24*time.Hour), time.Now())
	// if err != nil { slog.Error("failed to sync orders", "error", err); return }
	// for _, o := range orders { db.Save(&o) }

	slog.Info("order sync completed (no-op: credentials not configured)")
}

// checkLowStock checks for products with stock < 10 and triggers purchase
func checkLowStock(db *gorm.DB) {
	slog.Info("checking low stock products...")

	// TODO: implement actual low stock check when credentials configured
	// var lowStock []model.Product
	// db.Where("stock < ?", 10).Find(&lowStock)
	// for _, p := range lowStock {
	//     orderID, err := alibaba.CreateDropShippingOrder(p.AliProductID, 50, "warehouse")
	//     if err != nil { slog.Error("auto purchase failed", "error", err); continue }
	//     slog.Info("auto purchase created", "order_id", orderID)
	// }

	slog.Info("low stock check completed (no-op: credentials not configured)")
}
