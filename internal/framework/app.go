// Package thinkgo — ThinkPHP-inspired Go web framework.
package thinkgo

import (
	"log/slog"
	"os"
	"sync"
)

// App is the framework application / IoC container.
// Manages service bindings, lifecycle, and configuration.
type App struct {
	mu       sync.RWMutex
	bindings map[string]any
	singletons map[string]any
	config   *Config
	logger   *slog.Logger
	db       any // *gorm.DB stored via any to avoid hard dependency
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

// Run starts the HTTP server.
func (a *App) Run(addr string) error {
	a.Logger().Info("starting server", "addr", addr)
	return nil // Router will be attached externally
}
