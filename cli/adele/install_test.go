package main

import (
	"os"
	"strings"
	"testing"
)

func TestInstallCommand_Registration(t *testing.T) {
	if InstallCommand.Name != "install" {
		t.Errorf("Expected InstallCommand.Name to be 'install', got %q", InstallCommand.Name)
	}

	if InstallCommand.Description == "" {
		t.Error("Expected InstallCommand.Description to be populated")
	}

	if InstallCommand.Usage == "" {
		t.Error("Expected InstallCommand.Usage to be populated")
	}

	if InstallCommand.Help == "" {
		t.Error("Expected InstallCommand.Help to be populated")
	}

	if len(InstallCommand.Examples) == 0 {
		t.Error("Expected InstallCommand.Examples to be populated")
	}

	cmd, exists := Registry.GetCommand("install")
	if !exists {
		t.Fatal("Expected 'install' command to be registered in Registry")
	}

	if cmd != InstallCommand {
		t.Error("Expected Registry's 'install' command to be the same as InstallCommand")
	}
}

func TestInstall_Handle_MissingKitArg(t *testing.T) {
	// Save and restore Registry args
	originalArgs := Registry.GetArgs()
	defer Registry.SetArgs(originalArgs)

	Registry.SetArgs([]string{"install"})

	install := NewInstall()
	err := install.Handle()
	if err == nil {
		t.Fatal("Expected error when no kit name provided")
	}

	if !strings.Contains(err.Error(), "starter-kit") {
		t.Errorf("Expected error to mention 'starter-kit' as available kit, got: %v", err)
	}
}

func TestInstall_Handle_UnknownKit(t *testing.T) {
	// Save and restore Registry args
	originalArgs := Registry.GetArgs()
	defer Registry.SetArgs(originalArgs)

	Registry.SetArgs([]string{"install", "bogus"})

	install := NewInstall()
	err := install.Handle()
	if err == nil {
		t.Fatal("Expected error when unknown kit name provided")
	}

	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("Expected error to mention 'bogus', got: %v", err)
	}

	if !strings.Contains(err.Error(), "starter-kit") {
		t.Errorf("Expected error to list 'starter-kit' as available, got: %v", err)
	}
}

func TestInstall_ResolveVariant_DefaultIsVanilla(t *testing.T) {
	originalOptions := Registry.GetOptions()
	defer Registry.SetOptions(originalOptions)

	Registry.SetOptions([]string{})

	got, err := resolveStarterKitVariant()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got.name != vanillaVariant.name {
		t.Errorf("Expected default variant %q, got %q", vanillaVariant.name, got.name)
	}
}

func TestInstall_ResolveVariant_BareVueIsVue3(t *testing.T) {
	originalOptions := Registry.GetOptions()
	defer Registry.SetOptions(originalOptions)

	Registry.SetOptions([]string{"--vue"})

	got, err := resolveStarterKitVariant()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got.name != vue3Variant.name {
		t.Errorf("Expected bare --vue to resolve to %q, got %q", vue3Variant.name, got.name)
	}
}

func TestInstall_ResolveVariant_VueExplicit3(t *testing.T) {
	originalOptions := Registry.GetOptions()
	defer Registry.SetOptions(originalOptions)

	Registry.SetOptions([]string{"--vue=3"})

	got, err := resolveStarterKitVariant()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got.name != vue3Variant.name {
		t.Errorf("Expected --vue=3 to resolve to %q, got %q", vue3Variant.name, got.name)
	}
}

func TestInstall_ResolveVariant_Vue2(t *testing.T) {
	originalOptions := Registry.GetOptions()
	defer Registry.SetOptions(originalOptions)

	Registry.SetOptions([]string{"--vue=2"})

	got, err := resolveStarterKitVariant()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got.name != vue2Variant.name {
		t.Errorf("Expected --vue=2 to resolve to %q, got %q", vue2Variant.name, got.name)
	}
}

