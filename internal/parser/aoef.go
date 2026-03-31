package parser

import (
	"encoding/xml"
	"fmt"
	"io"
)

// ParseAOEF parses an ArchiMate Open Exchange Format (XML) document
// and returns a Model.
func ParseAOEF(r io.Reader) (*Model, error) {
	var raw aoefModel
	if err := xml.NewDecoder(r).Decode(&raw); err != nil {
		return nil, fmt.Errorf("aoef: decode xml: %w", err)
	}
	return raw.toModel(), nil
}

// ---- raw AOEF XML structs ----

type aoefModel struct {
	XMLName       xml.Name           `xml:"model"`
	Name          string             `xml:"name"`
	PropertyDefs  []aoefPropertyDef  `xml:"propertyDefinitions>propertyDefinition"`
	Elements      []aoefElement      `xml:"elements>element"`
	Relationships []aoefRelationship `xml:"relationships>relationship"`
	Views         []aoefView         `xml:"views>diagrams>view"`
}

type aoefPropertyDef struct {
	ID   string `xml:"identifier,attr"`
	Name string `xml:"name"`
}

type aoefElement struct {
	ID            string         `xml:"identifier,attr"`
	Type          string         `xml:"type,attr"`
	Name          string         `xml:"name"`
	Documentation string         `xml:"documentation"`
	Properties    []aoefProperty `xml:"properties>property"`
}

type aoefProperty struct {
	DefinitionRef string `xml:"propertyDefinitionRef,attr"`
	Value         string `xml:"value"`
}

type aoefRelationship struct {
	ID            string `xml:"identifier,attr"`
	Type          string `xml:"type,attr"`
	Source        string `xml:"source,attr"`
	Target        string `xml:"target,attr"`
	Name          string `xml:"name"`
	Documentation string `xml:"documentation"`
}

type aoefView struct {
	ID            string     `xml:"identifier,attr"`
	Name          string     `xml:"name"`
	Documentation string     `xml:"documentation"`
	Nodes         []aoefNode `xml:"node"`
	Connections   []aoefConn `xml:"connection"`
}

type aoefNode struct {
	ElementRef string     `xml:"elementRef,attr"`
	X          int        `xml:"x,attr"`
	Y          int        `xml:"y,attr"`
	W          int        `xml:"w,attr"`
	H          int        `xml:"h,attr"`
	Children   []aoefNode `xml:"node"` // nested nodes (Container, Group...)
}

type aoefConn struct {
	RelationshipRef string      `xml:"relationshipRef,attr"`
	Bendpoints      []aoefPoint `xml:"bendpoint"`
}

type aoefPoint struct {
	X int `xml:"x,attr"`
	Y int `xml:"y,attr"`
}

func (m *aoefModel) toModel() *Model {
	out := &Model{Name: m.Name}

	// Build property definition ID → name lookup.
	propNames := make(map[string]string, len(m.PropertyDefs))
	for _, pd := range m.PropertyDefs {
		propNames[pd.ID] = pd.Name
	}

	for _, e := range m.Elements {
		elem := Element{
			ID:            e.ID,
			Type:          e.Type,
			Name:          e.Name,
			Documentation: e.Documentation,
		}
		for _, p := range e.Properties {
			key := propNames[p.DefinitionRef]
			if key == "" {
				key = p.DefinitionRef
			}
			if key != "" && p.Value != "" {
				elem.Properties = append(elem.Properties, Property{Key: key, Value: p.Value})
			}
		}
		out.Elements = append(out.Elements, elem)
	}

	for _, r := range m.Relationships {
		out.Relationships = append(out.Relationships, Relationship(r))
	}

	for _, v := range m.Views {
		d := Diagram{
			ID:            v.ID,
			Name:          v.Name,
			Documentation: v.Documentation,
		}
		collectNodes(v.Nodes, &d.Layout.Nodes)
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

// collectNodes recursively traverses nested AOEF nodes (Containers, Groups)
// and collects only nodes that reference an element (elementRef != "").
func collectNodes(nodes []aoefNode, out *[]NodeLayout) {
	for _, n := range nodes {
		if n.ElementRef != "" {
			*out = append(*out, NodeLayout{
				ElementID: n.ElementRef,
				X:         n.X, Y: n.Y, W: n.W, H: n.H,
			})
		}
		if len(n.Children) > 0 {
			collectNodes(n.Children, out)
		}
	}
}
