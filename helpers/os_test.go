package helpers

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cidekar/adele-framework/helpers/filesystem"
)

func TestCreateDirIfNotExist(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "test_dir")

	err := helpers.CreateDirIfNotExist(testPath)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Verify it exists and has correct permissions
	info, err := os.Stat(testPath)
	if err != nil {
		t.Fatal("Directory was not created")
	}
	if !info.IsDir() {
		t.Error("Created path is not a directory")
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("Wrong permissions: got %o, want 0755", info.Mode().Perm())
	}

	// Test that calling again doesn't error
	err = helpers.CreateDirIfNotExist(testPath)
	if err != nil {
		t.Fatalf("Should not error when directory exists: %v", err)
	}
}

func TestCreateDirIfNotExist_InvalidPath(t *testing.T) {

	// Test with empty path - should error
	err := helpers.CreateDirIfNotExist("")
	if err == nil {
		t.Error("Should fail with empty path")
	}
}

func TestCreateFileIfNotExist(t *testing.T) {
	helpers := &Helpers{}
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_file.txt")

	// Test creating new file
	err := helpers.CreateFileIfNotExist(testFile)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Verify file exists
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatal("File was not created")
	}
	if info.IsDir() {
		t.Error("Created a directory instead of file")
	}

	// Test that calling again doesn't error
	err = helpers.CreateFileIfNotExist(testFile)
	if err != nil {
		t.Fatalf("Should not error when file exists: %v", err)
	}
}

func TestCreateFileIfNotExist_InvalidPath(t *testing.T) {
	helpers := &Helpers{}

	// Test with invalid path (directory that doesn't exist)
	invalidPath := "/nonexistent/directory/file.txt"
	err := helpers.CreateFileIfNotExist(invalidPath)
	if err == nil {
		t.Error("Should fail when parent directory doesn't exist")
	}
}

func TestCreateFileIfNotExist_EmptyPath(t *testing.T) {
	helpers := &Helpers{}

	// Test with empty path
	err := helpers.CreateFileIfNotExist("")
	if err == nil {
		t.Error("Should fail with empty path")
	}
}

func TestHelpers_Getenv(t *testing.T) {
	helpers := &Helpers{}

	// Test 1: Environment variable exists
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	result := helpers.Getenv("TEST_VAR", "default")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}

	// Test 2: Environment variable doesn't exist, use default
	result = helpers.Getenv("NONEXISTENT_VAR", "default_value")
	if result != "default_value" {
		t.Errorf("Expected 'default_value', got '%s'", result)
	}

	// Test 3: Environment variable doesn't exist, no default
	result = helpers.Getenv("NONEXISTENT_VAR")
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}

	// Test 4: Environment variable is empty string
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")

	result = helpers.Getenv("EMPTY_VAR", "fallback")
	if result != "fallback" {
		t.Errorf("Expected 'fallback' for empty env var, got '%s'", result)
	}

	// Test 5: Multiple default values (should use first)
	result = helpers.Getenv("MISSING_VAR", "first", "second", "third")
	if result != "first" {
		t.Errorf("Expected 'first' from multiple defaults, got '%s'", result)
	}
}

func TestHelpers_Getenv_EmptyKey(t *testing.T) {
	helpers := &Helpers{}

	// Test with empty key
	result := helpers.Getenv("", "default")
	if result != "default" {
		t.Errorf("Expected 'default' for empty key, got '%s'", result)
	}
}

