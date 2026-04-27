package main

import (
	"os"
	"strings"
	"testing"
)

func TestMigrateCommand_Registration(t *testing.T) {
	if MigrateCommand.Name != "migrate" {
		t.Errorf("Expected MigrateCommand.Name to be 'migrate', got %q", MigrateCommand.Name)
	}
	if MigrateCommand.Description == "" {
		t.Error("Expected MigrateCommand.Description to be populated")
	}
	if MigrateCommand.Usage == "" {
		t.Error("Expected MigrateCommand.Usage to be populated")
	}
	if MigrateCommand.Help == "" {
		t.Error("Expected MigrateCommand.Help to be populated")
	}

	cmd, exists := Registry.GetCommand("migrate")
	if !exists {
		t.Fatal("Expected 'migrate' command to be registered in Registry")
	}
	if cmd != MigrateCommand {
		t.Error("Expected Registry's 'migrate' command to be the same as MigrateCommand")
	}
}

func TestMigrate_Handle_MissingAction(t *testing.T) {
	originalArgs := Registry.GetArgs()
	defer Registry.SetArgs(originalArgs)

	Registry.SetArgs([]string{"migrate"})

	err := NewMigrate().Handle()
	if err == nil {
		t.Fatal("Expected error when no action provided")
	}
	if !strings.Contains(err.Error(), "missing migrate action") {
		t.Errorf("Expected error to mention missing action, got: %v", err)
	}
}

func TestMigrate_Handle_UnknownAction(t *testing.T) {
	originalArgs := Registry.GetArgs()
	defer Registry.SetArgs(originalArgs)

	Registry.SetArgs([]string{"migrate", "bogus"})
	t.Setenv("DATABASE_TYPE", "postgres")
	t.Setenv("DATABASE_USER", "u")
	t.Setenv("DATABASE_PASSWORD", "p")
	t.Setenv("DATABASE_NAME", "d")

	err := NewMigrate().Handle()
	if err == nil {
		t.Fatal("Expected error for unknown action")
	}
	if !strings.Contains(err.Error(), "unknown migrate action") {
		t.Errorf("Expected error to mention unknown action, got: %v", err)
	}
}

func TestBuildMigrateDSN_PostgresWithDefaults(t *testing.T) {
	t.Setenv("DATABASE_TYPE", "postgres")
	t.Setenv("DATABASE_HOST", "")
	t.Setenv("DATABASE_PORT", "")
	t.Setenv("DATABASE_USER", "user")
	t.Setenv("DATABASE_PASSWORD", "pass")
	t.Setenv("DATABASE_NAME", "db")
	t.Setenv("DATABASE_SSL_MODE", "")

	dsn, err := buildMigrateDSN()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.HasPrefix(dsn, "postgres://user:pass@localhost:5432/db?sslmode=") {
		t.Errorf("Expected default host/port and prefer sslmode, got: %s", dsn)
	}
	if !strings.Contains(dsn, "sslmode=prefer") {
		t.Errorf("Expected default sslmode=prefer when env empty, got: %s", dsn)
	}
}

func TestBuildMigrateDSN_PostgresExplicit(t *testing.T) {
	t.Setenv("DATABASE_TYPE", "postgres")
	t.Setenv("DATABASE_HOST", "h")
	t.Setenv("DATABASE_PORT", "1234")
	t.Setenv("DATABASE_USER", "u")
	t.Setenv("DATABASE_PASSWORD", "p")
	t.Setenv("DATABASE_NAME", "d")
	t.Setenv("DATABASE_SSL_MODE", "disable")

	dsn, err := buildMigrateDSN()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := "postgres://u:p@h:1234/d?sslmode=disable"
	if dsn != expected {
		t.Errorf("Expected %q, got %q", expected, dsn)
	}
}

func TestBuildMigrateDSN_MySQL(t *testing.T) {
	t.Setenv("DATABASE_TYPE", "mysql")
	t.Setenv("DATABASE_HOST", "h")
	t.Setenv("DATABASE_PORT", "3306")
	t.Setenv("DATABASE_USER", "u")
	t.Setenv("DATABASE_PASSWORD", "p")
	t.Setenv("DATABASE_NAME", "d")

	dsn, err := buildMigrateDSN()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := "mysql://u:p@tcp(h:3306)/d"
	if dsn != expected {
		t.Errorf("Expected %q, got %q", expected, dsn)
	}
}

func TestBuildMigrateDSN_NoTypeErrors(t *testing.T) {
	os.Unsetenv("DATABASE_TYPE")

	_, err := buildMigrateDSN()
	if err == nil {
		t.Fatal("Expected error when DATABASE_TYPE unset")
	}
	if !strings.Contains(err.Error(), "DATABASE_TYPE") {
		t.Errorf("Expected error to mention DATABASE_TYPE, got: %v", err)
	}
}

func TestBuildMigrateDSN_UnknownTypeErrors(t *testing.T) {
	t.Setenv("DATABASE_TYPE", "sqlite")

	_, err := buildMigrateDSN()
	if err == nil {
		t.Fatal("Expected error for unsupported DATABASE_TYPE")
	}
	if !strings.Contains(err.Error(), "unsupported DATABASE_TYPE") {
		t.Errorf("Expected error to mention unsupported type, got: %v", err)
	}
}
