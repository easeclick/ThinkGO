package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// MD5 returns the MD5 hash of a string
func MD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// Now returns the current time as string
func Now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// Today returns today's date string
func Today() string {
	return time.Now().Format("2006-01-02")
}

// InArray checks if a string is in a slice
func InArray(needle string, haystack []string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// DD dumps a value and stops (ThinkPHP-style debug)
func DD(v ...interface{}) {
	for _, item := range v {
		fmt.Printf("%+v\n", item)
	}
}

// RandomString generates a random alphanumeric string
func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// PriceToString formats a price with 2 decimal places
func PriceToString(price float64) string {
	return fmt.Sprintf("%.2f", price)
}

// Truncate truncates a string to maxLen
func Truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen-3]) + "..."
}

// Slug converts a string to URL-friendly slug
func Slug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	// Remove non-alphanumeric except dash
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}
