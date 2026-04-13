-- Migration 011: relationship_attributes
-- Captures OEF relationship type-specific attributes that carry semantic meaning.

ALTER TABLE relationships
    ADD COLUMN IF NOT EXISTS access_type  TEXT,                          -- Access: Access|Read|Write|ReadWrite
    ADD COLUMN IF NOT EXISTS is_directed  BOOLEAN NOT NULL DEFAULT false, -- Association: directed or undirected
    ADD COLUMN IF NOT EXISTS modifier     TEXT;                          -- Influence: strength/sign (+, -, ++, etc.)
