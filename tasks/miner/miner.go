package miner

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	minermodel "github.com/buidl-labs/filecoin-chain-indexer/model/miner"
	powermodel "github.com/buidl-labs/filecoin-chain-indexer/model/power"
	"github.com/buidl-labs/filecoin-chain-indexer/util"
)

var log = logging.Logger("miner")

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
	if p.node == nil {
		node, closer, err := p.opener.Open(ctx)
		if err != nil {
			return nil, xerrors.Errorf("unable to open lens: %w", err)
		}
		p.node = node
		p.closer = closer
	}

	tsk := ts.Key()

	var addresses []address.Address
	activeMinersOnly, err := util.GetenvInt("ACTIVE_MINERS_ONLY")
	if err != nil {
		log.Error("getting miners", "error", err)
		return nil, err
	}
	if activeMinersOnly == 1 {
		addresses, _ = ActiveMinerAddresses()
		log.Debugw("fetching data for active miners", "count", len(addresses))
	} else {
		addresses, err := p.node.StateListMiners(context.Background(), tsk)
		if err != nil {
			return nil, err
		}
		log.Debugw("fetching data for all miners", "count", len(addresses))
	}

	for _, addr := range addresses {
		minerInfo := new(minermodel.MinerInfo)
		err = p.store.DB.Model(minerInfo).
			Where("state_root = ?", "-1").
			Where("miner_id = ?", addr.String()).
			Select()
		if err != nil {
			log.Debugw("never indexed", "miner", addr.String())

			// insert objects with state_root -1

			_, err := p.store.DB.Model(&minermodel.MinerInfo{
				MinerID:         addr.String(),
				Address:         "0",
				PeerID:          "0",
				OwnerID:         "0",
				WorkerID:        "0",
				Height:          int64(0),
				StateRoot:       "-1",
				StorageAskPrice: "0",
				MinPieceSize:    uint64(1),
				MaxPieceSize:    uint64(1),
			}).Insert()
			if err != nil {
				log.Errorw("inserting minerinfo",
					"miner", addr.String(),
					"error", err,
				)
				return nil, err
			} else {
				log.Debugw("inserted minerinfo", "miner", addr.String())
			}

			_, err = p.store.DB.Model(&powermodel.PowerActorClaim{
				MinerID:         addr.String(),
				Height:          int64(0),
				StateRoot:       "-1",
				RawBytePower:    "0",
				QualityAdjPower: "0",
			}).Insert()
			if err != nil {
				log.Errorw("inserting poweractorclaim",
					"miner", addr.String(),
					"error", err,
				)
				return nil, err
			} else {
				log.Debugw("inserted poweractorclaim", "miner", addr.String())
			}

			_, err = p.store.DB.Model(&minermodel.MinerFund{
				MinerID:           addr.String(),
				Height:            int64(0),
				StateRoot:         "-1",
				LockedFunds:       "0",
				InitialPledge:     "0",
				PreCommitDeposits: "0",
				AvailableBalance:  "0",
			}).Insert()
			if err != nil {
				log.Errorw("inserting minerfund",
					"miner", addr.String(),
					"error", err,
				)
				return nil, err
			} else {
				log.Debugw("inserted minerfund", "miner", addr.String())
			}
		}

		var info miner.MinerInfo
		var mpower *api.MinerPower

		info, err = p.node.StateMinerInfo(context.Background(), addr, tsk)
		if err != nil {
			log.Errorw("state minerinfo", "error", err)
			return nil, err
		}

		mpower, err = p.node.StateMinerPower(context.Background(), addr, tsk)
		if err != nil {
			log.Errorw("state minerpower", "error", err)
			return nil, err
		}

		availableBal, err := p.node.StateMinerAvailableBalance(context.Background(), addr, tsk)
		if err != nil {
			log.Errorw("state mineravailablebalance", "error", err)
			return nil, err
		}

		// c1 := make(chan *storagemarket.StorageAsk, 1)
		// if info.PeerId != nil {
		// 	go func() {
		// 		ask, _ := GetClientAsk(p, info, addr)
		// 		c1 <- ask
		// 	}()
		// 	select {
		// 	case ask := <-c1:
		// 		log.Debug("SLMAsk: {minerid:", ask.Miner, "price:", ask.Price, "verifiedP:", ask.VerifiedPrice, "minPS:", ask.MinPieceSize, "maxPS:", ask.MaxPieceSize, "timestamp:", ask.Timestamp, "Expiry:", ask.Expiry, "}")
		// 	case <-time.After(1 * time.Second): // set a higher value
		// 		log.Debug("Clientqueryask out of time :(")
		// 	}
		// }

		// allSectors, err := p.node.StateMinerSectors(context.Background(), addr, nil, tsk)
		// if err != nil {
		// 	log.Errorw("state minersectors", "error", err)
		// 	return nil, err
		// }
		// activeSectors, err := p.node.StateMinerActiveSectors(context.Background(), addr, tsk)
		// if err != nil {
		// 	log.Errorw("state mineractivesectors", "error", err)
		// 	return nil, err
		// }
		// faultySectors, err := p.node.StateMinerFaults(context.Background(), addr, tsk)
		// if err != nil {
		// 	log.Errorw("state minerfaults", "error", err)
		// 	return nil, err
		// }

		// var fsc uint64
		// fsc, _ = faultySectors.Count()

		// log.Debugw("allsectors",
		// 	"miner", addr.String(),
		// 	"count", len(allSectors),
		// )
		// log.Debugw("activesectors",
		// 	"miner", addr.String(),
		// 	"count", len(activeSectors),
		// )
		// log.Debugw("faultysectors",
		// 	"miner", addr.String(),
		// 	"count", fsc,
		// )

		peerID := "0"
		if info.PeerId != nil {
			peerID = info.PeerId.String()
		}

		mimod := &minermodel.MinerInfo{
			Height:          int64(ts.Height()),
			MinerID:         addr.String(),
			StateRoot:       "-1",
			Address:         "0",
			PeerID:          peerID,
			OwnerID:         info.Owner.String(),
			WorkerID:        info.Worker.String(),
			StorageAskPrice: "0",
			MinPieceSize:    uint64(1),
			MaxPieceSize:    uint64(1),
		}
		log.Info(mimod.Height)
		_, err = p.store.DB.Model(mimod).
			Set("height = ?", ts.Height()).
			Set("peer_id = ?", peerID).
			Set("owner_id = ?", info.Owner.String()).
			Set("worker_id = ?", info.Worker.String()).
			Where("state_root = ?", "-1").
			Where("miner_id = ?", addr.String()).
			Update()
		if err != nil {
			log.Errorw("updating minerinfo",
				"miner", addr.String(),
				"error", err,
			)
		} else {
			log.Infow("ht", "ht", ts.Height())
			log.Debugw("updated minerinfo",
				"miner", addr.String(),
			)
		}

		rbp := "0"
		if &mpower.MinerPower.RawBytePower != nil {
			rbp = mpower.MinerPower.RawBytePower.String()
		}
		qap := "0"
		if &mpower.MinerPower.QualityAdjPower != nil {
			qap = mpower.MinerPower.QualityAdjPower.String()
		}

		pac := &powermodel.PowerActorClaim{
			MinerID:         addr.String(),
			Height:          int64(ts.Height()),
			StateRoot:       "-1",
			RawBytePower:    rbp,
			QualityAdjPower: qap,
		}
		_, err = p.store.DB.Model(pac).
			Set("raw_byte_power = ?", rbp).
			Set("quality_adj_power = ?", qap).
			Set("height = ?", ts.Height()).
			Where("state_root = ?", "-1").
			Where("miner_id = ?", addr.String()).
			Update()
		if err != nil {
			log.Errorw("updating poweractorclaim",
				"miner", addr.String(),
				"error", err,
			)
		} else {
			log.Debugw("updated poweractorclaim",
				"miner", addr.String(),
			)
		}

		ec, err := NewMinerStateExtractionContext(p, context.Background(), addr, ts)
		if err != nil {
			log.Error(err)
		} else {
			availableBalStr := "0"
			if &availableBal != nil {
				availableBalStr = availableBal.String()
			}
			mlf, err := ExtractMinerLockedFunds(ec, addr, ts, availableBalStr)
			if err != nil {
				log.Error(err)
			} else {
				_, err = p.store.DB.Model(mlf).
					Set("available_balance = ?", mlf.AvailableBalance).
					Set("pre_commit_deposits = ?", mlf.PreCommitDeposits).
					Set("initial_pledge = ?", mlf.InitialPledge).
					Set("locked_funds = ?", mlf.LockedFunds).
					Where("state_root = ?", "-1").
					Where("miner_id = ?", addr.String()).
					Update()
				if err != nil {
					log.Errorw("updating minerfund",
						"miner", addr.String(),
						"error", err,
					)
				} else {
					log.Debugw("updated minerfund",
						"miner", addr.String(),
					)
				}
			}
		}
	}

	return nil, nil
}

