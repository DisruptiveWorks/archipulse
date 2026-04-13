-- Migration 013: property_definitions
-- Stores OEF <propertyDefinition> entries with their data type per workspace.
-- This is the schema-level definition; actual values live in element_properties.

CREATE TABLE IF NOT EXISTS property_definitions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    source_id       TEXT NOT NULL,          -- original identifier from AOEF file
    name            TEXT NOT NULL DEFAULT '',
    data_type       TEXT NOT NULL DEFAULT 'string', -- string|boolean|currency|date|time|number
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, source_id)
);

CREATE INDEX IF NOT EXISTS property_definitions_workspace_idx ON property_definitions (workspace_id);
