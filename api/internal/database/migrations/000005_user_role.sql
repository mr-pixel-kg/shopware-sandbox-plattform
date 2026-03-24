-- +goose Up
ALTER TABLE users ADD COLUMN role VARCHAR(32) NOT NULL DEFAULT '';
ALTER TABLE users ADD CONSTRAINT chk_users_role CHECK (role IN ('', 'admin', 'user'));

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;
ALTER TABLE users DROP COLUMN IF EXISTS role;
