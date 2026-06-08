package thinkgo

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBManager manages database connections.
// Supports multiple connections (like ThinkPHP's database config).
type DBManager struct {
	mu          sync.RWMutex
	connections map[string]*gorm.DB
	defaultName string
	logger      logger.Interface
}

// NewDBManager creates a new database manager.
func NewDBManager() *DBManager {
	return &DBManager{
		connections: make(map[string]*gorm.DB),
		defaultName: "default",
	}
}

// Register adds a GORM database connection.
func (m *DBManager) Register(name string, db *gorm.DB) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[name] = db
	if m.defaultName == "" {
		m.defaultName = name
	}
}

// Get returns a database connection by name.
// Returns the default connection if name is empty.
func (m *DBManager) Get(name ...string) *gorm.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()

	n := m.defaultName
	if len(name) > 0 && name[0] != "" {
		n = name[0]
	}

	db, ok := m.connections[n]
	if !ok {
		panic(fmt.Sprintf("db: connection %q not registered", n))
	}
	return db
}

// SetDefault sets the default connection name.
func (m *DBManager) SetDefault(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultName = name
}

// Close closes all database connections.
func (m *DBManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, db := range m.connections {
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("db: close %s: %w", name, err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("db: close %s: %w", name, err)
		}
	}
	return nil
}
