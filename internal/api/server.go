package api

import (
	"log/slog"
	"os"
	"time"

	gormlogger "gorm.io/gorm/logger"

	"github.com/user/thinkgo/internal/erpcore"
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

	// Database connection
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
	slog.Info("database connected", "dsn", dsn)

	// Auto-migrate
	if app.Config().GetBool("database.auto_migrate") {
		if err := model.MigrateDB(db); err != nil {
			slog.Error("migration failed", "error", err)
			os.Exit(1)
		}
		slog.Info("database migrated")
	}

	// Router setup
	router := thinkgo.NewRouter()
	router.Use(thinkgo.Recovery())
	router.Use(thinkgo.LoggerMW())
	router.Use(thinkgo.CORSMiddleware())

	registerRoutes(router, db)
	router.PrintRoutes()

	app.SetRouter(router)

	// Start server with graceful shutdown
	addr := thinkgo.ListenAddr(
		app.Config().GetString("server.host"),
		app.Config().GetInt("server.port"),
	)
	if err := app.Run(addr); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func registerRoutes(r *thinkgo.Router, db *gorm.DB) {
	r.Get("/ping", func(ctx *thinkgo.Context) error {
		return ctx.JSON(thinkgo.Map{
			"message": "pong",
		})
	})

	v1 := r.Group("/api/v1")
	v1.Get("/report/monthly", func(ctx *thinkgo.Context) error {
		return handleMonthlyReport(ctx, db)
	})
}

func handleMonthlyReport(ctx *thinkgo.Context, db *gorm.DB) error {
	year := ctx.QueryInt("year")
	if year <= 0 {
		year = time.Now().Year()
	}
	month := ctx.QueryInt("month")
	if month < 1 || month > 12 {
		month = int(time.Now().Month())
	}

	location := time.UTC
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, location)
	endDate := startDate.AddDate(0, 1, 0)

	var dailyProfits []*erpcore.DailyProfit
	totalProfit := 0.0
	totalSales := 0.0
	totalCost := 0.0

	current := startDate
	for current.Before(endDate) {
		profit, err := erpcore.CalculateDailyProfit(db, current)
		if err != nil {
			return ctx.Error("profit calculation failed: " + err.Error())
		}
		dailyProfits = append(dailyProfits, profit)
		totalProfit += profit.TotalProfit
		totalSales += profit.TotalSales
		totalCost += profit.TotalCost
		current = current.AddDate(0, 0, 1)
	}

	type SkuSales struct {
		Sku    string
		Amount float64
	}
	var rows []SkuSales
	db.Table("shop_orders").
		Select("sku, SUM(amount) as amount").
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Group("sku").
		Find(&rows)

	var bestSku, worstSku string
	var bestSales float64 = -1
	var worstSales float64 = -1
	for _, r := range rows {
		if bestSales < 0 || r.Amount > bestSales {
			bestSales = r.Amount
			bestSku = r.Sku
		}
		if worstSales < 0 || r.Amount < worstSales {
			worstSales = r.Amount
			worstSku = r.Sku
		}
	}

	return ctx.JSON(thinkgo.Map{
		"year":             year,
		"month":            month,
		"total_profit":     totalProfit,
		"total_sales":      totalSales,
		"total_cost":       totalCost,
		"total_orders":     len(rows),
		"best_seller_sku":  bestSku,
		"worst_seller_sku": worstSku,
		"daily_breakdown":  dailyProfits,
	})
}
