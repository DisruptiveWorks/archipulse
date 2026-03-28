// Package exporter serializes a workspace to AOEF (XML) and AJX (JSON).
package exporter

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/DisruptiveWorks/archipulse/internal/parser"
)

// WriteAOEF serializes m as ArchiMate Open Exchange Format XML into w.
func WriteAOEF(w io.Writer, m *parser.Model) error {
	raw := toAOEFModel(m)
	if _, err := fmt.Fprintf(w, "%s\n", xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(raw); err != nil {
		return fmt.Errorf("aoef: encode xml: %w", err)
	}
	return enc.Flush()
}

// ---- AOEF XML structs (output) ----

type aoefModel struct {
	XMLName       xml.Name           `xml:"model"`
	XMLNS         string             `xml:"xmlns,attr"`
	XMLNSxsi      string             `xml:"xmlns:xsi,attr"`
	SchemaLoc     string             `xml:"xsi:schemaLocation,attr"`
	Name          string             `xml:"name"`
	Elements      *aoefElements      `xml:"elements,omitempty"`
	Relationships *aoefRelationships `xml:"relationships,omitempty"`
	Views         *aoefViews         `xml:"views,omitempty"`
}

type aoefElements struct {
	Items []aoefElement `xml:"element"`
}

type aoefRelationships struct {
	Items []aoefRelationship `xml:"relationship"`
}

type aoefViews struct {
	Diagrams aoefDiagrams `xml:"diagrams"`
}

type aoefDiagrams struct {
	Items []aoefView `xml:"view"`
}

type aoefElement struct {
	ID            string `xml:"identifier,attr"`
	Type          string `xml:"xsi:type,attr"`
	Name          string `xml:"name"`
	Documentation string `xml:"documentation,omitempty"`
}

type aoefRelationship struct {
	ID            string `xml:"identifier,attr"`
	Type          string `xml:"xsi:type,attr"`
	Source        string `xml:"source,attr"`
	Target        string `xml:"target,attr"`
	Name          string `xml:"name,omitempty"`
	Documentation string `xml:"documentation,omitempty"`
}

type aoefView struct {
	ID            string     `xml:"identifier,attr"`
	Name          string     `xml:"name"`
	Documentation string     `xml:"documentation,omitempty"`
	Nodes         []aoefNode `xml:"node"`
	Connections   []aoefConn `xml:"connection"`
}

type aoefNode struct {
	ElementRef string `xml:"elementRef,attr"`
	X          int    `xml:"x,attr"`
	Y          int    `xml:"y,attr"`
	W          int    `xml:"w,attr"`
	H          int    `xml:"h,attr"`
}

type aoefConn struct {
	RelationshipRef string      `xml:"relationshipRef,attr"`
	Bendpoints      []aoefPoint `xml:"bendpoint"`
}

type aoefPoint struct {
	X int `xml:"x,attr"`
	Y int `xml:"y,attr"`
}

func toAOEFModel(m *parser.Model) *aoefModel {
	out := &aoefModel{
		XMLNS:     "http://www.opengroup.org/xsd/archimate/3.0/",
		XMLNSxsi:  "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLoc: "http://www.opengroup.org/xsd/archimate/3.0/ http://www.opengroup.org/xsd/archimate/3.1/archimate3_Diagram.xsd",
		Name:      m.Name,
	}

	if len(m.Elements) > 0 {
		elems := make([]aoefElement, len(m.Elements))
		for i, e := range m.Elements {
			elems[i] = aoefElement{
				ID:            e.ID,
				Type:          e.Type,
				Name:          e.Name,
				Documentation: e.Documentation,
			}
		}
		out.Elements = &aoefElements{Items: elems}
	}

	if len(m.Relationships) > 0 {
		rels := make([]aoefRelationship, len(m.Relationships))
		for i, r := range m.Relationships {
			rels[i] = aoefRelationship{
				ID:            r.ID,
				Type:          r.Type,
				Source:        r.Source,
				Target:        r.Target,
				Name:          r.Name,
				Documentation: r.Documentation,
			}
		}
		out.Relationships = &aoefRelationships{Items: rels}
	}

	if len(m.Diagrams) > 0 {
		views := make([]aoefView, len(m.Diagrams))
		for i, d := range m.Diagrams {
			v := aoefView{
				ID:            d.ID,
				Name:          d.Name,
				Documentation: d.Documentation,
			}
			for _, n := range d.Layout.Nodes {
				v.Nodes = append(v.Nodes, aoefNode{
					ElementRef: n.ElementID,
					X:          n.X, Y: n.Y, W: n.W, H: n.H,
				})
			}
			for _, c := range d.Layout.Connections {
				conn := aoefConn{RelationshipRef: c.RelationshipID}
				for _, bp := range c.Bendpoints {
					conn.Bendpoints = append(conn.Bendpoints, aoefPoint{X: bp.X, Y: bp.Y})
				}
				v.Connections = append(v.Connections, conn)
			}
			views[i] = v
		}
		out.Views = &aoefViews{Diagrams: aoefDiagrams{Items: views}}
	}

	return out
}
