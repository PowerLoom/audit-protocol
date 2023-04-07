package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/pruning-archival/models"
	"audit-protocol/pruning-archival/service"
)

type migrator struct {
	redisClient *redis.Client
	ipfsclient  *ipfsutils.IpfsClient
	rabbitmq    *amqp.Connection
	settingsObj *settings.SettingsObj
}

type dummy struct {
	ProjectID string `json:"projectID"`
	Tail      int    `json:"tail"`
}

func main() {
	settingsObj := settings.ParseSettings()
	settingsObj.PruningServiceSettings = settingsObj.GetDefaultPruneConfig()

	// make sure you have two connections to the different redis instances
	redisClient := redisutils.InitRedisClient(settingsObj.Redis.Host, settingsObj.Redis.Port, settingsObj.Redis.Db, 100, settingsObj.Redis.Password, -1)
	defer redisClient.Close()

	ipfsclient := ipfsutils.InitClient("localhost:5001", 100, nil, 10)

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	m := &migrator{
		redisClient: redisClient,
		ipfsclient:  ipfsclient,
		rabbitmq:    conn,
		settingsObj: settingsObj,
	}

	m.simulate()
}

func (m *migrator) simulate() {
	log.Println("filling dummy data...")

	dummyData := new(dummy)
	data, _ := json.Marshal(dummyDataJson)
	_ = json.Unmarshal(data, dummyData)

	// clear stale data if present
	m.clearStaleData(dummyData)

	// create stored projects
	log.Info("storing dummy project")
	_, err := m.redisClient.SAdd(context.Background(), redisutils.REDIS_KEY_STORED_PROJECTS, dummyData.ProjectID).Result()
	if err != nil {
		log.Panicln("error getting stored projects", err)
	}

	// create two project dag segments
	// assuming that the project has segment of height 720
	log.Info("creating project dag segments")
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_METADATA, dummyData.ProjectID)

	segmentHeight := m.settingsObj.PruningServiceSettings.SegmentSize
	segments := map[string]interface{}{
		fmt.Sprintf("%d", segmentHeight): jsonMarshal(&models.ProjectDAGSegment{
			BeginHeight: 0,
			EndHeight:   segmentHeight,
			EndDAGCID:   m.createDummyDagCID([]byte("1segment")),
			StorageType: "pending",
		}),
		fmt.Sprintf("%d", segmentHeight*2): jsonMarshal(&models.ProjectDAGSegment{
			BeginHeight: segmentHeight + 1,
			EndHeight:   segmentHeight * 2,
			EndDAGCID:   m.createDummyDagCID([]byte("2segment")),
			StorageType: "pending",
		}),
	}

	data, _ = json.Marshal(segments)
	err = m.redisClient.HSet(context.Background(), key, segments).Err()
	if err != nil {
		log.Panicln("error setting project metadata", err)
	}

	// create cids and payload cids
	safeHeight := segmentHeight*2 + m.settingsObj.PruningServiceSettings.SummaryProjectsPruneHeightBehindHead
	cids := make([]*redis.Z, safeHeight)
	payloadCids := make([]*redis.Z, safeHeight)

	log.Info("creating cids and payload cids, might take a while...be patient :)")
	for i := 0; i < safeHeight; i++ {
		cid := m.createDummyDagCID([]byte(strconv.Itoa(i) + "cid"))
		cids[i] = &redis.Z{Score: float64(i), Member: cid}

		payloadCid := m.createDummyDagCID([]byte(strconv.Itoa(i) + "payloadCid"))
		payloadCids[i] = &redis.Z{Score: float64(i), Member: payloadCid}
	}

	// add cids to project cids zset
	err = m.redisClient.ZAdd(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, dummyData.ProjectID), cids...).Err()
	if err != nil {
		log.Panicln("error adding cids to project cids zset", err)
	}

	// add payload cids to project payload cids zset
	err = m.redisClient.ZAdd(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, dummyData.ProjectID), payloadCids...).Err()
	if err != nil {
		log.Panicln("error adding payload cids to project payload cids zset", err)
	}

	// create oldest project indexed height
	err = m.redisClient.Set(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_TAIL_INDEX, dummyData.ProjectID, "7d"), dummyData.Tail, 0).Err()
	if err != nil {
		log.Panicln("error setting oldest project indexed height", err)
	}

	// create pruning status
	err = m.redisClient.HSet(context.Background(), redisutils.REDIS_KEY_PRUNING_STATUS, dummyData.ProjectID, 0).Err()
	if err != nil {
		log.Panicln("error setting pruning status", err)
	}

	log.Info("pushing archival task to message queue")
	m.pushTaskToMessageQueue(dummyData.ProjectID)

	log.Println("done")
}

// createDummyDagCID creates a dummy dag cid in local node and stores it in the redis instance for testing and simulation purposes.
func (m *migrator) createDummyDagCID(data []byte) string {
	cid, err := m.ipfsclient.AddFileToIPFS(data)
	if err != nil {
		log.Panicln("error adding file to ipfs", err)
	}

	return cid
}

func (m *migrator) pushTaskToMessageQueue(projectID string) {
	ch, err := m.rabbitmq.Channel()
	if err != nil {
		log.Panicln("error getting channel", err)
	}

	defer ch.Close()

	err = ch.ExchangeDeclare("audit-protocol-backend", "direct", true, false, false, false, nil)
	if err != nil {
		log.Panicln("error declaring exchange", err)
	}

	task := &service.PruningTask{
		ProjectID: projectID,
	}

	data, err := json.Marshal(task)
	if err != nil {
		log.Panicln("error marshalling task", err)
	}

	err = ch.Publish("audit-protocol-backend", "dag-pruning:task", true, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	})
	if err != nil {
		log.Panicln("error publishing message", err)
	}
}

func (m *migrator) clearStaleData(dummyData *dummy) {
	err := m.redisClient.SRem(context.Background(), redisutils.REDIS_KEY_STORED_PROJECTS, dummyData.ProjectID).Err()
	if err != nil {
		log.WithError(err).Error("error removing project from stored projects set")
	}

	segmentHeight := m.settingsObj.PruningServiceSettings.SegmentSize

	err = m.redisClient.HDel(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_METADATA, dummyData.ProjectID), strconv.Itoa(segmentHeight), strconv.Itoa(segmentHeight*2)).Err()
	if err != nil {
		log.WithError(err).Error("error removing project metadata")
	}

	err = m.redisClient.ZRemRangeByScore(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, dummyData.ProjectID), "0", strconv.Itoa(segmentHeight*2+100)).Err()
	if err != nil {
		log.WithError(err).Error("error removing project cids")
	}

	err = m.redisClient.ZRemRangeByScore(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, dummyData.ProjectID), "0", strconv.Itoa(segmentHeight*2+100)).Err()
	if err != nil {
		log.WithError(err).Error("error removing project payload cids")
	}

	err = m.redisClient.Del(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_TAIL_INDEX, dummyData.ProjectID, "7d")).Err()
	if err != nil {
		log.WithError(err).Error("error removing oldest project indexed height")
	}

	err = m.redisClient.HDel(context.Background(), redisutils.REDIS_KEY_PRUNING_STATUS, dummyData.ProjectID).Err()
	if err != nil {
		log.WithError(err).Error("error removing pruning status")
	}
}

func jsonMarshal(val interface{}) string {
	data, _ := json.Marshal(val)
	return string(data)
}

var dummyDataJson = map[string]interface{}{
	"projectID": "uniswap_pairContract_trade_volume_0x21b8065d10f73ee2e260e5b47d3344d3ced7596e_UNISWAPV2-ph15-dummy",
	"tail":      1600,
}
