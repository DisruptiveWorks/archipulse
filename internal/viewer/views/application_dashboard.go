package views

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// AppEntry is one application in scope with its resolved model properties.
type AppEntry struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties"`
}

// PropertyBucket is one slice of a property donut chart.
type PropertyBucket struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// ApplicationDashboardData is the full payload returned by the dashboard endpoint.
type ApplicationDashboardData struct {
	TotalApps    int                         `json:"total_apps"`
	Capabilities []string                    `json:"capabilities"`
	Apps         []AppEntry                  `json:"apps"`
	Properties   map[string][]PropertyBucket `json:"properties"`
}

// appTypesSQL is the SQL IN-list for ArchiMate application element types.
const appTypesSQL = `'ApplicationComponent','ApplicationService','ApplicationFunction','ApplicationCollaboration','ApplicationInterface'`

// ApplicationDashboard returns property distributions + the app list for Application-layer
// elements, optionally scoped to apps realizing a specific capability.
func ApplicationDashboard(db *sql.DB, workspaceID uuid.UUID, capabilityName string) (*ApplicationDashboardData, error) {
	caps, err := listCapabilities(db, workspaceID)
	if err != nil {
		return nil, err
	}

	apps, err := appsInScope(db, workspaceID, capabilityName)
	if err != nil {
		return nil, err
	}

	props := propertyDistributions(apps)

	return &ApplicationDashboardData{
		TotalApps:    len(apps),
		Capabilities: caps,
		Apps:         apps,
		Properties:   props,
	}, nil
}

// listCapabilities returns names of all Capability elements in the workspace.
func listCapabilities(db *sql.DB, workspaceID uuid.UUID) ([]string, error) {
	rows, err := db.Query(`
		SELECT name FROM elements
		WHERE workspace_id = $1 AND type = 'Capability'
		ORDER BY name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list capabilities: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// appsInScope returns AppEntry records with model properties for all apps in scope.
// When capabilityName is "" or "all", all application-layer elements are returned.
// Otherwise only apps linked to that capability via a Realization relationship
// (supports both directions and both "Realization" / "RealizationRelationship" type names).
func appsInScope(db *sql.DB, workspaceID uuid.UUID, capabilityName string) ([]AppEntry, error) {
	var queryStr string
	var args []any

	if capabilityName == "" || capabilityName == "all" {
		queryStr = fmt.Sprintf(`
			SELECT source_id, name, type
			FROM elements
			WHERE workspace_id = $1 AND type IN (%s)
			ORDER BY name`, appTypesSQL)
		args = []any{workspaceID}
	} else {
		// AOEF stores xsi:type="Realization" (no Relationship suffix).
		// Support both spellings and both source/target directions.
		queryStr = fmt.Sprintf(`
			SELECT DISTINCT e.source_id, e.name, e.type
			FROM elements e
			JOIN relationships r
				ON  r.workspace_id = $1
				AND r.type IN ('Realization', 'RealizationRelationship')
				AND (r.source_element = e.source_id OR r.target_element = e.source_id)
			JOIN elements cap
				ON  cap.workspace_id = $1
				AND cap.type = 'Capability'
				AND cap.name = $2
				AND (cap.source_id = r.target_element OR cap.source_id = r.source_element)
			WHERE e.workspace_id = $1
			  AND e.type IN (%s)
			ORDER BY e.name`, appTypesSQL)
		args = []any{workspaceID, capabilityName}
	}

	rows, err := db.Query(queryStr, args...)
	if err != nil {
		return nil, fmt.Errorf("apps in scope: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var apps []AppEntry
	var sourceIDs []string
	for rows.Next() {
		var a AppEntry
		if err := rows.Scan(&a.ID, &a.Name, &a.Type); err != nil {
			return nil, err
		}
		a.Properties = map[string]string{}
		apps = append(apps, a)
		sourceIDs = append(sourceIDs, a.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(apps) == 0 {
		return []AppEntry{}, nil
	}

	// Fetch model properties for all apps in one query using ANY($2::text[]).
	propRows, err := db.Query(`
		SELECT e.source_id, ep.key, ep.value
		FROM element_properties ep
		JOIN elements e ON e.id = ep.element_id
		WHERE e.workspace_id = $1
		  AND ep.source = 'model'
		  AND e.source_id = ANY($2)`,
		workspaceID, pq.Array(sourceIDs))
	if err != nil {
		return nil, fmt.Errorf("app properties: %w", err)
	}
	defer func() { _ = propRows.Close() }()

	idxMap := make(map[string]int, len(apps))
	for i, a := range apps {
		idxMap[a.ID] = i
	}
	for propRows.Next() {
		var sourceID, key, value string
		if err := propRows.Scan(&sourceID, &key, &value); err != nil {
			return nil, err
		}
		if i, ok := idxMap[sourceID]; ok {
			apps[i].Properties[key] = value
		}
	}
	return apps, propRows.Err()
}

// propertyDistributions derives bucket counts from the in-memory app list.
// No additional DB query needed since apps already carry their properties.
func propertyDistributions(apps []AppEntry) map[string][]PropertyBucket {
	if len(apps) == 0 {
		return map[string][]PropertyBucket{}
	}

	keySet := map[string]struct{}{}
	for _, a := range apps {
		for k := range a.Properties {
			keySet[k] = struct{}{}
		}
	}

	out := map[string][]PropertyBucket{}
	for key := range keySet {
		counts := map[string]int{}
		for _, a := range apps {
			v := a.Properties[key]
			if v == "" {
				counts["(unset)"]++
			} else {
				counts[v]++
			}
		}
		buckets := make([]PropertyBucket, 0, len(counts))
		for val, cnt := range counts {
			buckets = append(buckets, PropertyBucket{Value: val, Count: cnt})
		}
		sort.Slice(buckets, func(i, j int) bool {
			if buckets[i].Value == "(unset)" {
				return false
			}
			if buckets[j].Value == "(unset)" {
				return true
			}
			if buckets[i].Count != buckets[j].Count {
				return buckets[i].Count > buckets[j].Count
			}
			return buckets[i].Value < buckets[j].Value
		})
		out[key] = buckets
	}
	return out
}
