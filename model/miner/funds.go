package miner

type MinerFund struct {
	MinerID           string
	Height            int64
	StateRoot         string
	LockedFunds       string
	InitialPledge     string
	PreCommitDeposits string
	AvailableBalance  string
}
