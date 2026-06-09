package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/easeclick/ThinkGO/internal/api"
	"github.com/easeclick/ThinkGO/internal/framework"
	"github.com/easeclick/ThinkGO/internal/model"
	"github.com/easeclick/ThinkGO/internal/worker"

	// Auto-register plugins via init()
	_ "github.com/easeclick/ThinkGO/plugins/alibaba"
	_ "github.com/easeclick/ThinkGO/plugins/erpcore"
	_ "github.com/easeclick/ThinkGO/plugins/shopee"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "api":
		api.Run()
	case "worker":
		worker.Run()
	case "migrate":
		runMigrate()
	case "seed":
		runSeed()
	case "help", "--help", "-h":
		printUsage()
	default:
		slog.Error("unknown command", "command", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("ThinkGo ERP System")
	fmt.Println("")
	fmt.Println("Usage: go run main.go <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  api          Start API server")
	fmt.Println("  worker       Start background worker")
	fmt.Println("  migrate      Run database migration")
	fmt.Println("  seed         Seed database with mock data")
	fmt.Println("  help         Show this help")
}

func runSeed() {
	app := thinkgo.NewApp()
	if err := app.Config().Load("config/app.yaml"); err != nil {
		slog.Warn("config not loaded, using defaults", "error", err)
	}

	dsn := app.Config().GetString("database.dsn")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("seed: database connection failed", "error", err)
		os.Exit(1)
	}

	// Ensure tables exist first
	if err := model.MigrateDB(db); err != nil {
		slog.Error("seed: migration failed", "error", err)
		os.Exit(1)
	}

	if err := model.SeedData(db); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func runMigrate() {
	app := thinkgo.NewApp()
	if err := app.Config().Load("config/app.yaml"); err != nil {
		slog.Warn("config not loaded, using defaults", "error", err)
	}

	dsn := app.Config().GetString("database.dsn")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("migration: database connection failed", "error", err)
		os.Exit(1)
	}

	if err := model.MigrateDB(db); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}

	slog.Info("migration completed successfully")
}
