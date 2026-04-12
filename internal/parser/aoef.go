package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
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
	Organizations []aoefOrgItem      `xml:"organizations>item"`
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

// aoefOrgItem represents a node in <organizations>. Items with identifierRef
// are leaves (elements or views); items with <label> and children are folders.
type aoefOrgItem struct {
	IdentifierRef string           `xml:"identifierRef,attr"`
	Labels        []aoefLangString `xml:"label"`
	Children      []aoefOrgItem    `xml:"item"`
}

type aoefLangString struct {
	Value string `xml:",chardata"`
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
		collectNodes(v.Nodes, "", &d.Layout.Nodes)
		for _, c := range v.Connections {
			cl := ConnectionLayout{RelationshipID: c.RelationshipRef}
			for _, bp := range c.Bendpoints {
				cl.Bendpoints = append(cl.Bendpoints, Point(bp))
			}
			d.Layout.Connections = append(d.Layout.Connections, cl)
		}
		out.Diagrams = append(out.Diagrams, d)
	}

	// Build view ID set for organization traversal.
	viewIDs := make(map[string]bool, len(m.Views))
	for _, v := range m.Views {
		viewIDs[v.ID] = true
	}

	// Extract view folder hierarchy from <organizations>.
	for i, org := range m.Organizations {
		folders, diagFolders, hasViews := collectViewOrg(org, "", viewIDs, i)
		if hasViews {
			out.ViewFolders = append(out.ViewFolders, folders...)
			out.DiagramFolders = append(out.DiagramFolders, diagFolders...)
		}
	}

	return out
}

// collectViewOrg recursively traverses an organization item.
// Returns (folders, diagramFolderAssignments, containsAnyViewRef).
// Folders are returned in pre-order (parent before children) for safe DB insertion.
func collectViewOrg(item aoefOrgItem, parentSourceID string, viewIDs map[string]bool, pos int) ([]ViewFolder, []DiagramFolder, bool) {
	// Leaf: has identifierRef.
	if item.IdentifierRef != "" {
		if viewIDs[item.IdentifierRef] {
			return nil, []DiagramFolder{{
				DiagramSourceID: item.IdentifierRef,
				FolderSourceID:  parentSourceID,
			}}, true
		}
		return nil, nil, false // element ref — skip
	}

	// Folder item: compute this folder's source ID.
	label := orgItemLabel(item)
	sourceID := label
	if parentSourceID != "" && label != "" {
		sourceID = parentSourceID + "/" + label
	} else if parentSourceID != "" {
		sourceID = parentSourceID
	}

	var childFolders []ViewFolder
	var diagFolders []DiagramFolder
	containsViews := false

	for i, child := range item.Children {
		childParent := sourceID
		if label == "" {
			childParent = parentSourceID
		}
		cf, cd, hasViews := collectViewOrg(child, childParent, viewIDs, i)
		if hasViews {
			childFolders = append(childFolders, cf...)
			diagFolders = append(diagFolders, cd...)
			containsViews = true
		}
	}

	if !containsViews || label == "" {
		// No views here, or anonymous grouping — pass through without creating a folder.
		return childFolders, diagFolders, containsViews
	}

	// Prepend this folder so it appears before its children (parent-first order).
	folder := ViewFolder{
		SourceID: sourceID,
		Name:     label,
		ParentID: parentSourceID,
		Position: pos,
	}
	return append([]ViewFolder{folder}, childFolders...), diagFolders, true
}

// orgItemLabel returns the first non-empty label from an org item.
func orgItemLabel(item aoefOrgItem) string {
	for _, l := range item.Labels {
		v := strings.TrimSpace(l.Value)
		if v != "" {
			return v
		}
	}
	return ""
}

// collectNodes recursively traverses nested AOEF nodes and collects all nodes
// that reference an element (elementRef != ""), preserving the parent-child
// relationship via ParentElementID.
func collectNodes(nodes []aoefNode, parentElementID string, out *[]NodeLayout) {
	for _, n := range nodes {
		if n.ElementRef != "" {
			*out = append(*out, NodeLayout{
				ElementID:       n.ElementRef,
				ParentElementID: parentElementID,
				X:               n.X, Y: n.Y, W: n.W, H: n.H,
			})
			collectNodes(n.Children, n.ElementRef, out)
		} else {
			// Node without elementRef (pure grouping container) — pass parent through
			collectNodes(n.Children, parentElementID, out)
		}
	}
}
