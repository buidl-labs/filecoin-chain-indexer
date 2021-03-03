package chain

import (
	"context"
	"fmt"
	// "os"

	// "github.com/ipfs/go-cid"
	// log "github.com/sirupsen/logrus"
	// "github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	// marketmodel "github.com/buidl-labs/filecoin-chain-indexer/model/market"
)

type MinerInfoProcessor struct {
	node       lens.API
	opener     lens.APIOpener
	closer     lens.APICloser
	lastTipSet *types.TipSet
	store      db.Store
}

func NewMinerInfoProcessor(opener lens.APIOpener, store db.Store) *MinerInfoProcessor {
	p := &MinerInfoProcessor{
		opener: opener,
		store:  store,
	}
	return p
}

func (p *MinerInfoProcessor) ProcessTipSet(ctx context.Context, ts *types.TipSet) (model.Persistable, error) {
	var data model.Persistable
	if p.node == nil {
		node, closer, err := p.opener.Open(ctx)
		if err != nil {
			return data, xerrors.Errorf("unable to open lens: %w", err)
		}
		p.node = node
		p.closer = closer
	}

	headts, err := p.node.ChainHead(context.Background())
	if err != nil {
		fmt.Println("getting headts", err)
	}
	err = p.node.IndexActorCodes(context.Background(), headts)
	if err != nil {
		fmt.Println("idxactorcodes", err)
	}
	/*
		tsk := ts.Key()
		log.Info("minerinfotsk", tsk, " ht ", ts.Height())
		addrStr := os.Getenv("MINERID")
		addr, err := address.NewFromString(addrStr)
		if err != nil {
			fmt.Println("couldnt form addr")
			return data, err
		}
		mif, err := p.node.StateMinerInfo(context.Background(), addr, tsk)
		if err != nil {
			fmt.Println("couldnt get mif")
			return data, err
		}
		fmt.Println("MIF:\nminer:", addr, " owner:", mif.Owner,
			" worker:", mif.Worker, " caddrs:", mif.ControlAddresses,
			" workerchangeepoch:", mif.WorkerChangeEpoch,
			" newworker:", mif.NewWorker, " maddrs:", mif.Multiaddrs)

		// bcid, err := cid.Decode("bafy2bzacebikmwwxrtrwvjru62snnz6wwtd5le73pk4b5ksixnuawdxyiem24")
		bcid, err := cid.Decode("bafy2bzacect2j3h2swq2tcsxjuk6pn66ng5iptph3ez5e6a6qks7y37p7d3jw")
		if err != nil {
			fmt.Println("cant parse cid")
		}
		fmt.Println("bcid", bcid)
		mlup, err := p.node.StateSearchMsg(context.Background(), bcid)
		if err != nil {
			fmt.Println("cant mllokup")
		}
		fmt.Println("mlup", mlup.Height)
		// msgs, err := p.node.ChainGetParentMessages(context.Background(), ts.Cids()[0])
		// if err != nil {
		// 	return nil, xerrors.Errorf("get parent messages: %w", err)
		// }
		// fmt.Println("MSGs", msgs)
		// rcpts, err := p.node.ChainGetParentReceipts(context.Background(), ts.Cids()[0])
		// if err != nil {
		// 	return nil, xerrors.Errorf("get parent receipts: %w", err)
		// }
		// if len(rcpts) != len(msgs) {
		// 	return nil, xerrors.Errorf("mismatching number of receipts: got %d wanted %d", len(rcpts), len(msgs))
		// }
		// for index, m := range msgs {
		// 	fmt.Println("cid: ", m.Cid, "\nmsg: ", m.Message, "\nrcpt: ", rcpts[index])
		// 	// FromActorCode: getActorCode(m.Message.From),
		// 	// ToActorCode:   getActorCode(m.Message.To),
		// }
	*/
	return data, nil
}

func (p *MinerInfoProcessor) Close() error {
	if p.closer != nil {
		p.closer()
		p.closer = nil
	}
	p.node = nil
	return nil
}
