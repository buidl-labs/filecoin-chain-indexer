package miner

type MinerFunds struct {
	MinerID           string
	LockedFunds       float64
	InitialPledge     float64
	PreCommitDeposits float64
	AvailableBalance  float64
}
