package chain

import (
	"context"
	"fmt"
	"time"

	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	// "github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
)

var _ TipSetObserver = (*TipSetIndexer)(nil)

// TipSetIndexer waits for tipsets and persists their block data into a database.
type TipSetIndexer struct {
	window time.Duration
	// storage         model.Storage
	// processors      map[string]TipSetProcessor
	// actorProcessors map[string]ActorProcessor
	name        string
	persistSlot chan struct{}
	lastTipSet  *types.TipSet
	node        lens.API
	opener      lens.APIOpener
	closer      lens.APICloser
	// api *apistruct.FullNodeStruct
}

func NewTipSetIndexer(o lens.APIOpener, window time.Duration, name string, tasks []string) (*TipSetIndexer, error) {
	tsi := &TipSetIndexer{
		// storage:         d,
		window:      window,
		name:        name,
		persistSlot: make(chan struct{}, 1), // allow one concurrent persistence job
		// processors:      map[string]TipSetProcessor{},
		// actorProcessors: map[string]ActorProcessor{},
		opener: o,
		// api: api,
	}

	for _, task := range tasks {
		fmt.Println(task)
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
	fmt.Println(start, tctx)

	return nil
}

func (t *TipSetIndexer) Close() error {
	if t.closer != nil {
		t.closer()
		t.closer = nil
	}
	t.node = nil

	// for name, p := range t.processors {
	// 	if err := p.Close(); err != nil {
	// 		log.Errorw("error received while closing task processor", "error", err, "task", name)
	// 	}
	// }
	// for name, p := range t.actorProcessors {
	// 	if err := p.Close(); err != nil {
	// 		log.Errorw("error received while closing actor task processor", "error", err, "task", name)
	// 	}
	// }
	return nil
}
