package chain

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	builtin2 "github.com/filecoin-project/specs-actors/v2/actors/builtin"
	"github.com/ipfs/go-cid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	messagemodel "github.com/buidl-labs/filecoin-chain-indexer/model/messages"
)

type MessageProcessor struct {
	node       lens.API
	opener     lens.APIOpener
	closer     lens.APICloser
	lastTipSet *types.TipSet
	store      db.Store
}

func NewMessageProcessor(opener lens.APIOpener, store db.Store) *MessageProcessor {
	return &MessageProcessor{
		opener: opener,
		store:  store,
	}
}

func (p *MessageProcessor) ProcessTipSet(ctx context.Context, ts *types.TipSet) (model.Persistable, error) {
	if p.node == nil {
		node, closer, err := p.opener.Open(ctx)
		if err != nil {
			return nil, xerrors.Errorf("unable to open lens: %w", err)
		}
		p.node = node
		p.closer = closer
	}

	var data model.Persistable
	var err error
	// var txns []messagemodel.Transaction

	// fmt.Println("msgPT")
	if p.lastTipSet != nil {
		if p.lastTipSet.Height() > ts.Height() {
			// log.Info("p.lastTipSet.Height() > ts.Height()")
			// last tipset seen was the child
			data, _, err = p.processExecutedMessages(ctx, p.lastTipSet, ts)
		} else if p.lastTipSet.Height() < ts.Height() {
			// log.Info("p.lastTipSet.Height() < ts.Height()")
			// last tipset seen was the parent
			data, _, err = p.processExecutedMessages(ctx, ts, p.lastTipSet)
		} else {
			// log.Println("out of order tipsets", "height", ts.Height(), "last_height", p.lastTipSet.Height())
		}
	}

	// fmt.Println("Ptipset", data, txns)

	p.lastTipSet = ts

	if err != nil {
		log.Println("error received while processing messages, closing lens", "error", err)
		if cerr := p.Close(); cerr != nil {
			log.Println("error received while closing lens", "error", cerr)
		}
	}

	// log.Info("MTXNS", txns)
	// p.store.PersistTransactions(txns)

	return data, err
}

