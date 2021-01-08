package chain

import (
	"context"
	"fmt"
	"log"

	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/filecoin-project/go-state-types/abi"

	// "github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"golang.org/x/xerrors"
)

// func NewWalker(api *apistruct.FullNodeStruct, obs TipSetObserver, minHeight, maxHeight int64) *Walker {
func NewWalker(opener lens.APIOpener, obs TipSetObserver, minHeight, maxHeight int64) *Walker {
	return &Walker{
		// lens:      baselens.Lens{},
		// api:       api,
		opener:    opener,
		obs:       obs,
		finality:  900,
		minHeight: minHeight,
		maxHeight: maxHeight,
	}
}

type Walker struct {
	// lens      baselens.Lens
	opener lens.APIOpener
	// api       *apistruct.FullNodeStruct
	obs       TipSetObserver
	finality  int   // epochs after which chain state is considered final
	minHeight int64 // limit persisting to tipsets equal to or above this height
	maxHeight int64 // limit persisting to tipsets equal to or below this height}
}

func (c *Walker) Run(ctx context.Context) error {
	node, closer, err := c.opener.Open(ctx)
	if err != nil {
		return xerrors.Errorf("open lens: %w", err)
	}
	defer closer()

	ts, err := node.ChainHead(ctx)
	if err != nil {
		return xerrors.Errorf("get chain head: %w", err)
	}
	fmt.Println("after chainHead func call")

	if int64(ts.Height()) < c.minHeight {
		fmt.Println("int64(ts.Height()) < c.minHeight")
		return xerrors.Errorf("cannot walk history, chain head (%d) is earlier than minimum height (%d)", int64(ts.Height()), c.minHeight)
	}

	if int64(ts.Height()) > c.maxHeight {
		fmt.Println("int64(ts.Height()) > c.maxHeight")

		ts, err = node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(c.maxHeight), types.EmptyTSK)
		if err != nil {
			return xerrors.Errorf("get tipset by height: %w", err)
		}
		fmt.Println("now ts", ts)
	}

	if err := c.WalkChain(ctx, node, ts); err != nil {
		return xerrors.Errorf("walk chain: %w", err)
	}

	return nil
}

func (c *Walker) WalkChain(ctx context.Context, node lens.API, ts *types.TipSet) error {

	fmt.Println("in walkchain")
	log.Println("found tipset", "height", ts.Height())
	if err := c.obs.TipSet(ctx, ts); err != nil {
		return xerrors.Errorf("notify tipset: %w", err)
	}

	var err error
	for int64(ts.Height()) >= c.minHeight && ts.Height() > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ts, err = node.ChainGetTipSet(ctx, ts.Parents())
		if err != nil {
			return xerrors.Errorf("get tipset: %w", err)
		}

		log.Println("found tipset", "height", ts.Height())
		if err := c.obs.TipSet(ctx, ts); err != nil {
			return xerrors.Errorf("notify tipset: %w", err)
		}
	}

	return nil
}
