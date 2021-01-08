package chain

import (
	"context"

	"github.com/filecoin-project/lotus/chain/types"
)

// A TipSetObserver waits for notifications of new tipsets.
type TipSetObserver interface {
	TipSet(ctx context.Context, ts *types.TipSet) error
}
