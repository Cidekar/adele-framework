// Package database provides database connection management and query session
// creation across supported SQL backends.
//
// It opens and configures *sql.DB connection pools from a DataSourceName,
// resolves the configured database type to a driver, and builds upper/db query
// sessions via the mysqldriver and postgresdriver subpackages.
package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/cidekar/adele-framework/database/mysqldriver"
	"github.com/cidekar/adele-framework/database/postgresdriver"
	"github.com/upper/db/v4"
)

// NewSession creates a new sqlbuilder.Session instance based on the configured database type.
// Returns nil if no database is configured.
func (a *Database) NewSession() db.Session {
	if a.Pool == nil {
		return nil
	}

	dbType := getDBDriver(a.DataType)

	switch dbType {
	case "mysql":
		if session, err := mysqldriver.Session(a.Pool); err == nil {
			return session
		}
	case "pgx":
		if session, err := postgresdriver.Session(a.Pool); err == nil {
			return session
		}
	}

	return nil
}

// Get a connection to a database and return connection pool
func OpenDB(dbType string, config *DataSourceName) (*sql.DB, error) {

	driver := getDBDriver(dbType)

	if driver == "" {
		return nil, nil
	}

	dsn := ""
	switch driver {
	case "pgx":
		dsn = postgresdriver.BuildDSN(config.Host, config.Port, config.User, config.Password, config.DatabaseName, config.SslMode)
	case "mysql":
		dsn = mysqldriver.BuildDSN(config.Host, config.Port, config.User, config.Password, config.DatabaseName)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", dbType)
	}

	if dsn == "" {
		return nil, fmt.Errorf("failed to build DSN for database driver: %s", driver)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Set reasonable defaults
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// Convert database type to driver name
func getDBDriver(dbType string) string {
	switch strings.ToLower(strings.TrimSpace(dbType)) {
	case "postgres", "postgresql":
		return "pgx"
	case "mysql", "mariadb":
		return "mysql"
	default:
		return dbType
	}
}
