// Package workspace manages ArchiMate workspaces (baselines).
package workspace

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned when a workspace does not exist.
var ErrNotFound = errors.New("workspace not found")

// ErrConflict is returned when an optimistic lock conflict is detected.
var ErrConflict = errors.New("workspace was modified by another request")

// Workspace is an ArchiMate baseline (e.g. "Q1-2026-AS-IS").
type Workspace struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Purpose     string    `json:"purpose"`
	Description string    `json:"description"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Store provides CRUD operations for workspaces.
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store backed by the given database.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// List returns all workspaces ordered by name (used internally and by org admins).
func (s *Store) List() ([]Workspace, error) {
	rows, err := s.db.Query(`
		SELECT id, name, purpose, description, version, created_at, updated_at
		FROM workspaces
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Workspace
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.Purpose, &w.Description,
			&w.Version, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// ListForUser returns the workspaces the user is a member of.
// Org admins see all workspaces (pass isAdmin=true).
func (s *Store) ListForUser(userID string, isAdmin bool) ([]Workspace, error) {
	if isAdmin {
		return s.List()
	}
	rows, err := s.db.Query(`
		SELECT w.id, w.name, w.purpose, w.description, w.version, w.created_at, w.updated_at
		FROM   workspaces w
		JOIN   workspace_members wm ON wm.workspace_id = w.id
		WHERE  wm.user_id = $1
		ORDER  BY w.name`, userID)
	if err != nil {
		return nil, fmt.Errorf("list workspaces for user: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Workspace
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.Purpose, &w.Description,
			&w.Version, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// Get returns a single workspace by ID.
func (s *Store) Get(id uuid.UUID) (*Workspace, error) {
	var w Workspace
	err := s.db.QueryRow(`
		SELECT id, name, purpose, description, version, created_at, updated_at
		FROM workspaces WHERE id = $1`, id).
		Scan(&w.ID, &w.Name, &w.Purpose, &w.Description,
			&w.Version, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get workspace: %w", err)
	}
	return &w, nil
}

// Create inserts a new workspace and returns it.
func (s *Store) Create(name, purpose, description string) (*Workspace, error) {
	var w Workspace
	err := s.db.QueryRow(`
		INSERT INTO workspaces (name, purpose, description)
		VALUES ($1, $2, $3)
		RETURNING id, name, purpose, description, version, created_at, updated_at`,
		name, purpose, description).
		Scan(&w.ID, &w.Name, &w.Purpose, &w.Description,
			&w.Version, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	return &w, nil
}

// Update modifies an existing workspace. version must match the current value (optimistic locking).
func (s *Store) Update(id uuid.UUID, name, purpose, description string, version int) (*Workspace, error) {
	var w Workspace
	err := s.db.QueryRow(`
		UPDATE workspaces
		SET name = $1, purpose = $2, description = $3,
		    version = version + 1, updated_at = now()
		WHERE id = $4 AND version = $5
		RETURNING id, name, purpose, description, version, created_at, updated_at`,
		name, purpose, description, id, version).
		Scan(&w.ID, &w.Name, &w.Purpose, &w.Description,
			&w.Version, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		// Could be not found or version mismatch — check which.
		if _, err2 := s.Get(id); errors.Is(err2, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrConflict
	}
	if err != nil {
		return nil, fmt.Errorf("update workspace: %w", err)
	}
	return &w, nil
}

// Delete removes a workspace and all its elements, relationships, and diagrams (CASCADE).
func (s *Store) Delete(id uuid.UUID) error {
	res, err := s.db.Exec(`DELETE FROM workspaces WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete workspace: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
