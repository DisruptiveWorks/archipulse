package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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
		Scopes:       []string{gooidc.ScopeOpenID, "email", "profile"},
	}

	verifier := p.Verifier(&gooidc.Config{ClientID: cfg.OIDCClientID})

	return &OIDCProvider{provider: p, oauth2: oa, verifier: verifier}, nil
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

// ExchangeCode exchanges the callback code for an ID token and returns the email.
func (op *OIDCProvider) ExchangeCode(ctx context.Context, w http.ResponseWriter, r *http.Request) (email string, err error) {
	stateCookie, err := r.Cookie("oidc_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		return "", fmt.Errorf("invalid state")
	}
	// Clear state cookie
	http.SetCookie(w, &http.Cookie{Name: "oidc_state", MaxAge: -1, Path: "/"})

	code := r.URL.Query().Get("code")
	token, err := op.oauth2.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("exchange: %w", err)
	}

	rawID, ok := token.Extra("id_token").(string)
	if !ok {
		return "", fmt.Errorf("missing id_token")
	}

	idToken, err := op.verifier.Verify(ctx, rawID)
	if err != nil {
		return "", fmt.Errorf("verify id_token: %w", err)
	}

	var claims struct {
		Email string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return "", fmt.Errorf("parse id_token claims: %w", err)
	}
	if claims.Email == "" {
		return "", fmt.Errorf("id_token missing email claim")
	}
	return claims.Email, nil
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
