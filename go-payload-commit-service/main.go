package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-redis/redis/v8"
	shell "github.com/ipfs/go-ipfs-api"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"golang.org/x/time/rate"

	"github.com/powerloom/goutils/logger"
	"github.com/powerloom/goutils/settings"
)

var ctx = context.Background()

var redisClient *redis.Client
var ipfsClient *shell.Shell
var txMgrHttpClient http.Client
var dagFinalizerClient http.Client
var w3sHttpClient http.Client
var settingsObj *settings.SettingsObj
var consts ConstsObj
var rmqConnection *Conn
var exitChan chan bool
var WaitQueueForConsensus map[string]*PayloadCommit
var QueueLock sync.Mutex

//Rate Limiter Objects
var dagFinalizerClientRateLimiter *rate.Limiter
var web3StorageClientRateLimiter *rate.Limiter
var ipfsClientRateLimiter *rate.Limiter
var txClientRateLimiter *rate.Limiter

var REDIS_KEY_PROJECT_PAYLOAD_CIDS = "projectID:%s:payloadCids"
var REDIS_KEY_PROJECT_PENDING_TXNS = "projectID:%s:pendingTransactions"
var REDIS_KEY_PROJECT_TENTATIVE_BLOCK_HEIGHT = "projectID:%s:tentativeBlockHeight"

var WEB3_STORAGE_UPLOAD_URL_SUFFIX = "/upload"

var SKIP_SNAPSHOT_VALIDATION_ERR_STR = "skip validation"

var commonTxReqParams CommonTxRequestParams

type retryType int64

const (
	NO_RETRY_SUCCESS retryType = iota
	RETRY_IMMEDIATE            //TO be used in timeout scenarios or non server returned error scenarios.
	RETRY_WITH_DELAY           //TO be used when immediate error is returned so that server is not overloaded.
	NO_RETRY_FAILURE           //This is to be used for unexpected conditions which are not recoverable and hence no retry
)

func (r retryType) String() string {
	switch r {
	case NO_RETRY_SUCCESS:
		return "NO Retry"
	case RETRY_IMMEDIATE:
		return "Retry Immediately"
	case RETRY_WITH_DELAY:
		return "Retry with a delay"
	case NO_RETRY_FAILURE:
		return "Non recoverable error, hence not retrying."
	}
	return "unknown"
}

func RegisterSignalHandles() {
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		for {
			s := <-signalChanel
			switch s {
			// kill -SIGHUP XXXX [XXXX - PID for your program]
			case syscall.SIGHUP:
				log.Info("Signal hang up triggered.")

				// kill -SIGINT XXXX or Ctrl+c  [XXXX - PID for your program]
			case syscall.SIGINT:
				log.Info("Signal interrupt triggered.")
				GracefulShutDown()

				// kill -SIGTERM XXXX [XXXX - PID for your program]
			case syscall.SIGTERM:
				log.Info("Signal terminte triggered.")
				GracefulShutDown()

				// kill -SIGQUIT XXXX [XXXX - PID for your program]
			case syscall.SIGQUIT:
				log.Info("Signal quit triggered.")
				GracefulShutDown()

			default:
				log.Info("Unknown signal.", s)
				GracefulShutDown()
			}
		}
	}()

}

func GracefulShutDown() {
	rmqConnection.StopConsumer()
	time.Sleep(2 * time.Second)
	redisClient.Close()
	exitChan <- true
}

type SubmittedTransactionStates struct {
	CallbackPending  int `json:"callbackPending"`
	CallbackReceived int `json:"callbackReceived"`
}
type ConstsObj struct {
	SubmittedTxnStates SubmittedTransactionStates `json:"submittedTxnStates"`
}

func main() {

	RegisterSignalHandles()
	logger.InitLogger()
	settingsObj = settings.ParseSettings("../settings.json")
	if settingsObj.InstanceId == "" {
		log.Fatalf("InstanceID is set to null, please generate and set a unique instanceID")
		os.Exit(1)
	}
	ParseConsts("../dev_consts.json")
	InitIPFSClient()
	InitRedisClient()
	InitTxManagerClient()
	InitDAGFinalizerCallbackClient()
	log.Info("Starting RabbitMq Consumer")
	InitW3sClient()
	var wg sync.WaitGroup
	if settingsObj.UseConsensus {
		wg.Add(1)
		InitConsensusClient()
		WaitQueueForConsensus = make(map[string]*PayloadCommit, 100) //TODO Make this queueSize configurable
		go func() {
			defer wg.Done()
			PollConsensusForConfirmations()
		}()
	}

	InitRabbitmqConsumer()
	if settingsObj.UseConsensus {
		wg.Wait()
	}
}

func ParseConsts(constsFile string) {
	log.Info("Reading Consts File:", constsFile)
	data, err := os.ReadFile(constsFile)
	if err != nil {
		log.Error("Cannot read the file:", err)
		panic(err)
	}

	log.Debug("Consts json data is", string(data))
	err = json.Unmarshal(data, &consts)
	if err != nil {
		log.Error("Cannot unmarshal the Consts json ", err)
		panic(err)
	}
}

func InitRabbitmqConsumer() {
	var err error
	rabbitMqURL := fmt.Sprintf("amqp://%s:%s@%s:%d/", settingsObj.Rabbitmq.User, settingsObj.Rabbitmq.Password, settingsObj.Rabbitmq.Host, settingsObj.Rabbitmq.Port)
	rmqConnection, err = GetConn(rabbitMqURL)
	log.Infof("Starting rabbitMQ consumer connecting to URL: %s with concurreny %d", rabbitMqURL, settingsObj.PayloadCommitConcurrency)
	if err != nil {
		panic(err)
	}
	rmqExchangeName := settingsObj.Rabbitmq.Setup.Core.Exchange
	//TODO: These settings need to be moved to json config.
	rmqQueueName := "audit-protocol-commit-payloads"
	rmqRoutingKey := "commit-payloads"

	err = rmqConnection.StartConsumer(rmqQueueName, rmqExchangeName, rmqRoutingKey, RabbitmqMsgHandler, settingsObj.PayloadCommitConcurrency)
	if err != nil {
		panic(err)
	}

	exitChan = make(chan bool)
	<-exitChan
}

