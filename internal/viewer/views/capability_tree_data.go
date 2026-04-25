package views

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// CapabilityNode is a single node in the capability tree.
type CapabilityNode struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	Type           string              `json:"type"`
	ParentID       string              `json:"parent_id"`
	SupportingApps []CapabilitySuppApp `json:"supporting_apps"`
}

// CapabilitySuppApp is an application element that serves a capability or process.
type CapabilitySuppApp struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	LifecycleStatus string `json:"lifecycle_status"`
}

// CapabilityTreeData returns all capability/process nodes with parent references
// and supporting application elements, ready for the frontend to render as a tree.
func CapabilityTreeData(db *sql.DB, workspaceID uuid.UUID) ([]CapabilityNode, error) {
	// 1. Fetch all capability/process elements with their parent ID.
	nodeRows, err := db.Query(`
		SELECT
			e.source_id,
			e.name,
			e.type,
			COALESCE(r.source_element, '') AS parent_id
		FROM elements e
		LEFT JOIN relationships r
			ON  r.workspace_id    = e.workspace_id
			AND r.target_element  = e.source_id
			AND r.type IN ('CompositionRelationship', 'AggregationRelationship', 'Composition', 'Aggregation')
		WHERE e.workspace_id = $1
		  AND e.type = 'Capability'
		ORDER BY parent_id, e.name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("capability tree nodes: %w", err)
	}
	defer func() { _ = nodeRows.Close() }()

	nodes := []CapabilityNode{}
	nodeIndex := map[string]int{}
	for nodeRows.Next() {
		var n CapabilityNode
		if err := nodeRows.Scan(&n.ID, &n.Name, &n.Type, &n.ParentID); err != nil {
			return nil, err
		}
		n.SupportingApps = []CapabilitySuppApp{}
		nodeIndex[n.ID] = len(nodes)
		nodes = append(nodes, n)
	}
	if err := nodeRows.Err(); err != nil {
		return nil, err
	}

	if len(nodes) == 0 {
		return nodes, nil
	}

	// 2. Fetch supporting application elements.
	// Two paths are accepted:
	//   a) Direct:  App → Serving/Association → Capability
	//   b) Via process: App → Serving → BusinessProcess → Realization → Capability
	appRows, err := db.Query(`
		SELECT DISTINCT cap_id, app_id, app_name, app_type, lifecycle_status FROM (
			-- Path A: App → Serving/Association → Capability (direct)
			SELECT
				r.target_element AS cap_id,
				a.source_id      AS app_id,
				a.name           AS app_name,
				a.type           AS app_type,
				COALESCE(
					(SELECT ep.value FROM element_properties ep
					 WHERE ep.element_id = a.id AND ep.key = 'lifecycle_status'
					   AND ep.source = 'model' LIMIT 1), ''
				) AS lifecycle_status
			FROM relationships r
			JOIN elements a
				ON  a.workspace_id = r.workspace_id
				AND a.source_id    = r.source_element
				AND a.layer        = 'Application'
			JOIN elements cap
				ON  cap.workspace_id = r.workspace_id
				AND cap.source_id    = r.target_element
				AND cap.type         = 'Capability'
			WHERE r.workspace_id = $1
			  AND r.type IN ('ServingRelationship','Serving','AssociationRelationship','Association')

			UNION

			-- Path B: App → Serving → BusinessProcess → Realization → Capability
			SELECT
				cap.source_id AS cap_id,
				a.source_id   AS app_id,
				a.name        AS app_name,
				a.type        AS app_type,
				COALESCE(
					(SELECT ep.value FROM element_properties ep
					 WHERE ep.element_id = a.id AND ep.key = 'lifecycle_status'
					   AND ep.source = 'model' LIMIT 1), ''
				) AS lifecycle_status
			FROM relationships r1
			JOIN elements a
				ON  a.workspace_id = r1.workspace_id
				AND a.source_id    = r1.source_element
				AND a.layer        = 'Application'
			JOIN elements bp
				ON  bp.workspace_id = r1.workspace_id
				AND bp.source_id    = r1.target_element
				AND bp.type IN ('BusinessProcess','BusinessFunction','BusinessService','BusinessInteraction')
			JOIN relationships r2
				ON  r2.workspace_id    = $1
				AND r2.source_element  = bp.source_id
				AND r2.type IN ('RealizationRelationship','Realization')
			JOIN elements cap
				ON  cap.workspace_id = r2.workspace_id
				AND cap.source_id    = r2.target_element
				AND cap.type         = 'Capability'
			WHERE r1.workspace_id = $1
			  AND r1.type IN ('ServingRelationship','Serving','AssociationRelationship','Association')
		) combined
		ORDER BY cap_id, app_name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("capability tree apps: %w", err)
	}
	defer func() { _ = appRows.Close() }()

	for appRows.Next() {
		var capID string
		var app CapabilitySuppApp
		if err := appRows.Scan(&capID, &app.ID, &app.Name, &app.Type, &app.LifecycleStatus); err != nil {
			return nil, err
		}
		if idx, ok := nodeIndex[capID]; ok {
			nodes[idx].SupportingApps = append(nodes[idx].SupportingApps, app)
		}
	}
	return nodes, appRows.Err()
}
