package main

import (
	log "github.com/sirupsen/logrus"

	"audit-protocol/caching"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/logger"
	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/smartcontract"
	taskmgr "audit-protocol/goutils/taskmgr/rabbitmq"
	w3storage "audit-protocol/goutils/w3s"
	"audit-protocol/payload-commit/service"
	"audit-protocol/payload-commit/worker"
)

func main() {
	logger.InitLogger()
	settingsObj := settings.ParseSettings()

	ipfsutils.InitClient(
		settingsObj.IpfsConfig.URL,
		settingsObj.PayloadCommit.Concurrency,
		settingsObj.IpfsConfig.IPFSRateLimiter,
		settingsObj.IpfsConfig.Timeout,
	)

	redisClient := redisutils.InitRedisClient(
		settingsObj.Redis.Host,
		settingsObj.Redis.Port,
		settingsObj.Redis.Db,
		settingsObj.DagVerifierSettings.RedisPoolSize,
		settingsObj.Redis.Password,
		-1,
	)

	caching.NewRedisCache()
	smartcontract.InitContractAPI()
	taskmgr.NewRabbitmqTaskMgr()
	w3storage.InitW3S()

	service.InitPayloadCommitService()

	mqWorker := worker.NewWorker()

	defer func() {
		mqWorker.ShutdownWorker()
		err := redisClient.Close()
		if err != nil {
			log.WithError(err).Error("error while closing redis client")
		}
	}()

	for {
		err := mqWorker.ConsumeTask()
		if err != nil {
			log.WithError(err).Error("error while consuming task, starting again")
		}
	}

	// InitTxManagerClient()
	// InitDAGFinalizerCallbackClient()
	// InitW3sClient()
	// var wg sync.WaitGroup
	// if settingsObj.UseConsensus {
	// 	wg.Add(1)
	// 	InitConsensusClient()
	// 	WaitQueueForConsensus = make(map[string]*datamodel.PayloadCommit, 100)
	// 	go func() {
	// 		defer wg.Done()
	// 		PollConsensusForConfirmations()
	// 	}()
	// }
	//
	// log.Info("Starting RabbitMq Consumer")
	// RegisterSignalHandles()
	//
	// InitRabbitmqConsumer()
	// if settingsObj.UseConsensus {
	// 	wg.Wait()
	// }
}

// const (
// 	NO_RETRY_SUCCESS retryType = iota
// 	RETRY_IMMEDIATE            // TO be used in timeout scenarios or non server returned error scenarios.
// 	RETRY_WITH_DELAY           // TO be used when immediate error is returned so that server is not overloaded.
// 	NO_RETRY_FAILURE           // This is to be used for unexpected conditions which are not recoverable and hence no retry
// )

// func (r retryType) String() string {
// 	switch r {
// 	case NO_RETRY_SUCCESS:
// 		return "NO Retry"
// 	case RETRY_IMMEDIATE:
// 		return "Retry Immediately"
// 	case RETRY_WITH_DELAY:
// 		return "Retry with a delay"
// 	case NO_RETRY_FAILURE:
// 		return "Non recoverable error, hence not retrying."
// 	}
// 	return "unknown"
// }

// func RegisterSignalHandles() {
// 	log.Info("Setting up signal Handlers")
// 	signalChanel := make(chan os.Signal, 1)
// 	signal.Notify(signalChanel,
// 		syscall.SIGHUP,
// 		syscall.SIGINT,
// 		syscall.SIGTERM,
// 		syscall.SIGQUIT)
//
// 	go func() {
// 		for {
// 			s := <-signalChanel
// 			switch s {
// 			// kill -SIGHUP XXXX [XXXX - PID for your program]
// 			case syscall.SIGHUP:
// 				log.Info("Signal hang up triggered.")
//
// 				// kill -SIGINT XXXX or Ctrl+c  [XXXX - PID for your program]
// 			case syscall.SIGINT:
// 				log.Info("Signal interrupt triggered.")
// 				GracefulShutDown()
//
// 				// kill -SIGTERM XXXX [XXXX - PID for your program]
// 			case syscall.SIGTERM:
// 				log.Info("Signal terminte triggered.")
// 				GracefulShutDown()
//
// 				// kill -SIGQUIT XXXX [XXXX - PID for your program]
// 			case syscall.SIGQUIT:
// 				log.Info("Signal quit triggered.")
// 				GracefulShutDown()
//
// 			default:
// 				log.Info("Unknown signal.", s)
// 				GracefulShutDown()
// 			}
// 		}
// 	}()
//
// }

