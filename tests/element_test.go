package tests

import (
	"errors"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/element"
	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

func TestElementStore_CRUD(t *testing.T) {
	conn := openTestDB(t)
	wsStore := workspace.NewStore(conn)
	eStore := element.NewStore(conn)

	ws, err := wsStore.Create("test-element-ws", "as-is", "")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	t.Cleanup(func() { _ = wsStore.Delete(ws.ID) })

	// Create
	e, err := eStore.Create(ws.ID, "src-001", "ApplicationComponent", "Application", "PaymentService", "docs")
	if err != nil {
		t.Fatalf("Create element: %v", err)
	}
	if e.Name != "PaymentService" || e.Version != 1 {
		t.Errorf("unexpected element after create: %+v", e)
	}

	// Get
	got, err := eStore.Get(e.ID)
	if err != nil {
		t.Fatalf("Get element: %v", err)
	}
	if got.SourceID != "src-001" {
		t.Errorf("Get returned wrong SourceID: %q", got.SourceID)
	}

	// List
	list, err := eStore.List(ws.ID)
	if err != nil {
		t.Fatalf("List elements: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List count = %d, want 1", len(list))
	}

	// Update
	updated, err := eStore.Update(e.ID, "ApplicationComponent", "Application", "PaymentServiceV2", "updated", e.Version)
	if err != nil {
		t.Fatalf("Update element: %v", err)
	}
	if updated.Name != "PaymentServiceV2" || updated.Version != 2 {
		t.Errorf("unexpected element after update: %+v", updated)
	}

	// Optimistic lock conflict
	_, err = eStore.Update(e.ID, "ApplicationComponent", "Application", "conflict", "", e.Version)
	if !errors.Is(err, element.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}

	// Delete
	if err := eStore.Delete(e.ID); err != nil {
		t.Fatalf("Delete element: %v", err)
	}
	_, err = eStore.Get(e.ID)
	if !errors.Is(err, element.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}
