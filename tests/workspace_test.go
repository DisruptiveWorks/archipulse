package tests

import (
	"errors"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

func TestWorkspaceStore_CRUD(t *testing.T) {
	conn := openTestDB(t)
	store := workspace.NewStore(conn)

	// Create
	ws, err := store.Create("test-ws", "as-is", "integration test workspace")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if ws.Name != "test-ws" || ws.Version != 1 {
		t.Errorf("unexpected workspace after create: %+v", ws)
	}
	t.Cleanup(func() { _ = store.Delete(ws.ID) })

	// Get
	got, err := store.Get(ws.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != ws.ID {
		t.Errorf("Get returned wrong ID: %v", got.ID)
	}

	// List
	list, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, w := range list {
		if w.ID == ws.ID {
			found = true
		}
	}
	if !found {
		t.Error("created workspace not found in List")
	}

	// Update
	updated, err := store.Update(ws.ID, "test-ws-updated", "to-be", "updated desc", ws.Version)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "test-ws-updated" || updated.Version != 2 {
		t.Errorf("unexpected workspace after update: %+v", updated)
	}

	// Optimistic lock conflict
	_, err = store.Update(ws.ID, "conflict", "as-is", "", ws.Version) // stale version
	if !errors.Is(err, workspace.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}

	// Delete
	if err := store.Delete(ws.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Get after delete
	_, err = store.Get(ws.ID)
	if !errors.Is(err, workspace.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestWorkspaceStore_GetNotFound(t *testing.T) {
	conn := openTestDB(t)
	store := workspace.NewStore(conn)

	_, err := store.Get(nonExistentUUID())
	if !errors.Is(err, workspace.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
