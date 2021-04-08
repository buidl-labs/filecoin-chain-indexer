package messages

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	builtin3 "github.com/filecoin-project/specs-actors/v3/actors/builtin"

	"github.com/buidl-labs/filecoin-chain-indexer/lens"
	messagemodel "github.com/buidl-labs/filecoin-chain-indexer/model/messages"
)

func updateMinerAddressChanges(p *Task, m *lens.ExecutedMessage, transaction *messagemodel.Transaction) error {
	if transaction.Method == 3 && transaction.ActorName == "fil/3/storageminer" {
		// change in worker/control addr
		tts, _ := p.node.ChainGetTipSetByHeight(context.Background(), abi.ChainEpoch(transaction.Height), types.EmptyTSK)
		nts, _ := p.node.ChainGetTipSetByHeight(context.Background(), abi.ChainEpoch(transaction.Height+1), types.EmptyTSK)
		tmi, _ := p.node.StateMinerInfo(context.Background(), m.Message.To, tts.Key())
		nmi, _ := p.node.StateMinerInfo(context.Background(), m.Message.To, nts.Key())
		if !sameStringSlice(tmi.ControlAddresses, nmi.ControlAddresses) {
			// control addr changed
			jsonFile, err := os.Open(os.Getenv("ADDR_CHANGES"))
			if err != nil {
				return err
			}
			byteValue, _ := ioutil.ReadAll(jsonFile)
			var minerAddressChanges map[string]MinerAddressChanges
			json.Unmarshal(byteValue, &minerAddressChanges)
			jsonFile.Close()
			changedMiner := minerAddressChanges[m.Message.To.String()]
			controlChanges := changedMiner.ControlChanges
			var fromCAs []string
			var toCAs []string
			for _, ca := range tmi.ControlAddresses {
				fromCAs = append(fromCAs, ca.String())
			}
			for _, ca := range nmi.ControlAddresses {
				toCAs = append(toCAs, ca.String())
			}
			if len(fromCAs) == 0 {
				fromCAs = make([]string, 0)
			}
			if len(toCAs) == 0 {
				toCAs = make([]string, 0)
			}
			controlChanges = append(controlChanges, ControlChange{
				Epoch: transaction.Height + 1,
				From:  fromCAs,
				To:    toCAs,
			})
			changedMiner.ControlChanges = controlChanges
			minerAddressChanges[m.Message.To.String()] = changedMiner

			mdata, err := json.MarshalIndent(minerAddressChanges, "", "	")
			if err != nil {
				return err
			}
			jsonStr := string(mdata)
			f, err := os.OpenFile(os.Getenv("ADDR_CHANGES"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			defer f.Close()
			w := bufio.NewWriter(f)
			n4, _ := w.WriteString(jsonStr)
			log.Debug("wrote %d bytes\n", n4)
			w.Flush()
		} else {
			// worker changed
			if nmi.NewWorker.String() != "<empty>" {
				jsonFile, err := os.Open(os.Getenv("ADDR_CHANGES"))
				if err != nil {
					return err
				}
				byteValue, _ := ioutil.ReadAll(jsonFile)
				var minerAddressChanges map[string]MinerAddressChanges
				json.Unmarshal(byteValue, &minerAddressChanges)
				jsonFile.Close()
				changedMiner := minerAddressChanges[m.Message.To.String()]
				workerChanges := changedMiner.WorkerChanges
				workerChanges = append(workerChanges, WorkerChange{
					Epoch: transaction.Height + 900 + 1,
					From:  tmi.Worker.String(),
					To:    nmi.NewWorker.String(),
				})
				changedMiner.WorkerChanges = workerChanges
				minerAddressChanges[m.Message.To.String()] = changedMiner

				mdata, err := json.MarshalIndent(minerAddressChanges, "", "	")
				if err != nil {
					return err
				}
				jsonStr := string(mdata)
				f, err := os.OpenFile(os.Getenv("ADDR_CHANGES"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
				if err != nil {
					return err
				}
				defer f.Close()
				w := bufio.NewWriter(f)
				n4, _ := w.WriteString(jsonStr)
				log.Debug("wrote %d bytes\n", n4)
				w.Flush()
			}
		}
	}
	if transaction.Method == 23 && transaction.ActorName == "fil/3/storageminer" {
		// change in owner addr
		tts, _ := p.node.ChainGetTipSetByHeight(context.Background(), abi.ChainEpoch(transaction.Height), types.EmptyTSK)
		nts, _ := p.node.ChainGetTipSetByHeight(context.Background(), abi.ChainEpoch(transaction.Height+1), types.EmptyTSK)
		tmi, _ := p.node.StateMinerInfo(context.Background(), m.Message.To, tts.Key())
		nmi, _ := p.node.StateMinerInfo(context.Background(), m.Message.To, nts.Key())

		jsonFile, err := os.Open(os.Getenv("ADDR_CHANGES"))
		if err != nil {
			return err
		}
		byteValue, _ := ioutil.ReadAll(jsonFile)
		var minerAddressChanges map[string]MinerAddressChanges
		json.Unmarshal(byteValue, &minerAddressChanges)
		jsonFile.Close()
		changedMiner := minerAddressChanges[m.Message.To.String()]
		ownerChanges := changedMiner.OwnerChanges
		ownerChanges = append(ownerChanges, OwnerChange{
			Epoch: transaction.Height + 1,
			From:  tmi.Owner.String(),
			To:    nmi.Owner.String(),
		})
		changedMiner.OwnerChanges = ownerChanges
		minerAddressChanges[m.Message.To.String()] = changedMiner

		mdata, err := json.MarshalIndent(minerAddressChanges, "", "	")
		if err != nil {
			return err
		}
		jsonStr := string(mdata)
		f, err := os.OpenFile(os.Getenv("ADDR_CHANGES"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		n4, _ := w.WriteString(jsonStr)
		log.Debug("wrote %d bytes\n", n4)
		w.Flush()
	}
	return nil
}

func generateParamsStr(toAddr address.Address, params []byte, method abi.MethodNum, p *Task) string {
	decodedParams, err := p.node.StateDecodeParams(context.Background(), toAddr, method, params, types.EmptyTSK)
	if err != nil {
		return ""
	}
	decodedParamsStr := fmt.Sprintf("%v", decodedParams)
	return decodedParamsStr
}

func computeTransferredAmount(m *lens.ExecutedMessage, actorName string, p *Task) string {
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

func GetMethodName(actorName string, methodNum abi.MethodNum) string {
	if methodNum == 0 {
		return commonMethods[0]
	} else if methodNum == 1 {
		return commonMethods[1]
	}
	switch actorName {
	case "accountActor", "accountActorV2", "accountActorV3":
		return accountMethods[methodNum]
	case "initActor", "initActorV2", "initActorV3":
		return initMethods[methodNum]
	case "cronActor", "cronActorV2", "cronActorV3":
		return cronMethods[methodNum]
	case "rewardActor", "rewardActorV2", "rewardActorV3":
		return rewardMethods[methodNum]
	case "multisigActor", "multisigActorV2", "multisigActorV3":
		return multisigMethods[methodNum]
	case "paymentChannelActor", "paymentChannelActorV2", "paymentChannelActorV3":
		return paychMethods[methodNum]
	case "storageMarketActor", "storageMarketActorV2", "storageMarketActorV3":
		return marketMethods[methodNum]
	case "storagePowerActor", "storagePowerActorV2", "storagePowerActorV3":
		return powerMethods[methodNum]
	case "storageMinerActor", "storageMinerActorV2", "storageMinerActorV3":
		return minerMethods[methodNum]
	case "verifiedRegistryActor","verifiedRegistryActorV2","verifiedRegistryActorV3":
		return verifiedRegistryMethods[methodNum]
	default:
		return "0"
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

type OwnerChange struct {
	Epoch int64  `json:"epoch,omitempty"`
	From  string `json:"from,omitempty"`
	To    string `json:"to,omitempty"`
}

type WorkerChange struct {
	Epoch int64  `json:"epoch,omitempty"`
	From  string `json:"from,omitempty"`
	To    string `json:"to,omitempty"`
}

type ControlChange struct {
	Epoch int64    `json:"epoch,omitempty"`
	From  []string `json:"from,omitempty"`
	To    []string `json:"to,omitempty"`
}

type MinerAddressChanges struct {
	OwnerChanges   []OwnerChange   `json:"ownerChanges,omitempty"`
	WorkerChanges  []WorkerChange  `json:"workerChanges,omitempty"`
	ControlChanges []ControlChange `json:"controlChanges,omitempty"`
}

func sameStringSlice(x, y []address.Address) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[address.Address]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y] -= 1
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	if len(diff) == 0 {
		return true
	}
	return false
}