func RabbitmqMsgHandler(d amqp.Delivery) bool {
	if d.Body == nil {
		log.Errorf("Received message %+v from rabbitmq without message body! Ignoring and not processing it.", d)
		return true
	}
	log.Tracef("Received Message from rabbitmq %v", d.Body)

	var payloadCommit PayloadCommit

	err := json.Unmarshal(d.Body, &payloadCommit)
	if err != nil {
		log.Warnf("CRITICAL: Json unmarshal failed for payloadCommit %v, with err %v. Ignoring", d.Body, err)
		return true
	}

	if !payloadCommit.Resubmitted {
		if d.Redelivered {
			log.Warnf("Message got redelivered from rabbitmq for project %s and commitId %s",
				payloadCommit.ProjectId, payloadCommit.CommitId)
		}
		lastTentativeBlockHeight, err := GetTentativeBlockHeight(payloadCommit.ProjectId)
		if err != nil {
			return false
		}
		//Note: Once we bring in consensus, tentativeBlockHeight should be updated only once consensus is achieved
		payloadCommit.TentativeBlockHeight = lastTentativeBlockHeight + 1
		if lastTentativeBlockHeight > 0 {
			isValidSnapshot, err := validSnapshot(&payloadCommit, lastTentativeBlockHeight)
			if err != nil {
				log.Errorf("Could not validate current snapshot for project %s with commitId %s",
					payloadCommit.ProjectId, payloadCommit.CommitId)
				return false
			}
			if !isValidSnapshot {
				log.Warnf("Invalid snapshot received for project %s at tentativeBlockHeight %d and ignoring it",
					payloadCommit.ProjectId, lastTentativeBlockHeight)
				return true
			}
		}
		var ipfsStatus bool
		if payloadCommit.Web3Storage {
			var wg sync.WaitGroup
			var w3sStatus bool
			log.Debugf("Received incoming Payload commit message at tentative DAG Height %d for project %s with commitId %s from rabbitmq. Uploading payload to web3.storage and IPFS.",
				payloadCommit.TentativeBlockHeight, payloadCommit.ProjectId, payloadCommit.CommitId)
			wg.Add(1)
			go func() {
				defer wg.Done()
				ipfsStatus = UploadSnapshotToIPFS(&payloadCommit)
				if !ipfsStatus {
					return
				}
				log.Debugf("IPFS add Successful. Snapshot CID is %s for project %s with commitId %s at tentativeBlockHeight %d",
					payloadCommit.SnapshotCID, payloadCommit.ProjectId, payloadCommit.CommitId, payloadCommit.TentativeBlockHeight)
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				var snapshotCid string
				snapshotCid, w3sStatus = UploadToWeb3Storage(&payloadCommit)
				if !w3sStatus {
					return
				}
				log.Debugf("web3.storage upload Successful. Snapshot CID is %s for project %s with commitId %s at tentativeBlockHeight %d",
					snapshotCid, payloadCommit.ProjectId, payloadCommit.CommitId, payloadCommit.TentativeBlockHeight)
				//SnapshotCID from web3.storage is not used directly.
				//payloadCommit.SnapshotCID = snapshotCid
			}()

			wg.Wait()
			if !ipfsStatus {
				log.Errorf("Failed to add to IPFS. IPFSStatus %b , web3.storage status %b", ipfsStatus, w3sStatus)
				return false
			}
			if !w3sStatus {
				//Since web3.storage is only used as a backup, we proceed even if web3.storage upload fails as IPFS upload is successful
				log.Errorf("Failed to upload to web3.storage status %b. Continuing with processing.", w3sStatus)
			}
		} else {
			log.Debugf("Received incoming Payload commit message at tentative DAG Height %d for project %s with commitId %s from rabbitmq. Adding payload to IPFS.",
				payloadCommit.TentativeBlockHeight, payloadCommit.ProjectId, payloadCommit.CommitId)
			ipfsStatus = UploadSnapshotToIPFS(&payloadCommit)
			if !ipfsStatus {
				return false
			}
		}

		err = StorePayloadCidInRedis(&payloadCommit)
		if err != nil {
			log.Errorf("Failed to store payloadCid in redis for the project %s with commitId %s due to error %+v",
				payloadCommit.ProjectId, payloadCommit.CommitId, err)
			return false
		}
		//Update TentativeBlockHeight for the project
		err = UpdateTentativeBlockHeight(&payloadCommit)
		if err != nil {
			log.Errorf("Failed to update tentativeBlockHeight for the project %s with commitId %s due to error %+v",
				payloadCommit.ProjectId, payloadCommit.CommitId, err)
			return false
		}
	} else {
		if payloadCommit.SnapshotCID == "" && payloadCommit.Payload == nil {
			log.Warnf("Received incoming Payload commit message without snapshotCID and empty payload at tentative DAG Height %d for project %s for resubmission at block %d from rabbitmq. Discarding this message without processing.",
				payloadCommit.TentativeBlockHeight, payloadCommit.ProjectId, payloadCommit.ResubmissionBlock)
			return true
		} else {
			//TODO: What if payload is present and snapshotCID is empty?? Currently there is no scenario where this can happen, but need to handle in future.
			//This would require soem kind of reorg of DAGChain if required as this is a resubmission of payload already submitted.
			log.Debugf("Received incoming Payload commit message at tentative DAG Height %d for project %s for resubmission at block %d from rabbitmq.",
				payloadCommit.TentativeBlockHeight, payloadCommit.ProjectId, payloadCommit.ResubmissionBlock)
		}
	}

	var txHash string
	var retryType retryType
	if !payloadCommit.SkipAnchorProof {
		txHash, retryType = PrepareAndSubmitTxnToChain(&payloadCommit)
		if retryType == RETRY_IMMEDIATE || retryType == RETRY_WITH_DELAY {
			//TODO: Not retrying further..need to think of project recovery from this point.
			log.Warnf("MAX Retries reached while trying to invoke tx-manager for project %s and commitId %s with tentativeBlockHeight %d.",
				payloadCommit.ProjectId, payloadCommit.CommitId, payloadCommit.TentativeBlockHeight, " Not retrying further.")
		} else if retryType == NO_RETRY_SUCCESS {
			log.Trace("Submitted txn to chain for %+v", payloadCommit)
		} else {
			log.Errorf("Irrecoverable error occurred and hence ignoring snapshots for processing.")
			return true
		}
	} else {
		var status bool
		txHash, status = GenerateTokenHash(&payloadCommit)
		payloadCommit.ApiKeyHash = txHash
		if !status {
			log.Errorf("Irrecoverable error occurred for project %s and commitId %s with tentativeBlockHeight %d and hence ignoring snapshot for processing.",
				payloadCommit.ProjectId, payloadCommit.CommitId, payloadCommit.TentativeBlockHeight)
			return true
		}
		//Wait for consensus
		if settingsObj.UseConsensus {
			//TODO: Move this queue to redis
			status, err := SubmitSnapshotForConsensus(&payloadCommit)
			if status == SNAPSHOT_CONSENSUS_STATUS_ACCEPTED {
				QueueLock.Lock()
				WaitQueueForConsensus[payloadCommit.ProjectId] = &payloadCommit
				QueueLock.Unlock()
				//TODO: Notify polling go-routine
				return true
			}
			if err != nil {
				return false
			}
		}
	}
	return AddToPendingTxns(&payloadCommit, txHash)
}

