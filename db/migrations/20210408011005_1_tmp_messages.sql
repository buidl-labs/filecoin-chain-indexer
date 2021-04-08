-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE tmp_transactions
(
    "cid" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "sender" TEXT NOT NULL,
    "receiver" TEXT NOT NULL,
    "amount" TEXT NOT NULL,
    "type" INT,
    "gas_fee_cap" TEXT NOT NULL,
    "gas_premium" TEXT NOT NULL,
    "gas_limit" BIGINT,
    "size_bytes" BIGINT,
    "nonce" BIGINT,
    "method" BIGINT,
    "method_name" TEXT,
    "params" TEXT,
    "params_bytes" bytea,
    "transferred" TEXT,
    "state_root" TEXT NOT NULL,
    "exit_code" BIGINT NOT NULL,
    "gas_used" BIGINT NOT NULL,
    "parent_base_fee" TEXT NOT NULL,
    "base_fee_burn" TEXT NOT NULL,
    "over_estimation_burn" TEXT NOT NULL,
    "miner_penalty" TEXT NOT NULL,
    "miner_tip" TEXT NOT NULL,
    "refund" TEXT NOT NULL,
    "gas_refund" BIGINT NOT NULL,
    "gas_burned" BIGINT NOT NULL,
    "actor_name" TEXT
);

CREATE TABLE tmp_parsed_messages
(
    "cid" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "sender" TEXT NOT NULL,
    "receiver" TEXT NOT NULL,
    "value" TEXT NOT NULL,
    "method" TEXT NOT NULL,
    "params" jsonb
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE tmp_transactions;
DROP TABLE tmp_parsed_messages;
