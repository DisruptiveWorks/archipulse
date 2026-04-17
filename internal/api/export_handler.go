package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/DisruptiveWorks/archipulse/internal/exporter"
	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

type exportHandler struct {
	db *sql.DB
}

func (h *exportHandler) exportAOEF(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	m, err := exporter.LoadModel(h.db, id)
	if errors.Is(err, workspace.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="model.xml"`)
	if err := exporter.WriteAOEF(w, m); err != nil {
		// Headers already sent — log only.
		_ = err
	}
}

func (h *exportHandler) exportAJX(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	m, err := exporter.LoadModel(h.db, id)
	if errors.Is(err, workspace.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="model.ajx"`)
	if err := exporter.WriteAJX(w, m); err != nil {
		_ = err
	}
}

func registerExportRoutes(r chi.Router, db *sql.DB, svc *auth.Service) {
	h := &exportHandler{db: db}
	view := svc.RequireWorkspaceAccess(auth.RoleViewer)
	r.With(view).Get("/workspaces/{id}/export/aoef", h.exportAOEF)
	r.With(view).Get("/workspaces/{id}/export/ajx", h.exportAJX)
}
