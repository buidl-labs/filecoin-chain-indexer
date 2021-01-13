package lotus

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/filecoin-project/lotus/api/client"
	lru "github.com/hashicorp/golang-lru"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
)

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
	fmt.Println("tokenmaddr", tokenMaddr)
	toks := strings.Split(tokenMaddr, ":")
	if len(toks) != 2 {
		return nil, nil, fmt.Errorf("invalid api tokens, expected <token>:<maddr>, got: %s", tokenMaddr)
	}

	rawtoken = toks[0]
	rawaddr = toks[1]

	parsedAddr, err := ma.NewMultiaddr(rawaddr)
	if err != nil {
		return nil, nil, xerrors.Errorf("parse listen address: %w", err)
	}

	_, addr, err := manet.DialArgs(parsedAddr)
	if err != nil {
		return nil, nil, xerrors.Errorf("dial multiaddress: %w", err)
	}

	o := &APIOpener{
		cache:   ac,
		addr:    apiURI(addr),
		headers: apiHeaders(rawtoken),
	}

	return o, lens.APICloser(func() {}), nil
}

func (o *APIOpener) Open(ctx context.Context) (lens.API, lens.APICloser, error) {
	api, closer, err := client.NewFullNodeRPC(ctx, o.addr, o.headers)
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
