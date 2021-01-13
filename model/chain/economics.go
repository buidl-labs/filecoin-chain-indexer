package chain

import (
	"context"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

type ChainEconomics struct {
	tableName       struct{} `pg:"chain_economics"` // nolint: structcheck,unused
	ParentStateRoot string   `pg:",notnull"`
	CirculatingFil  string   `pg:",notnull"`
	VestedFil       string   `pg:",notnull"`
	MinedFil        string   `pg:",notnull"`
	BurntFil        string   `pg:",notnull"`
	LockedFil       string   `pg:",notnull"`
}

func (c *ChainEconomics) Persist(ctx context.Context, s model.StorageBatch) error {
	return s.PersistModel(ctx, c)
}

type ChainEconomicsList []*ChainEconomics

func (l ChainEconomicsList) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(l) == 0 {
		return nil
	}

	return s.PersistModel(ctx, l)
}
