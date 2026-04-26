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

	"github.com/DisruptiveWorks/archipulse/internal/audit"
	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/DisruptiveWorks/archipulse/internal/events"
	"github.com/DisruptiveWorks/archipulse/internal/snapshot"
	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

// NewRouter builds and returns the root HTTP router with all routes registered.
// Pass an embed.FS as the optional last argument to also serve the frontend SPA.
func NewRouter(db *sql.DB, svc *auth.Service, oidc *auth.OIDCProvider, version string, static ...embed.FS) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": version})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "application/json"))

		// Public auth endpoints — no authentication required.
		svc.RegisterRoutes(r, oidc)

		// Protected API: require a valid session.
		r.Group(func(r chi.Router) {
			r.Use(svc.RequireAuth)
			auditStore := audit.NewStore(db)
			snapStore := snapshot.NewStore(db)

			// Event bus: async dispatcher; audit writes are subscribers.
			bus := events.New(256)
			bus.Subscribe(func(e events.Event) {
				action := map[events.Kind]string{
					events.KindSavedViewCreated: audit.ActionCreate,
					events.KindSavedViewUpdated: audit.ActionUpdate,
					events.KindSavedViewDeleted: audit.ActionDelete,
				}[e.Kind]
				if action == "" {
					return
				}
				_ = auditStore.Record(audit.RecordParams{
					WorkspaceID: e.WorkspaceID,
					UserID:      e.ActorID,
					UserEmail:   e.ActorEmail,
					Action:      action,
					EntityType:  audit.EntitySavedView,
					EntityID:    e.ObjectID,
					EntityName:  e.ObjectName,
				})
			})

			registerWorkspaceRoutes(r, workspace.NewStore(db), svc)
			registerMembershipRoutes(r, svc, auditStore)
			registerUserRoutes(r, svc)
			registerElementRoutes(r, db, svc, auditStore)
			registerRelationshipRoutes(r, db, svc, auditStore)
			registerDiagramRoutes(r, db, svc, auditStore)
			registerFolderRoutes(r, db, svc)
			registerExportRoutes(r, db, svc)
			registerImportRoutes(r, db, svc, auditStore, snapStore)
			registerViewerRoutes(r, db, svc)
			registerSavedViewsRoutes(r, db, svc, bus)
			registerEventRoutes(r, auditStore, svc)
			registerSnapshotRoutes(r, db, snapStore, auditStore, svc)
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
