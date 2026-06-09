package api

import (
	"log/slog"
	"os"

	gormlogger "gorm.io/gorm/logger"

	"github.com/easeclick/ThinkGO/internal/framework"
	"github.com/easeclick/ThinkGO/internal/model"
	"github.com/easeclick/ThinkGO/plugin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Run() {
	app := thinkgo.NewApp()

	if err := app.Config().Load("config/app.yaml"); err != nil {
		slog.Warn("config not loaded, using defaults", "error", err)
	}

	app.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

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

	if app.Config().GetBool("database.auto_migrate") {
		if err := model.MigrateDB(db); err != nil {
			slog.Error("migration failed", "error", err)
			os.Exit(1)
		}
		slog.Info("database migrated")
	}

	router := thinkgo.NewRouter()
	router.Use(thinkgo.Recovery())
	router.Use(thinkgo.LoggerMW())
	router.Use(thinkgo.CORSMiddleware())

	router.Get("/ping", func(ctx *thinkgo.Context) error {
		return ctx.JSON(thinkgo.Map{"message": "pong"})
	})

	// ---- Plugin system ----
	pm := plugin.NewManager(app, router)
	for _, p := range plugin.Registered() {
		pm.Register(p)
	}
	if err := pm.Boot(); err != nil {
		slog.Error("plugin boot failed", "error", err)
		os.Exit(1)
	}

	router.PrintRoutes()

	app.SetRouter(router)

	addr := thinkgo.ListenAddr(
		app.Config().GetString("server.host"),
		app.Config().GetInt("server.port"),
	)
	if err := app.Run(addr); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
