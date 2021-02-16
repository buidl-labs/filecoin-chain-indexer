package miner

type MinerInfo struct {
	tableName       struct{} `pg:"miner_infos"` // nolint: structcheck,unused
	Height          int64    `pg:",pk,notnull,use_zero"`
	MinerID         string   `pg:",pk,notnull"`
	StateRoot       string   `pg:",pk,notnull"`
	OwnerID         string   `pg:",notnull"`
	WorkerID        string   `pg:",notnull"`
	PeerID          string   `pg:",notnull"`
	StorageAskPrice string   `pg:",notnull"`
	MinPieceSize    uint64   `pg:",notnull"`
	MaxPieceSize    uint64   `pg:",notnull"`
	Address         string
}
