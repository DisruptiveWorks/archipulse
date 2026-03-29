package parser_test

import (
	"errors"
	"os"
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/parser"
)

func TestValidate_ValidModel(t *testing.T) {
	f, err := os.Open("../../examples/minimal.xml")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer func() { _ = f.Close() }()

	m, err := parser.ParseAOEF(f)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := parser.Validate(m); err != nil {
		t.Errorf("expected valid model, got: %v", err)
	}
}

func TestValidate_UnknownElementType(t *testing.T) {
	m := &parser.Model{
		Elements: []parser.Element{
			{ID: "e1", Type: "NotARealType", Name: "X"},
		},
	}
	err := parser.Validate(m)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	var ve *parser.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Issues) != 1 {
		t.Errorf("expected 1 issue, got %d: %v", len(ve.Issues), ve.Issues)
	}
}

func TestValidate_MissingElementID(t *testing.T) {
	m := &parser.Model{
		Elements: []parser.Element{
			{ID: "", Type: "ApplicationComponent", Name: "X"},
		},
	}
	err := parser.Validate(m)
	if err == nil {
		t.Fatal("expected validation error for missing ID")
	}
}

func TestValidate_RelationshipReferencesUnknownElement(t *testing.T) {
	m := &parser.Model{
		Elements: []parser.Element{
			{ID: "e1", Type: "ApplicationComponent", Name: "A"},
		},
		Relationships: []parser.Relationship{
			{ID: "r1", Type: "AssociationRelationship", Source: "e1", Target: "e-missing"},
		},
	}
	err := parser.Validate(m)
	if err == nil {
		t.Fatal("expected validation error for unknown target")
	}
	var ve *parser.ValidationError
	errors.As(err, &ve)
	if len(ve.Issues) != 1 {
		t.Errorf("expected 1 issue, got %d: %v", len(ve.Issues), ve.Issues)
	}
}

func TestValidate_DuplicateElementID(t *testing.T) {
	m := &parser.Model{
		Elements: []parser.Element{
			{ID: "e1", Type: "ApplicationComponent", Name: "A"},
			{ID: "e1", Type: "BusinessProcess", Name: "B"},
		},
	}
	err := parser.Validate(m)
	if err == nil {
		t.Fatal("expected validation error for duplicate ID")
	}
}

func TestValidate_DiagramReferencesUnknownElement(t *testing.T) {
	m := &parser.Model{
		Elements: []parser.Element{
			{ID: "e1", Type: "ApplicationComponent", Name: "A"},
		},
		Diagrams: []parser.Diagram{
			{
				ID:   "d1",
				Name: "Test View",
				Layout: parser.DiagramLayout{
					Nodes: []parser.NodeLayout{
						{ElementID: "e-missing", X: 0, Y: 0, W: 100, H: 50},
					},
				},
			},
		},
	}
	err := parser.Validate(m)
	if err == nil {
		t.Fatal("expected validation error for unknown node element reference")
	}
}
