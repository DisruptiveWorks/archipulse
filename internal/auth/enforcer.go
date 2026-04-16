package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Workspace roles — ordered from most to least permissive.
const (
	RoleOwner  = "owner"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

// roleRank maps a workspace role to a numeric rank (higher = more permissions).
var roleRank = map[string]int{
	RoleOwner:  3,
	RoleEditor: 2,
	RoleViewer: 1,
}

// ErrNoMembership is returned when a user has no membership in a workspace.
var ErrNoMembership = errors.New("no workspace membership")

// WorkspaceMember represents a single membership row.
type WorkspaceMember struct {
	UserID    uuid.UUID  `json:"user_id"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	InvitedBy *uuid.UUID `json:"invited_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// Enforcer provides workspace membership checks and CRUD.
// Enforcement logic is implemented here in Go; the Casbin model is kept
// only as a schema reference and is not used for active enforcement.
type Enforcer struct {
	db *sql.DB
}

// NewEnforcer creates an Enforcer.
func NewEnforcer(db *sql.DB, _ *Config) (*Enforcer, error) {
	return &Enforcer{db: db}, nil
}

// ── Enforcement ───────────────────────────────────────────────────────────────

// WorkspaceRole returns the role the user holds in the workspace,
// or ErrNoMembership if there is none.
func (en *Enforcer) WorkspaceRole(userID, workspaceID string) (string, error) {
	var role string
	err := en.db.QueryRow(
		`SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, userID,
	).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNoMembership
	}
	return role, err
}

// hasRank reports whether the user's role in the workspace satisfies minRole.
func (en *Enforcer) hasRank(userID, workspaceID, minRole string) (bool, error) {
	role, err := en.WorkspaceRole(userID, workspaceID)
	if errors.Is(err, ErrNoMembership) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return roleRank[role] >= roleRank[minRole], nil
}

// CanView reports whether the user may read from the workspace.
func (en *Enforcer) CanView(userID, workspaceID string) (bool, error) {
	return en.hasRank(userID, workspaceID, RoleViewer)
}

// CanEdit reports whether the user may create/update resources in the workspace.
func (en *Enforcer) CanEdit(userID, workspaceID string) (bool, error) {
	return en.hasRank(userID, workspaceID, RoleEditor)
}

// CanManage reports whether the user may invite/remove members or delete the workspace.
func (en *Enforcer) CanManage(userID, workspaceID string) (bool, error) {
	return en.hasRank(userID, workspaceID, RoleOwner)
}

// ── Membership CRUD ───────────────────────────────────────────────────────────

// ListMembers returns all members of the workspace, joined with their email.
func (en *Enforcer) ListMembers(workspaceID string) ([]WorkspaceMember, error) {
	rows, err := en.db.Query(`
		SELECT wm.user_id, u.email, wm.role, wm.invited_by, wm.created_at
		FROM   workspace_members wm
		JOIN   users u ON u.id = wm.user_id
		WHERE  wm.workspace_id = $1
		ORDER  BY wm.created_at`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []WorkspaceMember
	for rows.Next() {
		var m WorkspaceMember
		var invBy sql.NullString
		if err := rows.Scan(&m.UserID, &m.Email, &m.Role, &invBy, &m.CreatedAt); err != nil {
			return nil, err
		}
		if invBy.Valid {
			id, _ := uuid.Parse(invBy.String)
			m.InvitedBy = &id
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// AddMember upserts a workspace membership.
// If the user is already a member, their role is updated.
func (en *Enforcer) AddMember(workspaceID, userID, role, invitedBy string) error {
	if _, ok := roleRank[role]; !ok {
		return fmt.Errorf("invalid workspace role %q", role)
	}
	var inv interface{} = nil
	if invitedBy != "" {
		inv = invitedBy
	}
	_, err := en.db.Exec(`
		INSERT INTO workspace_members (workspace_id, user_id, role, invited_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		workspaceID, userID, role, inv)
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

// UpdateMemberRole changes an existing member's role.
// Returns ErrNoMembership if the user is not a member.
func (en *Enforcer) UpdateMemberRole(workspaceID, userID, role string) error {
	if _, ok := roleRank[role]; !ok {
		return fmt.Errorf("invalid workspace role %q", role)
	}
	res, err := en.db.Exec(`
		UPDATE workspace_members SET role = $1
		WHERE workspace_id = $2 AND user_id = $3`,
		role, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNoMembership
	}
	return nil
}

// RemoveMember removes a user from a workspace.
// Returns ErrNoMembership if the user was not a member.
func (en *Enforcer) RemoveMember(workspaceID, userID string) error {
	res, err := en.db.Exec(`
		DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		workspaceID, userID)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNoMembership
	}
	return nil
}

// ListUserWorkspaces returns all workspace IDs the user is a member of.
func (en *Enforcer) ListUserWorkspaces(userID string) ([]string, error) {
	rows, err := en.db.Query(
		`SELECT workspace_id FROM workspace_members WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// SeedOwner adds the user as owner of the workspace if they are not already a member.
func (en *Enforcer) SeedOwner(workspaceID, userID string) error {
	_, err := en.db.Exec(`
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'owner')
		ON CONFLICT DO NOTHING`,
		workspaceID, userID)
	return err
}
