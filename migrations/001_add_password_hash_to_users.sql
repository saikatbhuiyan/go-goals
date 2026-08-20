-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;

-- Existing users may not have passwords. Keep the column nullable so they can
-- remain in the table until you migrate or reset those accounts.

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
