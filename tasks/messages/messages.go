package messages

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"os"

	// "github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/statediff"
	"github.com/filecoin-project/statediff/codec/fcjson"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-ipld-prime"
	"github.com/streadway/amqp"
	"golang.org/x/xerrors"

	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	// "github.com/buidl-labs/filecoin-chain-indexer/lens/lotus"
	"github.com/buidl-labs/filecoin-chain-indexer/model"
	messagemodel "github.com/buidl-labs/filecoin-chain-indexer/model/messages"
)

var log = logging.Logger("messages")

type Task struct {
	node       lens.API
	opener     lens.APIOpener
	closer     lens.APICloser
	lastTipSet *types.TipSet
	store      db.Store
	cfg        config.Config
}

func NewTask(opener lens.APIOpener, store db.Store) *Task {
	return &Task{
		opener: opener,
		store:  store,
	}
}

func failOnError(err error, msg string) error {
	if err != nil {
		log.Errorf("%s: %s", msg, err)
		return xerrors.Errorf("%s: %w", msg, err)
	}
	return nil
}

func (p *Task) ProcessMessages(ctx context.Context, ts *types.TipSet, pts *types.TipSet, emsgs []*lens.ExecutedMessage) (model.Persistable, error) {
	if p.node == nil {
		node, closer, err := p.opener.Open(ctx)
		if err != nil {
			return nil, xerrors.Errorf("unable to open lens: %w", err)
		}
		p.node = node
		p.closer = closer
	}

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Errorf("Failed to connect to RabbitMQ: %s", err)
		return nil, xerrors.Errorf("Failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("Failed to open a channel: %s", err)
		return nil, xerrors.Errorf("Failed to open a channel: %w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"A_rawcsojson", // name
		false,          // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		log.Errorf("Failed to declare a queue: %s", err)
		return nil, xerrors.Errorf("Failed to declare a queue: %w", err)
	}

	projectRoot := os.Getenv("ROOTDIR")

	log.Debugw("in ProcessMessages",
		"height", pts.Height(),
		"ph", pts.Height(),
	)
	cso, errx := p.node.StateCompute(context.Background(), pts.Height(), nil, pts.Key())
	log.Debug("completed StateCompute call")
	if errx != nil {
		log.Errorw("calling StateCompute",
			"error", errx,
			"tipset", pts.Height(),
		)
		// body := "failed fetching raw cso " + pts.Height().String()
		// err = ch.Publish(
		// 	"",     // exchange
		// 	q.Name, // routing key
		// 	false,  // mandatory
		// 	false,  // immediate
		// 	amqp.Publishing{
		// 		ContentType: "text/plain",
		// 		Body:        []byte(body),
		// 	})
		// if err != nil {
		// 	log.Errorf("Failed to publish a message: %s", err)
		// 	return nil, xerrors.Errorf("Failed to publish a message: %w", err)
		// }
	} else {
		csoFile, _ := json.MarshalIndent(cso, "", "	")
		err := ioutil.WriteFile(projectRoot+"/s3data/cso/"+pts.Height().String()+".json", csoFile, 0644)
		if err != nil {
			log.Errorw("write csofile",
				"error", err,
				"height", pts.Height(),
			)
			// body := "failed writing raw cso " + pts.Height().String()
			// err = ch.Publish(
			// 	"",     // exchange
			// 	q.Name, // routing key
			// 	false,  // mandatory
			// 	false,  // immediate
			// 	amqp.Publishing{
			// 		ContentType: "text/plain",
			// 		Body:        []byte(body),
			// 	})
			// if err != nil {
			// 	log.Errorf("Failed to publish a message: %s", err)
			// 	return nil, xerrors.Errorf("Failed to publish a message: %w", err)
			// }
		}
		log.Debugw("written csofile",
			"height", pts.Height(),
		)
		body := "successfully written raw cso " + pts.Height().String()
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		if err != nil {
			log.Errorf("Failed to publish a message: %s", err)
			return nil, xerrors.Errorf("Failed to publish a message: %w", err)
		}

		/*
			msgCount := 0
			subcallsCount := 0

			var execTraceResults messagemodel.Transactions
			var execTraceSubcallsResults messagemodel.Transactions

			for i, ir := range cso.Trace {
				log.Info("ir " + fmt.Sprintf("%d", i))

				log.Debug("ir StateLookupID start")
				var irToID, irFromID address.Address
				irToID, err := p.node.StateLookupID(context.Background(), ir.Msg.To, types.EmptyTSK)
				if err != nil {
					log.Warnw("couldn't lookup receiver's idaddress",
						"receiver", ir.Msg.To,
						"error", err,
					)
					irToID = ir.Msg.To
				}
				irFromID, err = p.node.StateLookupID(context.Background(), ir.Msg.From, types.EmptyTSK)
				if err != nil {
					log.Warnw("couldn't lookup sender's idaddress",
						"sender", ir.Msg.From,
						"error", err,
					)
					irFromID = ir.Msg.From
				}
				log.Debug("ir StateLookupID end")

				// log.Infow("ir "+fmt.Sprintf("%d", i)+": ",
				// 	"Cid", ir.MsgCid,
				// 	"From", irFromID,
				// 	"To", irToID,
				// 	"Method", ir.Msg.Method,
				// 	"Value", ir.Msg.Value,
				// 	"Nonce", ir.Msg.Nonce,
				// 	// "Params", generateParamsStr(ir.Msg.To, ir.Msg.Params, ir.Msg.Method, p),
				// 	"ExitCode", ir.MsgRct.ExitCode,
				// 	"Return", ir.MsgRct.Return,
				// 	"GasUsed", ir.MsgRct.GasUsed,

				// 	"GasFeeCap", ir.Msg.GasFeeCap,
				// 	"GasPremium", ir.Msg.GasPremium,
				// 	"GasLimit", ir.Msg.GasLimit,

				// 	"BaseFeeBurn", ir.GasCost.BaseFeeBurn,
				// 	"OverEstimationBurn", ir.GasCost.OverEstimationBurn,
				// 	"MinerPenalty", ir.GasCost.MinerPenalty,
				// 	"MinerTip", ir.GasCost.MinerTip,
				// 	"Refund", ir.GasCost.Refund,
				// )
				transaction := &messagemodel.Transaction{
					Height:     int64(pts.Height()),
					Cid:        ir.MsgCid.String(),
					Sender:     irFromID.String(),
					Receiver:   irToID.String(),
					Amount:     ir.Msg.Value.String(),
					Type:       0,
					GasFeeCap:  ir.Msg.GasFeeCap.String(),
					GasPremium: ir.Msg.GasPremium.String(),
					GasLimit:   ir.Msg.GasLimit,
					Nonce:      ir.Msg.Nonce,
					Method:     uint64(ir.Msg.Method),
					// MethodName:         msg.MethodName,
					StateRoot: cso.Root.String(),
					ExitCode:  int64(ir.MsgRct.ExitCode),
					GasUsed:   ir.MsgRct.GasUsed,
					// ParentBaseFee:      m.BlockHeader.ParentBaseFee.String(),
					// SizeBytes:          msgSize,
					BaseFeeBurn:        ir.GasCost.BaseFeeBurn.String(),
					OverEstimationBurn: ir.GasCost.OverEstimationBurn.String(),
					MinerPenalty:       ir.GasCost.MinerPenalty.String(),
					MinerTip:           ir.GasCost.MinerTip.String(),
					Refund:             ir.GasCost.Refund.String(),
					// GasRefund:          m.GasOutputs.GasRefund,
					// GasBurned:          m.GasOutputs.GasBurned,
					ParamsBytes: []byte{}, //keep empty for now
					Params:      "0",      //keep empty for now
					// Transferred:        computeTransferredAmount(m, lotus.ActorNameByCode(m.ToActorCode), p),
					// ActorName:          lotus.ActorNameByCode(m.ToActorCode),
				}
				execTraceResults = append(execTraceResults, transaction)

				// f02, _ := address.NewFromString("f02")
				// if ir.Msg.From == f02 && ir.Msg.Method == 14 {
				// 	// MethodsMiner.ApplyRewards
				// }
				msgCount += 1
				// log.Infow("et "+fmt.Sprintf("%d", i)+": ",
				// 	"Cid", ir.ExecutionTrace.Msg.Cid(),
				// 	"From", ir.ExecutionTrace.Msg.From,
				// 	"To", ir.ExecutionTrace.Msg.To,
				// 	"Method", ir.ExecutionTrace.Msg.Method,
				// 	"Params", generateParamsStr(
				// 		ir.ExecutionTrace.Msg.To,
				// 		ir.ExecutionTrace.Msg.Params,
				// 		ir.ExecutionTrace.Msg.Method,
				// 		p,
				// 	),
				// 	"ExitCode", ir.ExecutionTrace.MsgRct.ExitCode,
				// )
				for j, sc := range ir.ExecutionTrace.Subcalls {
					log.Info("sc " + fmt.Sprintf("%d.%d", i, j))

					log.Debug("sc StateLookupID start")
					var scToID, scFromID address.Address
					scToID, err := p.node.StateLookupID(context.Background(), sc.Msg.To, types.EmptyTSK)
					if err != nil {
						log.Warnw("couldn't lookup receiver's idaddress",
							"receiver", sc.Msg.To,
							"error", err,
						)
						scToID = sc.Msg.To
					}
					scFromID, err = p.node.StateLookupID(context.Background(), sc.Msg.From, types.EmptyTSK)
					if err != nil {
						log.Warnw("couldn't lookup sender's idaddress",
							"sender", sc.Msg.From,
							"error", err,
						)
						scFromID = sc.Msg.From
					}
					log.Debug("sc StateLookupID end")

					// log.Infow("sc "+fmt.Sprintf("%d.%d", i, j)+": ",
					// 	"Cid", sc.Msg.Cid(),
					// 	"From", scFromID,
					// 	"To", scToID,
					// 	"Method", sc.Msg.Method,
					// 	"Value", sc.Msg.Value,
					// 	"Nonce", sc.Msg.Nonce,
					// 	// "Params", generateParamsStr(sc.Msg.To, sc.Msg.Params, sc.Msg.Method, p),
					// 	"ExitCode", sc.MsgRct.ExitCode,
					// 	"Return", sc.MsgRct.Return,
					// 	"GasUsed", sc.MsgRct.GasUsed,

					// 	"GasFeeCap", sc.Msg.GasFeeCap,
					// 	"GasPremium", sc.Msg.GasPremium,
					// 	"GasLimit", sc.Msg.GasLimit,
					// )

					transaction := &messagemodel.Transaction{
						Height:     int64(pts.Height()),
						Cid:        sc.Msg.Cid().String(),
						Sender:     scFromID.String(),
						Receiver:   scToID.String(),
						Amount:     sc.Msg.Value.String(),
						Type:       0,
						GasFeeCap:  sc.Msg.GasFeeCap.String(),
						GasPremium: sc.Msg.GasPremium.String(),
						GasLimit:   sc.Msg.GasLimit,
						Nonce:      sc.Msg.Nonce,
						Method:     uint64(sc.Msg.Method),
						// MethodName:         msg.MethodName,
						StateRoot: cso.Root.String(),
						ExitCode:  int64(sc.MsgRct.ExitCode),
						GasUsed:   sc.MsgRct.GasUsed,
						// ParentBaseFee:      m.BlockHeader.ParentBaseFee.String(),
						// SizeBytes:          msgSize,
						// BaseFeeBurn:        sc.GasCost.BaseFeeBurn.String(),
						// OverEstimationBurn: sc.GasCost.OverEstimationBurn.String(),
						// MinerPenalty:       sc.GasCost.MinerPenalty.String(),
						// MinerTip:           sc.GasCost.MinerTip.String(),
						// Refund:             sc.GasCost.Refund.String(),
						// GasRefund:          m.GasOutputs.GasRefund,
						// GasBurned:          m.GasOutputs.GasBurned,
						ParamsBytes: []byte{}, //keep empty for now
						Params:      "0",      //keep empty for now
						// Transferred:        computeTransferredAmount(m, lotus.ActorNameByCode(m.ToActorCode), p),
						// ActorName:          lotus.ActorNameByCode(m.ToActorCode),
					}
					execTraceSubcallsResults = append(execTraceSubcallsResults, transaction)

					subcallsCount += 1
				}
			}
			log.Infow("counts",
				"msgCount", msgCount,
				"subcallsCount", subcallsCount,
			)

			etrsFile, _ := json.MarshalIndent(execTraceResults, "", "	")
			err = ioutil.WriteFile("./s3data/execTraceResults/"+pts.Height().String()+".json", etrsFile, 0644)
			if err != nil {
				log.Errorw("write execTraceResults",
					"error", err,
				)
			}

			etsrsFile, _ := json.MarshalIndent(execTraceSubcallsResults, "", "	")
			err = ioutil.WriteFile("./s3data/execTraceSubcallsResults/"+pts.Height().String()+".json", etsrsFile, 0644)
			if err != nil {
				log.Errorw("write execTraceSubcallsResults",
					"error", err,
				)
			}
		*/
	}

	var (
	// messageResults       = make(messagemodel.Messages, 0, len(emsgs))
	// receiptResults       = make(messagemodel.Receipts, 0, len(emsgs))
	// blockMessageResults = make(messagemodel.BlockMessages, 0, len(emsgs))
	// parsedMessageResults = make(messagemodel.ParsedMessages, 0, len(emsgs))
	// transactionsResults = make(messagemodel.Transactions, 0, len(emsgs))
	)

	var (
		// seen = make(map[cid.Cid]bool, len(emsgs))

		totalGasLimit     int64
		totalUniqGasLimit int64
	)
	/*
		for _, m := range emsgs {
			// Stop processing if we have been told to cancel
			select {
			case <-ctx.Done():
				return nil, xerrors.Errorf("context done: %w", ctx.Err())
			default:
			}

			// Record which blocks had which messages, regardless of duplicates
			for _, blockCid := range m.Blocks {
				blockMessageResults = append(blockMessageResults, &messagemodel.BlockMessage{
					Height:  int64(m.Height),
					Block:   blockCid.String(),
					Message: m.Cid.String(),
				})
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

			var toID, fromID address.Address

			toID, err := p.node.StateLookupID(context.Background(), m.Message.To, types.EmptyTSK)
			if err != nil {
				log.Warnw("couldn't lookup receiver's idaddress",
					"receiver", m.Message.To,
					"error", err,
				)
				toID = m.Message.To
			}
			fromID, err = p.node.StateLookupID(context.Background(), m.Message.From, types.EmptyTSK)
			if err != nil {
				log.Warnw("couldn't lookup sender's idaddress",
					"sender", m.Message.From,
					"error", err,
				)
				fromID = m.Message.From
			}

			// record all unique messages
			msg := &messagemodel.Message{
				Height:      int64(m.Height),
				Cid:         m.Cid.String(),
				From:        fromID.String(),
				To:          toID.String(),
				Value:       m.Message.Value.String(),
				GasFeeCap:   m.Message.GasFeeCap.String(),
				GasPremium:  m.Message.GasPremium.String(),
				GasLimit:    m.Message.GasLimit,
				SizeBytes:   msgSize,
				Nonce:       m.Message.Nonce,
				Method:      uint64(m.Message.Method),
				ParamsBytes: m.Message.Params,
				// Params:      generateParamsStr(m.Message.To, m.Message.Params, m.Message.Method, p),
				// MethodName:  getMethodName(lotus.ActorNameByCode(m.ToActorCode), m.Message.Method),
			}
			// messageResults = append(messageResults, msg)

			rcpt := &messagemodel.Receipt{
				Height:    int64(ts.Height()), // this is the child height
				Message:   msg.Cid,
				StateRoot: ts.ParentState().String(),
				Idx:       int(m.Index),
				ExitCode:  int64(m.Receipt.ExitCode),
				GasUsed:   m.Receipt.GasUsed,
			}
			// receiptResults = append(receiptResults, rcpt)

			transaction := &messagemodel.Transaction{
				Height:     msg.Height,
				Cid:        msg.Cid,
				Sender:     msg.From,
				Receiver:   msg.To,
				Amount:     msg.Value,
				Type:       0,
				GasFeeCap:  msg.GasFeeCap,
				GasPremium: msg.GasPremium,
				GasLimit:   msg.GasLimit,
				Nonce:      msg.Nonce,
				Method:     msg.Method,
				// MethodName:         msg.MethodName,
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
				ParamsBytes:        []byte{}, //keep empty for now
				Params:             "0",      //keep empty for now
				// Transferred:        computeTransferredAmount(m, lotus.ActorNameByCode(m.ToActorCode), p),
				// ActorName:          lotus.ActorNameByCode(m.ToActorCode),
			}
			transactionsResults = append(transactionsResults, transaction)
			// if transaction.ExitCode == 0 {
			// 	err = updateMinerAddressChanges(p, m, transaction)
			// 	if err != nil {
			// 		log.Errorw("updating miner address changes", "error", err)
			// 	}
			// }

			// method, params, err := p.parseMessageParams(m.Message, m.ToActorCode)
			// if err == nil {
			// 	pm := &messagemodel.ParsedMessage{
			// 		Height:   msg.Height,
			// 		Cid:      msg.Cid,
			// 		Sender:   msg.From,
			// 		Receiver: msg.To,
			// 		Value:    msg.Value,
			// 		Method:   method,
			// 		Params:   params,
			// 	}
			// 	parsedMessageResults = append(parsedMessageResults, pm)
			// } else {
			// 	log.Errorw("failed to parse message params", "error", err)
			// }
		}
		// _, err := p.store.DB.Model(&transactionsResults).Insert()
		// if err != nil {
		// 	log.Errorw("inserting transactions",
		// 		"error", err,
		// 		"tipset", pts.Height(),
		// 	)
		// } else {
		// 	log.Debugw(
		// 		"inserted transactions",
		// 		"tipset", pts.Height(),
		// 		"count", len(transactionsResults),
		// 	)
		// }
		// _, err = p.store.DB.Model(&parsedMessageResults).Insert()
		// if err != nil {
		// 	log.Errorw("inserting parsed_messages",
		// 		"error", err,
		// 		"tipset", pts.Height(),
		// 	)
		// } else {
		// 	log.Debugw(
		// 		"inserted parsed_messages",
		// 		"tipset", pts.Height(),
		// 		"count", len(parsedMessageResults),
		// 	)
		// }
		// _, err = p.store.DB.Model(&blockMessageResults).Insert()
		// if err != nil {
		// 	log.Errorw("inserting block_messages",
		// 		"error", err,
		// 		"tipset", pts.Height(),
		// 	)
		// } else {
		// 	log.Debugw(
		// 		"inserted block_messages",
		// 		"tipset", pts.Height(),
		// 		"count", len(blockMessageResults),
		// 	)
		// }

		trsFile, _ := json.MarshalIndent(transactionsResults, "", "	")

		err := ioutil.WriteFile("./s3data/transactionsResults/"+pts.Height().String()+".json", trsFile, 0644)
		if err != nil {
			log.Errorw("write transactionsResults",
				"error", err,
			)
		}
	*/

	log.Debug("starting messagegaseconomy calc")
	newBaseFee := store.ComputeNextBaseFee(pts.Blocks()[0].ParentBaseFee, totalUniqGasLimit, len(pts.Blocks()), pts.Height())
	baseFeeRat := new(big.Rat).SetFrac(newBaseFee.Int, new(big.Int).SetUint64(build.FilecoinPrecision))
	baseFee, _ := baseFeeRat.Float64()

	baseFeeChange := new(big.Rat).SetFrac(newBaseFee.Int, pts.Blocks()[0].ParentBaseFee.Int)
	baseFeeChangeF, _ := baseFeeChange.Float64()

	messageGasEconomyResult := &messagemodel.MessageGasEconomy{
		Height:              int64(pts.Height()),
		StateRoot:           pts.ParentState().String(),
		GasLimitTotal:       totalGasLimit,
		GasLimitUniqueTotal: totalUniqGasLimit,
		BaseFee:             baseFee,
		BaseFeeChangeLog:    math.Log(baseFeeChangeF) / math.Log(1.125),
		GasFillRatio:        float64(totalGasLimit) / float64(len(pts.Blocks())*build.BlockGasTarget),
		GasCapacityRatio:    float64(totalUniqGasLimit) / float64(len(pts.Blocks())*build.BlockGasTarget),
		GasWasteRatio:       float64(totalGasLimit-totalUniqGasLimit) / float64(len(pts.Blocks())*build.BlockGasTarget),
	}
	log.Debug("completed messagegaseconomy calc")

	mgersFile, _ := json.MarshalIndent(messageGasEconomyResult, "", "	")

	err = ioutil.WriteFile(projectRoot+"/s3data/messageGasEconomyResult/"+pts.Height().String()+".json", mgersFile, 0644)
	if err != nil {
		log.Errorw("write messageGasEconomyResult",
			"error", err,
		)
	}
	log.Debugw("written mgersfile",
		"height", pts.Height(),
	)
	// _, err = p.store.DB.Model(messageGasEconomyResult).Insert()
	// if err != nil {
	// 	log.Errorw("inserting message_gas_economy",
	// 		"error", err,
	// 		"tipset", pts.Height(),
	// 	)
	// } else {
	// 	log.Debugw(
	// 		"inserted message_gas_economy",
	// 		"tipset", pts.Height(),
	// 	)
	// }
	var pl model.PersistableList

	return pl, nil
}

func (p *Task) parseMessageParams(m *types.Message, destCode cid.Cid) (string, string, error) {
	actor, ok := statediff.LotusActorCodes[destCode.String()]
	if !ok {
		actor = statediff.LotusTypeUnknown
	}
	var params ipld.Node
	var method string
	var err error

	// TODO: the following closure is in place to handle the potential for panic
	// in ipld-prime. Can be removed once fixed upstream.
	// tracking issue: https://github.com/ipld/go-ipld-prime/issues/97
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = xerrors.Errorf("recovered panic: %v", r)
			}
		}()
		params, method, err = statediff.ParseParams(m.Params, int64(m.Method), actor)
	}()
	if err != nil && actor != statediff.LotusTypeUnknown {
		// fall back to generic cbor->json conversion.
		actor = statediff.LotusTypeUnknown
		params, method, err = statediff.ParseParams(m.Params, int64(m.Method), actor)
	}
	if method == "Unknown" {
		method = fmt.Sprintf("%s.%d", actor, m.Method)
	}
	if err != nil {
		log.Warnf("failed to parse parameters of message %s: %v", m.Cid, err)
		// this can occur when the message is not valid cbor
		return method, "", err
	}
	if params == nil {
		return method, "", nil
	}

	buf := bytes.NewBuffer(nil)
	if err := fcjson.Encoder(params, buf); err != nil {
		return "", "", xerrors.Errorf("json encode: %w", err)
	}

	encoded := string(bytes.ReplaceAll(bytes.ToValidUTF8(buf.Bytes(), []byte{}), []byte{0x00}, []byte{}))

	return method, encoded, nil
}

func (p *Task) Close() error {
	return nil
}

type MessageError struct {
	Cid   cid.Cid
	Error string
}
