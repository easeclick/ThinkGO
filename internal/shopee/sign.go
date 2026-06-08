package shopee

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// GenerateSign generates Shopee API signature.
// Rules: sort params by key, concat as key=value&key=value, append partnerKey, SHA256.
func GenerateSign(params map[string]string, partnerKey string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	base := strings.Join(parts, "&") + partnerKey

	h := sha256.New()
	h.Write([]byte(base))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateHmacSign generates Shopee HMAC-SHA256 signature.
func GenerateHmacSign(params map[string]string, partnerKey string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	base := strings.Join(parts, "&")

	mac := hmac.New(sha256.New, []byte(partnerKey))
	mac.Write([]byte(base))
	return hex.EncodeToString(mac.Sum(nil))
}
