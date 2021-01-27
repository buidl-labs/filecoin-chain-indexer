package miner

import (
	"context"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

type MinerCurrentDeadlineInfo struct {
	tableName struct{} `pg:"miner_current_deadline_infos"` // nolint: structcheck,unused
	Height    int64    `pg:",pk,notnull,use_zero"`
	MinerID   string   `pg:",pk,notnull"`
	StateRoot string   `pg:",pk,notnull"`

	DeadlineIndex uint64 `pg:",notnull,use_zero"`
	PeriodStart   int64  `pg:",notnull,use_zero"`
	Open          int64  `pg:",notnull,use_zero"`
	Close         int64  `pg:",notnull,use_zero"`
	Challenge     int64  `pg:",notnull,use_zero"`
	FaultCutoff   int64  `pg:",notnull,use_zero"`
}

func (m *MinerCurrentDeadlineInfo) Persist(ctx context.Context, s model.StorageBatch) error {
	return s.PersistModel(ctx, m)
}

type MinerCurrentDeadlineInfoList []*MinerCurrentDeadlineInfo

func (ml MinerCurrentDeadlineInfoList) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(ml) == 0 {
		return nil
	}
	return s.PersistModel(ctx, ml)
}
