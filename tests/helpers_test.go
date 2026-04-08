package tests

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/DisruptiveWorks/archipulse/internal/api"
	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/google/uuid"
)

const testJWTSecret = "test-secret-do-not-use-in-production"

// nonExistentUUID returns a valid UUID that is guaranteed not to exist in the database.
func nonExistentUUID() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

// repoRoot returns the absolute path to the repository root,
// regardless of the working directory when tests are run.
func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..")
}

// fixture returns the absolute path to a file under examples/.
func fixture(name string) string {
	return filepath.Join(repoRoot(), "examples", name)
}

// testAuthService builds an auth.Service suitable for integration tests.
// It uses a fixed JWT secret and connects to the test DB.
func testAuthService(t *testing.T, conn *sql.DB) *auth.Service {
	t.Helper()
	cfg := &auth.Config{
		JWTSecret:    testJWTSecret,
		TokenTTL:     time.Hour,
		CookieName:   "ap_session",
		CookieSecure: false,
	}
	svc, err := auth.NewService(conn, cfg)
	if err != nil {
		t.Fatalf("testAuthService: %v", err)
	}
	return svc
}

// testRouter returns a fully wired http.Handler with a real auth.Service using
// a fixed test secret. Use addAuthCookie to attach an admin session to requests.
func testRouter(t *testing.T, conn *sql.DB) http.Handler {
	t.Helper()
	svc := testAuthService(t, conn)
	return api.NewRouter(conn, svc, nil)
}

// addAuthCookie attaches a signed admin JWT cookie to req, so that the
// RequireAuth + RequireRole middleware passes in integration tests.
func addAuthCookie(t *testing.T, req *http.Request) {
	t.Helper()
	cfg := &auth.Config{
		JWTSecret: testJWTSecret,
		TokenTTL:  time.Hour,
		CookieName: "ap_session",
	}
	token, err := auth.IssueToken(cfg, fmt.Sprintf("%s", uuid.New()), "test@example.com", "admin")
	if err != nil {
		t.Fatalf("addAuthCookie: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "ap_session", Value: token})
}
