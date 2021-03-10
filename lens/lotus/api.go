package lotus

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	// builtininit "github.com/filecoin-project/lotus/chain/actors/builtin/init"
	miner "github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/state"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	builtin0 "github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/util/adt"
	builtin2 "github.com/filecoin-project/specs-actors/v2/actors/builtin"
	builtin3 "github.com/filecoin-project/specs-actors/v3/actors/builtin"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/lens"
)

func NewAPIWrapper(node api.FullNode, store adt.Store) *APIWrapper {
	return &APIWrapper{
		FullNode: node,
		store:    store,
	}
}

var _ lens.API = &APIWrapper{}

type APIWrapper struct {
	api.FullNode
	store adt.Store
}

func (aw *APIWrapper) Store() adt.Store {
	return aw.store
}

func (aw *APIWrapper) ChainGetBlock(ctx context.Context, msg cid.Cid) (*types.BlockHeader, error) {
	return aw.FullNode.ChainGetBlock(ctx, msg)
}

func (aw *APIWrapper) ChainGetMessage(ctx context.Context, bcid cid.Cid) (*types.Message, error) {
	return aw.FullNode.ChainGetMessage(ctx, bcid)
}

func (aw *APIWrapper) ChainGetBlockMessages(ctx context.Context, msg cid.Cid) (*api.BlockMessages, error) {
	return aw.FullNode.ChainGetBlockMessages(ctx, msg)
}

func (aw *APIWrapper) ChainGetGenesis(ctx context.Context) (*types.TipSet, error) {
	return aw.FullNode.ChainGetGenesis(ctx)
}

func (aw *APIWrapper) ChainGetParentMessages(ctx context.Context, bcid cid.Cid) ([]api.Message, error) {
	return aw.FullNode.ChainGetParentMessages(ctx, bcid)
}

func (aw *APIWrapper) StateGetReceipt(ctx context.Context, bcid cid.Cid, tsk types.TipSetKey) (*types.MessageReceipt, error) {
	return aw.FullNode.StateGetReceipt(ctx, bcid, tsk)
}

func (aw *APIWrapper) ChainGetParentReceipts(ctx context.Context, bcid cid.Cid) ([]*types.MessageReceipt, error) {
	return aw.FullNode.ChainGetParentReceipts(ctx, bcid)
}

func (aw *APIWrapper) ChainGetTipSet(ctx context.Context, tsk types.TipSetKey) (*types.TipSet, error) {
	return aw.FullNode.ChainGetTipSet(ctx, tsk)
}

func (aw *APIWrapper) ChainNotify(ctx context.Context) (<-chan []*api.HeadChange, error) {
	return aw.FullNode.ChainNotify(ctx)
}

func (aw *APIWrapper) ChainReadObj(ctx context.Context, obj cid.Cid) ([]byte, error) {
	return aw.FullNode.ChainReadObj(ctx, obj)
}

func (aw *APIWrapper) StateChangedActors(ctx context.Context, old cid.Cid, new cid.Cid) (map[string]types.Actor, error) {
	return aw.FullNode.StateChangedActors(ctx, old, new)
}

func (aw *APIWrapper) StateGetActor(ctx context.Context, actor address.Address, tsk types.TipSetKey) (*types.Actor, error) {
	// return aw.FullNode.StateGetActor(ctx, actor, tsk)
	// TODO idk how to get a store.ChainStore here
	return lens.OptimizedStateGetActorWithFallback(ctx, aw.Store(), aw.FullNode, aw.FullNode, actor, tsk)
}

func (aw *APIWrapper) StateListActors(ctx context.Context, tsk types.TipSetKey) ([]address.Address, error) {
	return aw.FullNode.StateListActors(ctx, tsk)
}

func (aw *APIWrapper) StateListMessages(ctx context.Context, match *api.MessageMatch, tsk types.TipSetKey, toht abi.ChainEpoch) ([]cid.Cid, error) {
	return aw.FullNode.StateListMessages(ctx, match, tsk, toht)
}

func (aw *APIWrapper) StateMarketDeals(ctx context.Context, tsk types.TipSetKey) (map[string]api.MarketDeal, error) {
	return aw.FullNode.StateMarketDeals(ctx, tsk)
}