func AddToPendingTxns(payloadCommit *PayloadCommit, txHash string) bool {
	/*Add to redis pendingTransactions*/
	err := AddToPendingTxnsInRedis(payloadCommit, txHash)
	if err != nil {
		//TODO: Not retrying further..need to think of project recovery from this point.
		log.Errorf("Unable to add transaction to PendingTxns after max retries for project %s and commitId %s with tentativeBlockHeight %d.",
			payloadCommit.ProjectId, payloadCommit.CommitId, payloadCommit.TentativeBlockHeight, " Not retrying further as this could be a system level issue!")
		return true
	}
	if payloadCommit.SkipAnchorProof {
		//Notify DAG finalizer service as we are skipping proof anchor on chain.
		retryType := InvokeDAGFinalizerCallback(payloadCommit)
		if retryType == RETRY_IMMEDIATE || retryType == RETRY_WITH_DELAY {
			//TODO: Not retrying further..need to think of project recovery from this point.
			log.Warnf("MAX Retries reached while trying to invoke DAG finalizer for project %s and commitId %s with tentativeBlockHeight %d.",
				payloadCommit.ProjectId, payloadCommit.CommitId, payloadCommit.TentativeBlockHeight, " Not retrying further as this could be a system level issue!")
			return true
		} else if retryType == NO_RETRY_SUCCESS {
			log.Trace("Submitted txn to chain for %+v", payloadCommit)
		}
	}
	return true
}

func UploadSnapshotToIPFS(payloadCommit *PayloadCommit) bool {
	for retryCount := 0; ; {

		err := ipfsClientRateLimiter.Wait(context.Background())
		if err != nil {
			log.Errorf("IPFSClient Rate Limiter wait timeout with error %+v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		snapshotCid, err := ipfsClient.Add(bytes.NewReader(payloadCommit.Payload), shell.CidVersion(1))

		if err != nil {
			if retryCount == *settingsObj.RetryCount {
				log.Errorf("IPFS Add failed for message %+v after max-retry of %d, with err %v", payloadCommit, *settingsObj.RetryCount, err)
				return false
			}
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			retryCount++
			log.Errorf("IPFS Add failed for message %v, with err %v..retryCount %d .", payloadCommit, err, retryCount)
			continue
		}
		log.Debugf("IPFS add Successful. Snapshot CID is %s for project %s with commitId %s at tentativeBlockHeight %d",
			snapshotCid, payloadCommit.ProjectId, payloadCommit.CommitId, payloadCommit.TentativeBlockHeight)
		payloadCommit.SnapshotCID = snapshotCid
		break
	}
	return true
}

func GetPreviousSnapshot(projectId string, lastTentativeBlockHeight int) (*PayloadData, error) {
	payloadCid, err := GetPayloadCidAtProjectHeightFromRedis(projectId, strconv.Itoa(lastTentativeBlockHeight))
	if err != nil {
		log.Errorf("Failed to fetch payloadCid for project %s at height %d from redis due to error %+v",
			projectId, lastTentativeBlockHeight, err)
		return nil, err
	}
	payload, err := GetPayloadFromIPFS(payloadCid, 1)
	if err != nil {
		log.Errorf("Failed to fetch payload from IPFS for CID %s for project %s at height %d from redis due to error %+v",
			payloadCid, projectId, lastTentativeBlockHeight, err)
		return nil, err
	}
	return payload, nil
}

func GetPayloadCidAtProjectHeightFromRedis(projectId string, startScore string) (string, error) {
	//key := projectId + ":payloadCids"
	key := fmt.Sprintf(REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectId)
	payloadCid := ""

	log.Debug("Fetching PayloadCid from redis at key:", key, ",with startScore: ", startScore)
	for i := 0; ; {
		zRangeByScore := redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
			Min: startScore,
			Max: startScore,
		})

		err := zRangeByScore.Err()
		log.Debug("Result for ZRangeByScoreWithScores : ", zRangeByScore)
		if err != nil {
			log.Errorf("Could not fetch PayloadCid from  redis for project %s at blockHeight %d error: %+v Query: %+v",
				projectId, startScore, err, zRangeByScore)
			if i == *settingsObj.RetryCount {
				log.Errorf("Could not fetch PayloadCid from  redis after max retries for project %s at blockHeight %d error: %+v Query: %+v",
					projectId, startScore, err, zRangeByScore)
				return "", err
			}
			i++
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			continue
		}

		res := zRangeByScore.Val()

		//dagPayloadsInfo = make([]DagPayload, len(res))
		log.Debugf("Fetched %d Payload CIDs for key %s", len(res), key)
		if len(res) == 1 {
			payloadCid = fmt.Sprintf("%v", res[0].Member)
			log.Debugf("PayloadCID %s fetched for project %s at height %s from redis", payloadCid, projectId, startScore)
		} else if len(res) > 1 {
			log.Errorf("Found more than 1 payload CIDS at height %d for project %s which means project state is messed up due to an issue that has occured while previous snapshot processing, considering the first one so that current snapshot processing can proceed",
				startScore, projectId)
			payloadCid = fmt.Sprintf("%v", res[0].Member)
		} else {
			log.Errorf("Could not find a payloadCid at height %s for project %s. Trying with lower height.", startScore, projectId)
			prevHeight, err := strconv.Atoi(startScore)
			if err != nil {
				log.Errorf("CRITICAL! Height passed in startScore is not int, hence failed to convert due to error %+v", err)
				return "", err
			}
			startScore = strconv.Itoa(prevHeight - 1)
			if prevHeight < 2 { //Safety check
				log.Errorf("Could not find any payloadCid at min height.Continuing with current snapshot processing in this case.")
				return "", errors.New(SKIP_SNAPSHOT_VALIDATION_ERR_STR)
			}
			i++
			continue
		}
		break
	}
	return payloadCid, nil
}