func (p *Task) Close() error {
	if p.closer != nil {
		p.closer()
		p.closer = nil
	}
	p.node = nil
	return nil
}

func NewMinerStateExtractionContext(p *Task, ctx context.Context, addr address.Address, ts *types.TipSet) (*MinerStateExtractionContext, error) {
	tsk := ts.Key()
	curActor, err := p.node.StateGetActor(ctx, addr, tsk)
	if err != nil {
		return nil, err
	}
	curTipset, err := p.node.ChainGetTipSet(ctx, tsk)
	if err != nil {
		return nil, err
	}
	curState, err := miner.Load(p.node.Store(), curActor)
	if err != nil {
		return nil, err
	}

	prevState := curState
	if ts.Height() != 0 {
		prevActor, err := p.node.StateGetActor(ctx, addr, ts.Parents())
		if err != nil {
			return nil, xerrors.Errorf("loading previous miner %s at tipset %s epoch %d: %w", addr, ts.Parents(), ts.Height(), err)
		}

		prevState, err = miner.Load(p.node.Store(), prevActor)
		if err != nil {
			return nil, xerrors.Errorf("loading previous miner actor state: %w", err)
		}
	}

	return &MinerStateExtractionContext{
		PrevState: prevState,
		CurrActor: curActor,
		CurrState: curState,
		CurrTs:    curTipset,
	}, nil
}

