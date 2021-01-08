package fetch

// github.com/buidl-labs/fmm

import (
	"context"

	"log"

	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"golang.org/x/sync/errgroup"

	baselens "github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/tasks"
)

// MinerFetcherTask fetches raw miner data
type MinerFetcherTask struct {
	lens *baselens.Lens
}

// NewMinerFetcherTask creates the task
func NewMinerFetcherTask(lens *baselens.Lens) tasks.Task {
	// func NewMinerFetcherTask(lens *baselens.Lens) interface{} {
	return &MinerFetcherTask{lens: lens}
}

// Run performs the task
func (t *MinerFetcherTask) Run(ctx context.Context) error {
	// payload := p.(*payload)
	payload := tasks.Payload{}
	tsk := payload.EpochTipset.Key()
	log.Println("tsk", tsk)

	deals, err := t.lens.Miner.GetMarketDeals(tsk)
	if err != nil {
		return err
	}
	payload.MinersDeals = deals

	addresses, err := t.lens.Miner.GetAddressesByTipset(tsk)
	if err != nil {
		return err
	}
	payload.MinersAddresses = addresses

	payload.MinersInfo = make([]*miner.MinerInfo, len(addresses))
	payload.MinersPower = make([]*api.MinerPower, len(addresses))
	payload.MinersFaults = make([]*bitfield.BitField, len(addresses))

	eg, _ := errgroup.WithContext(ctx)

	for i := range addresses {
		func(index int) {
			eg.Go(func() error {
				return fetchMinerData(index, t.lens, payload)
			})
		}(i)
	}

	return eg.Wait()
}

func fetchMinerData(index int, c *baselens.Lens, p tasks.Payload) error {
	address := p.MinersAddresses[index]
	tsk := p.EpochTipset.Key()

	info, err := c.Miner.GetInfoByTipset(address, tsk)
	if err != nil {
		return err
	}

	power, err := c.Miner.GetPowerByTipset(address, tsk)
	if err != nil {
		return err
	}

	faults, err := c.Miner.GetFaultsByTipset(address, tsk)
	if err != nil {
		return err
	}

	p.MinersInfo[index] = info
	p.MinersPower[index] = power
	p.MinersFaults[index] = faults

	return nil
}
