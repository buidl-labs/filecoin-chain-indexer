package services

import (
	"context"

	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/chain"
	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens/lotus"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	"github.com/buidl-labs/filecoin-chain-indexer/storage"
)

func Walk(cfg config.Config, tasks []string, taskType chain.TaskType, name string) error {
	lensOpener, lensCloser, err := lotus.NewAPIOpener(cfg, context.Background())
	if err != nil {
		log.Error("setup lens",
			"tasks", tasks,
			"taskType", taskType,
			"error", err,
		)
		return xerrors.Errorf("setup lens: %w", err)
	}
	defer func() {
		lensCloser()
	}()

	var strg model.Storage = &storage.NullStorage{}

	store, err := db.New(cfg.DBConnStr)
	if err != nil {
		log.Errorw("setup indexer, connecting db", "error", err)
		return xerrors.Errorf("setup indexer, connecting db: %w", err)
	}
	db0, _ := store.Conn()
	tsIndexer, err := chain.NewTipSetIndexer(lensOpener, db0, *store, strg, 0, name, tasks, cfg)
	if err != nil {
		log.Errorw("creating tipset indexer",
			"error", err,
			"tasks", tasks,
		)
		return xerrors.Errorf("setup indexer: %w", err)
	}
	defer func() {
		if err := tsIndexer.Close(); err != nil {
			log.Errorw("failed to close tipset indexer cleanly", "error", err)
		}
	}()

	ctx := context.Background()
	node, closer, err := lensOpener.Open(ctx)
	if err != nil {
		log.Errorw("open lens", "error", err)
		return xerrors.Errorf("open lens: %w", err)
	}
	defer closer()

	ts, err := node.ChainHead(ctx)
	if err != nil {
		log.Errorw("get chain head", "error", err)
		return xerrors.Errorf("get chain head: %w", err)
	}

	minHeight := cfg.From
	maxHeight := cfg.To

	if minHeight > maxHeight {
		log.Fatal("FROM should be less than TO")
	} else if minHeight >= int64(ts.Height())-900 {
		log.Fatal("FROM should be less than chainHead-finality")
	} else if maxHeight >= int64(ts.Height())-900 {
		log.Fatal("TO should be less than chainHead-finality")
	}

	log.Infow("creating walker",
		"minHeight", minHeight,
		"maxHeight", maxHeight,
	)

	walker := chain.NewWalker(lensOpener, tsIndexer, tasks, taskType, minHeight, maxHeight, cfg)
	err = walker.Run(context.Background())
	if err != nil {
		log.Errorw("running walker",
			"error", err,
			"tasks", tasks,
		)
		return err
	}
	return nil
}
