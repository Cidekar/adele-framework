//go:build aerra_template

package main

import (
	"net/http"

	"github.com/cidekar/adele-framework/mux"
	"github.com/go-chi/chi/v5"
)

func (a *application) WebRoutes() http.Handler {

	r := mux.NewRouter()

	r.Use(a.Middleware.NoSurf)

	r.Use(a.Middleware.CheckRemember)

	// Public routes
	r.Group(func(mux chi.Router) {

		r.Get("/", a.Handlers.Home)

		r.Get("/login", a.Handlers.Login)

		r.Post("/login", a.Handlers.LoginPost)

		r.Get("/logout", a.Handlers.Logout)

		r.Get("/forgot", a.Handlers.Forgot)

		r.Post("/forgot", a.Handlers.ForgotPost)

		r.Get("/registration", a.Handlers.Registration)

		r.Post("/registration", a.Handlers.RegistrationPost)

		r.Get("/reset-password", a.Handlers.ResetPassword)

		r.Post("/reset-password", a.Handlers.ResetPasswordPost)

		r.Get("/user", a.Handlers.CreateUser)

		r.NotFound(a.Handlers.NotFound)
	})

	// Private routes
	privateRoutes := mux.NewRouter()

	privateRoutes.Use(a.Middleware.AuthenticatedGuard)

	privateRoutes.Group(func(mux chi.Router) {

		privateRoutes.Get("/home", a.Handlers.Dashboard)

		privateRoutes.Get("/profile", a.Handlers.Profile)

		privateRoutes.Post("/profile", a.Handlers.ProfilePost)

		privateRoutes.Post("/profile-password", a.Handlers.ProfilePasswordPost)
	})

	a.App.Routes.Mount("/dashboard", privateRoutes)

	return r
}
