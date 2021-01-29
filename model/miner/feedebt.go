package miner

import (
	"context"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

type MinerFeeDebt struct {
	Height    int64  `pg:",pk,notnull,use_zero"`
	MinerID   string `pg:",pk,notnull"`
	StateRoot string `pg:",pk,notnull"`

	FeeDebt string `pg:",notnull"`
}

func (m *MinerFeeDebt) Persist(ctx context.Context, s model.StorageBatch) error {
	return s.PersistModel(ctx, m)
}

type MinerFeeDebtList []*MinerFeeDebt

func (ml MinerFeeDebtList) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(ml) == 0 {
		return nil
	}
	return s.PersistModel(ctx, ml)
}
