// Package api implements the ArchiPulse REST API handlers.
package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

// NewRouter builds and returns the root HTTP router with all routes registered.
func NewRouter(db *sql.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	r.Route("/api/v1", func(r chi.Router) {
		registerWorkspaceRoutes(r, workspace.NewStore(db))
		registerElementRoutes(r, db)
		registerRelationshipRoutes(r, db)
		registerDiagramRoutes(r, db)
		registerExportRoutes(r, db)
		registerImportRoutes(r, db)
		registerViewerRoutes(r, db)
	})

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
