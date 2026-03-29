package views

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// CapabilityTree returns all Capability and BusinessCapability elements
// with their parent-child composition relationships.
// Equivalent to Essential EAM "Business Capability Tree".
func CapabilityTree(db *sql.DB, workspaceID uuid.UUID) (string, []string, [][]any, error) {
	rows, err := db.Query(`
		SELECT
			e.source_id,
			e.name,
			e.type,
			COALESCE(parent.name, '') AS parent_name,
			COALESCE(parent.source_id, '') AS parent_id
		FROM elements e
		LEFT JOIN relationships r
			ON r.workspace_id = e.workspace_id
			AND r.target_element = e.source_id
			AND r.type IN ('CompositionRelationship', 'AggregationRelationship')
		LEFT JOIN elements parent
			ON parent.workspace_id = e.workspace_id
			AND parent.source_id = r.source_element
			AND parent.type IN ('Capability', 'BusinessFunction')
		WHERE e.workspace_id = $1
		  AND e.type IN ('Capability', 'BusinessFunction', 'BusinessProcess')
		ORDER BY parent_name, e.name`, workspaceID)
	if err != nil {
		return "", nil, nil, fmt.Errorf("capability tree: %w", err)
	}
	defer func() { _ = rows.Close() }()

	columns := []string{"ID", "Name", "Type", "Parent Name", "Parent ID"}
	var data [][]any
	for rows.Next() {
		var id, name, typ, parentName, parentID string
		if err := rows.Scan(&id, &name, &typ, &parentName, &parentID); err != nil {
			return "", nil, nil, err
		}
		data = append(data, []any{id, name, typ, parentName, parentID})
	}
	return "Business Capability Tree", columns, data, rows.Err()
}
