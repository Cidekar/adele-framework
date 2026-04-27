package postgresdriver

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     string
		user     string
		password string
		dbname   string
		sslmode  string
		expected string
	}{
		{
			name:     "standard postgres connection",
			host:     "localhost",
			port:     "5432",
			user:     "postgres",
			password: "secret",
			dbname:   "myapp",
			sslmode:  "disable",
			expected: "host=localhost port=5432 user=postgres password=secret dbname=myapp sslmode=disable timezone=UTC connect_timeout=5",
		},
		{
			name:     "remote database with SSL",
			host:     "db.example.com",
			port:     "5432",
			user:     "appuser",
			password: "complexpass123",
			dbname:   "production_db",
			sslmode:  "require",
			expected: "host=db.example.com port=5432 user=appuser password=complexpass123 dbname=production_db sslmode=require timezone=UTC connect_timeout=5",
		},
		{
			name:     "custom port",
			host:     "localhost",
			port:     "5433",
			user:     "dev",
			password: "devpass",
			dbname:   "testdb",
			sslmode:  "prefer",
			expected: "host=localhost port=5433 user=dev password=devpass dbname=testdb sslmode=prefer timezone=UTC connect_timeout=5",
		},
		{
			name:     "empty password",
			host:     "localhost",
			port:     "5432",
			user:     "postgres",
			password: "",
			dbname:   "myapp",
			sslmode:  "disable",
			expected: "host=localhost port=5432 user=postgres password= dbname=myapp sslmode=disable timezone=UTC connect_timeout=5",
		},
		{
			name:     "special characters in password",
			host:     "localhost",
			port:     "5432",
			user:     "user",
			password: "p@ssw0rd!#$",
			dbname:   "app",
			sslmode:  "disable",
			expected: "host=localhost port=5432 user=user password=p@ssw0rd!#$ dbname=app sslmode=disable timezone=UTC connect_timeout=5",
		},
		{
			name:     "all empty values defaults sslmode to prefer",
			host:     "",
			port:     "",
			user:     "",
			password: "",
			dbname:   "",
			sslmode:  "",
			expected: "host= port= user= password= dbname= sslmode=prefer timezone=UTC connect_timeout=5",
		},
		{
			name:     "SSL verify-full mode",
			host:     "secure-db.example.com",
			port:     "5432",
			user:     "secureuser",
			password: "securepass",
			dbname:   "securedb",
			sslmode:  "verify-full",
			expected: "host=secure-db.example.com port=5432 user=secureuser password=securepass dbname=securedb sslmode=verify-full timezone=UTC connect_timeout=5",
		},
		{
			name:     "localhost with default postgres settings",
			host:     "127.0.0.1",
			port:     "5432",
			user:     "postgres",
			password: "postgres",
			dbname:   "postgres",
			sslmode:  "disable",
			expected: "host=127.0.0.1 port=5432 user=postgres password=postgres dbname=postgres sslmode=disable timezone=UTC connect_timeout=5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDSN(tt.host, tt.port, tt.user, tt.password, tt.dbname, tt.sslmode)
			if result != tt.expected {
				t.Errorf("BuildDSN() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestSession(t *testing.T) {
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

	// Test the session, finally.
	session, err := Session(db)
	if err != nil {
		t.Errorf("Session() error = %v", err)
	}
	if session == nil {
		t.Error("Session() returned nil")
	}

}