func GetPayloadFromIPFS(payloadCid string, retryCount int) (*PayloadData, error) {
	var payload PayloadData
	for i := 0; ; {
		log.Debugf("Fetching payloadCid %s from IPFS", payloadCid)
		data, err := ipfsClient.Cat(payloadCid)
		if err != nil {
			if i >= retryCount {
				log.Errorf("Failed to fetch Payload with CID %s from IPFS even after max retries due to error %+v.", payloadCid, err)
				return &payload, errors.New(SKIP_SNAPSHOT_VALIDATION_ERR_STR)
			}
			log.Errorf("Failed to fetch Payload from IPFS, CID %s due to error %+v", payloadCid, err)
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			i++
			continue
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(data)

		err = json.Unmarshal(buf.Bytes(), &payload)
		if err != nil {
			log.Errorf("Failed to Unmarshal Json Payload from IPFS, CID %s, bytes: %+v due to error %+v ",
				payloadCid, buf, err)
			return nil, err
		}
		break
	}

	log.Debugf("Fetched Payload with CID %s from IPFS: %+v", payloadCid, payload)
	return &payload, nil
}

func validSnapshot(payloadCommit *PayloadCommit, lastTentativeBlockHeight int) (bool, error) {
	var currentSnapshotData PayloadData
	err := json.Unmarshal(payloadCommit.Payload, &currentSnapshotData)
	if err != nil {
		log.Warnf("CRITICAL:Unable to decode current snapshot for project %s with commitId %s as payload due to error %+v",
			payloadCommit.ProjectId, payloadCommit.CommitId, err)
		return false, nil
	}
	if currentSnapshotData.ChainHeightRange == nil {
		//If chainHeightRange is not present(like for summary projects), don't look for previousSnapshot.
		log.Debugf("Skip validating with previousSnapshot as chainHeightrange is not present in the snapshot for projectID %s", payloadCommit.ProjectId)
		return true, nil
	}
	payloadCommit.EpochEndBlockHeight = currentSnapshotData.ChainHeightRange.End
	// Check if this is a duplicate snapshot of previously submitted one and ignore if so.
	// We could've used snapshotCID to compare with previous one, but a snapshotter can submit incorrect snapshot at same chainHeightRange?
	// Fetch previously submitted payLoad and compare chainHeight to confirm duplicate.
	previousSnapshot, err := GetPreviousSnapshot(payloadCommit.ProjectId, lastTentativeBlockHeight)
	if previousSnapshot == nil {
		log.Warnf("Could not get previous snapshot for project %s with commitId %s, sending NACK to rabbitmq",
			payloadCommit.ProjectId, payloadCommit.CommitId)
		if err.Error() == SKIP_SNAPSHOT_VALIDATION_ERR_STR {
			log.Debugf("Skip validating with previousSnapshot as unable to fetch it")
			return true, nil
		}
		return false, errors.New("could not get previous snapshot for project")
	}

	if currentSnapshotData.ChainHeightRange != nil && previousSnapshot.ChainHeightRange != nil {
		if (currentSnapshotData.ChainHeightRange.Begin == previousSnapshot.ChainHeightRange.Begin) &&
			(currentSnapshotData.ChainHeightRange.End == previousSnapshot.ChainHeightRange.End) {
			log.Warnf("Duplicate snapshot received for project %s with commitId %s as chainHeightRange is same for previous and current %+v",
				payloadCommit.ProjectId, payloadCommit.CommitId, currentSnapshotData.ChainHeightRange)
			return false, nil
		}
		if currentSnapshotData.ChainHeightRange.Begin != previousSnapshot.ChainHeightRange.End+1 {
			log.Warnf("Snapshot received is out of sequence in comparison to previously submitted one for project %s and commitId %s.Either another go-routine is processing in parallel or there was a snapshot submission missed",
				payloadCommit.ProjectId, payloadCommit.CommitId)
		}
	}
	return true, nil
}

func StorePayloadCidInRedis(payload *PayloadCommit) error {
	for retryCount := 0; ; {
		key := fmt.Sprintf(REDIS_KEY_PROJECT_PAYLOAD_CIDS, payload.ProjectId)
		res := redisClient.ZAdd(ctx, key,
			&redis.Z{
				Score:  float64(payload.TentativeBlockHeight),
				Member: payload.SnapshotCID,
			})
		if res.Err() != nil {
			if retryCount == *settingsObj.RetryCount {
				log.Errorf("Failed to Add payload %s to redis Zset with key %s after max-retries of %d", payload.SnapshotCID, key, *settingsObj.RetryCount)
				return res.Err()
			}
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			retryCount++
			log.Errorf("Failed to Add payload %s to redis Zset with key %s..retryCount %d", payload.SnapshotCID, key, retryCount)
			continue
		}
		log.Debugf("Added payload %s to redis Zset with key %s successfully", payload.SnapshotCID, key)
		break
	}
	return nil
}

func AddToPendingTxnsInRedis(payload *PayloadCommit, txHash string) error {
	key := fmt.Sprintf(REDIS_KEY_PROJECT_PENDING_TXNS, payload.ProjectId)
	var pendingtxn PendingTransaction
	pendingtxn.EventData.ProjectId = payload.ProjectId
	pendingtxn.EventData.SnapshotCid = payload.SnapshotCID
	pendingtxn.EventData.PayloadCommitId = payload.CommitId
	pendingtxn.EventData.Timestamp = float64(time.Now().Unix())
	pendingtxn.EventData.TentativeBlockHeight = payload.TentativeBlockHeight
	pendingtxn.EventData.ApiKeyHash = payload.ApiKeyHash
	pendingtxn.EventData.SkipAnchorProof = payload.SkipAnchorProof

	if payload.Resubmitted {
		pendingtxn.LastTouchedBlock = payload.ResubmissionBlock
	} else {
		pendingtxn.LastTouchedBlock = consts.SubmittedTxnStates.CallbackPending
	}
	pendingtxn.TxHash = txHash
	pendingtxn.EventData.TxHash = txHash
	pendingtxnBytes, err := json.Marshal(pendingtxn)
	if err != nil {
		log.Errorf("Failed to marshal pendingTxn %+v", pendingtxn)
		return err
	}
	for retryCount := 0; ; {
		var zAddArgs redis.ZAddArgs
		zAddArgs.GT = true
		zAddArgs.Members = append(zAddArgs.Members, redis.Z{
			Score:  float64(payload.TentativeBlockHeight),
			Member: pendingtxnBytes,
		})
		res := redisClient.ZAddArgs(ctx, key,
			zAddArgs,
		)
		if res.Err() != nil {
			if retryCount == *settingsObj.RetryCount {
				log.Errorf("Failed to add payloadCid %s for project %s with commitID %s to pendingTxns in redis with err %+v after max retries of %d",
					payload.SnapshotCID, payload.ProjectId, payload.CommitId, res.Err(), *settingsObj.RetryCount)
				return res.Err()
			}
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			retryCount++
			log.Errorf("Failed to add payloadCid %s for project %s with commitID %s to pendingTxns in redis with err %+v ..retryCount %d",
				payload.SnapshotCID, payload.ProjectId, payload.CommitId, res.Err(), retryCount)
			continue
		}
		log.Debugf("Added payloadCid %s for project %s with commitID %s to pendingTxns redis Zset with key %s successfully",
			payload.SnapshotCID, payload.ProjectId, payload.CommitId, key)
		break
	}
	return nil
}

func GenerateTokenHash(payload *PayloadCommit) (string, bool) {
	bn := make([]byte, 32)
	ba := make([]byte, 32)
	bm := make([]byte, 32)
	bn[31] = 0
	ba[31] = 0
	bm[31] = 0
	var snapshot Snapshot
	snapshot.Cid = payload.SnapshotCID

	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		log.Errorf("CRITICAL. Failed to Json-Marshall snapshot %v for project %s with commitID %s , with err %v",
			snapshot, payload.ProjectId, payload.CommitId, err)
		return "", false
	}
	tokenHash := crypto.Keccak256Hash(snapshotBytes, bn, ba, bm).String()
	return tokenHash, true
}

