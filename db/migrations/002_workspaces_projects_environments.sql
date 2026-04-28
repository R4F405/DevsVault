ALTER TABLE workspaces
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS created_by TEXT NOT NULL DEFAULT 'system',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS created_by TEXT NOT NULL DEFAULT 'system',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE environments
    ADD COLUMN IF NOT EXISTS created_by TEXT NOT NULL DEFAULT 'system',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE secrets DROP CONSTRAINT IF EXISTS secrets_workspace_id_fkey;
ALTER TABLE secrets DROP CONSTRAINT IF EXISTS secrets_project_id_fkey;
ALTER TABLE secrets DROP CONSTRAINT IF EXISTS secrets_environment_id_fkey;

ALTER TABLE secrets
    ADD CONSTRAINT secrets_workspace_id_fkey
    FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE RESTRICT,
    ADD CONSTRAINT secrets_project_id_fkey
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE RESTRICT,
    ADD CONSTRAINT secrets_environment_id_fkey
    FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_workspaces_slug ON workspaces(slug);
CREATE INDEX IF NOT EXISTS idx_projects_workspace_slug ON projects(workspace_id, slug);
CREATE INDEX IF NOT EXISTS idx_environments_project_slug ON environments(project_id, slug);