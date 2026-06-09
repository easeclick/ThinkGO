// Package thinkgo — ThinkPHP-inspired Go web framework.
package thinkgo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// App is the framework application / IoC container.
// Manages service bindings, lifecycle, and configuration.
type App struct {
	mu         sync.RWMutex
	bindings   map[string]any
	singletons map[string]any
	config     *Config
	logger     *slog.Logger
	db         any // *gorm.DB stored via any to avoid hard dependency
	router     *Router
	server     *http.Server
}

// NewApp creates a new application instance.
func NewApp() *App {
	app := &App{
		bindings:   make(map[string]any),
		singletons: make(map[string]any),
	}

	// Default logger
	app.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load default config
	app.config = NewConfig()

	return app
}

// Bind registers a service factory into the container.
// factory is a function that returns the service.
func (a *App) Bind(name string, factory any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bindings[name] = factory
}

// Singleton registers a singleton service factory.
// The factory is called once, and the result is cached.
func (a *App) Singleton(name string, factory any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bindings[name] = factory
	// Mark as singleton by pre-allocating nil
	if _, exists := a.singletons[name]; !exists {
		a.singletons[name] = nil
	}
}

// Make resolves a service from the container.
// For singletons, the factory is called once and cached.
func (a *App) Make(name string) any {
	a.mu.RLock()
	factory, ok := a.bindings[name]
	a.mu.RUnlock()
	if !ok {
		return nil
	}

	// Check singleton cache
	if _, isSingleton := a.singletons[name]; isSingleton {
		a.mu.Lock()
		if cached := a.singletons[name]; cached != nil {
			a.mu.Unlock()
			return cached
		}
		// Call factory
		if fn, ok := factory.(func() any); ok {
			instance := fn()
			a.singletons[name] = instance
			a.mu.Unlock()
			return instance
		}
		a.mu.Unlock()
		return nil
	}

	// Transient: call factory each time
	if fn, ok := factory.(func() any); ok {
		return fn()
	}
	return nil
}

// Has checks if a service is registered.
func (a *App) Has(name string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	_, ok := a.bindings[name]
	return ok
}

// Config returns the application config.
func (a *App) Config() *Config {
	return a.config
}

// SetLogger sets the application logger.
func (a *App) SetLogger(logger *slog.Logger) {
	a.logger = logger
}

// Logger returns the application logger.
func (a *App) Logger() *slog.Logger {
	return a.logger
}

// SetDB stores the database instance.
func (a *App) SetDB(db any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.db = db
}

// GetDB returns the database instance.
func (a *App) GetDB() any {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.db
}

// SetRouter sets the HTTP router for the application.
func (a *App) SetRouter(router *Router) {
	a.router = router
}

// --- HTTP Server Lifecycle ---

// Run starts the HTTP server with graceful shutdown.
// This is the main entry point for the API service.
//
//	app := thinkgo.NewApp()
//	app.SetRouter(router)
//	if err := app.Run(":8888"); err != nil {
//	    log.Fatal(err)
//	}
//
// Run blocks until SIGINT/SIGTERM is received.
// Bind errors (e.g. port in use) are returned immediately.
func (a *App) Run(addr string) error {
	if a.router == nil {
		return errors.New("thinkgo: router not set, call SetRouter() before Run()")
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("thinkgo: listen %s: %w", addr, err)
	}

	a.server = &http.Server{
		Handler:      a.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	a.Logger().Info("server started", "addr", addr)

	go func() {
		if err := a.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Logger().Error("server error", "error", err)
		}
	}()

	// Wait for shutdown signal
	return a.waitForShutdown()
}

// Shutdown gracefully stops the HTTP server.
func (a *App) Shutdown() error {
	if a.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	a.Logger().Info("server shutting down...")
	return a.server.Shutdown(ctx)
}

// waitForShutdown blocks until SIGINT or SIGTERM is received.
func (a *App) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	a.Logger().Info("shutdown signal received", "signal", sig)

	// Close database connections if any
	if a.db != nil {
		if closer, ok := a.db.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				a.Logger().Error("error closing database", "error", err)
			}
		}
	}

	return a.Shutdown()
}

// ListenAddr returns the configured listen address.
// Helper for building addr string from config.
func ListenAddr(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
