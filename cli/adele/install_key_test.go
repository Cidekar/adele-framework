package main

import (
	"os"
	"strings"
	"testing"
)

// seedAdeleApp drops a minimal go.mod into the CWD so IsAdeleApp() reports
// true. Mirrors the pattern used in install_test.go.
func seedAdeleApp(t *testing.T) {
	t.Helper()
	mod := "module example.com/app\n\nrequire github.com/cidekar/adele-framework v0.0.0\n"
	if err := os.WriteFile("go.mod", []byte(mod), 0644); err != nil {
		t.Fatalf("seed go.mod: %v", err)
	}
}

func TestInstallKey_FreshEnv_AppendsKey(t *testing.T) {
	t.Chdir(t.TempDir())
	seedAdeleApp(t)
	// No .env file at all — command must create it.

	if err := NewInstallKey(false).Handle(); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	got, err := os.ReadFile(".env")
	if err != nil {
		t.Fatalf("read .env: %v", err)
	}
	if !strings.Contains(string(got), "APP_KEY=") {
		t.Errorf("expected .env to contain APP_KEY=, got: %q", got)
	}
}

func TestInstallKey_ExistingEnvWithoutAppKey_AppendsLine(t *testing.T) {
	t.Chdir(t.TempDir())
	seedAdeleApp(t)
	original := "APP_NAME=test\nAPP_DEBUG=true\n"
	if err := os.WriteFile(".env", []byte(original), 0644); err != nil {
		t.Fatalf("seed .env: %v", err)
	}

	if err := NewInstallKey(false).Handle(); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	got, err := os.ReadFile(".env")
	if err != nil {
		t.Fatalf("read .env: %v", err)
	}
	s := string(got)
	if !strings.Contains(s, "APP_NAME=test") {
		t.Errorf("expected original APP_NAME line preserved, got: %q", s)
	}
	if !strings.Contains(s, "APP_DEBUG=true") {
		t.Errorf("expected original APP_DEBUG line preserved, got: %q", s)
	}
	if !strings.Contains(s, "APP_KEY=") {
		t.Errorf("expected APP_KEY appended, got: %q", s)
	}
}

func TestInstallKey_ExistingEmptyAppKey_OverwritesWithoutPrompt(t *testing.T) {
	t.Chdir(t.TempDir())
	seedAdeleApp(t)
	// Empty APP_KEY value should NOT trigger the prompt — it's effectively unset.
	if err := os.WriteFile(".env", []byte("APP_KEY=\nOTHER=x\n"), 0644); err != nil {
		t.Fatalf("seed .env: %v", err)
	}

	if err := NewInstallKey(false).Handle(); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	got, _ := os.ReadFile(".env")
	s := string(got)
	if !strings.Contains(s, "APP_KEY=") || strings.Contains(s, "APP_KEY=\n") {
		t.Errorf("expected APP_KEY to be set to a non-empty value, got: %q", s)
	}
	if !strings.Contains(s, "OTHER=x") {
		t.Errorf("expected unrelated line preserved, got: %q", s)
	}
}

func TestInstallKey_ExistingValue_PromptCancelled(t *testing.T) {
	t.Chdir(t.TempDir())
	seedAdeleApp(t)
	const existing = "DO_NOT_OVERWRITE_ME"
	if err := os.WriteFile(".env", []byte("APP_KEY="+existing+"\n"), 0644); err != nil {
		t.Fatalf("seed .env: %v", err)
	}

	// Pipe "n\n" into stdin — command should cancel without rewriting.
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdin = r
	_, _ = w.WriteString("n\n")
	w.Close()

	err = NewInstallKey(false).Handle()
	if err == nil {
		t.Fatal("expected error from cancelled prompt, got nil")
	}
	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("expected cancellation error, got: %v", err)
	}

	got, _ := os.ReadFile(".env")
	if !strings.Contains(string(got), existing) {
		t.Errorf("expected existing key preserved on cancel, got: %q", got)
	}
}

func TestInstallKey_ExistingValue_PromptConfirmed_Overwrites(t *testing.T) {
	t.Chdir(t.TempDir())
	seedAdeleApp(t)
	const existing = "OLD_KEY_VALUE"
	if err := os.WriteFile(".env", []byte("APP_KEY="+existing+"\n"), 0644); err != nil {
		t.Fatalf("seed .env: %v", err)
	}

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdin = r
	_, _ = w.WriteString("yes\n")
	w.Close()

	if err := NewInstallKey(false).Handle(); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	got, _ := os.ReadFile(".env")
	if strings.Contains(string(got), existing) {
		t.Errorf("expected old key replaced, but %q still present in: %q", existing, got)
	}
	if !strings.Contains(string(got), "APP_KEY=") {
		t.Errorf("expected new APP_KEY in: %q", got)
	}
}

