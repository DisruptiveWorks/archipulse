package exporter

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/DisruptiveWorks/archipulse/internal/element"
	"github.com/DisruptiveWorks/archipulse/internal/parser"
	"github.com/DisruptiveWorks/archipulse/internal/relationship"
	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

// loadFolders reads diagram_folders and diagram.folder_id from the DB and
// populates ViewFolders + DiagramFolders on the model.
func loadFolders(db *sql.DB, workspaceID uuid.UUID, m *parser.Model) error {
	// Load all folders for this workspace.
	rows, err := db.Query(`
		SELECT f.source_id, f.name, COALESCE(parent.source_id, '')
		FROM   diagram_folders f
		LEFT JOIN diagram_folders parent ON parent.id = f.parent_id
		WHERE  f.workspace_id = $1
		ORDER  BY f.position`, workspaceID)
	if err != nil {
		return fmt.Errorf("load folders: %w", err)
	}
	defer func() { _ = rows.Close() }()

	pos := 0
	for rows.Next() {
		var sourceID, name, parentSourceID string
		if err := rows.Scan(&sourceID, &name, &parentSourceID); err != nil {
			return err
		}
		m.ViewFolders = append(m.ViewFolders, parser.ViewFolder{
			SourceID: sourceID,
			Name:     name,
			ParentID: parentSourceID,
			Position: pos,
		})
		pos++
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Load diagram→folder assignments.
	drows, err := db.Query(`
		SELECT d.source_id, COALESCE(f.source_id, '')
		FROM   diagrams d
		LEFT JOIN diagram_folders f ON f.id = d.folder_id
		WHERE  d.workspace_id = $1`, workspaceID)
	if err != nil {
		return fmt.Errorf("load diagram folders: %w", err)
	}
	defer func() { _ = drows.Close() }()

	for drows.Next() {
		var diagSourceID, folderSourceID string
		if err := drows.Scan(&diagSourceID, &folderSourceID); err != nil {
			return err
		}
		m.DiagramFolders = append(m.DiagramFolders, parser.DiagramFolder{
			DiagramSourceID: diagSourceID,
			FolderSourceID:  folderSourceID,
		})
	}
	return drows.Err()
}

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

	// Load model identifier.
	var modelIdentifier string
	_ = db.QueryRow(`SELECT COALESCE(model_identifier, '') FROM workspaces WHERE id = $1`,
		workspaceID).Scan(&modelIdentifier)

	m := &parser.Model{Identifier: modelIdentifier, Name: ws.Name}

	// --- Property definitions ---
	pdRows, err := db.Query(`
		SELECT source_id, name, data_type
		FROM property_definitions WHERE workspace_id = $1 ORDER BY name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("load property definitions: %w", err)
	}
	defer func() { _ = pdRows.Close() }()
	for pdRows.Next() {
		var pd parser.PropertyDefinition
		if err := pdRows.Scan(&pd.ID, &pd.Name, &pd.DataType); err != nil {
			return nil, err
		}
		m.PropertyDefinitions = append(m.PropertyDefinitions, pd)
	}
	if err := pdRows.Err(); err != nil {
		return nil, err
	}

	// --- Model-level properties ---
	mpRows, err := db.Query(`
		SELECT definition_ref, key, value
		FROM model_properties WHERE workspace_id = $1 ORDER BY key`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("load model properties: %w", err)
	}
	defer func() { _ = mpRows.Close() }()
	for mpRows.Next() {
		var defRef, key, val string
		if err := mpRows.Scan(&defRef, &key, &val); err != nil {
			return nil, err
		}
		m.Properties = append(m.Properties, parser.Property{
			DefinitionRef: defRef, Key: key, Value: val,
		})
	}
	if err := mpRows.Err(); err != nil {
		return nil, err
	}

	// --- Viewpoints ---
	vpRows, err := db.Query(`
		SELECT source_id, name, documentation, purpose, content,
		       concerns, allowed_element_types, allowed_relationship_types, modeling_notes
		FROM viewpoints WHERE workspace_id = $1 ORDER BY name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("load viewpoints: %w", err)
	}
	defer func() { _ = vpRows.Close() }()
	for vpRows.Next() {
		var vp parser.Viewpoint
		var concernsJSON, notesJSON []byte
		var allowedElems, allowedRels pq.StringArray
		if err := vpRows.Scan(&vp.ID, &vp.Name, &vp.Documentation, &vp.Purpose, &vp.Content,
			&concernsJSON, &allowedElems, &allowedRels, &notesJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(concernsJSON, &vp.Concerns); err != nil {
			return nil, fmt.Errorf("unmarshal concerns for viewpoint %q: %w", vp.ID, err)
		}
		if err := json.Unmarshal(notesJSON, &vp.ModelingNotes); err != nil {
			return nil, fmt.Errorf("unmarshal modeling notes for viewpoint %q: %w", vp.ID, err)
		}
		vp.AllowedElementTypes = []string(allowedElems)
		vp.AllowedRelationshipTypes = []string(allowedRels)
		m.Viewpoints = append(m.Viewpoints, vp)
	}
	if err := vpRows.Err(); err != nil {
		return nil, err
	}

	// --- Elements ---
	for _, e := range elems {
		m.Elements = append(m.Elements, parser.Element{
			ID:            e.SourceID,
			Type:          e.Type,
			Name:          e.Name,
			Documentation: e.Documentation,
		})
	}

	// --- Element properties (bulk load for all elements in workspace) ---
	epRows, err := db.Query(`
		SELECT e.source_id, ep.definition_ref, ep.key, ep.value
		FROM element_properties ep
		JOIN elements e ON e.id = ep.element_id
		WHERE e.workspace_id = $1 AND ep.source = 'model'
		ORDER BY e.source_id, ep.key`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("load element properties: %w", err)
	}
	defer func() { _ = epRows.Close() }()
	// Build index: element source_id → slice position in m.Elements
	elemIdx := make(map[string]int, len(m.Elements))
	for i, e := range m.Elements {
		elemIdx[e.ID] = i
	}
	for epRows.Next() {
		var elemSourceID, defRef, key, val string
		if err := epRows.Scan(&elemSourceID, &defRef, &key, &val); err != nil {
			return nil, err
		}
		if i, ok := elemIdx[elemSourceID]; ok {
			m.Elements[i].Properties = append(m.Elements[i].Properties, parser.Property{
				DefinitionRef: defRef,
				Key:           key,
				Value:         val,
			})
		}
	}
	if err := epRows.Err(); err != nil {
		return nil, err
	}

	// --- Relationships ---
	for _, r := range rels {
		m.Relationships = append(m.Relationships, parser.Relationship{
			ID:            r.SourceID,
			Type:          r.Type,
			Source:        r.SourceElement,
			Target:        r.TargetElement,
			Name:          r.Name,
			Documentation: r.Documentation,
			AccessType:    r.AccessType,
			IsDirected:    r.IsDirected,
			Modifier:      r.Modifier,
		})
	}

	// Build relationship index for subsequent queries.
	relIdx := make(map[string]int, len(m.Relationships))
	for i, r := range m.Relationships {
		relIdx[r.ID] = i
	}

	// --- Element lang variants ---
	enRows, err := db.Query(`
		SELECT e.source_id, en.field, en.lang, en.value
		FROM element_names en
		JOIN elements e ON e.id = en.element_id
		WHERE e.workspace_id = $1
		ORDER BY e.source_id, en.field, en.lang`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("load element names: %w", err)
	}
	defer func() { _ = enRows.Close() }()
	for enRows.Next() {
		var srcID, field, lang, value string
		if err := enRows.Scan(&srcID, &field, &lang, &value); err != nil {
			return nil, err
		}
		if i, ok := elemIdx[srcID]; ok {
			ls := parser.LangString{Lang: lang, Value: value}
			if field == "name" {
				m.Elements[i].Names = append(m.Elements[i].Names, ls)
			} else {
				m.Elements[i].Documentations = append(m.Elements[i].Documentations, ls)
			}
		}
	}
	if err := enRows.Err(); err != nil {
		return nil, err
	}

	// --- Relationship lang variants ---
	rnRows, err := db.Query(`
		SELECT r.source_id, rn.field, rn.lang, rn.value
		FROM relationship_names rn
		JOIN relationships r ON r.id = rn.relationship_id
		WHERE r.workspace_id = $1
		ORDER BY r.source_id, rn.field, rn.lang`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("load relationship names: %w", err)
	}
	defer func() { _ = rnRows.Close() }()
	for rnRows.Next() {
		var srcID, field, lang, value string
		if err := rnRows.Scan(&srcID, &field, &lang, &value); err != nil {
			return nil, err
		}
		if i, ok := relIdx[srcID]; ok {
			ls := parser.LangString{Lang: lang, Value: value}
			if field == "name" {
				m.Relationships[i].Names = append(m.Relationships[i].Names, ls)
			} else {
				m.Relationships[i].Documentations = append(m.Relationships[i].Documentations, ls)
			}
		}
	}
	if err := rnRows.Err(); err != nil {
		return nil, err
	}

	// --- Relationship properties (bulk load) ---
	rpRows, err := db.Query(`
		SELECT r.source_id, rp.definition_ref, rp.key, rp.value
		FROM relationship_properties rp
		JOIN relationships r ON r.id = rp.relationship_id
		WHERE r.workspace_id = $1 AND rp.source = 'model'
		ORDER BY r.source_id, rp.key`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("load relationship properties: %w", err)
	}
	defer func() { _ = rpRows.Close() }()
	for rpRows.Next() {
		var relSourceID, defRef, key, val string
		if err := rpRows.Scan(&relSourceID, &defRef, &key, &val); err != nil {
			return nil, err
		}
		if i, ok := relIdx[relSourceID]; ok {
			m.Relationships[i].Properties = append(m.Relationships[i].Properties, parser.Property{
				DefinitionRef: defRef,
				Key:           key,
				Value:         val,
			})
		}
	}
	if err := rpRows.Err(); err != nil {
		return nil, err
	}

	// --- Diagrams ---
	m.Diagrams = diagRows

	// Build diagram index: source_id → slice position.
	diagIdx := make(map[string]int, len(m.Diagrams))
	for i, d := range m.Diagrams {
		diagIdx[d.ID] = i
	}

	// --- Diagram lang variants ---
	dnRows, err := db.Query(`
		SELECT d.source_id, dn.field, dn.lang, dn.value
		FROM diagram_names dn
		JOIN diagrams d ON d.id = dn.diagram_id
		WHERE d.workspace_id = $1
		ORDER BY d.source_id, dn.field, dn.lang`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("load diagram names: %w", err)
	}
	defer func() { _ = dnRows.Close() }()
	for dnRows.Next() {
		var srcID, field, lang, value string
		if err := dnRows.Scan(&srcID, &field, &lang, &value); err != nil {
			return nil, err
		}
		if i, ok := diagIdx[srcID]; ok {
			ls := parser.LangString{Lang: lang, Value: value}
			if field == "name" {
				m.Diagrams[i].Names = append(m.Diagrams[i].Names, ls)
			} else {
				m.Diagrams[i].Documentations = append(m.Diagrams[i].Documentations, ls)
			}
		}
	}
	if err := dnRows.Err(); err != nil {
		return nil, err
	}

	// --- View properties ---
	vpropRows, err := db.Query(`
		SELECT d.source_id, vp.definition_ref, vp.key, vp.value
		FROM view_properties vp
		JOIN diagrams d ON d.id = vp.diagram_id
		WHERE d.workspace_id = $1
		ORDER BY d.source_id, vp.key`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("load view properties: %w", err)
	}
	defer func() { _ = vpropRows.Close() }()
	for vpropRows.Next() {
		var diagSourceID, defRef, key, val string
		if err := vpropRows.Scan(&diagSourceID, &defRef, &key, &val); err != nil {
			return nil, err
		}
		if i, ok := diagIdx[diagSourceID]; ok {
			m.Diagrams[i].Properties = append(m.Diagrams[i].Properties, parser.Property{
				DefinitionRef: defRef,
				Key:           key,
				Value:         val,
			})
		}
	}
	if err := vpropRows.Err(); err != nil {
		return nil, err
	}

	if err := loadFolders(db, workspaceID, m); err != nil {
		return nil, fmt.Errorf("load folders: %w", err)
	}

	return m, nil
}

// loadDiagrams deserializes the JSONB layout of each diagram into parser.Diagram.
func loadDiagrams(db *sql.DB, workspaceID uuid.UUID) ([]parser.Diagram, error) {
	rows, err := db.Query(`
		SELECT source_id, name, documentation, layout,
		       COALESCE(viewpoint, ''), COALESCE(viewpoint_ref, '')
		FROM diagrams WHERE workspace_id = $1 ORDER BY name`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []parser.Diagram
	for rows.Next() {
		var (
			sourceID, name, doc     string
			viewpoint, viewpointRef string
			rawLayout               []byte
		)
		if err := rows.Scan(&sourceID, &name, &doc, &rawLayout, &viewpoint, &viewpointRef); err != nil {
			return nil, err
		}

		// JSON tags must match what json.Marshal produces for parser.RGBColor /
		// parser.FontStyle / parser.NodeStyle / parser.ConnStyle — those structs have
		// no explicit json tags, so Go uses the field name as-is (PascalCase /
		// single-letter). Using snake_case or lowercase tags here would silently
		// deserialize every style as nil.
		type dbColor struct {
			R int  `json:"R"`
			G int  `json:"G"`
			B int  `json:"B"`
			A *int `json:"A,omitempty"`
		}
		type dbFont struct {
			Name  string   `json:"Name,omitempty"`
			Size  string   `json:"Size,omitempty"`
			Style string   `json:"Style,omitempty"`
			Color *dbColor `json:"Color,omitempty"`
		}
		type dbNodeStyle struct {
			FillColor *dbColor `json:"FillColor,omitempty"`
			LineColor *dbColor `json:"LineColor,omitempty"`
			Font      *dbFont  `json:"Font,omitempty"`
			LineWidth int      `json:"LineWidth,omitempty"`
		}
		type dbConnStyle struct {
			LineColor *dbColor `json:"LineColor,omitempty"`
			Font      *dbFont  `json:"Font,omitempty"`
			LineWidth int      `json:"LineWidth,omitempty"`
		}
		var layout struct {
			Nodes []struct {
				NodeID          string       `json:"NodeID"`
				ElementID       string       `json:"ElementID"`
				ParentElementID string       `json:"ParentElementID"`
				NodeType        string       `json:"NodeType"`
				Label           string       `json:"Label"`
				LabelExpression string       `json:"LabelExpression"`
				ElementType     string       `json:"ElementType"`
				X               int          `json:"X"`
				Y               int          `json:"Y"`
				W               int          `json:"W"`
				H               int          `json:"H"`
				Style           *dbNodeStyle `json:"Style"`
			} `json:"Nodes"`
			Connections []struct {
				ConnectionID    string `json:"ConnectionID"`
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
				Style *dbConnStyle `json:"Style"`
			} `json:"Connections"`
		}
		if err := json.Unmarshal(rawLayout, &layout); err != nil {
			return nil, fmt.Errorf("unmarshal layout for %s: %w", sourceID, err)
		}

		d := parser.Diagram{ID: sourceID, Name: name, Documentation: doc, Viewpoint: viewpoint, ViewpointRef: viewpointRef}
		for _, n := range layout.Nodes {
			nl := parser.NodeLayout{
				NodeID:          n.NodeID,
				ElementID:       n.ElementID,
				ParentElementID: n.ParentElementID,
				NodeType:        n.NodeType,
				Label:           n.Label,
				LabelExpression: n.LabelExpression,
				ElementType:     n.ElementType,
				X:               n.X,
				Y:               n.Y,
				W:               n.W,
				H:               n.H,
			}
			if s := n.Style; s != nil {
				ns := &parser.NodeStyle{LineWidth: s.LineWidth}
				if s.FillColor != nil {
					ns.FillColor = &parser.RGBColor{R: s.FillColor.R, G: s.FillColor.G, B: s.FillColor.B, A: s.FillColor.A}
				}
				if s.LineColor != nil {
					ns.LineColor = &parser.RGBColor{R: s.LineColor.R, G: s.LineColor.G, B: s.LineColor.B, A: s.LineColor.A}
				}
				if s.Font != nil {
					ns.Font = &parser.FontStyle{Name: s.Font.Name, Size: s.Font.Size, Style: s.Font.Style}
					if s.Font.Color != nil {
						ns.Font.Color = &parser.RGBColor{R: s.Font.Color.R, G: s.Font.Color.G, B: s.Font.Color.B, A: s.Font.Color.A}
					}
				}
				nl.Style = ns
			}
			d.Layout.Nodes = append(d.Layout.Nodes, nl)
		}
		for _, c := range layout.Connections {
			cl := parser.ConnectionLayout{
				ConnectionID:    c.ConnectionID,
				RelationshipID:  c.RelationshipID,
				SourceNodeID:    c.SourceNodeID,
				TargetNodeID:    c.TargetNodeID,
				SourceElementID: c.SourceElementID,
				TargetElementID: c.TargetElementID,
				Label:           c.Label,
			}
			for _, bp := range c.Bendpoints {
				cl.Bendpoints = append(cl.Bendpoints, parser.Point{X: bp.X, Y: bp.Y})
			}
			if s := c.Style; s != nil {
				cs := &parser.ConnStyle{LineWidth: s.LineWidth}
				if s.LineColor != nil {
					cs.LineColor = &parser.RGBColor{R: s.LineColor.R, G: s.LineColor.G, B: s.LineColor.B, A: s.LineColor.A}
				}
				if s.Font != nil {
					cs.Font = &parser.FontStyle{Name: s.Font.Name, Size: s.Font.Size, Style: s.Font.Style}
					if s.Font.Color != nil {
						cs.Font.Color = &parser.RGBColor{R: s.Font.Color.R, G: s.Font.Color.G, B: s.Font.Color.B, A: s.Font.Color.A}
					}
				}
				cl.Style = cs
			}
			d.Layout.Connections = append(d.Layout.Connections, cl)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