type MinerStateExtractionContext struct {
	PrevState miner.State

	CurrActor *types.Actor
	CurrState miner.State
	CurrTs    *types.TipSet
}

func (m *MinerStateExtractionContext) IsGenesis() bool {
	return m.CurrTs.Height() == 0
}

func ExtractMinerCurrentDeadlineInfo(ec *MinerStateExtractionContext, addr address.Address, ts *types.TipSet) (*minermodel.MinerCurrentDeadlineInfo, error) {
	currDeadlineInfo, err := ec.CurrState.DeadlineInfo(ec.CurrTs.Height())
	if err != nil {
		return nil, err
	}
	if !ec.IsGenesis() {
		prevDeadlineInfo, err := ec.PrevState.DeadlineInfo(ec.CurrTs.Height())
		if err != nil {
			return nil, err
		}
		if prevDeadlineInfo == currDeadlineInfo {
			return nil, nil
		}
	}

	tsPSstr := ""
	if ts != nil {
		tsPS := ts.ParentState()
		tsPSstr = tsPS.String()
	}
	return &minermodel.MinerCurrentDeadlineInfo{
		Height:        int64(ec.CurrTs.Height()),
		MinerID:       addr.String(),
		StateRoot:     tsPSstr,
		DeadlineIndex: currDeadlineInfo.Index,
		PeriodStart:   int64(currDeadlineInfo.PeriodStart),
		Open:          int64(currDeadlineInfo.Open),
		Close:         int64(currDeadlineInfo.Close),
		Challenge:     int64(currDeadlineInfo.Challenge),
		FaultCutoff:   int64(currDeadlineInfo.FaultCutoff),
	}, nil
}

func ExtractMinerLockedFunds(ec *MinerStateExtractionContext, addr address.Address, ts *types.TipSet, availableBal string) (*minermodel.MinerFund, error) {
	currLocked, err := ec.CurrState.LockedFunds()
	if err != nil {
		return nil, xerrors.Errorf("loading current miner locked funds: %w", err)
	}
	if !ec.IsGenesis() {
		prevLocked, err := ec.PrevState.LockedFunds()
		if err != nil {
			return nil, xerrors.Errorf("loading previous miner locked funds: %w", err)
		}
		if prevLocked == currLocked {
			return nil, nil
		}
	}
	tsPSstr := ""
	if ts != nil {
		tsPS := ts.ParentState()
		tsPSstr = tsPS.String()
	}
	return &minermodel.MinerFund{
		Height:            int64(ec.CurrTs.Height()),
		MinerID:           addr.String(),
		StateRoot:         tsPSstr,
		LockedFunds:       currLocked.VestingFunds.String(),
		InitialPledge:     currLocked.InitialPledgeRequirement.String(),
		PreCommitDeposits: currLocked.PreCommitDeposits.String(),
		AvailableBalance:  availableBal,
	}, nil
}

func GetClientAsk(p *Task, info miner.MinerInfo, addr address.Address) (*storagemarket.StorageAsk, error) {
	ask, err := p.node.ClientQueryAsk(context.Background(), *info.PeerId, addr)
	if err != nil {
		return ask, err
	}
	return ask, nil
}
