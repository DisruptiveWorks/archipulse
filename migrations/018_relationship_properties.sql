-- Migration 018: add relationship_properties table for AOEF <property> round-trip.
-- Mirrors element_properties so relationships can carry custom key/value pairs
-- from <relationship><properties><property> elements in the source AOEF file.
CREATE TABLE IF NOT EXISTS relationship_properties (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    relationship_id UUID        NOT NULL REFERENCES relationships(id) ON DELETE CASCADE,
    definition_ref  TEXT        NOT NULL DEFAULT '',
    key             TEXT        NOT NULL,
    value           TEXT        NOT NULL,
    source          TEXT        NOT NULL DEFAULT 'model',
    collected_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS relationship_properties_rel_idx  ON relationship_properties(relationship_id);
CREATE INDEX IF NOT EXISTS relationship_properties_key_idx  ON relationship_properties(key);
CREATE INDEX IF NOT EXISTS relationship_properties_src_idx  ON relationship_properties(source);
