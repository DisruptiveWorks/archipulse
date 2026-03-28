// Package tests contains integration tests that require a real PostgreSQL database.
// Set DATABASE_URL in the environment (or .env) before running.
package tests

import (
	"database/sql"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/DisruptiveWorks/archipulse/internal/db"
)

// openTestDB opens a connection to the test database and runs migrations.
// It skips the test if DATABASE_URL is not set.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	_ = godotenv.Load("../.env")

	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set — skipping integration test")
	}

	conn, err := db.Connect()
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}

	if err := db.Migrate(conn, "../migrations"); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	t.Cleanup(func() { _ = conn.Close() })
	return conn
}
