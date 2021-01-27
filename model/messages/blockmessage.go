package messages

import (
	"context"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

type BlockMessage struct {
	tableName struct{} `pg:"block_messages"` // nolint: structcheck,unused
	Height    int64    `pg:",pk,notnull,use_zero"`
	Block     string   `pg:",pk,notnull"`
	Message   string   `pg:",pk,notnull"`
}

func (bm *BlockMessage) Persist(ctx context.Context, s model.StorageBatch) error {
	return s.PersistModel(ctx, bm)
}

type BlockMessages []*BlockMessage

func (bms BlockMessages) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(bms) == 0 {
		return nil
	}

	return s.PersistModel(ctx, bms)
}
