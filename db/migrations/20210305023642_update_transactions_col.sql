-- +goose Up
-- SQL in this section is executed when the migration is applied.

ALTER TABLE transactions ADD COLUMN params_bytes bytea;
ALTER TABLE transactions ADD COLUMN params TEXT;
ALTER TABLE transactions ADD COLUMN transferred TEXT;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

ALTER TABLE transactions DROP COLUMN IF EXISTS params_bytes;
ALTER TABLE transactions DROP COLUMN IF EXISTS params;
ALTER TABLE transactions DROP COLUMN IF EXISTS transferred;
