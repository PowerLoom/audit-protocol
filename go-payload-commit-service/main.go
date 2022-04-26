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
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-redis/redis/v8"
	shell "github.com/ipfs/go-ipfs-api"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var ctx = context.Background()

var redisClient *redis.Client
var ipfsClient *shell.Shell
var httpClient http.Client
var settingsObj SettingsObj

//var httpTraceCtx context.Context

var REDIS_KEY_PAYLOAD_CIDS = "projectID:%s:payloadCids"
var REDIS_KEY_PENDING_TXNS = "projectID:%s:pendingTransactions"

const MAX_RETRY_COUNT = 3

var commonVigilParams CommonVigilRequestParams

func InitLogger() {
	if len(os.Args) < 2 {
		fmt.Println("Pass loglevel as an argument if you don't want default(INFO) to be set.")
		fmt.Println("Values to be passed for logLevel: ERROR(2),INFO(4),DEBUG(5)")
		log.SetLevel(log.DebugLevel)
	} else {
		logLevel, err := strconv.ParseUint(os.Args[1], 10, 32)
		if err != nil || logLevel > 6 {
			log.SetLevel(log.DebugLevel) //TODO: Change default level to error
		} else {
			//TODO: Need to come up with approach to dynamically update logLevel.
			log.SetLevel(log.Level(logLevel))
		}
	}
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {
	InitLogger()
	settingsObj = ParseSettings("../settings.json")
	InitIPFSClient()
	InitRedisClient()
	InitVigilClient()
	log.Info("Starting RabbitMq Consumer")
	InitRabbitmqConsumer()
}

func InitRabbitmqConsumer() {
	//conn, err := GetConn("amqp://powerloom:powerloom@172.31.23.45:5672/")
	concurrency := 50
	//rabbitMqURL := "amqp://guest:guest@localhost:5672/"
	//rabbitMqURL := "amqp://powerloom:powerloom@172.31.23.45:5672/" //Staging
	rabbitMqURL := fmt.Sprintf("amqp://%s:%s@%s:%d/", settingsObj.Rabbitmq.User, settingsObj.Rabbitmq.Password, settingsObj.Rabbitmq.Host, settingsObj.Rabbitmq.Port)
	conn, err := GetConn(rabbitMqURL)
	log.Infof("Starting rabbitMQ consumer connecting to URL: %s with concurreny %d", rabbitMqURL, concurrency)
	if err != nil {
		panic(err)
	}
	rmqExchangeName := "powerloom-backend"
	rmqExchangeName = settingsObj.Rabbitmq.Setup.Core.Exchange
	//TODO: These settings need to be moved to json config.
	rmqQueueName := "audit-protocol-commit-payloads"
	rmqRoutingKey := "commit-payloads"

	err = conn.StartConsumer(rmqQueueName, rmqExchangeName, rmqRoutingKey, RabbitmqMsgHandler, concurrency)
	if err != nil {
		panic(err)
	}

	forever := make(chan bool)
	<-forever
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
		log.Errorf("CRITICAL: Json unmarshal failed for payloadCommit %v, with err %v", d.Body, err)
		return true
	}

	if payloadCommit.Payload != nil {
		log.Debugf("Received incoming Payload commit message at tentative DAG Height %d for project %s with commitId %s from rabbitmq. Adding payload to IPFS.",
			payloadCommit.TentativeBlockHeight, payloadCommit.ProjectId, payloadCommit.CommitId)

		for retryCount := 0; ; {
			snapshotCid, err := ipfsClient.Add(bytes.NewReader(payloadCommit.Payload))
			if err != nil {
				if retryCount == MAX_RETRY_COUNT {
					log.Errorf("IPFS Add failed for message %+v after max-retry of %d, with err %v", payloadCommit, MAX_RETRY_COUNT, err)
					return false
				}
				retryCount++
				log.Errorf("Failed to add payload to IPFS %v, with err %v..retryCount %d .", d.Body, err, retryCount)
				continue
			}
			log.Debugf("IPFS add Successful. Snapshot CID is %s for payloadCommit %+v", snapshotCid, payloadCommit)
			payloadCommit.SnapshotCID = snapshotCid
			break
		}
		err = StorePayloadCidInRedis(&payloadCommit)
		if err != nil {
			return false
		}
	} else {
		log.Debugf("Received incoming Payload commit message at tentative DAG Height %d for project %s for resubmission at block %d from rabbitmq.",
			payloadCommit.TentativeBlockHeight, payloadCommit.ProjectId, payloadCommit.ResubmissionBlock)
	}
	err = PrepareAndSubmitTxnToChain(&payloadCommit)
	if err != nil {
		return false
	}
	log.Trace("Submitted txn to chain for %+v", payloadCommit)
	return true
}

