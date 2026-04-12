-- Migration 009: diagram_folders
-- Stores the folder hierarchy from AOEF <views> organization items.

CREATE TABLE diagram_folders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    parent_id       UUID REFERENCES diagram_folders (id) ON DELETE CASCADE,
    name            TEXT NOT NULL DEFAULT '',
    source_id       TEXT NOT NULL DEFAULT '',  -- original item label or id from AOEF
    position        INTEGER NOT NULL DEFAULT 0, -- ordering within parent
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, source_id)
);

CREATE INDEX diagram_folders_workspace_idx ON diagram_folders (workspace_id);
CREATE INDEX diagram_folders_parent_idx    ON diagram_folders (parent_id);
