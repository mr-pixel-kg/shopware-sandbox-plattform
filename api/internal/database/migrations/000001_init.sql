-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS images (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    tag VARCHAR(255) NOT NULL,
    title VARCHAR(255) NULL,
    description TEXT NULL,
    thumbnail_url VARCHAR(1024) NULL,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    user_id UUID NULL REFERENCES users(id) ON DELETE CASCADE,
    session_type VARCHAR(32) NOT NULL,
    token_id VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS sandboxes (
    id UUID PRIMARY KEY,
    image_id UUID NOT NULL REFERENCES images(id),
    created_by_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    guest_session_id UUID NULL REFERENCES sessions(id) ON DELETE SET NULL,
    status VARCHAR(32) NOT NULL,
    container_id VARCHAR(255) NOT NULL UNIQUE,
    container_name VARCHAR(255) NOT NULL UNIQUE,
    url VARCHAR(1024) NOT NULL UNIQUE,
    client_ip VARCHAR(128) NOT NULL,
    expires_at TIMESTAMPTZ NULL,
    last_seen_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS sandbox_events (
    id UUID PRIMARY KEY,
    sandbox_id UUID NOT NULL REFERENCES sandboxes(id) ON DELETE CASCADE,
    event_type VARCHAR(64) NOT NULL,
    description TEXT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(128) NOT NULL,
    ip_address VARCHAR(128) NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sandboxes_status ON sandboxes(status);
CREATE INDEX IF NOT EXISTS idx_sandboxes_expires_at ON sandboxes(expires_at);
CREATE INDEX IF NOT EXISTS idx_sandboxes_client_ip ON sandboxes(client_ip);
CREATE INDEX IF NOT EXISTS idx_sandbox_events_sandbox_id ON sandbox_events(sandbox_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token_id ON sessions(token_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS sandbox_events;
DROP TABLE IF EXISTS sandboxes;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS images;
DROP TABLE IF EXISTS users;
