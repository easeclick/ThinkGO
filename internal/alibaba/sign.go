package alibaba

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

// generateSign generates 1688 API signature.
// 1688 rule: secret + key1 + value1 + key2 + value2 + ... + secret, then SHA256.
func generateSign(params url.Values, appSecret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(appSecret)
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(params.Get(k))
	}
	sb.WriteString(appSecret)

	h := sha256.New()
	h.Write([]byte(sb.String()))
	return hex.EncodeToString(h.Sum(nil))
}
