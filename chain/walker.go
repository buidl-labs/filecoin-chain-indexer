package chain

import (
	"context"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/lens"
)

func NewWalker(opener lens.APIOpener, obs TipSetObserver, tasks []string, taskType int, minHeight, maxHeight int64) *Walker {
	return &Walker{
		opener:    opener,
		obs:       obs,
		finality:  900,
		minHeight: minHeight,
		maxHeight: maxHeight,
		tasks:     tasks,
		taskType:  taskType,
	}
}

type Walker struct {
	opener    lens.APIOpener
	obs       TipSetObserver
	finality  int   // epochs after which chain state is considered final
	minHeight int64 // limit persisting to tipsets equal to or above this height
	maxHeight int64 // limit persisting to tipsets equal to or below this height}
	tasks     []string
	taskType  int
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
	log.Info("after ChainHead func call")

	if c.taskType == 1 {
		log.Info("taskType 1 (currentEpochTasks): found tipset", "height", ts.Height())
		if err := c.obs.TipSet(ctx, ts); err != nil {
			return xerrors.Errorf("notify tipset: %w", err)
		}
	} else {
		log.Info("taskType 0 (allEpochsTasks): found tipset", "height", ts.Height())

		if int64(ts.Height()) < c.minHeight {
			log.Info("int64(ts.Height()) < c.minHeight")
			return xerrors.Errorf("cannot walk history, chain head (%d) is earlier than minimum height (%d)", int64(ts.Height()), c.minHeight)
		}

		if int64(ts.Height()) > c.maxHeight {
			log.Info("int64(ts.Height()) > c.maxHeight")
			ts, err = node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(c.maxHeight), types.EmptyTSK)
			if err != nil {
				return xerrors.Errorf("get tipset by height: %w", err)
			}
			log.Info("ts by height", ts)
		}

		if err := c.WalkChain(ctx, node, ts); err != nil {
			return xerrors.Errorf("walk chain: %w", err)
		}
	}

	return nil
}

func (c *Walker) WalkChain(ctx context.Context, node lens.API, ts *types.TipSet) error {
	log.Info("in WalkChain", "found tipset", "height", ts.Height())
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

		log.Info("found tipset", "height", ts.Height())
		if err := c.obs.TipSet(ctx, ts); err != nil {
			return xerrors.Errorf("notify tipset: %w", err)
		}
	}

	return nil
}
