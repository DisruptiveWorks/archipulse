package views

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// TechnologyCatalogue lists all technology-layer elements grouped by type.
// Equivalent to Essential EAM "Technology Product Catalogue as Table".
func TechnologyCatalogue(db *sql.DB, workspaceID uuid.UUID) (string, []string, [][]any, error) {
	rows, err := db.Query(`
		SELECT source_id, type, name, documentation
		FROM elements
		WHERE workspace_id = $1
		  AND layer = 'Technology'
		ORDER BY type, name`, workspaceID)
	if err != nil {
		return "", nil, nil, fmt.Errorf("technology catalogue: %w", err)
	}
	defer func() { _ = rows.Close() }()

	columns := []string{"ID", "Type", "Name", "Documentation"}
	var data [][]any
	for rows.Next() {
		var id, typ, name, doc string
		if err := rows.Scan(&id, &typ, &name, &doc); err != nil {
			return "", nil, nil, err
		}
		data = append(data, []any{id, typ, name, doc})
	}
	return "Technology Catalogue", columns, data, rows.Err()
}
