-- +goose Up
ALTER TABLE images ADD COLUMN IF NOT EXISTS status VARCHAR(32) NOT NULL DEFAULT 'ready';
ALTER TABLE images ADD COLUMN IF NOT EXISTS error TEXT NULL;
CREATE INDEX IF NOT EXISTS idx_images_status ON images(status);

-- +goose Down
DROP INDEX IF EXISTS idx_images_status;
ALTER TABLE images DROP COLUMN IF EXISTS error;
ALTER TABLE images DROP COLUMN IF EXISTS status;
