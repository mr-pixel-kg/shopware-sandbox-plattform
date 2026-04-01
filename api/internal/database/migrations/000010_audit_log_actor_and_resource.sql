-- +goose Up
ALTER TABLE audit_logs
    ADD COLUMN IF NOT EXISTS user_agent VARCHAR(512) NULL,
    ADD COLUMN IF NOT EXISTS client_token UUID NULL,
    ADD COLUMN IF NOT EXISTS resource_type VARCHAR(64) NULL,
    ADD COLUMN IF NOT EXISTS resource_id UUID NULL;

CREATE INDEX IF NOT EXISTS idx_audit_logs_client_token ON audit_logs(client_token);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs(resource_id);

-- +goose Down
DROP INDEX IF EXISTS idx_audit_logs_resource_id;
DROP INDEX IF EXISTS idx_audit_logs_resource_type;
DROP INDEX IF EXISTS idx_audit_logs_client_token;

ALTER TABLE audit_logs
    DROP COLUMN IF EXISTS resource_id,
    DROP COLUMN IF EXISTS resource_type,
    DROP COLUMN IF EXISTS client_token,
    DROP COLUMN IF EXISTS user_agent;