func PrepareAndSubmitTxnToChain(payload *PayloadCommit) (string, retryType) {
	var tokenHash, txHash string
	var status bool
	var err error

	if payload.ApiKeyHash == "" {
		tokenHash, status = GenerateTokenHash(payload)
		if !status {
			return "", NO_RETRY_FAILURE
		}
	} else {
		tokenHash = payload.ApiKeyHash
	}
	log.Tracef("Token hash generated for payload %+v for project %s with commitID %s  is : ",
		*payload, payload.ProjectId, payload.CommitId, tokenHash)

	var retryType retryType
	for retryCount := 0; ; {
		txHash, retryType, err = SubmitTxnToChain(payload, tokenHash)
		if err != nil {
			if retryType == NO_RETRY_FAILURE {
				return txHash, retryType
			} else if retryType == RETRY_WITH_DELAY {
				time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			}
			if retryCount == *settingsObj.RetryCount {
				log.Errorf("Failed to send txn for snapshot %s for project %s with commitID %s to tx-manager with err %+v after max retries of %d",
					payload.SnapshotCID, payload.ProjectId, payload.CommitId, err, *settingsObj.RetryCount)
				time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
				return txHash, RETRY_IMMEDIATE
			}
			retryCount++
			log.Errorf("Failed to send txn for snapshot %s for project %s with commitID %s to pendingTxns to tx-manager with err %+v ..retryCount %d",
				payload.SnapshotCID, payload.ProjectId, payload.CommitId, err, retryCount)
			continue
		}
		break
	}

	return txHash, NO_RETRY_SUCCESS
}

