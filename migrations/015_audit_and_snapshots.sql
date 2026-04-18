-- workspace_events: audit log of user actions within a workspace.
CREATE TABLE workspace_events (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id uuid        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id      text        NOT NULL,
    user_email   text        NOT NULL,
    action       text        NOT NULL, -- create | update | delete | import | add_member | remove_member | update_member_role | create_snapshot | restore_snapshot
    entity_type  text        NOT NULL, -- element | relationship | diagram | workspace | member | snapshot
    entity_id    text,
    entity_name  text,
    meta         jsonb,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX workspace_events_workspace_id_idx ON workspace_events (workspace_id, created_at DESC);

-- workspace_snapshots: point-in-time exports of a workspace model.
CREATE TABLE workspace_snapshots (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    uuid        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    created_by      text        NOT NULL,
    created_by_email text       NOT NULL,
    label           text,                   -- null = auto (import-triggered), set = manual
    trigger         text        NOT NULL,   -- import | manual
    payload         jsonb       NOT NULL,   -- full AOEF export
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX workspace_snapshots_workspace_id_idx ON workspace_snapshots (workspace_id, created_at DESC);
