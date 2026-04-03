package views

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// AppCatalogueEntry is a single row in the Application Catalogue rich view.
type AppCatalogueEntry struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Type          string            `json:"type"`
	Documentation string            `json:"documentation"`
	Properties    map[string]string `json:"properties"`
}

// AppCatalogueData is the payload for the application catalogue rich view.
type AppCatalogueData struct {
	Entries      []AppCatalogueEntry `json:"entries"`
	PropertyKeys []string            `json:"property_keys"`
}

// AppCatalogueEntries returns all application-layer elements with their
// model properties, ready for the rich catalogue view.
func AppCatalogueEntries(db *sql.DB, workspaceID uuid.UUID) (*AppCatalogueData, error) {
	// Step 1: fetch all application elements.
	rows, err := db.Query(fmt.Sprintf(`
		SELECT source_id, type, name, documentation
		FROM elements
		WHERE workspace_id = $1
		  AND layer = 'Application'
		  AND type IN (%s)
		ORDER BY name`, appTypesSQL), workspaceID)
	if err != nil {
		return nil, fmt.Errorf("app catalogue elements: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entries []AppCatalogueEntry
	var ids []string
	for rows.Next() {
		var e AppCatalogueEntry
		if err := rows.Scan(&e.ID, &e.Type, &e.Name, &e.Documentation); err != nil {
			return nil, err
		}
		e.Properties = map[string]string{}
		entries = append(entries, e)
		ids = append(ids, e.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return &AppCatalogueData{Entries: []AppCatalogueEntry{}, PropertyKeys: []string{}}, nil
	}

	// Step 2: fetch model properties for all entries.
	propsByID, propKeys, err := loadAppProperties(db, workspaceID, ids)
	if err != nil {
		return nil, err
	}

	for i, e := range entries {
		if p, ok := propsByID[e.ID]; ok {
			entries[i].Properties = p
		}
	}

	sort.Strings(propKeys)
	return &AppCatalogueData{Entries: entries, PropertyKeys: propKeys}, nil
}
