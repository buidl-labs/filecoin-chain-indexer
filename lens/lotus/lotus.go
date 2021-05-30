package lotus

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/filecoin-project/lotus/api/client"
	lru "github.com/hashicorp/golang-lru"
	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
)

var log = logging.Logger("lens/lotus")

type APIOpener struct {
	cache   *lru.ARCCache
	addr    string
	headers http.Header
}

func NewAPIOpener(cfg config.Config, cctx context.Context) (*APIOpener, lens.APICloser, error) {
	ac, err := lru.NewARC(cfg.CacheSize)
	if err != nil {
		return nil, nil, xerrors.Errorf("new arc cache: %w", err)
	}

	var rawaddr, rawtoken string

	tokenMaddr := cfg.FullNodeAPIInfo
	toks := strings.Split(tokenMaddr, ":")
	if len(toks) != 2 {
		return nil, nil, xerrors.Errorf("invalid api tokens, expected <token>:<maddr>, got: %s", tokenMaddr)
	}

	rawtoken = toks[0]
	rawaddr = toks[1]

	parsedAddr, err := ma.NewMultiaddr(rawaddr)
	if err != nil {
		return nil, nil, xerrors.Errorf("parse listen address: %w", err)
	}

	_, _, err = manet.DialArgs(parsedAddr)
	if err != nil {
		return nil, nil, xerrors.Errorf("dial multiaddress: %w", err)
	}
	_ = rawtoken

	o := &APIOpener{
		cache: ac,
		// addr:    apiURI(addr),
		// headers: apiHeaders(rawtoken),
		addr:    os.Getenv("LOTUS_RPC_ENDPOINT"),
		headers: http.Header{},
	}

	return o, lens.APICloser(func() {}), nil
}

func (o *APIOpener) Open(ctx context.Context) (lens.API, lens.APICloser, error) {
	api, closer, err := client.NewFullNodeRPCV1(ctx, o.addr, o.headers)
	if err != nil {
		return nil, nil, xerrors.Errorf("new full node rpc: %w", err)
	}

	cacheStore, err := NewCacheCtxStore(ctx, api, o.cache)
	if err != nil {
		return nil, nil, xerrors.Errorf("new cache store: %w", err)
	}

	lensAPI := NewAPIWrapper(api, cacheStore)

	return lensAPI, lens.APICloser(closer), nil
}

func apiURI(addr string) string {
	return "ws://" + addr + "/rpc/v0"
}

func apiHeaders(token string) http.Header {
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token)
	return headers
}