func TestInstall_ResolveVariant_InvalidValue(t *testing.T) {
	originalOptions := Registry.GetOptions()
	defer Registry.SetOptions(originalOptions)

	Registry.SetOptions([]string{"--vue=4"})

	_, err := resolveStarterKitVariant()
	if err == nil {
		t.Fatal("Expected error for --vue=4")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("Expected error to mention 'invalid', got: %v", err)
	}
}

// TestInstall_ResolveVariant_RealArgvBareVue exercises the full argv → Registry
// → resolveStarterKitVariant path that an end user actually hits when they run
// `adele install starter-kit --vue`. Earlier the resolver called GetOption,
// which leaks neighbor positional args (returning "starter-kit" as the value)
// and routed bare --vue into the invalid-value branch. This test pins the fix.
func TestInstall_ResolveVariant_RealArgvBareVue(t *testing.T) {
	originalArgs := os.Args
	originalOptions := Registry.GetOptions()
	originalParsedArgs := Registry.GetArgs()
	defer func() {
		os.Args = originalArgs
		Registry.SetOptions(originalOptions)
		Registry.SetArgs(originalParsedArgs)
	}()

	os.Args = []string{"adele", "install", "starter-kit", "--vue"}
	if err := Registry.ParseCmdArgs(); err != nil {
		t.Fatalf("ParseCmdArgs failed: %v", err)
	}

	got, err := resolveStarterKitVariant()
	if err != nil {
		t.Fatalf("Bare --vue should resolve to vue3 without error, got: %v", err)
	}
	if got.name != vue3Variant.name {
		t.Errorf("Expected bare --vue (real argv) to resolve to %q, got %q", vue3Variant.name, got.name)
	}
}

func TestInstall_Handle_DispatchesToStarterKit(t *testing.T) {
	// Save and restore Registry args
	originalArgs := Registry.GetArgs()
	defer Registry.SetArgs(originalArgs)

	// Save and restore stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Use temp dir so any side effects are isolated
	t.Chdir(t.TempDir())

	// Seed a fake adele go.mod so ensureAdeleApp() short-circuits and the
	// dispatch reaches StarterKit.Handle() (where the actual prompt fires).
	if err := os.WriteFile("go.mod", []byte("module example.com/app\n\nrequire github.com/cidekar/adele-framework v0.0.0\n"), 0644); err != nil {
		t.Fatalf("Failed to seed go.mod: %v", err)
	}

	// Seed a managed file so the StarterKit confirmation prompt fires; without
	// a pre-existing file the install would silently proceed (no prompt).
	if err := os.MkdirAll("resources/views/layouts", 0755); err != nil {
		t.Fatalf("Failed to create dirs: %v", err)
	}
	if err := os.WriteFile("resources/views/layouts/base.jet", []byte("STUB"), 0644); err != nil {
		t.Fatalf("Failed to seed stub: %v", err)
	}

	// Pipe with "n" to cancel — verifies dispatch landed in StarterKit.Handle()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r
	_, _ = w.WriteString("n\n")
	w.Close()

	Registry.SetArgs([]string{"install", "starter-kit"})

	install := NewInstall()
	err = install.Handle()
	if err == nil {
		t.Fatal("Expected error when starter-kit dispatch is cancelled at prompt")
	}

	if !strings.Contains(strings.ToLower(err.Error()), "cancelled") {
		t.Errorf("Expected dispatched starter-kit handler to return cancellation error, got: %v", err)
	}
}

func TestIsAdeleApp_TrueForFrameworkGoMod(t *testing.T) {
	t.Chdir(t.TempDir())

	mod := "module example.com/app\n\nrequire github.com/cidekar/adele-framework v0.0.0\n"
	if err := os.WriteFile("go.mod", []byte(mod), 0644); err != nil {
		t.Fatalf("Failed to seed go.mod: %v", err)
	}

	if !IsAdeleApp() {
		t.Error("Expected IsAdeleApp() to return true for go.mod referencing the framework")
	}
}

func TestIsAdeleApp_FalseForUnrelatedGoMod(t *testing.T) {
	t.Chdir(t.TempDir())

	mod := "module example.com/app\n\nrequire github.com/some/other v1.0.0\n"
	if err := os.WriteFile("go.mod", []byte(mod), 0644); err != nil {
		t.Fatalf("Failed to seed go.mod: %v", err)
	}

	if IsAdeleApp() {
		t.Error("Expected IsAdeleApp() to return false for go.mod that does not reference the framework")
	}
}

