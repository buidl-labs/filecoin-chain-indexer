package chain

import (
	"context"

	"github.com/filecoin-project/lotus/chain/types"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
	"github.com/buidl-labs/filecoin-chain-indexer/model/blocks"
)

type BlockProcessor struct {
}

func NewBlockProcessor() *BlockProcessor {
	return &BlockProcessor{}
}

func (p *BlockProcessor) ProcessTipSet(ctx context.Context, ts *types.TipSet) (model.Persistable, error) {
	var pl model.PersistableList
	for _, bh := range ts.Blocks() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		pl = append(pl, blocks.NewBlockHeader(bh))
		// pl = append(pl, blocks.NewBlockParents(bh))
		// pl = append(pl, blocks.NewDrandBlockEntries(bh))
	}

	// report := &visormodel.ProcessingReport{
	// 	Height:    int64(ts.Height()),
	// 	StateRoot: ts.ParentState().String(),
	// }

	return pl, nil
}

func (p *BlockProcessor) Close() error {
	return nil
}
