package thinkgo

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"reflect"
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

// Token generates a random hex token string.
// The resulting length is 2 * length (hex encoding).
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

// StructToMap converts a struct to map[string]any using field names as keys.
// Supports:
//   - Nested structs (flattened with dot notation)
//   - `json` tags for custom key names
//   - Unexported fields are skipped
//   - Pointer fields (nil pointers become zero values)
func StructToMap(obj any) map[string]any {
	result := make(map[string]any)
	v := reflect.ValueOf(obj)

	// Unwrap pointer
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return result
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		result["value"] = obj
		return result
	}

	structToMapRecursive(v, "", result)
	return result
}

func structToMapRecursive(v reflect.Value, prefix string, result map[string]any) {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Determine key name: json tag > field name
		key := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			if idx := strings.Index(jsonTag, ","); idx >= 0 {
				key = jsonTag[:idx]
			} else {
				key = jsonTag
			}
			if key == "-" {
				continue // explicitly skipped
			}
		}

		if prefix != "" {
			key = prefix + "." + key
		}

		// Unwrap pointer
		for fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				key = "" // signal to skip
				break
			}
			fieldVal = fieldVal.Elem()
		}
		if key == "" {
			continue
		}

		switch fieldVal.Kind() {
		case reflect.Struct:
			// Check if it's a time.Time or other well-known type
			if _, ok := fieldVal.Interface().(time.Time); ok {
				result[key] = fieldVal.Interface()
			} else {
				// Recurse into nested struct
				structToMapRecursive(fieldVal, key, result)
			}
		case reflect.Slice, reflect.Array:
			result[key] = fieldVal.Interface()
		case reflect.Map:
			result[key] = fieldVal.Interface()
		default:
			result[key] = fieldVal.Interface()
		}
	}
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
