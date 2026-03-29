package views

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// ElementCatalogue returns all elements in a workspace ordered by layer and name.
// Equivalent to Essential EAM "Application Capability Catalogue by Name" generalised to all types.
func ElementCatalogue(db *sql.DB, workspaceID uuid.UUID) (string, []string, [][]any, error) {
	rows, err := db.Query(`
		SELECT layer, type, name, documentation
		FROM elements
		WHERE workspace_id = $1
		ORDER BY layer, type, name`, workspaceID)
	if err != nil {
		return "", nil, nil, fmt.Errorf("element catalogue: %w", err)
	}
	defer func() { _ = rows.Close() }()

	columns := []string{"Layer", "Type", "Name", "Documentation"}
	var data [][]any
	for rows.Next() {
		var layer, typ, name, doc string
		if err := rows.Scan(&layer, &typ, &name, &doc); err != nil {
			return "", nil, nil, err
		}
		data = append(data, []any{layer, typ, name, doc})
	}
	return "Element Catalogue", columns, data, rows.Err()
}
