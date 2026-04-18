-- Change payload from jsonb to text so snapshots can store AOEF XML (not just AJX JSON).
ALTER TABLE workspace_snapshots ALTER COLUMN payload TYPE text USING payload::text;
