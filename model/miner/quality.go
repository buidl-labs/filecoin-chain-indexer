package miner

import (
	"math/big"

	filbig "github.com/filecoin-project/go-state-types/big"
)

type MinerQuality struct {
	MinerID          string
	QualityAdjPower  filbig.Int
	RawBytePower     filbig.Int
	WinCount         uint64
	DataStored       string
	BlocksMined      *big.Int
	MiningEfficiency string
	FaultySectors    uint64
}
