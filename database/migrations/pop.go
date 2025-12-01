package migrations

import (
	"github.com/gobuffalo/pop"
)

// PopConnect establishes a database connection using the Pop library.
// Currently defaults to the "development" environment configuration.
// Returns a Pop connection instance or an error if the connection fails.
// TODO: Do we want to default to development? Seems to me that a env pivot is helpful.
func (a *Migration) PopConnect() (*pop.Connection, error) {
	tx, err := pop.Connect("development")
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreatePopMigration generates new migration files using the Pop library.
// It creates both up and down migration files in the RootPath/migrations directory.
// Parameters:
//   - up: the SQL or Go code to run when migrating up
//   - down: the SQL or Go code to run when migrating down (rollback)
//   - migrationName: the name for the migration (used in filename)
//   - migrationType: the type of migration (e.g., "sql" or "fizz")
//
// Returns an error if the migration files cannot be created.
func (a *Migration) CreatePopMigration(up, down []byte, migrationName, migrationType string) error {
	var migrationPath = a.RootPath + "/migrations"
	err := pop.MigrationCreate(migrationPath, migrationName, migrationType, up, down)
	if err != nil {
		return err
	}
	return nil
}

// RunPopMigrations applies all pending migrations using the Pop library.
// It reads migration files from RootPath/migrations and executes them in order.
// Requires an active Pop database connection.
// Returns an error if the migrator cannot be created or if any migration fails.
func (a *Migration) RunPopMigrations(tx *pop.Connection) error {
	var migrationPath = a.RootPath + "/migrations"

	fm, err := pop.NewFileMigrator(migrationPath, tx)
	if err != nil {
		return err
	}

	err = fm.Up()
	if err != nil {
		return err
	}

	return nil
}

// PopMigrateDown reverts a specified number of migrations using the Pop library.
// The steps parameter is variadic; if not provided, it defaults to reverting 1 migration.
// If steps is provided, the first value determines how many migrations to roll back.
// Returns an error if the migrator cannot be created or if the rollback fails.
func (a *Migration) PopMigrateDown(tx *pop.Connection, steps ...int) error {
	var migrationPath = a.RootPath + "/migrations"

	step := 1
	if len(steps) > 0 {
		step = steps[0]
	}

	fm, err := pop.NewFileMigrator(migrationPath, tx)
	if err != nil {
		return err
	}

	err = fm.Down(step)
	if err != nil {
		return err
	}
	return nil
}

// PopMigrateReset rolls back all migrations and then re-applies them.
// This effectively rebuilds the entire database schema from scratch.
// Useful for development and testing environments to ensure a clean state.
// Returns an error if the migrator cannot be created or if the reset fails.
func (a *Migration) PopMigrateReset(tx *pop.Connection) error {
	var migrationPath = a.RootPath + "/migrations"
	fm, err := pop.NewFileMigrator(migrationPath, tx)
	if err != nil {
		return err
	}
	err = fm.Reset()
	if err != nil {
		return err
	}
	return nil
}
