package views

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// ApplicationCatalogue lists all application-layer elements.
// Equivalent to Essential EAM "Application Catalogue as Table".
func ApplicationCatalogue(db *sql.DB, workspaceID uuid.UUID) (string, []string, [][]any, error) {
	rows, err := db.Query(`
		SELECT source_id, type, name, documentation
		FROM elements
		WHERE workspace_id = $1
		  AND layer = 'Application'
		ORDER BY type, name`, workspaceID)
	if err != nil {
		return "", nil, nil, fmt.Errorf("application catalogue: %w", err)
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
	return "Application Catalogue", columns, data, rows.Err()
}
