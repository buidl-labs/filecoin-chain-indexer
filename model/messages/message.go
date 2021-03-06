package messages

import (
	"context"

	"github.com/buidl-labs/filecoin-chain-indexer/model"
)

type Message struct {
	tableName struct{} `pg:"messages"` // nolint: structcheck,unused
	Height    int64    `pg:",pk,notnull,use_zero"`
	Cid       string   `pg:",pk,notnull"`

	From       string `pg:",notnull"`
	To         string `pg:",notnull"`
	Value      string `pg:",notnull"`
	GasFeeCap  string `pg:",notnull"`
	GasPremium string `pg:",notnull"`

	GasLimit    int64  `pg:",use_zero"`
	SizeBytes   int    `pg:",use_zero"`
	Nonce       uint64 `pg:",use_zero"`
	Method      uint64 `pg:",use_zero"`
	MethodName  string `pg:",notnull"`
	ParamsBytes []byte `pg:""`
	Params      string `pg:""`
}

func (m *Message) Persist(ctx context.Context, s model.StorageBatch) error {
	return s.PersistModel(ctx, m)
}

type Messages []*Message

func (ms Messages) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(ms) == 0 {
		return nil
	}

	return s.PersistModel(ctx, ms)
}

type GasOutputs struct {
	tableName          struct{} `pg:"derived_gas_outputs"` //nolint: structcheck,unused
	Height             int64    `pg:",pk,use_zero,notnull"`
	Cid                string   `pg:",pk,notnull"`
	StateRoot          string   `pg:",pk,notnull"`
	From               string   `pg:",notnull"`
	To                 string   `pg:",notnull"`
	Value              string   `pg:",notnull"`
	GasFeeCap          string   `pg:",notnull"`
	GasPremium         string   `pg:",notnull"`
	GasLimit           int64    `pg:",use_zero,notnull"`
	SizeBytes          int      `pg:",use_zero,notnull"`
	Nonce              uint64   `pg:",use_zero,notnull"`
	Method             uint64   `pg:",use_zero,notnull"`
	MethodName         string   `pg:",notnull"`
	ActorName          string   `pg:",notnull"`
	ExitCode           int64    `pg:",use_zero,notnull"`
	GasUsed            int64    `pg:",use_zero,notnull"`
	ParentBaseFee      string   `pg:",notnull"`
	BaseFeeBurn        string   `pg:",notnull"`
	OverEstimationBurn string   `pg:",notnull"`
	MinerPenalty       string   `pg:",notnull"`
	MinerTip           string   `pg:",notnull"`
	Refund             string   `pg:",notnull"`
	GasRefund          int64    `pg:",use_zero,notnull"`
	GasBurned          int64    `pg:",use_zero,notnull"`
}

func (g *GasOutputs) Persist(ctx context.Context, s model.StorageBatch) error {
	return s.PersistModel(ctx, g)
}

type GasOutputsList []*GasOutputs

func (l GasOutputsList) Persist(ctx context.Context, s model.StorageBatch) error {
	if len(l) == 0 {
		return nil
	}

	return s.PersistModel(ctx, l)
}
