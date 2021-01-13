package storage

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

var _ model.Storage = (*NullStorage)(nil)

// A NullStorage ignores any requests to persist a model
type NullStorage struct {
}

func (*NullStorage) PersistBatch(ctx context.Context, p ...model.Persistable) error {
	log.Print("Not persisting data")
	return nil
}

func (*NullStorage) PersistModel(ctx context.Context, m interface{}) error {
	log.Print("Not persisting data")
	return nil
}
