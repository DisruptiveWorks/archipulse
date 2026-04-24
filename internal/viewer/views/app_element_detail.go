package views

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// AppDetailTechRef is a technology element that hosts this application.
type AppDetailTechRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// AppDetailRelRef is an application relationship (interface) to/from this application.
type AppDetailRelRef struct {
	TargetID   string `json:"target_id"`
	TargetName string `json:"target_name"`
	Direction  string `json:"direction"` // "out" | "in"
	RelType    string `json:"rel_type"`  // lowercase short form: "flow", "serving", etc.
}

// AppDetailProcRef is a business process that uses this application.
type AppDetailProcRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// AppElementDetailData is the payload for the application detail panel.
type AppElementDetailData struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	Type       string             `json:"type"`
	Properties map[string]string  `json:"properties"`
	RunsOn     []AppDetailTechRef `json:"runs_on"`
	Interfaces []AppDetailRelRef  `json:"interfaces"`
	Processes  []AppDetailProcRef `json:"processes"`
}

// AppElementDetail returns the rich detail for a single application element
// identified by its source_id within the workspace.
func AppElementDetail(db *sql.DB, workspaceID uuid.UUID, appSourceID string) (*AppElementDetailData, error) {
	// 1. Basic element info.
	var out AppElementDetailData
	out.ID = appSourceID
	err := db.QueryRow(`
		SELECT name, type FROM elements
		WHERE workspace_id = $1 AND source_id = $2`,
		workspaceID, appSourceID,
	).Scan(&out.Name, &out.Type)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("element not found: %s", appSourceID)
	}
	if err != nil {
		return nil, fmt.Errorf("app detail element: %w", err)
	}

	// 2. Properties.
	propRows, err := db.Query(`
		SELECT ep.key, ep.value
		FROM element_properties ep
		JOIN elements e ON e.id = ep.element_id
		WHERE e.workspace_id = $1 AND e.source_id = $2
		ORDER BY ep.key`,
		workspaceID, appSourceID,
	)
	if err != nil {
		return nil, fmt.Errorf("app detail properties: %w", err)
	}
	defer func() { _ = propRows.Close() }()
	out.Properties = map[string]string{}
	for propRows.Next() {
		var k, v string
		if err := propRows.Scan(&k, &v); err != nil {
			return nil, err
		}
		if v != "" {
			out.Properties[k] = v
		}
	}
	if err := propRows.Err(); err != nil {
		return nil, err
	}

	// 3. Runs On — technology elements linked via Assignment.
	techRows, err := db.Query(`
		SELECT DISTINCT e.source_id, e.name, e.type
		FROM relationships r
		JOIN elements e ON e.workspace_id = $1 AND e.layer = 'Technology'
		  AND (e.source_id = r.target_element OR e.source_id = r.source_element)
		WHERE r.workspace_id = $1
		  AND r.type IN ('Assignment', 'AssignmentRelationship')
		  AND (r.source_element = $2 OR r.target_element = $2)
		ORDER BY e.name`,
		workspaceID, appSourceID,
	)
	if err != nil {
		return nil, fmt.Errorf("app detail runs_on: %w", err)
	}
	defer func() { _ = techRows.Close() }()
	out.RunsOn = []AppDetailTechRef{}
	for techRows.Next() {
		var t AppDetailTechRef
		if err := techRows.Scan(&t.ID, &t.Name, &t.Type); err != nil {
			return nil, err
		}
		out.RunsOn = append(out.RunsOn, t)
	}
	if err := techRows.Err(); err != nil {
		return nil, err
	}

	// 4. Interfaces — outgoing application relationships.
	ifaceRows, err := db.Query(fmt.Sprintf(`
		SELECT e.source_id, e.name, r.type, 'out' AS direction
		FROM relationships r
		JOIN elements e ON e.workspace_id = $1
		  AND e.source_id = r.target_element
		  AND e.type IN (%s)
		WHERE r.workspace_id = $1
		  AND r.source_element = $2
		  AND r.type IN (
		    'Serving', 'ServingRelationship',
		    'Flow', 'FlowRelationship',
		    'Access', 'AccessRelationship',
		    'Triggering', 'TriggeringRelationship',
		    'Association', 'AssociationRelationship'
		  )
		UNION ALL
		SELECT e.source_id, e.name, r.type, 'in' AS direction
		FROM relationships r
		JOIN elements e ON e.workspace_id = $1
		  AND e.source_id = r.source_element
		  AND e.type IN (%s)
		WHERE r.workspace_id = $1
		  AND r.target_element = $2
		  AND r.type IN (
		    'Serving', 'ServingRelationship',
		    'Flow', 'FlowRelationship',
		    'Access', 'AccessRelationship',
		    'Triggering', 'TriggeringRelationship',
		    'Association', 'AssociationRelationship'
		  )
		ORDER BY direction, name`, appTypesSQL, appTypesSQL),
		workspaceID, appSourceID,
	)
	if err != nil {
		return nil, fmt.Errorf("app detail interfaces: %w", err)
	}
	defer func() { _ = ifaceRows.Close() }()
	out.Interfaces = []AppDetailRelRef{}
	for ifaceRows.Next() {
		var ref AppDetailRelRef
		var rawType string
		if err := ifaceRows.Scan(&ref.TargetID, &ref.TargetName, &rawType, &ref.Direction); err != nil {
			return nil, err
		}
		ref.RelType = shortRelType(rawType)
		out.Interfaces = append(out.Interfaces, ref)
	}
	if err := ifaceRows.Err(); err != nil {
		return nil, err
	}

	// 5. Used in Processes — business-layer elements that use this app.
	procRows, err := db.Query(`
		SELECT DISTINCT e.source_id, e.name, e.type
		FROM relationships r
		JOIN elements e ON e.workspace_id = $1
		  AND e.source_id = r.source_element
		  AND e.layer = 'Business'
		  AND e.type IN ('BusinessProcess', 'BusinessFunction', 'BusinessInteraction', 'BusinessService')
		WHERE r.workspace_id = $1
		  AND r.target_element = $2
		  AND r.type IN (
		    'Serving', 'ServingRelationship',
		    'Realization', 'RealizationRelationship',
		    'Assignment', 'AssignmentRelationship',
		    'Triggering', 'TriggeringRelationship'
		  )
		ORDER BY e.name`,
		workspaceID, appSourceID,
	)
	if err != nil {
		return nil, fmt.Errorf("app detail processes: %w", err)
	}
	defer func() { _ = procRows.Close() }()
	out.Processes = []AppDetailProcRef{}
	for procRows.Next() {
		var p AppDetailProcRef
		if err := procRows.Scan(&p.ID, &p.Name, &p.Type); err != nil {
			return nil, err
		}
		out.Processes = append(out.Processes, p)
	}
	if err := procRows.Err(); err != nil {
		return nil, err
	}

	return &out, nil
}

func shortRelType(t string) string {
	t = strings.TrimSuffix(t, "Relationship")
	return strings.ToLower(t)
}
