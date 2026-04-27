package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
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
	Identifier    string             `xml:"identifier,attr"`
	Version       string             `xml:"version,attr"`
	Names         []aoefLangString   `xml:"name"`
	PropertyDefs  []aoefPropertyDef  `xml:"propertyDefinitions>propertyDefinition"`
	Properties    []aoefProperty     `xml:"properties>property"`
	Elements      []aoefElement      `xml:"elements>element"`
	Relationships []aoefRelationship `xml:"relationships>relationship"`
	Views         []aoefView         `xml:"views>diagrams>view"`
	Viewpoints    []aoefViewpoint    `xml:"views>viewpoints>viewpoint"`
	Organizations []aoefOrgItem      `xml:"organizations>item"`
}

type aoefViewpoint struct {
	ID             string             `xml:"identifier,attr"`
	Names          []aoefLangString   `xml:"name"`
	Documentations []aoefLangString   `xml:"documentation"`
	Purpose        string             `xml:"viewpointPurpose"`
	Content        string             `xml:"viewpointContent"`
	Concerns       []aoefConcern      `xml:"concern"`
	AllowedElems   []aoefAllowedType  `xml:"allowedElementType"`
	AllowedRels    []aoefAllowedType  `xml:"allowedRelationshipType"`
	Notes          []aoefModelingNote `xml:"modelingNote"`
}

type aoefConcern struct {
	Labels         []aoefLangString  `xml:"label"`
	Documentations []aoefLangString  `xml:"documentation"`
	Stakeholders   []aoefStakeholder `xml:"stakeholders>stakeholder"`
}

type aoefStakeholder struct {
	Labels []aoefLangString `xml:"label"`
}

type aoefAllowedType struct {
	Type string `xml:"type,attr"`
}

type aoefModelingNote struct {
	Type           string           `xml:"type,attr"`
	Documentations []aoefLangString `xml:"documentation"`
}

type aoefPropertyDef struct {
	ID       string `xml:"identifier,attr"`
	DataType string `xml:"type,attr"`
	Name     string `xml:"name"`
}

type aoefElement struct {
	ID             string           `xml:"identifier,attr"`
	Type           string           `xml:"type,attr"`
	Names          []aoefLangString `xml:"name"`
	Documentations []aoefLangString `xml:"documentation"`
	Properties     []aoefProperty   `xml:"properties>property"`
}

type aoefProperty struct {
	DefinitionRef string           `xml:"propertyDefinitionRef,attr"`
	Values        []aoefLangString `xml:"value"`
}

type aoefRelationship struct {
	ID             string           `xml:"identifier,attr"`
	Type           string           `xml:"type,attr"`
	Source         string           `xml:"source,attr"`
	Target         string           `xml:"target,attr"`
	Names          []aoefLangString `xml:"name"`
	Documentations []aoefLangString `xml:"documentation"`
	Properties     []aoefProperty   `xml:"properties>property"`
	// Type-specific attributes.
	AccessType string `xml:"accessType,attr"` // Access relationship
	IsDirected string `xml:"isDirected,attr"` // Association relationship ("true"/"false")
	Modifier   string `xml:"modifier,attr"`   // Influence relationship
}

type aoefView struct {
	ID             string           `xml:"identifier,attr"`
	Viewpoint      string           `xml:"viewpoint,attr"`
	ViewpointRef   string           `xml:"viewpointRef,attr"`
	Names          []aoefLangString `xml:"name"`
	Documentations []aoefLangString `xml:"documentation"`
	Properties     []aoefProperty   `xml:"properties>property"`
	Nodes          []aoefNode       `xml:"node"`
	Connections    []aoefConn       `xml:"connection"`
}

type aoefNode struct {
	Identifier      string           `xml:"identifier,attr"`
	NodeType        string           `xml:"http://www.w3.org/2001/XMLSchema-instance type,attr"`
	ElementRef      string           `xml:"elementRef,attr"`
	LabelExpression string           `xml:"labelExpression,attr"`
	X               int              `xml:"x,attr"`
	Y               int              `xml:"y,attr"`
	W               int              `xml:"w,attr"`
	H               int              `xml:"h,attr"`
	Children        []aoefNode       `xml:"node"`
	Style           *aoefStyle       `xml:"style"`
	Labels          []aoefLangString `xml:"label"`
}

