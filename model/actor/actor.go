package actor

import (
	"context"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

type Actor struct {
	Height    int64  `pg:",pk,notnull,use_zero"`
	ID        string `pg:",pk,notnull"`
	StateRoot string `pg:",pk,notnull"`
	Code      string `pg:",notnull"`
	Head      string `pg:",notnull"`
	Balance   string `pg:",notnull"`
	Nonce     uint64 `pg:",use_zero"`
}

func (a *Actor) Persist(ctx context.Context, s model.StorageBatch) error {
	return s.PersistModel(ctx, a)
}

type ActorList []*Actor

func (actors ActorList) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(actors) == 0 {
		return nil
	}
	return s.PersistModel(ctx, actors)
}
