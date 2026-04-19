package parser

// Model is the in-memory representation of a parsed ArchiMate model.
type Model struct {
	Identifier          string // AOEF model/@identifier
	Name                string
	Version             string
	Properties          []Property // model-level <properties>
	Elements            []Element
	Relationships       []Relationship
	Diagrams            []Diagram
	ViewFolders         []ViewFolder         // folder hierarchy from <organizations>
	DiagramFolders      []DiagramFolder      // diagram → folder assignments
	PropertyDefinitions []PropertyDefinition // typed property definitions
	Viewpoints          []Viewpoint          // viewpoint definitions from <views><viewpoints>
}

// Viewpoint is an ArchiMate viewpoint definition from <viewpoints><viewpoint>.
type Viewpoint struct {
	ID                       string
	Name                     string
	Documentation            string
	Purpose                  string // space-separated: Designing|Deciding|Informing
	Content                  string // space-separated: Details|Coherence|Overview
	Concerns                 []ViewpointConcern
	AllowedElementTypes      []string
	AllowedRelationshipTypes []string
	ModelingNotes            []ViewpointModelingNote
}

// ViewpointConcern is a stakeholder concern attached to a viewpoint.
type ViewpointConcern struct {
	Label         string
	Documentation string
	Stakeholders  []string
}

// ViewpointModelingNote is a guidance note attached to a viewpoint.
type ViewpointModelingNote struct {
	Type          string
	Documentation string
}

// PropertyDefinition mirrors an OEF <propertyDefinition> with its data type.
type PropertyDefinition struct {
	ID       string // source identifier from AOEF
	Name     string
	DataType string // string|boolean|currency|date|time|number
}

// ViewFolder represents a folder node in the diagram view hierarchy.
type ViewFolder struct {
	SourceID string
	Name     string
	ParentID string // empty if root-level
	Position int
}

// DiagramFolder links a diagram to its folder.
type DiagramFolder struct {
	DiagramSourceID string
	FolderSourceID  string // empty if diagram is at root (no folder)
}

// LangString is a string value paired with an optional xml:lang tag.
type LangString struct {
	Lang  string // empty means no xml:lang attribute
	Value string
}

// Element represents an ArchiMate element (AOEF <element>).
type Element struct {
	ID             string
	Type           string
	Name           string       // preferred/first language — backward compat
	Documentation  string       // preferred/first language — backward compat
	Names          []LangString // all xml:lang variants
	Documentations []LangString // all xml:lang variants
	Properties     []Property
}

// Property is a key/value pair attached to an element or relationship, sourced from the model file.
type Property struct {
	DefinitionRef string // original AOEF propertyDefinitionRef — preserved for round-trip export
	Key           string
	Value         string
}

// Relationship represents an ArchiMate relationship (AOEF <relationship>).
type Relationship struct {
	ID             string
	Type           string
	Source         string       // element source_id
	Target         string       // element source_id
	Name           string       // preferred/first language — backward compat
	Documentation  string       // preferred/first language — backward compat
	Names          []LangString // all xml:lang variants
	Documentations []LangString // all xml:lang variants
	// Type-specific semantic attributes (OEF standard).
	AccessType string // Access relationship: Access|Read|Write|ReadWrite
	IsDirected bool   // Association relationship: directed flag
	Modifier   string // Influence relationship: strength/sign e.g. "+", "--", "5"
	Properties []Property
}

// Diagram represents an ArchiMate view (AOEF <view>).
type Diagram struct {
	ID             string
	Name           string       // preferred/first language — backward compat
	Documentation  string       // preferred/first language — backward compat
	Names          []LangString // all xml:lang variants
	Documentations []LangString // all xml:lang variants
	Viewpoint      string       // viewpoint attribute (named or free-form)
	ViewpointRef   string       // identifierRef to a defined viewpoint
	Properties     []Property   // view-level <properties>
	Layout         DiagramLayout
}

// DiagramLayout holds the visual positions of nodes and connections within a diagram.
type DiagramLayout struct {
	Nodes       []NodeLayout
	Connections []ConnectionLayout
}

// NodeLayout holds the position, size, and optional style of a node within a diagram.
type NodeLayout struct {
	NodeID          string // diagram-level node identifier from OEF (unique within the view)
	ElementID       string
	ParentElementID string // empty if top-level node
	NodeType        string // xsi:type: Element|Container|Label|etc. (empty = Element)
	Label           string // used for group nodes that have no element reference
	LabelExpression string // Archi-specific labelExpression attr (preserved for round-trip, not evaluated)
	ElementType     string // overrides DB lookup when set (e.g. "Group")
	X, Y            int
	W, H            int
	Style           *NodeStyle
}

// ConnectionLayout holds the visual path and optional style of a connection.
type ConnectionLayout struct {
	ConnectionID    string // diagram-level connection identifier (OEF @identifier)
	RelationshipID  string
	SourceNodeID    string // OEF node identifier of the visual source node
	TargetNodeID    string // OEF node identifier of the visual target node
	SourceElementID string // element ID of the connection's visual source node
	TargetElementID string // element ID of the connection's visual target node
	Label           string // override label shown on the connection in this diagram
	Bendpoints      []Point
	Style           *ConnStyle
}

// NodeStyle captures visual styling from OEF <style> on a node.
type NodeStyle struct {
	FillColor *RGBColor
	LineColor *RGBColor
	Font      *FontStyle
	LineWidth int // 0 means not set
}

// ConnStyle captures visual styling from OEF <style> on a connection.
type ConnStyle struct {
	LineColor *RGBColor
	Font      *FontStyle
	LineWidth int // 0 means not set
}

// RGBColor is an RGBA colour value from OEF.
// Alpha is 0–100 (0 = transparent, 100 = opaque); nil means not specified.
type RGBColor struct {
	R, G, B int
	A       *int
}

// FontStyle captures font metadata from OEF <font>.
type FontStyle struct {
	Name  string    // font family name
	Size  string    // size in points (stored as string to preserve half-granularity)
	Style string    // space-separated: plain|bold|italic|underline
	Color *RGBColor // text colour
}

// Point is a 2D coordinate.
type Point struct {
	X, Y int
}
