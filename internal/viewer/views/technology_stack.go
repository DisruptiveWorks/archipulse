package views

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// TechStackApp is an application row in the Technology Stack matrix.
type TechStackApp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// TechStackTech is a technology column in the matrix.
type TechStackTech struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Vendor    string `json:"vendor"`
	Version   string `json:"version"`
	Lifecycle string `json:"lifecycle"`
}

// TechStackData is the payload for the Technology Stack view.
type TechStackData struct {
	Apps    []TechStackApp  `json:"apps"`
	Tech    []TechStackTech `json:"tech"`
	AppTech [][2]string     `json:"app_tech"`
}

// TechnologyStack returns the matrix data for the Technology Stack view.
func TechnologyStack(db *sql.DB, workspaceID uuid.UUID) (*TechStackData, error) {
	// 1. All technology elements with key properties.
	techRows, err := db.Query(`
		SELECT e.source_id, e.name,
		    COALESCE(cat.value,  'Other')      AS category,
		    COALESCE(ven.value,  '')            AS vendor,
		    COALESCE(ver.value,  '')            AS version,
		    COALESCE(lc.value,   'Production')  AS lifecycle
		FROM elements e
		LEFT JOIN element_properties cat ON cat.element_id = e.id AND cat.key = 'category'
		LEFT JOIN element_properties ven ON ven.element_id = e.id AND ven.key = 'vendor'
		LEFT JOIN element_properties ver ON ver.element_id = e.id AND ver.key = 'version'
		LEFT JOIN element_properties lc  ON lc.element_id  = e.id AND lc.key  = 'lifecycle'
		WHERE e.workspace_id = $1
		  AND e.layer = 'Technology'
		  AND e.type IN (
		      'Node','SystemSoftware','TechnologyService','TechnologyInterface',
		      'TechnologyCollaboration','TechnologyFunction','TechnologyProcess',
		      'Artifact','CommunicationNetwork','Path','Device','TechnologyObject'
		  )
		ORDER BY category, e.name`,
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("technology-stack tech: %w", err)
	}
	defer func() { _ = techRows.Close() }()

	var tech []TechStackTech
	techIndex := map[string]int{}
	for techRows.Next() {
		var t TechStackTech
		if err := techRows.Scan(&t.ID, &t.Name, &t.Category, &t.Vendor, &t.Version, &t.Lifecycle); err != nil {
			return nil, err
		}
		techIndex[t.ID] = len(tech)
		tech = append(tech, t)
	}
	if err := techRows.Err(); err != nil {
		return nil, err
	}

	// 2. Application elements.
	appRows, err := db.Query(fmt.Sprintf(`
		SELECT source_id, name, type
		FROM elements
		WHERE workspace_id = $1
		  AND layer = 'Application'
		  AND type IN (%s)
		ORDER BY name`, appTypesSQL),
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("technology-stack apps: %w", err)
	}
	defer func() { _ = appRows.Close() }()

	var apps []TechStackApp
	for appRows.Next() {
		var a TechStackApp
		if err := appRows.Scan(&a.ID, &a.Name, &a.Type); err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	if err := appRows.Err(); err != nil {
		return nil, err
	}

	// 3. App → Tech relationships (Assignment / Realization).
	relRows, err := db.Query(`
		SELECT DISTINCT
		    CASE
		        WHEN src.layer = 'Application' THEN src.source_id
		        ELSE tgt.source_id
		    END AS app_id,
		    CASE
		        WHEN src.layer = 'Technology' THEN src.source_id
		        ELSE tgt.source_id
		    END AS tech_id
		FROM relationships r
		JOIN elements src ON src.workspace_id = $1 AND src.source_id = r.source_element
		JOIN elements tgt ON tgt.workspace_id = $1 AND tgt.source_id = r.target_element
		WHERE r.workspace_id = $1
		  AND r.type IN ('Assignment','AssignmentRelationship','Realization','RealizationRelationship')
		  AND (
		      (src.layer = 'Application' AND tgt.layer = 'Technology')
		   OR (src.layer = 'Technology'  AND tgt.layer = 'Application')
		  )`,
		workspaceID,
	)
	if err != nil {
		return nil, fmt.Errorf("technology-stack app-tech: %w", err)
	}
	defer func() { _ = relRows.Close() }()

	var appTech [][2]string
	for relRows.Next() {
		var appID, techID string
		if err := relRows.Scan(&appID, &techID); err != nil {
			return nil, err
		}
		appTech = append(appTech, [2]string{appID, techID})
	}
	if err := relRows.Err(); err != nil {
		return nil, err
	}

	if apps == nil {
		apps = []TechStackApp{}
	}
	if tech == nil {
		tech = []TechStackTech{}
	}
	if appTech == nil {
		appTech = [][2]string{}
	}

	return &TechStackData{Apps: apps, Tech: tech, AppTech: appTech}, nil
}
