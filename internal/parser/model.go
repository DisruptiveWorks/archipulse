package parser

// Model is the in-memory representation of a parsed ArchiMate model.
// It is format-agnostic — produced by both the AOEF and AJX parsers.
type Model struct {
	Name          string
	Elements      []Element
	Relationships []Relationship
	Diagrams      []Diagram
}

// Element represents an ArchiMate element (AOEF <element>).
type Element struct {
	ID            string
	Type          string
	Name          string
	Documentation string
}

// Relationship represents an ArchiMate relationship (AOEF <relationship>).
type Relationship struct {
	ID            string
	Type          string
	Source        string // element ID
	Target        string // element ID
	Name          string
	Documentation string
}

// Diagram represents an ArchiMate view (AOEF <view>).
type Diagram struct {
	ID            string
	Name          string
	Documentation string
	Layout        DiagramLayout
}

// DiagramLayout holds the visual positions of nodes and connections within a diagram.
type DiagramLayout struct {
	Nodes       []NodeLayout
	Connections []ConnectionLayout
}

// NodeLayout holds the position and size of an element within a diagram.
type NodeLayout struct {
	ElementID string
	X, Y      int
	W, H      int
}

// ConnectionLayout holds the visual path of a relationship within a diagram.
type ConnectionLayout struct {
	RelationshipID string
	Bendpoints     []Point
}

// Point is a 2D coordinate.
type Point struct {
	X, Y int
}
