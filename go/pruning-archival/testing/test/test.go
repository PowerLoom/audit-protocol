package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/pruning-archival/constants"
	"audit-protocol/pruning-archival/models"
)

func main() {
	projectID := "uniswap_pairContract_trade_volume_0x21b8065d10f73ee2e260e5b47d3344d3ced7596e_UNISWAPV2-ph15-stg"

	settingsObj := settings.ParseSettings()
	settingsObj.PruningServiceSettings = settingsObj.GetDefaultPruneConfig()

	redisClient := redisutils.InitRedisClient(settingsObj.Redis.Host, settingsObj.Redis.Port, settingsObj.Redis.Db, 100, settingsObj.Redis.Password, -1)
	defer redisClient.Close()

	segmentSize := settingsObj.PruningServiceSettings.SegmentSize
	segments, err := redisClient.HGetAll(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_METADATA, projectID)).Result()
	if err != nil {
		log.Panicln("TEST FAILED", err)
	}

	for _, segment := range segments {
		seg := new(models.ProjectDAGSegment)

		data, err := json.Marshal(segment)
		if err != nil {
			log.Panicln("TEST FAILED", err)
		}

		err = json.Unmarshal(data, seg)
		if err != nil {
			log.Panicln("TEST FAILED", err)
		}

		// check if the segment is cold or pruned
		if !(seg.StorageType == constants.DAG_CHAIN_STORAGE_TYPE_COLD || seg.StorageType == constants.DAG_CHAIN_STORAGE_TYPE_PRUNED) {
			log.Panicln("TEST FAILED: invalid storage type", seg.StorageType, "expected", constants.DAG_CHAIN_STORAGE_TYPE_COLD, "or", constants.DAG_CHAIN_STORAGE_TYPE_PRUNED)
		}
	}

	// check if CIDs are pruned
	cids, err := redisClient.ZRangeWithScores(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectID), 0, -1).Result()
	if err != nil {
		log.Panicln("TEST FAILED", err)
	}

	if len(cids) != 0 {
		log.Panicln("TEST FAILED: cids are not pruned")
	}

	// assuming dag segment height is 720
	if cids[0].Score <= 1440 {
		log.Panicln("TEST FAILED: invalid score", cids[0].Score, "expected", 1441)
	}

	payloadCIDs, err := redisClient.ZRangeWithScores(context.Background(), fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectID), 0, -1).Result()
	if err != nil {
		log.Panicln("TEST FAILED", err)
	}

	if len(payloadCIDs) != 0 {
		log.Panicln("TEST FAILED: cids are not pruned")
	}

	if int(payloadCIDs[0].Score) <= segmentSize*2 {
		log.Panicln("TEST FAILED: invalid score", payloadCIDs[0].Score, "expected", segmentSize*2+1)
	}

	// check project pruning status
	pruningStatus, err := redisClient.HGet(context.Background(), redisutils.REDIS_KEY_PRUNING_STATUS, projectID).Result()
	if err != nil {
		log.Panicln("TEST FAILED", err)
	}

	endHeight, err := strconv.Atoi(pruningStatus)
	if err != nil {
		log.Panicln("TEST FAILED", err)
	}

	if endHeight != segmentSize*2 {
		log.Panicln("TEST FAILED: invalid end height", endHeight, "expected", segmentSize*2)
	}

	// check if cids are backed up in local drive
	path := settingsObj.PruningServiceSettings.CARStoragePath
	for i := 0; i < 2; i++ {
		fileName := fmt.Sprintf("%s%s__%d_%d.json", path, projectID, i*(segmentSize+1), segmentSize*(i+1))
		if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
			log.Panicln("TEST FAILED: file cids are not backed up", fileName)
		}
	}

	log.Println("TEST PASSED")
}
