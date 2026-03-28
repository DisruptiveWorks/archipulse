package exporter

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/element"
	"github.com/DisruptiveWorks/archipulse/internal/parser"
	"github.com/DisruptiveWorks/archipulse/internal/relationship"
	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

// LoadModel builds a parser.Model from the database for the given workspace.
func LoadModel(db *sql.DB, workspaceID uuid.UUID) (*parser.Model, error) {
	wsStore := workspace.NewStore(db)
	ws, err := wsStore.Get(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("load workspace: %w", err)
	}

	elems, err := element.NewStore(db).List(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("load elements: %w", err)
	}

	rels, err := relationship.NewStore(db).List(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("load relationships: %w", err)
	}

	diagRows, err := loadDiagrams(db, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("load diagrams: %w", err)
	}

	m := &parser.Model{Name: ws.Name}

	for _, e := range elems {
		m.Elements = append(m.Elements, parser.Element{
			ID:            e.SourceID,
			Type:          e.Type,
			Name:          e.Name,
			Documentation: e.Documentation,
		})
	}

	for _, r := range rels {
		m.Relationships = append(m.Relationships, parser.Relationship{
			ID:            r.SourceID,
			Type:          r.Type,
			Source:        r.SourceElement,
			Target:        r.TargetElement,
			Name:          r.Name,
			Documentation: r.Documentation,
		})
	}

	m.Diagrams = diagRows
	return m, nil
}

// loadDiagrams deserializes the JSONB layout of each diagram into parser.Diagram.
func loadDiagrams(db *sql.DB, workspaceID uuid.UUID) ([]parser.Diagram, error) {
	rows, err := db.Query(`
		SELECT source_id, name, documentation, layout
		FROM diagrams WHERE workspace_id = $1 ORDER BY name`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []parser.Diagram
	for rows.Next() {
		var (
			sourceID, name, doc string
			rawLayout           []byte
		)
		if err := rows.Scan(&sourceID, &name, &doc, &rawLayout); err != nil {
			return nil, err
		}

		var layout struct {
			Nodes []struct {
				ElementRef string `json:"elementRef"`
				X, Y, W, H int
			} `json:"nodes"`
			Connections []struct {
				RelationshipRef string               `json:"relationshipRef"`
				Bendpoints      []struct{ X, Y int } `json:"bendpoints"`
			} `json:"connections"`
		}
		if err := json.Unmarshal(rawLayout, &layout); err != nil {
			return nil, fmt.Errorf("unmarshal layout for %s: %w", sourceID, err)
		}

		d := parser.Diagram{ID: sourceID, Name: name, Documentation: doc}
		for _, n := range layout.Nodes {
			d.Layout.Nodes = append(d.Layout.Nodes, parser.NodeLayout{
				ElementID: n.ElementRef, X: n.X, Y: n.Y, W: n.W, H: n.H,
			})
		}
		for _, c := range layout.Connections {
			cl := parser.ConnectionLayout{RelationshipID: c.RelationshipRef}
			for _, bp := range c.Bendpoints {
				cl.Bendpoints = append(cl.Bendpoints, parser.Point{X: bp.X, Y: bp.Y})
			}
			d.Layout.Connections = append(d.Layout.Connections, cl)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
