// Package diagram manages ArchiMate diagrams (views) within a workspace.
package diagram

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("diagram not found")
var ErrConflict = errors.New("diagram was modified by another request")

type Diagram struct {
	ID            uuid.UUID       `json:"id"`
	WorkspaceID   uuid.UUID       `json:"workspace_id"`
	SourceID      string          `json:"source_id"`
	Name          string          `json:"name"`
	Documentation string          `json:"documentation"`
	Layout        json.RawMessage `json:"layout"`
	Version       int             `json:"version"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) List(workspaceID uuid.UUID) ([]Diagram, error) {
	rows, err := s.db.Query(`
		SELECT id, workspace_id, source_id, name, documentation, layout, version, created_at, updated_at
		FROM diagrams WHERE workspace_id = $1 ORDER BY name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list diagrams: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Diagram
	for rows.Next() {
		var d Diagram
		if err := rows.Scan(&d.ID, &d.WorkspaceID, &d.SourceID, &d.Name,
			&d.Documentation, &d.Layout, &d.Version, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) Get(id uuid.UUID) (*Diagram, error) {
	var d Diagram
	err := s.db.QueryRow(`
		SELECT id, workspace_id, source_id, name, documentation, layout, version, created_at, updated_at
		FROM diagrams WHERE id = $1`, id).
		Scan(&d.ID, &d.WorkspaceID, &d.SourceID, &d.Name,
			&d.Documentation, &d.Layout, &d.Version, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get diagram: %w", err)
	}
	return &d, nil
}

func (s *Store) Create(workspaceID uuid.UUID, sourceID, name, documentation string, layout json.RawMessage) (*Diagram, error) {
	if layout == nil {
		layout = json.RawMessage("{}")
	}
	var d Diagram
	err := s.db.QueryRow(`
		INSERT INTO diagrams (workspace_id, source_id, name, documentation, layout)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, workspace_id, source_id, name, documentation, layout, version, created_at, updated_at`,
		workspaceID, sourceID, name, documentation, layout).
		Scan(&d.ID, &d.WorkspaceID, &d.SourceID, &d.Name,
			&d.Documentation, &d.Layout, &d.Version, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create diagram: %w", err)
	}
	return &d, nil
}

func (s *Store) Update(id uuid.UUID, name, documentation string, layout json.RawMessage, version int) (*Diagram, error) {
	if layout == nil {
		layout = json.RawMessage("{}")
	}
	var d Diagram
	err := s.db.QueryRow(`
		UPDATE diagrams
		SET name = $1, documentation = $2, layout = $3, version = version + 1, updated_at = now()
		WHERE id = $4 AND version = $5
		RETURNING id, workspace_id, source_id, name, documentation, layout, version, created_at, updated_at`,
		name, documentation, layout, id, version).
		Scan(&d.ID, &d.WorkspaceID, &d.SourceID, &d.Name,
			&d.Documentation, &d.Layout, &d.Version, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		if _, err2 := s.Get(id); errors.Is(err2, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrConflict
	}
	if err != nil {
		return nil, fmt.Errorf("update diagram: %w", err)
	}
	return &d, nil
}

func (s *Store) Delete(id uuid.UUID) error {
	res, err := s.db.Exec(`DELETE FROM diagrams WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete diagram: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
