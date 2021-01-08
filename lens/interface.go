package lens

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-bitfield"
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/api/apistruct"
	"github.com/filecoin-project/lotus/api/client"
	"github.com/filecoin-project/lotus/chain/actors/builtin/miner"
	"github.com/filecoin-project/lotus/chain/types"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"golang.org/x/xerrors"

	// "github.com/filecoin-project/lotus/chain/vm"
	"github.com/filecoin-project/specs-actors/actors/util/adt"
	"github.com/ipfs/go-cid"
)

type API interface {
	StoreAPI
	ChainAPI
	StateAPI

	// ComputeGasOutputs(gasUsed, gasLimit int64, baseFee, feeCap, gasPremium abi.TokenAmount) vm.GasOutputs
	// GetExecutedMessagesForTipset(ctx context.Context, ts, pts *types.TipSet) ([]*ExecutedMessage, error)
}

type StoreAPI interface {
	Store() adt.Store
}

type ChainAPI interface {
	ChainNotify(context.Context) (<-chan []*api.HeadChange, error)
	ChainHead(context.Context) (*types.TipSet, error)

	ChainHasObj(ctx context.Context, obj cid.Cid) (bool, error)
	ChainReadObj(ctx context.Context, obj cid.Cid) ([]byte, error)

	ChainGetGenesis(ctx context.Context) (*types.TipSet, error)
	ChainGetTipSet(context.Context, types.TipSetKey) (*types.TipSet, error)
	ChainGetTipSetByHeight(context.Context, abi.ChainEpoch, types.TipSetKey) (*types.TipSet, error)

	ChainGetBlockMessages(ctx context.Context, msg cid.Cid) (*api.BlockMessages, error)
	ChainGetParentMessages(ctx context.Context, blockCid cid.Cid) ([]api.Message, error)
	ChainGetParentReceipts(ctx context.Context, blockCid cid.Cid) ([]*types.MessageReceipt, error)
}

type StateAPI interface {
	StateGetActor(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*types.Actor, error)
	StateListActors(context.Context, types.TipSetKey) ([]address.Address, error)
	StateChangedActors(context.Context, cid.Cid, cid.Cid) (map[string]types.Actor, error)

	StateMinerSectors(ctx context.Context, addr address.Address, bf *bitfield.BitField, tsk types.TipSetKey) ([]*miner.SectorOnChainInfo, error)
	StateMinerPower(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*api.MinerPower, error)

	StateMarketDeals(context.Context, types.TipSetKey) (map[string]api.MarketDeal, error)

	StateReadState(ctx context.Context, addr address.Address, tsk types.TipSetKey) (*api.ActorState, error)
	StateGetReceipt(ctx context.Context, bcid cid.Cid, tsk types.TipSetKey) (*types.MessageReceipt, error)
	StateVMCirculatingSupplyInternal(context.Context, types.TipSetKey) (api.CirculatingSupply, error)
}

type APICloser func()

type APIOpener interface {
	Open(context.Context) (API, APICloser, error)
}

// Lens fetches data from a Filecoin node
type Lens struct {
	api    *apistruct.FullNodeStruct
	closer jsonrpc.ClientCloser

	Epoch       epochClient
	Miner       minerClient
	Transaction transactionClient
	Account     accountClient
}

// New creates a Filecoin client
func New(endpoint string) (*Lens, error) {
	var api apistruct.FullNodeStruct

	ctx := context.Background()
	// addr := "ws://" + endpoint + "/rpc/v0"
	// outs := []interface{}{&api.Internal, &api.CommonStruct.Internal}

	// DATABASE_DSN="postgresql://localhost:5432/filecoin_indexer"
	// RPC_ENDPOINT="filecoin.infura.io"
	// 1lERSegOP3IOaT8IMjTMam0ZcRP:5196eb0a290a324575f167462ee95d95
	// headers := http.Header{}
	// headers.Add("Authorization", "Basic "+"MWxFUlNlZ09QM0lPYVQ4SU1qVE1hbTBaY1JQOjUxOTZlYjBhMjkwYTMyNDU3NWYxNjc0NjJlZTk1ZDk1")
	// closer, err := jsonrpc.NewMergeClient(ctx, addr, "Filecoin", outs, headers)

	tokenMaddr := os.Getenv("FULLNODE_API_INFO")
	fmt.Println("tokenmaddr", tokenMaddr)
	toks := strings.Split(tokenMaddr, ":")
	if len(toks) != 2 {
		return nil, fmt.Errorf("invalid api tokens, expected <token>:<maddr>, got: %s", tokenMaddr)
	}
	rawtoken := toks[0]
	rawaddr := toks[1]

	parsedAddr, err := ma.NewMultiaddr(rawaddr)
	if err != nil {
		return nil, xerrors.Errorf("parse listen address: %w", err)
	}

	_, addr, err := manet.DialArgs(parsedAddr)
	if err != nil {
		return nil, xerrors.Errorf("dial multiaddress: %w", err)
	}

	uri := apiURI(addr)
	headers := apiHeaders(rawtoken)

	_, closer, err := client.NewFullNodeRPC(ctx, uri, headers)
	if err != nil {
		return nil, xerrors.Errorf("new full node rpc: %w", err)
	}

	// closer, err := jsonrpc.NewMergeClient(ctx, addr, "Filecoin", outs, http.Header{})
	// if err != nil {
	// 	return nil, err
	// }

	return &Lens{
		api:    &api,
		closer: closer,

		Epoch:       epochClient{api: &api},
		Miner:       minerClient{api: &api},
		Transaction: transactionClient{api: &api},
		Account:     accountClient{api: &api},
	}, nil
}

// Close closes the websocket connection
func (l *Lens) Close() {
	fmt.Println("calling closer")
	l.closer()
}

func apiURI(addr string) string {
	return "ws://" + addr + "/rpc/v0"
}

func apiHeaders(token string) http.Header {
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token)
	return headers
}
