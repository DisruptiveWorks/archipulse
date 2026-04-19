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

func TestWriteAOEF_RoundTrip_ArchiMetal(t *testing.T) {
	f, err := os.Open("../../examples/ArchiMetal.xml")
	if err != nil {
		t.Skip("ArchiMetal.xml not found:", err)
	}
	defer func() { _ = f.Close() }()
	original, err := parser.ParseAOEF(f)
	if err != nil {
		t.Fatalf("parse ArchiMetal: %v", err)
	}

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
			t.Errorf("diagram[%d] %q nodes: got %d, want %d", i, da.Name, len(db.Layout.Nodes), len(da.Layout.Nodes))
		}
		if len(da.Layout.Connections) != len(db.Layout.Connections) {
			t.Errorf("diagram[%d] %q connections: got %d, want %d", i, da.Name, len(db.Layout.Connections), len(da.Layout.Connections))
		}
		// Compare nodes by a composite key of NodeID+ElementID.
		// Using both fields handles nodes that lack an identifier (NodeID==""),
		// where ElementID alone is the only stable key.
		nodeKey := func(n parser.NodeLayout) string { return n.NodeID + ":" + n.ElementID }
		byNodeID := make(map[string]parser.NodeLayout, len(db.Layout.Nodes))
		for _, n := range db.Layout.Nodes {
			byNodeID[nodeKey(n)] = n
		}
		for _, na := range da.Layout.Nodes {
			nb, ok := byNodeID[nodeKey(na)]
			if !ok {
				t.Errorf("diagram[%d] %q: node %q (elem %q) missing after round-trip", i, da.Name, na.NodeID, na.ElementID)
				continue
			}
			if na.ElementID != nb.ElementID {
				t.Errorf("diagram[%d] %q node %q ElementID: got %q, want %q", i, da.Name, na.NodeID, nb.ElementID, na.ElementID)
			}
			if na.ParentElementID != nb.ParentElementID {
				t.Errorf("diagram[%d] %q node %q ParentElementID: got %q, want %q", i, da.Name, na.NodeID, nb.ParentElementID, na.ParentElementID)
			}
			if na.X != nb.X || na.Y != nb.Y || na.W != nb.W || na.H != nb.H {
				t.Errorf("diagram[%d] %q node %q bounds: got (%d,%d,%d,%d), want (%d,%d,%d,%d)",
					i, da.Name, na.NodeID, nb.X, nb.Y, nb.W, nb.H, na.X, na.Y, na.W, na.H)
			}
		}
	}
}
