package migrations

import (
	"testing"

	adele "github.com/cidekar/adele-framework"
)

func TestMigrateUp_InvalidDSN(t *testing.T) {
	m := &Migration{
		Adele: &adele.Adele{RootPath: "/nonexistent"},
	}

	err := m.MigrateUp("invalid-dsn")
	if err == nil {
		t.Error("expected error for invalid DSN, got nil")
	}
}

func TestMigrateDownAll_InvalidDSN(t *testing.T) {
	m := &Migration{
		Adele: &adele.Adele{RootPath: "/nonexistent"},
	}

	err := m.MigrateDownAll("invalid-dsn")
	if err == nil {
		t.Error("expected error for invalid DSN, got nil")
	}
}

func TestSteps_InvalidDSN(t *testing.T) {
	m := &Migration{
		Adele: &adele.Adele{RootPath: "/nonexistent"},
	}

	err := m.Steps(1, "invalid-dsn")
	if err == nil {
		t.Error("expected error for invalid DSN, got nil")
	}
}

func TestMigrateForce_InvalidDSN(t *testing.T) {
	m := &Migration{
		Adele: &adele.Adele{RootPath: "/nonexistent"},
	}

	err := m.MigrateForce("invalid-dsn")
	if err == nil {
		t.Error("expected error for invalid DSN, got nil")
	}
}
