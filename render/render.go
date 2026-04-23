package render

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/CloudyKit/jet/v6"
	"github.com/justinas/nosurf"
)

func (a *Render) defaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Secure = a.Secure
	td.ServerName = a.ServerName
	td.Port = a.Port
	td.CSRFToken = nosurf.Token(r)

	if a.Session.Exists(r.Context(), "userID") {
		td.IsAuthenticated = true
	}

	td.Error = a.Session.PopString(r.Context(), "error")
	td.Flash = a.Session.PopString(r.Context(), "flash")
	return td
}

func (a *Render) Page(w http.ResponseWriter, r *http.Request, view string, variables, data interface{}) error {
	switch strings.ToLower(a.Renderer) {
	case "go":
		return a.GoPage(w, r, view, data)
	case "jet":
		return a.JetPage(w, r, view, variables, data)
	default:
	}
	return errors.New("no rendering engine provided")
}

// Render with Inertia
func (a *Render) InertiaPage(w http.ResponseWriter, r *http.Request, template string) error {

	csrfToken := nosurf.Token(r)
	ctx := a.InertiaManager.WithViewData(r.Context(), "csrf", csrfToken)
	r = r.WithContext(ctx)

	flash := a.Session.Pop(r.Context(), "flash")

	err := a.InertiaManager.Render(w, r, template, map[string]interface{}{
		"flash": flash,
		"csrf":  csrfToken,
	})
	if err != nil {
		log.Printf("%s", err)
		return err
	}

	return nil
}

// Render with standard go templates
func (a *Render) GoPage(w http.ResponseWriter, r *http.Request, view string, data interface{}) error {
	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/%s/%s.page.tmpl", a.RootPath, a.Directory, view))
	if err != nil {
		return err
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData) // cast it
	}

	err = tmpl.Execute(w, &td)
	if err != nil {
		return err
	}

	return nil
}

// Render with Jet templates
func (a *Render) JetPage(w http.ResponseWriter, r *http.Request, templateName string, variables, data interface{}) error {
	// To render templates Jet needs this to pass data to the templates
	var vars jet.VarMap

	// Convert the vars and data into the right format
	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap) // cast it
	}

	td := &TemplateData{}
	if data != nil {
		td = data.(*TemplateData) // cast it, again
	}

	td = a.defaultData(td, r) // Add default data

	t, err := a.JetViews.GetTemplate(fmt.Sprintf("%s.jet", templateName))

	if err != nil {
		return err
	}

	if err = t.Execute(w, vars, td); err != nil {
		return err
	}

	return nil
}
