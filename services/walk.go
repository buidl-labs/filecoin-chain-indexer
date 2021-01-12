package services

import (
	"context"
	"log"

	"github.com/buidl-labs/filecoin-chain-indexer/chain"
	"github.com/buidl-labs/filecoin-chain-indexer/lens/lotus"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	"github.com/buidl-labs/filecoin-chain-indexer/storage"

	// "github.com/filecoin-project/lotus/api/apistruct"
	"golang.org/x/xerrors"
)

func Walk(tasks []string, taskType int) error {
	lensOpener, lensCloser, err := lotus.NewAPIOpener(context.Background(), 1)
	if err != nil {
		return xerrors.Errorf("setup lens: %w", err)
	}
	defer func() {
		lensCloser()
	}()

	var strg model.Storage = &storage.NullStorage{}
	// tasks := []string{"miners", "markets", "messages"}
	// tasks := []string{"markets"}
	tsIndexer, err := chain.NewTipSetIndexer(lensOpener, strg, 0, "somename", tasks)
	if err != nil {
		return xerrors.Errorf("setup indexer: %w", err)
	}
	defer func() {
		if err := tsIndexer.Close(); err != nil {
			log.Println("failed to close tipset indexer cleanly", "error", err)
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

	maxHeight := int64(ts.Height())

	// TODO: implement dataservice
	// height := dataservice.GetParsedTill()
	// minHeight := int64(192698) // setting a dummy value here
	minHeight := maxHeight - 10

	// walker := chain.NewWalker(&apistruct.FullNodeStruct{}, tsIndexer, 10, 1000)
	walker := chain.NewWalker(lensOpener, tsIndexer, tasks, taskType, minHeight, maxHeight)
	// walker := chain.NewWalker(lensOpener, tsIndexer, 10195, 10200)
	err = walker.Run(context.Background())
	if err != nil {
		return err
	}
	return nil
}
