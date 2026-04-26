package savedviews

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/pagination"
)

// SavedView represents a persisted automatic view with its filter state.
type SavedView struct {
	ID          uuid.UUID       `json:"id"`
	WorkspaceID uuid.UUID       `json:"workspace_id"`
	ViewType    string          `json:"view_type"`
	Name        string          `json:"name"`
	Filters     json.RawMessage `json:"filters"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) List(wsID uuid.UUID, p pagination.Params) ([]SavedView, int, error) {
	rows, err := s.db.Query(`
		SELECT id, workspace_id, view_type, name, filters, created_at, updated_at,
		       COUNT(*) OVER() AS total
		FROM saved_views
		WHERE workspace_id = $1
		ORDER BY name
		LIMIT $2 OFFSET $3`, wsID, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("list saved views: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []SavedView
	var total int
	for rows.Next() {
		var sv SavedView
		if err := rows.Scan(&sv.ID, &sv.WorkspaceID, &sv.ViewType, &sv.Name, &sv.Filters, &sv.CreatedAt, &sv.UpdatedAt, &total); err != nil {
			return nil, 0, err
		}
		out = append(out, sv)
	}
	return out, total, rows.Err()
}

func (s *Store) Get(wsID, id uuid.UUID) (*SavedView, error) {
	var sv SavedView
	err := s.db.QueryRow(`
		SELECT id, workspace_id, view_type, name, filters, created_at, updated_at
		FROM saved_views
		WHERE workspace_id = $1 AND id = $2`, wsID, id).
		Scan(&sv.ID, &sv.WorkspaceID, &sv.ViewType, &sv.Name, &sv.Filters, &sv.CreatedAt, &sv.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get saved view: %w", err)
	}
	return &sv, nil
}

func (s *Store) Create(wsID uuid.UUID, viewType, name string, filters json.RawMessage) (*SavedView, error) {
	if filters == nil {
		filters = json.RawMessage("{}")
	}
	var sv SavedView
	err := s.db.QueryRow(`
		INSERT INTO saved_views (workspace_id, view_type, name, filters)
		VALUES ($1, $2, $3, $4)
		RETURNING id, workspace_id, view_type, name, filters, created_at, updated_at`,
		wsID, viewType, name, filters).
		Scan(&sv.ID, &sv.WorkspaceID, &sv.ViewType, &sv.Name, &sv.Filters, &sv.CreatedAt, &sv.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create saved view: %w", err)
	}
	return &sv, nil
}

func (s *Store) Update(wsID, id uuid.UUID, name string, filters json.RawMessage) (*SavedView, error) {
	if filters == nil {
		filters = json.RawMessage("{}")
	}
	var sv SavedView
	err := s.db.QueryRow(`
		UPDATE saved_views
		SET name = $3, filters = $4, updated_at = now()
		WHERE workspace_id = $1 AND id = $2
		RETURNING id, workspace_id, view_type, name, filters, created_at, updated_at`,
		wsID, id, name, filters).
		Scan(&sv.ID, &sv.WorkspaceID, &sv.ViewType, &sv.Name, &sv.Filters, &sv.CreatedAt, &sv.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update saved view: %w", err)
	}
	return &sv, nil
}

func (s *Store) Delete(wsID, id uuid.UUID) error {
	res, err := s.db.Exec(`DELETE FROM saved_views WHERE workspace_id = $1 AND id = $2`, wsID, id)
	if err != nil {
		return fmt.Errorf("delete saved view: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}
