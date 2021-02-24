-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE miner_transactions
(
    "height" TEXT NOT NULL,
    "cid" TEXT NOT NULL,
    "sender" TEXT,
    "receiver" TEXT,
    "amount" TEXT,
    "gas_fee_cap" TEXT,
    "gas_premium" TEXT,
    "gas_limit" BIGINT,
    "nonce" BIGINT,
    "method" BIGINT,
    "from_actor_name" TEXT,
    "to_actor_name" TEXT,
    PRIMARY KEY ("height", "cid")
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE miner_transactions;
