# Database Schema

ArchiPulse uses PostgreSQL. All tables are created and updated via numbered SQL migrations in [`migrations/`](../migrations/).

## Entity-Relationship Diagram

```mermaid
erDiagram

    workspaces {
        uuid    id          PK
        text    name        UK
        text    purpose     "as-is|to-be|initiative|other"
        text    description
        integer version
        timestamptz created_at
        timestamptz updated_at
    }

    users {
        uuid    id    PK
        text    email UK
        text    password_hash
        text    role  "admin|architect|viewer"
        timestamptz created_at
        timestamptz updated_at
    }

    elements {
        uuid    id           PK
        uuid    workspace_id FK
        text    source_id    "OEF identifier"
        text    type         "e.g. ApplicationComponent"
        text    layer        "Business|Application|Technology|..."
        text    name
        text    documentation
        integer version
        timestamptz created_at
        timestamptz updated_at
    }

    element_properties {
        uuid    id         PK
        uuid    element_id FK
        text    key
        text    value
        text    source     "model|extractor"
        timestamptz collected_at
        timestamptz created_at
    }

    property_definitions {
        uuid    id           PK
        uuid    workspace_id FK
        text    source_id    "OEF identifier"
        text    name
        text    data_type    "string|boolean|currency|date|time|number"
        timestamptz created_at
    }

    relationships {
        uuid    id             PK
        uuid    workspace_id   FK
        text    source_id      "OEF identifier"
        text    type           "e.g. AssociationRelationship"
        text    source_element "elements.source_id"
        text    target_element "elements.source_id"
        text    name
        text    documentation
        text    access_type    "Access|Read|Write|ReadWrite (Access rel only)"
        boolean is_directed    "false by default (Association rel only)"
        text    modifier       "influence strength: +|++|-|--|0..10 (Influence rel only)"
        integer version
        timestamptz created_at
        timestamptz updated_at
    }

    diagram_folders {
        uuid    id           PK
        uuid    workspace_id FK
        uuid    parent_id    FK "self-ref, nullable"
        text    name
        text    source_id    "OEF organization label path"
        integer position     "ordering within parent"
        timestamptz created_at
    }

    diagrams {
        uuid    id            PK
        uuid    workspace_id  FK
        uuid    folder_id     FK "nullable"
        text    source_id     "OEF identifier"
        text    name
        text    documentation
        jsonb   layout        "nodes + connections with positions and styles"
        text    viewpoint     "OEF viewpoint name (e.g. Layered)"
        text    viewpoint_ref "OEF viewpointRef identifier"
        integer version
        timestamptz created_at
        timestamptz updated_at
    }

    workspaces    ||--o{ elements            : "contains"
    workspaces    ||--o{ relationships        : "contains"
    workspaces    ||--o{ diagrams             : "contains"
    workspaces    ||--o{ diagram_folders      : "contains"
    workspaces    ||--o{ property_definitions : "defines"

    elements      ||--o{ element_properties  : "has"
    diagram_folders ||--o{ diagram_folders   : "parent"
    diagram_folders ||--o{ diagrams          : "groups"
```

## Notes

### `layout` JSON structure (diagrams)

The `layout` column is a JSONB blob produced by the parser. Its structure mirrors the OEF diagram visual model:

```json
{
  "Nodes": [
    {
      "ElementID": "id-app-001",
      "ParentElementID": "",
      "NodeType": "Element",
      "X": 120, "Y": 80, "W": 120, "H": 55,
      "Style": {
        "FillColor": { "R": 255, "G": 251, "B": 235, "A": null },
        "LineColor": { "R": 217, "G": 119, "B": 6 },
        "Font": { "Name": "Arial", "Size": "9", "Style": "plain", "Color": null },
        "LineWidth": 1
      }
    }
  ],
  "Connections": [
    {
      "RelationshipID": "id-rel-001",
      "Label": "",
      "Bendpoints": [{ "X": 180, "Y": 108 }],
      "Style": null
    }
  ]
}
```

- **`Style`** is stored as parsed from the OEF file and is available for future custom rendering. ArchiPulse's current theme uses its own colour palette and ignores `Style.FillColor`/`Style.LineColor` by default.
- **`NodeType`** reflects the OEF `xsi:type` on each node (`Element`, `Container`, `Label`, etc.).

### Relationship semantic attributes

Three columns capture OEF type-specific semantics that affect how relationships are rendered:

| Column | Applies to | Values |
|---|---|---|
| `access_type` | `AccessRelationship` | `Access` (default) · `Read` · `Write` · `ReadWrite` |
| `is_directed` | `AssociationRelationship` | `false` (default) · `true` |
| `modifier` | `InfluenceRelationship` | `+` · `++` · `-` · `--` · `0`–`10` |

### Property definitions

OEF `<propertyDefinition>` entries are stored workspace-scoped in `property_definitions`. The `data_type` column records the declared type (`string`, `boolean`, `currency`, `date`, `time`, `number`), enabling typed property editing in a future editor feature.

### `source_id` vs `id`

Every ArchiMate concept table has both an `id` (internal UUID, stable across re-imports) and a `source_id` (the original identifier from the OEF/AJX file). Re-importing the same model is idempotent via `ON CONFLICT (workspace_id, source_id) DO UPDATE`.
```
