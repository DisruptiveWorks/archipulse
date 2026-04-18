// Package exporter serializes a workspace to AOEF (XML) and AJX (JSON).
package exporter

import (
	"encoding/xml"
	"fmt"
	"io"
	"sort"

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
	Identifier    string             `xml:"identifier,attr,omitempty"`
	XMLNS         string             `xml:"xmlns,attr"`
	XMLNSxsi      string             `xml:"xmlns:xsi,attr"`
	SchemaLoc     string             `xml:"xsi:schemaLocation,attr"`
	Name          string             `xml:"name"`
	Properties    *aoefProperties    `xml:"properties,omitempty"`
	Elements      *aoefElements      `xml:"elements,omitempty"`
	Relationships *aoefRelationships `xml:"relationships,omitempty"`
	Organizations *aoefOrgsBlock     `xml:"organizations,omitempty"`
	PropertyDefs  *aoefPropertyDefs  `xml:"propertyDefinitions,omitempty"`
	Views         *aoefViews         `xml:"views,omitempty"`
}

type aoefPropertyDefs struct {
	Items []aoefPropertyDef `xml:"propertyDefinition"`
}

type aoefPropertyDef struct {
	ID       string `xml:"identifier,attr"`
	DataType string `xml:"type,attr"`
	Name     string `xml:"name"`
}

type aoefProperties struct {
	Items []aoefProperty `xml:"property"`
}

type aoefProperty struct {
	DefinitionRef string `xml:"propertyDefinitionRef,attr"`
	Value         string `xml:"value"`
}

type aoefElements struct {
	Items []aoefElement `xml:"element"`
}

type aoefRelationships struct {
	Items []aoefRelationship `xml:"relationship"`
}

type aoefViews struct {
	Viewpoints *aoefViewpointsBlock `xml:"viewpoints,omitempty"`
	Diagrams   aoefDiagrams         `xml:"diagrams"`
}

type aoefDiagrams struct {
	Items []aoefView `xml:"view"`
}

type aoefViewpointsBlock struct {
	Items []aoefViewpointOut `xml:"viewpoint"`
}

type aoefViewpointOut struct {
	ID            string               `xml:"identifier,attr"`
	Name          string               `xml:"name"`
	Documentation string               `xml:"documentation,omitempty"`
	Purpose       string               `xml:"viewpointPurpose,omitempty"`
	Content       string               `xml:"viewpointContent,omitempty"`
	Concerns      []aoefConcernOut     `xml:"concern"`
	AllowedElems  []aoefAllowedTypeOut `xml:"allowedElementType"`
	AllowedRels   []aoefAllowedTypeOut `xml:"allowedRelationshipType"`
	Notes         []aoefNoteOut        `xml:"modelingNote"`
}

type aoefConcernOut struct {
	Label         string            `xml:"label"`
	Documentation string            `xml:"documentation,omitempty"`
	Stakeholders  *aoefStakeholders `xml:"stakeholders,omitempty"`
}

type aoefStakeholders struct {
	Items []aoefStakeholderOut `xml:"stakeholder"`
}

type aoefStakeholderOut struct {
	Label string `xml:"label"`
}

type aoefAllowedTypeOut struct {
	Type string `xml:"type,attr"`
}

type aoefNoteOut struct {
	Type          string `xml:"type,attr,omitempty"`
	Documentation string `xml:"documentation,omitempty"`
}

// aoefLangString is a multi-language string element (name, documentation, label, value).
type aoefLangString struct {
	Lang  string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	Value string `xml:",chardata"`
}

type aoefElement struct {
	ID             string           `xml:"identifier,attr"`
	Type           string           `xml:"xsi:type,attr"`
	Names          []aoefLangString `xml:"name"`
	Documentations []aoefLangString `xml:"documentation"`
	Properties     *aoefProperties  `xml:"properties,omitempty"`
}

type aoefRelationship struct {
	ID             string           `xml:"identifier,attr"`
	Type           string           `xml:"xsi:type,attr"`
	Source         string           `xml:"source,attr"`
	Target         string           `xml:"target,attr"`
	AccessType     string           `xml:"accessType,attr,omitempty"`
	IsDirected     string           `xml:"isDirected,attr,omitempty"`
	Modifier       string           `xml:"modifier,attr,omitempty"`
	Names          []aoefLangString `xml:"name"`
	Documentations []aoefLangString `xml:"documentation"`
	Properties     *aoefProperties  `xml:"properties,omitempty"`
}

