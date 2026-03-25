-- +goose Up
ALTER TABLE images ADD COLUMN IF NOT EXISTS metadata jsonb DEFAULT '[]';
ALTER TABLE images ADD COLUMN IF NOT EXISTS registry_ref VARCHAR(255) DEFAULT NULL;
ALTER TABLE sandboxes ADD COLUMN IF NOT EXISTS metadata jsonb DEFAULT '{}';

-- +goose Down
ALTER TABLE sandboxes DROP COLUMN IF EXISTS metadata;
ALTER TABLE images DROP COLUMN IF EXISTS registry_ref;
ALTER TABLE images DROP COLUMN IF EXISTS metadata;
