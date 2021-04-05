package chain

import (
	"context"
	"database/sql"
	"time"

	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	"github.com/buidl-labs/filecoin-chain-indexer/tasks/actorcodes"
	"github.com/buidl-labs/filecoin-chain-indexer/tasks/blocks"
	"github.com/buidl-labs/filecoin-chain-indexer/tasks/market"
	"github.com/buidl-labs/filecoin-chain-indexer/tasks/messages"
	"github.com/buidl-labs/filecoin-chain-indexer/tasks/miner"
)

const (
	MinerTxnsTask  = "minertxns"
	ActorCodesTask = "actorcodes"
	MinersTask     = "miners"
	BlocksTask     = "blocks"
	MessagesTask   = "messages"
	MarketsTask    = "markets"
)

var log = logging.Logger("chain")

var _ TipSetObserver = (*TipSetIndexer)(nil)

// TipSetIndexer waits for tipsets and persists their block data into a database.
type TipSetIndexer struct {
	db                *sql.DB
	store             db.Store
	storage           model.Storage
	window            time.Duration
	name              string
	persistSlot       chan struct{}
	processors        map[string]TipSetProcessor
	messageProcessors map[string]MessageProcessor
	actorProcessors   map[string]ActorProcessor
	lastTipSet        *types.TipSet
	node              lens.API
	opener            lens.APIOpener
	closer            lens.APICloser
	cfg               config.Config
}

func NewTipSetIndexer(o lens.APIOpener, db *sql.DB, store db.Store, s model.Storage, window time.Duration, name string, tasks []string, cfg config.Config) (*TipSetIndexer, error) {
	tsi := &TipSetIndexer{
		db:                db,
		store:             store,
		storage:           s,
		window:            window,
		name:              name,
		persistSlot:       make(chan struct{}, 1), // allow one concurrent persistence job
		processors:        map[string]TipSetProcessor{},
		messageProcessors: map[string]MessageProcessor{},
		actorProcessors:   map[string]ActorProcessor{},
		opener:            o,
		cfg:               cfg,
	}

	for _, task := range tasks {
		switch task {
		case MinerTxnsTask:
			tsi.processors[MinerTxnsTask] = NewMinerTxnsProcessor(o, store)
		case ActorCodesTask:
			tsi.processors[ActorCodesTask] = actorcodes.NewTask(o, store)
		case MinersTask:
			tsi.processors[MinersTask] = miner.NewTask(o, store)
		case BlocksTask:
			tsi.processors[BlocksTask] = blocks.NewTask(o, store)
		case MessagesTask:
			tsi.messageProcessors[MessagesTask] = messages.NewTask(o, store)
		case MarketsTask:
			tsi.processors[MarketsTask] = market.NewTask(o, store)
		}
	}
	return tsi, nil
}

// TipSet is called when a new tipset has been discovered
func (t *TipSetIndexer) TipSet(ctx context.Context, ts *types.TipSet) error {
	tctx := context.Background()

	inFlight := 0
	results := make(chan *TaskResult, len(t.processors)+len(t.actorProcessors))

	// Run each tipset processing task concurrently
	for name, p := range t.processors {
		inFlight++
		t.runProcessor(tctx, p, name, ts, results)
	}

	if len(t.actorProcessors) > 0 || len(t.messageProcessors) > 0 {
		var parent, child *types.TipSet
		if t.lastTipSet != nil {
			if t.lastTipSet.Height() > ts.Height() {
				// last tipset seen was the child
				child = t.lastTipSet
				parent = ts
			} else if t.lastTipSet.Height() < ts.Height() {
				// last tipset seen was the parent
				child = ts
				parent = t.lastTipSet
			} else {
				log.Errorw("out of order tipsets", "height", ts.Height(), "last_height", t.lastTipSet.Height())
			}
		}

		// If no parent tipset available then we need to skip processing. It's likely we received the last or first tipset
		// in a batch. No report is generated because a different run of the indexer could cover the parent and child
		// for this tipset.
		if parent != nil {
			if t.node == nil {
				node, closer, err := t.opener.Open(ctx)
				if err != nil {
					return xerrors.Errorf("unable to open lens: %w", err)
				}
				t.node = node
				t.closer = closer
			}

			// If we have message processors then extract the messages and receipts
			if len(t.messageProcessors) > 0 {
				// log.Debug("before GetExecutedMessagesForTipset")
				// t.node.GetExecutedMessagesForTipset(ctx, child, parent)
				// log.Debug("after GetExecutedMessagesForTipset")

				// emsgs, err := t.node.GetExecutedMessagesForTipset(ctx, child, parent)
				emsgs := []*lens.ExecutedMessage{}

				// if err == nil {
				// Start all the message processors
				for name, p := range t.messageProcessors {
					inFlight++
					t.runMessageProcessor(tctx, p, name, child, parent, emsgs, results)
				}
				// } else {
				// 	return xerrors.Errorf("failed to extract messages: %w", err)
				// }
			}
		}
	}

	// remember the last tipset we observed
	t.lastTipSet = ts

	return nil
}

func (t *TipSetIndexer) runProcessor(ctx context.Context, p TipSetProcessor, name string, ts *types.TipSet, results chan *TaskResult) {
	p.ProcessTipSet(ctx, ts)
	return
}

func (t *TipSetIndexer) runMessageProcessor(ctx context.Context, p MessageProcessor, name string, ts, pts *types.TipSet, emsgs []*lens.ExecutedMessage, results chan *TaskResult) {
	p.ProcessMessages(ctx, ts, pts, emsgs)
	return
}

func (t *TipSetIndexer) Close() error {
	if t.closer != nil {
		t.closer()
		t.closer = nil
	}
	t.node = nil

	for name, p := range t.processors {
		if err := p.Close(); err != nil {
			log.Errorw("error received while closing task processor", "error", err, "task", name)
		}
	}
	for name, p := range t.messageProcessors {
		if err := p.Close(); err != nil {
			log.Errorw("error received while closing message task processor", "error", err, "task", name)
		}
	}
	for name, p := range t.actorProcessors {
		if err := p.Close(); err != nil {
			log.Errorw("error received while closing actor task processor", "error", err, "task", name)
		}
	}

	return nil
}

// A TaskResult is either some data to persist or an error which indicates that the task did not complete. Partial
// completions are possible provided the Data contains a persistable log of the results.
type TaskResult struct {
	Task  string
	Error error
	Data  model.Persistable
}

type TipSetProcessor interface {
	// ProcessTipSet processes a tipset. If error is non-nil then the processor encountered a fatal error.
	// Any data returned must be accompanied by a processing report.
	ProcessTipSet(ctx context.Context, ts *types.TipSet) (model.Persistable, error)
	Close() error
}

type MessageProcessor interface {
	// ProcessMessages processes messages contained within a tipset. If error is non-nil then the processor encountered a fatal error.
	// pts is the tipset containing the messages, ts is the tipset containing the receipts
	// Any data returned must be accompanied by a processing report.
	ProcessMessages(ctx context.Context, ts *types.TipSet, pts *types.TipSet, emsgs []*lens.ExecutedMessage) (model.Persistable, error)
	Close() error
}

type ActorProcessor interface {
	// ProcessActor processes a set of actors. If error is non-nil then the processor encountered a fatal error.
	// Any data returned must be accompanied by a processing report.
	ProcessActors(ctx context.Context, ts *types.TipSet, pts *types.TipSet, actors map[string]types.Actor) (model.Persistable, error)
	Close() error
}
