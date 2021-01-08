package tasks

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	// "github.com/figment-networks/filecoin-indexer/model"
	// "github.com/figment-networks/indexing-engine/pipeline"
)

type Payload struct {
	currentHeight int64
	processed     bool

	// Fetcher stage
	EpochTipset          *types.TipSet
	MinersDeals          map[string]api.MarketDeal
	MinersAddresses      []address.Address
	MinersInfo           []*miner.MinerInfo
	MinersPower          []*api.MinerPower
	MinersFaults         []*bitfield.BitField
	TransactionsCIDs     []cid.Cid
	TransactionsMessages []*types.Message
}

type Task interface {
	//Run Task
	Run(context.Context) error

	// GetName gets name of task
	// GetName() string
}

func (p *Payload) SetCurrentHeight(height int64) {
	p.currentHeight = height
}

func (p *Payload) GetCurrentHeight() int64 {
	return p.currentHeight
}

func (p *Payload) MarkAsProcessed() {
	p.processed = true
}

func (p *Payload) IsProcessed() bool {
	return p.processed
}
