// Package snapshot manages point-in-time exports of a workspace model.
package snapshot

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned when a snapshot does not exist.
var ErrNotFound = errors.New("snapshot not found")

// TriggerImport and TriggerManual identify what caused the snapshot.
const (
	TriggerImport = "import"
	TriggerManual = "manual"
)

// Snapshot is a point-in-time export of a workspace.
// Payload holds the full AOEF XML — it is only populated by Get, not List.
type Snapshot struct {
	ID             uuid.UUID `json:"id"`
	WorkspaceID    uuid.UUID `json:"workspace_id"`
	CreatedBy      string    `json:"created_by"`
	CreatedByEmail string    `json:"created_by_email"`
	Label          string    `json:"label,omitempty"`
	Trigger        string    `json:"trigger"`
	Payload        string    `json:"-"` // AOEF XML, never sent to the client
	CreatedAt      time.Time `json:"created_at"`
}

// Store provides CRUD operations for snapshots.
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// List returns all snapshots for a workspace (newest first), without payload.
func (s *Store) List(workspaceID uuid.UUID) ([]Snapshot, error) {
	rows, err := s.db.Query(`
		SELECT id, workspace_id, created_by, created_by_email,
		       COALESCE(label, ''), trigger, created_at
		FROM   workspace_snapshots
		WHERE  workspace_id = $1
		ORDER  BY created_at DESC`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Snapshot
	for rows.Next() {
		var snap Snapshot
		if err := rows.Scan(&snap.ID, &snap.WorkspaceID, &snap.CreatedBy,
			&snap.CreatedByEmail, &snap.Label, &snap.Trigger, &snap.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, snap)
	}
	return out, rows.Err()
}

// Get returns a single snapshot including its AOEF XML payload.
func (s *Store) Get(id uuid.UUID) (*Snapshot, error) {
	var snap Snapshot
	err := s.db.QueryRow(`
		SELECT id, workspace_id, created_by, created_by_email,
		       COALESCE(label, ''), trigger, payload, created_at
		FROM   workspace_snapshots
		WHERE  id = $1`, id).
		Scan(&snap.ID, &snap.WorkspaceID, &snap.CreatedBy, &snap.CreatedByEmail,
			&snap.Label, &snap.Trigger, &snap.Payload, &snap.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get snapshot: %w", err)
	}
	return &snap, nil
}

// Create inserts a new snapshot and returns it (without payload).
// payload must be valid AOEF XML.
func (s *Store) Create(workspaceID uuid.UUID, createdBy, createdByEmail, label, trigger, payload string) (*Snapshot, error) {
	var snap Snapshot
	err := s.db.QueryRow(`
		INSERT INTO workspace_snapshots
		  (workspace_id, created_by, created_by_email, label, trigger, payload)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, workspace_id, created_by, created_by_email,
		          COALESCE(label, ''), trigger, created_at`,
		workspaceID, createdBy, createdByEmail, nullStr(label), trigger, payload).
		Scan(&snap.ID, &snap.WorkspaceID, &snap.CreatedBy, &snap.CreatedByEmail,
			&snap.Label, &snap.Trigger, &snap.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create snapshot: %w", err)
	}
	return &snap, nil
}

// Delete removes a snapshot by ID.
func (s *Store) Delete(id uuid.UUID) error {
	res, err := s.db.Exec(`DELETE FROM workspace_snapshots WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete snapshot: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
