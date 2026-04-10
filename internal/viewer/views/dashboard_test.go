package views_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/DisruptiveWorks/archipulse/internal/db"
	"github.com/DisruptiveWorks/archipulse/internal/viewer/views"
	"github.com/google/uuid"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	_ = godotenv.Load("../../../.env")
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set — skipping integration test")
	}
	conn, err := db.Connect()
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := db.Migrate(conn, "../../../migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestPropertyDistributions_Empty(t *testing.T) {
	result := views.PropertyDistributions([]views.AppEntry{})
	if len(result) != 0 {
		t.Errorf("expected empty map for empty input, got %v", result)
	}
}

func TestPropertyDistributions_SingleApp(t *testing.T) {
	apps := []views.AppEntry{
		{ID: "1", Name: "App A", Type: "ApplicationComponent", Properties: map[string]string{
			"lifecycle_status": "Production",
		}},
	}
	result := views.PropertyDistributions(apps)
	buckets, ok := result["lifecycle_status"]
	if !ok {
		t.Fatal("expected lifecycle_status key in result")
	}
	if len(buckets) != 1 || buckets[0].Value != "Production" || buckets[0].Count != 1 {
		t.Errorf("unexpected buckets: %v", buckets)
	}
}

func TestPropertyDistributions_MultipleApps(t *testing.T) {
	apps := []views.AppEntry{
		{ID: "1", Properties: map[string]string{"lifecycle_status": "Production"}},
		{ID: "2", Properties: map[string]string{"lifecycle_status": "Production"}},
		{ID: "3", Properties: map[string]string{"lifecycle_status": "Deprecated"}},
		{ID: "4", Properties: map[string]string{}}, // no lifecycle_status → (unset)
	}
	result := views.PropertyDistributions(apps)
	buckets := result["lifecycle_status"]

	counts := map[string]int{}
	for _, b := range buckets {
		counts[b.Value] = b.Count
	}
	if counts["Production"] != 2 {
		t.Errorf("Production count: got %d, want 2", counts["Production"])
	}
	if counts["Deprecated"] != 1 {
		t.Errorf("Deprecated count: got %d, want 1", counts["Deprecated"])
	}
	if counts["(unset)"] != 1 {
		t.Errorf("(unset) count: got %d, want 1", counts["(unset)"])
	}
	// Production should come first (highest count)
	if len(buckets) > 0 && buckets[0].Value != "Production" {
		t.Errorf("expected Production first (highest count), got %q", buckets[0].Value)
	}
}

func TestPropertyDistributions_UnsetLast(t *testing.T) {
	apps := []views.AppEntry{
		{ID: "1", Properties: map[string]string{"status": "Active"}},
		{ID: "2", Properties: map[string]string{}},
	}
	result := views.PropertyDistributions(apps)
	buckets := result["status"]
	if len(buckets) < 2 {
		t.Fatalf("expected 2 buckets, got %d", len(buckets))
	}
	if buckets[len(buckets)-1].Value != "(unset)" {
		t.Errorf("expected (unset) last, got %q", buckets[len(buckets)-1].Value)
	}
}

func TestApplicationDashboard_EmptyWorkspace(t *testing.T) {
	conn := openTestDB(t)
	wsID := uuid.New() // non-existent workspace → empty results

	data, err := views.ApplicationDashboard(conn, wsID, "")
	if err != nil {
		t.Fatalf("ApplicationDashboard: %v", err)
	}
	if data.TotalApps != 0 {
		t.Errorf("TotalApps: got %d, want 0", data.TotalApps)
	}
	if len(data.Apps) != 0 {
		t.Errorf("Apps: got %d, want 0", len(data.Apps))
	}
	if len(data.Properties) != 0 {
		t.Errorf("Properties: got %d keys, want 0", len(data.Properties))
	}
}
