package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RegisterRoutes attaches all /auth endpoints to the given mux prefix.
// The mux is expected to be a chi.Router already scoped to /api/v1.
func (svc *Service) RegisterRoutes(mux interface {
	Post(pattern string, handlerFn http.HandlerFunc)
	Get(pattern string, handlerFn http.HandlerFunc)
}, oidc *OIDCProvider) {
	mux.Post("/auth/login", svc.handleLogin)
	mux.Post("/auth/logout", svc.handleLogout)
	mux.Get("/auth/me", svc.handleMe)
	mux.Get("/auth/config", svc.handleConfig(oidc))

	if oidc != nil {
		mux.Get("/auth/oidc", svc.handleOIDCRedirect(oidc))
		mux.Get("/auth/oidc/callback", svc.handleOIDCCallback(oidc))
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (svc *Service) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     svc.Cfg.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   svc.Cfg.CookieSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(svc.Cfg.TokenTTL.Seconds()),
	})
}

func (svc *Service) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     svc.Cfg.CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   svc.Cfg.CookieSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

// ── Handlers ─────────────────────────────────────────────────────────────────

func (svc *Service) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, err := LoginLocal(svc, body.Email, body.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	svc.setSessionCookie(w, token)

	claims, _ := ParseToken(svc.Cfg, token)
	writeJSON(w, http.StatusOK, map[string]string{
		"email": claims.Email,
		"role":  claims.Role,
	})
}

func (svc *Service) handleLogout(w http.ResponseWriter, r *http.Request) {
	svc.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (svc *Service) handleMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(svc.Cfg.CookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	claims, err := ParseToken(svc.Cfg, cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"id":    claims.UserID,
		"email": claims.Email,
		"role":  claims.Role,
	})
}

func (svc *Service) handleConfig(oidc *OIDCProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"oidc_enabled": oidc != nil,
			"demo_mode":    svc.Cfg.DemoMode,
		}
		if svc.Cfg.DemoMode {
			resp["demo_email"] = svc.Cfg.DemoEmail
			resp["demo_password"] = svc.Cfg.DemoPassword
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func (svc *Service) handleOIDCRedirect(oidc *OIDCProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := oidc.AuthURL(w, r)
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func (svc *Service) handleOIDCCallback(oidc *OIDCProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, err := oidc.ExchangeCode(context.Background(), w, r)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("oidc: %v", err))
			return
		}

		u, err := svc.Users.GetByEmail(email)
		if err == ErrNotFound {
			// First OIDC login — assign admin if email matches bootstrap config,
			// otherwise provision as viewer.
			role := "viewer"
			if svc.Cfg.BootstrapEmail != "" && email == svc.Cfg.BootstrapEmail {
				role = "admin"
			}
			u, err = svc.Users.Create(email, "", role)
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "user lookup failed")
			return
		}

		// If the bootstrap admin logs in via OIDC and was previously provisioned
		// as viewer (e.g. before bootstrap config was set), promote to admin.
		if svc.Cfg.BootstrapEmail != "" && email == svc.Cfg.BootstrapEmail && u.Role != "admin" {
			if err := svc.Users.UpdateRole(u.ID.String(), "admin"); err == nil {
				u.Role = "admin"
			}
		}

		token, err := IssueToken(svc.Cfg, u.ID.String(), u.Email, u.Role)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "token issue failed")
			return
		}

		svc.setSessionCookie(w, token)
		// Redirect to SPA root after OIDC login.
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