func (p *MessageProcessor) processExecutedMessages(ctx context.Context, ts, pts *types.TipSet) (model.Persistable, []messagemodel.Transaction, error) {
	emsgs, err := p.node.GetExecutedMessagesForTipset(ctx, ts, pts)
	// fmt.Println("HEREemsgs", emsgs)
	if err != nil {
		return nil, nil, err
	}

	var (
		// messageResults      = make(messagemodel.Messages, 0, len(emsgs))
		// receiptResults      = make(messagemodel.Receipts, 0, len(emsgs))
		// blockMessageResults = make(messagemodel.BlockMessages, 0, len(emsgs))
		// parsedMessageResults = make(messagemodel.ParsedMessages, 0, len(emsgs))
		transactionsResults = make([]messagemodel.Transaction, 0, 1) //len(emsgs))
		// errorsDetected      = make([]*MessageError, 0, len(emsgs))
	)

	var (
		seen = make(map[cid.Cid]bool, len(emsgs))

		totalGasLimit     int64
		totalUniqGasLimit int64
	)

	for _, m := range emsgs {
		// Stop processing if we have been told to cancel
		select {
		case <-ctx.Done():
			return nil, nil, xerrors.Errorf("context done: %w", ctx.Err())
		default:
		}

		// Record which blocks had which messages, regardless of duplicates
		for _, blockCid := range m.Blocks {
			fmt.Println(blockCid)
			// blockMessageResults = append(blockMessageResults, &messagemodel.BlockMessage{
			// 	Height:  int64(m.Height),
			// 	Block:   blockCid.String(),
			// 	Message: m.Cid.String(),
			// })
			totalGasLimit += m.Message.GasLimit
		}

		if seen[m.Cid] {
			continue
		}
		seen[m.Cid] = true
		totalUniqGasLimit += m.Message.GasLimit

		var msgSize int
		if b, err := m.Message.Serialize(); err == nil {
			msgSize = len(b)
		} else {
			// errorsDetected = append(errorsDetected, &MessageError{
			// 	Cid:   m.Cid,
			// 	Error: xerrors.Errorf("failed to serialize message: %w", err).Error(),
			// })
		}

		// record all unique messages
		msg := &messagemodel.Message{
			Height:     int64(m.Height),
			Cid:        m.Cid.String(),
			From:       m.Message.From.String(),
			To:         m.Message.To.String(),
			Value:      m.Message.Value.String(),
			GasFeeCap:  m.Message.GasFeeCap.String(),
			GasPremium: m.Message.GasPremium.String(),
			GasLimit:   m.Message.GasLimit,
			SizeBytes:  msgSize,
			Nonce:      m.Message.Nonce,
			Method:     uint64(m.Message.Method),
			MethodName: methodNames[m.Message.Method],
		}
		// messageResults = append(messageResults, msg)
		// log.Info("messageResults", messageResults)

		rcpt := &messagemodel.Receipt{
			Height:    int64(ts.Height()), // this is the child height
			Message:   msg.Cid,
			StateRoot: ts.ParentState().String(),
			Idx:       int(m.Index),
			ExitCode:  int64(m.Receipt.ExitCode),
			GasUsed:   m.Receipt.GasUsed,
		}
		// receiptResults = append(receiptResults, rcpt)

		// outputs := p.node.ComputeGasOutputs(m.Receipt.GasUsed, m.Message.GasLimit, m.BlockHeader.ParentBaseFee, m.Message.GasFeeCap, m.Message.GasPremium)
		transaction := &messagemodel.Transaction{
			Height:             msg.Height,
			Cid:                msg.Cid,
			Sender:             msg.From,
			Receiver:           msg.To,
			Amount:             msg.Value,
			Type:               0,
			GasFeeCap:          msg.GasFeeCap,
			GasPremium:         msg.GasPremium,
			GasLimit:           msg.GasLimit,
			Nonce:              msg.Nonce,
			Method:             msg.Method,
			MethodName:         msg.MethodName,
			StateRoot:          m.BlockHeader.ParentStateRoot.String(),
			ExitCode:           rcpt.ExitCode,
			GasUsed:            rcpt.GasUsed,
			ParentBaseFee:      m.BlockHeader.ParentBaseFee.String(),
			SizeBytes:          msgSize,
			BaseFeeBurn:        m.GasOutputs.BaseFeeBurn.String(),
			OverEstimationBurn: m.GasOutputs.OverEstimationBurn.String(),
			MinerPenalty:       m.GasOutputs.MinerPenalty.String(),
			MinerTip:           m.GasOutputs.MinerTip.String(),
			Refund:             m.GasOutputs.Refund.String(),
			GasRefund:          m.GasOutputs.GasRefund,
			GasBurned:          m.GasOutputs.GasBurned,
			ActorName:          builtin2.ActorNameByCode(m.ToActorCode),
		}
		_, err := p.store.DB.Model(transaction).Insert()
		if err != nil {
			log.Info("insert txn", err)
		} else {
			// log.Info("inserted txn", r)
		}
		transaction = nil
		// transactionsResults = append(transactionsResults, transaction)

		// method, params, err := p.parseMessageParams(m.Message, m.ToActorCode)
		// if err == nil {
		// 	pm := &messagemodel.ParsedMessage{
		// 		Height: msg.Height,
		// 		Cid:    msg.Cid,
		// 		From:   msg.From,
		// 		To:     msg.To,
		// 		Value:  msg.Value,
		// 		Method: method,
		// 		Params: params,
		// 	}
		// 	parsedMessageResults = append(parsedMessageResults, pm)
		// } else {
		// 	errorsDetected = append(errorsDetected, &MessageError{
		// 		Cid:   m.Cid,
		// 		Error: xerrors.Errorf("failed to parse message params: %w", err).Error(),
		// 	})
		// }
	}

	// newBaseFee := store.ComputeNextBaseFee(pts.Blocks()[0].ParentBaseFee, totalUniqGasLimit, len(pts.Blocks()), pts.Height())
	// baseFeeRat := new(big.Rat).SetFrac(newBaseFee.Int, new(big.Int).SetUint64(build.FilecoinPrecision))
	// baseFee, _ := baseFeeRat.Float64()

	// baseFeeChange := new(big.Rat).SetFrac(newBaseFee.Int, pts.Blocks()[0].ParentBaseFee.Int)
	// baseFeeChangeF, _ := baseFeeChange.Float64()

	// messageGasEconomyResult := &messagemodel.MessageGasEconomy{
	// 	Height:              int64(pts.Height()),
	// 	StateRoot:           pts.ParentState().String(),
	// 	GasLimitTotal:       totalGasLimit,
	// 	GasLimitUniqueTotal: totalUniqGasLimit,
	// 	BaseFee:             baseFee,
	// 	BaseFeeChangeLog:    math.Log(baseFeeChangeF) / math.Log(1.125),
	// 	GasFillRatio:        float64(totalGasLimit) / float64(len(pts.Blocks())*build.BlockGasTarget),
	// 	GasCapacityRatio:    float64(totalUniqGasLimit) / float64(len(pts.Blocks())*build.BlockGasTarget),
	// 	GasWasteRatio:       float64(totalGasLimit-totalUniqGasLimit) / float64(len(pts.Blocks())*build.BlockGasTarget),
	// }

	// if len(errorsDetected) != 0 {
	// 	report.ErrorsDetected = errorsDetected
	// }

	return model.PersistableList{
		// messageResults,
		// receiptResults,
		// blockMessageResults,
		// parsedMessageResults,
		// transactionsResults,
		// messageGasEconomyResult,
	}, transactionsResults, nil
}

