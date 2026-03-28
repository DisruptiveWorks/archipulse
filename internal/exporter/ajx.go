package exporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DisruptiveWorks/archipulse/internal/parser"
)

// WriteAJX serializes m as ArchiMate JSON Exchange into w.
func WriteAJX(w io.Writer, m *parser.Model) error {
	raw := toAJXModel(m)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(raw); err != nil {
		return fmt.Errorf("ajx: encode json: %w", err)
	}
	return nil
}

// ---- AJX JSON structs (output) ----

type ajxModel struct {
	Name          string            `json:"name"`
	Elements      []ajxElement      `json:"elements"`
	Relationships []ajxRelationship `json:"relationships"`
	Views         []ajxView         `json:"views"`
}

type ajxElement struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	Documentation string `json:"documentation,omitempty"`
}

type ajxRelationship struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Source        string `json:"source"`
	Target        string `json:"target"`
	Name          string `json:"name,omitempty"`
	Documentation string `json:"documentation,omitempty"`
}

type ajxView struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Documentation string    `json:"documentation,omitempty"`
	Nodes         []ajxNode `json:"nodes"`
	Connections   []ajxConn `json:"connections"`
}

type ajxNode struct {
	ElementRef string `json:"elementRef"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	W          int    `json:"w"`
	H          int    `json:"h"`
}

type ajxConn struct {
	RelationshipRef string     `json:"relationshipRef"`
	Bendpoints      []ajxPoint `json:"bendpoints"`
}

type ajxPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func toAJXModel(m *parser.Model) *ajxModel {
	out := &ajxModel{
		Name:          m.Name,
		Elements:      make([]ajxElement, len(m.Elements)),
		Relationships: make([]ajxRelationship, len(m.Relationships)),
		Views:         make([]ajxView, len(m.Diagrams)),
	}

	for i, e := range m.Elements {
		out.Elements[i] = ajxElement{
			ID:            e.ID,
			Type:          e.Type,
			Name:          e.Name,
			Documentation: e.Documentation,
		}
	}

	for i, r := range m.Relationships {
		out.Relationships[i] = ajxRelationship{
			ID:            r.ID,
			Type:          r.Type,
			Source:        r.Source,
			Target:        r.Target,
			Name:          r.Name,
			Documentation: r.Documentation,
		}
	}

	for i, d := range m.Diagrams {
		v := ajxView{
			ID:            d.ID,
			Name:          d.Name,
			Documentation: d.Documentation,
			Nodes:         make([]ajxNode, len(d.Layout.Nodes)),
			Connections:   make([]ajxConn, len(d.Layout.Connections)),
		}
		for j, n := range d.Layout.Nodes {
			v.Nodes[j] = ajxNode{ElementRef: n.ElementID, X: n.X, Y: n.Y, W: n.W, H: n.H}
		}
		for j, c := range d.Layout.Connections {
			conn := ajxConn{
				RelationshipRef: c.RelationshipID,
				Bendpoints:      make([]ajxPoint, len(c.Bendpoints)),
			}
			for k, bp := range c.Bendpoints {
				conn.Bendpoints[k] = ajxPoint{X: bp.X, Y: bp.Y}
			}
			v.Connections[j] = conn
		}
		out.Views[i] = v
	}

	return out
}
