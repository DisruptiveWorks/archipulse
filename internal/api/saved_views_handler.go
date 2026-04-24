package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/DisruptiveWorks/archipulse/internal/savedviews"
)

type savedViewsHandler struct {
	store *savedviews.Store
}

func (h *savedViewsHandler) list(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "wsID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	views, err := h.store.List(wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	if views == nil {
		views = []savedviews.SavedView{}
	}
	respondJSON(w, http.StatusOK, views)
}

func (h *savedViewsHandler) get(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "wsID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	svID, err := uuid.Parse(chi.URLParam(r, "svID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	sv, err := h.store.Get(wsID, svID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	if sv == nil {
		respondError(w, http.StatusNotFound, errorf("saved view not found"))
		return
	}
	respondJSON(w, http.StatusOK, sv)
}

type savedViewInput struct {
	ViewType string          `json:"view_type"`
	Name     string          `json:"name"`
	Filters  json.RawMessage `json:"filters"`
}

func (h *savedViewsHandler) create(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "wsID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	var in savedViewInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondError(w, http.StatusBadRequest, errorf("invalid body: %w", err))
		return
	}
	if in.Name == "" || in.ViewType == "" {
		respondError(w, http.StatusBadRequest, errorf("name and view_type are required"))
		return
	}
	sv, err := h.store.Create(wsID, in.ViewType, in.Name, in.Filters)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusCreated, sv)
}

func (h *savedViewsHandler) update(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "wsID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	svID, err := uuid.Parse(chi.URLParam(r, "svID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	var in savedViewInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondError(w, http.StatusBadRequest, errorf("invalid body: %w", err))
		return
	}
	if in.Name == "" {
		respondError(w, http.StatusBadRequest, errorf("name is required"))
		return
	}
	sv, err := h.store.Update(wsID, svID, in.Name, in.Filters)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	if sv == nil {
		respondError(w, http.StatusNotFound, errorf("saved view not found"))
		return
	}
	respondJSON(w, http.StatusOK, sv)
}

func (h *savedViewsHandler) delete(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "wsID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	svID, err := uuid.Parse(chi.URLParam(r, "svID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.store.Delete(wsID, svID); err != nil {
		if err.Error() == "not found" {
			respondError(w, http.StatusNotFound, err)
			return
		}
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func registerSavedViewsRoutes(r chi.Router, db *sql.DB, svc *auth.Service) {
	h := &savedViewsHandler{store: savedviews.NewStore(db)}
	viewer := svc.RequireWorkspaceAccess(auth.RoleViewer)
	editor := svc.RequireWorkspaceAccess(auth.RoleEditor)

	r.With(viewer).Get("/workspaces/{wsID}/saved-views", h.list)
	r.With(viewer).Get("/workspaces/{wsID}/saved-views/{svID}", h.get)
	r.With(editor).Post("/workspaces/{wsID}/saved-views", h.create)
	r.With(editor).Put("/workspaces/{wsID}/saved-views/{svID}", h.update)
	r.With(editor).Delete("/workspaces/{wsID}/saved-views/{svID}", h.delete)
}
