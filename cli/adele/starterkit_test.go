package main

import (
	"os"
	"strings"
	"testing"
)

// withStdin replaces os.Stdin for the duration of the test with a pipe whose
// read end emits the given input. The original stdin is restored on cleanup.
func withStdin(t *testing.T, input string) {
	t.Helper()

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r

	if _, err := w.WriteString(input); err != nil {
		t.Fatalf("Failed to write to stdin pipe: %v", err)
	}
	w.Close()

	t.Cleanup(func() {
		os.Stdin = oldStdin
		r.Close()
	})
}

// expectedStarterKitFiles is the list of files the starter kit handler is
// expected to write into the working directory on success.
var expectedStarterKitFiles = []string{
	"resources/views/layouts/base.jet",
	"resources/views/home.jet",
	"resources/css/styles.css",
	"resources/js/script.ts",
	"package.json",
	"vite.config.ts",
	"tailwind.config.js",
	"postcss.config.js",
}

func TestStarterKit_Handle_CancelOnN(t *testing.T) {
	t.Chdir(t.TempDir())
	withStdin(t, "n\n")

	k := NewStarterKit()
	err := k.Handle()
	if err == nil {
		t.Fatal("Expected error when user cancels with 'n'")
	}

	if !strings.Contains(strings.ToLower(err.Error()), "cancelled") {
		t.Errorf("Expected error to contain 'cancelled', got: %v", err)
	}

	// No files should have been written.
	for _, f := range expectedStarterKitFiles {
		if fileExists(f) {
			t.Errorf("Expected file %q not to be written when cancelled", f)
		}
	}
}

func TestStarterKit_Handle_CancelOnEmpty(t *testing.T) {
	t.Chdir(t.TempDir())
	withStdin(t, "\n")

	k := NewStarterKit()
	err := k.Handle()
	if err == nil {
		t.Fatal("Expected error when user provides empty input")
	}

	if !strings.Contains(strings.ToLower(err.Error()), "cancelled") {
		t.Errorf("Expected error to contain 'cancelled', got: %v", err)
	}
}

func TestStarterKit_Handle_HappyPath(t *testing.T) {
	t.Chdir(t.TempDir())

	// The kit's WriteDir only creates resources/js and resources/css; the
	// views/layouts directory is assumed to exist (real adele project layout).
	if err := os.MkdirAll("resources/views/layouts", 0755); err != nil {
		t.Fatalf("Failed to create resources/views/layouts: %v", err)
	}

	withStdin(t, "y\n")

	k := NewStarterKit()
	err := k.Handle()
	if err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	for _, f := range expectedStarterKitFiles {
		if !fileExists(f) {
			t.Errorf("Expected file %q to be written", f)
			continue
		}

		info, err := os.Stat(f)
		if err != nil {
			t.Errorf("Failed to stat %q: %v", f, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("Expected file %q to be non-empty", f)
		}
	}

	// Spot-check package.json contains "vite" (matches embedded template).
	pkg, err := os.ReadFile("package.json")
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}
	if !strings.Contains(string(pkg), "vite") {
		t.Errorf("Expected package.json to mention 'vite', got: %s", string(pkg))
	}
}

func TestStarterKit_Handle_ReplacesExistingJetTemplates(t *testing.T) {
	t.Chdir(t.TempDir())

	// Pre-create the two jet templates with stub content.
	if err := os.MkdirAll("resources/views/layouts", 0755); err != nil {
		t.Fatalf("Failed to create dirs: %v", err)
	}

	stub := []byte("STUB CONTENT")
	if err := os.WriteFile("resources/views/layouts/base.jet", stub, 0644); err != nil {
		t.Fatalf("Failed to seed base.jet: %v", err)
	}
	if err := os.WriteFile("resources/views/home.jet", stub, 0644); err != nil {
		t.Fatalf("Failed to seed home.jet: %v", err)
	}

	withStdin(t, "y\n")

	k := NewStarterKit()
	err := k.Handle()
	if err != nil {
		t.Fatalf("Handle returned unexpected error: %v", err)
	}

	// Verify both templates were replaced (no longer match the stub).
	baseContent, err := os.ReadFile("resources/views/layouts/base.jet")
	if err != nil {
		t.Fatalf("Failed to read base.jet: %v", err)
	}
	if string(baseContent) == string(stub) {
		t.Error("Expected base.jet to be replaced with embedded template content")
	}

	homeContent, err := os.ReadFile("resources/views/home.jet")
	if err != nil {
		t.Fatalf("Failed to read home.jet: %v", err)
	}
	if string(homeContent) == string(stub) {
		t.Error("Expected home.jet to be replaced with embedded template content")
	}
}

func TestStarterKit_Handle_RerunFailsBecauseNonDeletedFilesExist(t *testing.T) {
	t.Chdir(t.TempDir())

	// Pre-create the views/layouts dir (kit's WriteDir doesn't make it).
	if err := os.MkdirAll("resources/views/layouts", 0755); err != nil {
		t.Fatalf("Failed to create resources/views/layouts: %v", err)
	}

	// First run: happy path.
	withStdin(t, "y\n")
	k := NewStarterKit()
	if err := k.Handle(); err != nil {
		t.Fatalf("First run failed: %v", err)
	}

	// Second run: non-jet files (styles.css is first in copy order) already
	// exist and are not pre-deleted, so copyFileFromTemplate should refuse.
	withStdin(t, "y\n")
	err := k.Handle()
	if err == nil {
		t.Fatal("Expected error on rerun because non-deleted files already exist")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected error mentioning 'already exists', got: %v", err)
	}
	if !strings.Contains(err.Error(), "styles.css") {
		t.Errorf("expected error mentioning 'styles.css' (first non-jet file in copy order), got: %v", err)
	}
}

func TestStarterKit_WriteDir_CreatesResourceDirs(t *testing.T) {
	t.Chdir(t.TempDir())

	k := NewStarterKit()
	if err := k.WriteDir(); err != nil {
		t.Fatalf("WriteDir returned unexpected error: %v", err)
	}

	for _, dir := range []string{"resources/js", "resources/css"} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("Expected directory %q to exist: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("Expected %q to be a directory", dir)
		}
	}
}

func TestStarterKit_Help_ReturnsNil(t *testing.T) {
	k := NewStarterKit()
	if err := k.Help(); err != nil {
		t.Errorf("Expected Help() to return nil, got: %v", err)
	}
}

func TestNewStarterKit_NotNil(t *testing.T) {
	k := NewStarterKit()
	if k == nil {
		t.Fatal("Expected NewStarterKit() to return non-nil")
	}
}

// Sanity check: ensure expectedStarterKitFiles doesn't drift by verifying
// the parent directories WriteDir creates are referenced in our list.
func TestStarterKit_ExpectedFiles_IncludeWriteDirOutputs(t *testing.T) {
	for _, dir := range []string{"resources/js", "resources/css"} {
		found := false
		for _, f := range expectedStarterKitFiles {
			if strings.HasPrefix(f, dir+"/") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected at least one file under %q in expectedStarterKitFiles", dir)
		}
	}
}
