// Package auth implements authentication and authorization for ArchiPulse.
// It provides JWT-based sessions, bcrypt passwords, optional OIDC login,
// and Casbin global RBAC (admin / architect / viewer).
package auth

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
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

	// Demo mode — shows pre-filled credentials on the login page.
	DemoMode     bool
	DemoEmail    string
	DemoPassword string

	// OIDC (all three must be set to enable)
	OIDCIssuerURL    string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string

	// OIDC role mapping — which claim carries the role/group values and
	// which values in that claim map to the "admin" org role.
	// Example: OIDCRolesClaim="groups", OIDCAdminValues=["archipulse-admin"]
	// For local Dex with staticPasswords, use OIDCRolesClaim="email" and
	// OIDCAdminValues=["admin@archipulse.org"] as a practical workaround.
	OIDCRolesClaim  string
	OIDCAdminValues []string
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

	demoEmail := os.Getenv("DEMO_EMAIL")
	demoPassword := os.Getenv("DEMO_PASSWORD")

	return &Config{
		JWTSecret:         secret,
		TokenTTL:          ttl,
		CookieName:        "ap_session",
		CookieSecure:      os.Getenv("COOKIE_SECURE") != "false",
		BootstrapEmail:    os.Getenv("ARCHIPULSE_BOOTSTRAP_EMAIL"),
		BootstrapPassword: os.Getenv("ARCHIPULSE_BOOTSTRAP_PASSWORD"),
		DemoMode:          os.Getenv("DEMO_MODE") == "true" && demoEmail != "" && demoPassword != "",
		DemoEmail:         demoEmail,
		DemoPassword:      demoPassword,
		OIDCIssuerURL:     os.Getenv("OIDC_ISSUER_URL"),
		OIDCClientID:      os.Getenv("OIDC_CLIENT_ID"),
		OIDCClientSecret:  os.Getenv("OIDC_CLIENT_SECRET"),
		OIDCRedirectURL:   os.Getenv("OIDC_REDIRECT_URL"),
		OIDCRolesClaim:    oidcRolesClaim(),
		OIDCAdminValues:   oidcAdminValues(),
	}, nil
}

// oidcRolesClaim returns the configured claim name, defaulting to "groups".
func oidcRolesClaim() string {
	if v := os.Getenv("OIDC_ROLES_CLAIM"); v != "" {
		return v
	}
	return "groups"
}

// oidcAdminValues returns the list of claim values that map to org_role "admin".
func oidcAdminValues() []string {
	v := os.Getenv("OIDC_ADMIN_VALUES")
	if v == "" {
		return nil
	}
	var out []string
	for _, s := range strings.Split(v, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
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
