package chain

import (
	"context"

	"github.com/filecoin-project/lotus/chain/types"
	log "github.com/sirupsen/logrus"

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
	var blockHeadersResults []*blocksmodel.BlockHeader
	for _, bh := range ts.Blocks() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		log.Info("thisbh", bh)
		// pl = append(pl, blocksmodel.NewBlockHeader(bh))
		blockHeadersResults = append(blockHeadersResults, blocksmodel.NewBlockHeader(bh))
		// pl = append(pl, blocksmodel.NewBlockParents(bh))
		// pl = append(pl, blocksmodel.NewDrandBlockEntries(bh))
	}
	log.Info("bhresults", blockHeadersResults)

	// p.store.PersistBlockHeaders(blockHeadersResults)
	r, err := p.store.DB.Model(&blockHeadersResults).Insert()
	if err != nil {
		log.Info("insert bhr", err)
	} else {
		log.Info("inserted bhr", r)
	}

	return pl, nil
}

func (p *BlockProcessor) Close() error {
	return nil
}
