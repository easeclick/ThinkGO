package plugin

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/easeclick/ThinkGO/internal/framework"
)

// Manager orchestrates plugin lifecycle.
type Manager struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	order   []string
	app     *thinkgo.App
	router  *thinkgo.Router
	booted  bool

	routes []RouteInfo // aggregated from all plugins
}

// NewManager creates a plugin manager bound to an app and router.
func NewManager(app *thinkgo.App, router *thinkgo.Router) *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
		app:     app,
		router:  router,
	}
}

// Register one or more plugins. Safe to call multiple times.
func (m *Manager) Register(plugins ...Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range plugins {
		id := p.ID()
		if _, exists := m.plugins[id]; exists {
			slog.Warn("plugin already registered, skipping", "id", id)
			continue
		}
		m.plugins[id] = p
		m.order = append(m.order, id)
		slog.Info("plugin registered", "id", id, "version", p.Version())
	}
	return nil
}

// Boot calls RegisterRoutes then Boot on all plugins in registration order.
// Also registers discovery routes. Safe to call once.
func (m *Manager) Boot() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.booted {
		return nil
	}

	// 1. Register AI discovery routes
	m.registerDiscoveryRoutes()

	// 2. RegisterRoutes phase — plugins declare their routes
	for _, id := range m.order {
		p := m.plugins[id]
		p.RegisterRoutes(m.router)
		m.routes = append(m.routes, p.Routes()...)
	}

	// 3. Boot phase — plugins initialize
	for _, id := range m.order {
		p := m.plugins[id]
		if err := p.Boot(m.app); err != nil {
			return err
		}
		slog.Info("plugin booted", "id", id)
	}

	m.booted = true
	return nil
}

// Shutdown gracefully stops all plugins in reverse registration order.
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.booted {
		return
	}

	for i := len(m.order) - 1; i >= 0; i-- {
		id := m.order[i]
		p := m.plugins[id]
		if err := p.Shutdown(); err != nil {
			slog.Error("plugin shutdown error", "id", id, "error", err)
		} else {
			slog.Info("plugin shut down", "id", id)
		}
	}
	m.booted = false
}

// All returns metadata for all registered plugins.
func (m *Manager) All() []PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]PluginInfo, 0, len(m.order))
	for _, id := range m.order {
		p := m.plugins[id]
		out = append(out, PluginInfo{
			ID:          p.ID(),
			Version:     p.Version(),
			Description: p.Description(),
			Routes:      p.Routes(),
		})
	}
	return out
}

// Get returns a plugin by ID, or nil.
func (m *Manager) Get(id string) Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.plugins[id]
}

// Booted returns whether Boot has been called.
func (m *Manager) Booted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.booted
}

// registerDiscoveryRoutes adds AI-friendly metadata endpoints.
func (m *Manager) registerDiscoveryRoutes() {
	m.router.Get("/-/plugins", func(c *thinkgo.Context) error {
		return c.JSON(m.All())
	})

	m.router.Get("/-/api.json", func(c *thinkgo.Context) error {
		spec := m.buildAPISpec()
		return c.JSON(spec)
	})
}

// apiSpec is the machine-readable API specification.
type apiSpec struct {
	Framework string            `json:"framework"`
	Version   string            `json:"version"`
	Plugins   []PluginInfo      `json:"plugins"`
	Routes    []json.RawMessage `json:"routes"`
}

func (m *Manager) buildAPISpec() apiSpec {
	spec := apiSpec{
		Framework: "ThinkGo",
		Version:   "1.0.0",
		Plugins:   m.All(),
		Routes:    make([]json.RawMessage, 0),
	}
	return spec
}
