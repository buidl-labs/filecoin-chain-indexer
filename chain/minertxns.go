package chain

import (
	"context"
	// "encoding/json"
	"fmt"
	// "io/ioutil"
	"os"
	// "sync"

	// "github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	// "github.com/filecoin-project/specs-actors/v2/actors/builtin"

	// builtin2 "github.com/filecoin-project/specs-actors/v2/actors/builtin"
	"github.com/ipfs/go-cid"
	// log "github.com/sirupsen/logrus"
	"github.com/filecoin-project/go-address"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	// "github.com/buidl-labs/filecoin-chain-indexer/lens/lotus"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	messagemodel "github.com/buidl-labs/filecoin-chain-indexer/model/messages"
)

type MinerTxnsProcessor struct {
	node       lens.API
	opener     lens.APIOpener
	closer     lens.APICloser
	lastTipSet *types.TipSet
	store      db.Store
}

func NewMinerTxnsProcessor(opener lens.APIOpener, store db.Store) *MinerTxnsProcessor {
	return &MinerTxnsProcessor{
		opener: opener,
		store:  store,
	}
}

func (p *MinerTxnsProcessor) ProcessTipSet(ctx context.Context, ts *types.TipSet) (model.Persistable, error) {
	if p.node == nil {
		node, closer, err := p.opener.Open(ctx)
		if err != nil {
			return nil, xerrors.Errorf("unable to open lens: %w", err)
		}
		p.node = node
		p.closer = closer
	}

	var data model.Persistable
	// var err error

	// epoch := abi.ChainEpoch(int64(507999))
	epoch := ts.Height()

	// addrs, err := ActiveMinerAddresses()
	// if err != nil {
	// 	return nil, err
	// }
	// addrs = addrs[:15]

	// plan, _ := ioutil.ReadFile(os.Getenv("ACTOR_CODES_JSON"))
	// var results map[string]string // address.Address]cid.Cid
	// err = json.Unmarshal([]byte(plan), &results)
	// if err != nil {
	// 	fmt.Println("load actorCodes: %w", err)
	// 	return nil, err
	// 	// return nil, xerrors.Errorf("load actorCodes: %w", err)
	// }
	// fmt.Println("HII", "f099", results["f099"])
	actorCodes := map[address.Address]cid.Cid{}
	// actorCodesStr := map[string]string{}
	// for k, v := range results {
	// 	actorCodesStr[k] = v
	// 	a, _ := address.NewFromString(k)
	// 	c, _ := cid.Decode(v)
	// 	actorCodes[a] = c
	// }

	fmt.Println("epoch ", epoch)
	// dir := totalAddrs / 120
	// for _, addr := range addrs {
	// 	fetchMessages(p, actorCodes, addr, ctx, epoch, ts)
	// }

	// addr:= "f09037"
	minerID := os.Getenv("MINERID")
	addr, _ := address.NewFromString(minerID)
	fetchMessages(p, actorCodes, addr, ctx, epoch, ts)

	/*
		l := 120
		totalAddrs := len(addrs)
		rem := totalAddrs % l
		for i := 0; i < totalAddrs-rem && i+l <= totalAddrs; i += l {
			fmt.Println("i ", i)
			var wg sync.WaitGroup
			wg.Add(l)
			for _, addr := range addrs[i : i+l] {
				fmt.Println("addr ", addr, " i", i, " i+l ", i+l)
				go func(addr address.Address) {
					fetchMessages(p, actorCodes, addr, ctx, epoch, ts)
					wg.Done()
				}(addr)
			}
			fmt.Println("bfw")
			wg.Wait()
			fmt.Println("donewt")
		}

		var wg2 sync.WaitGroup
		wg2.Add(rem)
		for i, addr := range addrs[totalAddrs-rem:] {
			fmt.Println("addr ", addr, " i", i) //, " i+l ", i+l)
			go func(addr address.Address) {
				fetchMessages(p, actorCodes, addr, ctx, epoch, ts)
				wg2.Done()
			}(addr)
		}
		fmt.Println("bfw")
		wg2.Wait()
		fmt.Println("donewt")
	*/

	return data, nil
}

