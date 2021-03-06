package chain

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/filecoin-project/lotus/chain/types"
	log "github.com/sirupsen/logrus"

	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

const (
	InitTask      = "init"
	MinerTxnsTask = "minertxns"
	BlocksTask    = "blocks"
	MinersTask    = "miners"
	MessagesTask  = "messages"
	MarketsTask   = "markets"
	MinerInfoTask = "minerinfo"
)

var _ TipSetObserver = (*TipSetIndexer)(nil)

// TipSetIndexer waits for tipsets and persists their block data into a database.
type TipSetIndexer struct {
	db              *sql.DB
	store           db.Store
	storage         model.Storage
	window          time.Duration
	name            string
	persistSlot     chan struct{}
	processors      map[string]TipSetProcessor
	actorProcessors map[string]ActorProcessor
	lastTipSet      *types.TipSet
	node            lens.API
	opener          lens.APIOpener
	closer          lens.APICloser
	cfg             config.Config
}

func NewTipSetIndexer(o lens.APIOpener, db *sql.DB, store db.Store, s model.Storage, window time.Duration, name string, tasks []string, cfg config.Config) (*TipSetIndexer, error) {
	tsi := &TipSetIndexer{
		db:              db,
		store:           store,
		storage:         s,
		window:          window,
		name:            name,
		persistSlot:     make(chan struct{}, 1), // allow one concurrent persistence job
		processors:      map[string]TipSetProcessor{},
		actorProcessors: map[string]ActorProcessor{},
		opener:          o,
		cfg:             cfg,
	}

	for _, task := range tasks {
		log.Info("task", task)
		switch task {
		// case InitTask:
		// 	tsi.processors[InitTask] = NewInitProcessor(store)
		case MinerInfoTask:
			tsi.processors[MinerInfoTask] = NewMinerInfoProcessor(o, store)
		case MinerTxnsTask:
			tsi.processors[MinerTxnsTask] = NewMinerTxnsProcessor(o, store)
		case BlocksTask:
			tsi.processors[BlocksTask] = NewBlockProcessor(store)
		case MinersTask:
			tsi.processors[MinersTask] = NewMinerProcessor(o, store)
		case MessagesTask:
			tsi.processors[MessagesTask] = NewMessageProcessor(o, store, cfg)
		case MarketsTask:
			tsi.processors[MarketsTask] = NewMarketProcessor(o, store)
		}
	}
	return tsi, nil
}

// TipSet is called when a new tipset has been discovered
func (t *TipSetIndexer) TipSet(ctx context.Context, ts *types.TipSet) error {
	var cancel func()
	var tctx context.Context // cancellable context for the task
	if t.window > 0 {
		// Do as much indexing as possible in the specified time window (usually one epoch when following head of chain)
		// Anything not completed in that time will be marked as incomplete
		tctx, cancel = context.WithTimeout(ctx, t.window)
	} else {
		// Ensure all goroutines are stopped when we exit
		tctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	start := time.Now()
	// fmt.Println(start, tctx)

	inFlight := 0
	results := make(chan *TaskResult, len(t.processors)+len(t.actorProcessors))

	// A map to gather the persistable outputs from each task
	taskOutputs := make(map[string]model.PersistableList, len(t.processors)+len(t.actorProcessors))

	// Run each tipset processing task concurrently
	for name, p := range t.processors {
		inFlight++
		go t.runProcessor(tctx, p, name, ts, results)
	}

	// Wait for all tasks to complete
	for inFlight > 0 {
		res := <-results
		inFlight--

		// Was there a fatal error?
		if res.Error != nil {
			log.Print("task returned with error", "error", res.Error.Error())
			// tell all the processors to close their connections to the lens, they can reopen when needed
			if err := t.Close(); err != nil {
				log.Print("error received while closing tipset indexer", "error", err)
			}
			return res.Error
		}

		// fmt.Println("torestak", res.Data)
		// Persist the processing report and the data in a single transaction
		taskOutputs[res.Task] = model.PersistableList{res.Data}
	}

	// remember the last tipset we observed
	t.lastTipSet = ts

	// fmt.Println("TOUTPUTS", taskOutputs)

	if len(taskOutputs) == 0 {
		// Nothing to persist
		log.Print("tipset complete, nothing to persist", "total_time", time.Since(start))
		return nil
	}

	// wait until there is an empty slot before persisting
	// log.Print("waiting to persist data", "time", time.Since(start))
	select {
	case <-ctx.Done():
		return ctx.Err()
	case t.persistSlot <- struct{}{}:
		// Slot is free so we can continue
	}

	// Persist all results
	go func() {
		// free up the slot when done
		defer func() {
			<-t.persistSlot
		}()

		// log.Println("persisting data", "time", time.Since(start))
		var wg sync.WaitGroup
		wg.Add(len(taskOutputs))

		// Persist each processor's data concurrently since they don't overlap
		for task, p := range taskOutputs {
			go func(task string, p model.Persistable) {
				defer wg.Done()
				if err := t.storage.PersistBatch(ctx, p); err != nil {
					// log.Println("persistence failed", "task", task, "error", err)
					return
				}
				// log.Println("task data persisted", "task", task, "time", time.Since(start), "typeofp", reflect.TypeOf(p), p)
			}(task, p)
		}
		wg.Wait()
		// log.Println("tipset complete", "total_time", time.Since(start))
	}()

	return nil
}

func (t *TipSetIndexer) runProcessor(ctx context.Context, p TipSetProcessor, name string, ts *types.TipSet, results chan *TaskResult) {
	data, err := p.ProcessTipSet(ctx, ts)
	if err != nil {
		results <- &TaskResult{
			Task:  name,
			Error: err,
		}
		return
	}
	results <- &TaskResult{
		Task: name,
		Data: data,
	}
}

func (t *TipSetIndexer) Close() error {
	if t.closer != nil {
		t.closer()
		t.closer = nil
	}
	t.node = nil

	for name, p := range t.processors {
		if err := p.Close(); err != nil {
			log.Println("error received while closing task processor", "error", err, "task", name)
		}
	}
	// for name, p := range t.actorProcessors {
	// 	if err := p.Close(); err != nil {
	// 		log.Println("error received while closing actor task processor", "error", err, "task", name)
	// 	}
	// }
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

type ActorProcessor interface {
	// ProcessActor processes a set of actors. If error is non-nil then the processor encountered a fatal error.
	// Any data returned must be accompanied by a processing report.
	ProcessActors(ctx context.Context, ts *types.TipSet, pts *types.TipSet, actors map[string]types.Actor) (model.Persistable, error)
	Close() error
}
