package chain

import (
	"context"
	"io/ioutil"
	"os"
	"strconv"
	"sync"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
)

type TaskType int

const (
	EPOCHS_RANGE TaskType = iota
	SINGLE_EPOCH TaskType = iota
)

func NewWalker(opener lens.APIOpener, obs TipSetObserver, tasks []string, taskType TaskType, minHeight, maxHeight int64, cfg config.Config) *Walker {
	return &Walker{
		opener:    opener,
		obs:       obs,
		finality:  900,
		minHeight: minHeight,
		maxHeight: maxHeight,
		tasks:     tasks,
		taskType:  taskType,
		cfg:       cfg,
	}
}

type Walker struct {
	opener    lens.APIOpener
	obs       TipSetObserver
	finality  int   // epochs after which chain state is considered final
	minHeight int64 // limit persisting to tipsets equal to or above this height
	maxHeight int64 // limit persisting to tipsets equal to or below this height}
	tasks     []string
	taskType  TaskType
	cfg       config.Config
}

func (c *Walker) Run(ctx context.Context) error {
	node, closer, err := c.opener.Open(ctx)
	if err != nil {
		return xerrors.Errorf("open lens: %w", err)
	}
	defer closer()

	var mints, maxts *types.TipSet
	ts, err := node.ChainHead(ctx)
	if err != nil {
		return xerrors.Errorf("get chain head: %w", err)
	}

	if c.taskType == SINGLE_EPOCH {
		for _, task := range c.tasks {
			if task == ActorCodesTask {
				ts, err := node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(ts.Height()-900), types.EmptyTSK)
				if err != nil {
					return xerrors.Errorf("get tipset by height: %w", err)
				}
				log.Debugw("found tipset", "height", ts.Height())
				if err := c.obs.TipSet(ctx, ts); err != nil {
					return xerrors.Errorf("notify tipset: %w", err)
				}
			} else {
				var singleEpochTs *types.TipSet
				if c.cfg.Epoch != -1 {
					if c.cfg.Epoch <= int64(ts.Height())-900 {
						singleEpochTs, err = node.ChainGetTipSetByHeight(context.Background(), abi.ChainEpoch(c.cfg.Epoch), types.EmptyTSK)
						if err != nil {
							return xerrors.Errorf("get singleepoch tipset by height: %w", err)
						}
					} else {
						singleEpochTs = ts
					}
				} else {
					singleEpochTs = ts
				}
				singleEpochTs, err = node.ChainGetTipSetByHeight(context.Background(), singleEpochTs.Height()-900, types.EmptyTSK)
				if err != nil {
					return xerrors.Errorf("get singleepoch tipset by height: %w", err)
				}
				log.Debugw("found tipset", "height", singleEpochTs.Height())
				if err := c.obs.TipSet(ctx, singleEpochTs); err != nil {
					return xerrors.Errorf("notify tipset: %w", err)
				}
			}
		}
	} else if c.taskType == EPOCHS_RANGE {
		log.Debugw("taskType 0 (EPOCHS_RANGE): found tipset", "height", ts.Height())

		if int64(ts.Height()) < c.minHeight {
			return xerrors.Errorf("cannot walk history, chain head (%d) is earlier than minimum height (%d)", int64(ts.Height()), c.minHeight)
		}

		if int64(ts.Height()) > c.maxHeight {
			log.Debug("int64(ts.Height()) > c.maxHeight")
			maxts, err = node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(c.maxHeight), types.EmptyTSK)
			if err != nil {
				return xerrors.Errorf("get tipset by height: %w", err)
			}
			// log.Debugw("maxts", "height", maxts)
			mints, err = node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(c.minHeight), types.EmptyTSK)
			if err != nil {
				return xerrors.Errorf("get tipset by height: %w", err)
			}
			// log.Debugw("mints", "height", mints)
		}

		if err := c.WalkChain(ctx, node, mints, maxts); err != nil {
			return xerrors.Errorf("walk chain: %w", err)
		}
	}

	return nil
}

