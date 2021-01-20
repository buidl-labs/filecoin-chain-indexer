-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE transactions
(
    "cid" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "from_addr" TEXT NOT NULL,
    "to_addr" TEXT NOT NULL,
    "amount" TEXT NOT NULL,
    "type" INT,
    "gas_fee_cap" TEXT NOT NULL,
    "gas_premium" TEXT NOT NULL,
    "gas_limit" BIGINT,
    "size_bytes" BIGINT,
    "nonce" BIGINT,
    "method" BIGINT,
    "method_name" TEXT,
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
    PRIMARY KEY ("cid")
);

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
    "timestamp" BIGINT NOT NULL
);

CREATE TABLE block_messages
(
    "block" TEXT NOT NULL,
    "message" TEXT NOT NULL,
    "height" BIGINT NOT NULL
);

CREATE TABLE receipts
(
    "message" TEXT NOT NULL,
    "state_root" TEXT NOT NULL,
    "idx" BIGINT NOT NULL,
    "exit_code" BIGINT NOT NULL,
    "gas_used" BIGINT NOT NULL,
    "height" BIGINT NOT NULL
);

CREATE TABLE miner_infos
(
    "miner_id" TEXT NOT NULL,
    "address" TEXT,
    "peer_id" TEXT NOT NULL,
    "owner_id" TEXT NOT NULL,
    "worker_id" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "storage_ask_price" TEXT NOT NULL,
    "min_piece_size" BIGINT NOT NULL,
    "max_piece_size" BIGINT NOT NULL
);

CREATE TABLE miner_funds
(
    "miner_id" TEXT NOT NULL,
    "height" BIGINT NOT NULL,
    "state_root" TEXT NOT NULL,
    "locked_funds" TEXT NOT NULL,
    "initial_pledge" TEXT NOT NULL,
    "pre_commit_deposits" TEXT NOT NULL,
    "available_balance" TEXT NOT NULL
);

CREATE TABLE miner_quality
(
    "miner_id" TEXT NOT NULL,
    "quality_adj_power" TEXT NOT NULL,
    "raw_byte_power" TEXT NOT NULL,
    "win_count" BIGINT NOT NULL ,
    "data_stored" TEXT NOT NULL,
    "blocks_mined" BIGINT NOT NULL,
    "mining_efficiency" TEXT NOT NULL,
    "faulty_sectors" BIGINT NOT NULL,
    "height" BIGINT NOT NULL
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
    "deal_id" BIGINT NOT NULL
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
    "miner_id" TEXT NOT NULL,
    "sector_id" BIGINT NOT NULL,
    "state_root" TEXT NOT NULL,
    "event" miner_sector_event_type NOT NULL
);

CREATE TABLE miner_sector_posts
(
    "height" BIGINT NOT NULL,
    "miner_id" TEXT NOT NULL,
    "sector_id" BIGINT NOT NULL,
    "post_message_cid" TEXT
);

CREATE TABLE miner_sector_faults
(
    "height" BIGINT NOT NULL,
    "miner_id" TEXT NOT NULL,
    "sector_id" BIGINT NOT NULL
);

CREATE TABLE market_deal_proposals
(
    deal_id bigint NOT NULL,
    state_root text NOT NULL,
    piece_cid text NOT NULL,
    padded_piece_size bigint NOT NULL,
    unpadded_piece_size bigint NOT NULL,
    is_verified boolean NOT NULL,
    client_id text NOT NULL,
    provider_id text NOT NULL,
    start_epoch bigint NOT NULL,
    end_epoch bigint NOT NULL,
    slashed_epoch bigint,
    storage_price_per_epoch text NOT NULL,
    provider_collateral text NOT NULL,
    client_collateral text NOT NULL,
    label text,
    height bigint NOT NULL
);

CREATE TABLE parsed_till
(
    "height" BIGINT NOT NULL
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE transactions;
DROP TABLE block_headers;
DROP TABLE block_messages;
DROP TABLE receipts;
DROP TABLE miner_infos;
DROP TABLE miner_funds;
DROP TABLE miner_quality;
DROP TABLE miner_current_deadline_infos;
DROP TABLE miner_sector_infos;
DROP TABLE miner_sector_deals;
DROP TABLE miner_sector_events;
DROP TABLE miner_sector_posts;
DROP TABLE miner_sector_faults;
DROP TYPE IF EXISTS miner_sector_event_type;
DROP TABLE market_deal_proposals;
DROP TABLE parsed_till;
