package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
)

type membershipHandler struct {
	enforcer *auth.Enforcer
}

func (h *membershipHandler) list(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "id")
	members, err := h.enforcer.ListMembers(wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	if members == nil {
		members = []auth.WorkspaceMember{}
	}
	respondJSON(w, http.StatusOK, members)
}

func (h *membershipHandler) add(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "id")
	var body struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, errorf("invalid request body"))
		return
	}
	if body.UserID == "" || body.Role == "" {
		respondError(w, http.StatusBadRequest, errorf("user_id and role are required"))
		return
	}
	caller := auth.ClaimsFromCtx(r.Context())
	invitedBy := ""
	if caller != nil {
		invitedBy = caller.UserID
	}
	if err := h.enforcer.AddMember(wsID, body.UserID, body.Role, invitedBy); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *membershipHandler) updateRole(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userId")
	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, errorf("invalid request body"))
		return
	}
	if body.Role == "" {
		respondError(w, http.StatusBadRequest, errorf("role is required"))
		return
	}
	if err := h.enforcer.UpdateMemberRole(wsID, userID, body.Role); errors.Is(err, auth.ErrNoMembership) {
		respondError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *membershipHandler) remove(w http.ResponseWriter, r *http.Request) {
	wsID := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userId")
	if err := h.enforcer.RemoveMember(wsID, userID); errors.Is(err, auth.ErrNoMembership) {
		respondError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// registerMembershipRoutes mounts workspace membership endpoints.
// These routes are nested under /workspaces/{id} and require at least owner access.
func registerMembershipRoutes(r chi.Router, svc *auth.Service) {
	h := &membershipHandler{enforcer: svc.Enforcer}
	// GET /workspaces/{id}/members — viewers can list members
	r.With(svc.RequireWorkspaceAccess(auth.RoleViewer)).
		Get("/workspaces/{id}/members", h.list)
	// POST /workspaces/{id}/members — owners can add members
	r.With(svc.RequireWorkspaceAccess(auth.RoleOwner)).
		Post("/workspaces/{id}/members", h.add)
	// PUT /workspaces/{id}/members/{userId} — owners can update roles
	r.With(svc.RequireWorkspaceAccess(auth.RoleOwner)).
		Put("/workspaces/{id}/members/{userId}", h.updateRole)
	// DELETE /workspaces/{id}/members/{userId} — owners can remove members
	r.With(svc.RequireWorkspaceAccess(auth.RoleOwner)).
		Delete("/workspaces/{id}/members/{userId}", h.remove)
}
