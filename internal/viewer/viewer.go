// Package viewer provides EAM analytical views generated from SQL queries
// against the workspace tables. Each view corresponds to a named report
// equivalent to an Essential EAM viewer portal page.
package viewer

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/viewer/views"
)

// View is the result of a tabular EAM view query.
type View struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
}

// Registry maps view names to their query functions.
type Registry struct {
	db    *sql.DB
	views map[string]queryFn
}

type queryFn func(db *sql.DB, workspaceID uuid.UUID) (*View, error)

// NewRegistry creates a Registry with all built-in EAM views registered.
func NewRegistry(db *sql.DB) *Registry {
	r := &Registry{db: db, views: make(map[string]queryFn)}

	r.register("element-catalogue", func(db *sql.DB, id uuid.UUID) (*View, error) {
		name, cols, rows, err := views.ElementCatalogue(db, id)
		return &View{Name: name, Columns: cols, Rows: rows}, err
	})
	r.register("capability-tree", func(db *sql.DB, id uuid.UUID) (*View, error) {
		name, cols, rows, err := views.CapabilityTree(db, id)
		return &View{Name: name, Columns: cols, Rows: rows}, err
	})
	r.register("application-landscape", func(db *sql.DB, id uuid.UUID) (*View, error) {
		name, cols, rows, err := views.ApplicationLandscape(db, id)
		return &View{Name: name, Columns: cols, Rows: rows}, err
	})
	r.register("application-catalogue", func(db *sql.DB, id uuid.UUID) (*View, error) {
		name, cols, rows, err := views.ApplicationCatalogue(db, id)
		return &View{Name: name, Columns: cols, Rows: rows}, err
	})
	r.register("technology-catalogue", func(db *sql.DB, id uuid.UUID) (*View, error) {
		name, cols, rows, err := views.TechnologyCatalogue(db, id)
		return &View{Name: name, Columns: cols, Rows: rows}, err
	})
	return r
}

func (r *Registry) register(name string, fn queryFn) {
	r.views[name] = fn
}

// Execute runs the named view for the given workspace.
func (r *Registry) Execute(name string, workspaceID uuid.UUID) (*View, error) {
	fn, ok := r.views[name]
	if !ok {
		return nil, fmt.Errorf("unknown view: %q", name)
	}
	v, err := fn(r.db, workspaceID)
	if v != nil && v.Rows == nil {
		v.Rows = [][]any{}
	}
	return v, err
}

// ApplicationDependencyGraph returns the graph data for the dependency viewer.
func (r *Registry) ApplicationDependencyGraph(workspaceID uuid.UUID) (*views.ApplicationDependencyGraph, error) {
	return views.ApplicationDependency(r.db, workspaceID)
}

// CapabilityTreeData returns the hierarchical capability tree nodes.
func (r *Registry) CapabilityTreeData(workspaceID uuid.UUID) ([]views.CapabilityNode, error) {
	return views.CapabilityTreeData(r.db, workspaceID)
}

// IntegrationMap returns the application integration topology graph.
func (r *Registry) IntegrationMap(workspaceID uuid.UUID) (*views.IntegrationGraph, error) {
	return views.IntegrationMap(r.db, workspaceID)
}

// ApplicationDashboard returns all property distributions for the Application layer,
// optionally filtered by capability name ("" or "all" = no filter).
func (r *Registry) ApplicationDashboard(workspaceID uuid.UUID, capability string) (*views.ApplicationDashboardData, error) {
	return views.ApplicationDashboard(r.db, workspaceID, capability)
}

// ApplicationLandscapeMap returns the L1 → L2 → apps hierarchy for the landscape map view.
func (r *Registry) ApplicationLandscapeMap(workspaceID uuid.UUID) (*views.ApplicationLandscapeMapData, error) {
	return views.ApplicationLandscapeMap(r.db, workspaceID)
}

// List returns the names of all registered tabular views, sorted.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.views))
	for name := range r.views {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
