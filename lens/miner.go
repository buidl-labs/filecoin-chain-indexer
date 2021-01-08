package lens

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
)

type minerClient struct {
	api *apistruct.FullNodeStruct
}

// GetMarketDeals fetches market deals for a given tipset
func (mc *minerClient) GetMarketDeals(tsk types.TipSetKey) (map[string]api.MarketDeal, error) {
	return mc.api.StateMarketDeals(context.Background(), tsk)
}

// GetAddressesByTipset fetches miners' addresses for a given tipset
func (mc *minerClient) GetAddressesByTipset(tsk types.TipSetKey) ([]address.Address, error) {
	return mc.api.StateListMiners(context.Background(), tsk)
}

// GetInfoByTipset fetches miner's information for a given tipset
func (mc *minerClient) GetInfoByTipset(address address.Address, tsk types.TipSetKey) (*miner.MinerInfo, error) {
	info, err := mc.api.StateMinerInfo(context.Background(), address, tsk)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// GetPowerByTipset fetches miner's power for a given tipset
func (mc *minerClient) GetPowerByTipset(address address.Address, tsk types.TipSetKey) (*api.MinerPower, error) {
	return mc.api.StateMinerPower(context.Background(), address, tsk)
}

// GetFaultsByTipset fetches miner's faults for a given tipset
func (mc *minerClient) GetFaultsByTipset(address address.Address, tsk types.TipSetKey) (*bitfield.BitField, error) {
	faults, err := mc.api.StateMinerFaults(context.Background(), address, tsk)
	if err != nil {
		return nil, err
	}

	return &faults, nil
}
