-- Migration 021: store all xml:lang variants of names and documentation.
-- The existing name/documentation TEXT columns remain the primary (first/preferred)
-- value for backward compatibility with all API consumers.
-- These tables capture the full set of language variants for round-trip AOEF fidelity.

CREATE TABLE IF NOT EXISTS element_names (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    element_id  UUID        NOT NULL REFERENCES elements(id) ON DELETE CASCADE,
    field       TEXT        NOT NULL CHECK (field IN ('name', 'documentation')),
    lang        TEXT        NOT NULL DEFAULT '',
    value       TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (element_id, field, lang)
);

CREATE TABLE IF NOT EXISTS relationship_names (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    relationship_id  UUID        NOT NULL REFERENCES relationships(id) ON DELETE CASCADE,
    field            TEXT        NOT NULL CHECK (field IN ('name', 'documentation')),
    lang             TEXT        NOT NULL DEFAULT '',
    value            TEXT        NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (relationship_id, field, lang)
);

CREATE TABLE IF NOT EXISTS diagram_names (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    diagram_id  UUID        NOT NULL REFERENCES diagrams(id) ON DELETE CASCADE,
    field       TEXT        NOT NULL CHECK (field IN ('name', 'documentation')),
    lang        TEXT        NOT NULL DEFAULT '',
    value       TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (diagram_id, field, lang)
);

CREATE INDEX IF NOT EXISTS element_names_element_idx      ON element_names(element_id);
CREATE INDEX IF NOT EXISTS relationship_names_rel_idx     ON relationship_names(relationship_id);
CREATE INDEX IF NOT EXISTS diagram_names_diagram_idx      ON diagram_names(diagram_id);
