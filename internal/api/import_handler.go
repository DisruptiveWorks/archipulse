package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

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
	case strings.HasSuffix(name, ".ajx"), strings.HasSuffix(name, ".json"):
		m, err = parser.ParseAJX(file)
	default:
		respondError(w, http.StatusBadRequest, errorf("unsupported file format: use .xml (AOEF) or .ajx/.json (AJX)"))
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

func importInTx(db *sql.DB, wsID uuid.UUID, m *parser.Model) (*ImportResult, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result := &ImportResult{WorkspaceID: wsID.String()}

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

		if len(e.Properties) > 0 {
			// Remove existing model properties so a re-import stays clean.
			if _, err := tx.Exec(`DELETE FROM element_properties WHERE element_id = $1 AND source = 'model'`, elemID); err != nil {
				return nil, errorf("clear model properties for element %q: %w", e.ID, err)
			}
			props := make([]struct{ Key, Value string }, len(e.Properties))
			for i, p := range e.Properties {
				props[i] = struct{ Key, Value string }{p.Key, p.Value}
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
		_, err := tx.Exec(`
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
			      version = relationships.version + 1, updated_at = now()`,
			wsID, r.ID, r.Type, r.Source, r.Target,
			r.Name, r.Documentation, accessType, r.IsDirected, modifier)
		if err != nil {
			return nil, errorf("upsert relationship %q: %w", r.ID, err)
		}
		result.Relationships++
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

		_, err = tx.Exec(`
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
			      version = diagrams.version + 1, updated_at = now()`,
			wsID, d.ID, d.Name, d.Documentation, layoutJSON, folderID,
			viewpoint, viewpointRef)
		if err != nil {
			return nil, errorf("upsert diagram %q: %w", d.ID, err)
		}
		result.Diagrams++
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
}
