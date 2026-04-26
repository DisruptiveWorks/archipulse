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
	"github.com/DisruptiveWorks/archipulse/internal/diagram"
	"github.com/DisruptiveWorks/archipulse/internal/pagination"
)

type diagramHandler struct {
	store *diagram.Store
	audit *audit.Store
}

func (h *diagramHandler) list(w http.ResponseWriter, r *http.Request) {
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

func (h *diagramHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	d, err := h.store.Get(id)
	if errors.Is(err, diagram.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, d)
}

func (h *diagramHandler) create(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "wsID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		SourceID      string          `json:"source_id"`
		Name          string          `json:"name"`
		Documentation string          `json:"documentation"`
		Layout        json.RawMessage `json:"layout"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	if body.SourceID == "" {
		respondError(w, http.StatusBadRequest, errorf("source_id is required"))
		return
	}
	d, err := h.store.Create(wsID, body.SourceID, body.Name, body.Documentation, body.Layout)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	if claims := auth.ClaimsFromCtx(r.Context()); claims != nil && h.audit != nil {
		_ = h.audit.Record(audit.RecordParams{
			WorkspaceID: wsID, UserID: claims.UserID, UserEmail: claims.Email,
			Action: audit.ActionCreate, EntityType: audit.EntityDiagram,
			EntityID: d.ID.String(), EntityName: d.Name,
		})
	}
	respondJSON(w, http.StatusCreated, d)
}

func (h *diagramHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		Name          string          `json:"name"`
		Documentation string          `json:"documentation"`
		Layout        json.RawMessage `json:"layout"`
		Version       int             `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	d, err := h.store.Update(id, body.Name, body.Documentation, body.Layout, body.Version)
	if errors.Is(err, diagram.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	}
	if errors.Is(err, diagram.ErrConflict) {
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
			Action: audit.ActionUpdate, EntityType: audit.EntityDiagram,
			EntityID: d.ID.String(), EntityName: d.Name,
		})
	}
	respondJSON(w, http.StatusOK, d)
}

func (h *diagramHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.store.Delete(id); errors.Is(err, diagram.ErrNotFound) {
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
			Action: audit.ActionDelete, EntityType: audit.EntityDiagram,
			EntityID: id.String(),
		})
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *diagramHandler) render(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	rd, err := h.store.Render(id)
	if errors.Is(err, diagram.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, rd)
}

func registerDiagramRoutes(r chi.Router, db *sql.DB, svc *auth.Service, auditStore *audit.Store) {
	h := &diagramHandler{store: diagram.NewStore(db), audit: auditStore}
	view := svc.RequireWorkspaceAccess(auth.RoleViewer)
	edit := svc.RequireWorkspaceAccess(auth.RoleEditor)
	r.With(view).Get("/workspaces/{wsID}/diagrams", h.list)
	r.With(edit).Post("/workspaces/{wsID}/diagrams", h.create)
	r.With(view).Get("/workspaces/{wsID}/diagrams/{id}", h.get)
	r.With(view).Get("/workspaces/{wsID}/diagrams/{id}/render", h.render)
	r.With(edit).Put("/workspaces/{wsID}/diagrams/{id}", h.update)
	r.With(edit).Delete("/workspaces/{wsID}/diagrams/{id}", h.delete)
}