func TestIsAdeleApp_FalseWhenNoGoMod(t *testing.T) {
	t.Chdir(t.TempDir())

	if IsAdeleApp() {
		t.Error("Expected IsAdeleApp() to return false when go.mod is absent")
	}
}

func TestEnsureAdeleApp_PassesThroughWhenInAdeleApp(t *testing.T) {
	t.Chdir(t.TempDir())

	mod := "module example.com/app\n\nrequire github.com/cidekar/adele-framework v0.0.0\n"
	if err := os.WriteFile("go.mod", []byte(mod), 0644); err != nil {
		t.Fatalf("Failed to seed go.mod: %v", err)
	}

	justScaffolded, err := ensureAdeleApp()
	if err != nil {
		t.Errorf("Expected ensureAdeleApp() to return nil inside adele app, got: %v", err)
	}
	if justScaffolded {
		t.Error("Expected justScaffolded=false when CWD is already an adele app")
	}
}

func TestEnsureAdeleApp_ErrorsWhenNotInAdeleAppAndUserDeclines(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "n\n")

	justScaffolded, err := ensureAdeleApp()
	if err == nil {
		t.Fatal("Expected ensureAdeleApp() to error when user declines scaffold")
	}
	if justScaffolded {
		t.Error("Expected justScaffolded=false when user declines scaffold")
	}
	if !strings.Contains(err.Error(), "must be run from the root of an adele application") {
		t.Errorf("Expected decline error to mention adele application root, got: %v", err)
	}
}

func TestEnsureAdeleApp_ErrorsWhenUserConfirmsButProvidesEmptyName(t *testing.T) {
	t.Chdir(t.TempDir())

	withStdin(t, "yes\n\n")

	justScaffolded, err := ensureAdeleApp()
	if err == nil {
		t.Fatal("Expected ensureAdeleApp() to error when user confirms but provides empty name")
	}
	if justScaffolded {
		t.Error("Expected justScaffolded=false when name is empty")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "empty") {
		t.Errorf("Expected error to mention empty name, got: %v", err)
	}
}

// TestInstall_WithAuth_RejectsVue3 verifies that combining --with-auth with
// --vue=3 returns a hard error before any project-touching work runs. Vue
// support for the auth scaffold is planned but not yet shipped.
func TestInstall_WithAuth_RejectsVue3(t *testing.T) {
	originalArgs := Registry.GetArgs()
	originalOptions := Registry.GetOptions()
	defer func() {
		Registry.SetArgs(originalArgs)
		Registry.SetOptions(originalOptions)
	}()

	t.Chdir(t.TempDir())

	Registry.SetArgs([]string{"install", "starter-kit"})
	Registry.SetOptions([]string{"--vue=3", "--with-auth"})

	err := NewInstall().Handle()
	if err == nil {
		t.Fatal("Expected error when --with-auth is combined with --vue=3")
	}
	if !strings.Contains(err.Error(), "vanilla") {
		t.Errorf("Expected error to mention 'vanilla', got: %v", err)
	}
}

// TestInstall_WithAuth_RejectsVue2 mirrors the vue3 rejection for vue2.
func TestInstall_WithAuth_RejectsVue2(t *testing.T) {
	originalArgs := Registry.GetArgs()
	originalOptions := Registry.GetOptions()
	defer func() {
		Registry.SetArgs(originalArgs)
		Registry.SetOptions(originalOptions)
	}()

	t.Chdir(t.TempDir())

	Registry.SetArgs([]string{"install", "starter-kit"})
	Registry.SetOptions([]string{"--vue=2", "--with-auth"})

	err := NewInstall().Handle()
	if err == nil {
		t.Fatal("Expected error when --with-auth is combined with --vue=2")
	}
	if !strings.Contains(err.Error(), "vanilla") {
		t.Errorf("Expected error to mention 'vanilla', got: %v", err)
	}
}
