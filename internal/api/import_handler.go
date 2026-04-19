package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/DisruptiveWorks/archipulse/internal/audit"
	"github.com/DisruptiveWorks/archipulse/internal/auth"
	"github.com/DisruptiveWorks/archipulse/internal/element"
	"github.com/DisruptiveWorks/archipulse/internal/parser"
	"github.com/DisruptiveWorks/archipulse/internal/snapshot"
	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

const maxUploadSize = 32 << 20 // 32 MB

type importHandler struct {
	db    *sql.DB
	audit *audit.Store
	snaps *snapshot.Store
}

// ImportResult summarises what was imported.
type ImportResult struct {
	WorkspaceID         string `json:"workspace_id"`
	Elements            int    `json:"elements"`
	Relationships       int    `json:"relationships"`
	Diagrams            int    `json:"diagrams"`
	Folders             int    `json:"folders"`
	PropertyDefinitions int    `json:"property_definitions"`
}

func (h *importHandler) importModel(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	// Verify workspace exists.
	if _, err := workspace.NewStore(h.db).Get(wsID); errors.Is(err, workspace.ErrNotFound) {
		respondError(w, http.StatusNotFound, err)
		return
	} else if err != nil {
		respondError(w, http.StatusInternalServerError, err)
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

	// Detect format from filename extension.
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

	// Semantic validation.
	if err := parser.Validate(m); err != nil {
		var ve *parser.ValidationError
		if errors.As(err, &ve) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":  "model validation failed",
				"issues": ve.Issues,
			})
			return
		}
		respondError(w, http.StatusUnprocessableEntity, err)
		return
	}

	claims := auth.ClaimsFromCtx(r.Context())

	// Auto-snapshot current state before overwriting.
	if h.snaps != nil {
		_, _ = takeSnapshot(h.db, h.snaps, wsID, claims.UserID, claims.Email, "", "import")
	}

	// Import inside a transaction — all or nothing.
	result, err := importInTx(h.db, wsID, m)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	// Record audit event.
	if h.audit != nil {
		_ = h.audit.Record(audit.RecordParams{
			WorkspaceID: wsID,
			UserID:      claims.UserID,
			UserEmail:   claims.Email,
			Action:      audit.ActionImport,
			EntityType:  audit.EntityWorkspace,
			EntityName:  header.Filename,
			Meta: map[string]any{
				"elements":      result.Elements,
				"relationships": result.Relationships,
			},
		})
	}

	respondJSON(w, http.StatusOK, result)
}

// insertLangStrings bulk-inserts xml:lang variants for a given entity and field.
// table must be one of element_names, relationship_names, diagram_names.
// idCol is the FK column name (element_id, relationship_id, diagram_id).
func insertLangStrings(tx *sql.Tx, table, idCol string, entityID interface{}, field string, langs []parser.LangString) error {
	for _, ls := range langs {
		if _, err := tx.Exec(`INSERT INTO `+table+` (`+idCol+`, field, lang, value)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (`+idCol+`, field, lang) DO UPDATE SET value = EXCLUDED.value`,
			entityID, field, ls.Lang, ls.Value); err != nil {
			return err
		}
	}
	return nil
}

