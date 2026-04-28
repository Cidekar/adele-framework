//go:build aerra_template

package handlers

import (
	"fmt"
	"$APPNAME$/models"
	"net/http"
	"strings"

	"github.com/cidekar/adele-framework/auth"
	"github.com/cidekar/adele-framework/mailer"
	"github.com/cidekar/adele-framework/urlsigner"
	"github.com/CloudyKit/jet/v6"
)

/*
|--------------------------------------------------------------------------
| Handlers
|--------------------------------------------------------------------------
|
| Here is where you can add your handlers for the application. These
| handlers are called from your routes.go files.
|
*/

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	if wantsJSON(r) {
		user := h.App.Auth.User(r)
		if user == nil {
			h.respondJSON(w, http.StatusUnauthorized, map[string]any{
				"ok":      false,
				"message": "unauthorized",
			})
			return
		}
		h.respondJSON(w, http.StatusOK, map[string]any{
			"ok":   true,
			"user": user,
		})
		return
	}
	h.render(w, r, "/dashboard/home", nil, nil)
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if wantsJSON(r) {
		h.respondJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	err := h.render(w, r, "login", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) LoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusBadRequest, map[string]any{
				"ok":      false,
				"message": "Unable to process your request at this time. Please try again later.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "Unable to process your request at this time. Please try again later.")
		return
	}

	validator := h.App.Validator(nil)
	validator.NotEmpty("email", r.Form.Get("email"), "The email field is required.")
	validator.NotEmpty("password", r.Form.Get("password"), "The password field is required.")

	if r.Form.Get("email") != "" {
		validator.IsEmail("email", r.Form.Get("email"))
	}

	if !validator.Valid() {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"ok":     false,
				"errors": validator.Errors,
			})
			return
		}
		vars := make(jet.VarMap)
		vars.Set("validatorBag", validator)
		h.render(w, r, "login", vars, nil)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")
	ok, err := h.App.Auth.Login(w, r, email, password)

	if err != nil && err != auth.InvalidPasswordOrUserError {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusInternalServerError, map[string]any{
				"ok":      false,
				"message": "Unable to process your request at this time. Please try again later.",
			})
			h.App.ErrorLog.Println(err)
			return
		}
		h.App.Session.Put(r.Context(), "error", "Unable to process your request at this time. Please try again later.")
		h.App.ErrorLog.Println(err)
		h.render(w, r, "login", nil, nil)
		return
	}

	if !ok {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusUnauthorized, map[string]any{
				"ok":      false,
				"message": "Sorry, the username or password you entered is incorrect. Please try again.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "Sorry, the username or password you entered is incorrect. Please try again.")
		h.render(w, r, "login", nil, nil)
		return
	}

	if wantsJSON(r) {
		h.respondJSON(w, http.StatusOK, map[string]any{
			"ok":       true,
			"redirect": "/dashboard/home",
			"message":  "Success! You are logged into the application.",
		})
		return
	}
	h.App.Session.Put(r.Context(), "flash", "Success! You are logged into the application.")
	http.Redirect(w, r, "/dashboard/home", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	_, err := h.App.Auth.Logout(w, r)
	if err != nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusInternalServerError, map[string]any{
				"ok":      false,
				"message": "Unable to process your request at this time. Please try again later.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "Unable to process your request at this time. Please try again later.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if wantsJSON(r) {
		h.respondJSON(w, http.StatusOK, map[string]any{
			"ok":       true,
			"redirect": "/login",
			"message":  "You have been logged out.",
		})
		return
	}
	h.App.Session.Put(r.Context(), "flash", "You have been logged out.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handlers) Forgot(w http.ResponseWriter, r *http.Request) {
	if wantsJSON(r) {
		h.respondJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	err := h.render(w, r, "forgot", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) ForgotPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		h.App.ErrorLog.Println("error parsing for:", err)
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusBadRequest, map[string]any{
				"ok":      false,
				"message": "Unable to process your request at this time. Please try again later.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "Unable to process your request at this time. Please try again later.")
		http.Redirect(w, r, "/forgot", http.StatusBadRequest)
		return
	}

	email := r.Form.Get("email")

	var u *models.User
	u, err = u.GetByEmail(email)
	if err != nil {
		h.App.ErrorLog.Println("error fetching user from database:", err)
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusInternalServerError, map[string]any{
				"ok":      false,
				"message": "Unable to process your request at this time. Please try again later.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "Unable to process your request at this time. Please try again later.")
		http.Redirect(w, r, "/forgot", http.StatusSeeOther)
		return
	}

	if u == nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusOK, map[string]any{
				"ok":      true,
				"message": "If an account with that email exists, a reset link has been sent.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "flash", "A password reset link was sent to your email address.")
		http.Redirect(w, r, "/forgot", http.StatusSeeOther)
		return
	}

	link := fmt.Sprintf("%s/reset-password?email=%s", h.App.Server.URL, email)

	sign := urlsigner.Signer{
		Secret: []byte(h.App.EncryptionKey),
	}

	signedLink := sign.GenerateTokenFromString(link)

	var data struct {
		Link string
	}
	data.Link = signedLink

	msg := mailer.Message{
		To:       u.Email,
		Subject:  "Password Reset",
		Template: "password-reset",
		Data:     data,
	}

	h.App.Mail.Jobs <- msg
	res := <-h.App.Mail.Results
	if res.Error != nil {
		h.App.ErrorLog.Println("error adding mail to queue:", res.Error)
	}

	if wantsJSON(r) {
		h.respondJSON(w, http.StatusOK, map[string]any{
			"ok":      true,
			"message": "If an account with that email exists, a reset link has been sent.",
		})
		return
	}
	h.App.Session.Put(r.Context(), "flash", "A password reset link was sent to your email address.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handlers) Registration(w http.ResponseWriter, r *http.Request) {
	if wantsJSON(r) {
		h.respondJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	err := h.render(w, r, "registration", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("error rendering:", err)
	}
}

func (h *Handlers) RegistrationPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusBadRequest, map[string]any{
				"ok":      false,
				"message": "Please try again later.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "Please try again later.")
		return
	}

	validator := h.App.Validator(nil)
	validator.NotEmpty("name", r.Form.Get("name"), "The name field is required.")
	validator.NotEmpty("email", r.Form.Get("email"), "The email field is required.")
	validator.NotEmpty("password", r.Form.Get("password"), "The password field is required.")
	validator.NotEmpty("verify-password", r.Form.Get("verify-password"), "The password confirmation field is required.")

	if r.Form.Get("password") != r.Form.Get("verify-password") {
		validator.AddError("password", "The password does not match the password confirmation.")
		validator.AddError("verify-password", "The confirmation password does not match the password.")
	}

	validator.Password("password", r.Form.Get("password"))

	if r.Form.Get("email") != "" {
		validator.IsEmail("email", r.Form.Get("email"))
		validator.IsEmailInPublicDomain("email", r.Form.Get("email"))
	}

	user, _ := h.Models.Users.GetByEmail(r.Form.Get("email"))
	if user != nil {
		validator.AddError("email", "Please choose another email address.")
	}

	if !validator.Valid() {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"ok":     false,
				"errors": validator.Errors,
			})
			return
		}
		vars := make(jet.VarMap)
		vars.Set("validatorBag", validator)
		vars.Set("name", r.Form.Get("name"))
		if user == nil {
			vars.Set("email", r.Form.Get("email"))
		}
		h.render(w, r, "registration", vars, nil)
		return
	}

	newUser := models.User{
		Email:    r.Form.Get("email"),
		Active:   1,
		Password: r.Form.Get("password"),
	}

	parts := strings.Fields(r.Form.Get("name"))
	if len(parts) == 1 {
		newUser.FirstName = parts[0]
	} else if len(parts) >= 2 {
		newUser.FirstName = parts[0]
		newUser.LastName = parts[len(parts)-1]
	}

	_, err = h.Models.Users.Insert(newUser)
	if err != nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusInternalServerError, map[string]any{
				"ok":      false,
				"message": http.StatusText(http.StatusInternalServerError),
			})
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	if wantsJSON(r) {
		h.respondJSON(w, http.StatusOK, map[string]any{
			"ok":       true,
			"redirect": "/login",
			"message":  "Registration complete - please check your email and login.",
		})
		return
	}
	h.App.Session.Put(r.Context(), "flash", "Registration complete - please check your email and login.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// NotFound is intentionally left in the user's existing handlers.go; this file
// does not duplicate it. The baseline NotFound renders the "404" template that
// ships with `adele new`.

func (h *Handlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	theURL := r.RequestURI
	testURL := fmt.Sprintf("%s%s", h.App.Server.URL, theURL)

	signer := urlsigner.Signer{
		Secret: []byte(h.App.EncryptionKey),
	}

	valid := signer.VerifyToken(testURL)
	if !valid {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusBadRequest, map[string]any{
				"ok":      false,
				"message": "Invalid or expired token",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "The password reset link is invalid. Please request a new one and try again.")
		h.render(w, r, "/forgot", nil, nil)
		return
	}

	expired := signer.Expired(testURL, 60)
	if expired {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusBadRequest, map[string]any{
				"ok":      false,
				"message": "Invalid or expired token",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "The password reset link expired, please request a new one.")
		h.render(w, r, "/forgot", nil, nil)
		return
	}

	if wantsJSON(r) {
		h.respondJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	encryptedEmail, _ := h.encrypt(email)

	vars := make(jet.VarMap)
	vars.Set("email", encryptedEmail)

	h.render(w, r, "reset-password", vars, nil)
}

func (h *Handlers) ResetPasswordPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusBadRequest, map[string]any{
				"ok":      false,
				"message": "Please request a new password reset and try again.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "Please request a new password reset and try again.")
		return
	}

	validator := h.App.Validator(nil)

	validator.NotEmpty("password", r.Form.Get("password"), "The password field is required.")
	validator.NotEmpty("verify-password", r.Form.Get("verify-password"), "The password confirmation field is required.")

	if r.Form.Get("password") != r.Form.Get("verify-password") {
		validator.AddError("password", "The password does not match the password confirmation.")
		validator.AddError("verify-password", "The confirmation password does not match the password.")
	}

	validator.Password("password", r.Form.Get("password"))

	if !validator.Valid() {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"ok":     false,
				"errors": validator.Errors,
			})
			return
		}
		vars := make(jet.VarMap)
		vars.Set("validatorBag", validator)

		vars.Set("email", r.Form.Get("email"))
		h.render(w, r, "reset-password", vars, nil)
		return
	}

	email, err := h.decrypt(r.Form.Get("email"))
	if err != nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusBadRequest, map[string]any{
				"ok":      false,
				"message": "The password reset link is invalid. Please request a new one.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "The password reset link is invalid. Please request a new one.")
		http.Redirect(w, r, "/forgot", http.StatusBadRequest)
		return
	}

	var user *models.User
	user, err = user.GetByEmail(email)
	if err != nil {
		h.App.ErrorLog.Println("error fetching user from database:", err)
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusInternalServerError, map[string]any{
				"ok":      false,
				"message": "Unable to process your request at this time. Please try again later.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "Unable to process your request at this time. Please try again later.")
		http.Redirect(w, r, "/forgot", http.StatusSeeOther)
		return
	}

	if user == nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusInternalServerError, map[string]any{
				"ok":      false,
				"message": "Unable to process your request at this time. Please try again later.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "Unable to process your request at this time. Please try again later.")
		http.Redirect(w, r, "/forgot", http.StatusSeeOther)
		return
	}

	err = user.ResetPassword(user.ID, r.Form.Get("password"))
	if err != nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusInternalServerError, map[string]any{
				"ok":      false,
				"message": "The password reset has failed. Please request a new password reset and try again.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "The password reset has failed. Please request a new password reset and try again.")
		return
	}

	if wantsJSON(r) {
		h.respondJSON(w, http.StatusOK, map[string]any{
			"ok":       true,
			"redirect": "/login",
			"message":  "Password reset complete - please login.",
		})
		return
	}
	h.App.Session.Put(r.Context(), "flash", "Password reset complete - please login.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	usr := models.User{
		FirstName: "Harrison",
		LastName:  "DeStefano",
		Email:     "harrison@cidekar.com",
		Password:  "123456",
	}

	_, err := h.Models.Users.Insert(usr)
	if err != nil {
		h.App.ErrorLog.Println("error creating user:", err)
	}

	h.render(w, r, "home", nil, nil)
}

func (h *Handlers) Profile(w http.ResponseWriter, r *http.Request) {
	if wantsJSON(r) {
		user := h.App.Auth.User(r)
		if user == nil {
			h.respondJSON(w, http.StatusUnauthorized, map[string]any{
				"ok":      false,
				"message": "unauthorized",
			})
			return
		}
		h.respondJSON(w, http.StatusOK, map[string]any{
			"ok":   true,
			"user": user,
		})
		return
	}
	h.render(w, r, "/dashboard/profile", nil, nil)
}

func (h *Handlers) ProfilePost(w http.ResponseWriter, r *http.Request) {

	// The user needs to be set or Jet cant render our template
	// Maybe we want to add this to a middleware so the user is auto added?
	user := h.App.Auth.User(r)
	vars := make(jet.VarMap)
	vars.Set("user", &user)

	err := r.ParseForm()
	if err != nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusBadRequest, map[string]any{
				"ok":      false,
				"message": "Please try again later.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "Please try again later.")
		h.render(w, r, r.URL.Path, vars, nil)
	}

	validator := h.App.Validator(nil)
	validator.NotEmpty("name", r.Form.Get("name"), "The name field is required.")
	validator.NotEmpty("email", r.Form.Get("email"), "The email field is required.")

	if r.Form.Get("email") != user.Email {

		if r.Form.Get("email") != "" {
			validator.IsEmail("email", r.Form.Get("email"))
			validator.IsEmailInPublicDomain("email", r.Form.Get("email"))
		}

		u, _ := h.Models.Users.GetByEmail(r.Form.Get("email"))

		if u != nil {
			validator.AddError("email", "Please choose another email address.")
		}
	}

	if !validator.Valid() {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"ok":     false,
				"errors": validator.Errors,
			})
			return
		}
		vars.Set("validatorBag", validator)
		h.render(w, r, r.URL.Path, vars, nil)
	}

	user.Email = r.Form.Get("email")

	parts := strings.Fields(r.Form.Get("name"))
	if len(parts) == 1 {
		user.FirstName = parts[0]
	} else if len(parts) >= 2 {
		user.FirstName = parts[0]
		user.LastName = parts[len(parts)-1]
	}

	fmt.Println(user)

	theuser := models.User{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Password:  user.Password,
	}

	err = h.Models.Users.Update(theuser)

	if err != nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusInternalServerError, map[string]any{
				"ok":      false,
				"message": "Unable to process your request at this time. Please try again later.",
			})
			h.App.ErrorLog.Println(err)
			return
		}
		h.App.Session.Put(r.Context(), "error", "Unable to process your request at this time. Please try again later.")
		h.App.ErrorLog.Println(err)
		h.render(w, r, r.URL.Path, vars, nil)
		return
	}

	if wantsJSON(r) {
		h.respondJSON(w, http.StatusOK, map[string]any{
			"ok":      true,
			"user":    theuser,
			"message": "Success! Your profile information was updated.",
		})
		return
	}
	h.App.Session.Put(r.Context(), "flash", "Success! Your profile information was updated.")
	h.render(w, r, r.URL.Path, vars, nil)

}

