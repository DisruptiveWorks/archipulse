package api

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/DisruptiveWorks/archipulse/internal/viewer"
)

type viewerHandler struct {
	registry *viewer.Registry
}

// listViews returns all available view names for a workspace.
func (h *viewerHandler) listViews(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string][]string{"views": h.registry.List()})
}

// getView executes a named tabular view and returns its data.
func (h *viewerHandler) getView(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	viewName := strings.ToLower(chi.URLParam(r, "view"))

	result, err := h.registry.Execute(viewName, wsID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "unknown view") {
			respondError(w, http.StatusNotFound, err)
			return
		}
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

// getIntegrationMap returns the application integration topology graph.
func (h *viewerHandler) getIntegrationMap(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	graph, err := h.registry.IntegrationMap(wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, graph)
}

// getDependencyGraph returns the application dependency graph for Cytoscape.js.
func (h *viewerHandler) getDependencyGraph(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	graph, err := h.registry.ApplicationDependencyGraph(wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, graph)
}

// getCapabilityTree returns the hierarchical capability tree data.
func (h *viewerHandler) getCapabilityTree(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	nodes, err := h.registry.CapabilityTreeData(wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"nodes": nodes})
}

// getApplicationDashboard returns property distribution stats.
// Accepts optional ?capability=<name> query param to filter by capability.
func (h *viewerHandler) getApplicationDashboard(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	capability := r.URL.Query().Get("capability")
	data, err := h.registry.ApplicationDashboard(wsID, capability)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, data)
}

// getCapabilityLandscapeMap returns the L1 → L2 → apps hierarchy for the capability landscape view.
func (h *viewerHandler) getCapabilityLandscapeMap(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	data, err := h.registry.CapabilityLandscapeMap(wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, data)
}

// getApplicationLandscapeDomain returns apps grouped by domain property.
func (h *viewerHandler) getApplicationLandscapeDomain(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	data, err := h.registry.ApplicationLandscapeDomain(wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, data)
}

// getAppCatalogueEntries returns the rich application catalogue payload.
func (h *viewerHandler) getAppCatalogueEntries(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	data, err := h.registry.AppCatalogueEntries(wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, data)
}

// getTechCatalogueEntries returns the rich technology catalogue payload.
func (h *viewerHandler) getTechCatalogueEntries(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	data, err := h.registry.TechCatalogueEntries(wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, data)
}

// getAppElementDetail returns the rich detail panel data for a single application element.
func (h *viewerHandler) getAppElementDetail(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}
	appID := chi.URLParam(r, "appID")
	data, err := h.registry.AppElementDetail(wsID, appID)
	if err != nil {
		if strings.Contains(err.Error(), "element not found") {
			respondError(w, http.StatusNotFound, err)
			return
		}
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, data)
}

func registerViewerRoutes(r chi.Router, db *sql.DB, svc *auth.Service) {
	h := &viewerHandler{registry: viewer.NewRegistry(db)}
	view := svc.RequireWorkspaceAccess(auth.RoleViewer)
	r.With(view).Get("/workspaces/{id}/views", h.listViews)
	r.With(view).Get("/workspaces/{id}/views/capability-tree/tree", h.getCapabilityTree)
	r.With(view).Get("/workspaces/{id}/views/application-dependency/graph", h.getDependencyGraph)
	r.With(view).Get("/workspaces/{id}/views/integration-map/graph", h.getIntegrationMap)
	r.With(view).Get("/workspaces/{id}/views/application-dashboard/stats", h.getApplicationDashboard)
	r.With(view).Get("/workspaces/{id}/views/capability-landscape/map", h.getCapabilityLandscapeMap)
	r.With(view).Get("/workspaces/{id}/views/application-landscape/map", h.getApplicationLandscapeDomain)
	r.With(view).Get("/workspaces/{id}/views/application-catalogue/entries", h.getAppCatalogueEntries)
	r.With(view).Get("/workspaces/{id}/views/technology-catalogue/entries", h.getTechCatalogueEntries)
	r.With(view).Get("/workspaces/{id}/views/{view}", h.getView)
	r.With(view).Get("/workspaces/{id}/elements/{appID}/app-detail", h.getAppElementDetail)
}