func UpdateTentativeBlockHeight(payload *PayloadCommit) error {
	key := fmt.Sprintf(REDIS_KEY_PROJECT_TENTATIVE_BLOCK_HEIGHT, payload.ProjectId)
	for retryCount := 0; ; retryCount++ {
		res := redisClient.Set(ctx, key, strconv.Itoa(payload.TentativeBlockHeight), 0)
		if res.Err() != nil {
			if retryCount > *settingsObj.RetryCount {
				return res.Err()
			}
			log.Errorf("Failed to update tentativeBlockHeight for project %s with commitId %s due to error %+v, retrying",
				payload.ProjectId, payload.CommitId, res.Err())
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			continue
		}
		break
	}
	return nil
}

func GetTentativeBlockHeight(projectId string) (int, error) {
	key := fmt.Sprintf(REDIS_KEY_PROJECT_TENTATIVE_BLOCK_HEIGHT, projectId)
	tentativeBlockHeight := 0
	var err error
	for retryCount := 0; ; retryCount++ {
		res := redisClient.Get(ctx, key)
		if res.Err() != nil {
			if res.Err() == redis.Nil {
				log.Infof("TentativeBlockHeight key is not present ")
				return tentativeBlockHeight, nil
			}
			if retryCount > *settingsObj.RetryCount {
				return tentativeBlockHeight, res.Err()
			}
			log.Errorf("Failed to fetch tentativeBlockHeight for project %s", projectId)
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			continue
		}
		tentativeBlockHeight, err = strconv.Atoi(res.Val())
		if err != nil {
			log.Fatalf("TentativeBlockHeight Corrupted for project %s with err %+v", projectId, err)
			return tentativeBlockHeight, err
		}
		break
	}
	log.Debugf("TenativeBlockHeight for project %s is %d", projectId, tentativeBlockHeight)
	return tentativeBlockHeight, nil
}

