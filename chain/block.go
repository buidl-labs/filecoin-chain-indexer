package chain

import (
	"context"
	"fmt"

	"github.com/filecoin-project/lotus/chain/types"

	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	blocksmodel "github.com/buidl-labs/filecoin-chain-indexer/model/blocks"
)

type BlockProcessor struct {
	store db.Store
}

func NewBlockProcessor(store db.Store) *BlockProcessor {
	return &BlockProcessor{
		store: store,
	}
}

func (p *BlockProcessor) ProcessTipSet(ctx context.Context, ts *types.TipSet) (model.Persistable, error) {
	var pl model.PersistableList
	var blockHeadersResults blocksmodel.BlockHeaders
	for _, bh := range ts.Blocks() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		pl = append(pl, blocksmodel.NewBlockHeader(bh))
		blockHeadersResults = append(blockHeadersResults, blocksmodel.NewBlockHeader(bh))
		// pl = append(pl, blocksmodel.NewBlockParents(bh))
		// pl = append(pl, blocksmodel.NewDrandBlockEntries(bh))
	}

	fmt.Println("blockHeadersResults", blockHeadersResults)
	for _, bhr := range blockHeadersResults {
		p.store.PersistBlockHeaders(*bhr)
		fmt.Println("bhr: miner", bhr.Miner, "wincount", bhr.WinCount)
	}

	return pl, nil
}

func (p *BlockProcessor) Close() error {
	return nil
}
