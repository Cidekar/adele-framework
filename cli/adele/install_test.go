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

func TestInstall_Handle_DispatchesToStarterKit(t *testing.T) {
	// Save and restore Registry args
	originalArgs := Registry.GetArgs()
	defer Registry.SetArgs(originalArgs)

	// Save and restore stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Pipe with "n" to cancel — verifies dispatch landed in StarterKit.Handle()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdin = r
	_, _ = w.WriteString("n\n")
	w.Close()

	// Use temp dir so any side effects are isolated
	t.Chdir(t.TempDir())

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
