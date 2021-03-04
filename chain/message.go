package chain

import (
	"context"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	// builtin2 "github.com/filecoin-project/specs-actors/v2/actors/builtin"
	builtin3 "github.com/filecoin-project/specs-actors/v3/actors/builtin"

	// "github.com/filecoin-project/specs-actors/v3/actors/builtin"
	"github.com/ipfs/go-cid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/lens/lotus"
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
			Height:      int64(m.Height),
			Cid:         m.Cid.String(),
			From:        m.Message.From.String(),
			To:          m.Message.To.String(),
			Value:       m.Message.Value.String(),
			GasFeeCap:   m.Message.GasFeeCap.String(),
			GasPremium:  m.Message.GasPremium.String(),
			GasLimit:    m.Message.GasLimit,
			SizeBytes:   msgSize,
			Nonce:       m.Message.Nonce,
			Method:      uint64(m.Message.Method),
			ParamsBytes: m.Message.Params,
			Params:      generateParamsStr(m.Message.To, m.Message.Params, m.Message.Method, p),
			MethodName:  getMethodName(lotus.ActorNameByCode(m.ToActorCode), m.Message.Method),
		}
		// fmt.Println("PSTRS", (m.Message.Params))
		// log.Info(msg.Cid, " version: ", m.ToActorCode.Version())
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
			ParamsBytes:        msg.ParamsBytes,
			Params:             msg.Params,
			Transferred:        computeTransferredAmount(m, lotus.ActorNameByCode(m.ToActorCode), p),
			ActorName:          lotus.ActorNameByCode(m.ToActorCode),
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

func generateParamsStr(toAddr address.Address, params []byte, method abi.MethodNum, p *MessageProcessor) string {
	decodedParams, err := p.node.StateDecodeParams(context.Background(), toAddr, method, params, types.EmptyTSK)
	if err != nil {
		return ""
	}
	// fmt.Println("DPSTR", decodedParams)
	decodedParamsStr := fmt.Sprintf("%v", decodedParams)
	return decodedParamsStr
}

func computeTransferredAmount(m *lens.ExecutedMessage, actorName string, p *MessageProcessor) string {
	if m.Message.Method == builtin3.MethodsMiner.WithdrawBalance &&
		actorName == "fil/3/storageminer" {
		decodedParams, err := p.node.StateDecodeParams(context.Background(), m.Message.To, m.Message.Method, m.Message.Params, types.EmptyTSK)
		if err != nil {
			return "0"
		}
		d := decodedParams.(map[string]interface{})
		amt := d["AmountRequested"].(string)
		return amt
	} else if m.Message.Method == builtin3.MethodsMarket.WithdrawBalance &&
		actorName == "fil/3/storagemarket" {
		decodedParams, err := p.node.StateDecodeParams(context.Background(), m.Message.To, m.Message.Method, m.Message.Params, types.EmptyTSK)
		if err != nil {
			return "0"
		}
		d := decodedParams.(map[string]interface{})
		amt := d["Amount"].(string)
		return amt
	}
	return "0"
}

func getMethodName(actorName string, methodNum abi.MethodNum) string {
	log.Info("act:", actorName, "meth:", methodNum)
	if methodNum == 0 {
		return commonMethods[0]
	} else if methodNum == 1 {
		return commonMethods[1]
	}
	switch actorName {
	case "fil/3/account":
		return accountMethods[methodNum]
	case "fil/3/init":
		return initMethods[methodNum]
	case "fil/3/cron":
		return cronMethods[methodNum]
	case "fil/3/reward":
		return rewardMethods[methodNum]
	case "fil/3/multisig":
		return multisigMethods[methodNum]
	case "fil/3/paymentchannel":
		return paychMethods[methodNum]
	case "fil/3/storagemarket":
		return marketMethods[methodNum]
	case "fil/3/storagepower":
		return powerMethods[methodNum]
	case "fil/3/storageminer":
		return minerMethods[methodNum]
	case "fil/3/verifiedregistry":
		return verifiedRegistryMethods[methodNum]
	default:
		return ""
	}
}

var commonMethods = map[abi.MethodNum]string{
	builtin3.MethodSend:        "Send",
	builtin3.MethodConstructor: "Constructor",
}

var accountMethods = map[abi.MethodNum]string{
	builtin3.MethodsAccount.Constructor:   "Constructor",
	builtin3.MethodsAccount.PubkeyAddress: "PubkeyAddress",
}

var initMethods = map[abi.MethodNum]string{
	builtin3.MethodsInit.Constructor: "Constructor",
	builtin3.MethodsInit.Exec:        "Exec",
}

var cronMethods = map[abi.MethodNum]string{
	builtin3.MethodsCron.Constructor: "Constructor",
	builtin3.MethodsCron.EpochTick:   "EpochTick",
}

var rewardMethods = map[abi.MethodNum]string{
	builtin3.MethodsReward.Constructor:      "Constructor",
	builtin3.MethodsReward.AwardBlockReward: "AwardBlockReward",
	builtin3.MethodsReward.ThisEpochReward:  "ThisEpochReward",
	builtin3.MethodsReward.UpdateNetworkKPI: "UpdateNetworkKPI",
}

