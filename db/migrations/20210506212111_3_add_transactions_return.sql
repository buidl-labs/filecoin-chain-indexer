-- +goose Up
-- SQL in this section is executed when the migration is applied.

ALTER TABLE transactions ADD COLUMN return_bytes TEXT;
ALTER TABLE tmp_transactions ADD COLUMN return_bytes TEXT;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

ALTER TABLE transactions DROP COLUMN return_bytes;
ALTER TABLE tmp_transactions DROP COLUMN return_bytes;
