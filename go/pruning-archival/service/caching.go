package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/pruning-archival/models"
)

// TODO move this to /audit-protocol/caching/redis.go
type caching struct {
	redisClient *redis.Client
}

// GetDAGCidsFromRedis fetches the DAGCids from redis for a given project
func (c *caching) GetDAGCidsFromRedis(projectID, taskID string, startScore int, endScore int) (map[int]string, error) {
	cids := make(map[int]string, endScore-startScore)
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectID)
	l := log.
		WithField("TaskID", taskID).
		WithField("Key", key).
		WithField("ProjectID", projectID)

	resp, err := c.redisClient.ZRangeByScoreWithScores(context.Background(), key, &redis.ZRangeBy{
		Min: strconv.Itoa(startScore),
		Max: strconv.Itoa(endScore),
	}).Result()
	if err != nil {
		l = l.WithError(err)
		if err == redis.Nil {
			l.Warn("Could not fetch dag cids due to key not found")

			return nil, err
		}

		l.Error("Could not fetch payloadCids")

		return nil, err
	}

	for j := range resp {
		cids[int(resp[j].Score)] = fmt.Sprintf("%v", resp[j].Member)
	}

	l.Debugf("Fetched %d dag cids from redis", len(cids))

	return cids, nil
}

// UpdatePrunedStatusToRedis updates the last pruned height of the project in redis
func (c *caching) UpdatePrunedStatusToRedis(projectID, taskID string, lastPrunedHeight int) error {
	l := log.WithField("TaskID", taskID).WithField("ProjectID", projectID)

	l.Infof("Updating Last Pruned Status at key %s", redisutils.REDIS_KEY_PRUNING_STATUS)

	_, err := c.redisClient.HSet(context.Background(), redisutils.REDIS_KEY_PRUNING_STATUS, projectID, lastPrunedHeight).Result()
	if err != nil {
		l.Warnf("Failed to update Last Pruned Status in redis")

		return err
	}

	l.Debugf("Updated last Pruned status successfully in redis")

	return nil
}

// UpdateDagSegmentStatusToRedis updates the dag segment status to redis
func (c *caching) UpdateDagSegmentStatusToRedis(cycleID string, projectID string, height int, dagSegment *models.ProjectDAGSegment) error {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_METADATA, projectID)

	bytes, err := json.Marshal(dagSegment)
	if err != nil {
		log.WithField("TaskID", cycleID).Fatalf("Failed to marshal dag segment due toe error %+v", err)

		return err
	}

	_, err = c.redisClient.HSet(context.Background(), key, height, bytes).Result()
	if err != nil {
		log.WithField("TaskID", cycleID).Warnf("Could not update key %s due to error %+v",
			key, err)

		return err
	}

	log.WithField("TaskID", cycleID).Debugf("Successfully updated archivedDAGSegments to redis for projectId %s with value %+v",
		projectID, *dagSegment)

	return nil
}

// GetPayloadCidsFromRedis fetches the payloadCids from redis for a given project
func (c *caching) GetPayloadCidsFromRedis(projectID, taskID string, startScore int, endScore int) (map[int]string, error) {
	cids := make(map[int]string, endScore-startScore)
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectID)
	l := log.
		WithField("TaskID", taskID).
		WithField("Key", key).
		WithField("ProjectID", projectID)

	resp, err := c.redisClient.ZRangeByScoreWithScores(context.Background(), key, &redis.ZRangeBy{
		Min: strconv.Itoa(startScore),
		Max: strconv.Itoa(endScore),
	}).Result()
	if err != nil {
		l = l.WithError(err)
		if err == redis.Nil {
			l.Warn("Could not fetch payloadCids due to key not found")

			return nil, err
		}

		l.Error("Could not fetch payloadCids")

		return nil, err
	}

	for j := range resp {
		cids[int(resp[j].Score)] = fmt.Sprintf("%v", resp[j].Member)
	}

	l.Debugf("Fetched %d payload Cids from redis", len(cids))

	return cids, nil
}

// PruneProjectCIDsInRedis prunes the given cids (ZSET) from redis
func (c *caching) PruneProjectCIDsInRedis(cycleID, projectId string, startScore int, endScore int) error {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectId)

	err := c.PruneZSetInRedis(cycleID, key, startScore, endScore)
	if err != nil {
		return err
	}

	key = fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectId)

	err = c.PruneZSetInRedis(cycleID, key, startScore, endScore)
	if err != nil {
		return err
	}

	return nil
}

