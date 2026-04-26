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
	"github.com/DisruptiveWorks/archipulse/internal/element"
	"github.com/DisruptiveWorks/archipulse/internal/pagination"
	"github.com/DisruptiveWorks/archipulse/internal/parser"
)

type elementHandler struct {
	store *element.Store
	audit *audit.Store
}

func (h *elementHandler) list(w http.ResponseWriter, r *http.Request) {
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

func (h *elementHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	e, err := h.store.Get(id)
	if errors.Is(err, element.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	props, err := h.store.ListProperties(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	// Group properties by source.
	bySource := make(map[string][]element.Property)
	for _, p := range props {
		bySource[p.Source] = append(bySource[p.Source], p)
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"element":    e,
		"properties": bySource,
	})
}

func (h *elementHandler) create(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "wsID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		SourceID      string `json:"source_id"`
		Type          string `json:"type"`
		Name          string `json:"name"`
		Documentation string `json:"documentation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	if body.SourceID == "" || body.Type == "" {
		respondError(w, http.StatusBadRequest, errorf("source_id and type are required"))
		return
	}
	layer := parser.ElementLayer(body.Type)
	e, err := h.store.Create(wsID, body.SourceID, body.Type, layer, body.Name, body.Documentation)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	if claims := auth.ClaimsFromCtx(r.Context()); claims != nil && h.audit != nil {
		_ = h.audit.Record(audit.RecordParams{
			WorkspaceID: wsID, UserID: claims.UserID, UserEmail: claims.Email,
			Action: audit.ActionCreate, EntityType: audit.EntityElement,
			EntityID: e.ID.String(), EntityName: e.Name,
		})
	}
	respondJSON(w, http.StatusCreated, e)
}

func (h *elementHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		Type          string `json:"type"`
		Name          string `json:"name"`
		Documentation string `json:"documentation"`
		Version       int    `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	e, err := h.store.Update(id, body.Type, parser.ElementLayer(body.Type), body.Name, body.Documentation, body.Version)
	if errors.Is(err, element.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	}
	if errors.Is(err, element.ErrConflict) {
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
			Action: audit.ActionUpdate, EntityType: audit.EntityElement,
			EntityID: e.ID.String(), EntityName: e.Name,
		})
	}
	respondJSON(w, http.StatusOK, e)
}

func (h *elementHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.store.Delete(id); errors.Is(err, element.ErrNotFound) {
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
			Action: audit.ActionDelete, EntityType: audit.EntityElement,
			EntityID: id.String(),
		})
	}
	w.WriteHeader(http.StatusNoContent)
}

func registerElementRoutes(r chi.Router, db *sql.DB, svc *auth.Service, auditStore *audit.Store) {
	h := &elementHandler{store: element.NewStore(db), audit: auditStore}
	view := svc.RequireWorkspaceAccess(auth.RoleViewer)
	edit := svc.RequireWorkspaceAccess(auth.RoleEditor)
	r.With(view).Get("/workspaces/{wsID}/elements", h.list)
	r.With(edit).Post("/workspaces/{wsID}/elements", h.create)
	r.With(view).Get("/workspaces/{wsID}/elements/{id}", h.get)
	r.With(edit).Put("/workspaces/{wsID}/elements/{id}", h.update)
	r.With(edit).Delete("/workspaces/{wsID}/elements/{id}", h.delete)
}
