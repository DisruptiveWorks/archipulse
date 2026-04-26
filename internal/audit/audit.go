// Package audit records workspace-scoped user actions.
package audit

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/pagination"
)

// Action constants.

const (
	ActionCreate           = "create"
	ActionUpdate           = "update"
	ActionDelete           = "delete"
	ActionImport           = "import"
	ActionAddMember        = "add_member"
	ActionRemoveMember     = "remove_member"
	ActionUpdateMemberRole = "update_member_role"
	ActionCreateSnapshot   = "create_snapshot"
	ActionRestoreSnapshot  = "restore_snapshot"
)

// EntityType constants.
const (
	EntityElement      = "element"
	EntityRelationship = "relationship"
	EntityDiagram      = "diagram"
	EntityWorkspace    = "workspace"
	EntityMember       = "member"
	EntitySnapshot     = "snapshot"
	EntitySavedView    = "saved_view"
)

// Event is a single audit log entry.
type Event struct {
	ID          uuid.UUID       `json:"id"`
	WorkspaceID uuid.UUID       `json:"workspace_id"`
	UserID      string          `json:"user_id"`
	UserEmail   string          `json:"user_email"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entity_type"`
	EntityID    string          `json:"entity_id,omitempty"`
	EntityName  string          `json:"entity_name,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// Store provides audit log operations.
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// RecordParams holds the fields needed to record an event.
type RecordParams struct {
	WorkspaceID uuid.UUID
	UserID      string
	UserEmail   string
	Action      string
	EntityType  string
	EntityID    string
	EntityName  string
	Meta        map[string]any
}

// Record inserts a new audit event. Errors are non-fatal — the caller should
// log them but not fail the request.
func (s *Store) Record(p RecordParams) error {
	var metaJSON []byte
	if len(p.Meta) > 0 {
		var err error
		metaJSON, err = json.Marshal(p.Meta)
		if err != nil {
			return fmt.Errorf("audit marshal meta: %w", err)
		}
	}
	_, err := s.db.Exec(`
		INSERT INTO workspace_events
		  (workspace_id, user_id, user_email, action, entity_type, entity_id, entity_name, meta)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		p.WorkspaceID, p.UserID, p.UserEmail, p.Action,
		p.EntityType, nullStr(p.EntityID), nullStr(p.EntityName), nullBytes(metaJSON),
	)
	if err != nil {
		return fmt.Errorf("audit record: %w", err)
	}
	return nil
}

// List returns events for a workspace (newest first).
func (s *Store) List(workspaceID uuid.UUID, p pagination.Params) ([]Event, int, error) {
	rows, err := s.db.Query(`
		SELECT id, workspace_id, user_id, user_email, action, entity_type,
		       COALESCE(entity_id, ''), COALESCE(entity_name, ''),
		       meta, created_at,
		       COUNT(*) OVER() AS total
		FROM   workspace_events
		WHERE  workspace_id = $1
		ORDER  BY created_at DESC
		LIMIT  $2 OFFSET $3`, workspaceID, p.Limit, p.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("audit list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Event
	var total int
	for rows.Next() {
		var e Event
		var meta []byte
		if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.UserID, &e.UserEmail,
			&e.Action, &e.EntityType, &e.EntityID, &e.EntityName,
			&meta, &e.CreatedAt, &total); err != nil {
			return nil, 0, err
		}
		if len(meta) > 0 {
			e.Meta = json.RawMessage(meta)
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

// ── helpers ───────────────────────────────────────────────────────────────────

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullBytes(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}
