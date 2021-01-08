package services

import (
	"context"
	"log"

	"github.com/buidl-labs/filecoin-chain-indexer/chain"
	"github.com/buidl-labs/filecoin-chain-indexer/lens/lotus"

	// "github.com/filecoin-project/lotus/api/apistruct"
	"golang.org/x/xerrors"
)

func Walk() error {
	lensOpener, lensCloser, err := lotus.NewAPIOpener(context.Background(), 1)
	if err != nil {
		return xerrors.Errorf("setup lens: %w", err)
	}
	defer func() {
		lensCloser()
	}()

	tasks := []string{"minertask", "transactiontask"}
	tsIndexer, err := chain.NewTipSetIndexer(lensOpener, 0, "somename", tasks)
	if err != nil {
		return xerrors.Errorf("setup indexer: %w", err)
	}
	defer func() {
		if err := tsIndexer.Close(); err != nil {
			log.Println("failed to close tipset indexer cleanly", "error", err)
		}
	}()

	// walker := chain.NewWalker(&apistruct.FullNodeStruct{}, tsIndexer, 10, 1000)
	walker := chain.NewWalker(lensOpener, tsIndexer, 10, 1000)
	err = walker.Run(context.Background())
	if err != nil {
		return err
	}
	return nil
}
