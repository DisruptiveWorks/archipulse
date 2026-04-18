-- Migration 019: store AOEF model/@identifier and model-level <properties>.
ALTER TABLE workspaces
    ADD COLUMN IF NOT EXISTS model_identifier TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS model_properties (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    definition_ref  TEXT        NOT NULL DEFAULT '',
    key             TEXT        NOT NULL,
    value           TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS model_properties_workspace_idx ON model_properties(workspace_id);
