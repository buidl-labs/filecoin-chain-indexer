package chain

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	minermodel "github.com/buidl-labs/filecoin-chain-indexer/model/miner"
	powermodel "github.com/buidl-labs/filecoin-chain-indexer/model/power"
)

type MinerProcessor struct {
	node       lens.API
	opener     lens.APIOpener
	closer     lens.APICloser
	lastTipSet *types.TipSet
	store      db.Store
}

func NewMinerProcessor(opener lens.APIOpener, store db.Store) *MinerProcessor {
	p := &MinerProcessor{
		opener: opener,
		store:  store,
	}
	return p
}

func (p *MinerProcessor) ProcessTipSet(ctx context.Context, ts *types.TipSet) (model.Persistable, error) {
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
	log.Info("mtsk", tsk)
	addresses, err := p.node.StateListMiners(context.Background(), tsk)
	if err != nil {
		return data, err
	}

	// var info miner.MinerInfo
	// var mpower *api.MinerPower
	// var allSectors []*miner.SectorOnChainInfo
	// var activeSectors []*miner.SectorOnChainInfo
	// var fsc uint64
	// var fsa []uint64
	// var minerinfoslist []minermodel.MinerInfo
	// // var minerinfoslist []interface{}
	// var claimedpowerlist []powermodel.PowerActorClaim
	// var minersectorslist []minermodel.MinerSectorInfo
	// var minersectorfaultslist []minermodel.MinerSectorFault
	// var minerdeadlineslist []minermodel.MinerCurrentDeadlineInfo
	// var minerfundslist []minermodel.MinerFund

	var minerinfoslist []*minermodel.MinerInfo
		// var minerinfoslist []interface{}
		var claimedpowerlist []*powermodel.PowerActorClaim
		// var minersectorslist []minermodel.MinerSectorInfo
		var minersectorfaultslist []*minermodel.MinerSectorFault
		var minerdeadlineslist []*minermodel.MinerCurrentDeadlineInfo
		var minerfundslist []*minermodel.MinerFund
	log.Info("SLM addresses", len(addresses))
	ads := addresses //[:10]
	log.Info("SLM SLICE", len(ads), ads)

	// var wg sync.WaitGroup
	// wg.Add(len(addresses))

	// i := 0
	for i, addr := range ads {
	// for i = 0; i < len(ads); i++ {
		fmt.Println("lindex", i)
		// jlim := len(ads)
		// if i+1 < jlim {
		// 	jlim = i + 1
		// }

		// var info miner.MinerInfo
		// var mpower *api.MinerPower
		// var allSectors []*miner.SectorOnChainInfo
		// var activeSectors []*miner.SectorOnChainInfo
		// var fsc uint64
		// var fsa []uint64


		// var minerinfoslist []*minermodel.MinerInfo
		// // var minerinfoslist []interface{}
		// var claimedpowerlist []*powermodel.PowerActorClaim
		// // var minersectorslist []minermodel.MinerSectorInfo
		// var minersectorfaultslist []*minermodel.MinerSectorFault
		// var minerdeadlineslist []*minermodel.MinerCurrentDeadlineInfo
		// var minerfundslist []*minermodel.MinerFund


		// for j := i; j < jlim; j++ {
		// 	fmt.Println("jindex", j)
		// 	addr := ads[j]

			var info miner.MinerInfo
			var mpower *api.MinerPower
			// var allSectors []*miner.SectorOnChainInfo
			// var activeSectors []*miner.SectorOnChainInfo
			var fsc uint64
			var fsa []uint64

			//*************************
			// go func(addr address.Address) {
			log.Info("miner", addr)
			// ida, err := p.node.StateAccountKey(context.Background(), addr, tsk)
			// if err != nil {
			// 	log.Println(err)
			// }
			// log.Info("IDA", ida)
			info, err = p.node.StateMinerInfo(context.Background(), addr, tsk)
			if err != nil {
				log.Println(err)
			}
			mpower, err = p.node.StateMinerPower(context.Background(), addr, tsk)
			if err != nil {
				log.Println(err)
			}

			// ask, err := p.node.ClientQueryAsk(context.Background(), *info.PeerId, addr)
			// if err != nil {
			// 	log.Info("SLMCLientqueryask", err)
			// } else {
			// 	log.Info("SLMAsk: {minerid:", ask.Miner, "price:", ask.Price, "verifiedP:", ask.VerifiedPrice, "minPS:", ask.MinPieceSize, "maxPS:", ask.MaxPieceSize, "timestamp:", ask.Timestamp, "Expiry:", ask.Expiry, "}")
			// }

			// var allSectors []*miner.SectorOnChainInfo
			// var activeSectors []*miner.SectorOnChainInfo
			// allSectors, err = p.node.StateMinerSectors(context.Background(), addr, nil, tsk)
			// if err != nil {
			// 	log.Println(err)
			// }
			// activeSectors, err = p.node.StateMinerActiveSectors(context.Background(), addr, tsk)
			// if err != nil {
			// 	log.Println(err)
			// }
			faultySectors, err := p.node.StateMinerFaults(context.Background(), addr, tsk)
			if err != nil {
				log.Println(err)
			}

			// log.Info("SLMallSec count", len(allSectors))
			// log.Info("SLMActSec count", len(activeSectors))
			fsc, _ = faultySectors.Count()
			fsa, _ = faultySectors.All(fsc)
			log.Info("SLMFaultySec count", fsa)
			log.Info("Info", info)
			peerID := ""
			if info.PeerId != nil {
				peerID = info.PeerId.String()
			}
			ownerID := ""
			// if info.Owner != nil {
			// 	ownerID = info.Owner.String()
			// }
			ownerID = info.Owner.String()
			workerID := ""
			// if info.Worker != nil {
			// 	workerID = info.Worker.String()
			// }
			workerID = info.Worker.String()
			minerinfoslist = append(minerinfoslist, &minermodel.MinerInfo{
				MinerID:         addr.String(),
				Address:         "",
				PeerID:          peerID,
				OwnerID:         ownerID,
				WorkerID:        workerID,
				Height:          int64(ts.Height()),
				StateRoot:       "",
				StorageAskPrice: "",
				MinPieceSize:    uint64(0),
				MaxPieceSize:    uint64(0),
			})
			rbp := ""
			if &mpower.MinerPower.RawBytePower != nil {
				rbp = mpower.MinerPower.RawBytePower.String()
			}
			qap := ""
			if &mpower.MinerPower.QualityAdjPower != nil {
				qap = mpower.MinerPower.QualityAdjPower.String()
			}
			claimedpowerlist = append(claimedpowerlist, &powermodel.PowerActorClaim{
				MinerID:         addr.String(),
				Height:          int64(ts.Height()),
				StateRoot:       "",
				RawBytePower:    rbp,
				QualityAdjPower: qap,
			})

			/*
				for _, s := range activeSectors {
					smsl := make([]minermodel.MinerSectorInfo, 1)
					tsPSstr := ""
					if ts != nil {
						tsPS := ts.ParentState()
						tsPSstr = tsPS.String()
					}
					smsl = append(smsl, minermodel.MinerSectorInfo{
						Height:                int64(ts.Height()),
						MinerID:               addr.String(),
						SectorID:              uint64(s.SectorNumber),
						StateRoot:             tsPSstr,
						SealedCID:             s.SealedCID.String(),
						ActivationEpoch:       int64(s.Activation),
						ExpirationEpoch:       int64(s.Expiration),
						DealWeight:            s.DealWeight.String(),
						VerifiedDealWeight:    s.VerifiedDealWeight.String(),
						InitialPledge:         s.InitialPledge.String(),
						ExpectedDayReward:     s.ExpectedDayReward.String(),
						ExpectedStoragePledge: s.ExpectedStoragePledge.String(),
					})
					p.store.PersistMinerSectors(smsl)
					// minersectorslist = append(minersectorslist, minermodel.MinerSectorInfo{
					// 	Height:                int64(ts.Height()),
					// 	MinerID:               addr.String(),
					// 	SectorID:              uint64(s.SectorNumber),
					// 	StateRoot:             ts.ParentState().String(),
					// 	SealedCID:             s.SealedCID.String(),
					// 	ActivationEpoch:       int64(s.Activation),
					// 	ExpirationEpoch:       int64(s.Expiration),
					// 	DealWeight:            s.DealWeight.String(),
					// 	VerifiedDealWeight:    s.VerifiedDealWeight.String(),
					// 	InitialPledge:         s.InitialPledge.String(),
					// 	ExpectedDayReward:     s.ExpectedDayReward.String(),
					// 	ExpectedStoragePledge: s.ExpectedStoragePledge.String(),
					// })
				}
			*/
			for _, fs := range fsa {
				minersectorfaultslist = append(minersectorfaultslist, &minermodel.MinerSectorFault{
					Height:   int64(ts.Height()),
					MinerID:  addr.String(),
					SectorID: fs,
				})
			}

			ec, err := NewMinerStateExtractionContext(p, context.Background(), addr, ts)
			if err != nil {
				log.Println(err)
			} else {
				mcdi, err := ExtractMinerCurrentDeadlineInfo(ec, addr, ts)
				if err != nil {
					log.Println(err)
				} else {
					minerdeadlineslist = append(minerdeadlineslist, mcdi)
				}
				mlf, err := ExtractMinerLockedFunds(ec, addr, ts)
				if err != nil {
					log.Println(err)
				} else {
					minerfundslist = append(minerfundslist, mlf)
				}
			}

			// wg.Done()
			// }(addr)
		// }
		// p.store.PersistMinerInfos(minerinfoslist)
		// p.store.PersistPowerActorClaims(claimedpowerlist)
		// // p.store.PersistMinerSectors(minersectorslist)
		// p.store.PersistMinerSectorFaults(minersectorfaultslist)
		// p.store.PersistMinerDeadlines(minerdeadlineslist)
		// p.store.PersistMinerFunds(minerfundslist)
	}
	// }
	// wg.Wait()
	// ***********
	p.store.PersistMinerInfos(minerinfoslist)
	p.store.PersistPowerActorClaims(claimedpowerlist)
	// p.store.PersistMinerSectors(minersectorslist)
	p.store.PersistMinerSectorFaults(minersectorfaultslist)
	p.store.PersistMinerDeadlines(minerdeadlineslist)
	p.store.PersistMinerFunds(minerfundslist)

	// p.store.PersistBatch(minerinfoslist, "miner_info")
	// p.store.PersistBatch(minersectorslist, "miner_sector_info")
	// p.store.PersistBatch(minersectorfaultslist, "miner_sector_fault")

	return data, nil
}

func (p *MinerProcessor) Close() error {
	if p.closer != nil {
		p.closer()
		p.closer = nil
	}
	p.node = nil
	return nil
}

func NewMinerStateExtractionContext(p *MinerProcessor, ctx context.Context, addr address.Address, ts *types.TipSet) (*MinerStateExtractionContext, error) {
	tsk := ts.Key()
	curActor, err := p.node.StateGetActor(ctx, addr, tsk)
	if err != nil {
		return nil, err
	}
	log.Info("CGTS", tsk, curActor)
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
		log.Info("TSP", ts.Parents())
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

func ExtractMinerLockedFunds(ec *MinerStateExtractionContext, addr address.Address, ts *types.TipSet) (*minermodel.MinerFund, error) {
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
		AvailableBalance:  "",
	}, nil
}
