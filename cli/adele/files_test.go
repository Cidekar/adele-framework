package main

import (
	"os"
	"path/filepath"
	"testing"
)

var dummyFile = "dummyfile.go"
var path = "./testdata/"

func TestCopyFileFromTemplate_fails_if_found(t *testing.T) {

	err := copyFileFromTemplate(path+"dummyfile.go", path+dummyFile)
	if err == nil {
		t.Error("unable to prevent overwrite of a file that already exists")
	}
}

func TestCopyDataToFile_returns_nil(t *testing.T) {

	bytes := []byte("1234abcd")
	err := copyDataToFile(bytes, path+"testCopyDataToFile.go")
	if err != nil {
		t.Error("unable to copy data to file")
	}

	_ = os.Remove(path + "testCopyDataToFile.go")
}

func TestFileExists(t *testing.T) {
	// Create temporary file for testing
	tempFile, err := os.CreateTemp("", "test_exists")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Test existing file
	if !fileExists(tempFile.Name()) {
		t.Error("fileExists should return true for existing file")
	}

	// Test non-existent file
	if fileExists("/this/file/does/not/exist") {
		t.Error("fileExists should return false for non-existent file")
	}
}

func TestCopyDataToFile(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "test_output.txt")
	testData := []byte("Hello, World!")

	err := copyDataToFile(testData, targetFile)
	if err != nil {
		t.Fatalf("copyDataToFile failed: %v", err)
	}

	// Verify file was created
	if !fileExists(targetFile) {
		t.Error("File was not created")
	}

	// Verify file contents
	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("File content mismatch: got %q, want %q", string(content), string(testData))
	}

	// Verify file permissions
	info, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	expectedMode := os.FileMode(0644)
	if info.Mode().Perm() != expectedMode {
		t.Errorf("File permissions mismatch: got %o, want %o", info.Mode().Perm(), expectedMode)
	}
}

func TestCopyDataToFile_InvalidPath(t *testing.T) {
	invalidPath := "/root/cannot_write_here"
	testData := []byte("test data")

	err := copyDataToFile(testData, invalidPath)
	if err == nil {
		t.Error("copyDataToFile should fail with invalid path")
	}
}

func TestCopyFileFromTemplate_FileExists(t *testing.T) {
	// Create temporary existing file
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "existing.txt")
	err := os.WriteFile(existingFile, []byte("existing content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = copyFileFromTemplate("templates/test.txt", existingFile)
	if err == nil {
		t.Error("copyFileFromTemplate should fail when target file exists")
	}

	expectedError := existingFile + " already exists!"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestCopyFileFromTemplate_NonExistentTemplate(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "output.txt")

	err := copyFileFromTemplate("templates/nonexistent.txt", targetFile)
	if err == nil {
		t.Error("copyFileFromTemplate should fail with non-existent template")
	}
}

func TestDeleteFile(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "to_delete.txt")

	if err := os.WriteFile(target, []byte("delete me"), 0644); err != nil {
		t.Fatalf("Failed to seed file: %v", err)
	}

	if err := deleteFile(target); err != nil {
		t.Errorf("deleteFile returned unexpected error: %v", err)
	}

	if fileExists(target) {
		t.Error("Expected file to be removed after deleteFile")
	}
}

func TestDeleteFile_MissingFile(t *testing.T) {
	err := deleteFile("/this/file/does/not/exist")
	if err == nil {
		t.Error("Expected deleteFile to return error for non-existent file")
	}
}

func TestCopyDataToFile_EmptyData(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "empty.txt")
	emptyData := []byte("")

	err := copyDataToFile(emptyData, targetFile)
	if err != nil {
		t.Fatalf("copyDataToFile failed with empty data: %v", err)
	}

	// Verify empty file was created
	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read empty file: %v", err)
	}

	if len(content) != 0 {
		t.Errorf("Expected empty file, got %d bytes", len(content))
	}
}