func fetchMessages(p *MinerTxnsProcessor, actorCodes map[address.Address]cid.Cid, addr address.Address, ctx context.Context, epoch abi.ChainEpoch, ts *types.TipSet) {
	// fmt.Println("miner ", addr)

	// getActorCode := func(a address.Address) cid.Cid {
	// 	c, ok := actorCodes[a]
	// 	if ok {
	// 		return c
	// 	}

	// 	return cid.Undef
	// }
	var cids []cid.Cid
	tsk := ts.Key()
	var mtxns []*messagemodel.MinerTransaction
	from, err := p.node.StateListMessages(context.Background(), &api.MessageMatch{From: addr}, tsk, epoch)
	if err != nil {
		// return nil, err
		fmt.Println("none ", addr, err)
	}
	for _, f := range from {
		msg, err := p.node.ChainGetMessage(context.Background(), f)
		if err == nil {
			fmt.Println(addr.String() == msg.From.String(), "msg ", msg.From, msg.To, msg.Method, msg.Value)
			mtxns = append(mtxns, &messagemodel.MinerTransaction{
				Height:   int64(epoch),
				Cid:      f.String(),
				Sender:   msg.From.String(),
				Receiver: msg.To.String(),
				Amount:   msg.Value.String(),
				// GasFeeCap:     msg.GasFeeCap.String(),
				// GasPremium:    msg.GasPremium.String(),
				// GasLimit:      msg.GasLimit,
				// Nonce:         msg.Nonce,
				Method: uint64(msg.Method),
				// FromActorName: lotus.ActorNameByCode(getActorCode(msg.From)),
				// ToActorName:   lotus.ActorNameByCode(getActorCode(msg.To)),
			})
		} else {
			fmt.Println("chaingetmsg", err)
		}
	}
	// p.store.DB.Model(transaction)
	cids = append(cids, from...)

	to, err := p.node.StateListMessages(ctx, &api.MessageMatch{To: addr}, tsk, epoch)
	if err != nil {
		// return nil, err
		fmt.Println("none ", addr, err)
	}
	for _, t := range to {
		msg, err := p.node.ChainGetMessage(context.Background(), t)
		if err == nil {
			fmt.Println(addr.String() == msg.To.String(), "msg ", msg.From, msg.To, msg.Method, msg.Value)
			mtxns = append(mtxns, &messagemodel.MinerTransaction{
				Height:   int64(epoch),
				Cid:      t.String(),
				Sender:   msg.From.String(),
				Receiver: msg.To.String(),
				Amount:   msg.Value.String(),
				// GasFeeCap:     msg.GasFeeCap.String(),
				// GasPremium:    msg.GasPremium.String(),
				// GasLimit:      msg.GasLimit,
				// Nonce:         msg.Nonce,
				Method: uint64(msg.Method),
				// FromActorName: lotus.ActorNameByCode(getActorCode(msg.From)),
				// ToActorName:   lotus.ActorNameByCode(getActorCode(msg.To)),
			})
		} else {
			fmt.Println("chaingetmsg", err)
		}
	}
	cids = append(cids, to...)
	_, err = p.store.DB.Model(&mtxns).Insert()
	if err != nil {
		fmt.Println("couldn't insert mtxns: ", err)
	}
	fmt.Println("miner ", addr, " cids len ", len(cids), " arr: ", cids)
	// for _, cid := range cids {
	// 	msg, _ := p.node.ChainGetMessage(context.Background(), cid)
	// 	fmt.Println("msg ", msg.From, msg.To, msg.Method, msg.Value)
	// }
}

func (p *MinerTxnsProcessor) Close() error {
	if p.closer != nil {
		p.closer()
		p.closer = nil
	}
	p.node = nil
	return nil
}
