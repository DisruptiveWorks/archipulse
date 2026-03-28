// Package element manages ArchiMate elements within a workspace.
package element

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("element not found")
var ErrConflict = errors.New("element was modified by another request")

type Element struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	SourceID      string
	Type          string
	Name          string
	Documentation string
	Version       int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) List(workspaceID uuid.UUID) ([]Element, error) {
	rows, err := s.db.Query(`
		SELECT id, workspace_id, source_id, type, name, documentation, version, created_at, updated_at
		FROM elements WHERE workspace_id = $1 ORDER BY name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list elements: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Element
	for rows.Next() {
		var e Element
		if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.SourceID, &e.Type, &e.Name,
			&e.Documentation, &e.Version, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) Get(id uuid.UUID) (*Element, error) {
	var e Element
	err := s.db.QueryRow(`
		SELECT id, workspace_id, source_id, type, name, documentation, version, created_at, updated_at
		FROM elements WHERE id = $1`, id).
		Scan(&e.ID, &e.WorkspaceID, &e.SourceID, &e.Type, &e.Name,
			&e.Documentation, &e.Version, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get element: %w", err)
	}
	return &e, nil
}

func (s *Store) Create(workspaceID uuid.UUID, sourceID, typ, name, documentation string) (*Element, error) {
	var e Element
	err := s.db.QueryRow(`
		INSERT INTO elements (workspace_id, source_id, type, name, documentation)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, workspace_id, source_id, type, name, documentation, version, created_at, updated_at`,
		workspaceID, sourceID, typ, name, documentation).
		Scan(&e.ID, &e.WorkspaceID, &e.SourceID, &e.Type, &e.Name,
			&e.Documentation, &e.Version, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create element: %w", err)
	}
	return &e, nil
}

func (s *Store) Update(id uuid.UUID, typ, name, documentation string, version int) (*Element, error) {
	var e Element
	err := s.db.QueryRow(`
		UPDATE elements
		SET type = $1, name = $2, documentation = $3, version = version + 1, updated_at = now()
		WHERE id = $4 AND version = $5
		RETURNING id, workspace_id, source_id, type, name, documentation, version, created_at, updated_at`,
		typ, name, documentation, id, version).
		Scan(&e.ID, &e.WorkspaceID, &e.SourceID, &e.Type, &e.Name,
			&e.Documentation, &e.Version, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		if _, err2 := s.Get(id); errors.Is(err2, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrConflict
	}
	if err != nil {
		return nil, fmt.Errorf("update element: %w", err)
	}
	return &e, nil
}

func (s *Store) Delete(id uuid.UUID) error {
	res, err := s.db.Exec(`DELETE FROM elements WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete element: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
