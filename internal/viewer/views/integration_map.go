package views

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// IntegrationNode is a node in the integration map graph.
type IntegrationNode struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Layer string `json:"layer"`
}

// IntegrationEdge is a directed edge in the integration map.
type IntegrationEdge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	Relationship string `json:"relationship"`
}

// IntegrationGraph holds graph data for the integration map view.
type IntegrationGraph struct {
	Nodes []IntegrationNode `json:"nodes"`
	Edges []IntegrationEdge `json:"edges"`
}

// IntegrationMap returns the application integration topology:
// components and services as nodes, integration relationships as edges,
// plus data objects that are accessed/exchanged between systems.
// Excludes structural relationships (Composition, Realization, Specialization).
func IntegrationMap(db *sql.DB, workspaceID uuid.UUID) (*IntegrationGraph, error) {
	graph := &IntegrationGraph{
		Nodes: []IntegrationNode{},
		Edges: []IntegrationEdge{},
	}

	// Nodes: application components, services, interfaces, and data objects
	// that participate in at least one integration relationship.
	nodeRows, err := db.Query(`
		SELECT DISTINCT e.source_id, e.name, e.type, e.layer
		FROM elements e
		JOIN relationships r
			ON  r.workspace_id = e.workspace_id
			AND (r.source_element = e.source_id OR r.target_element = e.source_id)
			AND r.type IN (
				'ServingRelationship',  'Serving',
				'AssociationRelationship', 'Association',
				'FlowRelationship',     'Flow',
				'AccessRelationship',   'Access',
				'TriggeringRelationship', 'Triggering',
				'AssignmentRelationship', 'Assignment'
			)
		WHERE e.workspace_id = $1
		  AND e.type IN (
			'ApplicationComponent',
			'ApplicationService',
			'ApplicationInterface',
			'ApplicationCollaboration',
			'DataObject'
		  )
		ORDER BY e.name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("integration map nodes: %w", err)
	}
	defer func() { _ = nodeRows.Close() }()

	seen := map[string]bool{}
	for nodeRows.Next() {
		var n IntegrationNode
		if err := nodeRows.Scan(&n.ID, &n.Name, &n.Type, &n.Layer); err != nil {
			return nil, err
		}
		if !seen[n.ID] {
			seen[n.ID] = true
			graph.Nodes = append(graph.Nodes, n)
		}
	}
	if err := nodeRows.Err(); err != nil {
		return nil, err
	}

	if len(graph.Nodes) == 0 {
		return graph, nil
	}

	// Edges: integration relationships between the node set above.
	edgeRows, err := db.Query(`
		SELECT r.source_id, r.source_element, r.target_element, r.type
		FROM relationships r
		JOIN elements src ON src.workspace_id = r.workspace_id
			AND src.source_id = r.source_element
			AND src.type IN (
				'ApplicationComponent', 'ApplicationService',
				'ApplicationInterface', 'ApplicationCollaboration', 'DataObject'
			)
		JOIN elements tgt ON tgt.workspace_id = r.workspace_id
			AND tgt.source_id = r.target_element
			AND tgt.type IN (
				'ApplicationComponent', 'ApplicationService',
				'ApplicationInterface', 'ApplicationCollaboration', 'DataObject'
			)
		WHERE r.workspace_id = $1
		  AND r.type IN (
			'ServingRelationship',     'Serving',
			'AssociationRelationship', 'Association',
			'FlowRelationship',        'Flow',
			'AccessRelationship',      'Access',
			'TriggeringRelationship',  'Triggering',
			'AssignmentRelationship',  'Assignment'
		  )
		ORDER BY r.source_element, r.target_element`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("integration map edges: %w", err)
	}
	defer func() { _ = edgeRows.Close() }()

	for edgeRows.Next() {
		var e IntegrationEdge
		if err := edgeRows.Scan(&e.ID, &e.Source, &e.Target, &e.Relationship); err != nil {
			return nil, err
		}
		graph.Edges = append(graph.Edges, e)
	}
	return graph, edgeRows.Err()
}
