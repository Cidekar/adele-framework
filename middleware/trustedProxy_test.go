package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// TestTrustedProxy_StripsForwardedHeadersFromUntrustedPeer is the core
// regression test for the IP-spoofing attack: an untrusted TCP peer must not
// be able to set X-Forwarded-For (or any header in trustedProxyHeaderNames)
// and influence what the framework's RealIP middleware writes to
// r.RemoteAddr. We model the production middleware order
// (TrustedProxy -> RealIP -> handler) and assert that a forged XFF from an
// untrusted peer is dropped before RealIP sees it.
//
// Without TrustedProxy this test would fail: RealIP would read the forged XFF
// and rewrite r.RemoteAddr to the attacker-controlled value, silently
// bypassing every downstream IP allowlist and per-IP throttle.
func TestTrustedProxy_StripsForwardedHeadersFromUntrustedPeer(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8")

	// Spy handler captures whatever r.RemoteAddr looks like AFTER both
	// TrustedProxy and RealIP have run. If the gate worked, the peer IP
	// "203.0.113.55" survives. If the gate failed, the spoofed "1.2.3.4"
	// from XFF leaks through.
	var captured string
	spy := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = r.RemoteAddr
	})

	// Wrap inner-out: TrustedProxy (outer) -> chi.RealIP -> spy.
	chain := TrustedProxy()(chimw.RealIP(spy))

	req := httptest.NewRequest(http.MethodPost, "/v1/campaign-monitor/subscriber", nil)
	req.RemoteAddr = "203.0.113.55:50000" // direct, NOT in TRUSTED_PROXIES
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	req.Header.Set("X-Real-IP", "1.2.3.4")
	req.Header.Set("True-Client-IP", "1.2.3.4")

	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	// chi.RealIP drops the port; the surviving RemoteAddr should be the
	// unmodified TCP peer (port 50000 still attached) because XFF/X-Real-
	// IP/TCIP were stripped before it ran.
	if captured != "203.0.113.55:50000" {
		t.Fatalf("expected r.RemoteAddr to remain TCP peer 203.0.113.55:50000 (XFF stripped), got %q — XFF spoofing was NOT blocked", captured)
	}
}

// TestTrustedProxy_PreservesForwardedHeadersFromTrustedPeer asserts the
// converse: a request whose immediate TCP peer IS in TRUSTED_PROXIES MUST
// retain its X-Forwarded-For header so RealIP can do its intended job of
// attributing requests to the real client IP behind the LB.
func TestTrustedProxy_PreservesForwardedHeadersFromTrustedPeer(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8")

	var captured string
	spy := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = r.RemoteAddr
	})

	chain := TrustedProxy()(chimw.RealIP(spy))

	req := httptest.NewRequest(http.MethodPost, "/v1/campaign-monitor/subscriber", nil)
	req.RemoteAddr = "10.5.5.5:60000" // trusted LB
	req.Header.Set("X-Forwarded-For", "203.0.113.99")

	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	// chi.RealIP drops the port and writes the bare IP from XFF.
	if captured != "203.0.113.99" {
		t.Fatalf("expected chi.RealIP to honor XFF from trusted peer (203.0.113.99), got %q", captured)
	}
}

// TestTrustedProxy_NoEnvVarTreatsAllPeersAsUntrusted documents the secure
// default: with TRUSTED_PROXIES unset, every peer is untrusted and every
// request has its forwarded headers stripped. This means a misconfigured
// deployment behind a real LB will attribute all activity to the LB IP —
// visible in logs/metrics — but no spoofing is possible.
func TestTrustedProxy_NoEnvVarTreatsAllPeersAsUntrusted(t *testing.T) {
	// Explicitly clear so a stray ambient env var can't taint this test.
	t.Setenv("TRUSTED_PROXIES", "")

	var captured string
	spy := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = r.RemoteAddr
	})

	chain := TrustedProxy()(chimw.RealIP(spy))

	req := httptest.NewRequest(http.MethodPost, "/v1/fingerprint-signup", nil)
	req.RemoteAddr = "10.5.5.5:60000" // would-be LB, but no env var
	req.Header.Set("X-Forwarded-For", "203.0.113.99")

	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if captured != "10.5.5.5:60000" {
		t.Fatalf("expected RemoteAddr to remain TCP peer 10.5.5.5:60000 with TRUSTED_PROXIES unset, got %q", captured)
	}
}

