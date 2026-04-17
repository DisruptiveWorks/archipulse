package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
)

type userHandler struct {
	users *auth.UserStore
}

// lookupByEmail handles GET /users?email=... — returns minimal user info for membership invitations.
// Accessible to any authenticated user (they need to know a UUID to add a member).
func (h *userHandler) lookupByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		respondError(w, http.StatusBadRequest, errorf("email query param required"))
		return
	}
	u, err := h.users.GetByEmail(email)
	if errors.Is(err, auth.ErrNotFound) {
		respondError(w, http.StatusNotFound, errorf("user not found"))
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{
		"id":    u.ID.String(),
		"email": u.Email,
	})
}

// registerUserRoutes mounts user endpoints on the given router.
func registerUserRoutes(r chi.Router, svc *auth.Service) {
	h := &userHandler{users: svc.Users}
	r.Get("/users/lookup", h.lookupByEmail)
}
