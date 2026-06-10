// Package mysqldriver provides MySQL-specific helpers for the database layer.
//
// It builds MySQL data source name strings and creates upper/db query sessions
// from an existing *sql.DB connection pool.
package mysqldriver

import (
	"database/sql"
	"fmt"

	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
)

// Build a data source name string
func BuildDSN(host, port, user, password, dbname string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC", user, password, host, port, dbname)
}

// Create a new MySQL builder session
func Session(pool *sql.DB) (db.Session, error) {
	session, err := mysql.New(pool)
	if err != nil {
		return nil, err
	}

	return session, nil
}