func (h *Handlers) ProfilePasswordPost(w http.ResponseWriter, r *http.Request) {

	user := h.App.Auth.User(r)
	password := r.Form.Get("current-password")
	ok, err := h.App.Auth.Login(w, r, user.Email, password)

	if !ok || err != nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusUnauthorized, map[string]any{
				"ok":      false,
				"message": "Sorry, the password you entered is incorrect. If you do not know your password, please logout and use our password reset.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "Sorry, the password you entered is incorrect. If you do not know your password, please logout and use our password reset.")
		h.render(w, r, "login", nil, nil)
		return
	}

	err = r.ParseForm()
	if err != nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusBadRequest, map[string]any{
				"ok":      false,
				"message": "We are having trouble processing your request. Please complete the form and try again.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "We are having trouble processing your request. Please complete the form and try again.")
		return
	}

	validator := h.App.Validator(nil)
	validator.NotEmpty("password", r.Form.Get("password"), "The password field is required.")
	validator.NotEmpty("verify-password", r.Form.Get("verify-password"), "The password confirmation field is required.")

	if r.Form.Get("password") != r.Form.Get("verify-password") {
		validator.AddError("password", "The password does not match the password confirmation.")
		validator.AddError("verify-password", "The confirmation password does not match the password.")
	}

	validator.Password("password", r.Form.Get("password"))

	if !validator.Valid() {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"ok":     false,
				"errors": validator.Errors,
			})
			return
		}
		vars := make(jet.VarMap)
		vars.Set("validatorBag", validator)

		h.render(w, r, "/profile", vars, nil)
		return
	}

	var u *models.User
	u, err = u.GetByEmail(user.Email)
	if err != nil {
		h.App.ErrorLog.Println("error fetching user from database:", err)
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusInternalServerError, map[string]any{
				"ok":      false,
				"message": "Unable to process your request at this time. Please try again later.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "Unable to process your request at this time. Please try again later.")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	if u == nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusInternalServerError, map[string]any{
				"ok":      false,
				"message": "Unable to process your request at this time. Please try again later.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "Unable to process your request at this time. Please try again later.")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	err = u.ResetPassword(user.ID, r.Form.Get("password"))

	if err != nil {
		if wantsJSON(r) {
			h.respondJSON(w, http.StatusInternalServerError, map[string]any{
				"ok":      false,
				"message": "The password reset has failed. Please complete the form and try again.",
			})
			return
		}
		h.App.Session.Put(r.Context(), "error", "The password reset has failed. Please complete the form and try again.")
		return
	}

	if wantsJSON(r) {
		h.respondJSON(w, http.StatusOK, map[string]any{
			"ok":       true,
			"redirect": "/login",
			"message":  "Password reset complete - please login.",
		})
		return
	}
	h.App.Session.Put(r.Context(), "flash", "Password reset complete - please login.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)

}
