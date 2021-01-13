package chain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/filecoin-project/lotus/chain/types"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	minermodel "github.com/buidl-labs/filecoin-chain-indexer/model/miner"
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
	fmt.Println("mtsk", tsk)
	addresses, err := p.node.StateListMiners(context.Background(), tsk)
	if err != nil {
		fmt.Println("SLMerr", err)
		return data, err
	}

	fmt.Println("SLM addresses", addresses)
	for _, addr := range addresses {
		fmt.Println("miner", addr)
		info, err := p.node.StateMinerInfo(context.Background(), addr, tsk)
		if err != nil {
			fmt.Println("SLMinfoerr", err)
			return data, err
		}
		fmt.Println("SLMInfo: {minerid:", addr, "ownerid:", info.Owner, "workerid:", info.Worker, "PeerId:", info.PeerId, "SectorSize:", info.SectorSize, "}")
		mpower, err := p.node.StateMinerPower(context.Background(), addr, tsk)
		if err != nil {
			fmt.Println("SLMpowererr", err)
			return data, err
		}
		fmt.Println("SLMpower", "raw:", mpower.MinerPower.RawBytePower, "totalraw:", mpower.TotalPower.RawBytePower, "qadj:", mpower.MinerPower.QualityAdjPower, "totalqadj:", mpower.TotalPower.QualityAdjPower)

		// ask, err := p.node.ClientQueryAsk(context.Background(), *info.PeerId, addr)
		// if err != nil {
		// 	fmt.Println("SLMCLientqueryask", err)
		// 	return data, err
		// }
		// fmt.Println("SLMAsk: {minerid:", ask.Miner, "price:", ask.Price, "verifiedP:", ask.VerifiedPrice, "minPS:", ask.MinPieceSize, "maxPS:", ask.MaxPieceSize, "timestamp:", ask.Timestamp, "Expiry:", ask.Expiry, "}")
		fmt.Println("**********")

		p.store.PersistMinerInfos(minermodel.MinerInfo{
			MinerID:         addr.String(),
			Address:         "",
			PeerID:          info.PeerId.String(),
			OwnerID:         info.Owner.String(),
			WorkerID:        info.Worker.String(),
			Height:          int64(ts.Height()),
			StorageAskPrice: "",
			MinPieceSize:    uint64(0),
			MaxPieceSize:    uint64(0),
		})

		p.store.PersistMinerQuality(minermodel.MinerQuality{
			MinerID:          addr.String(),
			QualityAdjPower:  mpower.MinerPower.QualityAdjPower,
			RawBytePower:     mpower.MinerPower.RawBytePower,
			WinCount:         uint64(0),
			DataStored:       "0.0",
			BlocksMined:      big.NewInt(0),
			MiningEfficiency: "0.0",
			FaultySectors:    uint64(0),
		})
	}

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
