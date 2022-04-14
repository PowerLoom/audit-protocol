package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
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

var REDIS_KEY_PAYLOAD_CIDS = "projectID:%s:payloadCids"
var REDIS_KEY_PENDING_TXNS = "projectID:%s:pendingTransactions"

func InitLogger() {
	if len(os.Args) < 2 {
		fmt.Println("Pass loglevel as an argument if you don't want default(INFO) to be set.")
		fmt.Println("Values to be passed for logLevel: ERROR(2),INFO(4),DEBUG(5)")
		log.SetLevel(log.DebugLevel)
	} else {
		logLevel, err := strconv.ParseUint(os.Args[1], 10, 32)
		if err != nil || logLevel > 6 {
			log.SetLevel(log.DebugLevel)
		} else {
			//TODO: Need to come up with approach to dynamically update logLevel.
			log.SetLevel(log.Level(logLevel))
		}
	}
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {
	InitLogger()
	InitIPFSClient()
	InitRedisClient()
	log.Info("Starting RabbitMq Consumer")
	InitRabbitmqConsumer()
}

func InitRabbitmqConsumer() {
	//conn, err := GetConn("amqp://powerloom:powerloom@172.31.23.45:5672/")
	concurrency := 50
	rabbitMqURL := "amqp://guest:guest@localhost:5672/"
	conn, err := GetConn(rabbitMqURL)
	log.Infof("Starting rabbitMQ consumer connecting to URL: %s with concurreny %d", rabbitMqURL, concurrency)
	if err != nil {
		panic(err)
	}

	err = conn.StartConsumer("audit-protocol-commit-payloads", "powerloom-backend", "commit-payloads", RabbitmqMsgHandler, concurrency)

	if err != nil {
		panic(err)
	}

	forever := make(chan bool)
	<-forever
}

func RabbitmqMsgHandler(d amqp.Delivery) bool {
	if d.Body == nil {
		log.Errorf("Error, no message body!")
		return false
	}
	log.Debugf("Message from rabbitmq %+v", d.Body)

	var payloadCommit PayloadCommit

	err := json.Unmarshal(d.Body, &payloadCommit)
	if err != nil {
		log.Errorf("CRITICAL: Json unmarshal failed for payloadCommit %v, with err %v", d.Body, err)
		return true
	}

	log.Debugf("Unmarshalled PayloadCommit %+v, payload %+v", payloadCommit, payloadCommit.Payload)
	if payloadCommit.Payload != nil {
		//TODO: Add retry with backoff.
		snapshotCid, err := ipfsClient.Add(bytes.NewReader(payloadCommit.Payload))
		if err != nil {
			log.Errorf("Failed to add to ipfs %v, with err %v", d.Body, err)
			return false
		}
		log.Debugf("Snapshot CID is %s for payloadCommit %+v", snapshotCid, payloadCommit)
		payloadCommit.SnapshotCID = snapshotCid
		//TODO: Capture processing logs in redis for success/error.
		err = StorePayloadCidInRedis(payloadCommit)
		if err != nil {
			return false
		}
	}
	err = PrepareAndSubmitTxnToChain(payloadCommit)
	if err != nil {
		return false
	}
	log.Trace("Submitted txn to chain for %+v", payloadCommit)
	return true
}

func StorePayloadCidInRedis(payload PayloadCommit) error {
	//TODO: Add retry in case of failure
	key := fmt.Sprintf(REDIS_KEY_PAYLOAD_CIDS, payload.ProjectId)
	res := redisClient.ZAdd(ctx, key,
		&redis.Z{
			Score:  float64(payload.TentativeBlockHeight),
			Member: payload.SnapshotCID,
		})
	if res.Err() != nil {
		log.Errorf("Failed to Add payload %s to redis Zset with key %s", payload.SnapshotCID, key)
		return res.Err()
	}
	log.Debugf("Added payload %s to redis Zset with key %s successfully", payload.SnapshotCID, key)
	return nil
}

func AddToPendingTxnsInRedis(payload PayloadCommit, tokenHash string, txHash string) error {
	key := fmt.Sprintf(REDIS_KEY_PENDING_TXNS, payload.ProjectId)
	var pendingtxn PendingTransaction
	pendingtxn.EventData.ProjectId = payload.ProjectId
	pendingtxn.EventData.SnapshotCid = payload.SnapshotCID
	pendingtxn.EventData.PayloadCommitId = payload.CommitId
	pendingtxn.EventData.Timestamp = float64(time.Now().Unix())
	pendingtxn.EventData.TentativeBlockHeight = payload.TentativeBlockHeight
	pendingtxn.EventData.ApiKeyHash = tokenHash
	pendingtxn.LastTouchedBlock = -1
	pendingtxn.TxHash = txHash
	pendingtxn.EventData.TxHash = txHash
	pendingtxnBytes, err := json.Marshal(pendingtxn)
	if err != nil {
		log.Errorf("Failed to marshal pendingTxn %+v", pendingtxn)
		return err
	}
	res := redisClient.ZAdd(ctx, key,
		&redis.Z{
			Score:  float64(payload.TentativeBlockHeight),
			Member: pendingtxnBytes,
		})
	if res.Err() != nil {
		log.Errorf("Failed to add to pendingTxns in redis with err %+v", res.Err())
		return res.Err()
	}
	return nil
}

func PrepareAndSubmitTxnToChain(payload PayloadCommit) error {
	var snapshot Snapshot
	var tokenHash string
	snapshot.Type = "HOT_IPFS"
	snapshot.Cid = payload.SnapshotCID
	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		log.Errorf("Failed to Json-Marshall snapshot %v, with err %v", snapshot, err)
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
	log.Debug("Token hash is", tokenHash)
	txHash, err := SubmitTxnToChain(payload, tokenHash)
	/*Add to redis pendingTransactions*/

	err = AddToPendingTxnsInRedis(payload, tokenHash, txHash)
	if err != nil {
		log.Errorf("Failed to add to pendingTxns in redis with error %v, hence sending this for re-processing.", err)
		return err
	}
	return nil
}

func SubmitTxnToChain(payload PayloadCommit, tokenHash string) (string, error) {
	log.Debugf("Sim: Submitted txn %s for payload %+v to ProstVigil-GW.", tokenHash, payload)
	//TODO: Invoke MaticVigil API.

	return "dummyHash" + fmt.Sprint(rand.Int()), nil
}

func InitRedisClient() {
	redisURL := "powerloom-main1.op9vva.ng.0001.use2.cache.amazonaws.com" + ":" + strconv.Itoa(6379)
	redisDb := 13
	//TODO: Change post testing to fetch from settings.
	redisURL = "localhost:6379"
	redisDb = 0
	log.Infof("Connecting to redis DB %d at %s", redisDb, redisURL)
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       redisDb,
	})
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Errorf("Unable to connect to redis at %s with error %+v", redisURL, err)
	}
	log.Info("Connected successfully to Redis and received ", pong, " back")
}

func InitIPFSClient() {
	url := "/ip4/localhost/tcp/5001"
	// Convert the URL from /ip4/172.31.16.206/tcp/5001 to IP:Port format.
	connectUrl := strings.Split(url, "/")[2] + ":" + strings.Split(url, "/")[4]
	timeout := time.Duration(5 * 1000000000)

	log.Infof("Initializing the IPFS client with IPFS Daemon URL %s with timeout of %f seconds", connectUrl, timeout.Seconds())
	ipfsClient = shell.NewShell(connectUrl)
	ipfsClient.SetTimeout(timeout)
}
