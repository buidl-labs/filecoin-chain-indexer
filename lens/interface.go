package lens

import (
	"context"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/vm"
	"github.com/filecoin-project/specs-actors/actors/util/adt"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"
)

var logger = logging.Logger("lens")

type API interface {
	StoreAPI
	ClientAPI
	ChainAPI
	StateAPI

	GetExecutedMessagesForTipset(ctx context.Context, ts, pts *types.TipSet) ([]*ExecutedMessage, error)
	IndexActorCodes(ctx context.Context, ts *types.TipSet) error
}

type StoreAPI interface {
	Store() adt.Store
}

type ClientAPI interface {
	ClientQueryAsk(ctx context.Context, p peer.ID, miner address.Address) (*storagemarket.StorageAsk, error)
}

type ChainAPI interface {
	ChainNotify(context.Context) (<-chan []*api.HeadChange, error)
	ChainHead(context.Context) (*types.TipSet, error)

	ChainHasObj(ctx context.Context, obj cid.Cid) (bool, error)
	ChainReadObj(ctx context.Context, obj cid.Cid) ([]byte, error)

	ChainGetGenesis(ctx context.Context) (*types.TipSet, error)
	ChainGetTipSet(context.Context, types.TipSetKey) (*types.TipSet, error)
	ChainGetTipSetByHeight(context.Context, abi.ChainEpoch, types.TipSetKey) (*types.TipSet, error)

	ChainGetMessage(context.Context, cid.Cid) (*types.Message, error)
	ChainGetBlockMessages(ctx context.Context, msg cid.Cid) (*api.BlockMessages, error)
	ChainGetParentMessages(ctx context.Context, blockCid cid.Cid) ([]api.Message, error)
	ChainGetParentReceipts(ctx context.Context, blockCid cid.Cid) ([]*types.MessageReceipt, error)
}

type StateAPI interface {
	StateAllMinerFaults(ctx context.Context, lookback abi.ChainEpoch, ts types.TipSetKey) ([]*api.Fault, error)
	StateChangedActors(context.Context, cid.Cid, cid.Cid) (map[string]types.Actor, error)
	StateCompute(context.Context, abi.ChainEpoch, []*types.Message, types.TipSetKey) (*api.ComputeStateOutput, error)
	StateDecodeParams(ctx context.Context, toAddr address.Address, method abi.MethodNum, params []byte, tsk types.TipSetKey) (interface{}, error)
	StateGetActor(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*types.Actor, error)
	StateListActors(context.Context, types.TipSetKey) ([]address.Address, error)
	StateListMiners(context.Context, types.TipSetKey) ([]address.Address, error)
	StateListMessages(ctx context.Context, match *api.MessageMatch, tsk types.TipSetKey, toht abi.ChainEpoch) ([]cid.Cid, error)
	StateLookupID(context.Context, address.Address, types.TipSetKey) (address.Address, error)
	StateAccountKey(context.Context, address.Address, types.TipSetKey) (address.Address, error)
	StateMarketDeals(context.Context, types.TipSetKey) (map[string]api.MarketDeal, error)
	StateMinerAvailableBalance(context.Context, address.Address, types.TipSetKey) (types.BigInt, error)
	StateMinerFaults(context.Context, address.Address, types.TipSetKey) (bitfield.BitField, error)
	StateMinerInfo(context.Context, address.Address, types.TipSetKey) (miner.MinerInfo, error)
	StateMinerSectors(ctx context.Context, addr address.Address, bf *bitfield.BitField, tsk types.TipSetKey) ([]*miner.SectorOnChainInfo, error)
	StateMinerActiveSectors(context.Context, address.Address, types.TipSetKey) ([]*miner.SectorOnChainInfo, error)
	StateMinerPower(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*api.MinerPower, error)
	StateReadState(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*api.ActorState, error)
	StateSearchMsg(ctx context.Context, from types.TipSetKey, msg cid.Cid, limit abi.ChainEpoch, allowReplaced bool) (*api.MsgLookup, error) //perm:read
	StateVMCirculatingSupplyInternal(context.Context, types.TipSetKey) (api.CirculatingSupply, error)
}

type APICloser func()

type APIOpener interface {
	Open(context.Context) (API, APICloser, error)
}

type MessageMatch struct {
	To   address.Address
	From address.Address
}

type ExecutedMessage struct {
	Cid         cid.Cid
	Height      abi.ChainEpoch
	Message     *types.Message
	Receipt     *types.MessageReceipt
	BlockHeader *types.BlockHeader
	Blocks      []cid.Cid // blocks this message appeared in
	Index       uint64    // Message and receipt sequence in tipset
	From        address.Address
	To          address.Address
	GasOutputs  vm.GasOutputs
	// FromActorCode cid.Cid   // code of the actor the message is from
	// ToActorCode   cid.Cid   // code of the actor the message is to
}

type Fault struct {
	Miner address.Address
	Epoch abi.ChainEpoch
}

type Deadline struct {
	PostSubmissions bitfield.BitField
}
