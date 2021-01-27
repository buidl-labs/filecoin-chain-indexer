package messages

import (
	"context"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

type ParsedMessage struct {
	tableName struct{} `pg:"parsed_messages"` // nolint: structcheck,unused
	Height    int64    `pg:",pk,notnull,use_zero"`
	Cid       string   `pg:",pk,notnull"`
	From      string   `pg:",notnull"`
	To        string   `pg:",notnull"`
	Value     string   `pg:",notnull"`
	Method    string   `pg:",notnull"`

	Params string `pg:",type:jsonb,notnull"`
}

func (pm *ParsedMessage) Persist(ctx context.Context, s model.StorageBatch) error {
	return s.PersistModel(ctx, pm)
}

type ParsedMessages []*ParsedMessage

func (pms ParsedMessages) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(pms) == 0 {
		return nil
	}

	return s.PersistModel(ctx, pms)
}
