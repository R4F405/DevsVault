CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    slug TEXT NOT NULL CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, slug)
);

CREATE TABLE environments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    slug TEXT NOT NULL CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,62}$'),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, slug)
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject TEXT NOT NULL UNIQUE,
    email TEXT,
    display_name TEXT,
    oidc_issuer TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    disabled_at TIMESTAMPTZ
);

CREATE TABLE service_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    subject TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    disabled_at TIMESTAMPTZ
);

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_type TEXT NOT NULL CHECK (actor_type IN ('user', 'service')),
    actor_id UUID NOT NULL,
    role_id UUID REFERENCES roles(id),
    action TEXT NOT NULL,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID REFERENCES environments(id) ON DELETE CASCADE,
    secret_id UUID,
    effect TEXT NOT NULL CHECK (effect IN ('allow', 'deny')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ
);

CREATE TABLE secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (name ~ '^[A-Z0-9_][A-Z0-9_./-]{0,127}$'),
    logical_path TEXT NOT NULL,
    active_version INTEGER,
    created_by_type TEXT NOT NULL CHECK (created_by_type IN ('user', 'service')),
    created_by_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ,
    UNIQUE (environment_id, name),
    UNIQUE (logical_path)
);

CREATE TABLE secret_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id UUID NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    ciphertext BYTEA NOT NULL,
    nonce BYTEA NOT NULL,
    wrapped_dek BYTEA NOT NULL,
    dek_nonce BYTEA NOT NULL,
    key_id TEXT NOT NULL,
    algorithm TEXT NOT NULL DEFAULT 'AES-256-GCM',
    created_by_type TEXT NOT NULL CHECK (created_by_type IN ('user', 'service')),
    created_by_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ,
    revoked_by_type TEXT CHECK (revoked_by_type IN ('user', 'service')),
    revoked_by_id UUID,
    UNIQUE (secret_id, version)
);

ALTER TABLE secrets
    ADD CONSTRAINT secrets_active_version_fk
    FOREIGN KEY (id, active_version) REFERENCES secret_versions(secret_id, version)
    DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE secret_access_grants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id UUID NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    actor_type TEXT NOT NULL CHECK (actor_type IN ('user', 'service')),
    actor_id UUID NOT NULL,
    action TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    UNIQUE (secret_id, actor_type, actor_id, action)
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ
);

CREATE TABLE service_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_identity_id UUID NOT NULL REFERENCES service_identities(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    actor_type TEXT NOT NULL CHECK (actor_type IN ('user', 'service', 'anonymous')),
    actor_id TEXT NOT NULL,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    workspace_id UUID,
    project_id UUID,
    environment_id UUID,
    outcome TEXT NOT NULL CHECK (outcome IN ('success', 'denied', 'error')),
    ip_address INET,
    user_agent TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX idx_projects_workspace_id ON projects(workspace_id);
CREATE INDEX idx_environments_project_id ON environments(project_id);
CREATE INDEX idx_secrets_environment_id ON secrets(environment_id);
CREATE INDEX idx_secret_versions_secret_id ON secret_versions(secret_id);
CREATE INDEX idx_policies_actor ON policies(actor_type, actor_id);
CREATE INDEX idx_audit_logs_occurred_at ON audit_logs(occurred_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);

INSERT INTO roles (name, description) VALUES
    ('admin', 'Full workspace administration'),
    ('developer', 'Manage metadata and write secrets in allowed scopes'),
    ('runtime-service', 'Read explicitly granted secret values at runtime'),
    ('auditor', 'Read audit logs and metadata without secret values')
ON CONFLICT (name) DO NOTHING;