// TestTrustedProxy_StripsAllListedHeaders is a defense-in-depth assertion that
// every header named in trustedProxyHeaderNames is actually deleted from
// r.Header for an untrusted peer. Any future addition to that list (e.g. a new
// Forwarded-* variant) must propagate without code changes here.
func TestTrustedProxy_StripsAllListedHeaders(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8")

	var got http.Header
	spy := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
	})

	chain := TrustedProxy()(spy) // bypass chi.RealIP for direct header inspection

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.10:1234" // untrusted
	for _, h := range trustedProxyHeaderNames {
		req.Header.Set(h, "attacker-value")
	}

	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	for _, h := range trustedProxyHeaderNames {
		if v := got.Get(h); v != "" {
			t.Errorf("expected header %q to be stripped for untrusted peer, got %q", h, v)
		}
	}
}

// TestTrustedProxy_SkipsProtoHostRewriteForUntrustedClientIP exercises case
// (iii): the immediate TCP peer IS trusted (so the forwarded headers survive
// the spoofing gate), but the client IP DERIVED from X-Forwarded-For is NOT in
// TRUSTED_PROXIES. The second security layer (proto/host rewrite,
// middleware.go ~119-138) must therefore be skipped entirely: a compromised or
// misbehaving trusted hop cannot smuggle an arbitrary X-Forwarded-Proto /
// X-Forwarded-Host on behalf of an untrusted origin. We inspect the request
// directly (no chi.RealIP) so the assertions are on the proto/host effects,
// not on RemoteAddr.
func TestTrustedProxy_SkipsProtoHostRewriteForUntrustedClientIP(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8")

	var (
		scheme string
		host   string
		hasTLS bool
	)
	spy := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		scheme = r.URL.Scheme
		host = r.Host
		hasTLS = r.TLS != nil
	})

	chain := TrustedProxy()(spy) // direct inspection — no chi.RealIP

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.5.5.5:60000" // trusted peer (inside 10.0.0.0/8)
	// XFF derives an UNtrusted client IP, so the proto/host layer is gated off.
	req.Header.Set("X-Forwarded-For", "203.0.113.99")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "evil.com")

	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if scheme == "https" {
		t.Errorf("expected r.URL.Scheme NOT to be rewritten to https for untrusted derived client IP, got %q", scheme)
	}
	if hasTLS {
		t.Errorf("expected r.TLS to remain nil for untrusted derived client IP, got non-nil")
	}
	if host == "evil.com" {
		t.Errorf("expected r.Host NOT to be rewritten to attacker X-Forwarded-Host %q for untrusted derived client IP", "evil.com")
	}
}

// TestTrustedProxy_RewritesProtoHostForTrustedClientIP is the positive path of
// the second security layer: peer trusted AND the X-Forwarded-For-derived
// client IP also trusted, so X-Forwarded-Proto / X-Forwarded-Host are honored
// and rewrite r.URL.Scheme / r.TLS and r.Host / r.URL.Host.
func TestTrustedProxy_RewritesProtoHostForTrustedClientIP(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8")

	var (
		scheme  string
		host    string
		urlHost string
		hasTLS  bool
	)
	spy := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		scheme = r.URL.Scheme
		host = r.Host
		urlHost = r.URL.Host
		hasTLS = r.TLS != nil
	})

	chain := TrustedProxy()(spy) // direct inspection — no chi.RealIP

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.5.5.5:60000"              // trusted peer
	req.Header.Set("X-Forwarded-For", "10.5.5.99") // derived client also trusted
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "example.com")

	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if scheme != "https" {
		t.Errorf("expected r.URL.Scheme to be rewritten to https, got %q", scheme)
	}
	if !hasTLS {
		t.Errorf("expected r.TLS to be set non-nil after https proto rewrite, got nil")
	}
	if host != "example.com" {
		t.Errorf("expected r.Host to be rewritten to example.com, got %q", host)
	}
	if urlHost != "example.com" {
		t.Errorf("expected r.URL.Host to be rewritten to example.com, got %q", urlHost)
	}
}

