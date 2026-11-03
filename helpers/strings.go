package helpers

import (
	"crypto/rand"
	"strings"
)

// RandomString generates a cryptographically secure random string of specified length using
// crypto/rand for entropy. The generated string contains alphanumeric characters (a-z, A-Z, 0-9)
// suitable for tokens, IDs, passwords, and other security-sensitive applications.
//
// Example:
//
//	sessionID := helpers.RandomString(32)
//	// Returns: "K7mX9pQr2NvB8zYtF3wL5cR1jP6sM4xH"
//
//	apiKey := helpers.RandomString(16)
//	// Returns: "aB9Xm2Qr7Kp3Nv8Z"
//
//	tempPassword := helpers.RandomString(12)
//	// Returns: "M9sK2pR7bN4X"
func (h *Helpers) RandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		randomByte := make([]byte, 1)
		rand.Read(randomByte)
		b[i] = charset[randomByte[0]%byte(len(charset))]
	}
	return string(b)
}

// IsEmpty reports whether s is the empty string. It checks for exact emptiness
// and does not trim whitespace; use IsBlank to treat whitespace-only strings as
// empty.
//
// Examples:
//
//	h.IsEmpty("")
//	// Result: true
//	h.IsEmpty(" ")
//	// Result: false
//	h.IsEmpty("x")
//	// Result: false
func (h *Helpers) IsEmpty(s string) bool {
	return IsEmpty(s)
}

// IsBlank reports whether s is empty or contains only whitespace. Whitespace is
// stripped with strings.TrimSpace before the comparison.
//
// Examples:
//
//	h.IsBlank("")
//	// Result: true
//	h.IsBlank("   ")
//	// Result: true
//	h.IsBlank("\t\n")
//	// Result: true
//	h.IsBlank("x")
//	// Result: false
//	h.IsBlank(" x ")
//	// Result: false
func (h *Helpers) IsBlank(s string) bool {
	return IsBlank(s)
}

// IsEmpty reports whether s is the empty string
func IsEmpty(s string) bool {
	return s == ""
}

// IsBlank reports whether s is empty or only whitespace
func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}
