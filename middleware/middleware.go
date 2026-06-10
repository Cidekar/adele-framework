// Package middleware provides HTTP middleware for Adele applications, including
// real-IP resolution, request IDs, panic recovery, rate limiting, session
// loading, maintenance mode, and trusted-proxy header handling.
package middleware

import (
	"net/http"

	chi "github.com/go-chi/chi/v5/middleware"
)

// trustedProxyHeaderNames is the set of request headers that are honored only
// when the immediate TCP peer is in TRUSTED_PROXIES. TrustedProxy strips these
// headers BEFORE RealIP runs when the peer is untrusted, which prevents a
// malicious client from spoofing its source IP, scheme, or host via
// X-Forwarded-For / X-Real-IP / True-Client-IP and defeating any per-IP
// authorization or throttling that consults r.RemoteAddr downstream.
var trustedProxyHeaderNames = []string{
	"X-Forwarded-For",
	"X-Real-IP",
	"True-Client-IP",
	"X-Forwarded-Proto",
	"X-Forwarded-Host",
	"X-Forwarded-Port",
}

// RealIP is a middleware that sets a http.Request's RemoteAddr to the results of parsing either the True-Client-IP,
//
//	X-Real-IP or the X-Forwarded-For headers (in that order).
//
// This middleware should be inserted fairly early in the middleware stack to ensure that subsequent layers (e.g.,
// request loggers) which examine the RemoteAddr will see the intended value.
// You should only use this middleware if you can trust the headers passed to you (in particular, the three headers
// this middleware uses), for example because you have placed a reverse proxy like HAProxy or nginx in front of chi.
// If your reverse proxies are configured to pass along arbitrary header values from the client, or if you use this
// middleware without a reverse proxy, malicious clients will be able to make you very sad (or, depending on how
// you're using RemoteAddr, vulnerable to an attack of some sort).
func RealIP() func(h http.Handler) http.Handler {
	return chi.RealIP
}

// RequestID is a middleware that injects a request ID into the context of each request. A request ID is a string of
// the form "host.example.com/random-0001", where "random" is a base62 random string that uniquely identifies this go
// process, and where the last number is an atomically incremented request counter.
func RequestID() func(h http.Handler) http.Handler {
	return chi.RequestID
}

// Recoverer is a middleware that recovers from panics, logs the panic (and a backtrace), and returns a HTTP 500
// (Internal Server Error) status if possible. Recoverer prints a request ID if one is provided.
func Recoverer() func(h http.Handler) http.Handler {
	return chi.Recoverer
}
