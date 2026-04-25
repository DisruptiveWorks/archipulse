package parser_test

import (
	"os"
	"strings"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/parser"
)

func badReader(s string) *strings.Reader { return strings.NewReader(s) }

func TestParseAOEF(t *testing.T) {
	f, err := os.Open("../testdata/minimal.xml")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	m, err := parser.ParseAOEF(f)
	if err != nil {
		t.Fatalf("ParseAOEF: %v", err)
	}

	if m.Name != "Minimal ArchiPulse Test Model" {
		t.Errorf("name = %q, want %q", m.Name, "Minimal ArchiPulse Test Model")
	}

	tests := []struct {
		name string
		got  int
		want int
	}{
		{"elements", len(m.Elements), 3},
		{"relationships", len(m.Relationships), 2},
		{"diagrams", len(m.Diagrams), 1},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s count = %d, want %d", tt.name, tt.got, tt.want)
		}
	}

	elem := m.Elements[0]
	if elem.ID != "id-app-001" {
		t.Errorf("element[0].ID = %q, want %q", elem.ID, "id-app-001")
	}
	if elem.Type != "ApplicationComponent" {
		t.Errorf("element[0].Type = %q, want %q", elem.Type, "ApplicationComponent")
	}
	if elem.Name != "PaymentService" {
		t.Errorf("element[0].Name = %q, want %q", elem.Name, "PaymentService")
	}

	rel := m.Relationships[0]
	if rel.Source != "id-app-001" || rel.Target != "id-app-002" {
		t.Errorf("relationship[0] source/target = %q/%q, want id-app-001/id-app-002", rel.Source, rel.Target)
	}

	diag := m.Diagrams[0]
	if diag.ID != "id-view-001" {
		t.Errorf("diagram[0].ID = %q, want %q", diag.ID, "id-view-001")
	}
	if len(diag.Layout.Nodes) != 2 {
		t.Errorf("diagram nodes = %d, want 2", len(diag.Layout.Nodes))
	}
	if len(diag.Layout.Connections) != 1 {
		t.Errorf("diagram connections = %d, want 1", len(diag.Layout.Connections))
	}
	if len(diag.Layout.Connections[0].Bendpoints) != 1 {
		t.Errorf("connection bendpoints = %d, want 1", len(diag.Layout.Connections[0].Bendpoints))
	}
}

func TestParseAOEF_InvalidXML(t *testing.T) {
	_, err := parser.ParseAOEF(badReader("not xml at all"))
	if err == nil {
		t.Error("expected error for invalid XML, got nil")
	}
}
