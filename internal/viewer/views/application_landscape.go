package views

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// ApplicationLandscape maps ApplicationComponents to the BusinessProcesses
// and Capabilities they serve or realize.
// Equivalent to Essential EAM "Application Landscape".
func ApplicationLandscape(db *sql.DB, workspaceID uuid.UUID) (string, []string, [][]any, error) {
	rows, err := db.Query(`
		SELECT
			app.name           AS application,
			app.source_id      AS application_id,
			COALESCE(bus.name, '') AS business_element,
			COALESCE(bus.type, '') AS business_type,
			COALESCE(r.type, '')   AS relationship_type
		FROM elements app
		LEFT JOIN relationships r
			ON r.workspace_id = app.workspace_id
			AND r.source_element = app.source_id
			AND r.type IN ('ServingRelationship', 'RealizationRelationship', 'AssignmentRelationship')
		LEFT JOIN elements bus
			ON bus.workspace_id = app.workspace_id
			AND bus.source_id = r.target_element
			AND bus.layer IN ('Business', 'Strategy')
		WHERE app.workspace_id = $1
		  AND app.type IN ('ApplicationComponent', 'ApplicationService', 'ApplicationFunction')
		ORDER BY app.name, bus.name`, workspaceID)
	if err != nil {
		return "", nil, nil, fmt.Errorf("application landscape: %w", err)
	}
	defer func() { _ = rows.Close() }()

	columns := []string{"Application", "Application ID", "Business Element", "Business Type", "Relationship"}
	var data [][]any
	for rows.Next() {
		var app, appID, bus, busType, relType string
		if err := rows.Scan(&app, &appID, &bus, &busType, &relType); err != nil {
			return "", nil, nil, err
		}
		data = append(data, []any{app, appID, bus, busType, relType})
	}
	return "Application Landscape", columns, data, rows.Err()
}
