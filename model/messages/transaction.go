package messages

import (
	"context"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

type TransactionType int

const (
	StorageDeal TransactionType = iota
	BlockReward
	Penalty
	NetworkFee
	Other
)

type Transaction struct {
	tableName          struct{}        `pg:"transactions"` // nolint: structcheck,unused
	Height             int64           `pg:",pk,use_zero,notnull"`
	Cid                string          `pg:",pk,notnull"`
	StateRoot          string          `pg:",pk,notnull"`
	Sender             string          `pg:",notnull"` // miner/owner/worker/from account
	Receiver           string          `pg:",notnull"` // miner/owner/worker/to account
	Amount             string          `pg:",notnull"` // amount (+/- FIL)
	Type               TransactionType `pg:""`
	GasFeeCap          string          `pg:",notnull"`
	GasPremium         string          `pg:",notnull"`
	GasLimit           int64           `pg:",use_zero,notnull"`
	SizeBytes          int             `pg:",use_zero,notnull"`
	Nonce              uint64          `pg:",use_zero,notnull"`
	Method             uint64          `pg:",use_zero,notnull"`
	MethodName         string          `pg:",notnull"`
	ParamsBytes        []byte          `pg:""`
	Params             string          `pg:""`
	Transferred        string          `pg:""`
	ActorName          string          `pg:",notnull"`
	ExitCode           int64           `pg:",use_zero,notnull"`
	GasUsed            int64           `pg:",use_zero,notnull"`
	ParentBaseFee      string          `pg:",notnull"`
	BaseFeeBurn        string          `pg:",notnull"`
	OverEstimationBurn string          `pg:",notnull"`
	MinerPenalty       string          `pg:",notnull"`
	MinerTip           string          `pg:",notnull"`
	Refund             string          `pg:",notnull"`
	GasRefund          int64           `pg:",use_zero,notnull"`
	GasBurned          int64           `pg:",use_zero,notnull"`
	Miner              string          `pg:""`
	ReturnBytes        string          `pg:""`
}

type MinerTransaction struct {
	tableName     struct{} `pg:"miner_transactions"` // nolint: structcheck,unused
	Height        int64    `pg:",pk,use_zero,notnull"`
	Cid           string   `pg:",pk,notnull"`
	Sender        string   `pg:""` // miner/owner/worker/from account
	Receiver      string   `pg:""` // miner/owner/worker/to account
	Amount        string   `pg:""` // amount (+/- FIL)
	GasFeeCap     string   `pg:""`
	GasPremium    string   `pg:""`
	GasLimit      int64    `pg:""`
	Nonce         uint64   `pg:""`
	Method        uint64   `pg:""`
	FromActorName string   `pg:""`
	ToActorName   string   `pg:""`
}

func (t *Transaction) Persist(ctx context.Context, s model.StorageBatch) error {
	return s.PersistModel(ctx, t)
}

type Transactions []*Transaction

func (ts Transactions) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(ts) == 0 {
		return nil
	}

	return s.PersistModel(ctx, ts)
}
