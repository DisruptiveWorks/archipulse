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

// aoefNode mirrors the parser's aoefNode, supporting recursive nesting.
type aoefNode struct {
	Identifier string     `xml:"identifier,attr"`
	ElementRef string     `xml:"elementRef,attr,omitempty"`
	X          int        `xml:"x,attr"`
	Y          int        `xml:"y,attr"`
	W          int        `xml:"w,attr"`
	H          int        `xml:"h,attr"`
	Children   []aoefNode `xml:"node"`
}

type aoefConn struct {
	RelationshipRef string      `xml:"relationshipRef,attr"`
	Source          string      `xml:"source,attr"`
	Target          string      `xml:"target,attr"`
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
			v.Nodes = buildNodeTree(d.Layout.Nodes)
			for _, c := range d.Layout.Connections {
				conn := aoefConn{
					RelationshipRef: c.RelationshipID,
					Source:          c.SourceNodeID,
					Target:          c.TargetNodeID,
				}
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

// buildNodeTree reconstructs the AOEF hierarchical node structure from the
// flat NodeLayout slice produced by the parser's collectNodes.
//
// The flat list encodes parent-child relationships via ParentElementID, which
// holds either the parent element's ElementID (for element nodes) or the
// parent group's NodeID (for group nodes).
//
// When the same element is referenced by two or more nodes in the same diagram
// (valid in ArchiMate), their children all share the same ParentElementID and
// would otherwise be duplicated under every instance.  Spatial containment is
// used to assign each child to the unique parent whose bounds enclose it.
func buildNodeTree(nodes []parser.NodeLayout) []aoefNode {
	// Build a map from "parent key" → child indices.
	// The parent key is ElementID for element nodes, NodeID for group nodes.
	childrenOf := make(map[string][]int, len(nodes))
	for i, n := range nodes {
		childrenOf[n.ParentElementID] = append(childrenOf[n.ParentElementID], i)
	}

	// Count how many nodes share each ElementID so we know when disambiguation
	// via spatial containment is necessary.
	elemIDCount := make(map[string]int, len(nodes))
	for _, n := range nodes {
		if n.ElementID != "" {
			elemIDCount[n.ElementID]++
		}
	}

	// Recursively build an aoefNode and all its descendants.
	var buildNode func(i int) aoefNode
	buildNode = func(i int) aoefNode {
		n := nodes[i]
		node := aoefNode{
			Identifier: n.NodeID,
			ElementRef: n.ElementID,
			X:          n.X,
			Y:          n.Y,
			W:          n.W,
			H:          n.H,
		}
		// Children reference this node via its ElementID (element node) or
		// NodeID (group node — ElementID is empty).
		parentKey := n.ElementID
		if parentKey == "" {
			parentKey = n.NodeID
		}
		for _, ci := range childrenOf[parentKey] {
			// When multiple nodes share the same ElementID, use spatial
			// containment to avoid assigning a child to the wrong parent.
			if elemIDCount[parentKey] > 1 && !nodeContains(n, nodes[ci]) {
				continue
			}
			node.Children = append(node.Children, buildNode(ci))
		}
		return node
	}

	// Roots are nodes whose ParentElementID is empty.
	var roots []aoefNode
	for i := range nodes {
		if nodes[i].ParentElementID == "" {
			roots = append(roots, buildNode(i))
		}
	}
	return roots
}

// nodeContains reports whether parent's bounds fully enclose child's bounds.
func nodeContains(parent, child parser.NodeLayout) bool {
	return child.X >= parent.X && child.Y >= parent.Y &&
		child.X+child.W <= parent.X+parent.W &&
		child.Y+child.H <= parent.Y+parent.H
}
