package chain

import (
	"context"
	"fmt"
	"strconv"

	"github.com/filecoin-project/lotus/chain/types"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	marketmodel "github.com/buidl-labs/filecoin-chain-indexer/model/market"
)

type MarketProcessor struct {
	node       lens.API
	opener     lens.APIOpener
	closer     lens.APICloser
	lastTipSet *types.TipSet
	store      db.Store
}

func NewMarketProcessor(opener lens.APIOpener, store db.Store) *MarketProcessor {
	p := &MarketProcessor{
		opener: opener,
		store:  store,
	}
	return p
}

func (p *MarketProcessor) ProcessTipSet(ctx context.Context, ts *types.TipSet) (model.Persistable, error) {
	var data model.Persistable

	if p.node == nil {
		node, closer, err := p.opener.Open(ctx)
		if err != nil {
			return data, xerrors.Errorf("unable to open lens: %w", err)
		}
		p.node = node
		p.closer = closer
	}
	// var pl model.PersistableList

	tsk := ts.Key()
	fmt.Println("mtsk", tsk)
	dealStates, err := p.node.StateMarketDeals(context.Background(), tsk)
	if err != nil {
		return nil, err
	}
	fmt.Println("dealStates", dealStates)

	states := make(marketmodel.MarketDealStates, len(dealStates))
	proposals := make(marketmodel.MarketDealProposals, len(dealStates))
	idx := 0
	for idStr, deal := range dealStates {
		dealID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return nil, err
		}
		states[idx] = &marketmodel.MarketDealState{
			Height:           int64(ts.Height()),
			DealID:           dealID,
			SectorStartEpoch: int64(deal.State.SectorStartEpoch),
			LastUpdateEpoch:  int64(deal.State.LastUpdatedEpoch),
			SlashEpoch:       int64(deal.State.SlashEpoch),
			StateRoot:        ts.ParentState().String(),
		}
		proposals[idx] = &marketmodel.MarketDealProposal{
			Height:               int64(ts.Height()),
			DealID:               dealID,
			StateRoot:            ts.ParentState().String(),
			PaddedPieceSize:      uint64(deal.Proposal.PieceSize),
			UnpaddedPieceSize:    uint64(deal.Proposal.PieceSize.Unpadded()),
			StartEpoch:           int64(deal.Proposal.StartEpoch),
			EndEpoch:             int64(deal.Proposal.EndEpoch),
			ClientID:             deal.Proposal.Client.String(),
			ProviderID:           deal.Proposal.Provider.String(),
			ClientCollateral:     deal.Proposal.ClientCollateral.String(),
			ProviderCollateral:   deal.Proposal.ProviderCollateral.String(),
			StoragePricePerEpoch: deal.Proposal.StoragePricePerEpoch.String(),
			PieceCID:             deal.Proposal.PieceCID.String(),
			IsVerified:           deal.Proposal.VerifiedDeal,
			Label:                deal.Proposal.Label,
		}
		p.store.PersistMarketDealProposals(*proposals[idx])
		idx++
	}

	return data, nil
}

func (p *MarketProcessor) Close() error {
	if p.closer != nil {
		p.closer()
		p.closer = nil
	}
	p.node = nil
	return nil
}
