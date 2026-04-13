package parser_test

import (
	"os"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/parser"
)

// TestParseAOEF_ArchiMetal validates full OEF coverage using the local ArchiMetal example.
// It exercises: multi-language names, xsi:type on elements, node styles, viewpoints,
// property definitions with data types, and relationship semantic attributes.
func TestParseAOEF_ArchiMetal(t *testing.T) {
	f, err := os.Open("../../examples/ArchiMetal.xml")
	if err != nil {
		t.Skipf("ArchiMetal.xml not found — skipping: %v", err)
	}
	defer func() { _ = f.Close() }()

	m, err := parser.ParseAOEF(f)
	if err != nil {
		t.Fatalf("ParseAOEF: %v", err)
	}

	if m.Name == "" {
		t.Error("model name is empty")
	}
	t.Logf("Model: %q  version=%q", m.Name, m.Version)

	if len(m.Elements) == 0 {
		t.Error("expected elements, got 0")
	}
	if len(m.Diagrams) == 0 {
		t.Error("expected diagrams, got 0")
	}

	t.Logf("Elements: %d  Relationships: %d  Diagrams: %d",
		len(m.Elements), len(m.Relationships), len(m.Diagrams))
	t.Logf("ViewFolders: %d  DiagramFolders: %d  PropertyDefinitions: %d",
		len(m.ViewFolders), len(m.DiagramFolders), len(m.PropertyDefinitions))

	// All elements must have a non-empty name (multi-lang parsing must work).
	for _, e := range m.Elements {
		if e.Name == "" {
			t.Errorf("element %q has empty name (multi-lang parsing failure)", e.ID)
		}
	}

	// Property definitions must carry their data type.
	for _, pd := range m.PropertyDefinitions {
		if pd.DataType == "" {
			t.Errorf("property definition %q has empty data_type", pd.Name)
		}
		t.Logf("  PropDef: %-30q  type=%q", pd.Name, pd.DataType)
	}

	// Count nodes with style information.
	styledNodes := 0
	for _, d := range m.Diagrams {
		for _, n := range d.Layout.Nodes {
			if n.Style != nil {
				styledNodes++
			}
		}
	}
	t.Logf("Nodes with style: %d", styledNodes)

	// Count connections with style.
	styledConns := 0
	for _, d := range m.Diagrams {
		for _, c := range d.Layout.Connections {
			if c.Style != nil {
				styledConns++
			}
		}
	}
	t.Logf("Connections with style: %d", styledConns)

	// Count diagrams with a viewpoint attribute.
	viewpointed := 0
	for _, d := range m.Diagrams {
		if d.Viewpoint != "" {
			viewpointed++
		}
	}
	t.Logf("Diagrams with viewpoint: %d", viewpointed)

	// Relationship semantic attributes.
	accessCount, directedCount, influenceCount := 0, 0, 0
	for _, r := range m.Relationships {
		if r.AccessType != "" {
			accessCount++
		}
		if r.IsDirected {
			directedCount++
		}
		if r.Modifier != "" {
			influenceCount++
		}
	}
	t.Logf("Access rels with accessType: %d  Directed assoc: %d  Influence with modifier: %d",
		accessCount, directedCount, influenceCount)
}
