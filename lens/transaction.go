package lens

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
)

type transactionClient struct {
	api *apistruct.FullNodeStruct
}

// GetCIDsByAddress returns CIDs of incoming and outgoing messages for a given address
func (tc *transactionClient) GetCIDsByAddress(addr address.Address, tsk types.TipSetKey, height int64) ([]cid.Cid, error) {
	ctx := context.Background()
	epoch := abi.ChainEpoch(height)

	var cids []cid.Cid

	from, err := tc.api.StateListMessages(ctx, &api.MessageMatch{From: addr}, tsk, epoch)
	if err != nil {
		return nil, err
	}
	cids = append(cids, from...)

	to, err := tc.api.StateListMessages(ctx, &api.MessageMatch{To: addr}, tsk, epoch)
	if err != nil {
		return nil, err
	}
	cids = append(cids, to...)

	return cids, nil
}

// GetMessage returns a message for a given CID
func (tc *transactionClient) GetMessage(cid cid.Cid) (*types.Message, error) {
	return tc.api.ChainGetMessage(context.Background(), cid)
}
