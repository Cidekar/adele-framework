package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"time"

	up "github.com/upper/db/v4"
	"golang.org/x/crypto/bcrypt"
)

// Check if a user is logged in.
func (a *Auth) Check(r *http.Request) bool {
	uid := a.Session.Get(r.Context(), "userID")
	if uid != nil {
		return true
	}

	return false
}

// Log the current user out.
func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) (bool, error) {

	// Delete remember token if exists
	if a.Session.Exists(r.Context(), "remember_token") {
		token := a.Session.GetString(r.Context(), "remember_token")
		session := a.DB.NewSession()
		collection := session.Collection("remember_tokens")
		res := collection.Find(up.Cond{"remember_token": token})
		err := res.Delete()
		if err != nil {
			return false, RememberTokenDeleteError
		}
	}

	var appName string
	value, ok := os.LookupEnv("APP_NAME")
	if !ok {
		appName = "adele"
	} else {
		appName = value
	}

	// Delete the cookie
	cookie := http.Cookie{
		Name:     fmt.Sprintf("%s_remember", appName),
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   a.Session.Cookie.Domain,
		MaxAge:   -1,
		Secure:   a.Session.Cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)

	// Log the user out of the system
	a.Session.RenewToken(r.Context())
	a.Session.Remove(r.Context(), "userID")
	a.Session.Remove(r.Context(), "remember_token")
	a.Session.Destroy(r.Context())
	a.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/login", http.StatusSeeOther)

	return true, nil
}

// Log a user in.
func (a *Auth) Login(w http.ResponseWriter, r *http.Request, email, password string) (bool, error) {

	// look up the current user
	var user User
	session := a.DB.NewSession()
	collection := session.Collection("users")
	res := collection.Find(up.Cond{"email =": email})
	err := res.One(&user)

	if err != nil {

		if err == up.ErrNoMoreRows {
			return false, nil
		}
		return false, err
	}

	if reflect.DeepEqual(user, User{}) {
		return false, InvalidPasswordOrUserError
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, InvalidPasswordOrUserError
		default:
			return false, err
		}
	}

	// create and store a remember token
	randomString := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321_+"
	s, rn := make([]rune, 12), []rune(randomString)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(rn))
		x, y := p.Uint64(), uint64(len(rn))
		s[i] = rn[x%y]
	}

	hasher := sha256.New()
	_, err = hasher.Write([]byte(string(s)))
	if err != nil {
		return false, RememberTokenHashError
	}

	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	collection = session.Collection("remember_tokens")
	rememberToken := RememberToken{
		UserID:        user.ID,
		RememberToken: sha,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_, err = collection.Insert(rememberToken)
	if err != nil {
		return false, RememberTokenStoreError
	}

	var appName string
	value, ok := os.LookupEnv("APP_NAME")
	if !ok {
		appName = "adele"
	} else {
		appName = value
	}

	// Set the cookie using the session configuration
	expiry := time.Now().Add(365 * 24 * 60 * 60 * time.Second)
	cookie := http.Cookie{
		Name:     fmt.Sprintf("%s_remember", appName),
		Value:    fmt.Sprintf("%d|%s", user.ID, sha),
		Path:     "/",
		Expires:  expiry,
		HttpOnly: true,
		Domain:   a.Session.Cookie.Domain,
		MaxAge:   315350000,
		Secure:   a.Session.Cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)

	// add remember token to session
	a.Session.Put(r.Context(), "remember_token", sha)

	// add user id to session
	a.Session.Put(r.Context(), "userID", user.ID)

	return true, nil
}

// Get the current authenticated user
func (a *Auth) User(r *http.Request) *User {

	uid := a.Session.Get(r.Context(), "userID")
	if uid != nil {
		var theUser User
		session := a.DB.NewSession()
		collection := session.Collection("users")
		res := collection.Find(up.Cond{"id =": uid})

		err := res.One(&theUser)
		if err != nil {
			return nil
		}
		return &theUser
	}

	return nil
}

// hash a plain text password for secure use in the application
func HashPassword(password string) (*[]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}
	return &hash, nil
}