func StorePayloadCidInRedis(payload *PayloadCommit) error {
	for retryCount := 0; ; {
		key := fmt.Sprintf(REDIS_KEY_PAYLOAD_CIDS, payload.ProjectId)
		res := redisClient.ZAdd(ctx, key,
			&redis.Z{
				Score:  float64(payload.TentativeBlockHeight),
				Member: payload.SnapshotCID,
			})
		if res.Err() != nil {
			if retryCount == MAX_RETRY_COUNT {
				log.Errorf("Failed to Add payload %s to redis Zset with key %s after max-retries of %d", payload.SnapshotCID, key, MAX_RETRY_COUNT)
				return res.Err()
			}
			retryCount++
			log.Errorf("Failed to Add payload %s to redis Zset with key %s..retryCount %d", payload.SnapshotCID, key, retryCount)
			continue
		}
		log.Debugf("Added payload %s to redis Zset with key %s successfully", payload.SnapshotCID, key)
		break
	}
	return nil
}

func AddToPendingTxnsInRedis(payload *PayloadCommit, tokenHash string, txHash string) error {
	key := fmt.Sprintf(REDIS_KEY_PENDING_TXNS, payload.ProjectId)
	var pendingtxn PendingTransaction
	pendingtxn.EventData.ProjectId = payload.ProjectId
	pendingtxn.EventData.SnapshotCid = payload.SnapshotCID
	pendingtxn.EventData.PayloadCommitId = payload.CommitId
	pendingtxn.EventData.Timestamp = float64(time.Now().Unix())
	pendingtxn.EventData.TentativeBlockHeight = payload.TentativeBlockHeight
	pendingtxn.EventData.ApiKeyHash = tokenHash
	if payload.Resubmitted {
		pendingtxn.LastTouchedBlock = payload.ResubmissionBlock
	} else {
		pendingtxn.LastTouchedBlock = -1
	}
	pendingtxn.TxHash = txHash
	pendingtxn.EventData.TxHash = txHash
	pendingtxnBytes, err := json.Marshal(pendingtxn)
	if err != nil {
		log.Errorf("Failed to marshal pendingTxn %+v", pendingtxn)
		return err
	}
	for retryCount := 0; ; {

		res := redisClient.ZAdd(ctx, key,
			&redis.Z{
				Score:  float64(payload.TentativeBlockHeight),
				Member: pendingtxnBytes,
			})
		if res.Err() != nil {
			if retryCount == MAX_RETRY_COUNT {
				log.Errorf("Failed to add payloadCid %s for project %s with commitID %s to pendingTxns in redis with err %+v after max retries of %d",
					payload.SnapshotCID, payload.ProjectId, payload.CommitId, res.Err(), MAX_RETRY_COUNT)
				return res.Err()
			}
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

func PrepareAndSubmitTxnToChain(payload *PayloadCommit) error {
	var snapshot Snapshot
	var tokenHash string
	snapshot.Type = "HOT_IPFS"
	snapshot.Cid = payload.SnapshotCID
	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		log.Errorf("CRITICAL. Failed to Json-Marshall snapshot %v for project %s with commitID %s , with err %v",
			snapshot, payload.ProjectId, payload.CommitId, err)
		return err
	}
	log.Trace("SnapshotBytes: ", snapshotBytes)
	if payload.ApiKeyHash == "" {
		bn := make([]byte, 32)
		ba := make([]byte, 32)
		bm := make([]byte, 32)
		bn[31] = 0
		ba[31] = 0
		bm[31] = 0

		tokenHash = crypto.Keccak256Hash(snapshotBytes, bn, ba, bm).String()
	} else {
		tokenHash = payload.ApiKeyHash
	}
	log.Tracef("Token hash generated for payload %+v for project %s with commitID %s  is : ",
		*payload, payload.ProjectId, payload.CommitId, tokenHash)
	var txHash string
	for retryCount := 0; ; {
		txHash, err = SubmitTxnToChain(payload, tokenHash)
		if err != nil {
			if retryCount == MAX_RETRY_COUNT {
				log.Errorf("Failed to send txn for snapshot %s for project %s with commitID %s and snapshotCID %s to Prost-Vigil with err %+v after max retries of %d",
					payload.SnapshotCID, payload.ProjectId, payload.CommitId, err, MAX_RETRY_COUNT)
				return err
			}
			retryCount++
			log.Errorf("Failed to send txn for snapshot %s for project %s with commitID %s to pendingTxns to Prost-Vigil with err %+v ..retryCount %d",
				payload.SnapshotCID, payload.ProjectId, payload.CommitId, err, retryCount)
			continue
		}
		break
	}
	/*Add to redis pendingTransactions*/
	err = AddToPendingTxnsInRedis(payload, tokenHash, txHash)
	if err != nil {
		return err
	}
	return nil
}

func SubmitTxnToChain(payload *PayloadCommit, tokenHash string) (string, error) {
	//log.Debugf("Sim: Submitted txn %s for payload %+v to ProstVigil-GW.", tokenHash, payload)
	// Directly calling ProstVigil LB via http://172.31.24.131:8888
	//reqURL = "http://172.31.24.131:8888/writer"
	reqURL := settingsObj.ContractCallBackend
	var reqParams AuditContractCommitParams
	reqParams.ApiKeyHash = tokenHash
	reqParams.PayloadCommitId = payload.CommitId
	reqParams.ProjectId = payload.ProjectId
	reqParams.SnapshotCid = payload.SnapshotCID
	reqParams.TentativeBlockHeight = payload.TentativeBlockHeight
	var err error
	commonVigilParams.Params, err = json.Marshal(reqParams)
	if err != nil {
		log.Errorf("Failed to marshal AuditContractCommitParams %+v towards Vigil GW with error %+v", reqParams, err)
		return "", err
	}
	body, err := json.Marshal(commonVigilParams)
	if err != nil {
		log.Errorf("Failed to marshal request %+v towards Vigil GW with error %+v", commonVigilParams, err)
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(body))
	if err != nil {
		log.Errorf("Failed to create new HTTP Req with URL %s for message %+v with error %+v",
			reqURL, commonVigilParams, err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("accept", "application/json")
	//req.Header.Add("X-API-KEY", "baf3b91a-bc68-484d-a377-741f8bf8de43") //TODO: Change to read from settings.
	log.Debugf("Sending Req %+v to Prost-Vigil URL %s for project %s with snapshotCID %s commitId %s tokenHash %s.",
		commonVigilParams, reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId, tokenHash)
	res, err := httpClient.Do(req)
	if err != nil {
		log.Errorf("Failed to send request %+v towards Prost-Vigil URL %s for project %s with snapshotCID %s commitId %s with error %+v",
			req, reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err)
		return "", err
	}
	defer res.Body.Close()
	var resp AuditContractCommitResp
	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Failed to read response body for project %s with snapshotCID %s commitId %s from Prost-Vigil with error %+v",
			payload.ProjectId, payload.SnapshotCID, payload.CommitId, err)
		return "", err
	}
	if res.StatusCode == http.StatusOK {
		err = json.Unmarshal(respBody, &resp)
		if err != nil {
			log.Errorf("Failed to unmarshal response %+v for project %s with snapshotCID %s commitId %s towards Prost-Vigil with error %+v",
				respBody, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err)
			return "", err
		}
		if resp.Success {
			log.Debugf("Received Success response %+v from Prost-Vigil for project %s with snapshotCID %s commitId %s.",
				resp, payload.ProjectId, payload.SnapshotCID, payload.CommitId)
			return resp.Data[0].TxHash, nil
		} else {
			var tmpRsp map[string]string
			_ = json.Unmarshal(respBody, &tmpRsp)
			log.Errorf("Received 200 OK with Error response %+v from Prost-Vigil for project %s with snapshotCID %s commitId %s resp bytes %+v ",
				resp, payload.ProjectId, payload.SnapshotCID, payload.CommitId, tmpRsp)
			return "", errors.New("Received Error response from Prost-Vigil : " + fmt.Sprintf("%+v", resp))
		}
	} else {
		log.Errorf("Received Error response %+v from Prost-Vigil for project %s with snapshotCID %s commitId %s with statusCode %d and status : %s ",
			resp, payload.ProjectId, payload.SnapshotCID, payload.CommitId, res.StatusCode, res.Status)
		return "", errors.New("Received Error response from Prost-Vigil" + fmt.Sprint(respBody))
	}
}

func InitVigilClient() {
	/*kl, err := os.OpenFile("ssl-keylog.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Error("Could not open keylog file due to error ", err)
	}*/
	//TODO: Move these to settings

	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        100,
		MaxConnsPerHost:     100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     0,
		DisableCompression:  true,
	}

	httpClient = http.Client{
		Timeout:   10 * time.Second,
		Transport: &t,
	}

	/* 	clientTrace := &httptrace.ClientTrace{
	   		GotConn: func(info httptrace.GotConnInfo) { log.Printf("conn was reused: %t", info.Reused) },
	   	}
	   	httpTraceCtx = httptrace.WithClientTrace(context.Background(), clientTrace)
	*/
	//commonVigilParams.Contract = "0xc8fe622716e7e2ab64a9c9595dd3b8df77477a9d"
	//commonVigilParams.Contract = "0xcd4fb4b7bb24cc0f623f6c5e9fd9b115360c2d66" //Staging PV
	commonVigilParams.Contract = strings.ToLower(settingsObj.AuditContract)
	commonVigilParams.Method = "commitRecord"
	commonVigilParams.NetworkId = 137
	commonVigilParams.HackerMan = false
	commonVigilParams.IgnoreGasEstimate = false
}

func InitRedisClient() {
	//redisURL := "powerloom-main1.op9vva.ng.0001.use2.cache.amazonaws.com" + ":" + strconv.Itoa(6379)
	//redisDb := 13
	//redisURL = "localhost:6379"
	//redisDb = 0
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
	//url := "/ip4/localhost/tcp/5001"
	//url = "/ip4/172.31.16.206/tcp/5001" //Old IPFS 0.7
	url := settingsObj.IpfsURL
	//url = "/ip4/172.31.31.46/tcp/5001" // new 0.13 IPFS
	// Convert the URL from /ip4/172.31.16.206/tcp/5001 to IP:Port format.
	connectUrl := strings.Split(url, "/")[2] + ":" + strings.Split(url, "/")[4]

	log.Infof("Initializing the IPFS client with IPFS Daemon URL %s.", connectUrl)
	//TODO: Move these to settings
	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        50,
		MaxConnsPerHost:     50,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     0,
		DisableCompression:  true,
	}

	ipfsHttpClient := http.Client{
		Timeout:   time.Duration(settingsObj.IpfsTimeout * 1000000000), //TODO: Read from settings.
		Transport: &t,
	}
	log.Debugf("Setting IPFS HTTP client timeout as %f seconds", ipfsHttpClient.Timeout.Seconds())
	ipfsClient = shell.NewShellWithClient(connectUrl, &ipfsHttpClient)
}
