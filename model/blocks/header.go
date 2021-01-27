package blocks

import (
	"context"

	"github.com/filecoin-project/lotus/chain/types"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

type BlockHeader struct {
	tableName       struct{} `pg:"block_headers"` // nolint: structcheck,unused
	Height          int64    `pg:",pk,use_zero,notnull"`
	Cid             string   `pg:",pk,notnull"`
	MinerID         string   `pg:",notnull"`
	ParentWeight    string   `pg:",notnull"`
	ParentBaseFee   string   `pg:",notnull"`
	ParentStateRoot string   `pg:",notnull"`

	WinCount      int64  `pg:",use_zero"`
	Timestamp     uint64 `pg:",use_zero"`
	ForkSignaling uint64 `pg:",use_zero"`
}

func NewBlockHeader(bh *types.BlockHeader) BlockHeader {
	return BlockHeader{
		Cid:             bh.Cid().String(),
		MinerID:         bh.Miner.String(),
		ParentWeight:    bh.ParentWeight.String(),
		ParentBaseFee:   bh.ParentBaseFee.String(),
		ParentStateRoot: bh.ParentStateRoot.String(),
		Height:          int64(bh.Height),
		WinCount:        bh.ElectionProof.WinCount,
		Timestamp:       bh.Timestamp,
		ForkSignaling:   bh.ForkSignaling,
	}
}

func (bh *BlockHeader) Persist(ctx context.Context, s model.StorageBatch) error {
	return s.PersistModel(ctx, bh)
}

type BlockHeaders []*BlockHeader

func (bhl BlockHeaders) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(bhl) == 0 {
		return nil
	}
	return s.PersistModel(ctx, bhl)
}