// PruneZSetInRedis prunes the given zset from redis
func (c *caching) PruneZSetInRedis(cycleID, key string, startScore, endScore int) error {
	val, err := c.redisClient.ZRemRangeByScore(
		context.Background(), key,
		"-inf", // Always prune from start
		strconv.Itoa(endScore),
	).Result()

	if err != nil {
		log.WithField("TaskID", cycleID).Warnf("Could not prune redis Zset %s between height %d to %d due to error %+v. Retrying %d.",
			key, startScore, endScore, err)

		return err
	}

	log.WithField("TaskID", cycleID).Debugf("Successfully pruned redis Zset %s of %d entries between height %d and %d",
		key, val, startScore, endScore)

	return nil
}

// DeleteContentFromLocalDrive deletes the content from local drive
func (c *caching) DeleteContentFromLocalDrive(projectId, cachePath string, dagCids map[int]string, payloadCids map[int]string) error {
	for _, cid := range dagCids {
		fileName := fmt.Sprintf("%s/%s/%s.json", cachePath, projectId, cid)
		if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
			continue
		}

		err := os.Remove(fileName)
		if err != nil {
			log.Errorf("Failed to remove file %s from local cache due to error %+v", fileName, err)
			// TODO: Need to have some sort of pruning files older than 8 days logic to handle failures.
			return err
		}

		log.Debugf("Deleted file %s successfully from local cache", fileName)
	}

	for _, cid := range payloadCids {
		fileName := fmt.Sprintf("%s/%s/%s.json", cachePath, projectId, cid)
		if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
			continue
		}

		err := os.Remove(fileName)
		if err != nil {
			log.Errorf("Failed to remove file %s from local cache due to error %+v", fileName, err)

			return err
		}

		log.Debugf("Deleted file %s successfully from local cache", fileName)
	}

	return nil
}

// UpdatePruningCycleDetailsInRedis updates the pruning cycle details in redis
func (c *caching) UpdatePruningCycleDetailsInRedis(taskDetails *models.PruningTaskDetails, intervalMins int) error {
	cycleDetailsStr, _ := json.Marshal(taskDetails)

	_, err := c.redisClient.ZAdd(context.Background(), redisutils.REDIS_KEY_PRUNING_CYCLE_DETAILS, &redis.Z{Score: float64(taskDetails.StartTime), Member: cycleDetailsStr}).Result()
	if err != nil {
		log.Warnf("Failed to update PruningTaskDetails in redis due to error %+v", err)

		return err
	}

	log.Debugf("Successfully update PruningCycle Details in redis as %+v", taskDetails)

	count, err := c.redisClient.ZCard(context.Background(), redisutils.REDIS_KEY_PRUNING_CYCLE_DETAILS).Result()
	if err != nil {
		log.Warnf("Failed to get length of pruningCycleDetails Zset in redis due to error %+v", err)
	} else {
		zsetLen := count
		if zsetLen > 20 {
			log.Debugf("Pruning entries from pruningCycleDetails Zset as length is %d", zsetLen)
			endRank := -1*(zsetLen-20) + 1
			res := c.redisClient.ZRemRangeByRank(context.Background(), redisutils.REDIS_KEY_PRUNING_CYCLE_DETAILS, 0, endRank)

			log.Debugf("Pruned %d entries from pruningCycleDetails Zset", res.Val())
		}

		key := fmt.Sprintf(redisutils.REDIS_KEY_PRUNING_CYCLE_PROJECT_DETAILS, taskDetails.TaskID)
		c.redisClient.Expire(context.Background(), key, time.Duration(25*intervalMins*int(time.Minute)))
	}

	// TODO: Migrate to using slack App.

	return nil
}

// UpdatePruningProjectReportInRedis updates the project pruning report in redis
func (c *caching) UpdatePruningProjectReportInRedis(taskDetails *models.PruningTaskDetails, projectPruningReport *models.ProjectPruningReport) error {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PRUNING_CYCLE_PROJECT_DETAILS, taskDetails.TaskID)
	projectReportStr, _ := json.Marshal(projectPruningReport)

	_, err := c.redisClient.HSet(context.Background(), key, projectPruningReport.ProjectID, projectReportStr).Result()
	if err != nil {
		log.Warnf("Failed to update projectPruningReport in redis due to error %+v", err)

		return err
	}

	log.Debugf("Successfully update projectPruningReport Details in redis as %+v", *projectPruningReport)
	// TODO: Migrate to using slack App.
	/* 		if projectPruningReport.DAGSegmentsArchivalFailed > 0 || projectPruningReport.UnPinFailed > 0 {
	   			projectPruningReport.HostName, _ = os.Hostname()
	   			report, _ := json.MarshalIndent(projectPruningReport, "", "\t")
	   			slackutils.NotifySlackWorkflow(string(report), "High")
	   			projectPruneState.ErrorInLastCycle = true
	   		} else {
	   			if projectPruneState.ErrorInLastCycle {
	   				projectPruningReport.HostName, _ = os.Hostname()
	   				//Send clear status
	   				report, _ := json.MarshalIndent(projectPruningReport, "", "\t")
	   				slackutils.NotifySlackWorkflow(string(report), "Cleared")
	   			}
	   			projectPruneState.ErrorInLastCycle = false
	   		} */
	return nil
}

