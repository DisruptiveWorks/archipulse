package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/audit"
	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/DisruptiveWorks/archipulse/internal/exporter"
	"github.com/DisruptiveWorks/archipulse/internal/parser"
	"github.com/DisruptiveWorks/archipulse/internal/snapshot"
)

type snapshotHandler struct {
	db    *sql.DB
	store *snapshot.Store
	audit *audit.Store
}

func (h *snapshotHandler) list(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, errorf("invalid workspace id"))
		return
	}
	snaps, err := h.store.List(wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	if snaps == nil {
		snaps = []snapshot.Snapshot{}
	}
	respondJSON(w, http.StatusOK, snaps)
}

func (h *snapshotHandler) create(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, errorf("invalid workspace id"))
		return
	}
	var body struct {
		Label string `json:"label"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	claims := auth.ClaimsFromCtx(r.Context())
	snap, err := takeSnapshot(h.db, h.store, wsID, claims.UserID, claims.Email, body.Label, snapshot.TriggerManual)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	_ = h.audit.Record(audit.RecordParams{
		WorkspaceID: wsID,
		UserID:      claims.UserID,
		UserEmail:   claims.Email,
		Action:      audit.ActionCreateSnapshot,
		EntityType:  audit.EntitySnapshot,
		EntityID:    snap.ID.String(),
		EntityName:  snap.Label,
	})

	respondJSON(w, http.StatusCreated, snap)
}

func (h *snapshotHandler) delete(w http.ResponseWriter, r *http.Request) {
	_, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, errorf("invalid workspace id"))
		return
	}
	snapID, err := uuid.Parse(chi.URLParam(r, "snapID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, errorf("invalid snapshot id"))
		return
	}
	if err := h.store.Delete(snapID); errors.Is(err, snapshot.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *snapshotHandler) restore(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, errorf("invalid workspace id"))
		return
	}
	snapID, err := uuid.Parse(chi.URLParam(r, "snapID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, errorf("invalid snapshot id"))
		return
	}

	snap, err := h.store.Get(snapID)
	if errors.Is(err, snapshot.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	// Before restoring, take an auto-snapshot of current state.
	claims := auth.ClaimsFromCtx(r.Context())
	_, _ = takeSnapshot(h.db, h.store, wsID, claims.UserID, claims.Email, "", snapshot.TriggerManual)

	m, err := parser.ParseAOEF(strings.NewReader(snap.Payload))
	if err != nil {
		respondError(w, http.StatusInternalServerError, errorf("parse snapshot: %v", err))
		return
	}
	if _, err := ImportModel(h.db, wsID, m); err != nil {
		respondError(w, http.StatusInternalServerError, errorf("restore failed: %v", err))
		return
	}

	_ = h.audit.Record(audit.RecordParams{
		WorkspaceID: wsID,
		UserID:      claims.UserID,
		UserEmail:   claims.Email,
		Action:      audit.ActionRestoreSnapshot,
		EntityType:  audit.EntitySnapshot,
		EntityID:    snap.ID.String(),
		EntityName:  snap.Label,
	})

	w.WriteHeader(http.StatusNoContent)
}

// takeSnapshot exports the current workspace model as AOEF XML and stores it.
func takeSnapshot(db *sql.DB, store *snapshot.Store, wsID uuid.UUID, userID, userEmail, label, trigger string) (*snapshot.Snapshot, error) {
	m, err := exporter.LoadModel(db, wsID)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := exporter.WriteAOEF(&buf, m); err != nil {
		return nil, err
	}
	return store.Create(wsID, userID, userEmail, label, trigger, buf.String())
}

func registerSnapshotRoutes(r chi.Router, db *sql.DB, store *snapshot.Store, auditStore *audit.Store, svc *auth.Service) {
	h := &snapshotHandler{db: db, store: store, audit: auditStore}
	view := svc.RequireWorkspaceAccess(auth.RoleViewer)
	edit := svc.RequireWorkspaceAccess(auth.RoleEditor)
	own := svc.RequireWorkspaceAccess(auth.RoleOwner)
	r.With(view).Get("/workspaces/{id}/snapshots", h.list)
	r.With(edit).Post("/workspaces/{id}/snapshots", h.create)
	r.With(own).Delete("/workspaces/{id}/snapshots/{snapID}", h.delete)
	r.With(edit).Post("/workspaces/{id}/snapshots/{snapID}/restore", h.restore)
}
