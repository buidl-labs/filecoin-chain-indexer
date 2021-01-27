package indexing

type ParsedTill struct {
	tableName struct{} `pg:"parsed_till"` // nolint: structcheck,unused
	Height    int64    `pg:",pk,notnull,use_zero"`
}
