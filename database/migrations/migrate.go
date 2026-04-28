// Package migrations provides database migration utilities built on top of golang-migrate.
package migrations

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	// Register the file:// source driver and the postgres + mysql database
	// drivers via side-effect imports. golang-migrate's New() looks them up by
	// scheme; without these imports it errors with "unknown driver 'file'" /
	// "unknown driver 'postgres'" at first call.
	//
	// The pgx postgres driver (rather than the default lib/pq one) is used
	// to match the framework's runtime DB driver. lib/pq rejects sslmode=prefer
	// which is the framework's default for a connection without explicit SSL
	// config; pgx accepts it.
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// stdoutLogger satisfies migrate.Logger. We filter out golang-migrate's
// lifecycle chatter ("Start buffering", "Read and execute", "Closing source
// and database") and rewrite the "Finished N/u name (read X, ran Y)" line
// into something a human can read at a glance. Verbose() must be true to
// receive the events at all.
type stdoutLogger struct{}

// finishedMigrationLine matches golang-migrate's verbose finish event:
//
//	Finished 1/u create_users_table (read 3.759042ms, ran 10.321583ms)
//
// Capture groups: 1=version, 2=direction (u|d), 3=name, 4=ran-duration.
var finishedMigrationLine = regexp.MustCompile(`^Finished (\d+)/(u|d) (\S+) \(read [^,]+, ran ([^)]+)\)`)

func (stdoutLogger) Printf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	trimmed := strings.TrimSpace(msg)

	if strings.HasPrefix(trimmed, "Start buffering") ||
		strings.HasPrefix(trimmed, "Read and execute") ||
		strings.HasPrefix(trimmed, "Closing source and database") {
		return
	}

	if match := finishedMigrationLine.FindStringSubmatch(trimmed); match != nil {
		version, dir, name, ran := match[1], match[2], match[3], match[4]
		direction := "up"
		if dir == "d" {
			direction = "down"
		}
		// Pad version to 4 digits so output lines up with the file naming.
		fmt.Fprintf(os.Stdout, "  %04s_%s.%s.sql  (%s)\n", version, name, direction, ran)
		return
	}

	fmt.Fprint(os.Stdout, msg)
}

func (stdoutLogger) Verbose() bool { return true }

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
	m.Log = stdoutLogger{}

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
	m.Log = stdoutLogger{}

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
	m.Log = stdoutLogger{}

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
	m.Log = stdoutLogger{}

	if err := m.Force(-1); err != nil {
		return err
	}

	return nil
}

// MigrateDrop wipes the entire database schema, including the
// schema_migrations tracking table golang-migrate maintains. Use this when
// down migrations are unrunnable (dirty state, broken migrations, schema
// drift) and you just want to start over.
func (adele *Migration) MigrateDrop(dsn string) error {
	m, err := migrate.New("file://"+adele.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()
	m.Log = stdoutLogger{}

	if err := m.Drop(); err != nil {
		return err
	}

	return nil
}
