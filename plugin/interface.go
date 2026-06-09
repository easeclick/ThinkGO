package plugin

import "github.com/easeclick/ThinkGO/internal/framework"

// Plugin is the interface every plugin must implement.
// Plugins self-register via init() — see registry.go.
type Plugin interface {
	// Identity
	ID() string
	Version() string
	Description() string

	// Routes returns a list of API routes this plugin registers.
	// Used by the discovery endpoint (/-/plugins).
	Routes() []RouteInfo

	// Lifecycle — called in order: RegisterRoutes → Boot → Shutdown
	RegisterRoutes(r *thinkgo.Router)
	Boot(app *thinkgo.App) error
	Shutdown() error
}

// RouteInfo describes a single route for AI discovery and documentation.
type RouteInfo struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	Auth        string `json:"auth,omitempty"` // "none" | "basic" | "token" | ...
}

// PluginInfo is the full metadata exposed by the discovery endpoint.
type PluginInfo struct {
	ID          string      `json:"id"`
	Version     string      `json:"version"`
	Description string      `json:"description"`
	Routes      []RouteInfo `json:"routes,omitempty"`
	Config      any         `json:"config,omitempty"` // JSON Schema when provided
}

// BasePlugin provides no-op defaults so plugins only override what they need.
type BasePlugin struct{}

func (BasePlugin) Version() string          { return "1.0.0" }
func (BasePlugin) Description() string      { return "" }
func (BasePlugin) Routes() []RouteInfo      { return nil }
func (BasePlugin) RegisterRoutes(r *thinkgo.Router) {}
func (BasePlugin) Boot(app *thinkgo.App) error  { return nil }
func (BasePlugin) Shutdown() error              { return nil }
