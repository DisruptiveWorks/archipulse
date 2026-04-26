package tests

import (
	"errors"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/relationship"
	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

func TestRelationshipStore_CRUD(t *testing.T) {
	conn := openTestDB(t)
	wsStore := workspace.NewStore(conn)
	rStore := relationship.NewStore(conn)

	ws, err := wsStore.Create("test-rel-ws", "as-is", "")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	t.Cleanup(func() { _ = wsStore.Delete(ws.ID) })

	// Create
	rel, err := rStore.Create(ws.ID, "rel-001", "AssociationRelationship", "src-a", "src-b", "calls", "")
	if err != nil {
		t.Fatalf("Create relationship: %v", err)
	}
	if rel.Type != "AssociationRelationship" || rel.Version != 1 {
		t.Errorf("unexpected relationship after create: %+v", rel)
	}

	// Get
	got, err := rStore.Get(rel.ID)
	if err != nil {
		t.Fatalf("Get relationship: %v", err)
	}
	if got.SourceElement != "src-a" || got.TargetElement != "src-b" {
		t.Errorf("unexpected source/target: %q/%q", got.SourceElement, got.TargetElement)
	}

	// List
	list, err := rStore.ListAll(ws.ID)
	if err != nil {
		t.Fatalf("List relationships: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List count = %d, want 1", len(list))
	}

	// Update
	updated, err := rStore.Update(rel.ID, "CompositionRelationship", "src-a", "src-b", "contains", "", rel.Version)
	if err != nil {
		t.Fatalf("Update relationship: %v", err)
	}
	if updated.Type != "CompositionRelationship" || updated.Version != 2 {
		t.Errorf("unexpected relationship after update: %+v", updated)
	}

	// Optimistic lock conflict
	_, err = rStore.Update(rel.ID, "AssociationRelationship", "src-a", "src-b", "", "", rel.Version)
	if !errors.Is(err, relationship.ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}

	// Delete
	if err := rStore.Delete(rel.ID); err != nil {
		t.Fatalf("Delete relationship: %v", err)
	}
	_, err = rStore.Get(rel.ID)
	if !errors.Is(err, relationship.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}
