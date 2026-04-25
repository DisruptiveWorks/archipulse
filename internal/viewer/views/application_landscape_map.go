package views

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// LandscapeApp is a single application entry in the landscape map.
type LandscapeApp struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties"`
}

// LandscapeL2 is an L2 capability with its realizing apps.
type LandscapeL2 struct {
	ID   string         `json:"id"`
	Name string         `json:"name"`
	Apps []LandscapeApp `json:"apps"`
}

// LandscapeL1 is a top-level capability group.
type LandscapeL1 struct {
	ID   string        `json:"id"`
	Name string        `json:"name"`
	L2   []LandscapeL2 `json:"l2"`
}

// ApplicationLandscapeMapData is the payload for the landscape map view.
type ApplicationLandscapeMapData struct {
	L1         []LandscapeL1 `json:"l1"`
	Properties []string      `json:"properties"` // distinct property keys available for the overlay
}

// ApplicationLandscapeMap builds the L1 → L2 → apps hierarchy for the
// Application Landscape Map view.
func ApplicationLandscapeMap(db *sql.DB, workspaceID uuid.UUID) (*ApplicationLandscapeMapData, error) {
	// Step 1: build the capability hierarchy (L1 → L2 via Composition).
	l1Map, l1Order, err := buildCapabilityHierarchy(db, workspaceID)
	if err != nil {
		return nil, err
	}

	// Step 2: load apps per L2 capability.
	appsByCapID, err := loadAppsByCapability(db, workspaceID)
	if err != nil {
		return nil, err
	}

	// Step 3: collect all app source_ids to load properties.
	allAppIDs := collectAppIDs(appsByCapID)
	propsByApp, propKeys, err := loadAppProperties(db, workspaceID, allAppIDs)
	if err != nil {
		return nil, err
	}

	// Step 4: attach properties to apps.
	for capID, apps := range appsByCapID {
		for i, a := range apps {
			if p, ok := propsByApp[a.ID]; ok {
				apps[i].Properties = p
			}
			_ = capID
		}
		appsByCapID[capID] = apps
	}

	// Step 5: assemble the final structure.
	l1List := make([]LandscapeL1, 0, len(l1Order))
	for _, l1ID := range l1Order {
		l1 := l1Map[l1ID]
		for i, l2 := range l1.L2 {
			apps := appsByCapID[l2.ID]
			if apps == nil {
				apps = []LandscapeApp{}
			}
			l1.L2[i].Apps = apps
		}
		l1List = append(l1List, l1)
	}

	sort.Strings(propKeys)
	return &ApplicationLandscapeMapData{L1: l1List, Properties: propKeys}, nil
}

// buildCapabilityHierarchy returns a map of L1 capability ID → LandscapeL1 (with L2 list)
// and an ordered list of L1 IDs (alphabetical).
func buildCapabilityHierarchy(db *sql.DB, workspaceID uuid.UUID) (map[string]LandscapeL1, []string, error) {
	// Query all (L1 source_id, L1 name, L2 source_id, L2 name) pairs via Composition.
	rows, err := db.Query(`
		SELECT
			p.source_id AS l1_id,
			p.name      AS l1_name,
			c.source_id AS l2_id,
			c.name      AS l2_name
		FROM relationships r
		JOIN elements p
			ON  p.workspace_id = $1
			AND p.source_id    = r.source_element
			AND p.type         = 'Capability'
		JOIN elements c
			ON  c.workspace_id = $1
			AND c.source_id    = r.target_element
			AND c.type         = 'Capability'
		WHERE r.workspace_id = $1
		  AND r.type IN ('Composition', 'CompositionRelationship')
		ORDER BY p.name, c.name`, workspaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("capability hierarchy: %w", err)
	}
	defer func() { _ = rows.Close() }()

	l1Map := map[string]LandscapeL1{}
	l1Order := []string{}

	for rows.Next() {
		var l1ID, l1Name, l2ID, l2Name string
		if err := rows.Scan(&l1ID, &l1Name, &l2ID, &l2Name); err != nil {
			return nil, nil, err
		}
		l1, exists := l1Map[l1ID]
		if !exists {
			l1 = LandscapeL1{ID: l1ID, Name: l1Name, L2: []LandscapeL2{}}
			l1Order = append(l1Order, l1ID)
		}
		l1.L2 = append(l1.L2, LandscapeL2{ID: l2ID, Name: l2Name, Apps: []LandscapeApp{}})
		l1Map[l1ID] = l1
	}
	return l1Map, l1Order, rows.Err()
}

