-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE transactions
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
    "actor_name" TEXT,
    PRIMARY KEY ("height", "cid", "state_root")
) PARTITION BY RANGE (height);

CREATE TABLE block_headers
(
    "height" BIGINT NOT NULL,
    "cid" TEXT NOT NULL,
    "miner_id" TEXT NOT NULL,
    "parent_weight" TEXT NOT NULL,
    "parent_state_root" TEXT NOT NULL,
    "parent_base_fee" TEXT NOT NULL,
    "fork_signaling" BIGINT NOT NULL,
    "win_count" BIGINT,
    "timestamp" BIGINT NOT NULL,
    PRIMARY KEY ("height", "cid")
);

CREATE TABLE block_messages
(
    "block" TEXT NOT NULL,
    "message" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    PRIMARY KEY ("height", "block", "message")
);

CREATE TABLE parsed_messages
(
    "cid" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "sender" TEXT NOT NULL,
    "receiver" TEXT NOT NULL,
    "value" TEXT NOT NULL,
    "method" TEXT NOT NULL,
    "params" jsonb,
    PRIMARY KEY ("height", "cid")
) PARTITION BY RANGE (height);

CREATE TABLE message_gas_economy
(
    "state_root" text NOT NULL,
    "height" bigint NOT NULL,
    "gas_limit_total" bigint,
    "gas_limit_unique_total" bigint,
    "base_fee" double precision,
    "base_fee_change_log" double precision,
    "gas_fill_ratio" double precision,
    "gas_capacity_ratio" double precision,
    "gas_waste_ratio" double precision,
    PRIMARY KEY ("height", "state_root")
);

CREATE TABLE receipts
(
    "message" TEXT NOT NULL,
    "state_root" TEXT NOT NULL,
    "idx" BIGINT NOT NULL,
    "exit_code" BIGINT NOT NULL,
    "gas_used" BIGINT NOT NULL,
    "height" BIGINT NOT NULL,
    PRIMARY KEY ("height", "message", "state_root")
);

CREATE TABLE miner_infos
(
    "miner_id" TEXT NOT NULL,
    "address" TEXT,
    "peer_id" TEXT NOT NULL,
    "owner_id" TEXT NOT NULL,
    "worker_id" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "state_root" TEXT NOT NULL,
    "storage_ask_price" TEXT NOT NULL,
    "min_piece_size" BIGINT NOT NULL,
    "max_piece_size" BIGINT NOT NULL,
    PRIMARY KEY ("height", "miner_id", "state_root")
);

CREATE TABLE miner_funds
(
    "miner_id" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "state_root" TEXT NOT NULL,
    "locked_funds" TEXT NOT NULL,
    "initial_pledge" TEXT NOT NULL,
    "pre_commit_deposits" TEXT NOT NULL,
    "available_balance" TEXT NOT NULL,
    PRIMARY KEY ("height", "miner_id", "state_root")
);

CREATE TABLE miner_fee_debts
(
    "height" BIGINT NOT NULL,
    "miner_id" TEXT NOT NULL,
    "state_root" TEXT NOT NULL,
    "fee_debt" TEXT NOT NULL,
    PRIMARY KEY ("height", "miner_id", "state_root")
);

CREATE TABLE miner_current_deadline_infos
(
    "height" BIGINT NOT NULL,
    "miner_id" TEXT NOT NULL,
    "state_root" TEXT NOT NULL,
    "deadline_index" BIGINT NOT NULL,
    "period_start" BIGINT NOT NULL,
    "open" BIGINT NOT NULL,
    "close" BIGINT NOT NULL,
    "challenge" BIGINT NOT NULL,
    "fault_cutoff" BIGINT NOT NULL,
    PRIMARY KEY ("height", "miner_id", "state_root")
);


CREATE TABLE miner_sector_infos
(
    "miner_id" TEXT NOT NULL,
    "sector_id" BIGINT NOT NULL,
    "state_root" TEXT NOT NULL,
    "sealed_cid" TEXT NOT NULL,
    "activation_epoch" BIGINT,
    "expiration_epoch" BIGINT,
    "deal_weight" TEXT NOT NULL,
    "verified_deal_weight" TEXT NOT NULL,
    "initial_pledge" TEXT NOT NULL,
    "expected_day_reward" TEXT NOT NULL,
    "expected_storage_pledge" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    PRIMARY KEY ("height", "miner_id", "sector_id", "state_root")
);

CREATE TABLE miner_sector_deals
(
    "height" BIGINT NOT NULL,
    "miner_id" TEXT NOT NULL,
    "sector_id" BIGINT NOT NULL,
    "deal_id" BIGINT NOT NULL,
    PRIMARY KEY ("height", "miner_id", "sector_id", "deal_id")
);

CREATE TYPE miner_sector_event_type AS ENUM
(
    'PRECOMMIT_ADDED',
    'PRECOMMIT_EXPIRED',
    'COMMIT_CAPACITY_ADDED',
    'SECTOR_ADDED',
    'SECTOR_EXTENDED',
    'SECTOR_EXPIRED',
    'SECTOR_FAULTED',
    'SECTOR_RECOVERING',
    'SECTOR_RECOVERED',
    'SECTOR_TERMINATED'
);