func (aw *APIWrapper) StateMinerPower(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*api.MinerPower, error) {
	return aw.FullNode.StateMinerPower(ctx, addr, tsk)
}

func (aw *APIWrapper) StateMinerSectors(ctx context.Context, addr address.Address, filter *bitfield.BitField, tsk types.TipSetKey) ([]*miner.SectorOnChainInfo, error) {
	return aw.FullNode.StateMinerSectors(ctx, addr, filter, tsk)
}

func (aw *APIWrapper) StateLookupID(ctx context.Context, addr address.Address, tsk types.TipSetKey) (address.Address, error) {
	return aw.FullNode.StateLookupID(ctx, addr, tsk)
}

func (aw *APIWrapper) StateAccountKey(ctx context.Context, addr address.Address, tsk types.TipSetKey) (address.Address, error) {
	return aw.FullNode.StateAccountKey(ctx, addr, tsk)
}

func (aw *APIWrapper) StateAllMinerFaults(ctx context.Context, lookback abi.ChainEpoch, tsk types.TipSetKey) ([]*api.Fault, error) {
	return aw.FullNode.StateAllMinerFaults(ctx, lookback, tsk)
}

func (aw *APIWrapper) StateMinerFaults(ctx context.Context, addr address.Address, tsk types.TipSetKey) (bitfield.BitField, error) {
	return aw.FullNode.StateMinerFaults(ctx, addr, tsk)
}

func (aw *APIWrapper) StateMinerAvailableBalance(ctx context.Context, addr address.Address, tsk types.TipSetKey) (types.BigInt, error) {
	return aw.FullNode.StateMinerAvailableBalance(ctx, addr, tsk)
}

// func (aw *APIWrapper) StateMinerDeadlines(ctx context.Context, addr address.Address, tsk types.TipSetKey) ([]api.Deadline, error) {
// 	return aw.FullNode.StateMinerDeadlines(ctx, addr, tsk)
// }

func (aw *APIWrapper) StateReadState(ctx context.Context, actor address.Address, tsk types.TipSetKey) (*api.ActorState, error) {
	return aw.FullNode.StateReadState(ctx, actor, tsk)
}

func (aw *APIWrapper) StateSearchMsg(ctx context.Context, bcid cid.Cid) (*api.MsgLookup, error) {
	return aw.FullNode.StateSearchMsg(ctx, bcid)
}

func (aw *APIWrapper) StateDecodeParams(ctx context.Context, toAddr address.Address, method abi.MethodNum, params []byte, tsk types.TipSetKey) (interface{}, error) {
	return aw.FullNode.StateDecodeParams(ctx, toAddr, method, params, tsk)
}

// func (aw *APIWrapper) ComputeGasOutputs(gasUsed, gasLimit int64, baseFee, feeCap, gasPremium abi.TokenAmount) vm.GasOutputs {
// 	return vm.ComputeGasOutputs(gasUsed, gasLimit, baseFee, feeCap, gasPremium)
// }

func (aw *APIWrapper) StateVMCirculatingSupplyInternal(ctx context.Context, tsk types.TipSetKey) (api.CirculatingSupply, error) {
	return aw.FullNode.StateVMCirculatingSupplyInternal(ctx, tsk)
}

