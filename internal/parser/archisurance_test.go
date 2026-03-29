package parser_test

import (
	"os"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/parser"
)

func TestParseAOEF_ArchiSurance(t *testing.T) {
	f, err := os.Open("../../examples/archisurance.xml")
	if err != nil {
		t.Fatalf("open archisurance fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	m, err := parser.ParseAOEF(f)
	if err != nil {
		t.Fatalf("ParseAOEF: %v", err)
	}

	if m.Name == "" {
		t.Error("model name is empty")
	}

	if len(m.Elements) < 100 {
		t.Errorf("expected at least 100 elements, got %d", len(m.Elements))
	}
	if len(m.Relationships) < 100 {
		t.Errorf("expected at least 100 relationships, got %d", len(m.Relationships))
	}
	if len(m.Diagrams) == 0 {
		t.Error("expected at least one diagram")
	}

	t.Logf("ArchiSurance: %d elements, %d relationships, %d diagrams",
		len(m.Elements), len(m.Relationships), len(m.Diagrams))
}

func TestValidate_ArchiSurance(t *testing.T) {
	f, err := os.Open("../../examples/archisurance.xml")
	if err != nil {
		t.Fatalf("open archisurance fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	m, err := parser.ParseAOEF(f)
	if err != nil {
		t.Fatalf("ParseAOEF: %v", err)
	}

	if err := parser.Validate(m); err != nil {
		t.Errorf("ArchiSurance failed validation: %v", err)
	}
}

func TestParseAOEF_ArchiSurance_ElementTypes(t *testing.T) {
	f, err := os.Open("../../examples/archisurance.xml")
	if err != nil {
		t.Fatalf("open archisurance fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	m, err := parser.ParseAOEF(f)
	if err != nil {
		t.Fatalf("ParseAOEF: %v", err)
	}

	// Count elements per layer using ElementLayer.
	layerCount := make(map[string]int)
	for _, e := range m.Elements {
		layer := parser.ElementLayer(e.Type)
		if layer == "" {
			t.Errorf("element %q has unknown type %q", e.ID, e.Type)
		}
		layerCount[layer]++
	}
	t.Logf("Elements by layer: %v", layerCount)

	// ArchiSurance spans multiple layers — expect at least Business and Application.
	if layerCount["Business"] == 0 {
		t.Error("expected Business layer elements in ArchiSurance")
	}
	if layerCount["Application"] == 0 {
		t.Error("expected Application layer elements in ArchiSurance")
	}
}
