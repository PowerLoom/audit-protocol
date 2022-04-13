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

func initLogger() {
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
	initLogger()
	InitIPFS()
	InitRedisClient()
	log.Info("Starting RabbitMq Consumer")
	initRabbitmqConsumer()
}

func initRabbitmqConsumer() {
	//conn, err := GetConn("amqp://powerloom:powerloom@172.31.23.45:5672/")
	conn, err := GetConn("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}

	err = conn.StartConsumer("audit-protocol-commit-payloads", "powerloom-backend", "commit-payloads", handler, 1)

	if err != nil {
		panic(err)
	}

	forever := make(chan bool)
	<-forever
}

func handler(d amqp.Delivery) bool {
	if d.Body == nil {
		fmt.Println("Error, no message body!")
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
	err = PrepareAndSubmitTxn(payloadCommit)
	if err != nil {
		return false
	}
	fmt.Println(string(d.Body))
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
		return res.Err()
	}
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
	pendingtxn.TxHash = txHash
	pendingtxnBytes, err := json.Marshal(pendingtxn)
	if err != nil {
		log.Error("Failed to marshal pendingTxn %+v", pendingtxn)
		return err
	}
	res := redisClient.ZAdd(ctx, key,
		&redis.Z{
			Score:  float64(payload.TentativeBlockHeight),
			Member: pendingtxnBytes,
		})
	if res.Err() != nil {
		log.Error("Failed to add to pendingTxns in redis with err %+v", res.Err())
		return res.Err()
	}
	return nil
}

func PrepareAndSubmitTxn(payload PayloadCommit) error {
	var snapshot Snapshot
	var tokenHash string
	snapshot.Type = "HOT_IPFS"
	snapshot.Cid = payload.SnapshotCID
	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		log.Errorf("Failed to Json-Marshall snapshot %v, with err %v", snapshot, err)
		return err
	}
	log.Debug("SnapshotBytes: ", snapshotBytes)
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
	txHash, err := makeTransaction(payload, tokenHash)
	/*Add to redis pendingTransactions*/

	err = AddToPendingTxnsInRedis(payload, tokenHash, txHash)
	if err != nil {
		log.Error("Failed to add to pendingTxns in redis with error %v, hence sending this for re-processing.", err)
		return err
	}
	return nil
}

func makeTransaction(payload PayloadCommit, tokenHash string) (string, error) {
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
	log.Info("Connecting to redis at:", redisURL)
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       redisDb,
	})
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Error("Unable to connect to redis at:")
	}
	log.Info("Connected successfully to Redis and received ", pong, " back")
}

func InitIPFS() {
	url := "/ip4/localhost/tcp/5001"
	// Convert the URL from /ip4/172.31.16.206/tcp/5001 to IP:Port format.
	connectUrl := strings.Split(url, "/")[2] + ":" + strings.Split(url, "/")[4]
	log.Debug("Initializing the IPFS client with IPFS Daemon URL:", connectUrl)
	ipfsClient = shell.NewShell(connectUrl)
	timeout := time.Duration(5 * 1000000000)
	ipfsClient.SetTimeout(timeout)
	log.Debugf("Setting IPFS timeout of %d seconds", timeout.Seconds())
}