CREATE TABLE miner_sector_events
(
    "height" BIGINT NOT NULL,
    "miner_id" TEXT NOT NULL,
    "sector_id" BIGINT NOT NULL,
    "state_root" TEXT NOT NULL,
    "event" miner_sector_event_type NOT NULL,
    PRIMARY KEY ("height", "sector_id", "event", "miner_id", "state_root")
);

CREATE TABLE miner_sector_posts
(
    "height" BIGINT NOT NULL,
    "miner_id" TEXT NOT NULL,
    "sector_id" BIGINT NOT NULL,
    "post_message_cid" TEXT,
    PRIMARY KEY ("height", "miner_id", "sector_id")
);

CREATE TABLE miner_sector_faults
(
    "height" BIGINT NOT NULL,
    "miner_id" TEXT NOT NULL,
    "sector_id" BIGINT NOT NULL
);

CREATE TABLE market_deal_states
(
    deal_id BIGINT NOT NULL,
    state_root TEXT NOT NULL,
    sector_start_epoch BIGINT NOT NULL,
    last_update_epoch BIGINT NOT NULL,
    slash_epoch BIGINT NOT NULL,
    height BIGINT NOT NULL,
    PRIMARY KEY ("deal_id")
);

CREATE TABLE market_deal_proposals
(
    deal_id BIGINT NOT NULL,
    state_root TEXT NOT NULL,
    piece_cid TEXT NOT NULL,
    padded_piece_size BIGINT NOT NULL,
    unpadded_piece_size BIGINT NOT NULL,
    is_verified boolean NOT NULL,
    client_id TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    start_epoch BIGINT NOT NULL,
    end_epoch BIGINT NOT NULL,
    slashed_epoch BIGINT,
    storage_price_per_epoch TEXT NOT NULL,
    provider_collateral TEXT NOT NULL,
    client_collateral TEXT NOT NULL,
    label TEXT,
    height BIGINT NOT NULL,
    PRIMARY KEY ("deal_id")
);

CREATE TABLE chain_economics
(
    parent_state_root TEXT NOT NULL,
    circulating_fil TEXT NOT NULL,
    vested_fil TEXT NOT NULL,
    mined_fil TEXT NOT NULL,
    burnt_fil TEXT NOT NULL,
    locked_fil TEXT NOT NULL,
    PRIMARY KEY ("parent_state_root")
);

CREATE TABLE chain_powers
(
    state_root TEXT NOT NULL,
    total_raw_bytes_power TEXT NOT NULL,
    total_raw_bytes_committed TEXT NOT NULL,
    total_qa_bytes_power TEXT NOT NULL,
    total_qa_bytes_committed TEXT NOT NULL,
    total_pledge_collateral TEXT NOT NULL,
    qa_smoothed_position_estimate TEXT NOT NULL,
    qa_smoothed_velocity_estimate TEXT NOT NULL,
    miner_count BIGINT,
    participating_miner_count BIGINT,
    height BIGINT NOT NULL,
    PRIMARY KEY ("height", "state_root")
);

CREATE TABLE power_actor_claims
(
    "height" BIGINT NOT NULL,
    "miner_id" TEXT NOT NULL,
    "state_root" TEXT NOT NULL,
    "raw_byte_power" TEXT NOT NULL,
    "quality_adj_power" TEXT NOT NULL,
    PRIMARY KEY ("height", "miner_id", "state_root")
);

CREATE TABLE actors
(
    id TEXT NOT NULL,
    code TEXT NOT NULL,
    head TEXT NOT NULL,
    nonce BIGINT NOT NULL,
    balance TEXT NOT NULL,
    state_root TEXT NOT NULL,
    height BIGINT NOT NULL,
    PRIMARY KEY ("height", "id", "state_root")
);

CREATE TABLE parsed_till
(
    "height" BIGINT NOT NULL,
    PRIMARY KEY ("height")
);

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

DROP TABLE transactions;
DROP TABLE block_headers;
DROP TABLE block_messages;
DROP TABLE parsed_messages;
DROP TABLE message_gas_economy;
DROP TABLE receipts;
DROP TABLE miner_infos;
DROP TABLE miner_funds;
DROP TABLE miner_fee_debts;
DROP TABLE miner_current_deadline_infos;
DROP TABLE miner_sector_infos;
DROP TABLE miner_sector_deals;
DROP TABLE miner_sector_events;
DROP TABLE miner_sector_posts;
DROP TABLE miner_sector_faults;
DROP TYPE IF EXISTS miner_sector_event_type;
DROP TABLE market_deal_proposals;
DROP TABLE chain_economics;
DROP TABLE chain_powers;
DROP TABLE power_actor_claims;
DROP TABLE actors;
DROP TABLE parsed_till;
DROP TABLE miner_transactions;
