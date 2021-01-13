package miner

import "math/big"

type MinerFunds struct {
	MinerID           string
	Height            big.Int
	LockedFunds       float64
	InitialPledge     float64
	PreCommitDeposits float64
	AvailableBalance  float64
}
