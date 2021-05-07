package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	// "net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// "github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/credentials"
	// "github.com/aws/aws-sdk-go/aws/session"
	// "github.com/aws/aws-sdk-go/service/s3"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	builtin3 "github.com/filecoin-project/specs-actors/v3/actors/builtin"
	"github.com/filecoin-project/statediff"
	"github.com/filecoin-project/statediff/codec/fcjson"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"golang.org/x/xerrors"

	// "github.com/buidl-labs/filecoin-chain-indexer/chain"
	"github.com/buidl-labs/filecoin-chain-indexer/config"
	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	"github.com/buidl-labs/filecoin-chain-indexer/tasks/messages"
	"github.com/buidl-labs/filecoin-chain-indexer/util"

	// "github.com/buidl-labs/filecoin-chain-indexer/db"
	"github.com/buidl-labs/filecoin-chain-indexer/lens/lotus"
	messagemodel "github.com/buidl-labs/filecoin-chain-indexer/model/messages"
	// "github.com/buidl-labs/filecoin-chain-indexer/model"
	// "github.com/buidl-labs/filecoin-chain-indexer/storage"
)

func Transform(cfg config.Config, csofilename string) error {
	projectRoot := os.Getenv("ROOTDIR")
	lensOpener, lensCloser, err := lotus.NewAPIOpener(cfg, context.Background())
	if err != nil {
		log.Error("setup lens",
			// "tasks", tasks,
			// "taskType", taskType,
			"error", err,
		)
		return xerrors.Errorf("setup lens: %w", err)
	}
	defer func() {
		lensCloser()
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
	log.Info(ts.Height())

	dirname := projectRoot + "/s3data/cso"

	f, err := os.Open(dirname)
	if err != nil {
		log.Error("opening cso dir: %w", err)
		return xerrors.Errorf("opening cso dir: %w", err)
	}
	// files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Error("reading files in cso dir: %w", err)
		return xerrors.Errorf("reading files in cso dir: %w", err)
	}

	plan, err := ioutil.ReadFile(os.Getenv("ACTOR_CODES_JSON"))
	if err != nil {
		log.Error("reading ACTOR_CODES_JSON: %w", err)
		return xerrors.Errorf("reading ACTOR_CODES_JSON: %w", err)
	}
	var results map[string]string // address.Address]cid.Cid
	err = json.Unmarshal([]byte(plan), &results)
	if err != nil {
		log.Error("unmarshal", err)
		return xerrors.Errorf("load actorCodes: %w", err)
	}

	actorCodes := map[address.Address]cid.Cid{}
	actorCodesStr := map[string]string{}
	for k, v := range results {
		actorCodesStr[k] = v
		a, _ := address.NewFromString(k)
		c, _ := cid.Decode(v)
		actorCodes[a] = c
	}

	getActorCode := func(a address.Address) cid.Cid {
		c, ok := actorCodes[a]
		if ok {
			return c
		}

		return cid.Undef
	}

	// for _, file := range files {
	// log.Info(file.Name())
	jsonFile, err := os.Open(dirname + "/" + csofilename)
	if err != nil {
		log.Error("opening csofile", err)
		return xerrors.Errorf("opening a file in cso dir: %w", err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var cso *api.ComputeStateOutput
	// log.Debug(cso)

	json.Unmarshal(byteValue, &cso)
	// log.Info(cso.Root)
	log.Info("len trace: ", len(cso.Trace), " height: ", csofilename)

	msgCount := 0
	subcallsCount := 0

	// var execTraceResults messagemodel.Transactions
	// var execTraceSubcallsResults messagemodel.Transactions

	var transactionsArr = [][]string{
		{"cid", "height", "sender", "receiver", "amount", "type", "gas_fee_cap", "gas_premium", "gas_limit", "size_bytes", "nonce", "method", "method_name", "params", "params_bytes", "transferred", "state_root", "exit_code", "gas_used", "parent_base_fee", "base_fee_burn", "over_estimation_burn", "miner_penalty", "miner_tip", "refund", "gas_refund", "gas_burned", "actor_name", "miner", "return_bytes"},
	}
	var parsedMessagesArr = [][]string{
		{"cid", "height", "sender", "receiver", "value", "method", "params", "miner"},
	}
	for i, ir := range cso.Trace {
		log.Info("ir " + fmt.Sprintf("%d", i))

		// log.Debug("ir StateLookupID start")
		var irToID, irFromID address.Address
		if len(ir.Msg.To.String()) > 10 {
			irToID, err = node.StateLookupID(context.Background(), ir.Msg.To, types.EmptyTSK)
			if err != nil {
				log.Warnw("couldn't lookup receiver's idaddress",
					"receiver", ir.Msg.To,
					"error", err,
				)
				irToID = ir.Msg.To
			}
		} else {
			irToID = ir.Msg.To
		}
		if len(ir.Msg.From.String()) > 10 {
			irFromID, err = node.StateLookupID(context.Background(), ir.Msg.From, types.EmptyTSK)
			if err != nil {
				log.Warnw("couldn't lookup sender's idaddress",
					"sender", ir.Msg.From,
					"error", err,
				)
				irFromID = ir.Msg.From
			}
		} else {
			irFromID = ir.Msg.From
		}
		// log.Debug("ir StateLookupID end")

		heightStr := strings.TrimSuffix(csofilename, filepath.Ext(csofilename))
		heightint64, _ := strconv.ParseInt(heightStr, 10, 64)

		actor, ok := statediff.LotusActorCodes[getActorCode(irToID).String()]
		if !ok {
			actor = statediff.LotusTypeUnknown
		}
		miner := "0"
		// log.Info("ir.MsgRct.Return:", ir.MsgRct.Return, " string ir.MsgRct.Return:", string(ir.MsgRct.Return[:]), " fmtsp:", fmt.Sprintf("%s", ir.MsgRct.Return))
		miner = util.DeriveMiner(&messagemodel.Transaction{
			Cid:         ir.MsgCid.String(),
			Height:      heightint64,
			Sender:      irFromID.String(),
			Receiver:    irToID.String(),
			Amount:      ir.Msg.Value.String(),
			Method:      uint64(ir.Msg.Method),
			ActorName:   string(actor),
			ReturnBytes: string(ir.MsgRct.Return),
		}, miner)
		// {"cid", "height", "sender", "receiver", "amount", "type", "gas_fee_cap", "gas_premium", "gas_limit", "size_bytes", "nonce", "method", "method_name", "params", "params_bytes", "transferred", "state_root", "exit_code", "gas_used", "parent_base_fee", "base_fee_burn", "over_estimation_burn", "miner_penalty", "miner_tip", "refund", "gas_refund", "gas_burned", "actor_name", "miner", "return_bytes"},
		am := []string{ir.MsgCid.String(), fmt.Sprintf("%d", heightint64),
			irFromID.String(), irToID.String(), ir.Msg.Value.String(), "0",
			ir.Msg.GasFeeCap.String(), ir.Msg.GasPremium.String(),
			fmt.Sprintf("%d", ir.Msg.GasLimit), fmt.Sprintf("%d", 0), fmt.Sprintf("%d", ir.Msg.Nonce),
			fmt.Sprintf("%d", ir.Msg.Method), messages.GetMethodName(string(actor), ir.Msg.Method), "0", "\\x",
			computeTransferredAmount(string(actor), ir.Msg.Method, irToID, ir.Msg.Params, node),
			"0", fmt.Sprintf("%d", ir.MsgRct.ExitCode), fmt.Sprintf("%d", ir.MsgRct.GasUsed),
			fmt.Sprintf("%d", 0),
			ir.GasCost.BaseFeeBurn.String(), ir.GasCost.OverEstimationBurn.String(),
			ir.GasCost.MinerPenalty.String(), ir.GasCost.MinerTip.String(),
			ir.GasCost.Refund.String(), fmt.Sprintf("%d", 0), fmt.Sprintf("%d", 0),
			string(actor), miner, fmt.Sprintf("%v", ir.MsgRct.Return)}

		transactionsArr = append(transactionsArr, am)

		// generateParamsStr(irToID, ir.Msg.Params, ir.Msg.Method, node)

		// transaction := &messagemodel.Transaction{
		// 	Cid:        ir.MsgCid.String(),
		// 	Height:     heightint64,
		// 	Sender:     irFromID.String(),
		// 	Receiver:   irToID.String(),
		// 	Amount:     ir.Msg.Value.String(),
		// 	Type:       0,
		// 	GasFeeCap:  ir.Msg.GasFeeCap.String(),
		// 	GasPremium: ir.Msg.GasPremium.String(),
		// 	GasLimit:   ir.Msg.GasLimit,
		// 	Nonce:      ir.Msg.Nonce,
		// 	Method:     uint64(ir.Msg.Method),
		// 	// MethodName:         msg.MethodName,
		// 	Params:      "0",      //keep empty for now
		// 	ParamsBytes: []byte{}, //keep empty for now
		// 	// Transferred:        computeTransferredAmount(m, lotus.ActorNameByCode(m.ToActorCode), p),
		// 	Transferred: computeTransferredAmount(string(actor), ir.Msg.Method, irToID, ir.Msg.Params, node),
		// 	StateRoot:   cso.Root.String(),
		// 	ExitCode:    int64(ir.MsgRct.ExitCode),
		// 	GasUsed:     ir.MsgRct.GasUsed,
		// 	// ParentBaseFee:      m.BlockHeader.ParentBaseFee.String(),
		// 	// SizeBytes:          msgSize,
		// 	BaseFeeBurn:        ir.GasCost.BaseFeeBurn.String(),
		// 	OverEstimationBurn: ir.GasCost.OverEstimationBurn.String(),
		// 	MinerPenalty:       ir.GasCost.MinerPenalty.String(),
		// 	MinerTip:           ir.GasCost.MinerTip.String(),
		// 	Refund:             ir.GasCost.Refund.String(),
		// 	ActorName:          string(actor),
		// 	// GasRefund:          m.GasOutputs.GasRefund,
		// 	// GasBurned:          m.GasOutputs.GasBurned,
		// 	// ActorName:          lotus.ActorNameByCode(m.ToActorCode),
		// }
		// execTraceResults = append(execTraceResults, transaction)
		msgCount += 1

		method, params, err := parseMessageParams(ir.Msg, getActorCode(irToID), node)
		if err == nil {
			// {"cid","height","sender","receiver","value","method","params"},
			pmir := []string{
				ir.MsgCid.String(), fmt.Sprintf("%d", heightint64),
				irFromID.String(), irToID.String(), ir.Msg.Value.String(),
				method, params, miner,
			}
			parsedMessagesArr = append(parsedMessagesArr, pmir)
		}

		for j, sc := range ir.ExecutionTrace.Subcalls {
			log.Info("sc " + fmt.Sprintf("%d.%d", i, j))

			// log.Debug("sc StateLookupID start")
			var scToID, scFromID address.Address
			if len(sc.Msg.To.String()) > 10 {
				scToID, err = node.StateLookupID(context.Background(), sc.Msg.To, types.EmptyTSK)
				if err != nil {
					log.Warnw("couldn't lookup receiver's idaddress",
						"receiver", sc.Msg.To,
						"error", err,
					)
					scToID = sc.Msg.To
				}
			} else {
				scToID = sc.Msg.To
			}
			if len(sc.Msg.From.String()) > 10 {
				scFromID, err = node.StateLookupID(context.Background(), sc.Msg.From, types.EmptyTSK)
				if err != nil {
					log.Warnw("couldn't lookup sender's idaddress",
						"sender", sc.Msg.From,
						"error", err,
					)
					scFromID = sc.Msg.From
				}
			} else {
				scFromID = sc.Msg.From
			}
			// log.Debug("sc StateLookupID end")

			actor, ok := statediff.LotusActorCodes[getActorCode(scToID).String()]
			if !ok {
				actor = statediff.LotusTypeUnknown
			}
			miner := "0"
			// log.Info("sc.MsgRct.Return:", sc.MsgRct.Return, " string sc.MsgRct.Return:", string(sc.MsgRct.Return[:]), " fmtsp:", fmt.Sprintf("%s", sc.MsgRct.Return))
			miner = util.DeriveMiner(&messagemodel.Transaction{
				Cid:         sc.Msg.Cid().String(),
				Height:      heightint64,
				Sender:      scFromID.String(),
				Receiver:    scToID.String(),
				Amount:      sc.Msg.Value.String(),
				Method:      uint64(sc.Msg.Method),
				ActorName:   string(actor),
				ReturnBytes: string(sc.MsgRct.Return),
			}, miner)
			// {"cid", "height", "sender", "receiver", "amount", "type", "gas_fee_cap", "gas_premium", "gas_limit", "size_bytes", "nonce", "method", "method_name", "params", "params_bytes", "transferred", "state_root", "exit_code", "gas_used", "parent_base_fee", "base_fee_burn", "over_estimation_burn", "miner_penalty", "miner_tip", "refund", "gas_refund", "gas_burned", "actor_name", "miner", "return_bytes"},
			amsc := []string{sc.Msg.Cid().String(), fmt.Sprintf("%d", heightint64),
				scFromID.String(), scToID.String(), sc.Msg.Value.String(), "0",
				sc.Msg.GasFeeCap.String(), sc.Msg.GasPremium.String(),
				fmt.Sprintf("%d", sc.Msg.GasLimit), fmt.Sprintf("%d", 0), fmt.Sprintf("%d", sc.Msg.Nonce),
				fmt.Sprintf("%d", sc.Msg.Method), messages.GetMethodName(string(actor), sc.Msg.Method), "0", "\\x",
				computeTransferredAmount(string(actor), sc.Msg.Method, scToID, sc.Msg.Params, node),
				"0", fmt.Sprintf("%d", sc.MsgRct.ExitCode), fmt.Sprintf("%d", sc.MsgRct.GasUsed),
				fmt.Sprintf("%d", 0),
				ir.GasCost.BaseFeeBurn.String(), ir.GasCost.OverEstimationBurn.String(),
				ir.GasCost.MinerPenalty.String(), ir.GasCost.MinerTip.String(),
				ir.GasCost.Refund.String(), fmt.Sprintf("%d", 0),
				fmt.Sprintf("%d", 0), string(actor), miner, fmt.Sprintf("%v", sc.MsgRct.Return)}
			// "0", "0", "0", "0", "0", string(actor)}

			// generateParamsStr(scToID, sc.Msg.Params, sc.Msg.Method, node)

			transactionsArr = append(transactionsArr, amsc)
			// transaction := &messagemodel.Transaction{
			// 	// Height:     int64(pts.Height()),
			// 	Cid:        sc.Msg.Cid().String(),
			// 	Sender:     scFromID.String(),
			// 	Receiver:   scToID.String(),
			// 	Amount:     sc.Msg.Value.String(),
			// 	Type:       0,
			// 	GasFeeCap:  sc.Msg.GasFeeCap.String(),
			// 	GasPremium: sc.Msg.GasPremium.String(),
			// 	GasLimit:   sc.Msg.GasLimit,
			// 	Nonce:      sc.Msg.Nonce,
			// 	Method:     uint64(sc.Msg.Method),
			// 	// MethodName:         msg.MethodName,
			// 	StateRoot:   cso.Root.String(),
			// 	ExitCode:    int64(sc.MsgRct.ExitCode),
			// 	GasUsed:     sc.MsgRct.GasUsed,
			// 	ParamsBytes: []byte{}, //keep empty for now
			// 	Params:      "0",      //keep empty for now
			// }
			// execTraceSubcallsResults = append(execTraceSubcallsResults, transaction)

			subcallsCount += 1

			method, params, err := parseMessageParams(sc.Msg, getActorCode(scToID), node)
			if err == nil {
				// {"cid","height","sender","receiver","value","method","params"},
				pmsc := []string{
					sc.Msg.Cid().String(), fmt.Sprintf("%d", heightint64),
					scFromID.String(), scToID.String(), sc.Msg.Value.String(),
					method, params, miner,
				}
				parsedMessagesArr = append(parsedMessagesArr, pmsc)
			}

		}
	}

	heightStr := strings.TrimSuffix(csofilename, filepath.Ext(csofilename))

	txnsfile, err := os.OpenFile(projectRoot+"/s3data/csvs/transactions/"+heightStr+".csv", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return xerrors.Errorf("creating transactions csv %w", err)
	}
	defer txnsfile.Close()
	txnswriter := csv.NewWriter(txnsfile)
	defer func() {
		txnswriter.Flush()
	}()
	for _, value := range transactionsArr {
		err := txnswriter.Write(value)
		if err != nil {
			return xerrors.Errorf("writing to transactions csv %w", err)
		}
	}

	pmsgsfile, err := os.OpenFile(projectRoot+"/s3data/csvs/parsed_messages/"+heightStr+".csv", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return xerrors.Errorf("creating parsed_messages csv %w", err)
	}
	defer pmsgsfile.Close()
	pmsgswriter := csv.NewWriter(pmsgsfile)
	defer func() {
		pmsgswriter.Flush()
	}()
	for _, value := range parsedMessagesArr {
		err := pmsgswriter.Write(value)
		if err != nil {
			return xerrors.Errorf("writing to parsed_messages csv %w", err)
		}
	}
	/*
		defer func() {
			// s3 upload
			AccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
			SecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
			//  MyRegion := os.Getenv("AWS_REGION")

			sess := session.New(&aws.Config{
				Region: aws.String("us-west-2"),
				Credentials: credentials.NewStaticCredentials(
					AccessKeyID,
					SecretAccessKey,
					"", // a token will be created when the session it's used.
				)})

			txnsfile1, err := os.Open(projectRoot + "/s3data/csvs/transactions/" + heightStr + ".csv")
			if err != nil {
				log.Error("txns1 fopen", err)
				return
			}
			defer txnsfile1.Close()
			// Get file size and read the file content into a buffer
			txnsfileInfo, _ := txnsfile1.Stat()
			var txnssize int64 = txnsfileInfo.Size()
			txnsbuffer := make([]byte, txnssize)
			txnsfile1.Read(txnsbuffer)

			_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
				Bucket:               aws.String("fmm-csv-transactions"),
				Key:                  aws.String(heightStr + ".csv"),
				ACL:                  aws.String("private"),
				Body:                 bytes.NewReader(txnsbuffer),
				ContentLength:        aws.Int64(txnssize),
				ContentType:          aws.String(http.DetectContentType(txnsbuffer)),
				ContentDisposition:   aws.String("attachment"),
				ServerSideEncryption: aws.String("AES256"),
			})
			if err != nil {
				log.Error("txns1 putting to s3: ", err)
				return
			}
			log.Debug("txns1 uploaded to s3")
			e := os.Remove(projectRoot + "/s3data/cso/" + heightStr + ".json")
			if e != nil {
				log.Error("deleting csojson: ", err)
				return
			}
			e = os.Remove(projectRoot + "/s3data/csvs/transactions/" + heightStr + ".csv")
			if e != nil {
				log.Error("deleting txns1: ", err)
				return
			}

			pmsfile, err := os.Open(projectRoot + "/s3data/csvs/parsed_messages/" + heightStr + ".csv")
			if err != nil {
				log.Error("pms fopen", err)
				return
			}
			defer pmsfile.Close()
			// Get file size and read the file content into a buffer
			pmsfileInfo, _ := pmsfile.Stat()
			var pmssize int64 = pmsfileInfo.Size()
			pmsbuffer := make([]byte, pmssize)
			pmsfile.Read(pmsbuffer)

			_, err = s3.New(sess).PutObject(&s3.PutObjectInput{
				Bucket:               aws.String("fmm-csv-parsed-messages"),
				Key:                  aws.String(heightStr + ".csv"),
				ACL:                  aws.String("private"),
				Body:                 bytes.NewReader(pmsbuffer),
				ContentLength:        aws.Int64(pmssize),
				ContentType:          aws.String(http.DetectContentType(pmsbuffer)),
				ContentDisposition:   aws.String("attachment"),
				ServerSideEncryption: aws.String("AES256"),
			})
			if err != nil {
				log.Error("pms putting to s3: ", err)
				return
			}
			log.Debug("pms uploaded to s3")
			e = os.Remove(projectRoot + "/s3data/csvs/parsed_messages/" + heightStr + ".csv")
			if e != nil {
				log.Error("deleting pms: ", err)
				return
			}
		}()
	*/

	// }

	return nil
}

func computeTransferredAmount(actorName string, method abi.MethodNum, toID address.Address, params []byte, node lens.API) string {
	if method == builtin3.MethodsMiner.WithdrawBalance &&
		strings.HasPrefix(actorName, "storageMinerActor") {
		decodedParams, err := node.StateDecodeParams(context.Background(), toID, method, params, types.EmptyTSK)
		if err != nil {
			return "0"
		}
		d := decodedParams.(map[string]interface{})
		amt := d["AmountRequested"].(string)
		return amt
	} else if method == builtin3.MethodsMarket.WithdrawBalance &&
		strings.HasPrefix(actorName, "storageMarketActor") {
		decodedParams, err := node.StateDecodeParams(context.Background(), toID, method, params, types.EmptyTSK)
		if err != nil {
			return "0"
		}
		d := decodedParams.(map[string]interface{})
		amt := d["Amount"].(string)
		return amt
	}
	return "0"
}

func parseMessageParams(m *types.Message, destCode cid.Cid, node lens.API) (string, string, error) {
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
		return method, "0", err
	}
	if params == nil {
		return method, "0", nil
	}

	buf := bytes.NewBuffer(nil)
	if err := fcjson.Encoder(params, buf); err != nil {
		return "", "0", xerrors.Errorf("json encode: %w", err)
	}

	json_bytes := bytes.ReplaceAll(bytes.ToValidUTF8(buf.Bytes(), []byte{}), []byte{0x00}, []byte{})
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, json_bytes); err != nil {
		return method, "0", err
	}

	encoded := buffer.String()
	if encoded == "\"\"" {
		encoded = "0"
	}
	// log.Info("encoded ", encoded)
	return method, encoded, nil
}

func generateParamsStr(toAddr address.Address, params []byte, method abi.MethodNum, node lens.API) string {
	decodedParams, err := node.StateDecodeParams(context.Background(), toAddr, method, params, types.EmptyTSK)
	if err != nil {
		return ""
	}
	decodedParamsStr := fmt.Sprintf("%v", decodedParams)
	return decodedParamsStr
}