type aoefView struct {
	ID             string           `xml:"identifier,attr"`
	Type           string           `xml:"xsi:type,attr,omitempty"`
	Viewpoint      string           `xml:"viewpoint,attr,omitempty"`
	ViewpointRef   string           `xml:"viewpointRef,attr,omitempty"`
	Names          []aoefLangString `xml:"name"`
	Documentations []aoefLangString `xml:"documentation"`
	Properties     *aoefProperties  `xml:"properties,omitempty"`
	Nodes          []aoefNode       `xml:"node"`
	Connections    []aoefConn       `xml:"connection"`
}

// aoefNode mirrors the parser's aoefNode, supporting recursive nesting.
type aoefNode struct {
	Identifier      string     `xml:"identifier,attr"`
	ElementRef      string     `xml:"elementRef,attr,omitempty"`
	NodeType        string     `xml:"xsi:type,attr,omitempty"`
	LabelExpression string     `xml:"labelExpression,attr,omitempty"`
	X               int        `xml:"x,attr"`
	Y               int        `xml:"y,attr"`
	W               int        `xml:"w,attr"`
	H               int        `xml:"h,attr"`
	Style           *aoefStyle `xml:"style"`
	Children        []aoefNode `xml:"node"`
}

type aoefConn struct {
	Identifier      string      `xml:"identifier,attr,omitempty"`
	RelationshipRef string      `xml:"relationshipRef,attr"`
	Type            string      `xml:"xsi:type,attr,omitempty"`
	Source          string      `xml:"source,attr"`
	Target          string      `xml:"target,attr"`
	Style           *aoefStyle  `xml:"style"`
	Bendpoints      []aoefPoint `xml:"bendpoint"`
}

type aoefPoint struct {
	X int `xml:"x,attr"`
	Y int `xml:"y,attr"`
}

type aoefOrgsBlock struct {
	Items []aoefOrgItem `xml:"item"`
}

type aoefOrgItem struct {
	IdentifierRef string        `xml:"identifierRef,attr,omitempty"`
	Label         string        `xml:"label,omitempty"`
	Children      []aoefOrgItem `xml:"item"`
}

type aoefStyle struct {
	LineWidth int           `xml:"lineWidth,attr,omitempty"`
	LineColor *aoefRGBColor `xml:"lineColor"`
	FillColor *aoefRGBColor `xml:"fillColor"`
	Font      *aoefFont     `xml:"font"`
}

type aoefRGBColor struct {
	R int    `xml:"r,attr"`
	G int    `xml:"g,attr"`
	B int    `xml:"b,attr"`
	A string `xml:"a,attr,omitempty"`
}

type aoefFont struct {
	Name  string        `xml:"name,attr,omitempty"`
	Size  string        `xml:"size,attr,omitempty"`
	Style string        `xml:"style,attr,omitempty"`
	Color *aoefRGBColor `xml:"color"`
}

