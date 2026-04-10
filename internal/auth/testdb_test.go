package auth_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/DisruptiveWorks/archipulse/internal/db"
)

// openTestDB opens the test database and runs migrations.
// Skips the test if DATABASE_URL is not set.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	_ = godotenv.Load("../../.env")

	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set — skipping integration test")
	}

	conn, err := db.Connect()
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	if err := db.Migrate(conn, "../../migrations"); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// newTestService creates an auth.Service backed by the test DB.
func newTestService(t *testing.T, conn *sql.DB) *auth.Service {
	t.Helper()
	cfg := &auth.Config{
		JWTSecret:  "test-secret-do-not-use-in-production",
		TokenTTL:   time.Hour,
		CookieName: "ap_session",
	}
	svc, err := auth.NewService(conn, cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	return svc
}