func createMultipartRequest(t *testing.T, fieldName, fileName, content string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = part.Write([]byte(content))
	if err != nil {
		t.Fatalf("Failed to write file content: %v", err)
	}

	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestUploadFile_Success(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	destDir := t.TempDir()

	helpers := &Helpers{}

	config := FileUploadConfig{
		MaxSize:          1024,
		AllowedMimeTypes: []string{"text/plain; charset=utf-8"},
		TempDir:          tempDir,
		Destination:      destDir,
	}

	req := createMultipartRequest(t, "file", "test.txt", "Hello World")

	// Test
	result, err := helpers.UploadFile(req, "file", config, nil)

	// Assertions
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.OriginalName != "test.txt" {
		t.Errorf("Expected original name 'test.txt', got '%s'", result.OriginalName)
	}
	if !strings.Contains(result.SavedName, "test_") {
		t.Errorf("Expected saved name to contain 'test_', got '%s'", result.SavedName)
	}
	if result.MimeType != "text/plain; charset=utf-8" {
		t.Errorf("Expected MIME type 'text/plain; charset=utf-8', got '%s'", result.MimeType)
	}

	// Check file was created
	if _, err := os.Stat(result.Path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist at %s", result.Path)
	}
}

func TestUploadFile_InvalidField(t *testing.T) {
	helpers := &Helpers{}
	config := FileUploadConfig{
		MaxSize:          1024,
		AllowedMimeTypes: []string{"text/plain; charset=utf-8"},
	}

	req := createMultipartRequest(t, "file", "test.txt", "Hello World")

	// Test with wrong field name
	result, err := helpers.UploadFile(req, "wrong_field", config, nil)

	if err == nil {
		t.Fatal("Expected error for invalid field name")
	}
	if result != nil {
		t.Error("Expected nil result on error")
	}
	if !strings.Contains(err.Error(), "wrong_field") {
		t.Errorf("Expected error to mention field name, got: %v", err)
	}
}

func TestUploadFile_FileTooLarge(t *testing.T) {
	helpers := &Helpers{}
	config := FileUploadConfig{
		MaxSize:          5, // Very small limit
		AllowedMimeTypes: []string{"text/plain; charset=utf-8"},
	}

	req := createMultipartRequest(t, "file", "large.txt", "This content is too large")

	result, err := helpers.UploadFile(req, "file", config, nil)

	if err == nil {
		t.Fatal("Expected error for file too large")
	}
	if result != nil {
		t.Error("Expected nil result on error")
	}
	if !strings.Contains(err.Error(), "exceeds maximum") {
		t.Errorf("Expected error about file size, got: %v", err)
	}
}

func TestUploadFile_InvalidMimeType(t *testing.T) {
	helpers := &Helpers{}
	config := FileUploadConfig{
		MaxSize:          1024,
		AllowedMimeTypes: []string{"image/jpeg"}, // Only allow JPEG
	}

	req := createMultipartRequest(t, "file", "test.txt", "Plain text content")

	result, err := helpers.UploadFile(req, "file", config, nil)

	if err == nil {
		t.Fatal("Expected error for invalid MIME type")
	}
	if result != nil {
		t.Error("Expected nil result on error")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("Expected error about file type not allowed, got: %v", err)
	}
}

// Mock filesystem for testing
type MockFS struct {
	shouldFail bool
}

func (m *MockFS) Put(string, string, ...string) error {
	if m.shouldFail {
		return errors.New("mock filesystem error")
	}
	return nil
}

func (m *MockFS) Delete([]string) bool {
	return true
}

func (m *MockFS) Get(string, ...string) error {
	return nil
}

func (m *MockFS) List(string) ([]filesystem.Listing, error) {
	return []filesystem.Listing{}, nil
}

func TestUploadFile_WithFilesystem(t *testing.T) {
	tempDir := t.TempDir()
	helpers := &Helpers{}
	mockFS := &MockFS{shouldFail: false}

	config := FileUploadConfig{
		MaxSize:          1024,
		AllowedMimeTypes: []string{"text/plain; charset=utf-8"},
		TempDir:          tempDir,
		Destination:      "/cloud/storage",
	}

	req := createMultipartRequest(t, "file", "cloud.txt", "Cloud content")

	result, err := helpers.UploadFile(req, "file", config, mockFS)

	if err != nil {
		t.Fatalf("Expected no error with mock filesystem, got: %v", err)
	}
	if result.Path != "" {
		t.Error("Expected empty path when using filesystem interface")
	}
}

func TestUploadFile_FilesystemError(t *testing.T) {
	tempDir := t.TempDir()
	helpers := &Helpers{}
	mockFS := &MockFS{shouldFail: true}

	config := FileUploadConfig{
		MaxSize:          1024,
		AllowedMimeTypes: []string{"text/plain; charset=utf-8"},
		TempDir:          tempDir,
		Destination:      "/cloud/storage",
	}

	req := createMultipartRequest(t, "file", "fail.txt", "Content")

	result, err := helpers.UploadFile(req, "file", config, mockFS)

	if err == nil {
		t.Fatal("Expected error from failing filesystem")
	}
	if result != nil {
		t.Error("Expected nil result on filesystem error")
	}
}

func TestGenerateSafeFilename(t *testing.T) {
	helpers := &Helpers{}

	testCases := []struct {
		input     string
		extension string
		checks    []string // Strings that should be present
	}{
		{
			input:     "normal.txt",
			extension: ".txt",
			checks:    []string{"normal_", ".txt"},
		},
		{
			input:     "../../../etc/passwd",
			extension: ".txt",
			checks:    []string{"passwd_", ".txt"},
		},
		{
			input:     "file with spaces.doc",
			extension: ".doc",
			checks:    []string{"file_with_spaces_", ".doc"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := helpers.generateSafeFilename(tc.input, tc.extension)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			for _, check := range tc.checks {
				if !strings.Contains(result, check) {
					t.Errorf("Expected '%s' to contain '%s'", result, check)
				}
			}

			// Should have random component
			if !strings.Contains(result, "_") {
				t.Errorf("Expected filename to have random component: %s", result)
			}
		})
	}
}
