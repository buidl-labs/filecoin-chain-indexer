package miner

import (
	"context"

	// filbig "github.com/filecoin-project/go-state-types/big"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

type MinerSectorInfo struct {
	tableName struct{} `pg:"miner_sector_infos"` // nolint: structcheck,unused
	Height    int64    `pg:",pk,notnull,use_zero"`
	MinerID   string   `pg:",pk,notnull"`
	SectorID  uint64   `pg:",pk,use_zero"`
	StateRoot string   `pg:",pk,notnull"`

	SealedCID string `pg:",notnull"`

	ActivationEpoch int64 `pg:",use_zero"`
	ExpirationEpoch int64 `pg:",use_zero"`

	DealWeight         string `pg:",notnull"`
	VerifiedDealWeight string `pg:",notnull"`

	InitialPledge         string `pg:",notnull"`
	ExpectedDayReward     string `pg:",notnull"`
	ExpectedStoragePledge string `pg:",notnull"`
}

func (msi *MinerSectorInfo) Persist(ctx context.Context, s model.StorageBatch) error {
	return s.PersistModel(ctx, msi)
}

type MinerSectorInfoList []*MinerSectorInfo

func (ml MinerSectorInfoList) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(ml) == 0 {
		return nil
	}
	return s.PersistModel(ctx, ml)
}
