// Package session configures and initializes a scs session manager from
// cookie settings such as lifetime, persistence, security, name, and domain.
package session

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
)

func (s *Session) InitSession() *scs.SessionManager {

	// exisiting code from previous function
	var persist, secure bool

	// sessions last for?
	minutes, err := strconv.Atoi(s.CookieLifetime)
	if err != nil {
		minutes = 60
	}

	// do cookies persist?
	if strings.ToLower(s.CookiePersist) == "true" {
		persist = true
	} else {
		persist = false
	}

	// cookies are secure?
	if strings.ToLower(s.CookieSecure) == "true" {
		secure = true
	} else {
		secure = false
	}

	// create our session
	session := scs.New()
	session.Lifetime = time.Duration(minutes) * time.Minute
	session.Cookie.Persist = persist
	session.Cookie.Name = s.CookieName
	session.Cookie.Secure = secure
	session.Cookie.Domain = s.CookieDomain
	session.Cookie.SameSite = http.SameSiteLaxMode

	// Use cookie for sessions
	// TODO: allow for other session store types

	return session
}
