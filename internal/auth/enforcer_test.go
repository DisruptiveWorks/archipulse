package auth_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
)

func TestEnforcer_WorkspaceRole(t *testing.T) {
	conn := openTestDB(t)
	svc := newTestService(t, conn)

	email := fmt.Sprintf("enf-role-%s@example.com", t.Name())
	hash, _ := auth.HashPassword("pass")
	u, err := svc.Users.Create(email, hash, "member")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = conn.Exec("DELETE FROM users WHERE email = $1", email) })

	wsID := mustCreateWorkspace(t, conn)

	// No membership yet.
	_, err = svc.Enforcer.WorkspaceRole(u.ID.String(), wsID)
	if err != auth.ErrNoMembership {
		t.Errorf("expected ErrNoMembership before adding, got %v", err)
	}

	if err := svc.Enforcer.AddMember(wsID, u.ID.String(), auth.RoleViewer, ""); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	role, err := svc.Enforcer.WorkspaceRole(u.ID.String(), wsID)
	if err != nil {
		t.Fatalf("WorkspaceRole: %v", err)
	}
	if role != auth.RoleViewer {
		t.Errorf("role: got %q, want %q", role, auth.RoleViewer)
	}
}

func TestEnforcer_CanView_CanEdit_CanManage(t *testing.T) {
	conn := openTestDB(t)
	svc := newTestService(t, conn)

	email := fmt.Sprintf("enf-perms-%s@example.com", t.Name())
	hash, _ := auth.HashPassword("pass")
	u, err := svc.Users.Create(email, hash, "member")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = conn.Exec("DELETE FROM users WHERE email = $1", email) })

	wsID := mustCreateWorkspace(t, conn)

	// viewer: can view, cannot edit or manage.
	if err := svc.Enforcer.AddMember(wsID, u.ID.String(), auth.RoleViewer, ""); err != nil {
		t.Fatalf("AddMember viewer: %v", err)
	}
	assertCan(t, svc, u.ID.String(), wsID, "viewer", true, false, false)

	// Upgrade to editor.
	if err := svc.Enforcer.UpdateMemberRole(wsID, u.ID.String(), auth.RoleEditor); err != nil {
		t.Fatalf("UpdateMemberRole editor: %v", err)
	}
	assertCan(t, svc, u.ID.String(), wsID, "editor", true, true, false)

	// Upgrade to owner.
	if err := svc.Enforcer.UpdateMemberRole(wsID, u.ID.String(), auth.RoleOwner); err != nil {
		t.Fatalf("UpdateMemberRole owner: %v", err)
	}
	assertCan(t, svc, u.ID.String(), wsID, "owner", true, true, true)
}

func TestEnforcer_RemoveMember(t *testing.T) {
	conn := openTestDB(t)
	svc := newTestService(t, conn)

	email := fmt.Sprintf("enf-remove-%s@example.com", t.Name())
	hash, _ := auth.HashPassword("pass")
	u, err := svc.Users.Create(email, hash, "member")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = conn.Exec("DELETE FROM users WHERE email = $1", email) })

	wsID := mustCreateWorkspace(t, conn)

	if err := svc.Enforcer.AddMember(wsID, u.ID.String(), auth.RoleEditor, ""); err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	if err := svc.Enforcer.RemoveMember(wsID, u.ID.String()); err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}

	_, err = svc.Enforcer.WorkspaceRole(u.ID.String(), wsID)
	if err != auth.ErrNoMembership {
		t.Errorf("expected ErrNoMembership after remove, got %v", err)
	}

	// Removing again should return ErrNoMembership.
	if err := svc.Enforcer.RemoveMember(wsID, u.ID.String()); err != auth.ErrNoMembership {
		t.Errorf("expected ErrNoMembership on double remove, got %v", err)
	}
}

