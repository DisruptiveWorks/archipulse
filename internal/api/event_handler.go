package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/audit"
	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/DisruptiveWorks/archipulse/internal/pagination"
)

type eventHandler struct {
	store *audit.Store
}

func (h *eventHandler) list(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, errorf("invalid workspace id"))
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

func registerEventRoutes(r chi.Router, store *audit.Store, svc *auth.Service) {
	h := &eventHandler{store: store}
	r.With(svc.RequireWorkspaceAccess(auth.RoleViewer)).
		Get("/workspaces/{id}/events", h.list)
}
