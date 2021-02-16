package chain

import (
	"context"
	"fmt"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
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

	tsk := ts.Key()
	log.Info("mtsk", tsk)
	addresses, err := p.node.StateListMiners(context.Background(), tsk)
	if err != nil {
		return data, err
	}

	log.Info("SLM addresses", len(addresses))
	ads := addresses //[181:183]
	// m1, _ := address.NewFromString("f0107995")
	// m2, _ := address.NewFromString("f080444")
	// m3, _ := address.NewFromString("f067170")
	// ads := [3]address.Address{m1, m2, m3}
	// ads := [1]address.Address{m1}
	log.Info("SLM SLICE", len(ads), ads)
	// minerinfoslist := make([]*minermodel.MinerInfo, 0, len(ads))
	// claimedpowerlist := make([]*powermodel.PowerActorClaim, 0, len(ads))
	// minerdeadlineslist := make([]*minermodel.MinerCurrentDeadlineInfo, 0, len(ads))
	// minerfundslist := make([]*minermodel.MinerFund, 0, len(ads))
	// var wg sync.WaitGroup
	// wg.Add(len(addresses))

	// i := 0
	for i, addr := range ads {
		fmt.Println("lindex", i)

		var info miner.MinerInfo
		var mpower *api.MinerPower
		// var allSectors []*miner.SectorOnChainInfo
		var activeSectors []*miner.SectorOnChainInfo
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

		c1 := make(chan *storagemarket.StorageAsk, 1)

		if info.PeerId != nil {
			go func() {
				fmt.Println("infom", info)
				ask, _ := GetClientAsk(p, info, addr)
				c1 <- ask
			}()

			select {
			case ask := <-c1:
				fmt.Println(ask)
				log.Info("SLMAsk: {minerid:", ask.Miner, "price:", ask.Price, "verifiedP:", ask.VerifiedPrice, "minPS:", ask.MinPieceSize, "maxPS:", ask.MaxPieceSize, "timestamp:", ask.Timestamp, "Expiry:", ask.Expiry, "}")
			case <-time.After(1 * time.Second): // set a higher value
				fmt.Println("Clientqueryask out of time :(")
			}
		}

		// ask, err := GetClientAsk(p, info, addr)
		// if err != nil {
		// 	log.Info("SLMCLientqueryask", err)
		// } else {
		// 	log.Info("SLMAsk: {minerid:", ask.Miner, "price:", ask.Price, "verifiedP:", ask.VerifiedPrice, "minPS:", ask.MinPieceSize, "maxPS:", ask.MaxPieceSize, "timestamp:", ask.Timestamp, "Expiry:", ask.Expiry, "}")
		// }

		// allSectors, err = p.node.StateMinerSectors(context.Background(), addr, nil, tsk)
		// if err != nil {
		// 	log.Println(err)
		// }
		activeSectors, err = p.node.StateMinerActiveSectors(context.Background(), addr, tsk)
		if err != nil {
			log.Println(err)
		}
		faultySectors, err := p.node.StateMinerFaults(context.Background(), addr, tsk)
		if err != nil {
			log.Println(err)
		}

		// log.Info("SLMallSec count", len(allSectors))
		log.Info("SLMActSec count", len(activeSectors))
		fsc, _ = faultySectors.Count()
		fsa, _ = faultySectors.All(fsc)
		log.Info("SLMFaultySec count", fsa)
		log.Info("Info", info)

		peerID := "0"
		if info.PeerId != nil {
			peerID = info.PeerId.String()
		}
		ownerID := "0"
		// if info.Owner != nil {
		// 	ownerID = info.Owner.String()
		// }
		ownerID = info.Owner.String()
		workerID := "0"
		// if info.Worker != nil {
		// 	workerID = info.Worker.String()
		// }
		workerID = info.Worker.String()

		mimod := &minermodel.MinerInfo{
			MinerID:         addr.String(),
			Address:         "0",
			PeerID:          peerID,
			OwnerID:         ownerID,
			WorkerID:        workerID,
			Height:          int64(ts.Height()),
			StateRoot:       "0",
			StorageAskPrice: "0",
			MinPieceSize:    uint64(1),
			MaxPieceSize:    uint64(1),
		}
		r1, err := p.store.DB.Model(mimod).Insert()
		if err != nil {
			fmt.Println("MILERR", err)
		} else {
			fmt.Println("MIR", r1)
		}
		mimod = nil

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
			StateRoot:       "0",
			RawBytePower:    rbp,
			QualityAdjPower: qap,
		}
		r1, err = p.store.DB.Model(pac).Insert()
		if err != nil {
			fmt.Println("PALERR", err)
		} else {
			fmt.Println("PAR", r1)
		}
		pac = nil

		for _, s := range activeSectors {
			tsPSstr := ""
			if ts != nil {
				tsPS := ts.ParentState()
				tsPSstr = tsPS.String()
			}

			msi := &minermodel.MinerSectorInfo{
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
			}
			r, err := p.store.DB.Model(msi).Insert()
			if err != nil {
				fmt.Println("MSLERR", err)
			} else {
				fmt.Println("MSR", r)
			}
			msi = nil
		}

		// fmt.Println("MSFLLEN", len(minersectorslist))
		// p.store.PersistMinerSectors(minersectorslist)
		// r, err := p.store.DB.Model(&minersectorslist).Insert()
		// if err != nil {
		// 	fmt.Println("MSLERR", err)
		// } else {
		// 	fmt.Println("MSR", r)

		for _, fs := range fsa {
			// minersectorfaultslist = append(minersectorfaultslist, &minermodel.MinerSectorFault{
			// 	Height:   int64(ts.Height()),
			// 	MinerID:  addr.String(),
			// 	SectorID: fs,
			// })
			msf := &minermodel.MinerSectorFault{
				Height:   int64(ts.Height()),
				MinerID:  addr.String(),
				SectorID: fs,
			}
			r1, err := p.store.DB.Model(msf).Insert()
			if err != nil {
				fmt.Println("MSFLERR", err)
			} else {
				fmt.Println("MSFR", r1)
			}
			msf = nil
		}
		// p.store.PersistMinerSectorFaults(minersectorfaultslist)
		// r1, err := p.store.DB.Model(&minersectorfaultslist).Insert()
		// if err != nil {
		// 	fmt.Println("MSFLERR", err)
		// } else {
		// 	fmt.Println("MSFR", r1)
		// }
		ec, err := NewMinerStateExtractionContext(p, context.Background(), addr, ts)
		if err != nil {
			log.Println(err)
		} else {
			mcdi, err := ExtractMinerCurrentDeadlineInfo(ec, addr, ts)
			if err != nil {
				log.Println(err)
			} else {
				// minerdeadlineslist = append(minerdeadlineslist, mcdi)
				r1, err = p.store.DB.Model(mcdi).Insert()
				if err != nil {
					fmt.Println("MDLERR", err)
				} else {
					fmt.Println("MDR", r1)
				}
			}
			mcdi = nil
			mlf, err := ExtractMinerLockedFunds(ec, addr, ts)
			if err != nil {
				log.Println(err)
			} else {
				// minerfundslist = append(minerfundslist, mlf)
				r1, err = p.store.DB.Model(mlf).Insert()
				if err != nil {
					fmt.Println("MFLERR", err)
				} else {
					fmt.Println("MFR", r1)
				}
			}
			mlf = nil
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

	// p.store.PersistMinerInfos(minerinfoslist)
	// r1, err := p.store.DB.Model(&minerinfoslist).Insert()
	// if err != nil {
	// 	fmt.Println("MILERR", err)
	// } else {
	// 	fmt.Println("MIR", r1)
	// }
	// // p.store.PersistPowerActorClaims(claimedpowerlist)
	// r1, err = p.store.DB.Model(&claimedpowerlist).Insert()
	// if err != nil {
	// 	fmt.Println("PALERR", err)
	// } else {
	// 	fmt.Println("PAR", r1)
	// }

	/*fmt.Println("MSFLLEN", len(minersectorslist))
	// p.store.PersistMinerSectors(minersectorslist)
	r, err = p.store.DB.Model(&minersectorslist).Insert()
	if err != nil {
		fmt.Println("MSLERR", err)
	} else {
		fmt.Println("MSR", r)
	}
	// p.store.PersistMinerSectorFaults(minersectorfaultslist)
	r, err = p.store.DB.Model(&minersectorfaultslist).Insert()
	if err != nil {
		fmt.Println("MSFLERR", err)
	} else {
		fmt.Println("MSFR", r)
	}*/
	// p.store.PersistMinerDeadlines(minerdeadlineslist)
	// r1, err = p.store.DB.Model(&minerdeadlineslist).Insert()
	// if err != nil {
	// 	fmt.Println("MDLERR", err)
	// } else {
	// 	fmt.Println("MDR", r1)
	// }
	// // p.store.PersistMinerFunds(minerfundslist)
	// r1, err = p.store.DB.Model(&minerfundslist).Insert()
	// if err != nil {
	// 	fmt.Println("MFLERR", err)
	// } else {
	// 	fmt.Println("MFR", r1)
	// }

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
		AvailableBalance:  "0",
	}, nil
}

func GetClientAsk(p *MinerProcessor, info miner.MinerInfo, addr address.Address) (*storagemarket.StorageAsk, error) {
	ask, err := p.node.ClientQueryAsk(context.Background(), *info.PeerId, addr)
	if err != nil {
		return ask, err
	}
	return ask, nil
}
