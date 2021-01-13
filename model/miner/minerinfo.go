package miner

type MinerInfo struct {
	MinerID         string
	Address         string
	PeerID          string
	OwnerID         string
	WorkerID        string
	Height          int64
	StorageAskPrice string
	MinPieceSize    uint64
	MaxPieceSize    uint64
}
