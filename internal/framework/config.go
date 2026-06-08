package thinkgo

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config manages application configuration.
// Supports YAML files and direct key-value access with dot notation.
type Config struct {
	mu     sync.RWMutex
	data   map[string]any
	loaded bool
}

// NewConfig creates an empty config.
func NewConfig() *Config {
	return &Config{
		data: make(map[string]any),
	}
}

// Load reads a YAML config file and merges it into the current config.
func (c *Config) Load(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("config: read file %s: %w", path, err)
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return fmt.Errorf("config: parse yaml %s: %w", path, err)
	}

	if c.data == nil {
		c.data = make(map[string]any)
	}
	deepMerge(c.data, parsed)
	c.loaded = true
	return nil
}

// Get returns a config value by dot-notation key.
// Returns nil if the key is not found.
func (c *Config) Get(key string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return deepGet(c.data, key)
}

// GetString returns a string config value.
func (c *Config) GetString(key string) string {
	v := c.Get(key)
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// GetInt returns an int config value.
func (c *Config) GetInt(key string) int {
	v := c.Get(key)
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case int64:
		return int(n)
	}
	return 0
}

// GetBool returns a bool config value.
func (c *Config) GetBool(key string) bool {
	v := c.Get(key)
	if v == nil {
		return false
	}
	b, _ := v.(bool)
	return b
}

// Set sets a config value by dot-notation key.
func (c *Config) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := parseDotKey(key)
	if len(keys) == 0 {
		return
	}
	setDeep(c.data, keys, value)
}

// All returns a copy of all config data.
func (c *Config) All() map[string]any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]any, len(c.data))
	for k, v := range c.data {
		result[k] = v
	}
	return result
}

// IsLoaded returns whether config has been loaded from a file.
func (c *Config) IsLoaded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded
}

// deepMerge merges src into dst recursively.
func deepMerge(dst, src map[string]any) {
	for k, sv := range src {
		dv, exists := dst[k]
		if !exists {
			dst[k] = sv
			continue
		}
		srcMap, srcOk := sv.(map[string]any)
		dstMap, dstOk := dv.(map[string]any)
		if srcOk && dstOk {
			deepMerge(dstMap, srcMap)
		} else {
			dst[k] = sv
		}
	}
}

// deepGet traverses a nested map by dot-notation key.
func deepGet(m map[string]any, key string) any {
	keys := parseDotKey(key)
	current := m
	for i, k := range keys {
		if i == len(keys)-1 {
			return current[k]
		}
		if next, ok := current[k].(map[string]any); ok {
			current = next
		} else {
			return nil
		}
	}
	return nil
}

// setDeep sets a value in a nested map by key path.
func setDeep(m map[string]any, keys []string, value any) {
	current := m
	for i, k := range keys {
		if i == len(keys)-1 {
			current[k] = value
			return
		}
		if next, ok := current[k].(map[string]any); ok {
			current = next
		} else {
			next = make(map[string]any)
			current[k] = next
			current = next
		}
	}
}

// parseDotKey splits "a.b.c" into ["a","b","c"].
func parseDotKey(key string) []string {
	if key == "" {
		return nil
	}
	var keys []string
	start := 0
	for i := 0; i < len(key); i++ {
		if key[i] == '.' {
			if i > start {
				keys = append(keys, key[start:i])
			}
			start = i + 1
		}
	}
	if start < len(key) {
		keys = append(keys, key[start:])
	}
	return keys
}
