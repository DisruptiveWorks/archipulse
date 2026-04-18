-- Migration 020: store ArchiMate viewpoint definitions from <views><viewpoints>.
-- concerns and modeling_notes are JSONB because they are complex nested objects
-- that are always read/written as a unit and never individually queried.
-- allowed_element_types and allowed_relationship_types are TEXT[] (flat enum arrays).
CREATE TABLE IF NOT EXISTS viewpoints (
    id                          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id                UUID        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    source_id                   TEXT        NOT NULL,
    name                        TEXT        NOT NULL DEFAULT '',
    documentation               TEXT        NOT NULL DEFAULT '',
    purpose                     TEXT        NOT NULL DEFAULT '',
    content                     TEXT        NOT NULL DEFAULT '',
    concerns                    JSONB       NOT NULL DEFAULT '[]',
    allowed_element_types       TEXT[]      NOT NULL DEFAULT '{}',
    allowed_relationship_types  TEXT[]      NOT NULL DEFAULT '{}',
    modeling_notes              JSONB       NOT NULL DEFAULT '[]',
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, source_id)
);

CREATE INDEX IF NOT EXISTS viewpoints_workspace_idx ON viewpoints(workspace_id);
