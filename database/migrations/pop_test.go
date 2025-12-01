package migrations

import (
	"testing"

	adele "github.com/cidekar/adele-framework"
)

func TestMigrationStruct(t *testing.T) {
	m := &Migration{
		Adele: &adele.Adele{RootPath: "/test/path"},
	}

	if m.RootPath != "/test/path" {
		t.Errorf("expected RootPath to be '/test/path', got '%s'", m.RootPath)
	}
}

func TestCreatePopMigration_InvalidPath(t *testing.T) {
	m := &Migration{
		Adele: &adele.Adele{RootPath: "/nonexistent"},
	}

	err := m.CreatePopMigration([]byte("up"), []byte("down"), "test_migration", "sql")
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}
