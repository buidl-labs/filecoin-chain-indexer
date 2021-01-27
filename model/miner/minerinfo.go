package miner

type MinerInfo struct {
	tableName       struct{} `pg:"miner_infos"` // nolint: structcheck,unused
	Height          int64    `pg:",pk,notnull,use_zero"`
	MinerID         string   `pg:",pk,notnull"`
	OwnerID         string   `pg:",notnull"`
	WorkerID        string   `pg:",notnull"`
	Address         string
	PeerID          string
	StorageAskPrice string
	MinPieceSize    uint64
	MaxPieceSize    uint64
}
