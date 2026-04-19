package api

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/parser"
	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

// PreviewItem describes a single added or modified entity.
type PreviewItem struct {
	SourceID string `json:"source_id"`
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`
}

// PreviewCategory summarises changes for one category of entities.
type PreviewCategory struct {
	Added     int           `json:"added"`
	Modified  int           `json:"modified"`
	Unchanged int           `json:"unchanged"`
	Details   []PreviewItem `json:"details"` // only added + modified items
}

// ImportPreview is the full diff returned by the preview endpoint.
type ImportPreview struct {
	Elements            PreviewCategory `json:"elements"`
	Relationships       PreviewCategory `json:"relationships"`
	Diagrams            PreviewCategory `json:"diagrams"`
	PropertyDefinitions PreviewCategory `json:"property_definitions"`
}

func (h *importHandler) previewImport(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	if _, err := workspace.NewStore(h.db).Get(wsID); err != nil {
		if isNotFound(err) {
			respondError(w, http.StatusNotFound, err)
		} else {
			respondError(w, http.StatusInternalServerError, err)
		}
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		respondError(w, http.StatusBadRequest, errorf("parse multipart form: %v", err))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, errorf("file field required"))
		return
	}
	defer func() { _ = file.Close() }()

	name := strings.ToLower(header.Filename)
	var m *parser.Model
	switch {
	case strings.HasSuffix(name, ".xml"):
		m, err = parser.ParseAOEF(file)
	default:
		respondError(w, http.StatusBadRequest, errorf("unsupported file format: use .xml (AOEF / Open Exchange Format)"))
		return
	}
	if err != nil {
		respondError(w, http.StatusUnprocessableEntity, errorf("parse error: %v", err))
		return
	}

	preview, err := buildImportPreview(h.db, wsID, m)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, http.StatusOK, preview)
}

// buildImportPreview compares the incoming model against the current workspace
// state and returns a diff without writing anything to the database.
func buildImportPreview(db *sql.DB, wsID uuid.UUID, m *parser.Model) (*ImportPreview, error) {
	preview := &ImportPreview{}
	var err error

	preview.Elements, err = diffElements(db, wsID, m)
	if err != nil {
		return nil, err
	}
	preview.Relationships, err = diffRelationships(db, wsID, m)
	if err != nil {
		return nil, err
	}
	preview.Diagrams, err = diffDiagrams(db, wsID, m)
	if err != nil {
		return nil, err
	}
	preview.PropertyDefinitions, err = diffPropertyDefinitions(db, wsID, m)
	if err != nil {
		return nil, err
	}
	return preview, nil
}

// diffElements compares incoming elements against the workspace DB.
func diffElements(db *sql.DB, wsID uuid.UUID, m *parser.Model) (PreviewCategory, error) {
	rows, err := db.Query(`
		SELECT e.source_id, COALESCE(n.value, e.name, ''), e.type
		FROM elements e
		LEFT JOIN element_names n ON n.element_id = e.id AND n.field = 'name'
		WHERE e.workspace_id = $1`, wsID)
	if err != nil {
		return PreviewCategory{}, err
	}
	defer func() { _ = rows.Close() }()

	type existing struct{ name, typ string }
	current := make(map[string]existing)
	for rows.Next() {
		var sid, name, typ string
		if err := rows.Scan(&sid, &name, &typ); err != nil {
			return PreviewCategory{}, err
		}
		current[sid] = existing{name, typ}
	}
	if err := rows.Err(); err != nil {
		return PreviewCategory{}, err
	}

	var cat PreviewCategory
	for _, el := range m.Elements {
		ex, exists := current[el.ID]
		displayName := el.Name
		if displayName == "" {
			displayName = firstLangValue(el.Names)
		}
		if !exists {
			cat.Added++
			cat.Details = append(cat.Details, PreviewItem{
				SourceID: el.ID,
				Name:     displayName,
				Type:     el.Type,
			})
		} else if ex.typ != el.Type || nameChanged(ex.name, displayName) {
			cat.Modified++
			cat.Details = append(cat.Details, PreviewItem{
				SourceID: el.ID,
				Name:     displayName,
				Type:     el.Type,
			})
		} else {
			cat.Unchanged++
		}
	}
	return cat, nil
}

// diffRelationships compares incoming relationships against the workspace DB.
func diffRelationships(db *sql.DB, wsID uuid.UUID, m *parser.Model) (PreviewCategory, error) {
	rows, err := db.Query(`
		SELECT source_id, type, source_element, target_element
		FROM relationships
		WHERE workspace_id = $1`, wsID)
	if err != nil {
		return PreviewCategory{}, err
	}
	defer func() { _ = rows.Close() }()

	type existing struct{ typ, src, tgt string }
	current := make(map[string]existing)
	for rows.Next() {
		var sid, typ, src, tgt string
		if err := rows.Scan(&sid, &typ, &src, &tgt); err != nil {
			return PreviewCategory{}, err
		}
		current[sid] = existing{typ, src, tgt}
	}
	if err := rows.Err(); err != nil {
		return PreviewCategory{}, err
	}

	var cat PreviewCategory
	for _, rel := range m.Relationships {
		ex, exists := current[rel.ID]
		displayName := rel.Name
		if displayName == "" {
			displayName = firstLangValue(rel.Names)
		}
		if !exists {
			cat.Added++
			cat.Details = append(cat.Details, PreviewItem{
				SourceID: rel.ID,
				Name:     displayName,
				Type:     rel.Type,
			})
		} else if ex.typ != rel.Type || ex.src != rel.Source || ex.tgt != rel.Target {
			cat.Modified++
			cat.Details = append(cat.Details, PreviewItem{
				SourceID: rel.ID,
				Name:     displayName,
				Type:     rel.Type,
			})
		} else {
			cat.Unchanged++
		}
	}
	return cat, nil
}

// diffDiagrams compares incoming diagrams against the workspace DB.
func diffDiagrams(db *sql.DB, wsID uuid.UUID, m *parser.Model) (PreviewCategory, error) {
	rows, err := db.Query(`
		SELECT d.source_id, COALESCE(n.value, d.name, '')
		FROM diagrams d
		LEFT JOIN diagram_names n ON n.diagram_id = d.id AND n.field = 'name'
		WHERE d.workspace_id = $1`, wsID)
	if err != nil {
		return PreviewCategory{}, err
	}
	defer func() { _ = rows.Close() }()

	current := make(map[string]string)
	for rows.Next() {
		var sid, name string
		if err := rows.Scan(&sid, &name); err != nil {
			return PreviewCategory{}, err
		}
		current[sid] = name
	}
	if err := rows.Err(); err != nil {
		return PreviewCategory{}, err
	}

	var cat PreviewCategory
	for _, d := range m.Diagrams {
		existingName, exists := current[d.ID]
		displayName := d.Name
		if displayName == "" {
			displayName = firstLangValue(d.Names)
		}
		if !exists {
			cat.Added++
			cat.Details = append(cat.Details, PreviewItem{
				SourceID: d.ID,
				Name:     displayName,
			})
		} else if nameChanged(existingName, displayName) {
			cat.Modified++
			cat.Details = append(cat.Details, PreviewItem{
				SourceID: d.ID,
				Name:     displayName,
			})
		} else {
			cat.Unchanged++
		}
	}
	return cat, nil
}

// diffPropertyDefinitions compares incoming property definitions against the workspace DB.
func diffPropertyDefinitions(db *sql.DB, wsID uuid.UUID, m *parser.Model) (PreviewCategory, error) {
	rows, err := db.Query(`
		SELECT source_id, name, data_type FROM property_definitions WHERE workspace_id = $1`, wsID)
	if err != nil {
		return PreviewCategory{}, err
	}
	defer func() { _ = rows.Close() }()

	type existing struct{ name, typ string }
	current := make(map[string]existing)
	for rows.Next() {
		var sid, name, typ string
		if err := rows.Scan(&sid, &name, &typ); err != nil {
			return PreviewCategory{}, err
		}
		current[sid] = existing{name, typ}
	}
	if err := rows.Err(); err != nil {
		return PreviewCategory{}, err
	}

	var cat PreviewCategory
	for _, pd := range m.PropertyDefinitions {
		ex, exists := current[pd.ID]
		if !exists {
			cat.Added++
			cat.Details = append(cat.Details, PreviewItem{
				SourceID: pd.ID,
				Name:     pd.Name,
				Type:     pd.DataType,
			})
		} else if ex.name != pd.Name || ex.typ != pd.DataType {
			cat.Modified++
			cat.Details = append(cat.Details, PreviewItem{
				SourceID: pd.ID,
				Name:     pd.Name,
				Type:     pd.DataType,
			})
		} else {
			cat.Unchanged++
		}
	}
	return cat, nil
}

// firstLangValue returns the first non-empty value from a []parser.LangString.
func firstLangValue(langs []parser.LangString) string {
	for _, l := range langs {
		if l.Value != "" {
			return l.Value
		}
	}
	return ""
}

// nameChanged returns true if the names differ, ignoring empty values.
func nameChanged(existing, incoming string) bool {
	if incoming == "" {
		return false
	}
	return existing != incoming
}

// isNotFound checks for workspace not-found errors.
func isNotFound(err error) bool {
	return err != nil && err.Error() == "workspace not found"
}
