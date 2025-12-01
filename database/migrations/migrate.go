// Package migrations provides database migration utilities built on top of golang-migrate.
package migrations

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
)

// MigrateUp applies all available migrations to bring the database schema up to date.
// It reads migration files from the configured RootPath/migrations directory and
// connects to the database using the provided DSN (Data Source Name).
// Returns an error if the migration instance cannot be created or if any migration fails.
func (adele *Migration) MigrateUp(dsn string) error {

	m, err := migrate.New("file://"+adele.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		log.Println("Error running migration: ", err)
		return err
	}
	return nil
}

// MigrateDownAll reverts all applied migrations, rolling the database schema back
// to its initial state. This is useful for testing or completely resetting the database.
// Returns an error if the migration instance cannot be created or if any rollback fails.
func (adele *Migration) MigrateDownAll(dsn string) error {
	m, err := migrate.New("file://"+adele.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Down(); err != nil {
		return err
	}
	return nil
}

// Steps applies or reverts a specific number of migrations based on the value of n.
// If n is positive, it applies n migrations forward.
// If n is negative, it reverts n migrations backward.
// This provides fine-grained control over migration state changes.
// Returns an error if the migration instance cannot be created or if any step fails.
func (adele *Migration) Steps(n int, dsn string) error {
	m, err := migrate.New("file://"+adele.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(n); err != nil {
		return err
	}
	return nil
}

// MigrateForce resets the migration version to -1 (no version) without running any migrations.
// This is useful for recovering from a dirty migration state where a migration failed
// partway through and left the database in an inconsistent state.
// Use with caution as it does not modify the actual database schema.
// Returns an error if the migration instance cannot be created or if the force operation fails.
func (adele *Migration) MigrateForce(dsn string) error {
	m, err := migrate.New("file://"+adele.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Force(-1); err != nil {
		return err
	}

	return nil
}
