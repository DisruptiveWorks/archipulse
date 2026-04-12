// Package diagramfolder manages the folder hierarchy for ArchiMate diagram views.
package diagramfolder

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("folder not found")

// Folder represents a node in the diagram folder tree.
type Folder struct {
	ID          uuid.UUID  `json:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Name        string     `json:"name"`
	SourceID    string     `json:"source_id"`
	Position    int        `json:"position"`
}

// Store handles persistence for diagram folders.
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Upsert inserts or updates a folder by (workspace_id, source_id).
// Returns the folder's UUID.
func (s *Store) Upsert(workspaceID uuid.UUID, parentID *uuid.UUID, name, sourceID string, position int) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.db.QueryRow(`
		INSERT INTO diagram_folders (workspace_id, parent_id, name, source_id, position)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (workspace_id, source_id) DO UPDATE
		  SET parent_id = EXCLUDED.parent_id,
		      name      = EXCLUDED.name,
		      position  = EXCLUDED.position
		RETURNING id`,
		workspaceID, parentID, name, sourceID, position,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert folder %q: %w", sourceID, err)
	}
	return id, nil
}

// ListByWorkspace returns all folders for a workspace (unordered — caller builds tree).
func (s *Store) ListByWorkspace(workspaceID uuid.UUID) ([]Folder, error) {
	rows, err := s.db.Query(`
		SELECT id, workspace_id, parent_id, name, source_id, position
		FROM diagram_folders
		WHERE workspace_id = $1
		ORDER BY position`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var folders []Folder
	for rows.Next() {
		var f Folder
		var parentID sql.NullString
		if err := rows.Scan(&f.ID, &f.WorkspaceID, &parentID, &f.Name, &f.SourceID, &f.Position); err != nil {
			return nil, fmt.Errorf("scan folder: %w", err)
		}
		if parentID.Valid {
			id, _ := uuid.Parse(parentID.String)
			f.ParentID = &id
		}
		folders = append(folders, f)
	}
	return folders, rows.Err()
}

// DeleteByWorkspace removes all folders for a workspace (used before re-import).
func (s *Store) DeleteByWorkspace(workspaceID uuid.UUID) error {
	_, err := s.db.Exec(`DELETE FROM diagram_folders WHERE workspace_id = $1`, workspaceID)
	return err
}
