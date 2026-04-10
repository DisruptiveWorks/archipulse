package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/diagram"
)

type diagramHandler struct {
	store *diagram.Store
}

func (h *diagramHandler) list(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "wsID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	diags, err := h.store.List(wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, diags)
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

func registerDiagramRoutes(r chi.Router, db *sql.DB) {
	h := &diagramHandler{store: diagram.NewStore(db)}
	r.Get("/workspaces/{wsID}/diagrams", h.list)
	r.Post("/workspaces/{wsID}/diagrams", h.create)
	r.Get("/workspaces/{wsID}/diagrams/{id}", h.get)
	r.Get("/workspaces/{wsID}/diagrams/{id}/render", h.render)
	r.Put("/workspaces/{wsID}/diagrams/{id}", h.update)
	r.Delete("/workspaces/{wsID}/diagrams/{id}", h.delete)
}
