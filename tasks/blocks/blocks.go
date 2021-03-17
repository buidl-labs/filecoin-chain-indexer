package blocks

import (
	"context"

	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"

	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	blocksmodel "github.com/buidl-labs/filecoin-chain-indexer/model/blocks"
)

var log = logging.Logger("blocks")

type Task struct {
	store db.Store
}

func NewTask(store db.Store) *Task {
	return &Task{
		store: store,
	}
}

func (p *Task) ProcessTipSet(ctx context.Context, ts *types.TipSet) (model.Persistable, error) {
	var pl model.PersistableList
	var blockHeadersResults []*blocksmodel.BlockHeader
	for _, bh := range ts.Blocks() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		blockHeadersResults = append(blockHeadersResults, blocksmodel.NewBlockHeader(bh))
	}

	_, err := p.store.DB.Model(&blockHeadersResults).Insert()
	if err != nil {
		log.Errorw("inserting block_headers",
			"error", err,
			"tipset", ts.Height(),
		)
	} else {
		log.Debugw(
			"inserted block_headers",
			"tipset", ts.Height(),
			"count", len(blockHeadersResults),
		)
	}

	return pl, nil
}

func (p *Task) Close() error {
	return nil
}
