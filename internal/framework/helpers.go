package thinkgo

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// --- ThinkPHP-style helper functions ---

// DD dumps the given values and exits (like ThinkPHP's dump+die).
func DD(vals ...any) {
	for _, v := range vals {
		fmt.Printf("%+v\n", v)
	}
}

// Dump prints a variable in readable format.
func Dump(vals ...any) {
	for _, v := range vals {
		fmt.Printf("%+v ", v)
	}
	fmt.Println()
}

// MD5 returns the MD5 hash of a string.
func MD5(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

// Token generates a random token string.
func Token(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Now returns the current time in the given format.
// Format follows ThinkPHP convention: "Y-m-d H:i:s" maps to Go format.
func Now(format string) string {
	goFormat := thinkFormatToGo(format)
	return time.Now().Format(goFormat)
}

// thinkFormatToGo converts ThinkPHP date format to Go format.
func thinkFormatToGo(format string) string {
	replacements := map[string]string{
		"Y": "2006",
		"m": "01",
		"d": "02",
		"H": "15",
		"i": "04",
		"s": "05",
	}
	result := format
	for tp, goFmt := range replacements {
		result = strings.ReplaceAll(result, tp, goFmt)
	}
	return result
}

// InArray checks if a value exists in a slice (like PHP in_array).
func InArray[T comparable](needle T, haystack []T) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

// ArrayColumn extracts a column from a slice of maps.
func ArrayColumn[T any](items []map[string]T, key string) []T {
	result := make([]T, 0, len(items))
	for _, item := range items {
		if v, ok := item[key]; ok {
			result = append(result, v)
		}
	}
	return result
}

// Default returns the first non-zero value (like PHP ?? operator).
func Default[T comparable](val, def T) T {
	var zero T
	if val == zero {
		return def
	}
	return val
}

// StructToMap converts a struct to a map using field names.
// Uses fmt.Sprintf for values - use with simple types.
func StructToMap(obj any) map[string]any {
	result := make(map[string]any)
	// This is a simplified placeholder.
	// In production, use reflection or json marshal/unmarshal.
	return result
}

// Pagination is a ThinkPHP-style pagination result.
type Pagination struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Pages    int   `json:"pages"`
	HasPrev  bool  `json:"has_prev"`
	HasNext  bool  `json:"has_next"`
}

// NewPagination creates a pagination info object.
func NewPagination(total int64, page, pageSize int) Pagination {
	pages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		pages++
	}

	return Pagination{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
		HasPrev:  page > 1,
		HasNext:  page < pages,
	}
}
