-- Migration 022: store view-level <properties> from AOEF <view> elements.
-- Mirrors element_properties so diagrams can carry custom key/value pairs
-- from <view><properties><property> in the source AOEF file.
CREATE TABLE IF NOT EXISTS view_properties (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    diagram_id  UUID        NOT NULL REFERENCES diagrams(id) ON DELETE CASCADE,
    definition_ref TEXT     NOT NULL DEFAULT '',
    key         TEXT        NOT NULL,
    value       TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS view_properties_diagram_idx ON view_properties(diagram_id);
CREATE INDEX IF NOT EXISTS view_properties_key_idx     ON view_properties(key);
