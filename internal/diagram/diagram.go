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

// RenderNode is a node enriched with element metadata for rendering.
type RenderNode struct {
	ElementID   string `json:"element_id"`
	ElementName string `json:"element_name"`
	ElementType string `json:"element_type"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	W           int    `json:"w"`
	H           int    `json:"h"`
}

// RenderConnection is a connection enriched with relationship metadata.
type RenderConnection struct {
	RelationshipID   string  `json:"relationship_id"`
	RelationshipType string  `json:"relationship_type"`
	SourceElementID  string  `json:"source_element_id"`
	TargetElementID  string  `json:"target_element_id"`
	Bendpoints       []Point `json:"bendpoints"`
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
			ElementID string `json:"ElementID"`
			X         int    `json:"X"`
			Y         int    `json:"Y"`
			W         int    `json:"W"`
			H         int    `json:"H"`
		} `json:"Nodes"`
		Connections []struct {
			RelationshipID string `json:"RelationshipID"`
			Bendpoints     []struct {
				X int `json:"X"`
				Y int `json:"Y"`
			} `json:"Bendpoints"`
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

	// Resolve relationship metadata.
	relIDs := make([]string, 0, len(layout.Connections))
	for _, c := range layout.Connections {
		if c.RelationshipID != "" {
			relIDs = append(relIDs, c.RelationshipID)
		}
	}

	type relMeta struct {
		typ    string
		source string
		target string
	}
	relMap := map[string]relMeta{}
	if len(relIDs) > 0 {
		rows, err := s.db.Query(`
			SELECT source_id, type, source_element, target_element FROM relationships
			WHERE workspace_id = $1 AND source_id = ANY($2)`,
			d.WorkspaceID, relIDs)
		if err != nil {
			return nil, fmt.Errorf("resolve relationship metadata: %w", err)
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var sid, typ, src, tgt string
			if err := rows.Scan(&sid, &typ, &src, &tgt); err != nil {
				return nil, err
			}
			relMap[sid] = relMeta{typ: typ, source: src, target: tgt}
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
		meta := elemMeta[n.ElementID]
		rd.Nodes = append(rd.Nodes, RenderNode{
			ElementID:   n.ElementID,
			ElementName: meta[0],
			ElementType: meta[1],
			X:           n.X,
			Y:           n.Y,
			W:           n.W,
			H:           n.H,
		})
	}

	for _, c := range layout.Connections {
		meta := relMap[c.RelationshipID]
		bps := make([]Point, 0, len(c.Bendpoints))
		for _, bp := range c.Bendpoints {
			bps = append(bps, Point{X: bp.X, Y: bp.Y})
		}
		rd.Connections = append(rd.Connections, RenderConnection{
			RelationshipID:   c.RelationshipID,
			RelationshipType: meta.typ,
			SourceElementID:  meta.source,
			TargetElementID:  meta.target,
			Bendpoints:       bps,
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
