package util

import (
	"bytes"
	"os"
	"strconv"
	"strings"

	power0 "github.com/filecoin-project/specs-actors/actors/builtin/power"
	logging "github.com/ipfs/go-log/v2"

	messagemodel "github.com/buidl-labs/filecoin-chain-indexer/model/messages"
)

var log = logging.Logger("util")

func GetenvInt(key string) (int, error) {
	v, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return -1, err
	}
	return v, nil
}

func DeriveMiner(transaction *messagemodel.Transaction, miner string) string {
	method := transaction.Method
	actorName := transaction.ActorName
	if strings.HasPrefix(actorName, "accountActor") {
		switch method {
		case 0:
			// Send
			// sender: account, receiver: account
		case 1:
			// Constructor
			// sender: system actor (f00), receiver: account
		case 2:
			// PubkeyAddress
			// sender: any, receiver: account
		}
	} else if strings.HasPrefix(actorName, "initActor") {
		switch method {
		case 2:
			// Exec
			// sender: any, receiver: init (f01)
		}
	} else if strings.HasPrefix(actorName, "cronActor") {
		switch method {
		case 2:
			// epochtick
			// sender: f00, receiver: cron (f03)
		}
	} else if strings.HasPrefix(actorName, "rewardActor") {
		switch method {
		case 2:
			// AwardBlockReward
			// f00 -> f02
		case 3:
			// ThisEpochReward
			// any (miner) -> f02
			// TODO: check if the miner calling is random or earning reward
		case 4:
			// UpdateNetworkKPI
			// power actor (f04) -> f02
		}
	} else if strings.HasPrefix(actorName, "multisigActor") {
	} else if strings.HasPrefix(actorName, "paymentChannelActor") {
	} else if strings.HasPrefix(actorName, "storageMarketActor") {
		switch method {
		case 2:
			// addbalance
			// sender: signable (client/provider), receiver: f05
			// amount gets transferred to f05
			// eg: https://filfox.info/en/message/bafy2bzaced7brnlsqal32hjgxbtpytll7rrwh26j27cc7dofejm2m7mo4ykve
			// basically, locking in to escrow (market actor)
			// TODO: derive miner from signable (if possible)
		case 3:
			// withdrawbalance
			// sender: account(owner/worker/client), receiver: f05
			// NOTE: the amount gets transferred to account from f05
			// eg: https://filfox.info/en/message/bafy2bzaceds5j7i4u3wrff7txnnb4zk23gj5vpet36ewsvbto35sizfvtppn6
			// so, this is basically withdrawing from escrow
			// TODO: derive miner from account (if possible)
		case 4:
			// publishstoragedeals
			// sender: signable(worker/control), receiver: f05
			// NOTE: amount=0 (just miner and burn fee spent)
			// eg: https://filfox.info/en/message/bafy2bzaced6z6qkxq2pyjwxm4z7h67t6e7qvk554ktu33rliknhghyw7xhkvi
			// TODO: derive miner from signable (if possible)
		case 5:
			// VerifyDealsForActivation
			// sender: miner, receiver: f05
			// amount=0
			miner = transaction.Sender
		case 6:
			// ActivateDeals
			// sender: miner, receiver:
			miner = transaction.Sender
			// TODO check
		case 7:
			// OnMinerSectorsTerminate
			// sender: miner, receiver:
			miner = transaction.Sender
			// TODO check
		case 8:
			// ComputeDataCommitment
			// sender: miner, receiver:
			miner = transaction.Sender
			// TODO check
		case 9:
			// CronTick
		}
	} else if strings.HasPrefix(actorName, "storagePowerActor") {
		switch method {
		case 2:
			// CreateMiner
			// sender: signable (account/multisig), receiver: f04 (storagePowerActor)
			// NOTE: amount=0 (just miner and burn fee spent)
			// log.Debug("CREATEMINER got ", []byte(transaction.ReturnBytes))
			var ret power0.CreateMinerReturn
			err := ret.UnmarshalCBOR(bytes.NewReader([]byte(transaction.ReturnBytes)))
			if err != nil {
				log.Errorf("failed to decode return value: %w", err)
				return miner
			}
			log.Infow("CreateMiner return", "Return.IDAddress", ret.IDAddress)
			miner = ret.IDAddress.String()
		case 3:
			// UpdateClaimedPower
			// sender: miner, receiver: f04
			miner = transaction.Sender
		case 4:
			// EnrollCronEvent
			// sender: miner
			miner = transaction.Sender
		case 5:
			// OnEpochTickEnd
			// sender: f03 (cron), receiver: f04
		case 6:
			// UpdatePledgeTotal
			// sender: miner,
			miner = transaction.Sender
		case 8:
			// SubmitPoRepForBulkVerify
			// sender: miner,
			miner = transaction.Sender
		case 9:
			// CurrentTotalPower
		}
	} else if strings.HasPrefix(actorName, "storageMinerActor") {
		switch method {
		case 1:
			// Constructor
			// sender: system actor (f00), receiver: miner
			miner = transaction.Receiver
		case 2:
			// ControlAddresses
			// sender: any, receiver: miner
			miner = transaction.Receiver
		case 3:
			// ChangeWorkerAddress
			// sender: owner, receiver: miner
			miner = transaction.Receiver
		case 4:
			// ChangePeerID
			// sender: owner/worker/control, receiver: miner
			miner = transaction.Receiver
		case 5:
			// SubmitWindowedPoSt
			// sender: owner/worker/control, receiver: miner
			miner = transaction.Receiver
		case 6:
			// PreCommitSector
			// sender: control/owner/worker, receiver: miner
			miner = transaction.Receiver
		case 7:
			// ProveCommitSector
			// sender: any, receiver: miner
			miner = transaction.Receiver
		case 8:
			// ExtendSectorExpiration
			// sender: control/owner/worker, receiver: miner
			miner = transaction.Receiver
		case 9:
			// TerminateSectors
			// sender: control/owner/worker, receiver: miner
			miner = transaction.Receiver
		case 10:
			// DeclareFaults
			// sender: control/owner/worker, receiver: miner
			miner = transaction.Receiver
		case 11:
			// DeclareFaultsRecovered
			// sender: control/owner/worker, receiver: miner
			miner = transaction.Receiver
		case 12:
			// OnDeferredCronEvent
			// sender: power actor, receiver:
			// TODO check
			miner = transaction.Receiver
		case 13:
			// CheckSectorProven
			// sender: any, receiver:
			// TODO check
			miner = transaction.Receiver
		case 14:
			// ApplyRewards
			// sender: f02, receiver: miner
			// amount!=0
			miner = transaction.Receiver
		case 15:
			// ReportConsensusFault
			// sender: signable, receiver: miner
			// eg: https://filfox.info/en/message/bafy2bzacecaapgafjyquysif3rpjcewahdelabu4zssy5a5jtylddvzkpavhq
			// burn: miner -> f099
			// transfer: miner -> caller (sender)
			miner = transaction.Receiver
		case 16:
			// WithdrawBalance
			// sender: owner, receiver: miner
			// transfer is from miner to owner though
			miner = transaction.Receiver
		case 17:
			// ConfirmSectorProofsValid
			// sender: storagePowerActor, receiver:
			// TODO check
			miner = transaction.Receiver
		case 18:
			// ChangeMultiaddrs
			// sender: control/owner/worker, receiver: miner
			miner = transaction.Receiver
		case 19:
			// CompactPartitions
			// sender: control/owner/worker, receiver:
			// TODO check
			miner = transaction.Receiver
		case 20:
			// CompactSectorNumbers
			// sender: control/owner/worker, receiver: miner
			miner = transaction.Receiver
		case 21:
			// ConfirmUpdateWorkerKey
			// sender: owner, receiver: miner
			miner = transaction.Receiver
		case 22:
			// RepayDebt
			// sender: control/owner/worker, receiver: miner
			// eg: https://filfox.info/en/message/bafy2bzaceaoqgncosqiozpf3sr4ad6cjsyawy7lo55thdr3njni3xht573umi
			// sender -> miner (transfer), miner -> f099 (burn) {the same transferred amount is burnt}
			miner = transaction.Receiver
		case 23:
			// ChangeOwnerAddress
			// sender: owner, receiver: miner
			miner = transaction.Receiver
		case 24:
			// DisputeWindowedPoSt
			// sender: signable, receiver: miner
			miner = transaction.Receiver
		}
	} else if strings.HasPrefix(actorName, "verifiedRegistryActor") {
	}
	return miner
}
