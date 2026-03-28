package parser_test

import (
	"os"
	"strings"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/parser"
)

func TestParseAJX(t *testing.T) {
	f, err := os.Open("../../examples/minimal.ajx")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	m, err := parser.ParseAJX(f)
	if err != nil {
		t.Fatalf("ParseAJX: %v", err)
	}

	if m.Name != "Minimal ArchiPulse Test Model" {
		t.Errorf("name = %q, want %q", m.Name, "Minimal ArchiPulse Test Model")
	}

	counts := []struct {
		name string
		got  int
		want int
	}{
		{"elements", len(m.Elements), 3},
		{"relationships", len(m.Relationships), 2},
		{"diagrams", len(m.Diagrams), 1},
	}
	for _, tt := range counts {
		if tt.got != tt.want {
			t.Errorf("%s count = %d, want %d", tt.name, tt.got, tt.want)
		}
	}

	elem := m.Elements[0]
	if elem.ID != "id-app-001" {
		t.Errorf("element[0].ID = %q, want %q", elem.ID, "id-app-001")
	}
	if elem.Documentation != "Handles payment processing" {
		t.Errorf("element[0].Documentation = %q", elem.Documentation)
	}

	diag := m.Diagrams[0]
	if len(diag.Layout.Nodes) != 2 {
		t.Errorf("diagram nodes = %d, want 2", len(diag.Layout.Nodes))
	}
	conn := diag.Layout.Connections[0]
	if len(conn.Bendpoints) != 1 || conn.Bendpoints[0].X != 220 {
		t.Errorf("unexpected bendpoint: %+v", conn.Bendpoints)
	}
}

func TestParseAJX_InvalidJSON(t *testing.T) {
	_, err := parser.ParseAJX(strings.NewReader("{bad json"))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestAOEFandAJXProduceSameModel(t *testing.T) {
	xmlFile, err := os.Open("../../examples/minimal.xml")
	if err != nil {
		t.Fatalf("open xml fixture: %v", err)
	}
	defer func() { _ = xmlFile.Close() }()

	jsonFile, err := os.Open("../../examples/minimal.ajx")
	if err != nil {
		t.Fatalf("open ajx fixture: %v", err)
	}
	defer func() { _ = jsonFile.Close() }()

	aoef, err := parser.ParseAOEF(xmlFile)
	if err != nil {
		t.Fatalf("ParseAOEF: %v", err)
	}
	ajx, err := parser.ParseAJX(jsonFile)
	if err != nil {
		t.Fatalf("ParseAJX: %v", err)
	}

	if len(aoef.Elements) != len(ajx.Elements) {
		t.Errorf("element count mismatch: aoef=%d ajx=%d", len(aoef.Elements), len(ajx.Elements))
	}
	if len(aoef.Relationships) != len(ajx.Relationships) {
		t.Errorf("relationship count mismatch: aoef=%d ajx=%d", len(aoef.Relationships), len(ajx.Relationships))
	}
	if len(aoef.Diagrams) != len(ajx.Diagrams) {
		t.Errorf("diagram count mismatch: aoef=%d ajx=%d", len(aoef.Diagrams), len(ajx.Diagrams))
	}

	for i := range aoef.Elements {
		a, b := aoef.Elements[i], ajx.Elements[i]
		if a.ID != b.ID || a.Type != b.Type || a.Name != b.Name {
			t.Errorf("element[%d] mismatch: aoef=%+v ajx=%+v", i, a, b)
		}
	}
}

// badReader wraps a string as an io.Reader (used by AOEF invalid XML test).
func badReader(s string) *strings.Reader {
	return strings.NewReader(s)
}
