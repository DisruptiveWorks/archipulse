package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/element"
	"github.com/DisruptiveWorks/archipulse/internal/parser"
)

type elementHandler struct {
	store *element.Store
}

func (h *elementHandler) list(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "wsID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	elems, err := h.store.List(wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, elems)
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
	respondJSON(w, http.StatusOK, e)
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
	w.WriteHeader(http.StatusNoContent)
}

func registerElementRoutes(r chi.Router, db *sql.DB) {
	h := &elementHandler{store: element.NewStore(db)}
	r.Get("/workspaces/{wsID}/elements", h.list)
	r.Post("/workspaces/{wsID}/elements", h.create)
	r.Get("/workspaces/{wsID}/elements/{id}", h.get)
	r.Put("/workspaces/{wsID}/elements/{id}", h.update)
	r.Delete("/workspaces/{wsID}/elements/{id}", h.delete)
}
