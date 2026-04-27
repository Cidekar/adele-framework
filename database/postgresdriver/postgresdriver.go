package postgresdriver

import (
	"database/sql"
	"fmt"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/postgresql"
)

// Build a data source name string. If sslmode is empty we fall back to "prefer"
// — pgx rejects an empty value, and "prefer" is the safest default (tries SSL,
// falls back to plain TCP if the server doesn't support it).
func BuildDSN(host, port, user, password, dbname, sslmode string) string {
	if sslmode == "" {
		sslmode = "prefer"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
		host, port, user, password, dbname, sslmode)
}

// Create a new Postgres builder session
func Session(pool *sql.DB) (db.Session, error) {
	session, err := postgresql.New(pool)
	if err != nil {
		return nil, err
	}

	return session, nil
}