func TestEnforcer_ListMembers(t *testing.T) {
	conn := openTestDB(t)
	svc := newTestService(t, conn)

	email1 := fmt.Sprintf("enf-list1-%s@example.com", t.Name())
	email2 := fmt.Sprintf("enf-list2-%s@example.com", t.Name())
	hash, _ := auth.HashPassword("pass")
	u1, _ := svc.Users.Create(email1, hash, "member")
	u2, _ := svc.Users.Create(email2, hash, "member")
	t.Cleanup(func() {
		_, _ = conn.Exec("DELETE FROM users WHERE email IN ($1, $2)", email1, email2)
	})

	wsID := mustCreateWorkspace(t, conn)

	_ = svc.Enforcer.AddMember(wsID, u1.ID.String(), auth.RoleOwner, "")
	_ = svc.Enforcer.AddMember(wsID, u2.ID.String(), auth.RoleViewer, "")

	members, err := svc.Enforcer.ListMembers(wsID)
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	if len(members) < 2 {
		t.Errorf("expected at least 2 members, got %d", len(members))
	}
}

func TestEnforcer_SeedOwner(t *testing.T) {
	conn := openTestDB(t)
	svc := newTestService(t, conn)

	email := fmt.Sprintf("enf-seed-%s@example.com", t.Name())
	hash, _ := auth.HashPassword("pass")
	u, err := svc.Users.Create(email, hash, "admin")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = conn.Exec("DELETE FROM users WHERE email = $1", email) })

	wsID := mustCreateWorkspace(t, conn)

	if err := svc.Enforcer.SeedOwner(wsID, u.ID.String()); err != nil {
		t.Fatalf("SeedOwner: %v", err)
	}
	role, err := svc.Enforcer.WorkspaceRole(u.ID.String(), wsID)
	if err != nil {
		t.Fatalf("WorkspaceRole after seed: %v", err)
	}
	if role != auth.RoleOwner {
		t.Errorf("role: got %q, want owner", role)
	}

	// Second seed call must be a no-op (ON CONFLICT DO NOTHING).
	if err := svc.Enforcer.SeedOwner(wsID, u.ID.String()); err != nil {
		t.Fatalf("SeedOwner (idempotent): %v", err)
	}
}

// assertCan checks CanView/CanEdit/CanManage against expectations.
func assertCan(t *testing.T, svc *auth.Service, userID, wsID, label string, wantView, wantEdit, wantManage bool) {
	t.Helper()
	canView, err := svc.Enforcer.CanView(userID, wsID)
	if err != nil {
		t.Errorf("[%s] CanView error: %v", label, err)
	}
	if canView != wantView {
		t.Errorf("[%s] CanView: got %v, want %v", label, canView, wantView)
	}

	canEdit, err := svc.Enforcer.CanEdit(userID, wsID)
	if err != nil {
		t.Errorf("[%s] CanEdit error: %v", label, err)
	}
	if canEdit != wantEdit {
		t.Errorf("[%s] CanEdit: got %v, want %v", label, canEdit, wantEdit)
	}

	canManage, err := svc.Enforcer.CanManage(userID, wsID)
	if err != nil {
		t.Errorf("[%s] CanManage error: %v", label, err)
	}
	if canManage != wantManage {
		t.Errorf("[%s] CanManage: got %v, want %v", label, canManage, wantManage)
	}
}

// mustCreateWorkspace inserts a workspace and returns its UUID string.
// Cleanup is registered automatically.
func mustCreateWorkspace(t *testing.T, conn *sql.DB) string {
	t.Helper()
	var id string
	err := conn.QueryRow(
		`INSERT INTO workspaces (name, purpose) VALUES ($1, 'other') RETURNING id`,
		fmt.Sprintf("test-ws-%s", t.Name()),
	).Scan(&id)
	if err != nil {
		t.Fatalf("create test workspace: %v", err)
	}
	t.Cleanup(func() { _, _ = conn.Exec("DELETE FROM workspaces WHERE id = $1", id) })
	return id
}
