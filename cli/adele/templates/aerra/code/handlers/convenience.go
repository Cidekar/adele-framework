//go:build aerra_template

package handlers

import (
	"net/http"

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
