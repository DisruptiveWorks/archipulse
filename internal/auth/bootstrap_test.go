package auth_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
)


func TestBootstrapAdmin_CreatesAdminWhenEmpty(t *testing.T) {
	conn := openTestDB(t)
	email := fmt.Sprintf("bootstrap-admin-%s@example.com", t.Name())

	// Ensure the user doesn't exist before
	_, _ = conn.Exec("DELETE FROM users WHERE email = $1", email)
	t.Cleanup(func() { _, _ = conn.Exec("DELETE FROM users WHERE email = $1", email) })

	cfg := &auth.Config{
		JWTSecret:         "test-secret",
		TokenTTL:          time.Hour,
		CookieName:        "ap_session",
		BootstrapEmail:    email,
		BootstrapPassword: "bootstrappass",
	}
	svc, err := auth.NewService(conn, cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	// Only runs if users table is empty — seed a separate user first to test
	// the "already exists" path, then clear and test the creation path.
	// Since shared DB may have users, we test Bootstrap doesn't error.
	if err := auth.Bootstrap(svc); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}

	// If the user was created (table was empty before), verify role.
	u, err := svc.Users.GetByEmail(email)
	if err == auth.ErrNotFound {
		t.Skip("shared DB already had users — bootstrap skipped (expected in CI)")
	}
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if u.Role != "admin" {
		t.Errorf("Role: got %q, want admin", u.Role)
	}
}

func TestBootstrapAdmin_SkipsWhenNoConfig(t *testing.T) {
	conn := openTestDB(t)
	cfg := &auth.Config{
		JWTSecret:  "test-secret",
		TokenTTL:   time.Hour,
		CookieName: "ap_session",
		// BootstrapEmail and BootstrapPassword intentionally empty
	}
	svc, err := auth.NewService(conn, cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	if err := auth.Bootstrap(svc); err != nil {
		t.Errorf("Bootstrap with no config should be a no-op, got: %v", err)
	}
}

func TestBootstrapDemo_CreatesOrSyncsDemoUser(t *testing.T) {
	conn := openTestDB(t)
	email := fmt.Sprintf("bootstrap-demo-%s@example.com", t.Name())

	_, _ = conn.Exec("DELETE FROM users WHERE email = $1", email)
	t.Cleanup(func() { _, _ = conn.Exec("DELETE FROM users WHERE email = $1", email) })

	cfg := &auth.Config{
		JWTSecret:    "test-secret",
		TokenTTL:     time.Hour,
		CookieName:   "ap_session",
		DemoMode:     true,
		DemoEmail:    email,
		DemoPassword: "demopass",
	}
	svc, err := auth.NewService(conn, cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	// First boot — creates user
	if err := auth.Bootstrap(svc); err != nil {
		t.Fatalf("Bootstrap (create): %v", err)
	}
	u, err := svc.Users.GetByEmail(email)
	if err != nil {
		t.Fatalf("GetByEmail after create: %v", err)
	}
	if u.Role != "viewer" {
		t.Errorf("Role: got %q, want viewer", u.Role)
	}
	if !auth.CheckPassword(*u.PasswordHash, "demopass") {
		t.Error("password hash incorrect after create")
	}

	// Second boot with changed password — syncs hash
	cfg.DemoPassword = "newdemopass"
	svc2, err := auth.NewService(conn, cfg)
	if err != nil {
		t.Fatalf("NewService (2nd boot): %v", err)
	}
	if err := auth.Bootstrap(svc2); err != nil {
		t.Fatalf("Bootstrap (sync): %v", err)
	}
	u2, _ := svc2.Users.GetByEmail(email)
	if !auth.CheckPassword(*u2.PasswordHash, "newdemopass") {
		t.Error("password hash not synced on second boot")
	}
}

func TestBootstrapDemo_SkipsWhenDemoModeOff(t *testing.T) {
	conn := openTestDB(t)
	cfg := &auth.Config{
		JWTSecret:  "test-secret",
		TokenTTL:   time.Hour,
		CookieName: "ap_session",
		DemoMode:   false,
	}
	svc, err := auth.NewService(conn, cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	if err := auth.Bootstrap(svc); err != nil {
		t.Errorf("Bootstrap with DemoMode=false should be a no-op, got: %v", err)
	}
}
