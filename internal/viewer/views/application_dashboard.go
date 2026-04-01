package views

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// PropertyBucket is one slice of a property donut chart.
type PropertyBucket struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// ApplicationDashboardData is the full payload returned by the dashboard endpoint.
type ApplicationDashboardData struct {
	TotalApps    int                          `json:"total_apps"`
	Capabilities []string                     `json:"capabilities"`
	Properties   map[string][]PropertyBucket  `json:"properties"`
}

// appTypes are the ArchiMate element types treated as "applications" in the dashboard.
var appTypes = []string{
	"'ApplicationComponent'",
	"'ApplicationService'",
	"'ApplicationFunction'",
	"'ApplicationCollaboration'",
	"'ApplicationInterface'",
}

// ApplicationDashboard returns all property distributions for Application-layer
// elements in a workspace, optionally filtered by capability name.
func ApplicationDashboard(db *sql.DB, workspaceID uuid.UUID, capabilityName string) (*ApplicationDashboardData, error) {
	caps, err := listCapabilities(db, workspaceID)
	if err != nil {
		return nil, err
	}

	appIDs, total, err := appsInScope(db, workspaceID, capabilityName)
	if err != nil {
		return nil, err
	}

	props, err := propertyDistributions(db, workspaceID, appIDs, total)
	if err != nil {
		return nil, err
	}

	return &ApplicationDashboardData{
		TotalApps:    total,
		Capabilities: caps,
		Properties:   props,
	}, nil
}

// listCapabilities returns the names of all L1/L2 Capability elements in the workspace.
func listCapabilities(db *sql.DB, workspaceID uuid.UUID) ([]string, error) {
	rows, err := db.Query(`
		SELECT name
		FROM elements
		WHERE workspace_id = $1
		  AND type = 'Capability'
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

// appsInScope returns the UUIDs (pk) and total count of Application elements
// in scope. When capabilityName is empty all apps are returned; otherwise only
// apps that have a RealizationRelationship targeting that capability.
func appsInScope(db *sql.DB, workspaceID uuid.UUID, capabilityName string) ([]uuid.UUID, int, error) {
	typeList := "ApplicationComponent, ApplicationService, ApplicationFunction, ApplicationCollaboration, ApplicationInterface"

	var queryStr string
	var args []any

	if capabilityName == "" || capabilityName == "all" {
		queryStr = fmt.Sprintf(`
			SELECT id FROM elements
			WHERE workspace_id = $1
			  AND type IN (%s)`, sqlInList(appTypes))
		args = []any{workspaceID}
	} else {
		queryStr = fmt.Sprintf(`
			SELECT DISTINCT e.id
			FROM elements e
			JOIN relationships r
				ON r.source_element = e.source_id
				AND r.workspace_id  = e.workspace_id
				AND r.type          = 'RealizationRelationship'
			JOIN elements cap
				ON cap.source_id    = r.target_element
				AND cap.workspace_id = e.workspace_id
				AND cap.type        = 'Capability'
				AND cap.name        = $2
			WHERE e.workspace_id = $1
			  AND e.type IN (%s)`, sqlInList(appTypes))
		args = []any{workspaceID, capabilityName}
	}

	_ = typeList // used in comment above for clarity

	rows, err := db.Query(queryStr, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("apps in scope: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return ids, len(ids), nil
}

// propertyDistributions returns, for every distinct property key found among
// the given app IDs, a slice of (value, count) sorted by count descending.
// Apps missing a given key are counted as "(unset)".
func propertyDistributions(db *sql.DB, workspaceID uuid.UUID, appIDs []uuid.UUID, total int) (map[string][]PropertyBucket, error) {
	if len(appIDs) == 0 {
		return map[string][]PropertyBucket{}, nil
	}

	// Build $2, $3, ... placeholders for the app IDs.
	placeholders := make([]string, len(appIDs))
	args := make([]any, 0, 1+len(appIDs))
	args = append(args, workspaceID)
	for i, id := range appIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}
	inClause := joinStrings(placeholders, ",")

	rows, err := db.Query(fmt.Sprintf(`
		SELECT key, value, COUNT(*) AS cnt
		FROM element_properties
		WHERE element_id IN (%s)
		  AND source = 'model'
		GROUP BY key, value
		ORDER BY key, cnt DESC`, inClause), args...)
	if err != nil {
		return nil, fmt.Errorf("property distributions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	// key → value → count
	raw := map[string]map[string]int{}
	for rows.Next() {
		var key, val string
		var cnt int
		if err := rows.Scan(&key, &val, &cnt); err != nil {
			return nil, err
		}
		if raw[key] == nil {
			raw[key] = map[string]int{}
		}
		raw[key][val] = cnt
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := map[string][]PropertyBucket{}
	for key, valMap := range raw {
		set := 0
		buckets := make([]PropertyBucket, 0, len(valMap)+1)
		for val, cnt := range valMap {
			buckets = append(buckets, PropertyBucket{Value: val, Count: cnt})
			set += cnt
		}
		// sort by count desc, then alpha
		sort.Slice(buckets, func(i, j int) bool {
			if buckets[i].Count != buckets[j].Count {
				return buckets[i].Count > buckets[j].Count
			}
			return buckets[i].Value < buckets[j].Value
		})
		unset := total - set
		if unset > 0 {
			buckets = append(buckets, PropertyBucket{Value: "(unset)", Count: unset})
		}
		out[key] = buckets
	}
	return out, nil
}

func sqlInList(items []string) string {
	return joinStrings(items, ", ")
}

func joinStrings(s []string, sep string) string {
	result := ""
	for i, v := range s {
		if i > 0 {
			result += sep
		}
		result += v
	}
	return result
}
