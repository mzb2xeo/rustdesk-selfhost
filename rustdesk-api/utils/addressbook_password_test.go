package utils

import "testing"

func TestDecodeAddressBookPassword(t *testing.T) {
	plain := "Rd@86079"
	encoded := "UmRAODYwNzk="
	if got := DecodeAddressBookPassword(plain); got != plain {
		t.Fatalf("plain = %q", got)
	}
	if got := DecodeAddressBookPassword(encoded); got != plain {
		t.Fatalf("decoded = %q", got)
	}
	if got := DecodeAddressBookPassword(""); got != "" {
		t.Fatalf("empty = %q", got)
	}
}