func toAOEFModel(m *parser.Model) *aoefModel {
	out := &aoefModel{
		Identifier: m.Identifier,
		XMLNS:      "http://www.opengroup.org/xsd/archimate/3.0/",
		XMLNSxsi:   "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLoc:  "http://www.opengroup.org/xsd/archimate/3.0/ http://www.opengroup.org/xsd/archimate/3.1/archimate3_Diagram.xsd",
		Name:       m.Name,
		Properties: buildProperties(m.Properties),
	}

	if len(m.PropertyDefinitions) > 0 {
		defs := make([]aoefPropertyDef, len(m.PropertyDefinitions))
		for i, pd := range m.PropertyDefinitions {
			defs[i] = aoefPropertyDef{ID: pd.ID, DataType: pd.DataType, Name: pd.Name}
		}
		out.PropertyDefs = &aoefPropertyDefs{Items: defs}
	}

	if len(m.Elements) > 0 {
		elems := make([]aoefElement, len(m.Elements))
		for i, e := range m.Elements {
			elems[i] = aoefElement{
				ID:             e.ID,
				Type:           e.Type,
				Names:          toLangStrings(e.Names, e.Name),
				Documentations: toLangStrings(e.Documentations, e.Documentation),
				Properties:     buildProperties(e.Properties),
			}
		}
		out.Elements = &aoefElements{Items: elems}
	}

	if len(m.Relationships) > 0 {
		rels := make([]aoefRelationship, len(m.Relationships))
		for i, r := range m.Relationships {
			rel := aoefRelationship{
				ID:             r.ID,
				Type:           r.Type,
				Source:         r.Source,
				Target:         r.Target,
				AccessType:     r.AccessType,
				Modifier:       r.Modifier,
				Names:          toLangStrings(r.Names, r.Name),
				Documentations: toLangStrings(r.Documentations, r.Documentation),
				Properties:     buildProperties(r.Properties),
			}
			if r.IsDirected {
				rel.IsDirected = "true"
			}
			rels[i] = rel
		}
		out.Relationships = &aoefRelationships{Items: rels}
	}

	if len(m.Diagrams) > 0 || len(m.Viewpoints) > 0 {
		diagViews := make([]aoefView, len(m.Diagrams))
		for i, d := range m.Diagrams {
			v := aoefView{
				ID:             d.ID,
				Type:           "Diagram",
				Viewpoint:      d.Viewpoint,
				ViewpointRef:   d.ViewpointRef,
				Names:          toLangStrings(d.Names, d.Name),
				Documentations: toLangStrings(d.Documentations, d.Documentation),
				Properties:     buildProperties(d.Properties),
			}
			v.Nodes = buildNodeTree(d.Layout.Nodes)
			for _, c := range d.Layout.Connections {
				conn := aoefConn{
					Identifier:      c.ConnectionID,
					RelationshipRef: c.RelationshipID,
					Type:            "Relationship",
					Source:          c.SourceNodeID,
					Target:          c.TargetNodeID,
					Style:           convertConnStyle(c.Style),
				}
				for _, bp := range c.Bendpoints {
					conn.Bendpoints = append(conn.Bendpoints, aoefPoint{X: bp.X, Y: bp.Y})
				}
				v.Connections = append(v.Connections, conn)
			}
			diagViews[i] = v
		}
		outViews := &aoefViews{Diagrams: aoefDiagrams{Items: diagViews}}
		if len(m.Viewpoints) > 0 {
			outViews.Viewpoints = buildViewpoints(m.Viewpoints)
		}
		out.Views = outViews
	}

	if len(m.ViewFolders) > 0 || len(m.DiagramFolders) > 0 {
		out.Organizations = buildOrganizations(m.ViewFolders, m.DiagramFolders)
	}

	return out
}

// buildOrganizations reconstructs the AOEF <organizations> tree from the flat
// ViewFolders (sorted parent-first) and DiagramFolders (diagram→folder links).
func buildOrganizations(folders []parser.ViewFolder, diagFolders []parser.DiagramFolder) *aoefOrgsBlock {
	// Map folder SourceID → children SourceIDs (folder order preserved by Position).
	type folderNode struct {
		sourceID string
		name     string
		parentID string
		position int
	}
	nodes := make([]folderNode, len(folders))
	for i, f := range folders {
		nodes[i] = folderNode{f.SourceID, f.Name, f.ParentID, f.Position}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].position < nodes[j].position })

	// Map parentID → child folder nodes.
	childFolders := make(map[string][]folderNode, len(nodes))
	for _, n := range nodes {
		childFolders[n.parentID] = append(childFolders[n.parentID], n)
	}

	// Map folderSourceID → diagram sourceIDs.
	folderDiags := make(map[string][]string, len(diagFolders))
	for _, df := range diagFolders {
		folderDiags[df.FolderSourceID] = append(folderDiags[df.FolderSourceID], df.DiagramSourceID)
	}

	var buildItem func(n folderNode) aoefOrgItem
	buildItem = func(n folderNode) aoefOrgItem {
		item := aoefOrgItem{Label: n.name}
		for _, diagID := range folderDiags[n.sourceID] {
			item.Children = append(item.Children, aoefOrgItem{IdentifierRef: diagID})
		}
		for _, child := range childFolders[n.sourceID] {
			item.Children = append(item.Children, buildItem(child))
		}
		return item
	}

	// Root folders: parentID == "". Each root folder becomes a top-level <item>.
	block := &aoefOrgsBlock{}
	for _, root := range childFolders[""] {
		block.Items = append(block.Items, buildItem(root))
	}
	// Diagrams with no folder go directly as top-level <item identifierRef="...">.
	for _, diagID := range folderDiags[""] {
		block.Items = append(block.Items, aoefOrgItem{IdentifierRef: diagID})
	}
	return block
}

