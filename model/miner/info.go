package miner

import "math/big"

type MinerInfo struct {
	MinerID         string
	Address         string
	PeerID          string
	OwnerID         string
	WorkerID        string
	CreatedAt       int64 // epoch
	StorageAskPrice big.Int
	MinPieceSize    uint64
	MaxPieceSize    uint64
}
