package lotus

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	miner "github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/state"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	"github.com/filecoin-project/specs-actors/actors/util/adt"
	"github.com/ipfs/go-cid"
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
	return aw.FullNode.StateGetActor(ctx, actor, tsk)
}

func (aw *APIWrapper) StateListActors(ctx context.Context, tsk types.TipSetKey) ([]address.Address, error) {
	return aw.FullNode.StateListActors(ctx, tsk)
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

func (aw *APIWrapper) ComputeGasOutputs(gasUsed, gasLimit int64, baseFee, feeCap, gasPremium abi.TokenAmount) vm.GasOutputs {
	return vm.ComputeGasOutputs(gasUsed, gasLimit, baseFee, feeCap, gasPremium)
}

func (aw *APIWrapper) StateVMCirculatingSupplyInternal(ctx context.Context, tsk types.TipSetKey) (api.CirculatingSupply, error) {
	return aw.FullNode.StateVMCirculatingSupplyInternal(ctx, tsk)
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

	// TODO: load from db (or some key-value store)
	// Listen for new actors and insert them into the store
	// Doing this since loading actors from state tree is time consuming.
	plan, _ := ioutil.ReadFile("actorCodes.json")
	// var data interface{}
	var results map[string]string // address.Address]cid.Cid
	err = json.Unmarshal([]byte(plan), &results)
	if err != nil {
		return nil, xerrors.Errorf("load actorCodes: %w", err)
	}
	fmt.Println("HII", "f099", results["f099"])
	actorCodes := map[address.Address]cid.Cid{}
	for k, v := range results {
		a, _ := address.NewFromString(k)
		c, _ := cid.Decode(v)
		actorCodes[a] = c
	}
	fmt.Println(actorCodes)
	// Build a lookup of actor codes
	// actorCodes := map[address.Address]cid.Cid{}
	// if err := stateTree.ForEach(func(a address.Address, act *types.Actor) error {
	// 	actorCodes[a] = act.Code
	// 	fmt.Println("somerr1", err, a, act.Code)
	// 	return nil
	// }); err != nil {
	// 	fmt.Println("somerr2", err)
	// 	return nil, xerrors.Errorf("iterate actors: %w", err)
	// }
	// fmt.Println("aCs", actorCodes)
	// actorCodesData, err := json.Marshal(actorCodes)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return nil, xerrors.Errorf("json.Marshal: %w", err)
	// }
	// jsonStr := string(actorCodesData)

	// f, _ := os.Create("actors.json")
	// defer f.Close()
	// w := bufio.NewWriter(f)
	// n4, _ := w.WriteString(jsonStr)
	// fmt.Printf("wrote %d bytes\n", n4)
	// w.Flush()

	getActorCode := func(a address.Address) cid.Cid {
		c, ok := actorCodes[a]
		if ok {
			return c
		}

		return cid.Undef
	}

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

	for index, m := range msgs {
		emsgs = append(emsgs, &lens.ExecutedMessage{
			Cid:           m.Cid,
			Height:        pts.Height(),
			Message:       m.Message,
			Receipt:       rcpts[index],
			BlockHeader:   blockHeaders[messageBlocks[m.Cid][0]],
			Blocks:        messageBlocks[m.Cid],
			Index:         uint64(index),
			FromActorCode: getActorCode(m.Message.From),
			ToActorCode:   getActorCode(m.Message.To),
		})
	}

	return emsgs, nil
}
