package thinkgo

import (
	"time"

	"gorm.io/gorm"
)

// Model is the base model providing Active Record-style operations.
// Embedded in your model structs, similar to ThinkPHP's Model class.
//
// Usage:
//
//	type User struct {
//	    thinkgo.Model
//	    Name  string `gorm:"size:100"`
//	    Email string `gorm:"uniqueIndex"`
//	}
//
//	func (u *User) TableName() string { return "users" }
//
//	// Usage:
//	user := &User{Name: "Alice", Email: "alice@example.com"}
//	thinkgo.DB.Create(user)
//
//	var users []User
//	thinkgo.DB.Where("age > ?", 18).Find(&users)
type Model struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ModelOps provides chainable query operations on a model type.
// Similar to ThinkPHP's query chaining: User.Where(...).Order(...).Find()
type ModelOps struct {
	db    *gorm.DB
	model any // model instance (for table name)
}

// DB is the package-level database for model operations.
// Set this during app initialization.
var DB *gorm.DB

// UseModel creates a ModelOps for the given model.
//
// Example:
//
//	thinkgo.UseModel(&User{}).Where("age > ?", 18).Find(&users)
func UseModel(model any) *ModelOps {
	if DB == nil {
		panic("model: DB not initialized. Call thinkgo.InitDB() or set thinkgo.DB")
	}
	return &ModelOps{db: DB.Model(model), model: model}
}

// Where adds a WHERE condition (same as GORM's Where).
func (m *ModelOps) Where(query any, args ...any) *ModelOps {
	m.db = m.db.Where(query, args...)
	return m
}

// Order adds an ORDER BY clause.
func (m *ModelOps) Order(value any) *ModelOps {
	m.db = m.db.Order(value)
	return m
}

// Limit adds a LIMIT clause.
func (m *ModelOps) Limit(limit int) *ModelOps {
	m.db = m.db.Limit(limit)
	return m
}

// Offset adds an OFFSET clause.
func (m *ModelOps) Offset(offset int) *ModelOps {
	m.db = m.db.Offset(offset)
	return m
}

// Select specifies fields to retrieve.
func (m *ModelOps) Select(query any, args ...any) *ModelOps {
	m.db = m.db.Select(query, args...)
	return m
}

// Preload specifies associations to preload (eager loading).
func (m *ModelOps) Preload(query string, args ...any) *ModelOps {
	m.db = m.db.Preload(query, args...)
	return m
}

// Group adds a GROUP BY clause.
func (m *ModelOps) Group(name string) *ModelOps {
	m.db = m.db.Group(name)
	return m
}

// Having adds a HAVING clause.
func (m *ModelOps) Having(query any, args ...any) *ModelOps {
	m.db = m.db.Having(query, args...)
	return m
}

// Find executes the query and returns results in dest.
func (m *ModelOps) Find(dest any, conds ...any) *gorm.DB {
	return m.db.Find(dest, conds...)
}

// First returns the first matching record.
func (m *ModelOps) First(dest any, conds ...any) *gorm.DB {
	return m.db.First(dest, conds...)
}

// Take returns a single record without ordering.
func (m *ModelOps) Take(dest any, conds ...any) *gorm.DB {
	return m.db.Take(dest, conds...)
}

// Pluck queries a single column and stores the result.
func (m *ModelOps) Pluck(column string, dest any) *gorm.DB {
	return m.db.Pluck(column, dest)
}

// Count returns the count of matching records.
func (m *ModelOps) Count(count *int64) *gorm.DB {
	return m.db.Count(count)
}

// Paginate is a ThinkPHP-style pagination helper.
// Returns paginated results and total count.
func (m *ModelOps) Paginate(page, pageSize int, dest any) (total int64, err error) {
	query := m.db.Session(&gorm.Session{})

	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	if err := m.db.Offset(offset).Limit(pageSize).Find(dest).Error; err != nil {
		return 0, err
	}

	return total, nil
}

// AutoMigrate runs auto-migration for the given models.
func AutoMigrate(models ...any) error {
	return DB.AutoMigrate(models...)
}
