-- Migration 002: elements
-- Maps AOEF <element> to a row. type uses ArchiMate 3.2 element type names.

CREATE TABLE elements (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    source_id       TEXT NOT NULL,              -- original id from AOEF/AJX file
    type            TEXT NOT NULL,              -- e.g. "ApplicationComponent", "BusinessProcess"
    name            TEXT NOT NULL DEFAULT '',
    documentation   TEXT NOT NULL DEFAULT '',
    version         INTEGER NOT NULL DEFAULT 1, -- optimistic locking
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, source_id)
);

CREATE INDEX elements_workspace_idx ON elements (workspace_id);
CREATE INDEX elements_type_idx      ON elements (workspace_id, type);