// func GracefulShutDown() {
// 	// rmqConnection.StopConsumer()
// 	time.Sleep(2 * time.Second)
// 	redisClient.Close()
// 	exitChan <- true
// }
//
// type SubmittedTransactionStates struct {
// 	CallbackPending  int `json:"callbackPending"`
// 	CallbackReceived int `json:"callbackReceived"`
// }
// type ConstsObj struct {
// 	SubmittedTxnStates SubmittedTransactionStates `json:"submittedTxnStates"`
// }
//
// func ParseSettings() {
// 	settingsObj = settings.ParseSettings()
// 	if settingsObj.UseConsensus && settingsObj.InstanceId == "" {
// 		log.Fatalf("InstanceID is set to null, please generate and set a unique instanceID")
// 		os.Exit(1)
// 	}
//
// 	ParseConsts(os.Getenv("CONFIG_PATH") + "/dev_consts.json")
// }
//
// func ParseConsts(constsFile string) {
// 	log.Info("Reading Consts File:", constsFile)
// 	data, err := os.ReadFile(constsFile)
// 	if err != nil {
// 		log.Error("Cannot read the file:", err)
// 		panic(err)
// 	}
//
// 	log.Debug("Consts json data is", string(data))
// 	err = json.Unmarshal(data, &consts)
// 	if err != nil {
// 		log.Error("Cannot unmarshal the Consts json ", err)
// 		panic(err)
// 	}
// }
//
// func InitRabbitmqConsumer() {
// 	// var err error
// 	// rabbitMqURL := fmt.Sprintf("amqp://%s:%s@%s:%d/", settingsObj.Rabbitmq.User, settingsObj.Rabbitmq.Password, settingsObj.Rabbitmq.Host, settingsObj.Rabbitmq.Port)
// 	// // rmqConnection, err = GetConn(rabbitMqURL)
// 	// log.Infof("Starting rabbitMQ consumer connecting to URL: %s with concurreny %d", rabbitMqURL, settingsObj.PayloadCommit.Concurrency)
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	// rmqExchangeName := settingsObj.Rabbitmq.Setup.Core.Exchange
// 	// rmqQueueName := settingsObj.Rabbitmq.Setup.Queues.CommitPayloads.QueueNamePrefix + settingsObj.InstanceId
// 	// rmqRoutingKey := settingsObj.Rabbitmq.Setup.Queues.CommitPayloads.RoutingKeyPrefix + settingsObj.InstanceId
// 	//
// 	// err = rmqConnection.StartConsumer(rmqQueueName,
// 	// 	rmqExchangeName,
// 	// 	rmqRoutingKey,
// 	// 	RabbitmqMsgHandler,
// 	// 	settingsObj.PayloadCommit.Concurrency)
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	//
// 	// exitChan = make(chan bool)
// 	// <-exitChan
// }
//
// func AssignTentativeHeight(payloadCommit *datamodel.PayloadCommit) int {
// 	tentativeBlockHeight := 0
// 	firstEpochEndHeight, epochSize := GetFirstEpochDetails(payloadCommit)
// 	if epochSize == 0 || firstEpochEndHeight == 0 {
// 		tentativeBlockHeight = 0
// 	} else {
// 		tentativeBlockHeight = (payloadCommit.SourceChainDetails.EpochEndHeight-firstEpochEndHeight)/epochSize + 1
// 		log.Debugf("Assigning tentativeBlockHeight as %d for project %s and payloadCommitID %s",
// 			tentativeBlockHeight, payloadCommit.ProjectId, payloadCommit.CommitId)
// 	}
// 	return tentativeBlockHeight
// }
//
// func GetFirstEpochDetails(payloadCommit *datamodel.PayloadCommit) (int, int) {
//
// 	// Record Project epoch size for the first epoch and also the endHeight.
// 	epochSize := FetchProjectEpochSize(payloadCommit.ProjectId)
// 	firstEpochEndHeight := 0
// 	if epochSize == 0 {
// 		epochSize = payloadCommit.SourceChainDetails.EpochEndHeight - payloadCommit.SourceChainDetails.EpochStartHeight + 1
// 		status := SetProjectEpochSize(payloadCommit.ProjectId, epochSize)
// 		if !status {
// 			return 0, 0
// 		}
// 		firstEpochEndHeight := payloadCommit.SourceChainDetails.EpochEndHeight
// 		status = SetFirstEpochEndHeight(payloadCommit.ProjectId, firstEpochEndHeight)
// 		if !status {
// 			return 0, 0
// 		}
// 	}
// 	firstEpochEndHeight = FetchFirstEpochEndHeight(payloadCommit.ProjectId)
// 	log.Debugf("Fetched firstEpochEndHeight as %d and epochSize as %d from redis for project %s",
// 		firstEpochEndHeight, epochSize, payloadCommit.ProjectId)
// 	return firstEpochEndHeight, epochSize
// }
//
// func ProcessUnCommittedSnapshot(payloadCommit *datamodel.PayloadCommit) bool {
// 	tentativeBlockHeight := 0
// 	updateTentativeBlockHeightState := false
// 	// TODO: Need to think of a better way to do this in future,
// 	// probably some kind of projectLevel Data which indicates ordering requirements for the project.
// 	if payloadCommit.SourceChainDetails.EpochStartHeight == 0 || payloadCommit.SourceChainDetails.EpochEndHeight == 0 {
// 		/*In case of projects which need not be ordered based on sourceChainHeight such as SummaryProjects
// 		  Arrive at tentativeHeight by storing it in redis and assigning the next one*/
// 		lastTentativeBlockHeight, err := GetTentativeBlockHeight(payloadCommit.ProjectId)
// 		if err != nil {
// 			return false
// 		}
// 		tentativeBlockHeight = lastTentativeBlockHeight + 1
// 		updateTentativeBlockHeightState = true
// 		// TODO: For now taking this shortcut..but this should ideally be derived from projectState.
// 		payloadCommit.IsSummaryProject = true
// 	} else {
// 		tentativeBlockHeight = AssignTentativeHeight(payloadCommit)
// 		if tentativeBlockHeight == 0 {
// 			return false
// 		}
// 	}
// 	payloadCommit.TentativeBlockHeight = tentativeBlockHeight
// 	var ipfsStatus bool
// 	if payloadCommit.Web3Storage && settingsObj.Web3Storage.APIToken != "" {
// 		var wg sync.WaitGroup
// 		var w3sStatus bool
// 		log.Debugf("Received incoming Payload commit message at tentative DAG Height %d for project %s with commitId %s from rabbitmq. Uploading payload to web3.storage and IPFS.",
// 			payloadCommit.TentativeBlockHeight, payloadCommit.ProjectId, payloadCommit.CommitId)
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			ipfsStatus = ipfsClient.UploadSnapshotToIPFS(payloadCommit)
// 			if !ipfsStatus {
// 				return
// 			}
// 			log.Debugf("IPFS add Successful. Snapshot CID is %s for project %s with commitId %s at tentativeBlockHeight %d",
// 				payloadCommit.SnapshotCID, payloadCommit.ProjectId, payloadCommit.CommitId, payloadCommit.TentativeBlockHeight)
// 		}()
//
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			var snapshotCid string
// 			snapshotCid, w3sStatus = UploadToWeb3Storage(payloadCommit)
// 			if !w3sStatus {
// 				return
// 			}
// 			log.Debugf("web3.storage upload Successful. Snapshot CID is %s for project %s with commitId %s at tentativeBlockHeight %d",
// 				snapshotCid, payloadCommit.ProjectId, payloadCommit.CommitId, payloadCommit.TentativeBlockHeight)
// 			// SnapshotCID from web3.storage is not used directly.
// 			// payloadCommit.SnapshotCID = snapshotCid
// 		}()
//
// 		wg.Wait()
// 		if !ipfsStatus {
// 			log.Errorf("Failed to add to IPFS. IPFSStatus %b , web3.storage status %b", ipfsStatus, w3sStatus)
// 			return false
// 		}
// 		if !w3sStatus {
// 			// Since web3.storage is only used as a backup, we proceed even if web3.storage upload fails as IPFS upload is successful
// 			log.Errorf("Failed to upload to web3.storage status %b. Continuing with processing.", w3sStatus)
// 		}
// 	} else {
// 		log.Debugf("Received incoming Payload commit message at tentative DAG Height %d for project %s with commitId %s from rabbitmq. Adding payload to IPFS.",
// 			payloadCommit.TentativeBlockHeight, payloadCommit.ProjectId, payloadCommit.CommitId)
// 		ipfsStatus = ipfsClient.UploadSnapshotToIPFS(payloadCommit,
// 			settingsObj.RetryIntervalSecs, *settingsObj.RetryCount)
// 		if !ipfsStatus {
// 			return false
// 		}
// 	}
// 	err := filecache.StorePayloadToCache(settingsObj.PayloadCachePath, payloadCommit.ProjectId, payloadCommit.SnapshotCID, payloadCommit.Payload)
// 	if err != nil {
// 		log.Errorf("Failed to store payload in cache for the project %s with commitId %s due to error %+v",
// 			payloadCommit.ProjectId, payloadCommit.CommitId, err)
// 	}
// 	err = redisutils.AddPayloadCidToZSet(ctx, redisClient, payloadCommit,
// 		settingsObj.RetryIntervalSecs, *settingsObj.RetryCount)
// 	if err != nil {
// 		log.Errorf("Failed to store payloadCid in redis for the project %s with commitId %s due to error %+v",
// 			payloadCommit.ProjectId, payloadCommit.CommitId, err)
// 		return false
// 	}
// 	if updateTentativeBlockHeightState {
// 		// Update TentativeBlockHeight for the project
// 		err = UpdateTentativeBlockHeight(payloadCommit)
// 		if err != nil {
// 			log.Errorf("Failed to update tentativeBlockHeight for the project %s with commitId %s due to error %+v",
// 				payloadCommit.ProjectId, payloadCommit.CommitId, err)
// 			return false
// 		}
// 	}
// 	return true
// }
//
// func SetProjectEpochSize(projectId string, epochSize int) bool {
// 	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_EPOCH_SIZE, projectId)
// 	log.Debugf("Setting  epoch size as %d for project %s", epochSize, projectId)
//
// 	return redisutils.SetIntFieldInRedis(
// 		ctx, redisClient, key, epochSize,
// 		settingsObj.RetryIntervalSecs, *settingsObj.RetryCount)
// }
//
// func SetFirstEpochEndHeight(projectId string, epochEndHeight int) bool {
// 	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_FIRST_EPOCH_END_HEIGHT, projectId)
// 	log.Debugf("Setting First epoch endHeight as %d for project %s", epochEndHeight, projectId)
//
// 	return redisutils.SetIntFieldInRedis(
// 		ctx, redisClient, key, epochEndHeight,
// 		settingsObj.RetryIntervalSecs, *settingsObj.RetryCount)
// }
//
// func FetchProjectEpochSize(projectID string) int {
// 	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_EPOCH_SIZE, projectID)
//
// 	epochSize := redisutils.FetchIntFieldFromRedis(
// 		ctx, redisClient, key,
// 		settingsObj.RetryIntervalSecs, *settingsObj.RetryCount)
//
// 	return epochSize
// }
//
// func FetchFirstEpochEndHeight(projectID string) int {
// 	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_FIRST_EPOCH_END_HEIGHT, projectID)
//
// 	firstEpochEndHeight := redisutils.FetchIntFieldFromRedis(
// 		ctx, redisClient, key,
// 		settingsObj.RetryIntervalSecs, *settingsObj.RetryCount)
//
// 	return firstEpochEndHeight
// }
//
// func RabbitmqMsgHandler(d amqp.Delivery) bool {
// 	if d.Body == nil {
// 		log.Errorf("Received message %+v from rabbitmq without message body! Ignoring and not processing it.", d)
// 		return true
// 	}
// 	log.Tracef("Received Message from rabbitmq %v", d.Body)
//
// 	var payloadCommit datamodel.PayloadCommit
//
// 	err := json.Unmarshal(d.Body, &payloadCommit)
// 	if err != nil {
// 		log.Warnf("CRITICAL: Json unmarshal failed for payloadCommit %v, with err %v. Ignoring", d.Body, err)
// 		return true
// 	}
//
// 	if !payloadCommit.Resubmitted {
// 		if d.Redelivered {
// 			log.Warnf("Message got redelivered from rabbitmq for project %s and commitId %s",
// 				payloadCommit.ProjectId, payloadCommit.CommitId)
// 		}
// 		if !ProcessUnCommittedSnapshot(&payloadCommit) {
// 			return true
// 		}
// 	} else {
// 		if payloadCommit.SnapshotCID == "" && payloadCommit.Payload == nil {
// 			log.Warnf("Received incoming Payload commit message without snapshotCID and empty payload at tentative DAG Height %d for project %s for resubmission at block %d from rabbitmq. Discarding this message without processing.",
// 				payloadCommit.TentativeBlockHeight, payloadCommit.ProjectId, payloadCommit.ResubmissionBlock)
// 			return true
// 		} else {
// 			// What if payload is present and snapshotCID is empty?? Currently there is no scenario where this can happen, but need to handle in future.
// 			// This would require soem kind of reorg of DAGChain if required as this is a resubmission of payload already submitted.
// 			log.Debugf("Received incoming Payload commit message at tentative DAG Height %d for project %s for resubmission at block %d from rabbitmq.",
// 				payloadCommit.TentativeBlockHeight, payloadCommit.ProjectId, payloadCommit.ResubmissionBlock)
// 		}
// 	}
//
// 	var requestID, txHash string
// 	var retryType retryType
// 	if !payloadCommit.SkipAnchorProof {
// 		requestID, txHash, retryType = PrepareAndSubmitTxnToChain(&payloadCommit)
// 		if retryType == RETRY_IMMEDIATE || retryType == RETRY_WITH_DELAY {
// 			// Not retrying further, expecting sel-healing to take care of recovery
// 			log.Warnf("MAX Retries reached while trying to invoke tx-manager for project %s and commitId %s with tentativeBlockHeight %d.",
// 				payloadCommit.ProjectId, payloadCommit.CommitId, payloadCommit.TentativeBlockHeight, " Not retrying further.")
// 			requestID = ""
// 		} else if retryType == NO_RETRY_SUCCESS {
// 			log.Trace("Submitted txn to chain for %+v", payloadCommit)
// 		} else {
// 			log.Errorf("Irrecoverable error occurred and hence ignoring snapshots for processing.")
// 			return true
// 		}
// 	} else {
// 		requestID = uuid.New().String()
// 		payloadCommit.RequestID = requestID
// 		// Wait for consensus.
// 		// Skip consensus in case of summmaryProject until aggregation logic is fixed.
// 		if settingsObj.UseConsensus && !payloadCommit.IsSummaryProject && !payloadCommit.Resubmitted {
// 			// In case of resubmission, no need to go for consensus again.
// 			// Not storing this queue to redis, because in a worst-case scenario
// 			// if the process crashes and we loose this state information, self-healing shall take care of it.
// 			status, err := SubmitSnapshotForConsensus(&payloadCommit)
// 			if status == SNAPSHOT_CONSENSUS_STATUS_ACCEPTED {
// 				QueueLock.Lock()
// 				WaitQueueForConsensus[payloadCommit.ProjectId] = &payloadCommit
// 				QueueLock.Unlock()
// 				return true
// 			} else if status == "" { // This check is added as protection, ideally this condition should not be hit.
// 				log.Fatalf("Snapshot is not accepted for project %s due to error %+v", payloadCommit.ProjectId, err)
// 				return true
// 			}
// 			if err != nil {
// 				if status == "" {
// 					log.Fatalf("Snapshot is not accepted for project %s due to error %+v", payloadCommit.ProjectId, err)
// 					return true
// 				}
// 				return true
// 			}
// 		}
// 	}
// 	AddToPendingTxns(&payloadCommit, txHash, requestID)
// 	return true
// }
//
// func AddToPendingTxns(payloadCommit *datamodel.PayloadCommit, txHash string, requestID string) bool {
// 	/*Add to redis pendingTransactions*/
// 	err := UpdatePendingTxnInRedis(payloadCommit, requestID, txHash)
// 	if err != nil {
// 		// Not retrying further..expecting self-healing to take care.
// 		log.Errorf("Unable to add transaction to PendingTxns after max retries for project %s and commitId %s with tentativeBlockHeight %d.",
// 			payloadCommit.ProjectId, payloadCommit.CommitId, payloadCommit.TentativeBlockHeight, " Not retrying further as this could be a system level issue!")
// 		return true
// 	}
// 	if payloadCommit.SkipAnchorProof {
// 		// Notify DAG finalizer service as we are skipping proof anchor on chain.
// 		retryType := InvokeDAGFinalizerCallback(payloadCommit, requestID)
// 		if retryType == RETRY_IMMEDIATE || retryType == RETRY_WITH_DELAY {
// 			// Not retrying further..expecting self-healing to take care.
// 			log.Warnf("MAX Retries reached while trying to invoke DAG finalizer for project %s and commitId %s with tentativeBlockHeight %d.",
// 				payloadCommit.ProjectId, payloadCommit.CommitId, payloadCommit.TentativeBlockHeight, " Not retrying further as this could be a system level issue!")
// 			return true
// 		} else if retryType == NO_RETRY_SUCCESS {
// 			log.Trace("Submitted txn to chain for %+v", payloadCommit)
// 		}
// 	}
// 	return true
// }
//
// func ReadPayloadFromCache(projectID string, payloadCid string) (*datamodel.DagPayload, error) {
// 	var payload datamodel.DagPayload
// 	log.Debugf("Fetching payloadCid %s from local Cache", payloadCid)
// 	bytes, err := filecache.ReadFromCache(settingsObj.PayloadCachePath+"/", projectID, payloadCid)
// 	if err != nil {
// 		log.Errorf("Failed to fetch payloadCid from local Cache, CID %s, due to error %+v ",
// 			payloadCid, err)
// 		return nil, err
// 	}
// 	err = json.Unmarshal(bytes, &payload)
// 	if err != nil {
// 		log.Errorf("Failed to Unmarshal Json Payload from local Cache, CID %s, bytes: %+v due to error %+v ",
// 			payloadCid, bytes, err)
// 		return nil, err
// 	}
// 	log.Debugf("Fetched Payload with CID %s from local cache: %+v", payloadCid, payload)
// 	return &payload, nil
// }
//
// func UpdatePendingTxnInRedis(payload *datamodel.PayloadCommit, requestID string, txHash string) error {
// 	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PENDING_TXNS, payload.ProjectId)
// 	var pendingtxn datamodel.PendingTransaction
// 	pendingtxn.EventData.ProjectId = payload.ProjectId
// 	pendingtxn.EventData.SnapshotCid = payload.SnapshotCID
// 	pendingtxn.EventData.PayloadCommitId = payload.CommitId
// 	pendingtxn.EventData.Timestamp = float64(time.Now().Unix())
// 	pendingtxn.EventData.TentativeBlockHeight = payload.TentativeBlockHeight
// 	pendingtxn.EventData.ApiKeyHash = payload.ApiKeyHash
// 	pendingtxn.EventData.SkipAnchorProof = payload.SkipAnchorProof
// 	pendingtxn.RequestID = requestID
//
// 	if payload.Resubmitted {
// 		pendingtxn.LastTouchedBlock = payload.ResubmissionBlock
// 		// Check if a pendingEntry Exists, if so remove the old one and then create the new entry
// 		res := redisClient.ZRangeByScore(ctx, key,
// 			&redis.ZRangeBy{
// 				Min: strconv.Itoa(payload.TentativeBlockHeight),
// 				Max: strconv.Itoa(payload.TentativeBlockHeight),
// 			})
// 		if res.Err() == nil {
// 			log.Debugf("Removing old pendingTxn entries %+v for project %s at tentativeHeight %d", res.Val(), payload.ProjectId, payload.TentativeBlockHeight)
// 			removeRes := redisClient.ZRemRangeByScore(ctx, key,
// 				strconv.Itoa(payload.TentativeBlockHeight),
// 				strconv.Itoa(payload.TentativeBlockHeight))
// 			if removeRes.Err() != nil {
// 				log.Warnf("Failed to remove pendingTxn entry %+v from redis due to error %+v", res.Val(), removeRes.Err())
// 			}
// 		} else {
// 			if res.Err() != redis.Nil {
// 				log.Warnf("Failed to fetch pendingTxns for payloadCid %s for project %s with commitID %s from redis with err %+v",
// 					payload.SnapshotCID, payload.ProjectId, payload.CommitId, res.Err(), *settingsObj.RetryCount)
// 			} else {
// 				log.Debugf("No pendingTxn entry present for project %s at tentativeHeight %d. Payload is %+v",
// 					payload.ProjectId, payload.TentativeBlockHeight, payload)
// 			}
// 		}
// 	} else {
// 		pendingtxn.LastTouchedBlock = consts.SubmittedTxnStates.CallbackPending
// 	}
// 	pendingtxn.TxHash = txHash
// 	pendingtxn.EventData.TxHash = txHash
// 	pendingtxnBytes, err := json.Marshal(pendingtxn)
// 	if err != nil {
// 		log.Errorf("Failed to marshal pendingTxn %+v", pendingtxn)
// 		return err
// 	}
// 	for retryCount := 0; ; {
// 		var zAddArgs redis.ZAddArgs
// 		zAddArgs.GT = true
// 		zAddArgs.Members = append(zAddArgs.Members, redis.Z{
// 			Score:  float64(payload.TentativeBlockHeight),
// 			Member: pendingtxnBytes,
// 		})
// 		res := redisClient.ZAddArgs(ctx, key,
// 			zAddArgs,
// 		)
// 		if res.Err() != nil {
// 			if retryCount == *settingsObj.RetryCount {
// 				log.Errorf("Failed to add payloadCid %s for project %s with commitID %s to pendingTxns in redis with err %+v after max retries of %d",
// 					payload.SnapshotCID, payload.ProjectId, payload.CommitId, res.Err(), *settingsObj.RetryCount)
// 				return res.Err()
// 			}
// 			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
// 			retryCount++
// 			log.Errorf("Failed to add payloadCid %s for project %s with commitID %s to pendingTxns in redis with err %+v ..retryCount %d",
// 				payload.SnapshotCID, payload.ProjectId, payload.CommitId, res.Err(), retryCount)
// 			continue
// 		}
// 		log.Debugf("Added payloadCid %s for project %s with commitID %s to pendingTxns redis Zset with key %s successfully",
// 			payload.SnapshotCID, payload.ProjectId, payload.CommitId, key)
// 		break
// 	}
// 	return nil
// }
//
// func GenerateTokenHash(payload *datamodel.PayloadCommit) (string, bool) {
// 	bn := make([]byte, 32)
// 	ba := make([]byte, 32)
// 	bm := make([]byte, 32)
// 	bn[31] = 0
// 	ba[31] = 0
// 	bm[31] = 0
// 	var snapshot datamodel.Snapshot
// 	snapshot.Cid = payload.SnapshotCID
//
// 	snapshotBytes, err := json.Marshal(snapshot)
// 	if err != nil {
// 		log.Errorf("CRITICAL. Failed to Json-Marshall snapshot %v for project %s with commitID %s , with err %v",
// 			snapshot, payload.ProjectId, payload.CommitId, err)
// 		return "", false
// 	}
// 	tokenHash := crypto.Keccak256Hash(snapshotBytes, bn, ba, bm).String()
// 	return tokenHash, true
// }
//
// func PrepareAndSubmitTxnToChain(payload *datamodel.PayloadCommit) (string, string, retryType) {
// 	var tokenHash, requestID, txHash string
// 	var status bool
// 	var err error
//
// 	if payload.ApiKeyHash == "" {
// 		tokenHash, status = GenerateTokenHash(payload)
// 		if !status {
// 			return "", "", NO_RETRY_FAILURE
// 		}
// 	} else {
// 		tokenHash = payload.ApiKeyHash
// 	}
// 	log.Tracef("Token hash generated for payload %+v for project %s with commitID %s  is : ",
// 		*payload, payload.ProjectId, payload.CommitId, tokenHash)
//
// 	var retryType retryType
// 	for retryCount := 0; ; {
// 		requestID, txHash, retryType, err = SubmitTxnToChain(payload, tokenHash)
// 		if err != nil {
// 			if retryType == NO_RETRY_FAILURE {
// 				return requestID, txHash, retryType
// 			} else if retryType == RETRY_WITH_DELAY {
// 				time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
// 			}
// 			if retryCount == *settingsObj.RetryCount {
// 				log.Errorf("Failed to send txn for snapshot %s for project %s with commitID %s to tx-manager with err %+v after max retries of %d",
// 					payload.SnapshotCID, payload.ProjectId, payload.CommitId, err, *settingsObj.RetryCount)
// 				time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
// 				return requestID, txHash, RETRY_IMMEDIATE
// 			}
// 			retryCount++
// 			log.Errorf("Failed to send txn for snapshot %s for project %s with commitID %s to pendingTxns to tx-manager with err %+v ..retryCount %d",
// 				payload.SnapshotCID, payload.ProjectId, payload.CommitId, err, retryCount)
// 			continue
// 		}
// 		break
// 	}
//
// 	return requestID, txHash, NO_RETRY_SUCCESS
// }
//
// func UpdateTentativeBlockHeight(payload *datamodel.PayloadCommit) error {
// 	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_TENTATIVE_BLOCK_HEIGHT, payload.ProjectId)
//
// 	status := redisutils.SetIntFieldInRedis(
// 		ctx, redisClient, key,
// 		payload.TentativeBlockHeight,
// 		settingsObj.RetryIntervalSecs, *settingsObj.RetryCount)
//
// 	if !status {
// 		return errors.New("failed t update tentativeBlockHeight in redis")
// 	}
// 	return nil
// }
//
// func GetTentativeBlockHeight(projectId string) (int, error) {
// 	tentativeBlockHeight := 0
// 	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_TENTATIVE_BLOCK_HEIGHT, projectId)
//
// 	retVal := redisutils.FetchIntFieldFromRedis(ctx, redisClient,
// 		key, settingsObj.RetryIntervalSecs, *settingsObj.RetryCount)
//
// 	if retVal == -1 {
// 		return tentativeBlockHeight, errors.New("failed to fetch tentativeBlockHeight from redis")
// 	}
// 	return retVal, nil
//
// }
//
// // TODO: Optimize code for all HTTP client's to reuse retry logic like tenacity retry of Python.
// // As of now it is copy pasted and looks ugly.
// func InvokeDAGFinalizerCallback(payload *datamodel.PayloadCommit, requestID string) retryType {
// 	reqURL := fmt.Sprintf("http://%s:%d/", settingsObj.DAGFinalizer.Host, settingsObj.DAGFinalizer.Port)
// 	var req datamodel.AuditContractSimWebhookCallbackRequest
// 	req.EventName = "RecordAppended"
// 	req.Type = "event"
// 	req.Ctime = time.Now().Unix()
// 	req.TxHash = payload.ApiKeyHash
// 	req.RequestID = requestID
// 	req.EventData.ApiKeyHash = payload.ApiKeyHash
// 	req.EventData.PayloadCommitId = payload.CommitId
// 	req.EventData.ProjectId = payload.ProjectId
// 	req.EventData.SnapshotCid = payload.SnapshotCID
// 	req.EventData.TentativeBlockHeight = payload.TentativeBlockHeight
// 	req.EventData.Timestamp = time.Now().Unix()
// 	reqParams, err := json.Marshal(req)
//
// 	if err != nil {
// 		log.Fatalf("CRITICAL. Failed to Json-Marshall DAG finalizer request for project %s with commitID %s , with err %v",
// 			payload.ProjectId, payload.CommitId, err)
// 		return NO_RETRY_FAILURE
// 	}
// 	for retryCount := 0; ; {
// 		if retryCount == *settingsObj.RetryCount {
// 			log.Errorf("Webhook invocation failed for snapshot %s project %s with commitId %s after max-retry of %d",
// 				payload.SnapshotCID, payload.ProjectId, payload.CommitId, *settingsObj.RetryCount)
// 			return RETRY_IMMEDIATE
// 		}
// 		req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(reqParams))
// 		if err != nil {
// 			log.Fatalf("Failed to create new HTTP Req with URL %s for message %+v with error %+v",
// 				reqURL, commonTxReqParams, err)
// 			return RETRY_IMMEDIATE
// 		}
// 		req.Header.Add("Authorization", "Bearer "+settingsObj.Web3Storage.APIToken)
// 		req.Header.Add("accept", "application/json")
// 		req.Header.Add("Content-Type", "application/json")
//
// 		err = dagFinalizerClientRateLimiter.Wait(context.Background())
// 		if err != nil {
// 			log.Errorf("WebhookClient Rate Limiter wait timeout with error %+v", err)
// 			time.Sleep(1 * time.Second)
// 			continue
// 		}
// 		log.Debugf("Sending Req to DAG finalizer URL %s for project %s with snapshotCID %s commitId %s ",
// 			reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId)
// 		res, err := dagFinalizerClient.Do(req)
// 		if err != nil {
// 			retryCount++
// 			log.Errorf("Failed to send request %+v towards DAG finalizer URL %s for project %s with snapshotCID %s commitId %s with error %+v.  Retrying %d",
// 				req, reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
// 			continue
// 		}
// 		defer res.Body.Close()
// 		var resp map[string]json.RawMessage
// 		respBody, err := io.ReadAll(res.Body)
// 		if err != nil {
// 			retryCount++
// 			log.Errorf("Failed to read response body for project %s with snapshotCID %s commitId %s from DAG finalizer with error %+v. Retrying %d",
// 				payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
// 			continue
// 		}
// 		if res.StatusCode == http.StatusOK {
// 			err = json.Unmarshal(respBody, &resp)
// 			if err != nil {
// 				retryCount++
// 				log.Errorf("Failed to unmarshal response %+v for project %s with snapshotCID %s commitId %s from DAG finalizer with error %+v. Retrying %d",
// 					respBody, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
// 				time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
// 				continue
// 			}
// 			log.Debugf("Received 200 OK with body %+v from DAG finalizer for project %s with snapshotCID %s commitId %s ",
// 				resp, payload.ProjectId, payload.SnapshotCID, payload.CommitId)
// 			return NO_RETRY_SUCCESS
// 		} else {
// 			retryCount++
// 			log.Errorf("Received Error response %+v from DAG finalizer for project %s with commitId %s with statusCode %d and status : %s ",
// 				resp, payload.ProjectId, payload.CommitId, res.StatusCode, res.Status)
// 			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
// 			continue
// 		}
// 	}
// }
//
// func UploadToWeb3Storage(payload *datamodel.PayloadCommit) (string, bool) {
//
// 	reqURL := settingsObj.Web3Storage.URL + settingsObj.Web3Storage.UploadURLSuffix
// 	for retryCount := 0; ; {
// 		if retryCount == *settingsObj.RetryCount {
// 			log.Errorf("web3.storage upload failed for snapshot %s project %s with commitId %s after max-retry of %d",
// 				payload.SnapshotCID, payload.ProjectId, payload.CommitId, *settingsObj.RetryCount)
// 			return "", false
// 		}
// 		req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(payload.Payload))
// 		if err != nil {
// 			log.Fatalf("Failed to create new HTTP Req with URL %s for snapshot %s project %s with commitId %s with error %+v",
// 				reqURL, payload.SnapshotCID, payload.ProjectId, payload.CommitId, err)
// 			return "", false
// 		}
// 		req.Header.Add("Authorization", "Bearer "+settingsObj.Web3Storage.APIToken)
// 		req.Header.Add("accept", "application/json")
//
// 		err = web3StorageClientRateLimiter.Wait(context.Background())
// 		if err != nil {
// 			log.Errorf("Web3Storage Rate Limiter wait timeout with error %+v", err)
// 			time.Sleep(1 * time.Second)
// 			continue
// 		}
// 		log.Debugf("Sending Req to web3.storage URL %s for project %s with snapshotCID %s commitId %s ",
// 			reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId)
// 		res, err := w3sHttpClient.Do(req)
// 		if err != nil {
// 			retryCount++
// 			log.Errorf("Failed to send request %+v towards web3.storage URL %s for project %s with snapshotCID %s commitId %s with error %+v.  Retrying %d",
// 				req, reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
// 			continue
// 		}
// 		defer res.Body.Close()
// 		var resp datamodel.Web3StoragePutResponse
// 		respBody, err := io.ReadAll(res.Body)
// 		if err != nil {
// 			retryCount++
// 			log.Errorf("Failed to read response body for project %s with snapshotCID %s commitId %s from web3.storage with error %+v. Retrying %d",
// 				payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
// 			continue
// 		}
// 		if res.StatusCode == http.StatusOK {
// 			err = json.Unmarshal(respBody, &resp)
// 			if err != nil {
// 				retryCount++
// 				log.Errorf("Failed to unmarshal response %+v for project %s with snapshotCID %s commitId %s towards web3.storage with error %+v. Retrying %d",
// 					respBody, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
// 				time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
// 				continue
// 			}
// 			log.Debugf("Received 200 OK from web3.storage for project %s with snapshotCID %s commitId %s ",
// 				payload.ProjectId, resp.CID, payload.CommitId)
// 			return resp.CID, true
// 		} else {
// 			retryCount++
// 			var resp datamodel.Web3StorageErrResponse
// 			err = json.Unmarshal(respBody, &resp)
// 			if err != nil {
// 				log.Errorf("Failed to unmarshal error response %+v for project %s with snapshotCID %s commitId %s towards web3.storage with error %+v. Retrying %d",
// 					respBody, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
// 			} else {
// 				log.Errorf("Received Error response %+v from web3.storage for project %s with commitId %s with statusCode %d and status : %s ",
// 					resp, payload.ProjectId, payload.CommitId, res.StatusCode, res.Status)
// 			}
// 			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
// 			continue
// 		}
// 	}
// }
//
// func SubmitTxnToChain(payload *datamodel.PayloadCommit, tokenHash string) (requestID string, txHash string, retry retryType, err error) {
// 	reqURL := settingsObj.ContractCallBackend.URL
// 	var reqParams datamodel.AuditContractCommitParams
// 	reqParams.RequestID = payload.RequestID
// 	reqParams.ApiKeyHash = tokenHash
// 	reqParams.PayloadCommitId = payload.CommitId
// 	reqParams.ProjectId = payload.ProjectId
// 	reqParams.SnapshotCid = payload.SnapshotCID
// 	reqParams.TentativeBlockHeight = payload.TentativeBlockHeight
// 	commonTxReqParams.Params, err = json.Marshal(reqParams)
// 	if err != nil {
// 		log.Fatalf("Failed to marshal AuditContractCommitParams %+v towards tx-manager with error %+v", reqParams, err)
// 		return "", "", NO_RETRY_FAILURE, err
// 	}
// 	body, err := json.Marshal(commonTxReqParams)
// 	if err != nil {
// 		log.Fatalf("Failed to marshal request %+v towards tx-manager with error %+v", commonTxReqParams, err)
// 		return "", "", NO_RETRY_FAILURE, err
// 	}
//
// 	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(body))
// 	if err != nil {
// 		log.Fatalf("Failed to create new HTTP Req with URL %s for message %+v with error %+v",
// 			reqURL, commonTxReqParams, err)
// 		return "", "", NO_RETRY_FAILURE, err
// 	}
// 	req.Header.Add("Content-Type", "application/json")
// 	req.Header.Add("accept", "application/json")
//
// 	err = txClientRateLimiter.Wait(context.Background())
// 	if err != nil {
// 		log.Errorf("tx-manager Rate Limiter wait timeout with error %+v", err)
// 		time.Sleep(1 * time.Second)
// 		return "", "", RETRY_IMMEDIATE, err
// 	}
//
// 	log.Debugf("Sending Req with params %+v to tx-manager URL %s for project %s with snapshotCID %s commitId %s tokenHash %s.",
// 		reqParams, reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId, tokenHash)
// 	res, err := txMgrHttpClient.Do(req)
// 	if err != nil {
// 		log.Errorf("Failed to send request %+v towards tx-manager URL %s for project %s with snapshotCID %s commitId %s with error %+v",
// 			req, reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err)
// 		return "", "", RETRY_WITH_DELAY, err
// 	}
// 	defer res.Body.Close()
// 	var resp datamodel.AuditContractCommitResp
// 	respBody, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		log.Errorf("Failed to read response body for project %s with snapshotCID %s commitId %s from tx-manager with error %+v",
// 			payload.ProjectId, payload.SnapshotCID, payload.CommitId, err)
// 		return "", "", RETRY_WITH_DELAY, err
// 	}
// 	if res.StatusCode == http.StatusOK {
// 		err = json.Unmarshal(respBody, &resp)
// 		if err != nil {
// 			log.Errorf("Failed to unmarshal response %+v for project %s with snapshotCID %s commitId %s towards tx-manager with error %+v",
// 				respBody, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err)
// 			return "", "", RETRY_WITH_DELAY, err
// 		}
// 		if resp.Success {
// 			log.Debugf("Received Success response %+v from tx-manager for project %s with snapshotCID %s commitId %s.",
// 				resp, payload.ProjectId, payload.SnapshotCID, payload.CommitId)
// 			return resp.Data[0].RequestID, resp.Data[0].TxHash, NO_RETRY_SUCCESS, nil
// 		} else {
// 			var tmpRsp map[string]string
// 			_ = json.Unmarshal(respBody, &tmpRsp)
// 			log.Errorf("Received 200 OK with Error response %+v from tx-manager for project %s with snapshotCID %s commitId %s resp bytes %+v ",
// 				resp, payload.ProjectId, payload.SnapshotCID, payload.CommitId, tmpRsp)
// 			return "", "", RETRY_WITH_DELAY, errors.New("Received Error response from tx-manager : " + fmt.Sprintf("%+v", resp))
// 		}
// 	} else {
// 		log.Errorf("Received Error response %+v from tx-manager for project %s with snapshotCID %s commitId %s with statusCode %d and status : %s ",
// 			resp, payload.ProjectId, payload.SnapshotCID, payload.CommitId, res.StatusCode, res.Status)
// 		return "", "", RETRY_WITH_DELAY, errors.New("Received Error response from tx-manager" + fmt.Sprint(respBody))
// 	}
// }
//
// func InitTxManagerClient() {
// 	log.Info("InitTxManagerClient")
// 	t := http.Transport{
// 		// TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
// 		MaxIdleConns:        settingsObj.PayloadCommit.Concurrency,
// 		MaxConnsPerHost:     settingsObj.PayloadCommit.Concurrency,
// 		MaxIdleConnsPerHost: settingsObj.PayloadCommit.Concurrency,
// 		IdleConnTimeout:     0,
// 		DisableCompression:  true,
// 	}
//
// 	txMgrHttpClient = http.Client{
// 		Timeout:   time.Duration(settingsObj.HttpClientTimeoutSecs) * time.Second,
// 		Transport: &t,
// 	}
//
// 	commonTxReqParams.Method = "commitRecord"
//
// 	// Default values
// 	tps := rate.Limit(50) // 50 TPS
// 	burst := 50
// 	if settingsObj.ContractCallBackend.RateLimiter != nil {
// 		burst = settingsObj.ContractCallBackend.RateLimiter.Burst
// 		if settingsObj.ContractCallBackend.RateLimiter.RequestsPerSec == -1 {
// 			tps = rate.Inf
// 			burst = 0
// 		} else {
// 			tps = rate.Limit(settingsObj.ContractCallBackend.RateLimiter.RequestsPerSec)
// 		}
// 	}
// 	log.Infof("Rate Limit configured for tx-manager Client at %v TPS with a burst of %d", tps, burst)
// 	txClientRateLimiter = rate.NewLimiter(tps, burst)
// }
//
// func InitW3sClient() {
// 	log.Info("InitW3sClient")
//
// 	t := http.Transport{
// 		// TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
// 		MaxIdleConns:        settingsObj.Web3Storage.MaxIdleConns,
// 		MaxConnsPerHost:     settingsObj.Web3Storage.MaxIdleConns,
// 		MaxIdleConnsPerHost: settingsObj.Web3Storage.MaxIdleConns,
// 		IdleConnTimeout:     time.Duration(settingsObj.Web3Storage.IdleConnTimeout),
// 		DisableCompression:  true,
// 	}
//
// 	w3sHttpClient = http.Client{
// 		Timeout:   time.Duration(settingsObj.Web3Storage.TimeoutSecs) * time.Second,
// 		Transport: &t,
// 	}
//
// 	// Default values
// 	tps := rate.Limit(3) // 3 TPS
// 	burst := 3
// 	if settingsObj.Web3Storage.RateLimiter != nil {
// 		burst = settingsObj.Web3Storage.RateLimiter.Burst
// 		if settingsObj.Web3Storage.RateLimiter.RequestsPerSec == -1 {
// 			tps = rate.Inf
// 			burst = 0
// 		} else {
// 			tps = rate.Limit(settingsObj.Web3Storage.RateLimiter.RequestsPerSec)
// 		}
// 	}
// 	log.Infof("Rate Limit configured for web3.storage at %v TPS with a burst of %d", tps, burst)
// 	web3StorageClientRateLimiter = rate.NewLimiter(tps, burst)
// }
//
// func InitDAGFinalizerCallbackClient() {
// 	log.Info("InitDAGFinalizerCallbackClient")
//
// 	t := http.Transport{
// 		// TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
// 		MaxIdleConns:        settingsObj.PayloadCommit.Concurrency,
// 		MaxConnsPerHost:     settingsObj.PayloadCommit.Concurrency,
// 		MaxIdleConnsPerHost: settingsObj.PayloadCommit.Concurrency,
// 		IdleConnTimeout:     0,
// 		DisableCompression:  true,
// 	}
//
// 	dagFinalizerClient = http.Client{
// 		Timeout:   time.Duration(settingsObj.HttpClientTimeoutSecs) * time.Second,
// 		Transport: &t,
// 	}
// 	// Default values
// 	tps := rate.Limit(50) // 50 TPS
// 	burst := 20
// 	if settingsObj.PayloadCommit.DAGFinalizerRateLimiter != nil {
// 		burst = settingsObj.PayloadCommit.DAGFinalizerRateLimiter.Burst
// 		if settingsObj.PayloadCommit.DAGFinalizerRateLimiter.RequestsPerSec == -1 {
// 			tps = rate.Inf
// 			burst = 0
// 		} else {
// 			tps = rate.Limit(settingsObj.PayloadCommit.DAGFinalizerRateLimiter.RequestsPerSec)
// 		}
// 	}
// 	log.Infof("Rate Limit configured for dagFinalizerClient at %v TPS with a burst of %d", tps, burst)
// 	dagFinalizerClientRateLimiter = rate.NewLimiter(tps, burst)
// }
