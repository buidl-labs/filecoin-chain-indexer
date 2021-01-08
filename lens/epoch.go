package lens

import (
	"context"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
)

type epochClient struct {
	api *apistruct.FullNodeStruct
}

// GetCurrentHeight fetches the height of the current epoch
func (ec *epochClient) GetCurrentHeight() (int64, error) {
	tipset, err := ec.api.ChainHead(context.Background())
	if err != nil {
		return 0, err
	}

	return int64(tipset.Height()), nil
}

// GetTipsetByHeight fetches a tipset for a given height
func (ec *epochClient) GetTipsetByHeight(height int64) (*types.TipSet, error) {
	ctx := context.Background()
	epoch := abi.ChainEpoch(height)

	return ec.api.ChainGetTipSetByHeight(ctx, epoch, types.EmptyTSK)
}