//TODO: Optimize code for all HTTP client's to reuse retry logic like tenacity retry of Python.
//As of now it is copy pasted and looks ugly.
func InvokeDAGFinalizerCallback(payload *PayloadCommit) retryType {
	reqURL := fmt.Sprintf("http://%s:%d/", settingsObj.WebhookListener.Host, settingsObj.WebhookListener.Port)
	var req AuditContractSimWebhookCallbackRequest
	req.EventName = "RecordAppended"
	req.Type = "event"
	req.Ctime = time.Now().Unix()
	req.TxHash = payload.ApiKeyHash
	req.EventData.ApiKeyHash = payload.ApiKeyHash
	req.EventData.PayloadCommitId = payload.CommitId
	req.EventData.ProjectId = payload.ProjectId
	req.EventData.SnapshotCid = payload.SnapshotCID
	req.EventData.TentativeBlockHeight = payload.TentativeBlockHeight
	req.EventData.Timestamp = time.Now().Unix()
	reqParams, err := json.Marshal(req)

	if err != nil {
		log.Fatalf("CRITICAL. Failed to Json-Marshall DAG finalizer request for project %s with commitID %s , with err %v",
			payload.ProjectId, payload.CommitId, err)
		return NO_RETRY_FAILURE
	}
	for retryCount := 0; ; {
		if retryCount == *settingsObj.RetryCount {
			log.Errorf("Webhook invocation failed for snapshot %s project %s with commitId %s after max-retry of %d",
				payload.SnapshotCID, payload.ProjectId, payload.CommitId, *settingsObj.RetryCount)
			return RETRY_IMMEDIATE
		}
		req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(reqParams))
		if err != nil {
			log.Fatalf("Failed to create new HTTP Req with URL %s for message %+v with error %+v",
				reqURL, commonTxReqParams, err)
			return RETRY_IMMEDIATE
		}
		req.Header.Add("Authorization", "Bearer "+settingsObj.Web3Storage.APIToken)
		req.Header.Add("accept", "application/json")
		req.Header.Add("Content-Type", "application/json")

		err = dagFinalizerClientRateLimiter.Wait(context.Background())
		if err != nil {
			log.Errorf("WebhookClient Rate Limiter wait timeout with error %+v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.Debugf("Sending Req to DAG finalizer URL %s for project %s with snapshotCID %s commitId %s ",
			reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId)
		res, err := dagFinalizerClient.Do(req)
		if err != nil {
			retryCount++
			log.Errorf("Failed to send request %+v towards DAG finalizer URL %s for project %s with snapshotCID %s commitId %s with error %+v.  Retrying %d",
				req, reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
			continue
		}
		defer res.Body.Close()
		var resp map[string]json.RawMessage
		respBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			retryCount++
			log.Errorf("Failed to read response body for project %s with snapshotCID %s commitId %s from DAG finalizer with error %+v. Retrying %d",
				payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
			continue
		}
		if res.StatusCode == http.StatusOK {
			err = json.Unmarshal(respBody, &resp)
			if err != nil {
				retryCount++
				log.Errorf("Failed to unmarshal response %+v for project %s with snapshotCID %s commitId %s from DAG finalizer with error %+v. Retrying %d",
					respBody, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
				time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
				continue
			}
			log.Debugf("Received 200 OK with body %+v from DAG finalizer for project %s with snapshotCID %s commitId %s ",
				resp, payload.ProjectId, payload.SnapshotCID, payload.CommitId)
			return NO_RETRY_SUCCESS
		} else {
			retryCount++
			log.Errorf("Received Error response %+v from DAG finalizer for project %s with commitId %s with statusCode %d and status : %s ",
				resp, payload.ProjectId, payload.CommitId, res.StatusCode, res.Status)
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			continue
		}
	}
}

func UploadToWeb3Storage(payload *PayloadCommit) (string, bool) {

	reqURL := settingsObj.Web3Storage.URL + settingsObj.Web3Storage.UploadURLSuffix
	for retryCount := 0; ; {
		if retryCount == *settingsObj.RetryCount {
			log.Errorf("web3.storage upload failed for snapshot %s project %s with commitId %s after max-retry of %d",
				payload.SnapshotCID, payload.ProjectId, payload.CommitId, *settingsObj.RetryCount)
			return "", false
		}
		req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(payload.Payload))
		if err != nil {
			log.Fatalf("Failed to create new HTTP Req with URL %s for snapshot %s project %s with commitId %s with error %+v",
				reqURL, payload.SnapshotCID, payload.ProjectId, payload.CommitId, err)
			return "", false
		}
		req.Header.Add("Authorization", "Bearer "+settingsObj.Web3Storage.APIToken)
		req.Header.Add("accept", "application/json")

		err = web3StorageClientRateLimiter.Wait(context.Background())
		if err != nil {
			log.Errorf("Web3Storage Rate Limiter wait timeout with error %+v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.Debugf("Sending Req to web3.storage URL %s for project %s with snapshotCID %s commitId %s ",
			reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId)
		res, err := w3sHttpClient.Do(req)
		if err != nil {
			retryCount++
			log.Errorf("Failed to send request %+v towards web3.storage URL %s for project %s with snapshotCID %s commitId %s with error %+v.  Retrying %d",
				req, reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
			continue
		}
		defer res.Body.Close()
		var resp Web3StoragePutResponse
		respBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			retryCount++
			log.Errorf("Failed to read response body for project %s with snapshotCID %s commitId %s from web3.storage with error %+v. Retrying %d",
				payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
			continue
		}
		if res.StatusCode == http.StatusOK {
			err = json.Unmarshal(respBody, &resp)
			if err != nil {
				retryCount++
				log.Errorf("Failed to unmarshal response %+v for project %s with snapshotCID %s commitId %s towards web3.storage with error %+v. Retrying %d",
					respBody, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
				time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
				continue
			}
			log.Debugf("Received 200 OK from web3.storage for project %s with snapshotCID %s commitId %s ",
				payload.ProjectId, resp.CID, payload.CommitId)
			return resp.CID, true
		} else {
			retryCount++
			var resp Web3StorageErrResponse
			err = json.Unmarshal(respBody, &resp)
			if err != nil {
				log.Errorf("Failed to unmarshal error response %+v for project %s with snapshotCID %s commitId %s towards web3.storage with error %+v. Retrying %d",
					respBody, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
			} else {
				log.Errorf("Received Error response %+v from web3.storage for project %s with commitId %s with statusCode %d and status : %s ",
					resp, payload.ProjectId, payload.CommitId, res.StatusCode, res.Status)
			}
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			continue
		}
	}
}

func SubmitTxnToChain(payload *PayloadCommit, tokenHash string) (txHash string, retry retryType, err error) {
	reqURL := settingsObj.ContractCallBackend
	var reqParams AuditContractCommitParams
	reqParams.ApiKeyHash = tokenHash
	reqParams.PayloadCommitId = payload.CommitId
	reqParams.ProjectId = payload.ProjectId
	reqParams.SnapshotCid = payload.SnapshotCID
	reqParams.TentativeBlockHeight = payload.TentativeBlockHeight
	commonTxReqParams.Params, err = json.Marshal(reqParams)
	if err != nil {
		log.Fatalf("Failed to marshal AuditContractCommitParams %+v towards tx-manager with error %+v", reqParams, err)
		return "", NO_RETRY_FAILURE, err
	}
	body, err := json.Marshal(commonTxReqParams)
	if err != nil {
		log.Fatalf("Failed to marshal request %+v towards tx-manager with error %+v", commonTxReqParams, err)
		return "", NO_RETRY_FAILURE, err
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Failed to create new HTTP Req with URL %s for message %+v with error %+v",
			reqURL, commonTxReqParams, err)
		return "", NO_RETRY_FAILURE, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("accept", "application/json")

	err = txClientRateLimiter.Wait(context.Background())
	if err != nil {
		log.Errorf("tx-manager Rate Limiter wait timeout with error %+v", err)
		time.Sleep(1 * time.Second)
		return "", RETRY_IMMEDIATE, err
	}

	log.Debugf("Sending Req with params %+v to tx-manager URL %s for project %s with snapshotCID %s commitId %s tokenHash %s.",
		reqParams, reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId, tokenHash)
	res, err := txMgrHttpClient.Do(req)
	if err != nil {
		log.Errorf("Failed to send request %+v towards tx-manager URL %s for project %s with snapshotCID %s commitId %s with error %+v",
			req, reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err)
		return "", RETRY_WITH_DELAY, err
	}
	defer res.Body.Close()
	var resp AuditContractCommitResp
	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Failed to read response body for project %s with snapshotCID %s commitId %s from tx-manager with error %+v",
			payload.ProjectId, payload.SnapshotCID, payload.CommitId, err)
		return "", RETRY_WITH_DELAY, err
	}
	if res.StatusCode == http.StatusOK {
		err = json.Unmarshal(respBody, &resp)
		if err != nil {
			log.Errorf("Failed to unmarshal response %+v for project %s with snapshotCID %s commitId %s towards tx-manager with error %+v",
				respBody, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err)
			return "", RETRY_WITH_DELAY, err
		}
		if resp.Success {
			log.Debugf("Received Success response %+v from tx-manager for project %s with snapshotCID %s commitId %s.",
				resp, payload.ProjectId, payload.SnapshotCID, payload.CommitId)
			return resp.Data[0].TxHash, NO_RETRY_SUCCESS, nil
		} else {
			var tmpRsp map[string]string
			_ = json.Unmarshal(respBody, &tmpRsp)
			log.Errorf("Received 200 OK with Error response %+v from tx-manager for project %s with snapshotCID %s commitId %s resp bytes %+v ",
				resp, payload.ProjectId, payload.SnapshotCID, payload.CommitId, tmpRsp)
			return "", RETRY_WITH_DELAY, errors.New("Received Error response from tx-manager : " + fmt.Sprintf("%+v", resp))
		}
	} else {
		log.Errorf("Received Error response %+v from tx-manager for project %s with snapshotCID %s commitId %s with statusCode %d and status : %s ",
			resp, payload.ProjectId, payload.SnapshotCID, payload.CommitId, res.StatusCode, res.Status)
		return "", RETRY_WITH_DELAY, errors.New("Received Error response from tx-manager" + fmt.Sprint(respBody))
	}
}

func InitTxManagerClient() {

	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        settingsObj.PayloadCommitConcurrency,
		MaxConnsPerHost:     settingsObj.PayloadCommitConcurrency,
		MaxIdleConnsPerHost: settingsObj.PayloadCommitConcurrency,
		IdleConnTimeout:     0,
		DisableCompression:  true,
	}

	txMgrHttpClient = http.Client{
		Timeout:   time.Duration(settingsObj.HttpClientTimeoutSecs) * time.Second,
		Transport: &t,
	}

	commonTxReqParams.Contract = strings.ToLower(settingsObj.AuditContract)
	commonTxReqParams.Method = "commitRecord"
	commonTxReqParams.NetworkId = 137
	commonTxReqParams.HackerMan = false
	commonTxReqParams.IgnoreGasEstimate = false

	//Default values
	tps := rate.Limit(50) //50 TPS
	burst := 50
	if settingsObj.ContractRateLimiter != nil {
		burst = settingsObj.ContractRateLimiter.Burst
		if settingsObj.ContractRateLimiter.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(settingsObj.ContractRateLimiter.RequestsPerSec)
		}
	}
	log.Infof("Rate Limit configured for tx-manager Client at %v TPS with a burst of %d", tps, burst)
	txClientRateLimiter = rate.NewLimiter(tps, burst)
}

func InitRedisClient() {
	redisURL := fmt.Sprintf("%s:%d", settingsObj.Redis.Host, settingsObj.Redis.Port)
	redisDb := settingsObj.Redis.Db
	log.Infof("Connecting to redis DB %d at %s", redisDb, redisURL)
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: settingsObj.Redis.Password,
		DB:       redisDb,
	})
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Errorf("Unable to connect to redis at %s with error %+v", redisURL, err)
	}
	log.Info("Connected successfully to Redis and received ", pong, " back")
}

