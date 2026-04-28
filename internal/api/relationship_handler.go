package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/DisruptiveWorks/archipulse/internal/audit"
	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/DisruptiveWorks/archipulse/internal/pagination"
	"github.com/DisruptiveWorks/archipulse/internal/parser"
	"github.com/DisruptiveWorks/archipulse/internal/relationship"
)

type relationshipHandler struct {
	db    *sql.DB
	store *relationship.Store
	audit *audit.Store
}

// relResponse wraps a Relationship with any structural violations detected.
type relResponse struct {
	*relationship.Relationship
	Violations []parser.RelationshipViolation `json:"violations,omitempty"`
}

// elementTypesBySourceIDs returns a source_id→type map for the given IDs in a workspace.
func (h *relationshipHandler) elementTypesBySourceIDs(wsID uuid.UUID, sourceIDs []string) (map[string]string, error) {
	rows, err := h.db.Query(
		`SELECT source_id, type FROM elements WHERE workspace_id = $1 AND source_id = ANY($2)`,
		wsID, pq.Array(sourceIDs),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	m := make(map[string]string, len(sourceIDs))
	for rows.Next() {
		var sid, typ string
		if err := rows.Scan(&sid, &typ); err != nil {
			return nil, err
		}
		m[sid] = typ
	}
	return m, rows.Err()
}

// validateRel looks up element types and runs ArchiMate structural validation.
// Returns an empty slice when valid or element types are unknown.
func (h *relationshipHandler) validateRel(wsID uuid.UUID, sourceElement, targetElement, relType string) []parser.RelationshipViolation {
	types, err := h.elementTypesBySourceIDs(wsID, []string{sourceElement, targetElement})
	if err != nil {
		return nil // don't block on lookup failure
	}
	return parser.ValidateRelationship(types[sourceElement], types[targetElement], relType)
}

func (h *relationshipHandler) list(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "wsID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	p := parsePage(r)
	items, total, err := h.store.List(wsID, p)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, pagination.NewPage(items, total, p.Page, p.Limit))
}

func (h *relationshipHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	rel, err := h.store.Get(id)
	if errors.Is(err, relationship.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, rel)
}

func (h *relationshipHandler) create(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "wsID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		SourceID      string `json:"source_id"`
		Type          string `json:"type"`
		SourceElement string `json:"source_element"`
		TargetElement string `json:"target_element"`
		Name          string `json:"name"`
		Documentation string `json:"documentation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	if body.SourceID == "" || body.Type == "" || body.SourceElement == "" || body.TargetElement == "" {
		respondError(w, http.StatusBadRequest, errorf("source_id, type, source_element and target_element are required"))
		return
	}
	rel, err := h.store.Create(wsID, body.SourceID, body.Type, body.SourceElement, body.TargetElement, body.Name, body.Documentation)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	if claims := auth.ClaimsFromCtx(r.Context()); claims != nil && h.audit != nil {
		_ = h.audit.Record(audit.RecordParams{
			WorkspaceID: wsID, UserID: claims.UserID, UserEmail: claims.Email,
			Action: audit.ActionCreate, EntityType: audit.EntityRelationship,
			EntityID: rel.ID.String(), EntityName: rel.Name,
		})
	}
	respondJSON(w, http.StatusCreated, relResponse{
		Relationship: rel,
		Violations:   h.validateRel(wsID, body.SourceElement, body.TargetElement, body.Type),
	})
}

func (h *relationshipHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		Type          string `json:"type"`
		SourceElement string `json:"source_element"`
		TargetElement string `json:"target_element"`
		Name          string `json:"name"`
		Documentation string `json:"documentation"`
		Version       int    `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	rel, err := h.store.Update(id, body.Type, body.SourceElement, body.TargetElement, body.Name, body.Documentation, body.Version)
	if errors.Is(err, relationship.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	}
	if errors.Is(err, relationship.ErrConflict) {
		respondError(w, http.StatusConflict, err)
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	wsID, _ := uuid.Parse(chi.URLParam(r, "wsID"))
	if claims := auth.ClaimsFromCtx(r.Context()); claims != nil && h.audit != nil {
		_ = h.audit.Record(audit.RecordParams{
			WorkspaceID: wsID, UserID: claims.UserID, UserEmail: claims.Email,
			Action: audit.ActionUpdate, EntityType: audit.EntityRelationship,
			EntityID: rel.ID.String(), EntityName: rel.Name,
		})
	}
	respondJSON(w, http.StatusOK, relResponse{
		Relationship: rel,
		Violations:   h.validateRel(wsID, body.SourceElement, body.TargetElement, body.Type),
	})
}

func (h *relationshipHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.store.Delete(id); errors.Is(err, relationship.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	if claims := auth.ClaimsFromCtx(r.Context()); claims != nil && h.audit != nil {
		wsID, _ := uuid.Parse(chi.URLParam(r, "wsID"))
		_ = h.audit.Record(audit.RecordParams{
			WorkspaceID: wsID, UserID: claims.UserID, UserEmail: claims.Email,
			Action: audit.ActionDelete, EntityType: audit.EntityRelationship,
			EntityID: id.String(),
		})
	}
	w.WriteHeader(http.StatusNoContent)
}

func registerRelationshipRoutes(r chi.Router, db *sql.DB, svc *auth.Service, auditStore *audit.Store) {
	h := &relationshipHandler{db: db, store: relationship.NewStore(db), audit: auditStore}
	view := svc.RequireWorkspaceAccess(auth.RoleViewer)
	edit := svc.RequireWorkspaceAccess(auth.RoleEditor)
	r.With(view).Get("/workspaces/{wsID}/relationships", h.list)
	r.With(edit).Post("/workspaces/{wsID}/relationships", h.create)
	r.With(view).Get("/workspaces/{wsID}/relationships/{id}", h.get)
	r.With(edit).Put("/workspaces/{wsID}/relationships/{id}", h.update)
	r.With(edit).Delete("/workspaces/{wsID}/relationships/{id}", h.delete)
}
