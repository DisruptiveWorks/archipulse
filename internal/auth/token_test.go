package auth_test

import (
	"testing"
	"time"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/google/uuid"
)

func testConfig() *auth.Config {
	return &auth.Config{
		JWTSecret: "test-secret",
		TokenTTL:  time.Hour,
	}
}

func TestIssueToken_RoundTrip(t *testing.T) {
	cfg := testConfig()
	id := uuid.New().String()

	raw, err := auth.IssueToken(cfg, id, "user@example.com", "viewer")
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}

	claims, err := auth.ParseToken(cfg, raw)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}

	if claims.UserID != id {
		t.Errorf("UserID: got %q, want %q", claims.UserID, id)
	}
	if claims.Email != "user@example.com" {
		t.Errorf("Email: got %q", claims.Email)
	}
	if claims.Role != "viewer" {
		t.Errorf("Role: got %q", claims.Role)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	cfg := testConfig()
	raw, _ := auth.IssueToken(cfg, uuid.New().String(), "u@x.com", "admin")

	wrongCfg := &auth.Config{JWTSecret: "wrong-secret", TokenTTL: time.Hour}
	if _, err := auth.ParseToken(wrongCfg, raw); err == nil {
		t.Error("expected error parsing token with wrong secret")
	}
}

func TestParseToken_Expired(t *testing.T) {
	cfg := &auth.Config{JWTSecret: "test-secret", TokenTTL: -time.Second}
	raw, _ := auth.IssueToken(cfg, uuid.New().String(), "u@x.com", "viewer")

	parseCfg := testConfig()
	if _, err := auth.ParseToken(parseCfg, raw); err == nil {
		t.Error("expected error parsing expired token")
	}
}

func TestParseToken_Malformed(t *testing.T) {
	if _, err := auth.ParseToken(testConfig(), "not.a.jwt"); err == nil {
		t.Error("expected error parsing malformed token")
	}
}

func TestParseToken_Empty(t *testing.T) {
	if _, err := auth.ParseToken(testConfig(), ""); err == nil {
		t.Error("expected error parsing empty token")
	}
}