// loadAppsByCapability returns a map of capability source_id → []LandscapeApp.
// Supports two paths:
//
//	A) App → Realization → Capability (direct)
//	B) App → Serving → BusinessProcess → Realization → Capability (2-hop)
func loadAppsByCapability(db *sql.DB, workspaceID uuid.UUID) (map[string][]LandscapeApp, error) {
	rows, err := db.Query(fmt.Sprintf(`
		SELECT DISTINCT cap_id, app_id, app_name, app_type FROM (
			-- Path A: App → Realization → Capability (direct)
			SELECT
				cap.source_id AS cap_id,
				e.source_id   AS app_id,
				e.name        AS app_name,
				e.type        AS app_type
			FROM relationships r
			JOIN elements e
				ON  e.workspace_id = $1
				AND e.source_id    = r.source_element
				AND e.type IN (%s)
			JOIN elements cap
				ON  cap.workspace_id = $1
				AND cap.source_id    = r.target_element
				AND cap.type         = 'Capability'
			WHERE r.workspace_id = $1
			  AND r.type IN ('Realization', 'RealizationRelationship')

			UNION

			-- Path B: App → Serving → BusinessProcess → Realization → Capability
			SELECT
				cap.source_id AS cap_id,
				a.source_id   AS app_id,
				a.name        AS app_name,
				a.type        AS app_type
			FROM relationships r1
			JOIN elements a
				ON  a.workspace_id = r1.workspace_id
				AND a.source_id    = r1.source_element
				AND a.type IN (%s)
			JOIN elements bp
				ON  bp.workspace_id = r1.workspace_id
				AND bp.source_id    = r1.target_element
				AND bp.type IN ('BusinessProcess','BusinessFunction','BusinessService','BusinessInteraction')
			JOIN relationships r2
				ON  r2.workspace_id   = $1
				AND r2.source_element = bp.source_id
				AND r2.type IN ('Realization', 'RealizationRelationship')
			JOIN elements cap
				ON  cap.workspace_id = r2.workspace_id
				AND cap.source_id    = r2.target_element
				AND cap.type         = 'Capability'
			WHERE r1.workspace_id = $1
			  AND r1.type IN ('ServingRelationship','Serving','AssociationRelationship','Association')
		) combined
		ORDER BY cap_id, app_name`, appTypesSQL, appTypesSQL), workspaceID)
	if err != nil {
		return nil, fmt.Errorf("apps by capability: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := map[string][]LandscapeApp{}
	for rows.Next() {
		var capID, appID, appName, appType string
		if err := rows.Scan(&capID, &appID, &appName, &appType); err != nil {
			return nil, err
		}
		out[capID] = append(out[capID], LandscapeApp{
			ID:         appID,
			Name:       appName,
			Type:       appType,
			Properties: map[string]string{},
		})
	}
	return out, rows.Err()
}

func collectAppIDs(appsByCapID map[string][]LandscapeApp) []string {
	seen := map[string]struct{}{}
	var ids []string
	for _, apps := range appsByCapID {
		for _, a := range apps {
			if _, ok := seen[a.ID]; !ok {
				seen[a.ID] = struct{}{}
				ids = append(ids, a.ID)
			}
		}
	}
	return ids
}

// loadAppProperties returns a map of app source_id → {key: value} and the
// list of distinct property keys present across all apps.
func loadAppProperties(db *sql.DB, workspaceID uuid.UUID, appSourceIDs []string) (map[string]map[string]string, []string, error) {
	if len(appSourceIDs) == 0 {
		return map[string]map[string]string{}, []string{}, nil
	}

	rows, err := db.Query(`
		SELECT e.source_id, ep.key, ep.value
		FROM element_properties ep
		JOIN elements e ON e.id = ep.element_id
		WHERE e.workspace_id = $1
		  AND ep.source = 'model'
		  AND e.source_id = ANY($2)`,
		workspaceID, pq.Array(appSourceIDs))
	if err != nil {
		return nil, nil, fmt.Errorf("landscape app properties: %w", err)
	}
	defer func() { _ = rows.Close() }()

	props := map[string]map[string]string{}
	keySet := map[string]struct{}{}
	for rows.Next() {
		var sourceID, key, value string
		if err := rows.Scan(&sourceID, &key, &value); err != nil {
			return nil, nil, err
		}
		if props[sourceID] == nil {
			props[sourceID] = map[string]string{}
		}
		props[sourceID][key] = value
		keySet[key] = struct{}{}
	}

	keys := make([]string, 0, len(keySet))
	for k := range keySet {
		keys = append(keys, k)
	}
	return props, keys, rows.Err()
}
