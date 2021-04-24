package services

import (
	"fmt"

	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/db"
)

func Partition(cfg config.Config) error {
	store, err := db.New(cfg.DBConnStr)
	if err != nil {
		log.Errorw("setup indexer, connecting db", "error", err)
		return xerrors.Errorf("setup indexer, connecting db: %w", err)
	}
	for i := 0; i < 900000; i += 5000 {
		sql := fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS transactions_%v_%v PARTITION OF transactions FOR VALUES FROM (%v) TO (%v);`,
			i, i+5000, i, i+5000,
		)
		_, err := store.PqDB.Exec(sql)
		if err != nil {
			log.Errorw("create transactions partition",
				"i", i,
				"i+5000", i+5000,
				"error", err,
			)
			return err
		}

        sql = fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS parsed_messages_%v_%v PARTITION OF parsed_messages FOR VALUES FROM (%v) TO (%v);`,
			i, i+5000, i, i+5000,
		)
		_, err = store.PqDB.Exec(sql)
		if err != nil {
			log.Errorw("create parsed_messages partition",
				"i", i,
				"i+5000", i+5000,
				"error", err,
			)
			return err
		}
	}
	return nil
}