type aoefConn struct {
	Identifier      string           `xml:"identifier,attr"`
	RelationshipRef string           `xml:"relationshipRef,attr"`
	Source          string           `xml:"source,attr"`
	Target          string           `xml:"target,attr"`
	Bendpoints      []aoefPoint      `xml:"bendpoint"`
	Style           *aoefStyle       `xml:"style"`
	Labels          []aoefLangString `xml:"label"`
}

type aoefPoint struct {
	X int `xml:"x,attr"`
	Y int `xml:"y,attr"`
}

type aoefStyle struct {
	LineWidth int           `xml:"lineWidth,attr"`
	LineColor *aoefRGBColor `xml:"lineColor"`
	FillColor *aoefRGBColor `xml:"fillColor"`
	Font      *aoefFont     `xml:"font"`
}

type aoefRGBColor struct {
	R int    `xml:"r,attr"`
	G int    `xml:"g,attr"`
	B int    `xml:"b,attr"`
	A string `xml:"a,attr"` // optional 0–100; empty string means not set
}

type aoefFont struct {
	Name  string        `xml:"name,attr"`
	Size  string        `xml:"size,attr"`
	Style string        `xml:"style,attr"`
	Color *aoefRGBColor `xml:"color"`
}

// aoefOrgItem represents a node in <organizations>.
// Items with identifierRef are leaves (elements or views); items with <label>
// and children are folders.
type aoefOrgItem struct {
	IdentifierRef string           `xml:"identifierRef,attr"`
	Labels        []aoefLangString `xml:"label"`
	Children      []aoefOrgItem    `xml:"item"`
}

type aoefLangString struct {
	Lang  string `xml:"http://www.w3.org/XML/1998/namespace lang,attr"`
	Value string `xml:",chardata"`
}

// toLangStrings converts raw AOEF lang strings to the model type, skipping blanks.
func toLangStrings(ss []aoefLangString) []LangString {
	out := make([]LangString, 0, len(ss))
	for _, s := range ss {
		v := strings.TrimSpace(s.Value)
		if v != "" {
			out = append(out, LangString{Lang: s.Lang, Value: v})
		}
	}
	return out
}

// firstLang returns the first non-empty value from a slice of lang strings.
// If lang is specified, it prefers the matching entry, falling back to the first non-empty.
func firstLang(ss []aoefLangString, prefer string) string {
	fallback := ""
	for _, s := range ss {
		v := strings.TrimSpace(s.Value)
		if v == "" {
			continue
		}
		if fallback == "" {
			fallback = v
		}
		if prefer == "" || s.Lang == prefer || s.Lang == "" {
			return v
		}
	}
	return fallback
}

