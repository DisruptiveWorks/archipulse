package exporter_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/exporter"
	"github.com/DisruptiveWorks/archipulse/internal/parser"
)

func fixtureModel(t *testing.T) *parser.Model {
	t.Helper()
	f, err := os.Open("../../examples/minimal.xml")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()
	m, err := parser.ParseAOEF(f)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	return m
}

func TestWriteAOEF_RoundTrip(t *testing.T) {
	original := fixtureModel(t)

	var buf bytes.Buffer
	if err := exporter.WriteAOEF(&buf, original); err != nil {
		t.Fatalf("WriteAOEF: %v", err)
	}

	reparsed, err := parser.ParseAOEF(&buf)
	if err != nil {
		t.Fatalf("re-parse exported AOEF: %v", err)
	}

	assertModelsEqual(t, original, reparsed)
}

func TestWriteAJX_RoundTrip(t *testing.T) {
	original := fixtureModel(t)

	var buf bytes.Buffer
	if err := exporter.WriteAJX(&buf, original); err != nil {
		t.Fatalf("WriteAJX: %v", err)
	}

	reparsed, err := parser.ParseAJX(&buf)
	if err != nil {
		t.Fatalf("re-parse exported AJX: %v", err)
	}

	assertModelsEqual(t, original, reparsed)
}

func assertModelsEqual(t *testing.T, a, b *parser.Model) {
	t.Helper()

	if len(a.Elements) != len(b.Elements) {
		t.Errorf("elements: got %d, want %d", len(b.Elements), len(a.Elements))
	}
	if len(a.Relationships) != len(b.Relationships) {
		t.Errorf("relationships: got %d, want %d", len(b.Relationships), len(a.Relationships))
	}
	if len(a.Diagrams) != len(b.Diagrams) {
		t.Errorf("diagrams: got %d, want %d", len(b.Diagrams), len(a.Diagrams))
	}

	for i := range a.Elements {
		ea, eb := a.Elements[i], b.Elements[i]
		if ea.ID != eb.ID || ea.Type != eb.Type || ea.Name != eb.Name {
			t.Errorf("element[%d] mismatch: %+v vs %+v", i, ea, eb)
		}
	}

	for i := range a.Relationships {
		ra, rb := a.Relationships[i], b.Relationships[i]
		if ra.ID != rb.ID || ra.Type != rb.Type || ra.Source != rb.Source || ra.Target != rb.Target {
			t.Errorf("relationship[%d] mismatch: %+v vs %+v", i, ra, rb)
		}
	}

	for i := range a.Diagrams {
		da, db := a.Diagrams[i], b.Diagrams[i]
		if da.ID != db.ID || da.Name != db.Name {
			t.Errorf("diagram[%d] mismatch: %+v vs %+v", i, da, db)
		}
		if len(da.Layout.Nodes) != len(db.Layout.Nodes) {
			t.Errorf("diagram[%d] nodes: got %d, want %d", i, len(db.Layout.Nodes), len(da.Layout.Nodes))
		}
	}
}
