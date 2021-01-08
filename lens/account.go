package lens

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/types"
)

type accountClient struct {
	api *apistruct.FullNodeStruct
}

// GetActor fetches account details for a given address
func (ac *accountClient) GetActor(addr address.Address) (*types.Actor, error) {
	return ac.api.StateGetActor(context.Background(), addr, types.EmptyTSK)
}

// GetIDAddress fetches the ID address of a given address
func (ac *accountClient) GetIDAddress(addr address.Address) string {
	id, _ := ac.api.StateLookupID(context.Background(), addr, types.EmptyTSK)

	return id.String()
}

// GetPublicKeyAddress fetches the public key address of a given address
func (ac *accountClient) GetPublicKeyAddress(addr address.Address) string {
	pubkey, _ := ac.api.StateAccountKey(context.Background(), addr, types.EmptyTSK)

	return pubkey.String()
}
