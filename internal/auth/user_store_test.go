package auth_test

import (
	"fmt"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
)

func TestUserStore_CreateAndGetByEmail(t *testing.T) {
	conn := openTestDB(t)
	store := auth.NewUserStore(conn)
	email := fmt.Sprintf("store-create-%s@example.com", t.Name())

	hash, _ := auth.HashPassword("pass123")
	u, err := store.Create(email, hash, "viewer")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = conn.Exec("DELETE FROM users WHERE email = $1", email) })

	if u.Email != email {
		t.Errorf("Email: got %q, want %q", u.Email, email)
	}
	if u.Role != "viewer" {
		t.Errorf("Role: got %q, want viewer", u.Role)
	}
	if u.PasswordHash == nil || *u.PasswordHash != hash {
		t.Error("PasswordHash not stored correctly")
	}

	got, err := store.GetByEmail(email)
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if got.ID != u.ID {
		t.Errorf("ID mismatch: got %v, want %v", got.ID, u.ID)
	}
}

func TestUserStore_GetByEmail_NotFound(t *testing.T) {
	conn := openTestDB(t)
	store := auth.NewUserStore(conn)

	_, err := store.GetByEmail("nobody@nowhere.com")
	if err != auth.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUserStore_GetByID(t *testing.T) {
	conn := openTestDB(t)
	store := auth.NewUserStore(conn)
	email := fmt.Sprintf("store-byid-%s@example.com", t.Name())

	hash, _ := auth.HashPassword("pass")
	u, err := store.Create(email, hash, "architect")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = conn.Exec("DELETE FROM users WHERE email = $1", email) })

	got, err := store.GetByID(u.ID.String())
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Email != email {
		t.Errorf("Email: got %q, want %q", got.Email, email)
	}
}

func TestUserStore_GetByID_NotFound(t *testing.T) {
	conn := openTestDB(t)
	store := auth.NewUserStore(conn)

	_, err := store.GetByID("00000000-0000-0000-0000-000000000099")
	if err != auth.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUserStore_Exists_Empty(t *testing.T) {
	// We can't guarantee an empty DB in shared test DB, so we just verify it returns without error.
	conn := openTestDB(t)
	store := auth.NewUserStore(conn)

	_, err := store.Exists()
	if err != nil {
		t.Errorf("Exists: unexpected error: %v", err)
	}
}

func TestUserStore_Exists_AfterCreate(t *testing.T) {
	conn := openTestDB(t)
	store := auth.NewUserStore(conn)
	email := fmt.Sprintf("store-exists-%s@example.com", t.Name())

	hash, _ := auth.HashPassword("pass")
	if _, err := store.Create(email, hash, "viewer"); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = conn.Exec("DELETE FROM users WHERE email = $1", email) })

	ok, err := store.Exists()
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !ok {
		t.Error("expected Exists to return true after creating a user")
	}
}

func TestUserStore_UpdatePasswordHash(t *testing.T) {
	conn := openTestDB(t)
	store := auth.NewUserStore(conn)
	email := fmt.Sprintf("store-updatehash-%s@example.com", t.Name())

	hash, _ := auth.HashPassword("oldpass")
	u, err := store.Create(email, hash, "viewer")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = conn.Exec("DELETE FROM users WHERE email = $1", email) })

	newHash, _ := auth.HashPassword("newpass")
	if err := store.UpdatePasswordHash(u.ID.String(), newHash); err != nil {
		t.Fatalf("UpdatePasswordHash: %v", err)
	}

	got, _ := store.GetByEmail(email)
	if got.PasswordHash == nil || !auth.CheckPassword(*got.PasswordHash, "newpass") {
		t.Error("password hash was not updated correctly")
	}
}

func TestUserStore_UpdateRole(t *testing.T) {
	conn := openTestDB(t)
	store := auth.NewUserStore(conn)
	email := fmt.Sprintf("store-updaterole-%s@example.com", t.Name())

	hash, _ := auth.HashPassword("pass")
	u, err := store.Create(email, hash, "viewer")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = conn.Exec("DELETE FROM users WHERE email = $1", email) })

	if err := store.UpdateRole(u.ID.String(), "architect"); err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}

	got, _ := store.GetByEmail(email)
	if got.Role != "architect" {
		t.Errorf("Role: got %q, want architect", got.Role)
	}
}

func TestUserStore_Create_OIDCUser_NoPassword(t *testing.T) {
	conn := openTestDB(t)
	store := auth.NewUserStore(conn)
	email := fmt.Sprintf("store-oidc-%s@example.com", t.Name())

	u, err := store.Create(email, "", "viewer")
	if err != nil {
		t.Fatalf("Create OIDC user: %v", err)
	}
	t.Cleanup(func() { _, _ = conn.Exec("DELETE FROM users WHERE email = $1", email) })

	if u.PasswordHash != nil {
		t.Error("expected nil PasswordHash for OIDC user")
	}
}
