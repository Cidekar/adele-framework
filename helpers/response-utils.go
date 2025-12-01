package helpers

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
)

// ReadJSON reads and decodes a JSON request body into the provided data structure.
// It limits the request body size to prevent memory exhaustion attacks.
// The maxBytes parameter is optional; if not provided, it defaults to 1048576 (1 MB).
// Returns an error if decoding fails or if the body contains more than one JSON value.
//
// Example:
//
//	type CreateUserRequest struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	func (a *App) CreateUser(w http.ResponseWriter, r *http.Request) {
//	    var req CreateUserRequest
//	    // Use default 1MB limit
//	    if err := a.Helpers.ReadJSON(w, r, &req); err != nil {
//	        a.Helpers.JsonError(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
//	        return
//	    }
//	    // Or specify custom limit (e.g., 512KB)
//	    if err := a.Helpers.ReadJSON(w, r, &req, 524288); err != nil {
//	        a.Helpers.JsonError(w, map[string]string{"error": err.Error()}, http.StatusBadRequest)
//	        return
//	    }
//	}
func (h *Helpers) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}, maxBytes ...int) error {
	limit := 1048576 // default 1 MB limit for JSON payload
	if len(maxBytes) > 0 {
		limit = maxBytes[0]
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(limit))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only have a single json value")
	}

	return nil
}

// WriteJSON encodes data as JSON and writes it to the response with the specified status code.
// The response is pretty-printed with indentation for readability.
// Optional headers can be provided to add custom HTTP headers to the response.
//
// Example:
//
//	type UserResponse struct {
//	    ID    int    `json:"id"`
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	func (a *App) GetUser(w http.ResponseWriter, r *http.Request) {
//	    user := UserResponse{ID: 1, Name: "John", Email: "john@example.com"}
//	    // Basic usage
//	    a.Helpers.WriteJSON(w, http.StatusOK, user)
//	    // With custom headers
//	    headers := http.Header{"X-Custom-Header": []string{"value"}}
//	    a.Helpers.WriteJSON(w, http.StatusOK, user, headers)
//	}
func (h *Helpers) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

// JsonError writes an error response in JSON format following RFC 7807 (Problem Details).
// Sets the Content-Type to "application/problem+json" and includes security headers.
//
// Example:
//
//	func (a *App) CreateUser(w http.ResponseWriter, r *http.Request) {
//	    // Validation error
//	    a.Helpers.JsonError(w, map[string]string{
//	        "error": "validation failed",
//	        "field": "email",
//	    }, http.StatusBadRequest)
//
//	    // Or with a custom error struct
//	    type APIError struct {
//	        Code    string `json:"code"`
//	        Message string `json:"message"`
//	    }
//	    a.Helpers.JsonError(w, APIError{Code: "USER_EXISTS", Message: "User already exists"}, http.StatusConflict)
//	}
func (h *Helpers) JsonError(w http.ResponseWriter, err interface{}, status int) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(err)
}

// WriteXML encodes data as XML and writes it to the response with the specified status code.
// The response is pretty-printed with indentation for readability.
// Optional headers can be provided to add custom HTTP headers to the response.
//
// Example:
//
//	type ProductResponse struct {
//	    XMLName xml.Name `xml:"product"`
//	    ID      int      `xml:"id"`
//	    Name    string   `xml:"name"`
//	    Price   float64  `xml:"price"`
//	}
//
//	func (a *App) GetProduct(w http.ResponseWriter, r *http.Request) {
//	    product := ProductResponse{ID: 1, Name: "Widget", Price: 29.99}
//	    a.Helpers.WriteXML(w, http.StatusOK, product)
//	}
func (h *Helpers) WriteXML(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := xml.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

// DownloadFile serves a file as a downloadable attachment to the client.
// It sets the Content-Disposition header to trigger a browser download dialog.
// The file path is sanitized using filepath.Clean to prevent directory traversal attacks.
// Returns the cleaned file path and any error that occurred.
//
// Example:
//
//	func (a *App) DownloadReport(w http.ResponseWriter, r *http.Request) {
//	    filePath, err := a.Helpers.DownloadFile(w, r, "/var/reports", "monthly-report.pdf")
//	    if err != nil {
//	        a.Helpers.Error500(w, r)
//	        return
//	    }
//	    log.Printf("Served file: %s", filePath)
//	}
func (h *Helpers) DownloadFile(w http.ResponseWriter, r *http.Request, pathToFile, fileName string) (string, error) {
	fp := path.Join(pathToFile, fileName)

	// clean path up
	fileToServe := filepath.Clean(fp)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	http.ServeFile(w, r, fileToServe)
	return fileToServe, nil
}

// Error404 sends a 404 Not Found response to the client.
//
// Example:
//
//	func (a *App) GetUser(w http.ResponseWriter, r *http.Request) {
//	    user, err := a.DB.FindUser(id)
//	    if err != nil {
//	        a.Helpers.Error404(w, r)
//	        return
//	    }
//	}
func (h *Helpers) Error404(w http.ResponseWriter, r *http.Request) {
	h.ErrorStatus(w, http.StatusNotFound)
}

// Error500 sends a 500 Internal Server Error response to the client.
//
// Example:
//
//	func (a *App) ProcessData(w http.ResponseWriter, r *http.Request) {
//	    if err := a.Service.Process(); err != nil {
//	        log.Printf("Processing failed: %v", err)
//	        a.Helpers.Error500(w, r)
//	        return
//	    }
//	}
func (h *Helpers) Error500(w http.ResponseWriter, r *http.Request) {
	h.ErrorStatus(w, http.StatusInternalServerError)
}

// ErrorUnauthorized sends a 401 Unauthorized response to the client.
// Use this when authentication is required but not provided or invalid.
//
// Example:
//
//	func (a *App) ProtectedEndpoint(w http.ResponseWriter, r *http.Request) {
//	    token := r.Header.Get("Authorization")
//	    if token == "" || !a.Auth.ValidateToken(token) {
//	        a.Helpers.ErrorUnauthorized(w, r)
//	        return
//	    }
//	}
func (h *Helpers) ErrorUnauthorized(w http.ResponseWriter, r *http.Request) {
	h.ErrorStatus(w, http.StatusUnauthorized)
}

// ErrorForbidden sends a 403 Forbidden response to the client.
// Use this when the user is authenticated but lacks permission for the requested resource.
//
// Example:
//
//	func (a *App) AdminEndpoint(w http.ResponseWriter, r *http.Request) {
//	    user := a.Auth.GetUser(r)
//	    if !user.IsAdmin {
//	        a.Helpers.ErrorForbidden(w, r)
//	        return
//	    }
//	}
func (h *Helpers) ErrorForbidden(w http.ResponseWriter, r *http.Request) {
	h.ErrorStatus(w, http.StatusForbidden)
}

// ErrorStatus sends an error response with the specified HTTP status code.
// The response body contains the standard HTTP status text for the given code.
//
// Example:
//
//	func (a *App) CustomError(w http.ResponseWriter, r *http.Request) {
//	    // Send a 429 Too Many Requests
//	    a.Helpers.ErrorStatus(w, http.StatusTooManyRequests)
//	    // Send a 503 Service Unavailable
//	    a.Helpers.ErrorStatus(w, http.StatusServiceUnavailable)
//	}
func (h *Helpers) ErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
