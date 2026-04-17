package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/audit"
	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/DisruptiveWorks/archipulse/internal/relationship"
)

type relationshipHandler struct {
	store *relationship.Store
	audit *audit.Store
}

func (h *relationshipHandler) list(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "wsID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	rels, err := h.store.List(wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, rels)
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
	respondJSON(w, http.StatusCreated, rel)
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
	if claims := auth.ClaimsFromCtx(r.Context()); claims != nil && h.audit != nil {
		wsID, _ := uuid.Parse(chi.URLParam(r, "wsID"))
		_ = h.audit.Record(audit.RecordParams{
			WorkspaceID: wsID, UserID: claims.UserID, UserEmail: claims.Email,
			Action: audit.ActionUpdate, EntityType: audit.EntityRelationship,
			EntityID: rel.ID.String(), EntityName: rel.Name,
		})
	}
	respondJSON(w, http.StatusOK, rel)
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
	h := &relationshipHandler{store: relationship.NewStore(db), audit: auditStore}
	view := svc.RequireWorkspaceAccess(auth.RoleViewer)
	edit := svc.RequireWorkspaceAccess(auth.RoleEditor)
	r.With(view).Get("/workspaces/{wsID}/relationships", h.list)
	r.With(edit).Post("/workspaces/{wsID}/relationships", h.create)
	r.With(view).Get("/workspaces/{wsID}/relationships/{id}", h.get)
	r.With(edit).Put("/workspaces/{wsID}/relationships/{id}", h.update)
	r.With(edit).Delete("/workspaces/{wsID}/relationships/{id}", h.delete)
}
