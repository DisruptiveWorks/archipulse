// Package api implements the ArchiPulse REST API handlers.
package api

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

// NewRouter builds and returns the root HTTP router with all routes registered.
// Pass an embed.FS as the optional second argument to also serve the frontend SPA.
func NewRouter(db *sql.DB, svc *auth.Service, oidc *auth.OIDCProvider, static ...embed.FS) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "application/json"))

		// Public auth endpoints — no authentication required.
		svc.RegisterRoutes(r, oidc)

		// Protected API: require a valid session + RBAC check.
		r.Group(func(r chi.Router) {
			r.Use(svc.RequireAuth)
			r.Use(svc.RequireRole)
			registerWorkspaceRoutes(r, workspace.NewStore(db))
			registerElementRoutes(r, db)
			registerRelationshipRoutes(r, db)
			registerDiagramRoutes(r, db)
			registerExportRoutes(r, db)
			registerImportRoutes(r, db)
			registerViewerRoutes(r, db)
		})
	})

	if len(static) > 0 {
		serveFrontend(r, static[0])
	}

	return r
}

// respondJSON writes v as JSON with the given status code.
func respondJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// respondError writes an error payload.
func respondError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
