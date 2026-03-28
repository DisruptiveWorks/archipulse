package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

type workspaceHandler struct {
	store *workspace.Store
}

func (h *workspaceHandler) list(w http.ResponseWriter, r *http.Request) {
	wss, err := h.store.List()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, wss)
}

func (h *workspaceHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	ws, err := h.store.Get(id)
	if errors.Is(err, workspace.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, ws)
}

func (h *workspaceHandler) create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Purpose     string `json:"purpose"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	if body.Name == "" || body.Purpose == "" {
		respondError(w, http.StatusBadRequest, errorf("name and purpose are required"))
		return
	}
	ws, err := h.store.Create(body.Name, body.Purpose, body.Description)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusCreated, ws)
}

func (h *workspaceHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	var body struct {
		Name        string `json:"name"`
		Purpose     string `json:"purpose"`
		Description string `json:"description"`
		Version     int    `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	ws, err := h.store.Update(id, body.Name, body.Purpose, body.Description, body.Version)
	if errors.Is(err, workspace.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	}
	if errors.Is(err, workspace.ErrConflict) {
		respondError(w, http.StatusConflict, err)
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, ws)
}

func (h *workspaceHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.store.Delete(id); errors.Is(err, workspace.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// registerWorkspaceRoutes mounts workspace endpoints on the given router.
func registerWorkspaceRoutes(r chi.Router, store *workspace.Store) {
	h := &workspaceHandler{store: store}
	r.Get("/workspaces", h.list)
	r.Post("/workspaces", h.create)
	r.Get("/workspaces/{id}", h.get)
	r.Put("/workspaces/{id}", h.update)
	r.Delete("/workspaces/{id}", h.delete)
}

func parseUUID(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, param))
}