var multisigMethods = map[abi.MethodNum]string{
	builtin3.MethodsMultisig.Constructor:                 "Constructor",
	builtin3.MethodsMultisig.Propose:                     "Propose",
	builtin3.MethodsMultisig.Approve:                     "Approve",
	builtin3.MethodsMultisig.Cancel:                      "Cancel",
	builtin3.MethodsMultisig.AddSigner:                   "AddSigner",
	builtin3.MethodsMultisig.RemoveSigner:                "RemoveSigner",
	builtin3.MethodsMultisig.SwapSigner:                  "SwapSigner",
	builtin3.MethodsMultisig.ChangeNumApprovalsThreshold: "ChangeNumApprovalsThreshold",
	builtin3.MethodsMultisig.LockBalance:                 "LockBalance",
}

var paychMethods = map[abi.MethodNum]string{
	builtin3.MethodsPaych.Constructor:        "Constructor",
	builtin3.MethodsPaych.UpdateChannelState: "UpdateChannelState",
	builtin3.MethodsPaych.Settle:             "Settle",
	builtin3.MethodsPaych.Collect:            "Collect",
}

var marketMethods = map[abi.MethodNum]string{
	builtin3.MethodsMarket.Constructor:              "Constructor",
	builtin3.MethodsMarket.AddBalance:               "AddBalance",
	builtin3.MethodsMarket.WithdrawBalance:          "WithdrawBalance",
	builtin3.MethodsMarket.PublishStorageDeals:      "PublishStorageDeals",
	builtin3.MethodsMarket.VerifyDealsForActivation: "VerifyDealsForActivation",
	builtin3.MethodsMarket.ActivateDeals:            "ActivateDeals",
	builtin3.MethodsMarket.OnMinerSectorsTerminate:  "OnMinerSectorsTerminate",
	builtin3.MethodsMarket.ComputeDataCommitment:    "ComputeDataCommitment",
	builtin3.MethodsMarket.CronTick:                 "CronTick",
}

var powerMethods = map[abi.MethodNum]string{
	builtin3.MethodsPower.Constructor:              "Constructor",
	builtin3.MethodsPower.CreateMiner:              "CreateMiner",
	builtin3.MethodsPower.UpdateClaimedPower:       "UpdateClaimedPower",
	builtin3.MethodsPower.EnrollCronEvent:          "EnrollCronEvent",
	builtin3.MethodsPower.OnEpochTickEnd:           "OnEpochTickEnd",
	builtin3.MethodsPower.UpdatePledgeTotal:        "UpdatePledgeTotal",
	builtin3.MethodsPower.Deprecated1:              "Deprecated1",
	builtin3.MethodsPower.SubmitPoRepForBulkVerify: "SubmitPoRepForBulkVerify",
	builtin3.MethodsPower.CurrentTotalPower:        "CurrentTotalPower",
}

var minerMethods = map[abi.MethodNum]string{
	builtin3.MethodsMiner.Constructor:              "Constructor",
	builtin3.MethodsMiner.ControlAddresses:         "ControlAddresses",
	builtin3.MethodsMiner.ChangeWorkerAddress:      "ChangeWorkerAddress",
	builtin3.MethodsMiner.ChangePeerID:             "ChangePeerID",
	builtin3.MethodsMiner.SubmitWindowedPoSt:       "SubmitWindowedPoSt",
	builtin3.MethodsMiner.PreCommitSector:          "PreCommitSector",
	builtin3.MethodsMiner.ProveCommitSector:        "ProveCommitSector",
	builtin3.MethodsMiner.ExtendSectorExpiration:   "ExtendSectorExpiration",
	builtin3.MethodsMiner.TerminateSectors:         "TerminateSectors",
	builtin3.MethodsMiner.DeclareFaults:            "DeclareFaults",
	builtin3.MethodsMiner.DeclareFaultsRecovered:   "DeclareFaultsRecovered",
	builtin3.MethodsMiner.OnDeferredCronEvent:      "OnDeferredCronEvent",
	builtin3.MethodsMiner.CheckSectorProven:        "CheckSectorProven",
	builtin3.MethodsMiner.ApplyRewards:             "ApplyRewards",
	builtin3.MethodsMiner.ReportConsensusFault:     "ReportConsensusFault",
	builtin3.MethodsMiner.WithdrawBalance:          "WithdrawBalance",
	builtin3.MethodsMiner.ConfirmSectorProofsValid: "ConfirmSectorProofsValid",
	builtin3.MethodsMiner.ChangeMultiaddrs:         "ChangeMultiaddrs",
	builtin3.MethodsMiner.CompactPartitions:        "CompactPartitions",
	builtin3.MethodsMiner.CompactSectorNumbers:     "CompactSectorNumbers",
	builtin3.MethodsMiner.ConfirmUpdateWorkerKey:   "ConfirmUpdateWorkerKey",
	builtin3.MethodsMiner.RepayDebt:                "RepayDebt",
	builtin3.MethodsMiner.ChangeOwnerAddress:       "ChangeOwnerAddress",
	builtin3.MethodsMiner.DisputeWindowedPoSt:      "DisputeWindowedPoSt",
}

var verifiedRegistryMethods = map[abi.MethodNum]string{
	builtin3.MethodsVerifiedRegistry.Constructor:       "Constructor",
	builtin3.MethodsVerifiedRegistry.AddVerifier:       "AddVerifier",
	builtin3.MethodsVerifiedRegistry.RemoveVerifier:    "RemoveVerifier",
	builtin3.MethodsVerifiedRegistry.AddVerifiedClient: "AddVerifiedClient",
	builtin3.MethodsVerifiedRegistry.UseBytes:          "UseBytes",
	builtin3.MethodsVerifiedRegistry.RestoreBytes:      "RestoreBytes",
}
