-- Migration 010: add folder_id to diagrams
-- Links each diagram to its folder in the hierarchy.

ALTER TABLE diagrams
    ADD COLUMN folder_id UUID REFERENCES diagram_folders (id) ON DELETE SET NULL;

CREATE INDEX diagrams_folder_idx ON diagrams (folder_id);
