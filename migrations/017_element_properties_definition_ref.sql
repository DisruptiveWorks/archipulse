-- Migration 017: add definition_ref to element_properties for AOEF round-trip fidelity.
-- Stores the original propertyDefinitionRef identifier so the exporter can reconstruct
-- the correct <property propertyDefinitionRef="..."> attribute without a name→ID lookup.
ALTER TABLE element_properties
    ADD COLUMN IF NOT EXISTS definition_ref TEXT NOT NULL DEFAULT '';
