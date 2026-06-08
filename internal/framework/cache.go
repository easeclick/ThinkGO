package thinkgo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CacheStore defines the interface for cache operations.
// ThinkPHP-style: Set/Get/Delete/Clear with TTL support.
type CacheStore interface {
	// Get retrieves a cached value. Returns nil if not found or expired.
	Get(key string) (any, bool)

	// Set stores a value with TTL (0 = forever).
	Set(key string, value any, ttl time.Duration) error

	// Delete removes a cached value.
	Delete(key string) error

	// Clear removes all cached values.
	Clear() error

	// Has checks if a key exists and is not expired.
	Has(key string) bool
}

// MemoryCache is an in-memory cache implementation.
// Default cache store for development.
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]memItem
}

type memItem struct {
	value     any
	expiresAt time.Time
}

// NewMemoryCache creates an in-memory cache.
func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{
		items: make(map[string]memItem),
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

func (c *MemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		c.Delete(key)
		return nil, false
	}

	return item.value, true
}

func (c *MemoryCache) Set(key string, value any, ttl time.Duration) error {
	item := memItem{value: value}
	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
	}

	c.mu.Lock()
	c.items[key] = item
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) Delete(key string) error {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) Clear() error {
	c.mu.Lock()
	c.items = make(map[string]memItem)
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) Has(key string) bool {
	_, ok := c.Get(key)
	return ok
}

func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, v := range c.items {
			if !v.expiresAt.IsZero() && now.After(v.expiresAt) {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}

// FileCache is a file-based cache implementation.
// Stores cached items as JSON files in a directory.
type FileCache struct {
	mu   sync.RWMutex
	dir  string
	ttl  time.Duration
}

// NewFileCache creates a file-based cache in the given directory.
func NewFileCache(dir string) *FileCache {
	os.MkdirAll(dir, 0755)
	return &FileCache{
		dir:  dir,
		ttl:  0, // forever by default
	}
}

func (c *FileCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := os.ReadFile(c.filePath(key))
	if err != nil {
		return nil, false
	}

	var item memItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, false
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		os.Remove(c.filePath(key))
		return nil, false
	}

	return item.value, true
}

func (c *FileCache) Set(key string, value any, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := memItem{value: value}
	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
	}

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return os.WriteFile(c.filePath(key), data, 0644)
}

func (c *FileCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return os.Remove(c.filePath(key))
}

func (c *FileCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			os.Remove(filepath.Join(c.dir, entry.Name()))
		}
	}
	return nil
}

func (c *FileCache) Has(key string) bool {
	_, ok := c.Get(key)
	return ok
}

func (c *FileCache) filePath(key string) string {
	return filepath.Join(c.dir, key+".cache")
}

// Ensure MemoryCache implements CacheStore
var _ CacheStore = (*MemoryCache)(nil)
var _ CacheStore = (*FileCache)(nil)
