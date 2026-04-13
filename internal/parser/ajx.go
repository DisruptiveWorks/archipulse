package parser

import (
	"encoding/json"
	"fmt"
	"io"
)

// ParseAJX parses an ArchiMate JSON Exchange (AJX) document
// and returns a Model.
func ParseAJX(r io.Reader) (*Model, error) {
	var raw ajxModel
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, fmt.Errorf("ajx: decode json: %w", err)
	}
	return raw.toModel(), nil
}

// ---- raw AJX JSON structs ----

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
	Documentation string `json:"documentation"`
}

type ajxRelationship struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Source        string `json:"source"`
	Target        string `json:"target"`
	Name          string `json:"name"`
	Documentation string `json:"documentation"`
}

type ajxView struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Documentation string    `json:"documentation"`
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

func (m *ajxModel) toModel() *Model {
	out := &Model{Name: m.Name}

	for _, e := range m.Elements {
		out.Elements = append(out.Elements, Element{
			ID:            e.ID,
			Type:          e.Type,
			Name:          e.Name,
			Documentation: e.Documentation,
		})
	}

	for _, r := range m.Relationships {
		out.Relationships = append(out.Relationships, Relationship{
			ID:            r.ID,
			Type:          r.Type,
			Source:        r.Source,
			Target:        r.Target,
			Name:          r.Name,
			Documentation: r.Documentation,
		})
	}

	for _, v := range m.Views {
		d := Diagram{
			ID:            v.ID,
			Name:          v.Name,
			Documentation: v.Documentation,
		}
		for _, n := range v.Nodes {
			d.Layout.Nodes = append(d.Layout.Nodes, NodeLayout{
				ElementID: n.ElementRef,
				X:         n.X, Y: n.Y, W: n.W, H: n.H,
			})
		}
		for _, c := range v.Connections {
			cl := ConnectionLayout{RelationshipID: c.RelationshipRef}
			for _, bp := range c.Bendpoints {
				cl.Bendpoints = append(cl.Bendpoints, Point(bp))
			}
			d.Layout.Connections = append(d.Layout.Connections, cl)
		}
		out.Diagrams = append(out.Diagrams, d)
	}

	return out
}
