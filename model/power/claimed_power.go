package power

import (
	"context"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

type PowerActorClaim struct {
	Height          int64  `pg:",pk,notnull,use_zero"`
	MinerID         string `pg:",pk,notnull"`
	StateRoot       string `pg:",pk,notnull"`
	RawBytePower    string `pg:",notnull"`
	QualityAdjPower string `pg:",notnull"`
}

func (p *PowerActorClaim) Persist(ctx context.Context, s model.StorageBatch) error {
	return s.PersistModel(ctx, p)
}

type PowerActorClaimList []*PowerActorClaim

func (pl PowerActorClaimList) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(pl) == 0 {
		return nil
	}
	return s.PersistModel(ctx, pl)
}
