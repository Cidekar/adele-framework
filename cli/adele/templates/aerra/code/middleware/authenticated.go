//go:build aerra_template

package middleware

import (
	"net/http"
	"strings"
)

func (m *Middleware) AuthenticatedGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.App.Auth.Check(r) {
			// JSON callers (Vue SPA fetch) need a hard 401 so the client can
			// route to /login itself. A 307 redirect would auto-follow into
			// the public /login GET and look like success to fetch().
			if strings.Contains(r.Header.Get("Accept"), "application/json") ||
				strings.Contains(r.Header.Get("Content-Type"), "application/json") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"ok":false,"message":"unauthorized"}`))
				return
			}
			r.URL.Path = "/login"
			m.App.Session.Put(r.Context(), "error", "Sorry, you are not authorized to access this page.")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
