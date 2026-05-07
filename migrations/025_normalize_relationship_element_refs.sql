-- Normalize source_element / target_element for relationships created via the
-- MCP server or REST API before the resolveElementRef fix. Those paths stored
-- element UUIDs instead of ArchiMate source_ids, breaking all viewer queries
-- that JOIN on elements.source_id = relationships.source_element/target_element.

UPDATE relationships r
SET source_element = COALESCE(
        (SELECT e.source_id FROM elements e
         WHERE e.id::text = r.source_element
           AND e.workspace_id = r.workspace_id
           AND e.source_id <> ''),
        r.source_element)
WHERE r.source_element ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';

UPDATE relationships r
SET target_element = COALESCE(
        (SELECT e.source_id FROM elements e
         WHERE e.id::text = r.target_element
           AND e.workspace_id = r.workspace_id
           AND e.source_id <> ''),
        r.target_element)
WHERE r.target_element ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';
