-- Migration 023: saved_views
-- Persists user-defined bookmarks of automatic views (Dashboard, Landscape, etc.)
-- with their filter state. folder_id reserved for future folder support.

CREATE TABLE saved_views (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    view_type    TEXT NOT NULL,
    name         TEXT NOT NULL,
    filters      JSONB NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX saved_views_workspace_idx ON saved_views (workspace_id);
