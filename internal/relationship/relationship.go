// Package relationship manages ArchiMate relationships within a workspace.
package relationship

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("relationship not found")
var ErrConflict = errors.New("relationship was modified by another request")

type Relationship struct {
	ID            uuid.UUID `json:"id"`
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	SourceID      string    `json:"source_id"`
	Type          string    `json:"type"`
	SourceElement string    `json:"source_element"`
	TargetElement string    `json:"target_element"`
	Name          string    `json:"name"`
	Documentation string    `json:"documentation"`
	Version       int       `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) List(workspaceID uuid.UUID) ([]Relationship, error) {
	rows, err := s.db.Query(`
		SELECT id, workspace_id, source_id, type, source_element, target_element,
		       name, documentation, version, created_at, updated_at
		FROM relationships WHERE workspace_id = $1 ORDER BY type, name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list relationships: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Relationship
	for rows.Next() {
		var rel Relationship
		if err := rows.Scan(&rel.ID, &rel.WorkspaceID, &rel.SourceID, &rel.Type,
			&rel.SourceElement, &rel.TargetElement, &rel.Name, &rel.Documentation,
			&rel.Version, &rel.CreatedAt, &rel.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, rel)
	}
	return out, rows.Err()
}

func (s *Store) Get(id uuid.UUID) (*Relationship, error) {
	var rel Relationship
	err := s.db.QueryRow(`
		SELECT id, workspace_id, source_id, type, source_element, target_element,
		       name, documentation, version, created_at, updated_at
		FROM relationships WHERE id = $1`, id).
		Scan(&rel.ID, &rel.WorkspaceID, &rel.SourceID, &rel.Type,
			&rel.SourceElement, &rel.TargetElement, &rel.Name, &rel.Documentation,
			&rel.Version, &rel.CreatedAt, &rel.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get relationship: %w", err)
	}
	return &rel, nil
}

func (s *Store) Create(workspaceID uuid.UUID, sourceID, typ, sourceEl, targetEl, name, documentation string) (*Relationship, error) {
	var rel Relationship
	err := s.db.QueryRow(`
		INSERT INTO relationships (workspace_id, source_id, type, source_element, target_element, name, documentation)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, workspace_id, source_id, type, source_element, target_element,
		          name, documentation, version, created_at, updated_at`,
		workspaceID, sourceID, typ, sourceEl, targetEl, name, documentation).
		Scan(&rel.ID, &rel.WorkspaceID, &rel.SourceID, &rel.Type,
			&rel.SourceElement, &rel.TargetElement, &rel.Name, &rel.Documentation,
			&rel.Version, &rel.CreatedAt, &rel.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create relationship: %w", err)
	}
	return &rel, nil
}

func (s *Store) Update(id uuid.UUID, typ, sourceEl, targetEl, name, documentation string, version int) (*Relationship, error) {
	var rel Relationship
	err := s.db.QueryRow(`
		UPDATE relationships
		SET type = $1, source_element = $2, target_element = $3,
		    name = $4, documentation = $5, version = version + 1, updated_at = now()
		WHERE id = $6 AND version = $7
		RETURNING id, workspace_id, source_id, type, source_element, target_element,
		          name, documentation, version, created_at, updated_at`,
		typ, sourceEl, targetEl, name, documentation, id, version).
		Scan(&rel.ID, &rel.WorkspaceID, &rel.SourceID, &rel.Type,
			&rel.SourceElement, &rel.TargetElement, &rel.Name, &rel.Documentation,
			&rel.Version, &rel.CreatedAt, &rel.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		if _, err2 := s.Get(id); errors.Is(err2, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrConflict
	}
	if err != nil {
		return nil, fmt.Errorf("update relationship: %w", err)
	}
	return &rel, nil
}

func (s *Store) Delete(id uuid.UUID) error {
	res, err := s.db.Exec(`DELETE FROM relationships WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete relationship: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
