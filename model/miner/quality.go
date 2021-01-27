package miner

// filbig "github.com/filecoin-project/go-state-types/big"

type MinerQuality struct {
	tableName        struct{} `pg:"miner_quality"` // nolint: structcheck,unused
	Height           int64    `pg:",pk,notnull,use_zero"`
	MinerID          string   `pg:",pk,notnull"`
	QualityAdjPower  string
	RawBytePower     string
	WinCount         uint64
	DataStored       string
	BlocksMined      uint64
	MiningEfficiency string
	FaultySectors    uint64
}
