package miner

import "math/big"

type MinerQuality struct {
	MinerID          string
	QualityAdjPower  big.Int
	RawBytePower     big.Int
	WinCount         uint64
	DataStored       big.Int
	BlocksMined      big.Int
	MiningEfficiency uint64
	FaultySectors    uint64
}
