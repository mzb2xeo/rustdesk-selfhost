package utils

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestEncodeRustDeskConfig_roundTripShape(t *testing.T) {
	cfg := EncodeRustDeskConfig("rd.example.com:21116", "rd.example.com:21117", "https://api.example.com", "abcKEY")
	if cfg == "" {
		t.Fatal("expected non-empty config string")
	}
	if strings.Contains(cfg, "=") {
		t.Fatal("config string must not contain padding")
	}

	reversed := []rune(cfg)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	padding := (4 - len(reversed)%4) % 4
	raw, err := base64.StdEncoding.DecodeString(string(reversed) + strings.Repeat("=", padding))
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	var decoded map[string]string
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if decoded["host"] != "rd.example.com" {
		t.Fatalf("host = %q", decoded["host"])
	}
	if decoded["relay"] != "rd.example.com" {
		t.Fatalf("relay = %q", decoded["relay"])
	}
	if decoded["api"] != "https://api.example.com" {
		t.Fatalf("api = %q", decoded["api"])
	}
	if decoded["key"] != "abcKEY" {
		t.Fatalf("key = %q", decoded["key"])
	}
}
