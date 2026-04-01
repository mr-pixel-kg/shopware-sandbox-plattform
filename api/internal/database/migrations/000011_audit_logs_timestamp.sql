-- +goose Up
ALTER TABLE audit_logs
    RENAME COLUMN created_at TO timestamp;

ALTER INDEX IF EXISTS idx_audit_logs_created_at
    RENAME TO idx_audit_logs_timestamp;

-- +goose Down
ALTER TABLE audit_logs
    RENAME COLUMN timestamp TO created_at;

ALTER INDEX IF EXISTS idx_audit_logs_timestamp
    RENAME TO idx_audit_logs_created_at;
