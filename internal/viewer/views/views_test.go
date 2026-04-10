package views_test

import (
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/viewer/views"
	"github.com/google/uuid"
)

// emptyWS returns a UUID guaranteed not to exist in the test DB.
func emptyWS() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000099")
}

// --- CapabilityTreeData ---

func TestCapabilityTreeData_EmptyWorkspace(t *testing.T) {
	conn := openTestDB(t)
	nodes, err := views.CapabilityTreeData(conn, emptyWS())
	if err != nil {
		t.Fatalf("CapabilityTreeData: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(nodes))
	}
}

func TestCapabilityTreeData_ReturnsSlice(t *testing.T) {
	conn := openTestDB(t)
	nodes, err := views.CapabilityTreeData(conn, emptyWS())
	if err != nil {
		t.Fatalf("CapabilityTreeData: %v", err)
	}
	// Result must be a non-nil slice (not nil) for JSON marshaling.
	if nodes == nil {
		t.Error("expected non-nil slice, got nil")
	}
}

// --- AppCatalogueEntries ---

func TestAppCatalogueEntries_EmptyWorkspace(t *testing.T) {
	conn := openTestDB(t)
	data, err := views.AppCatalogueEntries(conn, emptyWS())
	if err != nil {
		t.Fatalf("AppCatalogueEntries: %v", err)
	}
	if len(data.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(data.Entries))
	}
	if data.PropertyKeys == nil {
		t.Error("PropertyKeys must be non-nil")
	}
}

// --- TechCatalogueEntries ---

func TestTechCatalogueEntries_EmptyWorkspace(t *testing.T) {
	conn := openTestDB(t)
	data, err := views.TechCatalogueEntries(conn, emptyWS())
	if err != nil {
		t.Fatalf("TechCatalogueEntries: %v", err)
	}
	if len(data.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(data.Entries))
	}
}

// --- ApplicationLandscape ---

func TestApplicationLandscape_EmptyWorkspace(t *testing.T) {
	conn := openTestDB(t)
	title, cols, data, err := views.ApplicationLandscape(conn, emptyWS())
	if err != nil {
		t.Fatalf("ApplicationLandscape: %v", err)
	}
	if title != "Application Landscape" {
		t.Errorf("title: got %q, want %q", title, "Application Landscape")
	}
	if len(cols) == 0 {
		t.Error("expected column headers, got none")
	}
	if len(data) != 0 {
		t.Errorf("expected 0 rows, got %d", len(data))
	}
}

// --- ApplicationDependency ---

func TestApplicationDependency_EmptyWorkspace(t *testing.T) {
	conn := openTestDB(t)
	graph, err := views.ApplicationDependency(conn, emptyWS())
	if err != nil {
		t.Fatalf("ApplicationDependency: %v", err)
	}
	if graph == nil {
		t.Fatal("expected non-nil graph")
	}
	if graph.Nodes == nil {
		t.Error("Nodes must be non-nil")
	}
	if graph.Edges == nil {
		t.Error("Edges must be non-nil")
	}
	if len(graph.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(graph.Nodes))
	}
}

// --- ElementCatalogue ---

func TestElementCatalogue_EmptyWorkspace(t *testing.T) {
	conn := openTestDB(t)
	title, cols, data, err := views.ElementCatalogue(conn, emptyWS())
	if err != nil {
		t.Fatalf("ElementCatalogue: %v", err)
	}
	if title == "" {
		t.Error("expected non-empty title")
	}
	if len(cols) == 0 {
		t.Error("expected column headers")
	}
	if len(data) != 0 {
		t.Errorf("expected 0 rows, got %d", len(data))
	}
}

// --- ApplicationDashboard with capability filter ---

func TestApplicationDashboard_WithCapabilityFilter_EmptyWorkspace(t *testing.T) {
	conn := openTestDB(t)
	data, err := views.ApplicationDashboard(conn, emptyWS(), "NonExistentCapability")
	if err != nil {
		t.Fatalf("ApplicationDashboard: %v", err)
	}
	if data.TotalApps != 0 {
		t.Errorf("expected 0 apps, got %d", data.TotalApps)
	}
	// Capabilities may be nil when workspace has no capability elements — that's fine.
	if data.TotalApps != 0 {
		t.Errorf("expected 0 apps with capability filter, got %d", data.TotalApps)
	}
}
