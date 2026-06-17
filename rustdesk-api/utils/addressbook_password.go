package utils

import (
	"encoding/base64"
	"strings"
)

// DecodeAddressBookPassword normalizes deploy payloads to plain text.
// Shared address books expect a plain password; older deploy scripts sent base64.
func DecodeAddressBookPassword(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return value
	}
	if len(decoded) == 0 {
		return value
	}
	for _, b := range decoded {
		if b < 0x20 || b > 0x7E {
			return value
		}
	}
	return string(decoded)
}
