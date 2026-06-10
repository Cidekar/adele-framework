package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cidekar/adele-framework/filesystem"
	"github.com/gabriel-vasile/mimetype"
)

// Get environment variable or return default if the value is an empty string.
// Example:
//
//	Helpers.Getenv("ADELE_API_ADDR", "localhost")
func (h *Helpers) Getenv(key string, defaultValue ...string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// Ensure that a file at the given path exists. If it doesn't, it attempts to create
// the file.
// Example:
//
//	Helpers.CreateFileIfNotExist(path)
func (h *Helpers) CreateFileIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

// Ensure that a specific directory exists at the given path. If the directory
// is absent, it proceeds to create it with predefined permissions. This function
// is useful in scenarios where you need to guarantee that a directory is present
// before performing operations that require its existence. A directory that is
// created will have octal value allows the owner to read, write, and execute files
// within the directory, while the group and others can only read and execute, not
// alter the content.
// Example:
//
//	Helpers.CreateDirIfNotExist(path)
func (h *Helpers) CreateDirIfNotExist(path string) error {
	const mode = 0755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, mode)
		if err != nil {
			return err
		}
	}
	return nil
}

// UploadFile safely handles multipart file uploads with validation, secure filename generation,
// and automatic cleanup. Validates MIME type and size, prevents path traversal attacks, and
// supports both local filesystem and custom storage backends.
// Example:
//
//	config := UploadConfig{
//	    MaxSize:          10 << 20, // 10MB
//	    AllowedMimeTypes: []string{"image/jpeg", "image/png"},
//	    TempDir:          "storage/tmp",
//	    Destination:      "storage/uploads",
//	}
//	result, err := app.UploadFile(r, "avatar", config, nil)
//	if err != nil {
//	    return fmt.Errorf("upload failed: %w", err)
//	}
//	log.Printf("Uploaded %s as %s", result.OriginalName, result.SavedName)
func (h *Helpers) UploadFile(r *http.Request, field string, config FileUploadConfig, fs filesystem.FS) (*FileUploadResult, error) {
	// Parse multipart form
	if err := r.ParseMultipartForm(config.MaxSize); err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// Get file from form
	file, header, err := r.FormFile(field)
	if err != nil {
		return nil, fmt.Errorf("failed to get file from field '%s': %w", field, err)
	}
	defer file.Close()

	// Validate file
	result, err := h.validateAndPrepareFile(file, header, config)
	if err != nil {
		return nil, err
	}

	// Create temporary file with safe name
	tempPath, cleanup, err := h.createTempFile(file, result.SavedName, config.TempDir)
	if err != nil {
		return nil, err
	}
	defer cleanup() // Always cleanup temp file

	// Move to final destination
	if fs != nil {
		err = fs.Put(tempPath, config.Destination)
	} else {
		finalPath := filepath.Join(config.Destination, result.SavedName)
		err = os.Rename(tempPath, finalPath)
		result.Path = finalPath
	}

	if err != nil {
		return nil, fmt.Errorf("failed to move file to destination: %w", err)
	}

	return result, nil
}

// validateAndPrepareFile validates an uploaded file against the supplied config and
// builds the metadata for the saved file. It enforces the maximum size, detects the
// MIME type by reading the file content (resetting the read pointer afterward), and
// rejects types not present in config.AllowedMimeTypes. On success it generates a safe
// filename and returns a FileUploadResult populated with the original name, generated
// name, detected MIME type, and size. It returns an error if the file exceeds the size
// limit, the MIME type cannot be detected or is disallowed, the pointer cannot be reset,
// or a safe filename cannot be generated.
func (h *Helpers) validateAndPrepareFile(file multipart.File, header *multipart.FileHeader, config FileUploadConfig) (*FileUploadResult, error) {
	// Check file size
	if header.Size > config.MaxSize {
		return nil, fmt.Errorf("file size %d exceeds maximum %d bytes", header.Size, config.MaxSize)
	}

	// Detect MIME type
	mimeType, err := mimetype.DetectReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to detect file type: %w", err)
	}

	// Reset file pointer
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Validate MIME type
	if !slices.Contains(config.AllowedMimeTypes, mimeType.String()) {
		return nil, fmt.Errorf("file type '%s' not allowed. Allowed types: %v",
			mimeType.String(), config.AllowedMimeTypes)
	}

	// Generate safe filename
	safeName, err := h.generateSafeFilename(header.Filename, mimeType.Extension())
	if err != nil {
		return nil, fmt.Errorf("failed to generate safe filename: %w", err)
	}

	return &FileUploadResult{
		OriginalName: header.Filename,
		SavedName:    safeName,
		MimeType:     mimeType.String(),
		Size:         header.Size,
	}, nil
}

// createTempFile copies the contents of src into a new file named filename inside
// tempDir, creating tempDir (with 0755 permissions) if it does not exist. When tempDir
// is empty it defaults to "./tmp". It returns the path to the temporary file and a
// cleanup function that removes that file when called. If the directory or file cannot
// be created the copy is aborted and an error is returned; if copying fails the partial
// temp file is removed before the error is returned.
func (h *Helpers) createTempFile(src multipart.File, filename, tempDir string) (string, func(), error) {
	if tempDir == "" {
		tempDir = "./tmp"
	}

	// Ensure temp directory exists
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	tempPath := filepath.Join(tempDir, filename)
	dst, err := os.Create(tempPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	// Copy file content
	_, err = io.Copy(dst, src)
	dst.Close()

	if err != nil {
		os.Remove(tempPath) // Cleanup on error
		return "", nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.Remove(tempPath)
	}

	return tempPath, cleanup, nil
}

// generateSafeFilename builds a collision-resistant, sanitized filename from originalName
// and the given extension. It strips any directory components, replaces spaces with
// underscores, removes ".." sequences to guard against path traversal, drops the original
// extension to avoid double extensions, and truncates the base name to 50 characters. A
// random 8-byte hex suffix (sourced from crypto/rand) is appended before the extension to
// ensure uniqueness. It returns an error only if reading random bytes fails.
func (h *Helpers) generateSafeFilename(originalName, extension string) (string, error) {
	// Generate random part
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	randomStr := hex.EncodeToString(randomBytes)

	// Clean original name (remove path components and dangerous characters)
	cleanName := filepath.Base(originalName)
	cleanName = strings.ReplaceAll(cleanName, " ", "_")
	cleanName = strings.ReplaceAll(cleanName, "..", "")

	// Remove extension to avoid double extensions
	cleanName = strings.TrimSuffix(cleanName, filepath.Ext(cleanName))

	// Limit length
	if len(cleanName) > 50 {
		cleanName = cleanName[:50]
	}

	return fmt.Sprintf("%s_%s%s", cleanName, randomStr, extension), nil
}
