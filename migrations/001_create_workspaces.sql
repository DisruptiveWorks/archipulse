-- Migration 001: workspaces
-- A workspace is a named baseline of an architecture model.

CREATE TABLE workspaces (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    purpose     TEXT NOT NULL CHECK (purpose IN ('as-is', 'to-be', 'initiative', 'other')),
    description TEXT NOT NULL DEFAULT '',
    version     INTEGER NOT NULL DEFAULT 1,  -- optimistic locking
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX workspaces_name_idx ON workspaces (name);
