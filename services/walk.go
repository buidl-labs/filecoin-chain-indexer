package services

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/chain"
	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens/lotus"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	"github.com/buidl-labs/filecoin-chain-indexer/storage"
)

func Walk(cfg config.Config, tasks []string, taskType int) error {
	lensOpener, lensCloser, err := lotus.NewAPIOpener(cfg, context.Background())
	if err != nil {
		return xerrors.Errorf("setup lens: %w", err)
	}
	fmt.Println("deferlensclose")
	defer func() {
		lensCloser()
	}()

	var strg model.Storage = &storage.NullStorage{}

	store, err := db.New(cfg.DBConnStr)
	if err != nil {
		return xerrors.Errorf("setup indexer, connecting db: %w", err)
	}
	db0, _ := store.Conn()
	fmt.Println("gonna open TSIDXR")
	tsIndexer, err := chain.NewTipSetIndexer(lensOpener, db0, *store, strg, 0, "somename", tasks)
	if err != nil {
		return xerrors.Errorf("setup indexer: %w", err)
	}
	defer func() {
		if err := tsIndexer.Close(); err != nil {
			log.Error("failed to close tipset indexer cleanly", err)
		}
	}()

	ctx := context.Background()
	node, closer, err := lensOpener.Open(ctx)
	if err != nil {
		return xerrors.Errorf("open lens: %w", err)
	}
	defer closer()

	ts, err := node.ChainHead(ctx)
	if err != nil {
		return xerrors.Errorf("get chain head: %w", err)
	}

	maxHeight := int64(ts.Height()) - 100
	// maxHeight = int64(1000)
	log.Info("maxHeight", maxHeight)

	// TODO: start indexing from a certain height
	// height := dataservice.GetParsedTill()
	// minHeight := int64(192698) // setting a dummy value here
	minHeight := maxHeight - 2

	// walker := chain.NewWalker(&apistruct.FullNodeStruct{}, tsIndexer, 10, 1000)
	walker := chain.NewWalker(lensOpener, tsIndexer, tasks, taskType, minHeight, maxHeight)
	// walker := chain.NewWalker(lensOpener, tsIndexer, 10195, 10200)
	err = walker.Run(context.Background())
	if err != nil {
		return err
	}
	return nil
}
