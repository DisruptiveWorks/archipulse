-- Migration 004: diagrams (views)
-- Maps AOEF <view> to a row. The layout is stored as JSONB.

CREATE TABLE diagrams (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    source_id       TEXT NOT NULL,              -- original id from AOEF/AJX file
    name            TEXT NOT NULL DEFAULT '',
    documentation   TEXT NOT NULL DEFAULT '',
    layout          JSONB NOT NULL DEFAULT '{}',-- node positions and connection paths
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, source_id)
);

CREATE INDEX diagrams_workspace_idx ON diagrams (workspace_id);
