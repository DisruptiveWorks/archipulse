package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
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

// RequireOrgAdmin is middleware that additionally checks the user has org_role == "admin".
func (svc *Service) RequireOrgAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := ClaimsFromCtx(r.Context())
		if claims == nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		if claims.OrgRole != "admin" {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireWorkspaceAccess returns middleware that checks the caller holds at least
// minRole in the workspace identified by the chi URL param "id".
func (svc *Service) RequireWorkspaceAccess(minRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromCtx(r.Context())
			if claims == nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			// Org admins bypass workspace-level checks.
			if claims.OrgRole == "admin" {
				next.ServeHTTP(w, r)
				return
			}
			wsID := chi.URLParam(r, "id")
			if wsID == "" {
				wsID = chi.URLParam(r, "wsID")
			}
			if wsID == "" {
				http.Error(w, `{"error":"workspace id required"}`, http.StatusBadRequest)
				return
			}
			var ok bool
			var err error
			switch minRole {
			case RoleOwner:
				ok, err = svc.Enforcer.CanManage(claims.UserID, wsID)
			case RoleEditor:
				ok, err = svc.Enforcer.CanEdit(claims.UserID, wsID)
			default:
				ok, err = svc.Enforcer.CanView(claims.UserID, wsID)
			}
			if err != nil && !errors.Is(err, ErrNoMembership) {
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
				return
			}
			if !ok {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ClaimsFromCtx retrieves auth Claims stored by RequireAuth.
// Returns nil when the request was not authenticated.
func ClaimsFromCtx(ctx context.Context) *Claims {
	c, _ := ctx.Value(claimsKey).(*Claims)
	return c
}
