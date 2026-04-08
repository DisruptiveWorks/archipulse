// Package auth implements authentication and authorization for ArchiPulse.
// It provides JWT-based sessions, bcrypt passwords, optional OIDC login,
// and Casbin global RBAC (admin / architect / viewer).
package auth

import (
	"database/sql"
	"fmt"
	"os"
	"time"
)

// Config holds all auth configuration derived from environment variables.
type Config struct {
	// JWT
	JWTSecret string
	TokenTTL  time.Duration

	// Cookie
	CookieName   string
	CookieSecure bool

	// Bootstrap admin
	BootstrapEmail    string
	BootstrapPassword string

	// OIDC (all three must be set to enable)
	OIDCIssuerURL   string
	OIDCClientID    string
	OIDCClientSecret string
	OIDCRedirectURL string
}

// OIDCEnabled reports whether OIDC is configured.
func (c *Config) OIDCEnabled() bool {
	return c.OIDCIssuerURL != "" && c.OIDCClientID != "" && c.OIDCClientSecret != ""
}

// ConfigFromEnv builds a Config from environment variables.
func ConfigFromEnv() (*Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	ttl := 24 * time.Hour
	if v := os.Getenv("JWT_TTL_HOURS"); v != "" {
		var h int
		if _, err := fmt.Sscanf(v, "%d", &h); err == nil && h > 0 {
			ttl = time.Duration(h) * time.Hour
		}
	}

	return &Config{
		JWTSecret:         secret,
		TokenTTL:          ttl,
		CookieName:        "ap_session",
		CookieSecure:      os.Getenv("COOKIE_SECURE") != "false",
		BootstrapEmail:    os.Getenv("ARCHIPULSE_BOOTSTRAP_EMAIL"),
		BootstrapPassword: os.Getenv("ARCHIPULSE_BOOTSTRAP_PASSWORD"),
		OIDCIssuerURL:     os.Getenv("OIDC_ISSUER_URL"),
		OIDCClientID:      os.Getenv("OIDC_CLIENT_ID"),
		OIDCClientSecret:  os.Getenv("OIDC_CLIENT_SECRET"),
		OIDCRedirectURL:   os.Getenv("OIDC_REDIRECT_URL"),
	}, nil
}

// Service bundles everything needed to handle auth.
type Service struct {
	DB       *sql.DB
	Cfg      *Config
	Users    *UserStore
	Enforcer *Enforcer
}

// NewService creates a new Service. Call Bootstrap after to ensure the first admin exists.
func NewService(db *sql.DB, cfg *Config) (*Service, error) {
	enf, err := NewEnforcer(db, cfg)
	if err != nil {
		return nil, fmt.Errorf("new enforcer: %w", err)
	}
	return &Service{
		DB:       db,
		Cfg:      cfg,
		Users:    NewUserStore(db),
		Enforcer: enf,
	}, nil
}
