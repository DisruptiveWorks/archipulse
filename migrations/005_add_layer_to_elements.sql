-- Migration 005: add layer column to elements
-- layer is derived from the ArchiMate element type and stored for efficient EAM view queries.

ALTER TABLE elements ADD COLUMN layer TEXT NOT NULL DEFAULT '';

CREATE INDEX elements_layer_idx ON elements (workspace_id, layer);
