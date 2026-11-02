package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// VerifyHMACSHA256 verifies a signed webhook. Providers like GitHub, Stripe,
// and Slack sign each request body with a shared secret and send the result in
// a header shaped as "<prefix><hex>" (for example GitHub's "sha256=" followed
// by the hex digest). This function recomputes that signature over body under
// secret and reports whether it matches the one in header.
//
// The comparison runs in constant time via hmac.Equal, never == or
// bytes.Equal. A byte-by-byte string compare returns early on the first
// mismatched byte, so its run time leaks how many leading bytes an attacker
// guessed. hmac.Equal always reads both values in full, which denies that
// timing side channel.
//
// The contract is fail-closed: it returns false on an empty secret, an empty
// or missing prefix, malformed hex, or a digest-length mismatch. An empty
// secret means there is nothing to trust, so there is no valid signature. An
// empty prefix is a misconfiguration, not a pass: strings.HasPrefix against ""
// is always true, which would silently disable prefix checking, so an empty
// prefix is rejected outright.
//
// The caller owns how much of the request body is read before calling. This
// function hashes exactly the bytes it receives, so an unbounded body becomes
// an unbounded hash. Wrap the body with http.MaxBytesReader (or read a capped
// amount) before passing it here.
//
// Example:
//
//	body, _ := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
//	ok := helpers.VerifyHMACSHA256(secret, "sha256=", body, r.Header.Get("X-Hub-Signature-256"))
func VerifyHMACSHA256(secret, prefix string, body []byte, header string) bool {
	if secret == "" {
		return false
	}
	if prefix == "" {
		return false
	}
	if !strings.HasPrefix(header, prefix) {
		return false
	}

	want, err := hex.DecodeString(strings.TrimPrefix(header, prefix))
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	got := mac.Sum(nil)

	return hmac.Equal(got, want)
}
