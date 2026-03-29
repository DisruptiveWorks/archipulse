package views

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// ApplicationDependencyNode is a node in the dependency graph.
type ApplicationDependencyNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ApplicationDependencyEdge is an edge in the dependency graph.
type ApplicationDependencyEdge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	Relationship string `json:"relationship"`
}

// ApplicationDependencyGraph holds the graph data for Cytoscape.js.
type ApplicationDependencyGraph struct {
	Nodes []ApplicationDependencyNode `json:"nodes"`
	Edges []ApplicationDependencyEdge `json:"edges"`
}

// ApplicationDependency returns the application dependency graph
// (nodes + edges) for use with Cytoscape.js.
// Equivalent to Essential EAM "Application Dependency Viewer".
func ApplicationDependency(db *sql.DB, workspaceID uuid.UUID) (*ApplicationDependencyGraph, error) {
	// Nodes: all application-layer elements.
	nodeRows, err := db.Query(`
		SELECT source_id, name, type
		FROM elements
		WHERE workspace_id = $1 AND layer = 'Application'
		ORDER BY name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("application dependency nodes: %w", err)
	}
	defer func() { _ = nodeRows.Close() }()

	graph := &ApplicationDependencyGraph{}
	for nodeRows.Next() {
		var n ApplicationDependencyNode
		if err := nodeRows.Scan(&n.ID, &n.Name, &n.Type); err != nil {
			return nil, err
		}
		graph.Nodes = append(graph.Nodes, n)
	}
	if err := nodeRows.Err(); err != nil {
		return nil, err
	}

	// Edges: relationships between application-layer elements.
	edgeRows, err := db.Query(`
		SELECT r.source_id, r.source_element, r.target_element, r.type
		FROM relationships r
		JOIN elements src ON src.workspace_id = r.workspace_id
			AND src.source_id = r.source_element AND src.layer = 'Application'
		JOIN elements tgt ON tgt.workspace_id = r.workspace_id
			AND tgt.source_id = r.target_element AND tgt.layer = 'Application'
		WHERE r.workspace_id = $1
		ORDER BY r.source_element, r.target_element`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("application dependency edges: %w", err)
	}
	defer func() { _ = edgeRows.Close() }()

	for edgeRows.Next() {
		var e ApplicationDependencyEdge
		if err := edgeRows.Scan(&e.ID, &e.Source, &e.Target, &e.Relationship); err != nil {
			return nil, err
		}
		graph.Edges = append(graph.Edges, e)
	}
	return graph, edgeRows.Err()
}
