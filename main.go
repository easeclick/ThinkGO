package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/user/thinkgo/internal/api"
	"github.com/user/thinkgo/internal/framework"
	"github.com/user/thinkgo/internal/model"
	"github.com/user/thinkgo/internal/worker"
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
	fmt.Println("  help         Show this help")
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
