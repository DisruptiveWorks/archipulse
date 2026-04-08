package auth

import (
	"context"
	"net/http"
)

type contextKey string

const claimsKey contextKey = "auth_claims"

// RequireAuth is HTTP middleware that enforces a valid ap_session cookie.
// On success it stores the Claims in the request context.
func (svc *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(svc.Cfg.CookieName)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		claims, err := ParseToken(svc.Cfg, cookie.Value)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole is middleware that additionally enforces a Casbin RBAC check.
func (svc *Service) RequireRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := ClaimsFromCtx(r.Context())
		if claims == nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		allowed, err := svc.Enforcer.Allow(claims.Role, r.URL.Path, r.Method)
		if err != nil || !allowed {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ClaimsFromCtx retrieves auth Claims stored by RequireAuth.
// Returns nil when the request was not authenticated.
func ClaimsFromCtx(ctx context.Context) *Claims {
	c, _ := ctx.Value(claimsKey).(*Claims)
	return c
}
