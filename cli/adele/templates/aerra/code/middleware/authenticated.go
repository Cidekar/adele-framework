//go:build aerra_template

package middleware

import (
	"net/http"
)

func (m *Middleware) AuthenticatedGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.App.Auth.Check(r) {
			r.URL.Path = "/login"
			m.App.Session.Put(r.Context(), "error", "Sorry, you are not authorized to access this page.")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
