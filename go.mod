module github.com/buidl-labs/filecoin-chain-indexer

go 1.15

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/AndreasBriese/bbloom v0.0.0-20190825152654-46b345b51c96 // indirect
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/fatih/color v1.10.0 // indirect
	github.com/filecoin-project/go-address v0.0.5-0.20201103152444-f2023ef3f5bb
	github.com/filecoin-project/go-bitfield v0.2.3-0.20201110211213-fe2c1862e816
	github.com/filecoin-project/go-fil-markets v1.0.4
	github.com/filecoin-project/go-jsonrpc v0.1.2-0.20201008195726-68c6a2704e49
	github.com/filecoin-project/go-multistore v0.0.3
	github.com/filecoin-project/go-state-types v0.0.0-20201013222834-41ea465f274f
	github.com/filecoin-project/lotus v1.1.4-0.20201116232018-a152d98af82b
	github.com/filecoin-project/sentinel-visor v0.3.0
	github.com/filecoin-project/specs-actors v0.9.12
	github.com/filecoin-project/specs-actors/v2 v2.2.0
	github.com/filecoin-project/statediff v0.0.8-0.20201027195725-7eaa5391a639
	github.com/go-pg/migrations/v8 v8.0.1
	github.com/go-pg/pg/v10 v10.3.1
	github.com/go-pg/pgext v0.1.4
	github.com/hashicorp/golang-lru v0.5.4
	github.com/ipfs/go-block-format v0.0.2
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-datastore v0.4.5
	github.com/ipfs/go-ipfs-blockstore v1.0.3 // indirect
	github.com/ipfs/go-ipld-cbor v0.0.5
	github.com/ipfs/go-log/v2 v2.1.2-0.20200626104915-0016c0b4b3e4
	github.com/ipld/go-ipld-prime v0.5.1-0.20201021195245-109253e8a018
	github.com/jackc/pgx/v4 v4.9.0
	github.com/jasonlvhit/gocron v0.0.1
	github.com/lib/pq v1.9.0
	github.com/libp2p/go-libp2p-core v0.7.0
	github.com/mattn/go-sqlite3 v1.14.6 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/multiformats/go-multibase v0.0.3
	github.com/multiformats/go-multihash v0.0.14
	github.com/pressly/goose v2.6.0+incompatible
	github.com/prometheus/client_golang v1.6.0
	github.com/raulk/clock v1.1.0
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli/v2 v2.3.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20200826160007-0b9f6c5fb163
	github.com/willscott/carbs v0.0.3
	github.com/ziutek/mymysql v1.5.4 // indirect
	go.opencensus.io v0.22.4
	go.opentelemetry.io/otel v0.12.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.12.0
	go.opentelemetry.io/otel/sdk v0.12.0
	golang.org/x/sys v0.0.0-20201112073958-5cba982894dd // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
)

replace github.com/filecoin-project/filecoin-ffi => github.com/filecoin-project/statediff/extern/filecoin-ffi v0.0.0-20201028183730-8e6631500d49

// replace github.com/supranational/blst => ./extern/fil-blst/blst

// replace github.com/filecoin-project/fil-blst => ./extern/fil-blst

// require (
// 	github.com/figment-networks/filecoin-indexer v0.0.0-20201219063208-e7473de0ccd0
// 	github.com/figment-networks/indexing-engine v0.1.14
// 	github.com/filecoin-project/go-address v0.0.5-0.20201103152444-f2023ef3f5bb
// 	github.com/filecoin-project/go-bitfield v0.2.3-0.20201110211213-fe2c1862e816
// 	github.com/filecoin-project/go-fil-markets v1.1.1
// 	github.com/filecoin-project/go-jsonrpc v0.1.2
// 	github.com/filecoin-project/go-state-types v0.0.0-20201203022337-7cab7f0d4bfb
// 	github.com/filecoin-project/lotus v1.4.0
// 	github.com/filecoin-project/specs-actors v0.9.13
// 	github.com/influxdata/influxdb1-client v0.0.0-20191209144304-8bf82d3c094d
// 	github.com/ipfs/go-cid v0.0.7
// 	github.com/libp2p/go-libp2p-core v0.8.0
// 	github.com/multiformats/go-multiaddr v0.3.1
// 	github.com/prometheus/common v0.10.0
// 	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
// 	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
// )

// replace github.com/filecoin-project/filecoin-ffi => github.com/filecoin-project/statediff/extern/filecoin-ffi v0.0.0-20201028183730-8e6631500d49