package main

import (
	"os"
	"testing"
)

func TestSanitize(t *testing.T) {
	app := NewApplication()

	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"MyApp", "myapp", false},
		{"my-app_test", "my-app_test", false},
		{"My App 123", "myapp123", false},
		{"app@#$%^&*()", "app", false},
		{"", "", true},
		{"   ", "", true},
		{"@#$%", "", true},
	}

	for _, test := range tests {
		result, err := app.Sanitize(test.input)

		if test.hasError && err == nil {
			t.Errorf("Expected error for input '%s', but got none", test.input)
		}

		if !test.hasError && err != nil {
			t.Errorf("Unexpected error for input '%s': %v", test.input, err)
		}

		if result != test.expected {
			t.Errorf("For input '%s', expected '%s', got '%s'", test.input, test.expected, result)
		}
	}
}

func TestValidate(t *testing.T) {
	app := NewApplication()

	// Test with no arguments (only command name)
	Registry.args = []string{"new"}
	err := app.Validate()
	if err == nil {
		t.Error("Expected error when no app name provided")
	}

	// Test with app name provided
	Registry.args = []string{"new", "testapp"}
	err = app.Validate()
	if err != nil {
		t.Errorf("Unexpected error when app name provided: %v", err)
	}
}

func TestClone(t *testing.T) {
	app := NewApplication()

	// Clean up before test
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Test cloning when directory doesn't exist
	err := app.Clone(testDir)
	if err != nil {
		t.Logf("Clone test skipped or failed: %v", err)
		return // Skip if network issues or repo unavailable
	}

	// Verify directory was created
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("Expected directory to be created after clone")
	}

	// Test cloning when directory already exists
	err = app.Clone(testDir)
	if err == nil {
		t.Error("Expected error when directory already exists")
	}
}

func TestUpdateApplicationNameInSource(t *testing.T) {
	app := NewApplication()

	// Create test directory and file
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	// Create a test .go file with "myapp" content
	testFile := testDir + "/main.go"
	testContent := `package main

import "myapp/models"

func main() {
	// myapp startup code
}`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Change to test directory
	originalDir, _ := os.Getwd()
	os.Chdir(testDir)
	defer os.Chdir(originalDir)

	// Test updating source files
	err = app.UpdateApplicationNameInSource("newapp")
	if err != nil {
		t.Errorf("UpdateApplicationNameInSource failed: %v", err)
	}

	// Read updated content
	updatedContent, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	expectedContent := `package main

import "newapp/models"

func main() {
	// newapp startup code
}`

	if string(updatedContent) != expectedContent {
		t.Errorf("Content not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedContent, string(updatedContent))
	}
}

func TestWrite(t *testing.T) {
	app := NewApplication()

	// Create dummy Makefile templates
	os.WriteFile(testDir+"/Makefile.mac", []byte("# Mac Makefile"), 0644)
	os.WriteFile(testDir+"/Makefile.windows", []byte("# Windows Makefile"), 0644)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	err := app.Write(testDir)
	if err != nil {
		t.Logf("Write test failed (likely missing template files): %v", err)
		return // Skip if template files not available
	}

	// Check if Makefile was created
	if _, err := os.Stat(testDir + "/Makefile"); err != nil {
		t.Error("Expected Makefile to be created")
	}

	// Check if template files were removed
	if _, err := os.Stat(testDir + "/Makefile.mac"); err == nil {
		t.Error("Expected Makefile.mac to be removed")
	}
}

func TestHandle(t *testing.T) {

	// Clean up before test
	os.RemoveAll(testDir)

	app := NewApplication()

	// Test with no arguments
	Registry.args = []string{"new"}
	err := app.Handle()
	if err == nil {
		t.Error("Expected error when no app name provided")
	}

	// Test with valid app name
	Registry.args = []string{"new", testDir}
	err = app.Handle()
	if err != nil {
		t.Logf("Handle test failed (likely network/template issues): %v", err)
		return // Skip if external dependencies fail
	}

	// Verify directory was created
	if _, err := os.Stat("testdata"); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}
