// Package helpers provides common utility functions for Adele applications.
//
// It groups string generation, input sanitization, request validation, encryption,
// environment variable access, JSON request and response handling, template
// rendering, and file upload utilities under a single Helpers type.
package helpers

import "github.com/cidekar/adele-framework/render"

// Helpers aggregates the framework helper utilities, exposing file upload
// configuration and a reference to the render engine.
type Helpers struct {
	FileUploadConfig FileUploadConfig
	Redner           *render.Render
}

// UploadConfig holds upload configuration
type FileUploadConfig struct {
	MaxSize          int64
	AllowedMimeTypes []string
	TempDir          string
	Destination      string
}

// UploadResult contains information about uploaded file
type FileUploadResult struct {
	OriginalName string
	SavedName    string
	MimeType     string
	Size         int64
	Path         string
}
