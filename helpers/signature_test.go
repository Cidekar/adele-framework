package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

// sign computes a valid signature header for the given secret, prefix, and body.
func sign(secret, prefix string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return prefix + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyHMACSHA256(t *testing.T) {
	const secret = "s3cr3t"
	const prefix = "sha256="
	var body = []byte("payload")

	// validNoPrefix is a correct lowercase-hex signature with the prefix stripped.
	validNoPrefix := strings.TrimPrefix(sign(secret, prefix, body), prefix)

	// hex that decodes to 31 bytes (62 chars) and 33 bytes (66 chars).
	hex31 := strings.Repeat("ab", 31)
	hex33 := strings.Repeat("ab", 33)

	cases := []struct {
		name   string
		secret string
		prefix string
		body   []byte
		header string
		want   bool
	}{
		{
			name:   "valid signature",
			secret: secret,
			prefix: prefix,
			body:   body,
			header: sign(secret, prefix, body),
			want:   true,
		},
		{
			name:   "tampered body",
			secret: secret,
			prefix: prefix,
			body:   []byte("PAYLOAD"),
			header: sign(secret, prefix, []byte("payload")),
			want:   false,
		},
		{
			name:   "wrong secret",
			secret: secret,
			prefix: prefix,
			body:   body,
			header: sign("other", prefix, body),
			want:   false,
		},
		{
			name:   "missing header",
			secret: secret,
			prefix: prefix,
			body:   body,
			header: "",
			want:   false,
		},
		{
			name:   "bad hex",
			secret: secret,
			prefix: prefix,
			body:   body,
			header: "sha256=zzzz",
			want:   false,
		},
		{
			name:   "empty secret",
			secret: "",
			prefix: prefix,
			body:   body,
			header: sign(secret, prefix, body),
			want:   false,
		},
		{
			name:   "empty prefix",
			secret: secret,
			prefix: "",
			body:   body,
			header: validNoPrefix,
			want:   false,
		},
		{
			name:   "wrong prefix",
			secret: secret,
			prefix: prefix,
			body:   body,
			header: sign(secret, "sha1=", body),
			want:   false,
		},
		{
			name:   "no prefix at all",
			secret: secret,
			prefix: prefix,
			body:   body,
			header: validNoPrefix,
			want:   false,
		},
		{
			// hex.DecodeString is case-insensitive, so an uppercase digest
			// decodes to the same bytes and still verifies. Case is an encoding
			// artifact, not a property of the signature.
			name:   "uppercase hex",
			secret: secret,
			prefix: prefix,
			body:   body,
			header: prefix + strings.ToUpper(validNoPrefix),
			want:   true,
		},
		{
			name:   "valid over empty body",
			secret: secret,
			prefix: prefix,
			body:   []byte{},
			header: sign(secret, prefix, []byte{}),
			want:   true,
		},
		{
			name:   "nil body signed as nil",
			secret: secret,
			prefix: prefix,
			body:   nil,
			header: sign(secret, prefix, nil),
			want:   true,
		},
		{
			name:   "nil body signed as empty slice",
			secret: secret,
			prefix: prefix,
			body:   nil,
			header: sign(secret, prefix, []byte{}),
			want:   true,
		},
		{
			name:   "length mismatch 31 bytes",
			secret: secret,
			prefix: prefix,
			body:   body,
			header: prefix + hex31,
			want:   false,
		},
		{
			name:   "length mismatch 33 bytes",
			secret: secret,
			prefix: prefix,
			body:   body,
			header: prefix + hex33,
			want:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := VerifyHMACSHA256(tc.secret, tc.prefix, tc.body, tc.header)
			if got != tc.want {
				t.Fatalf("VerifyHMACSHA256(%q, %q, %q, %q) = %v, want %v",
					tc.secret, tc.prefix, tc.body, tc.header, got, tc.want)
			}
		})
	}
}

// TestVerifyHMACSHA256_KnownAnswer pins a hard-coded RFC-style test vector so a
// matching bug in both sign() and the implementation cannot cancel out.
func TestVerifyHMACSHA256_KnownAnswer(t *testing.T) {
	const secret = "key"
	body := []byte("The quick brown fox jumps over the lazy dog")
	const knownHex = "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"

	validHeader := "sha256=" + knownHex
	if !VerifyHMACSHA256(secret, "sha256=", body, validHeader) {
		t.Fatalf("VerifyHMACSHA256 with known-answer vector = false, want true")
	}

	// Flip one hex character; the result must no longer verify.
	tampered := "sha256=07bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"
	if VerifyHMACSHA256(secret, "sha256=", body, tampered) {
		t.Fatalf("VerifyHMACSHA256 with tampered known-answer vector = true, want false")
	}
}
