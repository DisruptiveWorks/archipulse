// Package element manages ArchiMate elements within a workspace.
package element

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/pagination"
)

// Property is a key/value pair stored in element_properties.
type Property struct {
	ID          uuid.UUID  `json:"id"`
	ElementID   uuid.UUID  `json:"element_id"`
	Key         string     `json:"key"`
	Value       string     `json:"value"`
	Source      string     `json:"source"`
	CollectedAt *time.Time `json:"collected_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

var ErrNotFound = errors.New("element not found")
var ErrConflict = errors.New("element was modified by another request")

type Element struct {
	ID            uuid.UUID `json:"id"`
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	SourceID      string    `json:"source_id"`
	Type          string    `json:"type"`
	Layer         string    `json:"layer"`
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

func (s *Store) List(workspaceID uuid.UUID, p pagination.Params) ([]Element, int, error) {
	rows, err := s.db.Query(`
		SELECT id, workspace_id, source_id, type, layer, name, documentation, version, created_at, updated_at,
		       COUNT(*) OVER() AS total
		FROM elements WHERE workspace_id = $1 ORDER BY name
		LIMIT $2 OFFSET $3`, workspaceID, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("list elements: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Element
	var total int
	for rows.Next() {
		var e Element
		if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.SourceID, &e.Type, &e.Layer, &e.Name,
			&e.Documentation, &e.Version, &e.CreatedAt, &e.UpdatedAt, &total); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

// ListAll returns every element in a workspace without pagination.
// Use for internal operations (export, snapshot) that need the full dataset.
func (s *Store) ListAll(workspaceID uuid.UUID) ([]Element, error) {
	rows, err := s.db.Query(`
		SELECT id, workspace_id, source_id, type, layer, name, documentation, version, created_at, updated_at
		FROM elements WHERE workspace_id = $1 ORDER BY name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list all elements: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Element
	for rows.Next() {
		var e Element
		if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.SourceID, &e.Type, &e.Layer, &e.Name,
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
		SELECT id, workspace_id, source_id, type, layer, name, documentation, version, created_at, updated_at
		FROM elements WHERE id = $1`, id).
		Scan(&e.ID, &e.WorkspaceID, &e.SourceID, &e.Type, &e.Layer, &e.Name,
			&e.Documentation, &e.Version, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get element: %w", err)
	}
	return &e, nil
}

func (s *Store) Create(workspaceID uuid.UUID, sourceID, typ, layer, name, documentation string) (*Element, error) {
	var e Element
	err := s.db.QueryRow(`
		INSERT INTO elements (workspace_id, source_id, type, layer, name, documentation)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, workspace_id, source_id, type, layer, name, documentation, version, created_at, updated_at`,
		workspaceID, sourceID, typ, layer, name, documentation).
		Scan(&e.ID, &e.WorkspaceID, &e.SourceID, &e.Type, &e.Layer, &e.Name,
			&e.Documentation, &e.Version, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create element: %w", err)
	}
	return &e, nil
}

func (s *Store) Update(id uuid.UUID, typ, layer, name, documentation string, version int) (*Element, error) {
	var e Element
	err := s.db.QueryRow(`
		UPDATE elements
		SET type = $1, layer = $2, name = $3, documentation = $4, version = version + 1, updated_at = now()
		WHERE id = $5 AND version = $6
		RETURNING id, workspace_id, source_id, type, layer, name, documentation, version, created_at, updated_at`,
		typ, layer, name, documentation, id, version).
		Scan(&e.ID, &e.WorkspaceID, &e.SourceID, &e.Type, &e.Layer, &e.Name,
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

// ModelProperty is a key/value pair with its original AOEF propertyDefinitionRef.
type ModelProperty struct {
	DefinitionRef string
	Key           string
	Value         string
}

// InsertProperties bulk-inserts properties for an element using the provided executor
// (either *sql.DB or *sql.Tx). source is the extractor name or "model".
// collectedAt may be nil when source is "model".
func InsertProperties(exec interface {
	Exec(query string, args ...any) (sql.Result, error)
}, elementID uuid.UUID, props []ModelProperty, source string, collectedAt *time.Time) error {
	for _, p := range props {
		_, err := exec.Exec(`
			INSERT INTO element_properties (element_id, key, value, source, collected_at, definition_ref)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			elementID, p.Key, p.Value, source, collectedAt, p.DefinitionRef)
		if err != nil {
			return fmt.Errorf("insert property %q for element %s: %w", p.Key, elementID, err)
		}
	}
	return nil
}

// ListProperties returns all properties for an element, grouped by source in the returned slice.
func (s *Store) ListProperties(elementID uuid.UUID) ([]Property, error) {
	rows, err := s.db.Query(`
		SELECT id, element_id, key, value, source, collected_at, created_at
		FROM element_properties
		WHERE element_id = $1
		ORDER BY source, key`, elementID)
	if err != nil {
		return nil, fmt.Errorf("list properties: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Property
	for rows.Next() {
		var p Property
		if err := rows.Scan(&p.ID, &p.ElementID, &p.Key, &p.Value, &p.Source, &p.CollectedAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
