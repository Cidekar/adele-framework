package main

import (
	"fmt"
	"os"

	adele "github.com/cidekar/adele-framework"
	"github.com/cidekar/adele-framework/database/migrations"
	"github.com/cidekar/adele-framework/helpers"
	"github.com/fatih/color"
)

var MigrateCommand = &Command{
	Name:        "migrate",
	Help:        "Run database migrations",
	Description: "Apply, revert, or force the migration version using the migration files in ./migrations/",
	Usage:       "adele migrate <up|down|force> [options]",
	Examples: []string{
		"adele migrate up",
		"adele migrate down",
		"adele migrate down --all",
		"adele migrate force",
	},
	Options: map[string]string{
		"--all": "with `down`, revert ALL migrations (default reverts a single step)",
	},
}

type Migrate struct{}

func NewMigrate() *Migrate {
	return &Migrate{}
}

func (c *Migrate) Handle() error {
	args := Registry.GetArgs()
	if len(args) < 2 {
		return fmt.Errorf("missing migrate action (up|down|force)\nusage: %s", MigrateCommand.Usage)
	}

	action := args[1]
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}

	dsn, err := buildMigrateDSN()
	if err != nil {
		return err
	}

	m := &migrations.Migration{Adele: &adele.Adele{RootPath: cwd}}

	switch action {
	case "up":
		if err := m.MigrateUp(dsn); err != nil {
			return fmt.Errorf("migrate up: %w", err)
		}
		color.Green("Migrations applied.")
	case "down":
		if HasOption("--all") {
			if err := m.MigrateDownAll(dsn); err != nil {
				return fmt.Errorf("migrate down --all: %w", err)
			}
			color.Green("All migrations reverted.")
		} else {
			if err := m.Steps(-1, dsn); err != nil {
				return fmt.Errorf("migrate down: %w", err)
			}
			color.Green("One migration reverted.")
		}
	case "force":
		if err := m.MigrateForce(dsn); err != nil {
			return fmt.Errorf("migrate force: %w", err)
		}
		color.Green("Migration version forced to -1.")
	default:
		return fmt.Errorf("unknown migrate action %q (expected up|down|force)", action)
	}
	return nil
}

// buildMigrateDSN constructs a golang-migrate URL-form DSN from the same env
// vars BootstrapDatabase reads. We hand-build the URL rather than reusing
// postgresdriver.BuildDSN because that helper returns key=value form which
// golang-migrate does not accept.
func buildMigrateDSN() (string, error) {
	dbType := os.Getenv("DATABASE_TYPE")
	if dbType == "" {
		return "", fmt.Errorf("DATABASE_TYPE is not set in .env; cannot run migrations")
	}
	h := helpers.Helpers{}
	host := h.Getenv("DATABASE_HOST", "localhost")
	port := h.Getenv("DATABASE_PORT", "5432")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_NAME")
	sslmode := os.Getenv("DATABASE_SSL_MODE")

	switch dbType {
	case "postgres", "postgresql", "pgx":
		if sslmode == "" {
			sslmode = "prefer"
		}
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbname, sslmode), nil
	case "mysql", "mariadb":
		return fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbname), nil
	default:
		return "", fmt.Errorf("unsupported DATABASE_TYPE %q (expected postgres or mysql)", dbType)
	}
}

func init() {
	if err := Registry.Register(MigrateCommand); err != nil {
		panic(fmt.Sprintf("Failed to register migrate command: %v", err))
	}
}
