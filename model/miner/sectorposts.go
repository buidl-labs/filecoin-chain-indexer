package miner

import (
	"context"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

type MinerSectorPost struct {
	tableName struct{} `pg:"miner_sector_posts"` // nolint: structcheck,unused
	Height    int64    `pg:",pk,notnull,use_zero"`
	MinerID   string   `pg:",pk,notnull"`
	SectorID  uint64   `pg:",pk,notnull,use_zero"`

	PostMessageCID string
}

type MinerSectorPostList []*MinerSectorPost

func (msp *MinerSectorPost) Persist(ctx context.Context, s model.StorageBatch) error {
	return s.PersistModel(ctx, msp)
}

func (ml MinerSectorPostList) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(ml) == 0 {
		return nil
	}
	return s.PersistModel(ctx, ml)
}
