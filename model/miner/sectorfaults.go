package miner

type MinerSectorFault struct {
	tableName struct{} `pg:"miner_sector_faults"` // nolint: structcheck,unused
	Height    int64    `pg:",pk,notnull,use_zero"`
	MinerID   string   `pg:",pk,notnull"`
	SectorID  uint64   `pg:",pk,use_zero"`
}
