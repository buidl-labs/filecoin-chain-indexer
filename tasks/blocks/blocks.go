package blocks

import (
	"context"

	// "github.com/filecoin-project/go-address"
	// "github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"

	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	blocksmodel "github.com/buidl-labs/filecoin-chain-indexer/model/blocks"
)

var log = logging.Logger("blocks")

type Task struct {
	node   lens.API
	opener lens.APIOpener
	closer lens.APICloser
	store  db.Store
	cfg    config.Config
}

func NewTask(opener lens.APIOpener, store db.Store) *Task {
	return &Task{
		opener: opener,
		store:  store,
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
