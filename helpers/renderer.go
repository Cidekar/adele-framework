package helpers

import (
	"net/http"

	"github.com/CloudyKit/jet/v6"
)

// Render renders the named Jet template to the response writer.
// The variables parameter, when non-nil, must be a jet.VarMap and is used as the
// template variables; when nil, an empty jet.VarMap is created. Render always sets
// the "view" variable to the template name and the "path" variable to the request's
// URL path before delegating to the underlying renderer's Page method. The data
// parameter is passed through to the template as its execution context.
// It returns any error produced while rendering the page.
//
// Render panics if variables is non-nil but not of type jet.VarMap.
func (h *Helpers) Render(w http.ResponseWriter, r *http.Request, template string, variables, data interface{}) error {
	vars := make(jet.VarMap)
	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap)
	}

	vars.Set("view", template)
	vars.Set("path", r.URL.Path)

	return h.Redner.Page(w, r, template, vars, data)
}
