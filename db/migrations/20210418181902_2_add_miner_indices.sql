-- +goose Up
-- SQL in this section is executed when the migration is applied.

ALTER TABLE transactions ADD COLUMN miner TEXT;
ALTER TABLE parsed_messages ADD COLUMN miner TEXT;

ALTER TABLE tmp_transactions ADD COLUMN miner TEXT;
ALTER TABLE tmp_parsed_messages ADD COLUMN miner TEXT;

CREATE INDEX idx_transactions_miner ON transactions(miner);
CREATE INDEX idx_parsed_messages_miner ON parsed_messages(miner);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP INDEX idx_transactions_miner;
DROP INDEX idx_parsed_messages_miner;

ALTER TABLE tmp_transactions DROP COLUMN miner;
ALTER TABLE tmp_parsed_messages DROP COLUMN miner;

ALTER TABLE transactions DROP COLUMN miner;
ALTER TABLE parsed_messages DROP COLUMN miner;
