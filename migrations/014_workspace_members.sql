-- Migration 014: workspace membership RBAC
--
-- 1. Rename users.role → users.org_role (admin | member).
--    Existing admins keep 'admin'; all other roles become 'member'.
-- 2. Add workspace_members table for per-workspace roles.
--    Roles: owner | editor | viewer.
-- 3. Seed the first admin as owner of every existing workspace so
--    no existing data is left inaccessible.

-- Step 1: rename column and normalise values.
ALTER TABLE users RENAME COLUMN role TO org_role;
UPDATE users SET org_role = 'member' WHERE org_role NOT IN ('admin', 'member');

-- Step 2: workspace membership table.
CREATE TABLE workspace_members (
    workspace_id UUID        NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id      UUID        NOT NULL REFERENCES users(id)      ON DELETE CASCADE,
    role         TEXT        NOT NULL CHECK (role IN ('owner', 'editor', 'viewer')),
    invited_by   UUID        REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (workspace_id, user_id)
);

-- Step 3: seed — every org admin becomes owner of every workspace.
INSERT INTO workspace_members (workspace_id, user_id, role)
SELECT w.id, u.id, 'owner'
FROM   workspaces w
CROSS  JOIN users u
WHERE  u.org_role = 'admin'
ON CONFLICT DO NOTHING;