func (p *MessageProcessor) Close() error {
	if p.closer != nil {
		p.closer()
		p.closer = nil
	}
	p.node = nil
	return nil
}

type MessageError struct {
	Cid   cid.Cid
	Error string
}

var methodNames = map[abi.MethodNum]string{
	builtin2.MethodSend:                            "Send",
	builtin2.MethodsMiner.Constructor:              "Constructor",
	builtin2.MethodsMiner.ControlAddresses:         "ControlAddresses",
	builtin2.MethodsMiner.ChangeWorkerAddress:      "ChangeWorkerAddress",
	builtin2.MethodsMiner.ChangePeerID:             "ChangePeerID",
	builtin2.MethodsMiner.SubmitWindowedPoSt:       "SubmitWindowedPoSt",
	builtin2.MethodsMiner.PreCommitSector:          "PreCommitSector",
	builtin2.MethodsMiner.ProveCommitSector:        "ProveCommitSector",
	builtin2.MethodsMiner.ExtendSectorExpiration:   "ExtendSectorExpiration",
	builtin2.MethodsMiner.TerminateSectors:         "TerminateSectors",
	builtin2.MethodsMiner.DeclareFaults:            "DeclareFaults",
	builtin2.MethodsMiner.DeclareFaultsRecovered:   "DeclareFaultsRecovered",
	builtin2.MethodsMiner.OnDeferredCronEvent:      "OnDeferredCronEvent",
	builtin2.MethodsMiner.CheckSectorProven:        "CheckSectorProven",
	builtin2.MethodsMiner.ApplyRewards:             "ApplyRewards",
	builtin2.MethodsMiner.ReportConsensusFault:     "ReportConsensusFault",
	builtin2.MethodsMiner.WithdrawBalance:          "WithdrawBalance",
	builtin2.MethodsMiner.ConfirmSectorProofsValid: "ConfirmSectorProofsValid",
	builtin2.MethodsMiner.ChangeMultiaddrs:         "ChangeMultiaddrs",
	builtin2.MethodsMiner.CompactPartitions:        "CompactPartitions",
	builtin2.MethodsMiner.CompactSectorNumbers:     "CompactSectorNumbers",
	builtin2.MethodsMiner.ConfirmUpdateWorkerKey:   "ConfirmUpdateWorkerKey",
	builtin2.MethodsMiner.RepayDebt:                "RepayDebt",
	builtin2.MethodsMiner.ChangeOwnerAddress:       "ChangeOwnerAddress",
}
