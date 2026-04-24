package views

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// ProcAppProcess is a business process row for the matrix.
type ProcAppProcess struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Capability string `json:"capability"`
}

// ProcAppApp is an application column for the matrix.
type ProcAppApp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ProcAppLink is a cell in the matrix: which process uses which app and how.
// Kind is "R" (Read), "W" (Write), "E" (Execute/Trigger), or "S" (Serving).
type ProcAppLink struct {
	ProcessID string `json:"process_id"`
	AppID     string `json:"app_id"`
	Kind      string `json:"kind"`
}

// ProcessApplicationData is the payload for the Process–Application usage matrix.
type ProcessApplicationData struct {
	Processes []ProcAppProcess `json:"processes"`
	Apps      []ProcAppApp     `json:"apps"`
	Links     []ProcAppLink    `json:"links"`
}

// ProcessApplication returns the matrix data for the Process-Application Usage view.
func ProcessApplication(db *sql.DB, workspaceID uuid.UUID) (*ProcessApplicationData, error) {
	// 1. Business processes with their capability property.
	procRows, err := db.Query(`
		SELECT e.source_id, e.name,
		       COALESCE((
		           SELECT ep.value FROM element_properties ep
		           JOIN elements ec ON ec.id = ep.element_id
		           WHERE ec.workspace_id = $1 AND ec.source_id = e.source_id AND ep.key = 'capability'
		           LIMIT 1
		       ), '') AS capability
		FROM elements e
		WHERE e.workspace_id = $1
		  AND e.layer = 'Business'
		  AND e.type IN ('BusinessProcess','BusinessFunction','BusinessInteraction','BusinessService')
		ORDER BY capability, e.name`,
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("process-application processes: %w", err)
	}
	defer func() { _ = procRows.Close() }()

	var processes []ProcAppProcess
	procIndex := map[string]int{}
	for procRows.Next() {
		var p ProcAppProcess
		if err := procRows.Scan(&p.ID, &p.Name, &p.Capability); err != nil {
			return nil, err
		}
		procIndex[p.ID] = len(processes)
		processes = append(processes, p)
	}
	if err := procRows.Err(); err != nil {
		return nil, err
	}

	if len(processes) == 0 {
		return &ProcessApplicationData{
			Processes: []ProcAppProcess{},
			Apps:      []ProcAppApp{},
			Links:     []ProcAppLink{},
		}, nil
	}

	// 2. Links: relationships between business processes and application elements
	//    in either direction (ArchiMate models vary: some draw App→Serving→Process,
	//    others draw Process→Association→App).
	linkRows, err := db.Query(fmt.Sprintf(`
		SELECT DISTINCT proc_id, app_id, app_name, app_type, kind FROM (
		    -- Direction A: Process (src) → App (tgt)
		    SELECT
		        src.source_id AS proc_id,
		        tgt.source_id AS app_id,
		        tgt.name      AS app_name,
		        tgt.type      AS app_type,
		        COALESCE((
		            SELECT rp.value FROM relationship_properties rp
		            WHERE rp.relationship_id = r.id AND rp.key = 'usage_kind' LIMIT 1
		        ), 'S') AS kind
		    FROM relationships r
		    JOIN elements src ON src.workspace_id = $1 AND src.source_id = r.source_element
		        AND src.layer = 'Business'
		        AND src.type IN ('BusinessProcess','BusinessFunction','BusinessInteraction','BusinessService')
		    JOIN elements tgt ON tgt.workspace_id = $1 AND tgt.source_id = r.target_element
		        AND tgt.layer = 'Application' AND tgt.type IN (%s)
		    WHERE r.workspace_id = $1
		      AND r.type IN (
		          'Serving','ServingRelationship',
		          'Realization','RealizationRelationship',
		          'Assignment','AssignmentRelationship',
		          'Triggering','TriggeringRelationship',
		          'Association','AssociationRelationship'
		      )
		    UNION
		    -- Direction B: App (src) → Process (tgt)  [most common for Serving]
		    SELECT
		        tgt.source_id AS proc_id,
		        src.source_id AS app_id,
		        src.name      AS app_name,
		        src.type      AS app_type,
		        COALESCE((
		            SELECT rp.value FROM relationship_properties rp
		            WHERE rp.relationship_id = r.id AND rp.key = 'usage_kind' LIMIT 1
		        ), 'S') AS kind
		    FROM relationships r
		    JOIN elements src ON src.workspace_id = $1 AND src.source_id = r.source_element
		        AND src.layer = 'Application' AND src.type IN (%s)
		    JOIN elements tgt ON tgt.workspace_id = $1 AND tgt.source_id = r.target_element
		        AND tgt.layer = 'Business'
		        AND tgt.type IN ('BusinessProcess','BusinessFunction','BusinessInteraction','BusinessService')
		    WHERE r.workspace_id = $1
		      AND r.type IN (
		          'Serving','ServingRelationship',
		          'Realization','RealizationRelationship',
		          'Assignment','AssignmentRelationship',
		          'Triggering','TriggeringRelationship',
		          'Association','AssociationRelationship'
		      )
		) combined
		ORDER BY app_name`, appTypesSQL, appTypesSQL),
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("process-application links: %w", err)
	}
	defer func() { _ = linkRows.Close() }()

	var links []ProcAppLink
	appSeen := map[string]bool{}
	appOrder := []string{}
	appMap := map[string]ProcAppApp{}

	for linkRows.Next() {
		var procID, appID, appName, appType, kind string
		if err := linkRows.Scan(&procID, &appID, &appName, &appType, &kind); err != nil {
			return nil, err
		}
		// Only include links for processes we have.
		if _, ok := procIndex[procID]; !ok {
			continue
		}
		links = append(links, ProcAppLink{ProcessID: procID, AppID: appID, Kind: kind})
		if !appSeen[appID] {
			appSeen[appID] = true
			appOrder = append(appOrder, appID)
			appMap[appID] = ProcAppApp{ID: appID, Name: appName, Type: appType}
		}
	}
	if err := linkRows.Err(); err != nil {
		return nil, err
	}

	apps := make([]ProcAppApp, 0, len(appOrder))
	for _, id := range appOrder {
		apps = append(apps, appMap[id])
	}

	if processes == nil {
		processes = []ProcAppProcess{}
	}
	if apps == nil {
		apps = []ProcAppApp{}
	}
	if links == nil {
		links = []ProcAppLink{}
	}

	return &ProcessApplicationData{
		Processes: processes,
		Apps:      apps,
		Links:     links,
	}, nil
}