func (m *aoefModel) toModel() *Model {
	out := &Model{
		Identifier: m.Identifier,
		Name:       firstLang(m.Names, ""),
		Version:    m.Version,
	}

	// Build property definition ID → (name, dataType) lookup and populate output.
	type propDef struct{ name, dataType string }
	propDefs := make(map[string]propDef, len(m.PropertyDefs))
	for _, pd := range m.PropertyDefs {
		dt := pd.DataType
		if dt == "" {
			dt = "string"
		}
		propDefs[pd.ID] = propDef{name: pd.Name, dataType: dt}
		out.PropertyDefinitions = append(out.PropertyDefinitions, PropertyDefinition{
			ID:       pd.ID,
			Name:     pd.Name,
			DataType: dt,
		})
	}

	// Model-level properties.
	for _, p := range m.Properties {
		def := propDefs[p.DefinitionRef]
		key := def.name
		if key == "" {
			key = p.DefinitionRef
		}
		val := firstLang(p.Values, "")
		if key != "" && val != "" {
			out.Properties = append(out.Properties, Property{
				DefinitionRef: p.DefinitionRef,
				Key:           key,
				Value:         val,
			})
		}
	}

	for _, e := range m.Elements {
		elem := Element{
			ID:             e.ID,
			Type:           e.Type,
			Name:           firstLang(e.Names, ""),
			Documentation:  firstLang(e.Documentations, ""),
			Names:          toLangStrings(e.Names),
			Documentations: toLangStrings(e.Documentations),
		}
		for _, p := range e.Properties {
			def := propDefs[p.DefinitionRef]
			key := def.name
			if key == "" {
				key = p.DefinitionRef
			}
			val := firstLang(p.Values, "")
			if key != "" && val != "" {
				elem.Properties = append(elem.Properties, Property{
					DefinitionRef: p.DefinitionRef,
					Key:           key,
					Value:         val,
				})
			}
		}
		out.Elements = append(out.Elements, elem)
	}

	for _, r := range m.Relationships {
		relType := r.Type
		if relType != "" && !strings.HasSuffix(relType, "Relationship") {
			relType += "Relationship"
		}
		rel := Relationship{
			ID:             r.ID,
			Type:           relType,
			Source:         r.Source,
			Target:         r.Target,
			Name:           firstLang(r.Names, ""),
			Documentation:  firstLang(r.Documentations, ""),
			Names:          toLangStrings(r.Names),
			Documentations: toLangStrings(r.Documentations),
			AccessType:     r.AccessType,
			Modifier:       r.Modifier,
		}
		if strings.EqualFold(r.IsDirected, "true") {
			rel.IsDirected = true
		}
		for _, p := range r.Properties {
			def := propDefs[p.DefinitionRef]
			key := def.name
			if key == "" {
				key = p.DefinitionRef
			}
			val := firstLang(p.Values, "")
			if key != "" && val != "" {
				rel.Properties = append(rel.Properties, Property{
					DefinitionRef: p.DefinitionRef,
					Key:           key,
					Value:         val,
				})
			}
		}
		out.Relationships = append(out.Relationships, rel)
	}

	for _, v := range m.Views {
		d := Diagram{
			ID:             v.ID,
			Name:           firstLang(v.Names, ""),
			Documentation:  firstLang(v.Documentations, ""),
			Names:          toLangStrings(v.Names),
			Documentations: toLangStrings(v.Documentations),
			Viewpoint:      v.Viewpoint,
			ViewpointRef:   v.ViewpointRef,
		}
		for _, p := range v.Properties {
			def := propDefs[p.DefinitionRef]
			key := def.name
			if key == "" {
				key = p.DefinitionRef
			}
			val := firstLang(p.Values, "")
			if key != "" && val != "" {
				d.Properties = append(d.Properties, Property{
					DefinitionRef: p.DefinitionRef,
					Key:           key,
					Value:         val,
				})
			}
		}
		// Build node-identifier → element-ID map for connection source/target resolution.
		nodeToElem := map[string]string{}
		var buildNodeMap func(nodes []aoefNode)
		buildNodeMap = func(nodes []aoefNode) {
			for _, n := range nodes {
				if n.Identifier != "" && n.ElementRef != "" {
					nodeToElem[n.Identifier] = n.ElementRef
				}
				buildNodeMap(n.Children)
			}
		}
		buildNodeMap(v.Nodes)

		collectNodes(v.Nodes, "", &d.Layout.Nodes)
		for _, c := range v.Connections {
			cl := ConnectionLayout{
				ConnectionID:    c.Identifier,
				RelationshipID:  c.RelationshipRef,
				SourceNodeID:    c.Source,
				TargetNodeID:    c.Target,
				SourceElementID: nodeToElem[c.Source],
				TargetElementID: nodeToElem[c.Target],
				Label:           firstLang(c.Labels, ""),
				Style:           convertConnStyle(c.Style),
			}
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

	// Parse viewpoint definitions.
	for _, vp := range m.Viewpoints {
		v := Viewpoint{
			ID:            vp.ID,
			Name:          firstLang(vp.Names, ""),
			Documentation: firstLang(vp.Documentations, ""),
			Purpose:       strings.TrimSpace(vp.Purpose),
			Content:       strings.TrimSpace(vp.Content),
		}
		for _, ae := range vp.AllowedElems {
			v.AllowedElementTypes = append(v.AllowedElementTypes, ae.Type)
		}
		for _, ar := range vp.AllowedRels {
			v.AllowedRelationshipTypes = append(v.AllowedRelationshipTypes, ar.Type)
		}
		for _, c := range vp.Concerns {
			concern := ViewpointConcern{
				Label:         firstLang(c.Labels, ""),
				Documentation: firstLang(c.Documentations, ""),
			}
			for _, s := range c.Stakeholders {
				concern.Stakeholders = append(concern.Stakeholders, firstLang(s.Labels, ""))
			}
			v.Concerns = append(v.Concerns, concern)
		}
		for _, n := range vp.Notes {
			v.ModelingNotes = append(v.ModelingNotes, ViewpointModelingNote{
				Type:          n.Type,
				Documentation: firstLang(n.Documentations, ""),
			})
		}
		out.Viewpoints = append(out.Viewpoints, v)
	}

	return out
}

// convertRGBColor converts an AOEF colour to the model type.
func convertRGBColor(c *aoefRGBColor) *RGBColor {
	if c == nil {
		return nil
	}
	col := &RGBColor{R: c.R, G: c.G, B: c.B}
	if c.A != "" {
		v, err := strconv.Atoi(c.A)
		if err == nil {
			col.A = &v
		}
	}
	return col
}

// convertFont converts an AOEF font to the model type.
func convertFont(f *aoefFont) *FontStyle {
	if f == nil {
		return nil
	}
	return &FontStyle{
		Name:  f.Name,
		Size:  f.Size,
		Style: f.Style,
		Color: convertRGBColor(f.Color),
	}
}

// convertNodeStyle converts an AOEF style to a NodeStyle.
func convertNodeStyle(s *aoefStyle) *NodeStyle {
	if s == nil {
		return nil
	}
	ns := &NodeStyle{
		FillColor: convertRGBColor(s.FillColor),
		LineColor: convertRGBColor(s.LineColor),
		Font:      convertFont(s.Font),
		LineWidth: s.LineWidth,
	}
	// Return nil if nothing was set.
	if ns.FillColor == nil && ns.LineColor == nil && ns.Font == nil && ns.LineWidth == 0 {
		return nil
	}
	return ns
}

// convertConnStyle converts an AOEF style to a ConnStyle.
func convertConnStyle(s *aoefStyle) *ConnStyle {
	if s == nil {
		return nil
	}
	cs := &ConnStyle{
		LineColor: convertRGBColor(s.LineColor),
		Font:      convertFont(s.Font),
		LineWidth: s.LineWidth,
	}
	if cs.LineColor == nil && cs.Font == nil && cs.LineWidth == 0 {
		return nil
	}
	return cs
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
		return nil, nil, false // element/relationship ref — skip
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
				NodeID:          n.Identifier,
				ElementID:       n.ElementRef,
				ParentElementID: parentElementID,
				NodeType:        n.NodeType,
				LabelExpression: n.LabelExpression,
				X:               n.X, Y: n.Y, W: n.W, H: n.H,
				Style: convertNodeStyle(n.Style),
			})
			collectNodes(n.Children, n.ElementRef, out)
		} else if n.Identifier != "" && n.W > 0 && n.H > 0 {
			// Group node: has a diagram-level identifier and optional label but no element reference.
			// Only emit if it has actual dimensions (skip degenerate placeholder nodes).
			label := firstLang(n.Labels, "")
			*out = append(*out, NodeLayout{
				NodeID:          n.Identifier,
				ElementID:       n.Identifier,
				ParentElementID: parentElementID,
				NodeType:        n.NodeType,
				Label:           label,
				LabelExpression: n.LabelExpression,
				ElementType:     "Group",
				X:               n.X, Y: n.Y, W: n.W, H: n.H,
				Style: convertNodeStyle(n.Style),
			})
			collectNodes(n.Children, n.Identifier, out)
		} else {
			// Node without identifier or elementRef — pass parent through.
			collectNodes(n.Children, parentElementID, out)
		}
	}
}
