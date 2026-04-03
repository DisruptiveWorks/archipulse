package views

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// TechCatalogueEntry is a single row in the Technology Catalogue rich view.
type TechCatalogueEntry struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Documentation string   `json:"documentation"`
	UsedByApps    []string `json:"used_by_apps"`
}

// TechCatalogueData is the payload for the technology catalogue rich view.
type TechCatalogueData struct {
	Entries []TechCatalogueEntry `json:"entries"`
}

// TechCatalogueEntries returns all technology-layer elements with the list of
// application elements they host/run (via Assignment relationships).
func TechCatalogueEntries(db *sql.DB, workspaceID uuid.UUID) (*TechCatalogueData, error) {
	// Step 1: fetch all technology elements.
	rows, err := db.Query(`
		SELECT source_id, type, name, documentation
		FROM elements
		WHERE workspace_id = $1
		  AND layer = 'Technology'
		ORDER BY type, name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("tech catalogue elements: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entries []TechCatalogueEntry
	var ids []string
	idxByID := map[string]int{}
	for rows.Next() {
		var e TechCatalogueEntry
		if err := rows.Scan(&e.ID, &e.Type, &e.Name, &e.Documentation); err != nil {
			return nil, err
		}
		e.UsedByApps = []string{}
		idxByID[e.ID] = len(entries)
		entries = append(entries, e)
		ids = append(ids, e.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return &TechCatalogueData{Entries: []TechCatalogueEntry{}}, nil
	}

	// Step 2: find which application elements are assigned to each tech element.
	// Assignment: tech element (source) → app element (target), or reverse.
	asgRows, err := db.Query(fmt.Sprintf(`
		SELECT
			CASE
				WHEN e_tech.source_id = ANY($2) THEN e_tech.source_id
				ELSE e_app.source_id
			END AS tech_id,
			CASE
				WHEN e_tech.source_id = ANY($2) THEN e_app.name
				ELSE e_tech.name
			END AS app_name
		FROM relationships r
		JOIN elements e_tech
			ON e_tech.workspace_id = $1
			AND e_tech.source_id   = r.source_element
			AND e_tech.layer       = 'Technology'
		JOIN elements e_app
			ON e_app.workspace_id = $1
			AND e_app.source_id   = r.target_element
			AND e_app.type IN (%s)
		WHERE r.workspace_id = $1
		  AND r.type IN ('Assignment', 'AssignmentRelationship')
		  AND e_tech.source_id = ANY($2)
		ORDER BY tech_id, app_name`, appTypesSQL),
		workspaceID, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("tech catalogue assignments: %w", err)
	}
	defer func() { _ = asgRows.Close() }()

	for asgRows.Next() {
		var techID, appName string
		if err := asgRows.Scan(&techID, &appName); err != nil {
			return nil, err
		}
		if idx, ok := idxByID[techID]; ok {
			entries[idx].UsedByApps = append(entries[idx].UsedByApps, appName)
		}
	}
	if err := asgRows.Err(); err != nil {
		return nil, err
	}

	return &TechCatalogueData{Entries: entries}, nil
}
