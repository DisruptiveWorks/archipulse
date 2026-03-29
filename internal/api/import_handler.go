package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DisruptiveWorks/archipulse/internal/parser"
	"github.com/DisruptiveWorks/archipulse/internal/workspace"
)

const maxUploadSize = 32 << 20 // 32 MB

type importHandler struct {
	db *sql.DB
}

// ImportResult summarises what was imported.
type ImportResult struct {
	WorkspaceID   string `json:"workspace_id"`
	Elements      int    `json:"elements"`
	Relationships int    `json:"relationships"`
	Diagrams      int    `json:"diagrams"`
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

	// Import inside a transaction — all or nothing.
	result, err := importInTx(h.db, wsID, m)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
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

	for _, e := range m.Elements {
		layer := parser.ElementLayer(e.Type)
		_, err := tx.Exec(`
			INSERT INTO elements (workspace_id, source_id, type, layer, name, documentation)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (workspace_id, source_id) DO UPDATE
			  SET type = EXCLUDED.type, layer = EXCLUDED.layer,
			      name = EXCLUDED.name, documentation = EXCLUDED.documentation,
			      version = elements.version + 1, updated_at = now()`,
			wsID, e.ID, e.Type, layer, e.Name, e.Documentation)
		if err != nil {
			return nil, errorf("upsert element %q: %w", e.ID, err)
		}
		result.Elements++
	}

	for _, r := range m.Relationships {
		_, err := tx.Exec(`
			INSERT INTO relationships (workspace_id, source_id, type, source_element, target_element, name, documentation)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (workspace_id, source_id) DO UPDATE
			  SET type = EXCLUDED.type, source_element = EXCLUDED.source_element,
			      target_element = EXCLUDED.target_element, name = EXCLUDED.name,
			      documentation = EXCLUDED.documentation,
			      version = relationships.version + 1, updated_at = now()`,
			wsID, r.ID, r.Type, r.Source, r.Target, r.Name, r.Documentation)
		if err != nil {
			return nil, errorf("upsert relationship %q: %w", r.ID, err)
		}
		result.Relationships++
	}

	for _, d := range m.Diagrams {
		layoutJSON, err := json.Marshal(d.Layout)
		if err != nil {
			return nil, errorf("marshal layout for diagram %q: %w", d.ID, err)
		}
		_, err = tx.Exec(`
			INSERT INTO diagrams (workspace_id, source_id, name, documentation, layout)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (workspace_id, source_id) DO UPDATE
			  SET name = EXCLUDED.name, documentation = EXCLUDED.documentation,
			      layout = EXCLUDED.layout,
			      version = diagrams.version + 1, updated_at = now()`,
			wsID, d.ID, d.Name, d.Documentation, layoutJSON)
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

func registerImportRoutes(r chi.Router, db *sql.DB) {
	h := &importHandler{db: db}
	r.Post("/workspaces/{id}/import", h.importModel)
}
