package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// FolderNode is a folder in the diagram tree response.
type FolderNode struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	SourceID string        `json:"source_id"`
	Children []*FolderNode `json:"children,omitempty"`
	Diagrams []DiagramLeaf `json:"diagrams,omitempty"`
}

// DiagramLeaf is a diagram entry inside a folder.
type DiagramLeaf struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	SourceID string `json:"source_id"`
}

type folderHandler struct {
	db *sql.DB
}

// diagramTree returns the full folder+diagram tree for a workspace.
func (h *folderHandler) diagramTree(w http.ResponseWriter, r *http.Request) {
	wsID, err := uuid.Parse(chi.URLParam(r, "wsID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	// Load all folders.
	type dbFolder struct {
		id       string
		parentID *string
		name     string
		sourceID string
		position int
	}
	rows, err := h.db.Query(`
		SELECT id, parent_id, name, source_id, position
		FROM diagram_folders
		WHERE workspace_id = $1
		ORDER BY position`, wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, errorf("list folders: %w", err))
		return
	}
	defer rows.Close()

	folders := map[string]*FolderNode{}
	var rootFolderIDs []string

	for rows.Next() {
		var f dbFolder
		var parentID sql.NullString
		if err := rows.Scan(&f.id, &parentID, &f.name, &f.sourceID, &f.position); err != nil {
			respondError(w, http.StatusInternalServerError, err)
			return
		}
		if parentID.Valid {
			f.parentID = &parentID.String
		}
		node := &FolderNode{ID: f.id, Name: f.name, SourceID: f.sourceID}
		folders[f.id] = node
		if f.parentID == nil {
			rootFolderIDs = append(rootFolderIDs, f.id)
		}
	}
	if err := rows.Close(); err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	// Re-query to wire parent→child relationships (need parentID per row).
	rows2, err := h.db.Query(`
		SELECT id, parent_id FROM diagram_folders WHERE workspace_id = $1`, wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err)
		return
	}
	defer rows2.Close()
	for rows2.Next() {
		var id string
		var parentID sql.NullString
		if err := rows2.Scan(&id, &parentID); err != nil {
			respondError(w, http.StatusInternalServerError, err)
			return
		}
		if parentID.Valid {
			if parent, ok := folders[parentID.String]; ok {
				parent.Children = append(parent.Children, folders[id])
			}
		}
	}
	_ = rows2.Close()

	// Load all diagrams with their folder assignment.
	dRows, err := h.db.Query(`
		SELECT id, name, source_id, folder_id
		FROM diagrams
		WHERE workspace_id = $1
		ORDER BY name`, wsID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, errorf("list diagrams: %w", err))
		return
	}
	defer dRows.Close()

	var rootDiagrams []DiagramLeaf

	for dRows.Next() {
		var id, name, sourceID string
		var folderID sql.NullString
		if err := dRows.Scan(&id, &name, &sourceID, &folderID); err != nil {
			respondError(w, http.StatusInternalServerError, err)
			return
		}
		leaf := DiagramLeaf{ID: id, Name: name, SourceID: sourceID}
		if folderID.Valid {
			if folder, ok := folders[folderID.String]; ok {
				folder.Diagrams = append(folder.Diagrams, leaf)
			} else {
				rootDiagrams = append(rootDiagrams, leaf)
			}
		} else {
			rootDiagrams = append(rootDiagrams, leaf)
		}
	}
	_ = dRows.Close()

	// Build root-level response.
	type treeResponse struct {
		Folders  []*FolderNode `json:"folders"`
		Diagrams []DiagramLeaf `json:"diagrams"`
	}

	roots := make([]*FolderNode, 0, len(rootFolderIDs))
	for _, id := range rootFolderIDs {
		roots = append(roots, folders[id])
	}

	respondJSON(w, http.StatusOK, treeResponse{
		Folders:  roots,
		Diagrams: rootDiagrams,
	})
}

func registerFolderRoutes(r chi.Router, db *sql.DB) {
	h := &folderHandler{db: db}
	r.Get("/workspaces/{wsID}/diagram-tree", h.diagramTree)
}
