-- +goose Up
ALTER TABLE images
    RENAME COLUMN created_by_user_id TO owner_id;

ALTER TABLE sandboxes
    RENAME COLUMN created_by_user_id TO owner_id;

-- +goose Down
ALTER TABLE sandboxes
    RENAME COLUMN owner_id TO created_by_user_id;

ALTER TABLE images
    RENAME COLUMN owner_id TO created_by_user_id;
