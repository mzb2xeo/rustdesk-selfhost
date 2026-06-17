package utils

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

// EncodeRustDeskConfig builds the obfuscated config string for `rustdesk --config`.
// Format: reverse(base64(JSON).replace("=", "")) with keys host, relay, api, key.
func EncodeRustDeskConfig(idServer, relayServer, apiServer, key string) string {
	host := stripPort(idServer, ":21116")
	relay := stripPort(relayServer, ":21117")
	if relay == "" {
		relay = host
	}
	key = strings.TrimSpace(key)
	payload, _ := json.Marshal(map[string]string{
		"host":  host,
		"relay": relay,
		"api":   strings.TrimRight(strings.TrimSpace(apiServer), "/"),
		"key":   key,
	})
	encoded := base64.StdEncoding.EncodeToString(payload)
	encoded = strings.ReplaceAll(encoded, "=", "")
	runes := []rune(encoded)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func stripPort(value, portSuffix string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasSuffix(value, portSuffix) {
		return strings.TrimSuffix(value, portSuffix)
	}
	if idx := strings.LastIndex(value, ":"); idx > 0 {
		return value[:idx]
	}
	return value
}