func importInTx(db *sql.DB, wsID uuid.UUID, m *parser.Model) (*ImportResult, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result := &ImportResult{WorkspaceID: wsID.String()}

	// --- Model identifier ---
	if m.Identifier != "" {
		if _, err := tx.Exec(`UPDATE workspaces SET model_identifier = $1 WHERE id = $2`,
			m.Identifier, wsID); err != nil {
			return nil, errorf("store model identifier: %w", err)
		}
	}

	// --- Model-level properties ---
	if _, err := tx.Exec(`DELETE FROM model_properties WHERE workspace_id = $1`, wsID); err != nil {
		return nil, errorf("clear model properties: %w", err)
	}
	for _, p := range m.Properties {
		if _, err := tx.Exec(`
			INSERT INTO model_properties (workspace_id, definition_ref, key, value)
			VALUES ($1, $2, $3, $4)`,
			wsID, p.DefinitionRef, p.Key, p.Value); err != nil {
			return nil, errorf("insert model property %q: %w", p.Key, err)
		}
	}

	// --- Viewpoints ---
	// Delete viewpoints no longer present (re-import is authoritative).
	if _, err := tx.Exec(`DELETE FROM viewpoints WHERE workspace_id = $1`, wsID); err != nil {
		return nil, errorf("clear viewpoints: %w", err)
	}
	for _, vp := range m.Viewpoints {
		concernsJSON, err := json.Marshal(vp.Concerns)
		if err != nil {
			return nil, errorf("marshal concerns for viewpoint %q: %w", vp.ID, err)
		}
		notesJSON, err := json.Marshal(vp.ModelingNotes)
		if err != nil {
			return nil, errorf("marshal modeling notes for viewpoint %q: %w", vp.ID, err)
		}
		allowedElems := vp.AllowedElementTypes
		if allowedElems == nil {
			allowedElems = []string{}
		}
		allowedRels := vp.AllowedRelationshipTypes
		if allowedRels == nil {
			allowedRels = []string{}
		}
		if _, err := tx.Exec(`
			INSERT INTO viewpoints
			  (workspace_id, source_id, name, documentation, purpose, content,
			   concerns, allowed_element_types, allowed_relationship_types, modeling_notes)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (workspace_id, source_id) DO UPDATE
			  SET name = EXCLUDED.name,
			      documentation = EXCLUDED.documentation,
			      purpose = EXCLUDED.purpose,
			      content = EXCLUDED.content,
			      concerns = EXCLUDED.concerns,
			      allowed_element_types = EXCLUDED.allowed_element_types,
			      allowed_relationship_types = EXCLUDED.allowed_relationship_types,
			      modeling_notes = EXCLUDED.modeling_notes`,
			wsID, vp.ID, vp.Name, vp.Documentation, vp.Purpose, vp.Content,
			concernsJSON, pq.Array(allowedElems), pq.Array(allowedRels), notesJSON); err != nil {
			return nil, errorf("upsert viewpoint %q: %w", vp.ID, err)
		}
	}

	// --- Property definitions ---
	for _, pd := range m.PropertyDefinitions {
		dt := pd.DataType
		if dt == "" {
			dt = "string"
		}
		_, err := tx.Exec(`
			INSERT INTO property_definitions (workspace_id, source_id, name, data_type)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (workspace_id, source_id) DO UPDATE
			  SET name = EXCLUDED.name, data_type = EXCLUDED.data_type`,
			wsID, pd.ID, pd.Name, dt)
		if err != nil {
			return nil, errorf("upsert property definition %q: %w", pd.ID, err)
		}
		result.PropertyDefinitions++
	}

	// --- Elements ---
	for _, e := range m.Elements {
		layer := parser.ElementLayer(e.Type)
		var elemID uuid.UUID
		err := tx.QueryRow(`
			INSERT INTO elements (workspace_id, source_id, type, layer, name, documentation)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (workspace_id, source_id) DO UPDATE
			  SET type = EXCLUDED.type, layer = EXCLUDED.layer,
			      name = EXCLUDED.name, documentation = EXCLUDED.documentation,
			      version = elements.version + 1, updated_at = now()
			RETURNING id`,
			wsID, e.ID, e.Type, layer, e.Name, e.Documentation).Scan(&elemID)
		if err != nil {
			return nil, errorf("upsert element %q: %w", e.ID, err)
		}
		result.Elements++

		// Lang variants for name and documentation — delete-then-insert for clean re-import.
		if _, err := tx.Exec(`DELETE FROM element_names WHERE element_id = $1`, elemID); err != nil {
			return nil, errorf("clear element names for %q: %w", e.ID, err)
		}
		if err := insertLangStrings(tx, "element_names", "element_id", elemID, "name", e.Names); err != nil {
			return nil, errorf("insert element names for %q: %w", e.ID, err)
		}
		if err := insertLangStrings(tx, "element_names", "element_id", elemID, "documentation", e.Documentations); err != nil {
			return nil, errorf("insert element docs for %q: %w", e.ID, err)
		}

		// Always clear then re-insert model properties so a re-import stays clean.
		if _, err := tx.Exec(`DELETE FROM element_properties WHERE element_id = $1 AND source = 'model'`, elemID); err != nil {
			return nil, errorf("clear model properties for element %q: %w", e.ID, err)
		}
		if len(e.Properties) > 0 {
			props := make([]element.ModelProperty, len(e.Properties))
			for i, p := range e.Properties {
				props[i] = element.ModelProperty{DefinitionRef: p.DefinitionRef, Key: p.Key, Value: p.Value}
			}
			if err := element.InsertProperties(tx, elemID, props, "model", nil); err != nil {
				return nil, err
			}
		}
	}

	// --- Relationships ---
	for _, r := range m.Relationships {
		// Nullable columns: access_type and modifier are empty strings when not applicable.
		var accessType, modifier *string
		if r.AccessType != "" {
			accessType = &r.AccessType
		}
		if r.Modifier != "" {
			modifier = &r.Modifier
		}
		var relID uuid.UUID
		err := tx.QueryRow(`
			INSERT INTO relationships
			  (workspace_id, source_id, type, source_element, target_element,
			   name, documentation, access_type, is_directed, modifier)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (workspace_id, source_id) DO UPDATE
			  SET type = EXCLUDED.type,
			      source_element = EXCLUDED.source_element,
			      target_element = EXCLUDED.target_element,
			      name = EXCLUDED.name,
			      documentation = EXCLUDED.documentation,
			      access_type = EXCLUDED.access_type,
			      is_directed = EXCLUDED.is_directed,
			      modifier = EXCLUDED.modifier,
			      version = relationships.version + 1, updated_at = now()
			RETURNING id`,
			wsID, r.ID, r.Type, r.Source, r.Target,
			r.Name, r.Documentation, accessType, r.IsDirected, modifier).Scan(&relID)
		if err != nil {
			return nil, errorf("upsert relationship %q: %w", r.ID, err)
		}
		result.Relationships++

		// Lang variants — delete-then-insert for clean re-import.
		if _, err := tx.Exec(`DELETE FROM relationship_names WHERE relationship_id = $1`, relID); err != nil {
			return nil, errorf("clear relationship names for %q: %w", r.ID, err)
		}
		if err := insertLangStrings(tx, "relationship_names", "relationship_id", relID, "name", r.Names); err != nil {
			return nil, errorf("insert relationship names for %q: %w", r.ID, err)
		}
		if err := insertLangStrings(tx, "relationship_names", "relationship_id", relID, "documentation", r.Documentations); err != nil {
			return nil, errorf("insert relationship docs for %q: %w", r.ID, err)
		}

		// Always clear then re-insert model properties so a re-import stays clean.
		if _, err := tx.Exec(`DELETE FROM relationship_properties WHERE relationship_id = $1 AND source = 'model'`, relID); err != nil {
			return nil, errorf("clear model properties for relationship %q: %w", r.ID, err)
		}
		for _, p := range r.Properties {
			if _, err := tx.Exec(`
				INSERT INTO relationship_properties (relationship_id, definition_ref, key, value, source)
				VALUES ($1, $2, $3, $4, 'model')`,
				relID, p.DefinitionRef, p.Key, p.Value); err != nil {
				return nil, errorf("insert property %q for relationship %q: %w", p.Key, r.ID, err)
			}
		}
	}

	// --- Diagram folders (parser returns them parent-first) ---
	// Build a map from source_id → DB UUID for assigning folder_id to diagrams.
	folderUUIDs := make(map[string]uuid.UUID, len(m.ViewFolders))
	for _, f := range m.ViewFolders {
		var parentID *uuid.UUID
		if f.ParentID != "" {
			if pid, ok := folderUUIDs[f.ParentID]; ok {
				parentID = &pid
			}
		}
		var id uuid.UUID
		err := tx.QueryRow(`
			INSERT INTO diagram_folders (workspace_id, parent_id, name, source_id, position)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (workspace_id, source_id) DO UPDATE
			  SET parent_id = EXCLUDED.parent_id,
			      name      = EXCLUDED.name,
			      position  = EXCLUDED.position
			RETURNING id`,
			wsID, parentID, f.Name, f.SourceID, f.Position,
		).Scan(&id)
		if err != nil {
			return nil, errorf("upsert folder %q: %w", f.SourceID, err)
		}
		folderUUIDs[f.SourceID] = id
		result.Folders++
	}

	// Build diagram source_id → folder_id lookup from DiagramFolders.
	diagFolderID := make(map[string]*uuid.UUID, len(m.DiagramFolders))
	for _, df := range m.DiagramFolders {
		if df.FolderSourceID == "" {
			diagFolderID[df.DiagramSourceID] = nil
			continue
		}
		if fid, ok := folderUUIDs[df.FolderSourceID]; ok {
			id := fid
			diagFolderID[df.DiagramSourceID] = &id
		}
	}

	// --- Diagrams ---
	for _, d := range m.Diagrams {
		layoutJSON, err := json.Marshal(d.Layout)
		if err != nil {
			return nil, errorf("marshal layout for diagram %q: %w", d.ID, err)
		}
		folderID := diagFolderID[d.ID] // nil if no folder

		// Nullable viewpoint fields.
		var viewpoint, viewpointRef *string
		if d.Viewpoint != "" {
			viewpoint = &d.Viewpoint
		}
		if d.ViewpointRef != "" {
			viewpointRef = &d.ViewpointRef
		}

		var diagID uuid.UUID
		err = tx.QueryRow(`
			INSERT INTO diagrams
			  (workspace_id, source_id, name, documentation, layout, folder_id, viewpoint, viewpoint_ref)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (workspace_id, source_id) DO UPDATE
			  SET name = EXCLUDED.name,
			      documentation = EXCLUDED.documentation,
			      layout = EXCLUDED.layout,
			      folder_id = EXCLUDED.folder_id,
			      viewpoint = EXCLUDED.viewpoint,
			      viewpoint_ref = EXCLUDED.viewpoint_ref,
			      version = diagrams.version + 1, updated_at = now()
			RETURNING id`,
			wsID, d.ID, d.Name, d.Documentation, layoutJSON, folderID,
			viewpoint, viewpointRef).Scan(&diagID)
		if err != nil {
			return nil, errorf("upsert diagram %q: %w", d.ID, err)
		}
		result.Diagrams++

		// Lang variants — delete-then-insert for clean re-import.
		if _, err := tx.Exec(`DELETE FROM diagram_names WHERE diagram_id = $1`, diagID); err != nil {
			return nil, errorf("clear diagram names for %q: %w", d.ID, err)
		}
		if err := insertLangStrings(tx, "diagram_names", "diagram_id", diagID, "name", d.Names); err != nil {
			return nil, errorf("insert diagram names for %q: %w", d.ID, err)
		}
		if err := insertLangStrings(tx, "diagram_names", "diagram_id", diagID, "documentation", d.Documentations); err != nil {
			return nil, errorf("insert diagram docs for %q: %w", d.ID, err)
		}

		// View-level properties — always clear then re-insert.
		if _, err := tx.Exec(`DELETE FROM view_properties WHERE diagram_id = $1`, diagID); err != nil {
			return nil, errorf("clear view properties for %q: %w", d.ID, err)
		}
		for _, p := range d.Properties {
			if _, err := tx.Exec(`
				INSERT INTO view_properties (diagram_id, definition_ref, key, value)
				VALUES ($1, $2, $3, $4)`,
				diagID, p.DefinitionRef, p.Key, p.Value); err != nil {
				return nil, errorf("insert view property %q for %q: %w", p.Key, d.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, errorf("commit transaction: %w", err)
	}
	return result, nil
}

// ImportModel imports a parsed model into the given workspace.
// It is exported so the seed command can use it without going through HTTP.
func ImportModel(db *sql.DB, wsID uuid.UUID, m *parser.Model) (*ImportResult, error) {
	return importInTx(db, wsID, m)
}

func registerImportRoutes(r chi.Router, db *sql.DB, svc *auth.Service, auditStore *audit.Store, snapStore *snapshot.Store) {
	h := &importHandler{db: db, audit: auditStore, snaps: snapStore}
	r.With(svc.RequireWorkspaceAccess(auth.RoleEditor)).Post("/workspaces/{id}/import", h.importModel)
	r.With(svc.RequireWorkspaceAccess(auth.RoleEditor)).Post("/workspaces/{id}/import/preview", h.previewImport)
}