func (c *Walker) WalkChain(ctx context.Context, node lens.API, mints, maxts *types.TipSet) error {
	log.Debugw("in WalkChain, found tipset", "height", maxts.Height())

	for _, task := range c.tasks {
		if task == MinerTxnsTask {
			log.Debug("minertxnstask", maxts.Height())
			x := 120
			var wg sync.WaitGroup
			l := int(c.maxHeight - c.minHeight) // + 1
			if l < 120 {
				x = l
			}
			rem := l % x
			log.Debug("maxh", c.maxHeight, "minh", c.minHeight, "Lhh", l)
			for j := int(c.minHeight); j < int(c.maxHeight)-rem && j+x <= int(c.maxHeight); j += x {
				// wg.Add(x)
				for i := j; i < j+x; i++ {
					wg.Add(1)
					go worker(c, node, ctx, &wg, i)
				}
				wg.Wait()
			}
			var wg2 sync.WaitGroup
			// wg2.Add(rem)
			for i := int(c.maxHeight) - rem; i <= int(c.maxHeight); i++ {
				wg2.Add(1)
				go worker(c, node, ctx, &wg2, i)
			}
			wg2.Wait()
		} else {
			if c.cfg.IndexForever == 1 {
				// Do forward indexing if forever

				maxheightB, err := ioutil.ReadFile(os.Getenv("ACS_PARSEDTILL"))
				if err != nil {
					return xerrors.Errorf("read acsparsedtill: %w", err)
				}
				maxheightStr := string(maxheightB)
				maxheight, _ := strconv.ParseInt(maxheightStr, 10, 64)

				head, err := node.ChainHead(context.Background())
				if err != nil {
					return xerrors.Errorf("get chain head: %w", err)
				}
				if maxheight >= int64(head.Height())-900 {
					maxheight = int64(head.Height()) - 900
				}

				for mints.Height() < abi.ChainEpoch(c.minHeight) {
					mints, err = node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(mints.Height()+2), types.EmptyTSK)
					if err != nil {
						return xerrors.Errorf("get next mintipset: %w", err)
					}
				}

				log.Debugw("found tipset", "height", mints.Height())
				if err := c.obs.TipSet(ctx, mints); err != nil {
					return xerrors.Errorf("notify tipset: %w", err)
				}

				currTs := mints

				for int64(currTs.Height()) >= c.minHeight {
					if (task == MessagesTask && int64(currTs.Height()) <= maxheight) ||
						(task != MessagesTask) {

						oldHt := currTs.Height()
						currTs, err = node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(currTs.Height()+1), types.EmptyTSK)
						if err != nil {
							return xerrors.Errorf("get next tipset: %w", err)
						}

						// Skip empty tipset
						// For example, tipset 563097 in mainnet
						for currTs.Height() == oldHt {
							currTs, err = node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(currTs.Height()+2), types.EmptyTSK)
							if err != nil {
								return xerrors.Errorf("get next tipset: %w", err)
							}
						}
						log.Debugw("found tipset", "height", currTs.Height())
						if err := c.obs.TipSet(ctx, currTs); err != nil {
							return xerrors.Errorf("notify tipset: %w", err)
						}

						// update maxheight to latest value of ACS_PARSEDTILL
						maxheightB, err := ioutil.ReadFile(os.Getenv("ACS_PARSEDTILL"))
						if err != nil {
							return xerrors.Errorf("read acsparsedtill: %w", err)
						}
						maxheightStr := string(maxheightB)
						maxheight, _ = strconv.ParseInt(maxheightStr, 10, 64)

						head, err := node.ChainHead(context.Background())
						if err != nil {
							return xerrors.Errorf("get chain head: %w", err)
						}
						if maxheight >= int64(head.Height())-900 {
							maxheight = int64(head.Height()) - 900
						}
					}
				}
			} else {
				// Do backward indexing

				log.Debugw("found tipset", "height", maxts.Height())
				if err := c.obs.TipSet(ctx, maxts); err != nil {
					return xerrors.Errorf("notify tipset: %w", err)
				}

				// maxheightB, err := ioutil.ReadFile(os.Getenv("ACS_PARSEDTILL"))
				// if err != nil {
				// 	return xerrors.Errorf("read acsparsedtill: %w", err)
				// }

				// maxheightStr := string(maxheightB)
				// maxheight, _ := strconv.ParseInt(maxheightStr, 10, 64)

				maxheight := int64(maxts.Height())
				head, err := node.ChainHead(context.Background())
				if err != nil {
					return xerrors.Errorf("get chain head: %w", err)
				}
				if c.maxHeight >= int64(head.Height())-900 {
					maxheight = int64(head.Height())
				}

				for int64(maxts.Height()) >= c.minHeight && maxts.Height() > 0 {
					if (task == MessagesTask && int64(maxts.Height()) <= maxheight) ||
						(task != MessagesTask) {

						maxts, err = node.ChainGetTipSet(ctx, maxts.Parents())
						if err != nil {
							return xerrors.Errorf("get tipset: %w", err)
						}

						log.Debugw("found tipset", "height", maxts.Height())
						if err := c.obs.TipSet(ctx, maxts); err != nil {
							return xerrors.Errorf("notify tipset: %w", err)
						}
					}
				}
			}
		}
	}
	return nil
}

func worker(c *Walker, node lens.API, ctx context.Context, wg *sync.WaitGroup, i int) {
	// defer wg.Done()
	log.Debug("currHeight", i, " starting")
	ts, err := node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(i), types.EmptyTSK)
	if err != nil {
		log.Errorw("get tipset by height",
			"height", i,
			"error", err,
		)
		wg.Done()
		// return xerrors.Errorf("get tipset: %w", err)
	} else {
		log.Debugw("found tipset", "height", ts.Height())
		if err := c.obs.TipSet(ctx, ts); err != nil {
			log.Errorw("notify tipset",
				"height", ts.Height(),
				"error", err,
			)
			wg.Done()
			// return xerrors.Errorf("notify tipset: %w", err)
		} else {
			log.Debugw("notified tipset",
				"height", ts.Height(),
			)
			wg.Done()
		}
	}
}
