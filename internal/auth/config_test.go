package auth_test

import (
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
)

func TestOIDCEnabled_AllSet(t *testing.T) {
	cfg := &auth.Config{
		OIDCIssuerURL:    "https://issuer.example.com",
		OIDCClientID:     "client-id",
		OIDCClientSecret: "client-secret",
	}
	if !cfg.OIDCEnabled() {
		t.Error("expected OIDCEnabled true when all fields set")
	}
}

func TestOIDCEnabled_MissingIssuer(t *testing.T) {
	cfg := &auth.Config{
		OIDCClientID:     "client-id",
		OIDCClientSecret: "client-secret",
	}
	if cfg.OIDCEnabled() {
		t.Error("expected OIDCEnabled false when issuer missing")
	}
}

func TestOIDCEnabled_MissingClientID(t *testing.T) {
	cfg := &auth.Config{
		OIDCIssuerURL:    "https://issuer.example.com",
		OIDCClientSecret: "client-secret",
	}
	if cfg.OIDCEnabled() {
		t.Error("expected OIDCEnabled false when client ID missing")
	}
}

func TestOIDCEnabled_MissingSecret(t *testing.T) {
	cfg := &auth.Config{
		OIDCIssuerURL: "https://issuer.example.com",
		OIDCClientID:  "client-id",
	}
	if cfg.OIDCEnabled() {
		t.Error("expected OIDCEnabled false when client secret missing")
	}
}

func TestConfigFromEnv_MissingJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	if _, err := auth.ConfigFromEnv(); err == nil {
		t.Error("expected error when JWT_SECRET is empty")
	}
}

func TestConfigFromEnv_Defaults(t *testing.T) {
	t.Setenv("JWT_SECRET", "mysecret")
	t.Setenv("DEMO_MODE", "")
	t.Setenv("OIDC_ISSUER_URL", "")
	t.Setenv("OIDC_CLIENT_ID", "")
	t.Setenv("OIDC_CLIENT_SECRET", "")

	cfg, err := auth.ConfigFromEnv()
	if err != nil {
		t.Fatalf("ConfigFromEnv: %v", err)
	}
	if cfg.JWTSecret != "mysecret" {
		t.Errorf("JWTSecret: got %q", cfg.JWTSecret)
	}
	if cfg.CookieName != "ap_session" {
		t.Errorf("CookieName: got %q", cfg.CookieName)
	}
	if cfg.OIDCEnabled() {
		t.Error("expected OIDC disabled by default")
	}
	if cfg.DemoMode {
		t.Error("expected DemoMode false by default")
	}
}

func TestConfigFromEnv_DemoModeRequiresBothVars(t *testing.T) {
	t.Setenv("JWT_SECRET", "mysecret")
	t.Setenv("DEMO_MODE", "true")
	t.Setenv("DEMO_EMAIL", "demo@example.com")
	t.Setenv("DEMO_PASSWORD", "") // missing

	cfg, err := auth.ConfigFromEnv()
	if err != nil {
		t.Fatalf("ConfigFromEnv: %v", err)
	}
	if cfg.DemoMode {
		t.Error("expected DemoMode false when DEMO_PASSWORD is empty")
	}
}