// TestTrustedProxy_HonorsOnlyConfiguredForwardHeaders verifies that
// TRUST_PROXY_HEADERS is honored selectively: with only "proto" configured,
// X-Forwarded-Proto rewrites the scheme but X-Forwarded-Host is ignored even
// though the peer and derived client IP are both trusted.
func TestTrustedProxy_HonorsOnlyConfiguredForwardHeaders(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8")
	t.Setenv("TRUST_PROXY_HEADERS", "proto")

	var (
		scheme string
		host   string
	)
	spy := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		scheme = r.URL.Scheme
		host = r.Host
	})

	chain := TrustedProxy()(spy) // direct inspection — no chi.RealIP

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.5.5.5:60000"              // trusted peer
	req.Header.Set("X-Forwarded-For", "10.5.5.99") // derived client trusted
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "evil.com")

	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if scheme != "https" {
		t.Errorf("expected r.URL.Scheme to be rewritten to https (proto configured), got %q", scheme)
	}
	if host == "evil.com" {
		t.Errorf("expected r.Host NOT to be rewritten (host NOT in TRUST_PROXY_HEADERS), got %q", host)
	}
}

// TestTrustedProxy_IPv6PeerHandled confirms tcpPeerIP correctly parses an
// IPv6 bracketed RemoteAddr and that a /128 single-host CIDR trusts the peer.
// When the IPv6 peer ::1 is trusted, the forwarded headers survive and
// chi.RealIP rewrites r.RemoteAddr from X-Forwarded-For. When TRUSTED_PROXIES
// is unset, the same IPv6 peer is untrusted and its XFF is stripped, so the
// bracketed RemoteAddr survives unchanged.
func TestTrustedProxy_IPv6PeerHandled(t *testing.T) {
	t.Run("trusted IPv6 peer keeps XFF", func(t *testing.T) {
		t.Setenv("TRUSTED_PROXIES", "::1/128")

		var captured string
		spy := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			captured = r.RemoteAddr
		})

		chain := TrustedProxy()(chimw.RealIP(spy))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "[::1]:8080" // trusted IPv6 peer
		req.Header.Set("X-Forwarded-For", "203.0.113.5")

		rec := httptest.NewRecorder()
		chain.ServeHTTP(rec, req)

		if captured != "203.0.113.5" {
			t.Fatalf("expected chi.RealIP to honor XFF from trusted IPv6 peer (203.0.113.5), got %q", captured)
		}
	})

	t.Run("untrusted IPv6 peer strips XFF", func(t *testing.T) {
		t.Setenv("TRUSTED_PROXIES", "")

		var captured string
		spy := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			captured = r.RemoteAddr
		})

		chain := TrustedProxy()(chimw.RealIP(spy))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "[::1]:8080" // would-be peer, but nothing trusted
		req.Header.Set("X-Forwarded-For", "203.0.113.5")

		rec := httptest.NewRecorder()
		chain.ServeHTTP(rec, req)

		if captured != "[::1]:8080" {
			t.Fatalf("expected IPv6 RemoteAddr to remain [::1]:8080 with TRUSTED_PROXIES unset, got %q", captured)
		}
	})
}

// TestParseTrustedProxies_RejectsMalformed is a contract test on the parser:
// garbage in TRUSTED_PROXIES must NOT crash the gate, and must NOT silently
// promote the malformed entry to a trusted CIDR.
func TestParseTrustedProxies_RejectsMalformed(t *testing.T) {
	nets := parseTrustedProxies("not-a-cidr,10.0.0.0/8,also-bad")
	if len(nets) != 1 {
		t.Fatalf("expected 1 valid network, got %d (%v)", len(nets), nets)
	}
	if nets[0].String() != "10.0.0.0/8" {
		t.Errorf("expected 10.0.0.0/8, got %s", nets[0].String())
	}
}
