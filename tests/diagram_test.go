package tests

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/diagram"
	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

func TestDiagramStore_CRUD(t *testing.T) {
	conn := openTestDB(t)
	wsStore := workspace.NewStore(conn)
	dStore := diagram.NewStore(conn)

	ws, err := wsStore.Create("test-diag-ws", "as-is", "")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	t.Cleanup(func() { _ = wsStore.Delete(ws.ID) })

	layout := json.RawMessage(`{"nodes":[{"elementRef":"src-001","x":0,"y":0,"w":120,"h":55}]}`)

	// Create
	d, err := dStore.Create(ws.ID, "view-001", "Application Overview", "desc", layout)
	if err != nil {
		t.Fatalf("Create diagram: %v", err)
	}
	if d.Name != "Application Overview" || d.Version != 1 {
		t.Errorf("unexpected diagram after create: %+v", d)
	}

	// Get
	got, err := dStore.Get(d.ID)
	if err != nil {
		t.Fatalf("Get diagram: %v", err)
	}
	if got.SourceID != "view-001" {
		t.Errorf("Get returned wrong SourceID: %q", got.SourceID)
	}

	// List
	list, err := dStore.List(ws.ID)
	if err != nil {
		t.Fatalf("List diagrams: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List count = %d, want 1", len(list))
	}

	// Update
	updated, err := dStore.Update(d.ID, "Updated Overview", "new desc", layout, d.Version)
	if err != nil {
		t.Fatalf("Update diagram: %v", err)
	}
	if updated.Name != "Updated Overview" || updated.Version != 2 {
		t.Errorf("unexpected diagram after update: %+v", updated)
	}

	// Optimistic lock conflict
	_, err = dStore.Update(d.ID, "conflict", "", nil, d.Version)
	if !errors.Is(err, diagram.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}

	// Delete
	if err := dStore.Delete(d.ID); err != nil {
		t.Fatalf("Delete diagram: %v", err)
	}
	_, err = dStore.Get(d.ID)
	if !errors.Is(err, diagram.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}
