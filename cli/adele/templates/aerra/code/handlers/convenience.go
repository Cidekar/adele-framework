//go:build aerra_template

package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cidekar/adele-framework/helpers"
	"github.com/CloudyKit/jet/v6"
)

func (h *Handlers) render(w http.ResponseWriter, r *http.Request, template string, variables, data interface{}) error {
	vars := make(jet.VarMap)
	if variables == nil {
		vars = make(jet.VarMap)
	} else {
		vars = variables.(jet.VarMap)
	}

	vars.Set("view", template)
	vars.Set("path", r.URL.Path)

	return h.App.Render.Page(w, r, template, vars, data)
}

func (h *Handlers) encrypt(text string) (string, error) {
	enc := helpers.Encryption{Key: []byte(h.App.EncryptionKey)}

	encrypted, err := enc.Encrypt(text)
	if err != nil {
		return "", err
	}
	return encrypted, nil
}

func (h *Handlers) decrypt(crypto string) (string, error) {
	enc := helpers.Encryption{Key: []byte(h.App.EncryptionKey)}

	decrypted, err := enc.Decrypt(crypto)
	if err != nil {
		return "", err
	}
	return decrypted, nil
}

// wantsJSON reports whether the caller requested a JSON response. We check
// both Accept and Content-Type because a JSON request body (SPA fetch with a
// JSON payload) generally expects a JSON response back, even if the client
// forgot to set an explicit Accept header.
func wantsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		return true
	}
	ct := r.Header.Get("Content-Type")
	return strings.Contains(ct, "application/json")
}

// respondJSON writes a JSON body with the given HTTP status code. The body
// shape is a flat map; callers pass {"ok": true, ...} on success or
// {"ok": false, "errors": {...}} / {"ok": false, "message": "..."} on failure.
func (h *Handlers) respondJSON(w http.ResponseWriter, status int, body map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		h.App.ErrorLog.Println("respondJSON encode:", err)
	}
}