func buildViewpoints(vps []parser.Viewpoint) *aoefViewpointsBlock {
	items := make([]aoefViewpointOut, len(vps))
	for i, vp := range vps {
		out := aoefViewpointOut{
			ID:            vp.ID,
			Name:          vp.Name,
			Documentation: vp.Documentation,
			Purpose:       vp.Purpose,
			Content:       vp.Content,
		}
		for _, ae := range vp.AllowedElementTypes {
			out.AllowedElems = append(out.AllowedElems, aoefAllowedTypeOut{Type: ae})
		}
		for _, ar := range vp.AllowedRelationshipTypes {
			out.AllowedRels = append(out.AllowedRels, aoefAllowedTypeOut{Type: ar})
		}
		for _, c := range vp.Concerns {
			co := aoefConcernOut{Label: c.Label, Documentation: c.Documentation}
			if len(c.Stakeholders) > 0 {
				sh := make([]aoefStakeholderOut, len(c.Stakeholders))
				for j, s := range c.Stakeholders {
					sh[j] = aoefStakeholderOut{Label: s}
				}
				co.Stakeholders = &aoefStakeholders{Items: sh}
			}
			out.Concerns = append(out.Concerns, co)
		}
		for _, n := range vp.ModelingNotes {
			out.Notes = append(out.Notes, aoefNoteOut{Type: n.Type, Documentation: n.Documentation})
		}
		items[i] = out
	}
	return &aoefViewpointsBlock{Items: items}
}

// toLangStrings converts parser.LangString slices to the exporter XML type.
// Falls back to a single entry from the plain string when no lang variants exist.
func toLangStrings(ls []parser.LangString, fallback string) []aoefLangString {
	if len(ls) > 0 {
		out := make([]aoefLangString, len(ls))
		for i, s := range ls {
			out[i] = aoefLangString{Lang: s.Lang, Value: s.Value}
		}
		return out
	}
	if fallback != "" {
		return []aoefLangString{{Value: fallback}}
	}
	return nil
}

func buildProperties(props []parser.Property) *aoefProperties {
	if len(props) == 0 {
		return nil
	}
	items := make([]aoefProperty, len(props))
	for i, p := range props {
		items[i] = aoefProperty{DefinitionRef: p.DefinitionRef, Value: p.Value}
	}
	return &aoefProperties{Items: items}
}

func convertRGBColor(c *parser.RGBColor) *aoefRGBColor {
	if c == nil {
		return nil
	}
	out := &aoefRGBColor{R: c.R, G: c.G, B: c.B}
	if c.A != nil {
		out.A = fmt.Sprintf("%d", *c.A)
	}
	return out
}

func convertFont(f *parser.FontStyle) *aoefFont {
	if f == nil {
		return nil
	}
	return &aoefFont{
		Name:  f.Name,
		Size:  f.Size,
		Style: f.Style,
		Color: convertRGBColor(f.Color),
	}
}

func convertNodeStyle(s *parser.NodeStyle) *aoefStyle {
	if s == nil {
		return nil
	}
	return &aoefStyle{
		LineWidth: s.LineWidth,
		FillColor: convertRGBColor(s.FillColor),
		LineColor: convertRGBColor(s.LineColor),
		Font:      convertFont(s.Font),
	}
}

func convertConnStyle(s *parser.ConnStyle) *aoefStyle {
	if s == nil {
		return nil
	}
	return &aoefStyle{
		LineWidth: s.LineWidth,
		LineColor: convertRGBColor(s.LineColor),
		Font:      convertFont(s.Font),
	}
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

	// placed tracks which node indices have already been emitted into the tree.
	// This prevents cycles (e.g. a node whose ElementID == its own ParentElementID)
	// and duplicate placement when the same node is claimed by multiple parents.
	placed := make(map[int]bool, len(nodes))

	// Recursively build an aoefNode and all its descendants.
	var buildNode func(i int) aoefNode
	buildNode = func(i int) aoefNode {
		placed[i] = true
		n := nodes[i]
		// Label/Group nodes have ElementID == NodeID (set as a placeholder by the
		// parser when there is no real elementRef). Do not emit elementRef for them.
		elementRef := n.ElementID
		if elementRef == n.NodeID {
			elementRef = ""
		}
		node := aoefNode{
			Identifier:      n.NodeID,
			ElementRef:      elementRef,
			NodeType:        n.NodeType,
			LabelExpression: n.LabelExpression,
			X:               n.X,
			Y:               n.Y,
			W:               n.W,
			H:               n.H,
			Style:           convertNodeStyle(n.Style),
		}
		// Children reference this node via its ElementID (element node) or
		// NodeID (group node — ElementID is empty).
		parentKey := n.ElementID
		if parentKey == "" {
			parentKey = n.NodeID
		}
		for _, ci := range childrenOf[parentKey] {
			if placed[ci] {
				continue // already in tree — skip to avoid cycles/duplicates
			}
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
		if nodes[i].ParentElementID == "" && !placed[i] {
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
