package actorcodes

import (
	"bufio"
	"context"
	"os"

	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log/v2"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

var log = logging.Logger("actorcodes")

type Task struct {
	node       lens.API
	opener     lens.APIOpener
	closer     lens.APICloser
	lastTipSet *types.TipSet
	store      db.Store
}

func NewTask(opener lens.APIOpener, store db.Store) *Task {
	return &Task{
		opener: opener,
		store:  store,
	}
}

func (p *Task) ProcessTipSet(ctx context.Context, ts *types.TipSet) (model.Persistable, error) {
	var pl model.PersistableList
	if p.node == nil {
		node, closer, err := p.opener.Open(ctx)
		if err != nil {
			return pl, xerrors.Errorf("unable to open lens: %w", err)
		}
		p.node = node
		p.closer = closer
	}

	headts, err := p.node.ChainHead(context.Background())
	if err != nil {
		log.Errorw("getting head tipset", "error", err)
		return nil, err
	}

	err = p.node.IndexActorCodes(context.Background(), headts)
	if err != nil {
		log.Errorw("indexactorcodes", "error", err)
	} else {
		f, err := os.OpenFile(os.Getenv("ACS_PARSEDTILL"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			log.Errorw("opening file ACS_PARSEDTILL", "error", err)
			return nil, err
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		_, err = w.WriteString(headts.Height().String())
		if err != nil {
			log.Errorw("writing to ACS_PARSEDTILL", "error", err)
			return nil, err
		}
		w.Flush()
		log.Infow("updated actorcodes", "height", headts.Height())
	}

	return pl, nil
}

func (p *Task) Close() error {
	return nil
}
