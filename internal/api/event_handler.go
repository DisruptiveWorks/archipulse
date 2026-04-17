package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/audit"
	"github.com/DisruptiveWorks/archipulse/internal/auth"
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
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	events, err := h.store.List(wsID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	if events == nil {
		events = []audit.Event{}
	}
	respondJSON(w, http.StatusOK, events)
}

func registerEventRoutes(r chi.Router, store *audit.Store, svc *auth.Service) {
	h := &eventHandler{store: store}
	r.With(svc.RequireWorkspaceAccess(auth.RoleViewer)).
		Get("/workspaces/{id}/events", h.list)
}
