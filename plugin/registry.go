package plugin

import "sync"

var (
	mu      sync.RWMutex
	entries []Plugin
)

// Register adds a plugin to the global registry.
// Plugins call this from their init() function.
// Safe for concurrent use; duplicates by ID are silently skipped.
func Register(p Plugin) {
	mu.Lock()
	defer mu.Unlock()
	for _, existing := range entries {
		if existing.ID() == p.ID() {
			return
		}
	}
	entries = append(entries, p)
}

// Registered returns a snapshot of all registered plugins.
func Registered() []Plugin {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Plugin, len(entries))
	copy(out, entries)
	return out
}

// Reset clears the global registry (for tests).
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	entries = nil
}
