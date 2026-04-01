package views

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// LifecycleBucket is one slice of the lifecycle status donut chart.
type LifecycleBucket struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// TypeBucket is one bar in the application type bar chart.
type TypeBucket struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// ApplicationDashboardData is the full payload returned by the dashboard endpoint.
type ApplicationDashboardData struct {
	TotalApps int               `json:"total_apps"`
	Lifecycle []LifecycleBucket `json:"lifecycle"`
	ByType    []TypeBucket      `json:"by_type"`
}

// ApplicationDashboard returns lifecycle-status and type distribution for the
// Application layer of a workspace.
func ApplicationDashboard(db *sql.DB, workspaceID uuid.UUID) (*ApplicationDashboardData, error) {
	lifecycle, err := lifecycleDistribution(db, workspaceID)
	if err != nil {
		return nil, err
	}

	byType, err := typeDistribution(db, workspaceID)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, b := range byType {
		total += b.Count
	}

	return &ApplicationDashboardData{
		TotalApps: total,
		Lifecycle: lifecycle,
		ByType:    byType,
	}, nil
}

func lifecycleDistribution(db *sql.DB, workspaceID uuid.UUID) ([]LifecycleBucket, error) {
	rows, err := db.Query(`
		SELECT
			COALESCE(ep.value, '(unset)') AS status,
			COUNT(DISTINCT e.id)           AS cnt
		FROM elements e
		LEFT JOIN element_properties ep
			ON ep.element_id = e.id
			AND ep.key       = 'lifecycle_status'
			AND ep.source    = 'model'
		WHERE e.workspace_id = $1
		  AND e.type IN (
			'ApplicationComponent',
			'ApplicationService',
			'ApplicationFunction',
			'ApplicationCollaboration',
			'ApplicationInterface'
		  )
		GROUP BY COALESCE(ep.value, '(unset)')
		ORDER BY cnt DESC`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("lifecycle distribution: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []LifecycleBucket
	for rows.Next() {
		var b LifecycleBucket
		if err := rows.Scan(&b.Status, &b.Count); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if out == nil {
		out = []LifecycleBucket{}
	}
	return out, rows.Err()
}

func typeDistribution(db *sql.DB, workspaceID uuid.UUID) ([]TypeBucket, error) {
	rows, err := db.Query(`
		SELECT type, COUNT(*) AS cnt
		FROM elements
		WHERE workspace_id = $1
		  AND layer = 'Application'
		GROUP BY type
		ORDER BY cnt DESC`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("type distribution: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []TypeBucket
	for rows.Next() {
		var b TypeBucket
		if err := rows.Scan(&b.Type, &b.Count); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if out == nil {
		out = []TypeBucket{}
	}
	return out, rows.Err()
}