func InitIPFSClient() {
	url := settingsObj.IpfsURL
	// Convert the URL from /ip4/<IPAddress>/tcp/<Port> to IP:Port format.
	connectUrl := strings.Split(url, "/")[2] + ":" + strings.Split(url, "/")[4]

	log.Infof("Initializing the IPFS client with IPFS Daemon URL %s.", connectUrl)
	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        settingsObj.PayloadCommitConcurrency,
		MaxConnsPerHost:     settingsObj.PayloadCommitConcurrency,
		MaxIdleConnsPerHost: settingsObj.PayloadCommitConcurrency,
		IdleConnTimeout:     0,
		DisableCompression:  true,
	}

	ipfsHttpClient := http.Client{
		Timeout:   time.Duration(settingsObj.IpfsTimeout * 1000000000),
		Transport: &t,
	}
	log.Debugf("Setting IPFS HTTP client timeout as %f seconds", ipfsHttpClient.Timeout.Seconds())
	ipfsClient = shell.NewShellWithClient(connectUrl, &ipfsHttpClient)

	//Default values
	tps := rate.Limit(100) //50 TPS
	burst := 100
	if settingsObj.IPFSRateLimiter != nil {
		burst = settingsObj.IPFSRateLimiter.Burst
		if settingsObj.IPFSRateLimiter.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(settingsObj.IPFSRateLimiter.RequestsPerSec)
		}
	}
	log.Infof("Rate Limit configured for IPFS Client at %v TPS with a burst of %d", tps, burst)
	ipfsClientRateLimiter = rate.NewLimiter(tps, burst)

}
func InitW3sClient() {
	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        settingsObj.Web3Storage.MaxIdleConns,
		MaxConnsPerHost:     settingsObj.Web3Storage.MaxIdleConns,
		MaxIdleConnsPerHost: settingsObj.Web3Storage.MaxIdleConns,
		IdleConnTimeout:     time.Duration(settingsObj.Web3Storage.IdleConnTimeout),
		DisableCompression:  true,
	}

	w3sHttpClient = http.Client{
		Timeout:   time.Duration(settingsObj.Web3Storage.TimeoutSecs) * time.Second,
		Transport: &t,
	}

	//Default values
	tps := rate.Limit(3) //3 TPS
	burst := 3
	if settingsObj.Web3Storage.RateLimiter != nil {
		burst = settingsObj.Web3Storage.RateLimiter.Burst
		if settingsObj.Web3Storage.RateLimiter.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(settingsObj.Web3Storage.RateLimiter.RequestsPerSec)
		}
	}
	log.Infof("Rate Limit configured for web3.storage at %v TPS with a burst of %d", tps, burst)
	web3StorageClientRateLimiter = rate.NewLimiter(tps, burst)
}

func InitDAGFinalizerCallbackClient() {
	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        settingsObj.PayloadCommitConcurrency,
		MaxConnsPerHost:     settingsObj.PayloadCommitConcurrency,
		MaxIdleConnsPerHost: settingsObj.PayloadCommitConcurrency,
		IdleConnTimeout:     0,
		DisableCompression:  true,
	}

	dagFinalizerClient = http.Client{
		Timeout:   time.Duration(settingsObj.HttpClientTimeoutSecs) * time.Second,
		Transport: &t,
	}
	//Default values
	tps := rate.Limit(50) //50 TPS
	burst := 20
	if settingsObj.WebhookListener.RateLimiter != nil {
		burst = settingsObj.WebhookListener.RateLimiter.Burst
		if settingsObj.WebhookListener.RateLimiter.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(settingsObj.WebhookListener.RateLimiter.RequestsPerSec)
		}
	}
	log.Infof("Rate Limit configured for dagFinalizerClient at %v TPS with a burst of %d", tps, burst)
	dagFinalizerClientRateLimiter = rate.NewLimiter(tps, burst)
}
