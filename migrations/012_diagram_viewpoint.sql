-- Migration 012: diagram_viewpoint
-- Stores the ArchiMate viewpoint associated with each view/diagram.

ALTER TABLE diagrams
    ADD COLUMN IF NOT EXISTS viewpoint     TEXT, -- e.g. "Layered", "Application Usage", free-form or enum value
    ADD COLUMN IF NOT EXISTS viewpoint_ref TEXT; -- identifierRef to a viewpoint defined in <viewpoints>