func (aw *APIWrapper) IndexActorCodes(ctx context.Context, ts *types.TipSet) error {
	initt := time.Now()
	log.Info("initial: ", initt)
	stateTree, err := state.LoadStateTree(aw.Store(), ts.ParentState())
	if err != nil {
		return xerrors.Errorf("load state tree: %w", err)
	}
	gotstatetree := time.Now()
	log.Info("gotstatetree: ", gotstatetree)
	// actorAddresses, err := aw.FullNode.StateListActors(ctx, ts.Key())
	// fmt.Println("count: ", len(actorAddresses))

	// initActor, err := stateTree.GetActor(builtininit.Address)
	// if err != nil {
	// 	return xerrors.Errorf("getting init actor: %w", err)
	// }
	// initActorState, err := builtininit.Load(aw.Store(), initActor)
	// if err != nil {
	// 	return xerrors.Errorf("loading init actor state: %w", err)
	// }

	// Build a lookup of actor codes
	actorCodesStr := map[string]string{}
	actorCodes := map[address.Address]cid.Cid{}
	if err := stateTree.ForEach(func(a address.Address, act *types.Actor) error {
		// fmt.Println("a:", a, "c:", act.Code)
		actorCodes[a] = act.Code
		actorCodesStr[a.String()] = act.Code.String()
		return nil
	}); err != nil {
		return xerrors.Errorf("iterate actors: %w", err)
	}
	createacmap := time.Now()
	log.Info("createacmap: ", createacmap)

	actorCodesData, err := json.Marshal(actorCodesStr)
	if err != nil {
		return xerrors.Errorf("json.Marshal: %w", err)
	}
	jsonStr := string(actorCodesData)
	fmt.Println("\n\njsonStr:\n\n", jsonStr)

	f, err := os.OpenFile(os.Getenv("ACTOR_CODES_JSON"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	n4, _ := w.WriteString(jsonStr)
	fmt.Printf("wrote %d bytes\n", n4)
	fmt.Println("finish time:", time.Now(), " time difference:", time.Now().Sub(gotstatetree))
	w.Flush()

	// getActorCode := func(a address.Address) cid.Cid {
	// 	ra, found, err := initActorState.ResolveAddress(a)
	// 	if err != nil || !found {
	// 		return cid.Undef
	// 	}

	// 	c, ok := actorCodes[ra]
	// 	if ok {
	// 		return c
	// 	}

	// 	return cid.Undef
	// }
	return nil
}

// GetExecutedMessagesForTipset returns a list of messages sent as part of pts (parent) with receipts found in ts (child).
// No attempt at deduplication of messages is made.
func (aw *APIWrapper) GetExecutedMessagesForTipset(ctx context.Context, ts, pts *types.TipSet) ([]*lens.ExecutedMessage, error) {
	if !types.CidArrsEqual(ts.Parents().Cids(), pts.Cids()) {
		return nil, xerrors.Errorf("child tipset (%s) is not on the same chain as parent (%s)", ts.Key(), pts.Key())
	}

	fmt.Println("gonnacall LST")
	stateTree, err := state.LoadStateTree(aw.Store(), ts.ParentState())
	if err != nil {
		return nil, xerrors.Errorf("load state tree: %w", err)
	}
	fmt.Println("LST", stateTree)

	parentStateTree, err := state.LoadStateTree(aw.Store(), pts.ParentState())
	if err != nil {
		return nil, xerrors.Errorf("load parent state tree: %w", err)
	}

	// initActor, err := stateTree.GetActor(builtininit.Address)
	// if err != nil {
	// 	return nil, xerrors.Errorf("getting init actor: %w", err)
	// }
	// initActorState, err := builtininit.Load(aw.Store(), initActor)
	// if err != nil {
	// 	return nil, xerrors.Errorf("loading init actor state: %w", err)
	// }

	// actorAddresses, err := aw.FullNode.StateListActors(ctx, ts.Key())
	// TODO: load from db (or some key-value store)
	// Listen for new actors and insert them into the store
	// Doing this since loading actors from state tree is time consuming.
	plan, _ := ioutil.ReadFile(os.Getenv("ACTOR_CODES_JSON"))
	var results map[string]string // address.Address]cid.Cid
	err = json.Unmarshal([]byte(plan), &results)
	if err != nil {
		return nil, xerrors.Errorf("load actorCodes: %w", err)
	}
	// fmt.Println("HII", "f099", results["f099"])
	actorCodes := map[address.Address]cid.Cid{}
	actorCodesStr := map[string]string{}
	for k, v := range results {
		actorCodesStr[k] = v
		a, _ := address.NewFromString(k)
		c, _ := cid.Decode(v)
		actorCodes[a] = c
	}
	// fmt.Println("lenac ", len(actorCodes), "lenaddrs", len(actorAddresses))
	// if len(actorCodes) != len(actorAddresses) {
	// 	c := 0
	// 	for _, addr := range actorAddresses {
	// 		if _, ok := actorCodes[addr]; !ok {
	// 			// fmt.Println("not yet", addr)
	// 			act, err := aw.FullNode.StateGetActor(ctx, addr, ts.Key())
	// 			if err != nil {
	// 				fmt.Println("stategetactor", err)
	// 			} else {
	// 				fmt.Println("addr ", addr, " actc", act.Code)
	// 				actorCodes[addr] = act.Code
	// 				actorCodesStr[addr.String()] = act.Code.String()
	// 			}
	// 			c++
	// 			//do something here
	// 		} else {
	// 			fmt.Println("no issues")
	// 		}
	// 	}
	// 	fmt.Println("COUNTER", c)
	// 	fmt.Println("NEWACTLEN: ", len(actorCodes), " str: ", len(actorCodesStr))
	// 	actorCodesData, err := json.Marshal(actorCodesStr)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return nil, xerrors.Errorf("json.Marshal: %w", err)
	// 	}
	// 	jsonStr := string(actorCodesData)
	// 	f, _ := os.Create(os.Getenv("ACTOR_CODES_JSON"))
	// 	defer f.Close()
	// 	w := bufio.NewWriter(f)
	// 	n4, _ := w.WriteString(jsonStr)
	// 	fmt.Printf("wrote %d bytes\n", n4)
	// 	w.Flush()
	// }

	// Build a lookup of actor codes
	/*
		actorCodes := map[address.Address]cid.Cid{}
		actorCodesStr := map[string]string{}
		if err := stateTree.ForEach(func(a address.Address, act *types.Actor) error {
			actorCodes[a] = act.Code
			actorCodesStr[a.String()] = act.Code.String()
			fmt.Println("somerr1", err, a, act.Code)
			return nil
		}); err != nil {
			fmt.Println("somerr2", err)
			return nil, xerrors.Errorf("iterate actors: %w", err)
		}
		actorCodesData, err := json.Marshal(actorCodesStr)
		if err != nil {
			fmt.Println(err)
			return nil, xerrors.Errorf("json.Marshal: %w", err)
		}
		jsonStr := string(actorCodesData)
		fmt.Println("jsonStr", jsonStr)

		currentTime := time.Now()
		dt := currentTime.Format("2006-01-02")
		f, _ := os.Create("/var/data/actorCodes " + dt + ".json")
		defer f.Close()
		w := bufio.NewWriter(f)
		n4, _ := w.WriteString(jsonStr)
		fmt.Printf("wrote %d bytes\n", n4)
		w.Flush()
	*/

	getActorCode := func(a address.Address) cid.Cid {
		c, ok := actorCodes[a]
		if ok {
			return c
		}

		return cid.Undef
	}
	// getActorCode := func(a address.Address) cid.Cid {
	// 	ra, found, err := initActorState.ResolveAddress(a)
	// 	if err != nil || !found {
	// 		return cid.Undef
	// 	}

	// 	c, ok := actorCodes[ra]
	// 	if ok {
	// 		return c
	// 	}

	// 	return cid.Undef
	// }

	// Build a lookup of which block headers indexed by their cid
	blockHeaders := map[cid.Cid]*types.BlockHeader{}
	for _, bh := range pts.Blocks() {
		blockHeaders[bh.Cid()] = bh
	}

	// Build a lookup of which blocks each message appears in
	messageBlocks := map[cid.Cid][]cid.Cid{}

	for _, blkCid := range pts.Cids() {
		blkMsgs, err := aw.ChainGetBlockMessages(ctx, blkCid)
		if err != nil {
			return nil, xerrors.Errorf("get block messages: %w", err)
		}

		for _, mcid := range blkMsgs.Cids {
			messageBlocks[mcid] = append(messageBlocks[mcid], blkCid)
		}
	}

	// Get messages that were processed in the parent tipset
	msgs, err := aw.ChainGetParentMessages(ctx, ts.Cids()[0])
	if err != nil {
		return nil, xerrors.Errorf("get parent messages: %w", err)
	}
	fmt.Println("MSGs", msgs)

	// Get receipts for parent messages
	rcpts, err := aw.ChainGetParentReceipts(ctx, ts.Cids()[0])
	if err != nil {
		return nil, xerrors.Errorf("get parent receipts: %w", err)
	}

	if len(rcpts) != len(msgs) {
		// logic error somewhere
		return nil, xerrors.Errorf("mismatching number of receipts: got %d wanted %d", len(rcpts), len(msgs))
	}

	// Start building a list of completed message with receipt
	emsgs := make([]*lens.ExecutedMessage, 0, len(msgs))

	// Create a skeleton vm just for calling ShouldBurn
	vmi, err := vm.NewVM(ctx, &vm.VMOpts{
		StateBase: pts.ParentState(),
		Epoch:     pts.Height(),
		Bstore:    &apiBlockstore{api: aw.FullNode}, // sadly vm wraps this to turn it back into an adt.Store
	})
	if err != nil {
		return nil, xerrors.Errorf("creating temporary vm: %w", err)
	}

	for index, m := range msgs {
		em := &lens.ExecutedMessage{
			Cid:           m.Cid,
			Height:        pts.Height(),
			Message:       m.Message,
			Receipt:       rcpts[index],
			BlockHeader:   blockHeaders[messageBlocks[m.Cid][0]],
			Blocks:        messageBlocks[m.Cid],
			Index:         uint64(index),
			FromActorCode: getActorCode(m.Message.From),
			ToActorCode:   getActorCode(m.Message.To),
		}
		// fmt.Println("em: fromAN:", ActorNameByCode(em.FromActorCode), " toAN:", ActorNameByCode(em.ToActorCode), " fromAC:", em.FromActorCode, " toAC: ", em.ToActorCode, " from: ", m.Message.From, " to: ", m.Message.To)

		burn, err := vmi.ShouldBurn(parentStateTree, m.Message, rcpts[index].ExitCode)
		if err != nil {
			return nil, xerrors.Errorf("deciding whether should burn failed: %w", err)
		}

		em.GasOutputs = vm.ComputeGasOutputs(em.Receipt.GasUsed, em.Message.GasLimit, em.BlockHeader.ParentBaseFee, em.Message.GasFeeCap, em.Message.GasPremium, burn)
		emsgs = append(emsgs, em)
	}

	return emsgs, nil
}

type apiBlockstore struct {
	api interface {
		ChainReadObj(context.Context, cid.Cid) ([]byte, error)
		ChainHasObj(context.Context, cid.Cid) (bool, error)
	}
}

func (a *apiBlockstore) Get(c cid.Cid) (blocks.Block, error) {
	data, err := a.api.ChainReadObj(context.Background(), c)
	if err != nil {
		return nil, err
	}

	return blocks.NewBlockWithCid(data, c)
}

func (a *apiBlockstore) Has(c cid.Cid) (bool, error) {
	return a.api.ChainHasObj(context.Background(), c)
}

func (a *apiBlockstore) DeleteBlock(c cid.Cid) error {
	return xerrors.Errorf("DeleteBlock not supported by apiBlockstore")
}

func (a *apiBlockstore) GetSize(c cid.Cid) (int, error) {
	data, err := a.api.ChainReadObj(context.Background(), c)
	if err != nil {
		return 0, err
	}

	return len(data), nil
}

func (a *apiBlockstore) Put(b blocks.Block) error {
	return xerrors.Errorf("Put not supported by apiBlockstore")
}

func (a *apiBlockstore) PutMany(bs []blocks.Block) error {
	return xerrors.Errorf("PutMany not supported by apiBlockstore")
}

func (a *apiBlockstore) AllKeysChan(ctx context.Context) (<-chan cid.Cid, error) {
	return nil, xerrors.Errorf("AllKeysChan not supported by apiBlockstore")
}

func (a *apiBlockstore) HashOnRead(enabled bool) {
}

func ActorNameByCode(c cid.Cid) string {
	switch {
	case builtin0.IsBuiltinActor(c):
		return builtin0.ActorNameByCode(c)
	case builtin2.IsBuiltinActor(c):
		return builtin2.ActorNameByCode(c)
	case builtin3.IsBuiltinActor(c):
		return builtin3.ActorNameByCode(c)
	default:
		return "<unknown>"
	}
}

// // ActorNameByCode returns the name of the actor code. Agnostic to the
// // version of specs-actors.
// func ActorNameByCode(code cid.Cid) string {
// 	if name := builtin.ActorNameByCode(code); name != "<unknown>" {
// 		return name
// 	}
// 	if name := builtin2.ActorNameByCode(code); name != "<unknown>" {
// 		return name
// 	}
// 	return builtin3.ActorNameByCode(code)
// }
