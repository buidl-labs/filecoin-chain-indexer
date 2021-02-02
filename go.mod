module github.com/buidl-labs/filecoin-chain-indexer

go 1.15

require (
	github.com/GeertJohan/go.rice v1.0.2 // indirect
	github.com/btcsuite/btcd v0.21.0-beta // indirect
	github.com/daaku/go.zipexe v1.0.1 // indirect
	github.com/davidlazar/go-crypto v0.0.0-20200604182044-b73af7476f6c // indirect
	github.com/dgraph-io/badger/v3 v3.2011.1
	github.com/elastic/go-sysinfo v1.5.0 // indirect
	github.com/elastic/go-windows v1.0.1 // indirect
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-bitfield v0.2.3
	github.com/filecoin-project/go-cbor-util v0.0.0-20201016124514-d0bbec7bfcc4 // indirect
	github.com/filecoin-project/go-data-transfer v1.2.8 // indirect
	github.com/filecoin-project/go-fil-markets v1.1.5
	github.com/filecoin-project/go-hamt-ipld/v3 v3.0.1 // indirect
	github.com/filecoin-project/go-padreader v0.0.0-20201016201355-9c5eb1faedb5 // indirect
	github.com/filecoin-project/go-state-types v0.0.0-20210119062722-4adba5aaea71
	github.com/filecoin-project/go-statemachine v0.0.0-20200925172917-aaed5359be39 // indirect
	github.com/filecoin-project/lotus v1.5.0-pre1
	github.com/filecoin-project/specs-actors v0.9.13
	github.com/filecoin-project/specs-actors/v2 v2.3.4
	github.com/gbrlsnchs/jwt/v3 v3.0.0 // indirect
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4
	github.com/ipfs/go-block-format v0.0.2
	github.com/ipfs/go-cid v0.0.7
	github.com/ipld/go-ipld-prime v0.7.0 // indirect
	github.com/lib/pq v1.9.0
	github.com/libp2p/go-libp2p-core v0.8.0
	github.com/magefile/mage v1.11.0 // indirect
	github.com/multiformats/go-multiaddr v0.3.1
	github.com/polydawn/refmt v0.0.0-20201211092308-30ac6d18308e // indirect
	github.com/pressly/goose v2.7.0+incompatible
	github.com/prometheus/procfs v0.3.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/shirou/gopsutil v3.21.1+incompatible // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/urfave/cli/v2 v2.3.0 // indirect
	github.com/whyrusleeping/cbor-gen v0.0.0-20210118024343-169e9d70c0c2
	go.opencensus.io v0.22.6 // indirect
	go.opentelemetry.io/otel v0.12.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/mod v0.4.1 // indirect
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	golang.org/x/tools v0.1.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
	honnef.co/go/tools v0.0.1-2020.1.3 // indirect
	howett.net/plist v0.0.0-20201203080718-1454fab16a06 // indirect
)

replace (
	github.com/filecoin-project/fil-blst => ./extern/fil-blst
	github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi-stub
	github.com/supranational/blst => ./extern/fil-blst/blst
)
