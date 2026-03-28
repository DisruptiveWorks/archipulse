-- Migration 003: relationships
-- Maps AOEF <relationship> to a row.

CREATE TABLE relationships (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    source_id       TEXT NOT NULL,              -- original id from AOEF/AJX file
    type            TEXT NOT NULL,              -- e.g. "AssociationRelationship", "CompositionRelationship"
    source_element  TEXT NOT NULL,              -- source_id of the source element
    target_element  TEXT NOT NULL,              -- source_id of the target element
    name            TEXT NOT NULL DEFAULT '',
    documentation   TEXT NOT NULL DEFAULT '',
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, source_id)
);

CREATE INDEX relationships_workspace_idx ON relationships (workspace_id);
CREATE INDEX relationships_source_idx    ON relationships (workspace_id, source_element);
CREATE INDEX relationships_target_idx    ON relationships (workspace_id, target_element);
