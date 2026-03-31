-- +goose Up
ALTER TABLE sandboxes ADD COLUMN IF NOT EXISTS state_reason VARCHAR(512) NULL;

-- +goose Down
ALTER TABLE sandboxes DROP COLUMN IF EXISTS state_reason;
