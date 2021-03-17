package market

import (
	"context"
	"strconv"
	"time"

	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	marketmodel "github.com/buidl-labs/filecoin-chain-indexer/model/market"
)

var log = logging.Logger("market")

type Task struct {
	node       lens.API
	opener     lens.APIOpener
	closer     lens.APICloser
	lastTipSet *types.TipSet
	store      db.Store
}

func NewTask(opener lens.APIOpener, store db.Store) *Task {
	return &Task{
		opener: opener,
		store:  store,
	}
}

func (p *Task) ProcessTipSet(ctx context.Context, ts *types.TipSet) (model.Persistable, error) {
	var data model.Persistable
	if p.node == nil {
		node, closer, err := p.opener.Open(ctx)
		if err != nil {
			return data, xerrors.Errorf("unable to open lens: %w", err)
		}
		p.node = node
		p.closer = closer
	}

	tsk := ts.Key()
	start := time.Now()
	log.Debugw("calling StateMarketDeals", "start time", start)
	dealStates, err := p.node.StateMarketDeals(context.Background(), tsk)
	if err != nil {
		log.Errorw("StateMarketDeals", "error", err)
		return nil, err
	}
	log.Debugw("calling StateMarketDeals", "end time", time.Now())
	log.Debugw("time taken to StateMarketDeals", "time", time.Since(start))

	for idStr, deal := range dealStates {
		dealID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return nil, err
		}
		mds := &marketmodel.MarketDealState{
			Height:           int64(ts.Height()),
			DealID:           dealID,
			SectorStartEpoch: int64(deal.State.SectorStartEpoch),
			LastUpdateEpoch:  int64(deal.State.LastUpdatedEpoch),
			SlashEpoch:       int64(deal.State.SlashEpoch),
			StateRoot:        ts.ParentState().String(),
		}
		mdp := &marketmodel.MarketDealProposal{
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
		r, err := p.store.DB.Model(mdp).Insert()
		if err != nil {
			log.Errorw("inserting marketDealProposal", "error", err)
		} else {
			log.Debug("inserted marketDealProposal", r)
		}
		r, err = p.store.DB.Model(mds).Insert()
		if err != nil {
			log.Errorw("inserting marketDealState", "error", err)
		} else {
			log.Debug("inserted marketDealState", r)
		}
	}

	return data, nil
}

func (p *Task) Close() error {
	if p.closer != nil {
		p.closer()
		p.closer = nil
	}
	p.node = nil
	return nil
}
