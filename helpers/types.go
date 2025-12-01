package helpers

import "github.com/cidekar/adele-framework/render"

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
