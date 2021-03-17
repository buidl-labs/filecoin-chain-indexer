package services

import (
	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/pressly/goose"
)

func RunMigrations(cfg config.Config, cmd string) error {
	store, err := db.New(cfg.DBConnStr)
	if err != nil {
		return err
	}
	defer store.Close()

	conn, err := store.Conn()
	if err != nil {
		return err
	}

	switch cmd {
	case "migrate":
		return goose.Up(conn, "db/migrations")
	case "rollback":
		return goose.Down(conn, "db/migrations")
	default:
		return nil
	}
}
