package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCProvider wraps the go-oidc provider and oauth2 config.
type OIDCProvider struct {
	provider *gooidc.Provider
	oauth2   oauth2.Config
	verifier *gooidc.IDTokenVerifier
}

// NewOIDCProvider initialises the OIDC provider from Config.
// Returns nil, nil when OIDC is not configured.
func NewOIDCProvider(ctx context.Context, cfg *Config) (*OIDCProvider, error) {
	if !cfg.OIDCEnabled() {
		return nil, nil
	}

	p, err := gooidc.NewProvider(ctx, cfg.OIDCIssuerURL)
	if err != nil {
		return nil, fmt.Errorf("oidc provider: %w", err)
	}

	oa := oauth2.Config{
		ClientID:     cfg.OIDCClientID,
		ClientSecret: cfg.OIDCClientSecret,
		Endpoint:     p.Endpoint(),
		RedirectURL:  cfg.OIDCRedirectURL,
		Scopes:       []string{gooidc.ScopeOpenID, "email", "profile", "groups"},
	}

	verifier := p.Verifier(&gooidc.Config{ClientID: cfg.OIDCClientID})

	return &OIDCProvider{provider: p, oauth2: oa, verifier: verifier}, nil
}

// OIDCIdentity holds the verified identity extracted from an ID token.
type OIDCIdentity struct {
	Email       string
	ClaimValues []string // raw values extracted from the configured roles claim
}

// AuthURL returns the OIDC authorisation redirect URL and sets a state cookie.
func (op *OIDCProvider) AuthURL(w http.ResponseWriter, r *http.Request) string {
	state := randomState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})
	return op.oauth2.AuthCodeURL(state)
}

// ExchangeCode exchanges the callback code for an ID token.
// Returns an OIDCIdentity with the email and the raw values of the configured
// roles claim (e.g. groups, roles, email) so the caller can map them to an org role.
func (op *OIDCProvider) ExchangeCode(ctx context.Context, w http.ResponseWriter, r *http.Request, rolesClaim string) (*OIDCIdentity, error) {
	stateCookie, err := r.Cookie("oidc_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		return nil, fmt.Errorf("invalid state")
	}
	// Clear state cookie.
	http.SetCookie(w, &http.Cookie{Name: "oidc_state", MaxAge: -1, Path: "/"})

	code := r.URL.Query().Get("code")
	token, err := op.oauth2.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange: %w", err)
	}

	rawID, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("missing id_token")
	}

	idToken, err := op.verifier.Verify(ctx, rawID)
	if err != nil {
		return nil, fmt.Errorf("verify id_token: %w", err)
	}

	// Parse all claims into a raw map so we can extract any field.
	var allClaims map[string]json.RawMessage
	if err := idToken.Claims(&allClaims); err != nil {
		return nil, fmt.Errorf("parse id_token claims: %w", err)
	}

	// Extract email.
	email, err := stringClaim(allClaims, "email")
	if err != nil || email == "" {
		return nil, fmt.Errorf("id_token missing email claim")
	}

	// Extract the roles claim — supports both string and []string values.
	claimValues := extractClaimValues(allClaims, rolesClaim)

	return &OIDCIdentity{Email: email, ClaimValues: claimValues}, nil
}

// OrgRoleFromClaims maps OIDCIdentity claim values to an org role using the
// provided admin values list. Returns "admin" if any claim value matches an
// admin value, "member" otherwise.
func OrgRoleFromClaims(identity *OIDCIdentity, adminValues []string) string {
	if len(adminValues) == 0 {
		return "member"
	}
	adminSet := make(map[string]struct{}, len(adminValues))
	for _, v := range adminValues {
		adminSet[v] = struct{}{}
	}
	for _, v := range identity.ClaimValues {
		if _, ok := adminSet[v]; ok {
			return "admin"
		}
	}
	return "member"
}

// ── helpers ───────────────────────────────────────────────────────────────────

// extractClaimValues extracts a claim that may be a single string or a JSON
// array of strings from the raw claims map.
func extractClaimValues(claims map[string]json.RawMessage, key string) []string {
	raw, ok := claims[key]
	if !ok {
		return nil
	}
	// Try array first.
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr
	}
	// Fall back to single string.
	var s string
	if err := json.Unmarshal(raw, &s); err == nil && s != "" {
		return []string{s}
	}
	return nil
}

func stringClaim(claims map[string]json.RawMessage, key string) (string, error) {
	raw, ok := claims[key]
	if !ok {
		return "", fmt.Errorf("claim %q not found", key)
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", err
	}
	return s, nil
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
