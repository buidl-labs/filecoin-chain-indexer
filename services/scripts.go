package services

import (
	// "bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// power0 "github.com/filecoin-project/specs-actors/actors/builtin/power"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/model/messages"
	"github.com/buidl-labs/filecoin-chain-indexer/util"
	"github.com/go-pg/pg/v10"
)

// func expectedBlockWinRate() {
// 	qpercI := types.BigDiv(types.BigMul(pow.MinerPower.QualityAdjPower, types.NewInt(1000000)), pow.TotalPower.QualityAdjPower)
// 	expWinChance := float64(types.BigMul(qpercI, types.NewInt(build.BlocksPerEpoch)).Int64()) / 1000000
// 	if expWinChance > 0 {
// 		if expWinChance > 1 {
// 			expWinChance = 1
// 		}
// 		winRate := time.Duration(float64(time.Second*time.Duration(build.BlockDelaySecs)) / expWinChance)
// 		winPerDay := float64(time.Hour*24) / float64(winRate)

// 		fmt.Print("Expected block win rate: ")
// 		color.Blue("%.4f/day (every %s)", winPerDay, winRate.Truncate(time.Second))
// 	}

// }

func UpdateTransactionMiner(cfg config.Config) error {
	from := cfg.From
	to := cfg.To

	if from > to {
		log.Errorf("%d > %d\n", from, to)
		return xerrors.Errorf("%d > %d", from, to)
	}

	store, err := db.New(cfg.DBConnStr)
	if err != nil {
		log.Errorw("setup indexer, connecting db", "error", err)
		return xerrors.Errorf("setup indexer, connecting db: %w", err)
	}

	tx, err := store.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Close()
	if err := setMiner(context.Background(), tx, from, to); err != nil {
		// Rollback on error.
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}
	return nil
}

func setMiner(ctx context.Context, tx *pg.Tx, from, to int64) error {
	var transactions []*messages.Transaction
	err := tx.Model(transactions).
		Where("height >= ? AND height <= ?", from, to).
		Select()
	if err != nil {
		return err
	}
	log.Info("count", len(transactions))
	for _, transaction := range transactions {
		var miner string = "0"
		miner = util.DeriveMiner(transaction, miner)
		tx.Model(transactions).
			Set("miner = ?", miner).
			Where("id = ?", transaction.Cid).
			Update()
	}
	return nil
}

func FixCsvs(cfg config.Config) error {
	projectRoot := os.Getenv("ROOTDIR")
	store, err := db.New(cfg.DBConnStr)
	if err != nil {
		log.Errorw("setup indexer, connecting db", "error", err)
		return xerrors.Errorf("setup indexer, connecting db: %w", err)
	}

	var files []string

	root := projectRoot + "/s3data/csvs/transactions"
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fmt.Println(file)
		ss := strings.Split(file, "/")
		// os.Open(file)
		filename := ss[len(ss)-1]
		ee := strings.Split(filename, ".")
		epoch := ee[0]
		log.Info("epoch:", epoch)

		_, err = store.PqDB.Exec("copy tmp_parsed_messages from '" + projectRoot + "/s3data/csvs/parsed_messages/" + epoch + ".csv' CSV HEADER")
		if err != nil {
			log.Error("copy tmp_parsed_messages", err)
			continue
			// return err
		}
		_, err = store.PqDB.Exec("copy tmp_transactions from '" + projectRoot + "/s3data/csvs/transactions/" + epoch + ".csv' CSV HEADER")
		if err != nil {
			log.Error("copy tmp_transactions", err)
			continue
			// return err
		}

		_, err = store.PqDB.Exec(`INSERT INTO parsed_messages SELECT * FROM `+
			`tmp_parsed_messages WHERE height = $1 `+
			`ON CONFLICT DO NOTHING`, string(epoch))
		if err != nil {
			log.Error("insert parsed_messages", err)
			continue
			// return err
		}
		_, err = store.PqDB.Exec(`INSERT INTO transactions SELECT * FROM `+
			`tmp_transactions WHERE height = $1 `+
			`ON CONFLICT DO NOTHING`, string(epoch))
		if err != nil {
			log.Error("insert transactions", err)
			continue
			// return err
		}

		// Delete inserted rows from tmp tables

		_, err = store.PqDB.Exec(`DELETE FROM tmp_parsed_messages `+
			`WHERE height = $1`, string(epoch))
		if err != nil {
			log.Error("delete tmp_parsed_messages", err)
			continue
			// return err
		}
		_, err = store.PqDB.Exec(`DELETE FROM tmp_transactions `+
			`WHERE height = $1`, string(epoch))
		if err != nil {
			log.Error("delete tmp_transactions", err)
			continue
			// return err
		}
	}

	return nil
}

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
