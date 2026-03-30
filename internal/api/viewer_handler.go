package api

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

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

func registerViewerRoutes(r chi.Router, db *sql.DB) {
	h := &viewerHandler{registry: viewer.NewRegistry(db)}
	r.Get("/workspaces/{id}/views", h.listViews)
	r.Get("/workspaces/{id}/views/capability-tree/tree", h.getCapabilityTree)
	r.Get("/workspaces/{id}/views/application-dependency/graph", h.getDependencyGraph)
	r.Get("/workspaces/{id}/views/integration-map/graph", h.getIntegrationMap)
	r.Get("/workspaces/{id}/views/{view}", h.getView)
}
