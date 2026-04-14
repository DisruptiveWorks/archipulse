package parser

import "fmt"

// ValidationError holds all semantic issues found in a model.
type ValidationError struct {
	Issues []string
}

func (e *ValidationError) Error() string {
	msg := fmt.Sprintf("model validation failed with %d issue(s):", len(e.Issues))
	for _, issue := range e.Issues {
		msg += "\n  - " + issue
	}
	return msg
}

// Validate performs semantic validation on a parsed Model.
// Returns a *ValidationError listing all issues, or nil if the model is valid.
func Validate(m *Model) error {
	var issues []string

	elementIDs := make(map[string]struct{}, len(m.Elements))

	for i, e := range m.Elements {
		if e.ID == "" {
			issues = append(issues, fmt.Sprintf("element[%d]: missing identifier", i))
		}
		if e.Type == "" {
			issues = append(issues, fmt.Sprintf("element %q: missing type", e.ID))
		} else if _, ok := ValidElementTypes[e.Type]; !ok {
			issues = append(issues, fmt.Sprintf("element %q: unknown ArchiMate type %q", e.ID, e.Type))
		}
		if e.ID != "" {
			if _, dup := elementIDs[e.ID]; dup {
				issues = append(issues, fmt.Sprintf("element %q: duplicate identifier", e.ID))
			}
			elementIDs[e.ID] = struct{}{}
		}
	}

	relIDs := make(map[string]struct{}, len(m.Relationships))

	for i, r := range m.Relationships {
		if r.ID == "" {
			issues = append(issues, fmt.Sprintf("relationship[%d]: missing identifier", i))
		}
		if r.Type == "" {
			issues = append(issues, fmt.Sprintf("relationship %q: missing type", r.ID))
		} else if _, ok := ValidRelationshipTypes[r.Type]; !ok {
			issues = append(issues, fmt.Sprintf("relationship %q: unknown ArchiMate relationship type %q", r.ID, r.Type))
		}
		if r.Source == "" {
			issues = append(issues, fmt.Sprintf("relationship %q: missing source", r.ID))
		} else if _, ok := elementIDs[r.Source]; !ok {
			issues = append(issues, fmt.Sprintf("relationship %q: source %q references unknown element", r.ID, r.Source))
		}
		if r.Target == "" {
			issues = append(issues, fmt.Sprintf("relationship %q: missing target", r.ID))
		} else if _, ok := elementIDs[r.Target]; !ok {
			issues = append(issues, fmt.Sprintf("relationship %q: target %q references unknown element", r.ID, r.Target))
		}
		if r.ID != "" {
			if _, dup := relIDs[r.ID]; dup {
				issues = append(issues, fmt.Sprintf("relationship %q: duplicate identifier", r.ID))
			}
			relIDs[r.ID] = struct{}{}
		}
	}

	for i, d := range m.Diagrams {
		if d.ID == "" {
			issues = append(issues, fmt.Sprintf("diagram[%d]: missing identifier", i))
		}
		for j, n := range d.Layout.Nodes {
			// Group nodes use a diagram-local identifier, not an element reference — skip validation.
			if n.ElementType == "Group" {
				continue
			}
			if _, ok := elementIDs[n.ElementID]; !ok {
				issues = append(issues, fmt.Sprintf("diagram %q node[%d]: references unknown element %q", d.ID, j, n.ElementID))
			}
		}
		for j, c := range d.Layout.Connections {
			if _, ok := relIDs[c.RelationshipID]; !ok {
				issues = append(issues, fmt.Sprintf("diagram %q connection[%d]: references unknown relationship %q", d.ID, j, c.RelationshipID))
			}
		}
	}

	if len(issues) > 0 {
		return &ValidationError{Issues: issues}
	}
	return nil
}
