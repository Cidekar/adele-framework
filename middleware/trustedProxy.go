package middleware

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"slices"
	"strings"
)

// TrustedProxy is a middleware that securely establishes the runtime trust
// model for requests arriving behind a reverse proxy. It is a single,
// self-contained middleware that performs BOTH security layers, because the
// framework has no outer server-level wrapper:
//
//  1. The spoofing gate. TrustedProxy is mounted in BootstrapMux BEFORE
//     RealIP (adele.go:379, ahead of RequestID:380 and RealIP:381), so the
//     request's r.RemoteAddr at this point is still the actual kernel-supplied
//     TCP peer. If that peer is NOT in TRUSTED_PROXIES, every header in
//     trustedProxyHeaderNames is deleted so that the framework's RealIP (which
//     rewrites r.RemoteAddr from True-Client-IP / X-Real-IP / X-Forwarded-For
//     with no trust validation) becomes a no-op and r.RemoteAddr stays the
//     true peer. This prevents a malicious client from spoofing its source IP
//     and defeating any per-IP authorization or throttling downstream.
//
//  2. The proto/host rewrite. When the peer IS trusted, the forwarded headers
//     survive, and if the derived client IP is also trusted, X-Forwarded-Proto
//     and X-Forwarded-Host are honored to rewrite r.URL.Scheme / r.TLS and
//     r.Host / r.URL.Host respectively.
//
// Configuration via environment variables (read at construction with
// os.Getenv — this is a package-level middleware with no a.Helpers access):
//
//	TRUSTED_PROXIES:     Comma-separated list of trusted proxy IPs/CIDRs.
//	                     Examples: "127.0.0.1,192.168.1.0/24" or "10.0.0.0/8"
//	TRUST_PROXY_HEADERS: Comma-separated list of headers to trust.
//	                     Examples: "proto,host" or "proto,host,port,for"
//
// Secure by default: with TRUSTED_PROXIES unset, no peer is trusted, so the
// forwarded headers are always stripped, RealIP becomes a no-op, and
// r.RemoteAddr remains the true TCP peer. A misconfigured deployment behind a
// real load balancer will attribute all activity to the LB IP (visible in
// logs/metrics) but no spoofing is possible.
//
// Security considerations:
//   - Never set TRUSTED_PROXIES to "*" or "0.0.0.0/0" in production.
//   - Only include your actual reverse-proxy IPs (LB subnets, ingress CIDRs).
//   - Headers from untrusted peers are completely ignored.
func TrustedProxy() func(h http.Handler) http.Handler {
	// Parse trusted proxy configuration once at construction.
	trustedProxies := parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))
	trustedHeaders := parseTrustedHeaders(os.Getenv("TRUST_PROXY_HEADERS"))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the TCP peer. RealIP has not run yet (TrustedProxy is
			// mounted before it), so r.RemoteAddr is the actual kernel-
			// supplied source. We deliberately do NOT consult any X-* header
			// here — that would defeat the purpose of the gate.
			peerIP := tcpPeerIP(r)

			if !isTrustedProxy(peerIP, trustedProxies) {
				// Untrusted peer — strip every header that downstream code
				// (including RealIP) could use to spoof a different source
				// IP / scheme / host, then continue.
				for _, h := range trustedProxyHeaderNames {
					r.Header.Del(h)
				}
				next.ServeHTTP(w, r)
				return
			}

			// Trusted peer — the forwarded headers survived. Derive the
			// client IP and, if it is also trusted, honor proto/host.
			clientIP := getClientIP(r)
			if isTrustedProxy(clientIP, trustedProxies) {
				if contains(trustedHeaders, "proto") {
					if proto := r.Header.Get("X-Forwarded-Proto"); proto == "https" {
						r.URL.Scheme = "https"
						r.TLS = &tls.ConnectionState{}
					}
				}

				if contains(trustedHeaders, "host") {
					if host := r.Header.Get("X-Forwarded-Host"); host != "" {
						r.Host = host
						r.URL.Host = host
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// tcpPeerIP returns the IP portion of r.RemoteAddr (drops the port). Used by
// TrustedProxy to evaluate whether the immediate TCP peer is a trusted reverse
// proxy. We deliberately do NOT consult any X-* headers here — that would
// defeat the purpose of the gate.
func tcpPeerIP(r *http.Request) string {
	if r.RemoteAddr == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// parseTrustedProxies converts environment string to list of trusted networks.
func parseTrustedProxies(proxyList string) []*net.IPNet {
	if proxyList == "" {
		return nil // No proxies trusted by default
	}

	var networks []*net.IPNet

	for proxy := range strings.SplitSeq(proxyList, ",") {
		proxy = strings.TrimSpace(proxy)
		if proxy == "" {
			continue
		}

		// Handle single IP (add /32 or /128)
		if !strings.Contains(proxy, "/") {
			if strings.Contains(proxy, ":") {
				proxy += "/128" // IPv6
			} else {
				proxy += "/32" // IPv4
			}
		}

		_, network, err := net.ParseCIDR(proxy)
		if err == nil {
			networks = append(networks, network)
		}
	}

	return networks
}

// parseTrustedHeaders converts environment string to list of trusted headers.
func parseTrustedHeaders(headerList string) []string {
	if headerList == "" {
		return []string{"proto", "host"} // Safe defaults
	}

	headers := strings.Split(headerList, ",")
	for i, header := range headers {
		headers[i] = strings.TrimSpace(header)
	}
	return headers
}

// getClientIP extracts the real client IP, handling various proxy scenarios.
func getClientIP(r *http.Request) string {
	// Try X-Real-IP first (most reliable from reverse proxy)
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Try X-Forwarded-For (could be comma-separated list)
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// Take the first IP (original client)
		if parts := strings.Split(ip, ","); len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// Fall back to direct connection IP
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}

// isTrustedProxy checks if the given IP is in the trusted proxy list.
func isTrustedProxy(ip string, trustedNetworks []*net.IPNet) bool {
	if len(trustedNetworks) == 0 {
		return false // No proxies trusted
	}

	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}

	for _, network := range trustedNetworks {
		if network.Contains(clientIP) {
			return true
		}
	}

	return false
}

// contains checks if a string slice contains a specific string.
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}