// FetchProjectDagSegments fetches the project's dag segments with their prune status from redis
// can be named as FetchProjectDagSegments
func (c *caching) FetchProjectDagSegments(projectId, taskID string) (*models.ProjectMetaData, error) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_METADATA, projectId)

	l := log.
		WithField("TaskID", taskID).
		WithField("Key", key).
		WithField("ProjectID", projectId)

	segmentsMap, err := c.redisClient.HGetAll(context.Background(), key).Result()
	if err != nil {
		if err == redis.Nil {
			l.WithError(err).Error("Failed to fetch project dag segments from redis as key does not exist.")

			return nil, err
		}

		l.WithError(err).Error("Failed to fetch project dag segments from redis")

		return nil, err
	}

	l.Debugf("Successfully fetched project metaData from redis for projectId %s with value %v", projectId, segmentsMap)

	projectMetaData := new(models.ProjectMetaData)
	projectMetaData.DagChains = segmentsMap

	return projectMetaData, nil

}

// GetOldestIndexedProjectHeight returns the oldest indexed project height
func (c *caching) GetOldestIndexedProjectHeight(projectID, taskID string, settingsObj *settings.SettingsObj) int {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_TAIL_INDEX, projectID, settingsObj.PruningServiceSettings.OldestProjectIndex)

	lastIndexHedeight := -1

	val, err := c.redisClient.Get(context.Background(), key).Result()

	l := log.
		WithField("TaskID", taskID).
		WithField("ProjectID", projectID)

	if err != nil && err == redis.Nil {
		l.Infof("Key %s does not exist", key)

		// For summary projects hard-code it to curBlockHeight-1000 as of now which gives safe values till 24hrs.
		key = fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_FINALIZED_HEIGHT, projectID)

		val, err = c.redisClient.Get(context.Background(), key).Result()
		if err != nil {
			if err == redis.Nil {
				l.Errorf("Key %s does not exist", key)

				return 0
			}
		}

		projectFinalizedHeight, err := strconv.Atoi(val)
		if err != nil {
			// fatal log as this should not happen
			l.Fatalf("Unable to convert retrieved projectFinalizedHeight for project %s to int due to error %+v ", projectID, err)

			return -1
		}

		// safe height for pruning in case not indexed yet
		lastIndexHedeight = projectFinalizedHeight - settingsObj.PruningServiceSettings.SummaryProjectsPruneHeightBehindHead

		return lastIndexHedeight
	}

	lastIndexHedeight, err = strconv.Atoi(val)
	if err != nil {
		l.Fatalf("Unable to convert retrieved lastIndexHeight for project %s to int due to error %+v ", projectID, err)

		return -1
	}

	l.Debugf("Fetched oldest index height %d for project %s from redis ", lastIndexHedeight, projectID)

	return lastIndexHedeight
}

// GetLastPrunedHeightOfProjectFromRedis gets the last pruned height for each project from redis
func (c *caching) GetLastPrunedHeightOfProjectFromRedis(projectID, taskID string) (int, error) {
	l := log.
		WithField("ProjectID", projectID).
		WithField("TaskID", taskID).
		WithField("Key", redisutils.REDIS_KEY_PRUNING_STATUS)

	l.Debug("Fetching Last Pruned Status of project")

	resp, err := c.redisClient.HGet(context.Background(), redisutils.REDIS_KEY_PRUNING_STATUS, projectID).Result()
	if err != nil {
		if err == redis.Nil {
			// Key doesn't exist.
			// this may happen when pruning service is running for the first time
			l.WithError(err).Error("Key doesn't exist..hence proceed from start of the block.")

			return 0, nil
		}

		l.Info("Failed to fetch Last Pruned Status of project from redis")

		return 0, err
	}

	lastPrunedHeight, err := strconv.Atoi(resp)
	if err != nil {
		l.WithError(err).Error("Failed to convert Last Pruned Status of project from redis to int")

		// if error then set lastPrunedHeight to 0
		lastPrunedHeight = 0
	}

	return lastPrunedHeight, nil
}
