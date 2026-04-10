package auth_test

import (
	"fmt"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
)

func TestLoginLocal_Success(t *testing.T) {
	conn := openTestDB(t)
	svc := newTestService(t, conn)

	email := fmt.Sprintf("login-test-%s@example.com", t.Name())
	hash, _ := auth.HashPassword("correctpassword")
	if _, err := svc.Users.Create(email, hash, "viewer"); err != nil {
		t.Fatalf("Create user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = conn.Exec("DELETE FROM users WHERE email = $1", email)
	})

	token, err := auth.LoginLocal(svc, email, "correctpassword")
	if err != nil {
		t.Fatalf("LoginLocal: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}

	// Token must be parseable and carry the right claims
	claims, err := auth.ParseToken(svc.Cfg, token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.Email != email {
		t.Errorf("claims.Email: got %q, want %q", claims.Email, email)
	}
	if claims.Role != "viewer" {
		t.Errorf("claims.Role: got %q, want viewer", claims.Role)
	}
}

func TestLoginLocal_WrongPassword(t *testing.T) {
	conn := openTestDB(t)
	svc := newTestService(t, conn)

	email := fmt.Sprintf("login-wrong-%s@example.com", t.Name())
	hash, _ := auth.HashPassword("correctpassword")
	if _, err := svc.Users.Create(email, hash, "viewer"); err != nil {
		t.Fatalf("Create user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = conn.Exec("DELETE FROM users WHERE email = $1", email)
	})

	if _, err := auth.LoginLocal(svc, email, "wrongpassword"); err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestLoginLocal_UnknownEmail(t *testing.T) {
	conn := openTestDB(t)
	svc := newTestService(t, conn)

	if _, err := auth.LoginLocal(svc, "nobody@nowhere.com", "password"); err == nil {
		t.Error("expected error for unknown email")
	}
}

func TestLoginLocal_OIDCUserHasNoPassword(t *testing.T) {
	conn := openTestDB(t)
	svc := newTestService(t, conn)

	// OIDC users have no password_hash (empty string → nil in Create)
	email := fmt.Sprintf("oidc-%s@example.com", t.Name())
	if _, err := svc.Users.Create(email, "", "viewer"); err != nil {
		t.Fatalf("Create oidc user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = conn.Exec("DELETE FROM users WHERE email = $1", email)
	})

	if _, err := auth.LoginLocal(svc, email, "anything"); err == nil {
		t.Error("expected error for OIDC user attempting local login")
	}
}
