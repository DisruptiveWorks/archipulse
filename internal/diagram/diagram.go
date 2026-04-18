// Package diagram manages ArchiMate diagrams (views) within a workspace.
package diagram

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("diagram not found")
var ErrConflict = errors.New("diagram was modified by another request")

type Diagram struct {
	ID            uuid.UUID       `json:"id"`
	WorkspaceID   uuid.UUID       `json:"workspace_id"`
	SourceID      string          `json:"source_id"`
	Name          string          `json:"name"`
	Documentation string          `json:"documentation"`
	Layout        json.RawMessage `json:"layout"`
	Version       int             `json:"version"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) List(workspaceID uuid.UUID) ([]Diagram, error) {
	rows, err := s.db.Query(`
		SELECT id, workspace_id, source_id, name, documentation, layout, version, created_at, updated_at
		FROM diagrams WHERE workspace_id = $1 ORDER BY name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list diagrams: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Diagram
	for rows.Next() {
		var d Diagram
		if err := rows.Scan(&d.ID, &d.WorkspaceID, &d.SourceID, &d.Name,
			&d.Documentation, &d.Layout, &d.Version, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) Get(id uuid.UUID) (*Diagram, error) {
	var d Diagram
	err := s.db.QueryRow(`
		SELECT id, workspace_id, source_id, name, documentation, layout, version, created_at, updated_at
		FROM diagrams WHERE id = $1`, id).
		Scan(&d.ID, &d.WorkspaceID, &d.SourceID, &d.Name,
			&d.Documentation, &d.Layout, &d.Version, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get diagram: %w", err)
	}
	return &d, nil
}

func (s *Store) Create(workspaceID uuid.UUID, sourceID, name, documentation string, layout json.RawMessage) (*Diagram, error) {
	if layout == nil {
		layout = json.RawMessage("{}")
	}
	var d Diagram
	err := s.db.QueryRow(`
		INSERT INTO diagrams (workspace_id, source_id, name, documentation, layout)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, workspace_id, source_id, name, documentation, layout, version, created_at, updated_at`,
		workspaceID, sourceID, name, documentation, layout).
		Scan(&d.ID, &d.WorkspaceID, &d.SourceID, &d.Name,
			&d.Documentation, &d.Layout, &d.Version, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create diagram: %w", err)
	}
	return &d, nil
}

func (s *Store) Update(id uuid.UUID, name, documentation string, layout json.RawMessage, version int) (*Diagram, error) {
	if layout == nil {
		layout = json.RawMessage("{}")
	}
	var d Diagram
	err := s.db.QueryRow(`
		UPDATE diagrams
		SET name = $1, documentation = $2, layout = $3, version = version + 1, updated_at = now()
		WHERE id = $4 AND version = $5
		RETURNING id, workspace_id, source_id, name, documentation, layout, version, created_at, updated_at`,
		name, documentation, layout, id, version).
		Scan(&d.ID, &d.WorkspaceID, &d.SourceID, &d.Name,
			&d.Documentation, &d.Layout, &d.Version, &d.CreatedAt, &d.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		if _, err2 := s.Get(id); errors.Is(err2, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrConflict
	}
	if err != nil {
		return nil, fmt.Errorf("update diagram: %w", err)
	}
	return &d, nil
}

// RGBColor is an RGBA colour value as stored in the layout JSON.
type RGBColor struct {
	R int  `json:"r"`
	G int  `json:"g"`
	B int  `json:"b"`
	A *int `json:"a,omitempty"` // 0–100; nil means not set
}

// FontStyle holds font metadata as stored in the layout JSON.
type FontStyle struct {
	Name  string    `json:"name,omitempty"`
	Size  string    `json:"size,omitempty"`
	Style string    `json:"style,omitempty"`
	Color *RGBColor `json:"color,omitempty"`
}

// NodeStyle holds visual styling for a diagram node.
type NodeStyle struct {
	FillColor *RGBColor  `json:"fill_color,omitempty"`
	LineColor *RGBColor  `json:"line_color,omitempty"`
	Font      *FontStyle `json:"font,omitempty"`
	LineWidth int        `json:"line_width,omitempty"`
}

// ConnStyle holds visual styling for a diagram connection.
type ConnStyle struct {
	LineColor *RGBColor  `json:"line_color,omitempty"`
	Font      *FontStyle `json:"font,omitempty"`
	LineWidth int        `json:"line_width,omitempty"`
}

// RenderNode is a node enriched with element metadata for rendering.
type RenderNode struct {
	NodeID          string     `json:"node_id,omitempty"` // OEF diagram node identifier (unique within view)
	ElementID       string     `json:"element_id"`
	ParentElementID string     `json:"parent_element_id,omitempty"`
	NodeType        string     `json:"node_type,omitempty"`
	ElementName     string     `json:"element_name"`
	ElementType     string     `json:"element_type"`
	X               int        `json:"x"`
	Y               int        `json:"y"`
	W               int        `json:"w"`
	H               int        `json:"h"`
	Style           *NodeStyle `json:"style,omitempty"`
}

// RenderConnection is a connection enriched with relationship metadata.
type RenderConnection struct {
	RelationshipID   string     `json:"relationship_id"`
	RelationshipType string     `json:"relationship_type"`
	SourceNodeID     string     `json:"source_node_id,omitempty"` // OEF diagram node identifier
	TargetNodeID     string     `json:"target_node_id,omitempty"` // OEF diagram node identifier
	SourceElementID  string     `json:"source_element_id"`
	TargetElementID  string     `json:"target_element_id"`
	Reversed         bool       `json:"reversed,omitempty"` // true when the connection is drawn opposite to the semantic relationship direction
	Label            string     `json:"label,omitempty"`
	AccessType       string     `json:"access_type,omitempty"`
	IsDirected       bool       `json:"is_directed,omitempty"`
	Modifier         string     `json:"modifier,omitempty"`
	Bendpoints       []Point    `json:"bendpoints"`
	Style            *ConnStyle `json:"style,omitempty"`
}

// Point is a 2D coordinate.
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// RenderData is the full enriched payload for rendering a diagram.
type RenderData struct {
	ID            uuid.UUID          `json:"id"`
	Name          string             `json:"name"`
	Documentation string             `json:"documentation"`
	Nodes         []RenderNode       `json:"nodes"`
	Connections   []RenderConnection `json:"connections"`
}

// Render returns the diagram layout enriched with element and relationship metadata.
func (s *Store) Render(diagramID uuid.UUID) (*RenderData, error) {
	d, err := s.Get(diagramID)
	if err != nil {
		return nil, err
	}

	// Parse the stored layout JSON.
	var layout struct {
		Nodes []struct {
			NodeID          string     `json:"NodeID"`
			ElementID       string     `json:"ElementID"`
			ParentElementID string     `json:"ParentElementID"`
			NodeType        string     `json:"NodeType"`
			Label           string     `json:"Label"`
			ElementType     string     `json:"ElementType"`
			X               int        `json:"X"`
			Y               int        `json:"Y"`
			W               int        `json:"W"`
			H               int        `json:"H"`
			Style           *NodeStyle `json:"Style"`
		} `json:"Nodes"`
		Connections []struct {
			RelationshipID  string `json:"RelationshipID"`
			SourceNodeID    string `json:"SourceNodeID"`
			TargetNodeID    string `json:"TargetNodeID"`
			SourceElementID string `json:"SourceElementID"`
			TargetElementID string `json:"TargetElementID"`
			Label           string `json:"Label"`
			Bendpoints      []struct {
				X int `json:"X"`
				Y int `json:"Y"`
			} `json:"Bendpoints"`
			Style *ConnStyle `json:"Style"`
		} `json:"Connections"`
	}
	if err := json.Unmarshal(d.Layout, &layout); err != nil {
		return nil, fmt.Errorf("parse layout: %w", err)
	}

	// Resolve element names and types in one query.
	elemIDs := make([]string, 0, len(layout.Nodes))
	for _, n := range layout.Nodes {
		if n.ElementID != "" {
			elemIDs = append(elemIDs, n.ElementID)
		}
	}

	elemMeta := map[string][2]string{} // source_id → [name, type]
	if len(elemIDs) > 0 {
		rows, err := s.db.Query(`
			SELECT source_id, name, type FROM elements
			WHERE workspace_id = $1 AND source_id = ANY($2)`,
			d.WorkspaceID, elemIDs)
		if err != nil {
			return nil, fmt.Errorf("resolve element names: %w", err)
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var sid, name, typ string
			if err := rows.Scan(&sid, &name, &typ); err != nil {
				return nil, err
			}
			elemMeta[sid] = [2]string{name, typ}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	// Resolve relationship metadata including semantic attributes.
	relIDs := make([]string, 0, len(layout.Connections))
	for _, c := range layout.Connections {
		if c.RelationshipID != "" {
			relIDs = append(relIDs, c.RelationshipID)
		}
	}

	type relMeta struct {
		typ        string
		source     string
		target     string
		accessType string
		isDirected bool
		modifier   string
	}
	relMap := map[string]relMeta{}
	if len(relIDs) > 0 {
		rows, err := s.db.Query(`
			SELECT source_id, type, source_element, target_element,
			       COALESCE(access_type, ''), is_directed, COALESCE(modifier, '')
			FROM relationships
			WHERE workspace_id = $1 AND source_id = ANY($2)`,
			d.WorkspaceID, relIDs)
		if err != nil {
			return nil, fmt.Errorf("resolve relationship metadata: %w", err)
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var sid, typ, src, tgt, accessType, modifier string
			var isDirected bool
			if err := rows.Scan(&sid, &typ, &src, &tgt, &accessType, &isDirected, &modifier); err != nil {
				return nil, err
			}
			relMap[sid] = relMeta{
				typ: typ, source: src, target: tgt,
				accessType: accessType, isDirected: isDirected, modifier: modifier,
			}
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	rd := &RenderData{
		ID:            d.ID,
		Name:          d.Name,
		Documentation: d.Documentation,
		Nodes:         make([]RenderNode, 0, len(layout.Nodes)),
		Connections:   make([]RenderConnection, 0, len(layout.Connections)),
	}

	for _, n := range layout.Nodes {
		// Skip Label nodes that have no displayable text (e.g. Archi labelExpression-only
		// nodes that reference view properties we cannot evaluate). Label nodes with a
		// fixed text label are kept so they render as text annotations.
		if n.NodeType == "Label" && n.Label == "" {
			continue
		}
		meta := elemMeta[n.ElementID]
		name := meta[0]
		typ := meta[1]
		// Group nodes have no element reference — use the layout-supplied label and type.
		if n.ElementType != "" {
			typ = n.ElementType
		}
		if n.Label != "" && name == "" {
			name = n.Label
		}
		rd.Nodes = append(rd.Nodes, RenderNode{
			NodeID:          n.NodeID,
			ElementID:       n.ElementID,
			ParentElementID: n.ParentElementID,
			NodeType:        n.NodeType,
			ElementName:     name,
			ElementType:     typ,
			X:               n.X,
			Y:               n.Y,
			W:               n.W,
			H:               n.H,
			Style:           n.Style,
		})
	}

	for _, c := range layout.Connections {
		meta := relMap[c.RelationshipID]
		bps := make([]Point, 0, len(c.Bendpoints))
		for _, bp := range c.Bendpoints {
			bps = append(bps, Point{X: bp.X, Y: bp.Y})
		}

		// Use the connection's visual source/target for path drawing (bendpoints follow this order).
		// Fall back to the relationship's direction for old layouts that lack this info.
		srcElem := c.SourceElementID
		tgtElem := c.TargetElementID
		if srcElem == "" {
			srcElem = meta.source
		}
		if tgtElem == "" {
			tgtElem = meta.target
		}

		// Detect whether the connection is drawn in the opposite direction to the relationship.
		// When reversed, the frontend must swap marker-start/end so the arrowhead stays at
		// the semantic target regardless of the visual drawing direction.
		reversed := srcElem != "" && meta.source != "" && srcElem == meta.target && tgtElem == meta.source

		rd.Connections = append(rd.Connections, RenderConnection{
			RelationshipID:   c.RelationshipID,
			RelationshipType: meta.typ,
			SourceNodeID:     c.SourceNodeID,
			TargetNodeID:     c.TargetNodeID,
			SourceElementID:  srcElem,
			TargetElementID:  tgtElem,
			Reversed:         reversed,
			Label:            c.Label,
			AccessType:       meta.accessType,
			IsDirected:       meta.isDirected,
			Modifier:         meta.modifier,
			Bendpoints:       bps,
			Style:            c.Style,
		})
	}

	return rd, nil
}

func (s *Store) Delete(id uuid.UUID) error {
	res, err := s.db.Exec(`DELETE FROM diagrams WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete diagram: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