func TestInstallKey_Force_OverwritesWithoutPrompt(t *testing.T) {
	t.Chdir(t.TempDir())
	seedAdeleApp(t)
	const existing = "WILL_BE_REPLACED"
	if err := os.WriteFile(".env", []byte("APP_KEY="+existing+"\n"), 0644); err != nil {
		t.Fatalf("seed .env: %v", err)
	}

	// No stdin pipe — if --force prompts, the test will hang. The fact that
	// it does not hang is part of the assertion.
	if err := NewInstallKey(true).Handle(); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	got, _ := os.ReadFile(".env")
	if strings.Contains(string(got), existing) {
		t.Errorf("expected --force to replace existing key, but %q still present in: %q", existing, got)
	}
}

func TestInstallKey_NotInAdeleApp_Errors(t *testing.T) {
	t.Chdir(t.TempDir())
	// Deliberately do NOT seed go.mod — IsAdeleApp() should return false.

	err := NewInstallKey(false).Handle()
	if err == nil {
		t.Fatal("expected error when not in an adele app, got nil")
	}
	if !strings.Contains(err.Error(), "adele application") {
		t.Errorf("expected error to mention 'adele application', got: %v", err)
	}

	if _, err := os.Stat(".env"); !os.IsNotExist(err) {
		t.Errorf("expected no .env to be created when not in an adele app, got Stat err: %v", err)
	}
}

func TestInstallKey_KeyHasExpectedShape(t *testing.T) {
	t.Chdir(t.TempDir())
	seedAdeleApp(t)

	if err := NewInstallKey(false).Handle(); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	got, _ := os.ReadFile(".env")
	lines := strings.Split(strings.TrimSpace(string(got)), "\n")
	var key string
	for _, line := range lines {
		if strings.HasPrefix(line, "APP_KEY=") {
			key = strings.TrimPrefix(line, "APP_KEY=")
			break
		}
	}
	if key == "" {
		t.Fatalf("APP_KEY missing or empty in .env: %q", got)
	}
	if len(key) != keyLength {
		t.Errorf("expected key length %d, got %d (%q)", keyLength, len(key), key)
	}
}

func TestInstallKey_TwoRunsProduceDifferentKeys(t *testing.T) {
	t.Chdir(t.TempDir())
	seedAdeleApp(t)

	if err := NewInstallKey(false).Handle(); err != nil {
		t.Fatalf("first Handle(): %v", err)
	}
	first, _ := readEnvValue(".env", "APP_KEY")

	if err := NewInstallKey(true).Handle(); err != nil {
		t.Fatalf("second Handle(): %v", err)
	}
	second, _ := readEnvValue(".env", "APP_KEY")

	if first == "" || second == "" {
		t.Fatalf("expected both keys to be non-empty (first=%q second=%q)", first, second)
	}
	if first == second {
		t.Errorf("expected two separate runs to produce different keys, both were %q", first)
	}
}

func TestInstallKey_PreservesCommentsAndOrder(t *testing.T) {
	t.Chdir(t.TempDir())
	seedAdeleApp(t)
	original := "# managed env\nAPP_NAME=test\n\n# auth\nAPP_KEY=OLD\nDB_HOST=localhost\n"
	if err := os.WriteFile(".env", []byte(original), 0644); err != nil {
		t.Fatalf("seed .env: %v", err)
	}

	if err := NewInstallKey(true).Handle(); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	got, _ := os.ReadFile(".env")
	s := string(got)
	if !strings.Contains(s, "# managed env") {
		t.Errorf("expected leading comment preserved: %q", s)
	}
	if !strings.Contains(s, "# auth") {
		t.Errorf("expected inline comment preserved: %q", s)
	}
	if !strings.Contains(s, "DB_HOST=localhost") {
		t.Errorf("expected trailing line preserved: %q", s)
	}
	// APP_KEY should remain in its original position (4th non-blank line).
	lines := strings.Split(strings.TrimSpace(s), "\n")
	var keyIdx = -1
	for i, line := range lines {
		if strings.HasPrefix(line, "APP_KEY=") {
			keyIdx = i
			break
		}
	}
	if keyIdx == -1 {
		t.Fatalf("APP_KEY missing: %q", s)
	}
	if keyIdx != 4 {
		t.Errorf("expected APP_KEY rewritten at index 4 (its original slot), got index %d in %v", keyIdx, lines)
	}
}

func TestInstallCommand_KeyExampleListed(t *testing.T) {
	// Sanity check that the help text advertises the `key` subcommand.
	var found bool
	for _, ex := range InstallCommand.Examples {
		if strings.Contains(ex, "adele install key") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected `adele install key` in InstallCommand.Examples, got %v", InstallCommand.Examples)
	}
}
