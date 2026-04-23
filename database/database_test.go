package database

import (
	"context"
	"database/sql"
	"log"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestOpenDatabaseConnectionPgx(t *testing.T) {

	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:15",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60*time.Second),
				wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second),
			),
		),
	)

	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Terminate(ctx)

	// Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		log.Fatal(err)
	}

	// The credentials you set
	database := "testdb"
	username := "testuser"
	password := "testpass"

	db, err := OpenDB("pgx", &DataSourceName{
		Host:         host,
		Port:         port.Port(),
		User:         username,
		Password:     password,
		DatabaseName: database,
		SslMode:      "disable",
	})

	if err != nil {
		t.Errorf("OpenDB() error = %v", err)
	}

	if db == nil {
		t.Errorf("db is nil when a pointer was expected")
	}

}

func TestOpenDatabaseConnectionSql(t *testing.T) {

	ctx := context.Background()

	container, err := mysql.Run(ctx,
		"mysql:8.0",
		mysql.WithDatabase("testdb"),
		mysql.WithUsername("testuser"),
		mysql.WithPassword("testpass"),
	)

	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Terminate(ctx)

	// Get connection details
	host, err := container.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}

	port, err := container.MappedPort(ctx, "3306/tcp")
	if err != nil {
		log.Fatal(err)
	}

	// The credentials you set
	database := "testdb"
	username := "testuser"
	password := "testpass"

	db, err := OpenDB("mysql", &DataSourceName{
		Host:         host,
		Port:         port.Port(),
		User:         username,
		Password:     password,
		DatabaseName: database,
	})

	if err != nil {
		t.Errorf("OpenDB() error = %v", err)
	}

	if db == nil {
		t.Errorf("db is nil when a pointer was expected")
	}

}

func TestNewSessionPgx(t *testing.T) {
	ctx := context.Background()

	// Start PostgreSQL container
	container, err := postgres.Run(ctx,
		"postgres:15",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)

	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Terminate(ctx)

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	d := Database{
		DataType: "pgx",
		Pool:     db,
	}

	session := d.NewSession()

	if session == nil {
		t.Fatalf("nil returned when session was expected")
	}
}

func TestNewSessionSql(t *testing.T) {
	ctx := context.Background()

	// Start MySQL container
	container, err := mysql.Run(ctx,
		"mysql:8.0",
		mysql.WithDatabase("testdb"),
		mysql.WithUsername("testuser"),
		mysql.WithPassword("testpass"),
	)

	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}
	defer container.Terminate(ctx)

	// Get connection string
	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Connect to database
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	d := Database{
		DataType: "mysql",
		Pool:     db,
	}

	session := d.NewSession()

	if session == nil {
		t.Fatalf("nil returned when session was expected")
	}
}
