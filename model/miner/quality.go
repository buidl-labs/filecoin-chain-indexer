package miner

// filbig "github.com/filecoin-project/go-state-types/big"

type MinerQuality struct {
	MinerID          string
	Height           int64
	QualityAdjPower  string
	RawBytePower     string
	WinCount         uint64
	DataStored       string
	BlocksMined      uint64
	MiningEfficiency string
	FaultySectors    uint64
}
