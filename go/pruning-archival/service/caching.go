package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/slackutils"
	"audit-protocol/pruning-archival/models"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type caching struct {
	redisClient *redis.Client
}

// GetDAGCidsFromRedis fetches the DAGCids from redis for a given project
func (c *caching) GetDAGCidsFromRedis(cycleID string, projectId string, startScore int, endScore int) *map[int]string {
	cids := make(map[int]string, endScore-startScore)

	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectId)
	for i := 0; i < 5; i++ {
		res := c.redisClient.ZRangeByScoreWithScores(context.Background(), key, &redis.ZRangeBy{
			Min: strconv.Itoa(startScore),
			Max: strconv.Itoa(endScore),
		})
		if res.Err() != nil {
			// caching need not worry about cycle details logging but service should log it.
			log.WithField("CycleID", cycleID).Warnf("Could not fetch DAGCids for project %s due to error %+v. Retrying %d.", projectId, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		for j := range res.Val() {
			cids[int(res.Val()[j].Score)] = fmt.Sprintf("%v", res.Val()[j].Member)
		}
		log.WithField("CycleID", cycleID).Debugf("Fetched %d DAG Cids from redis for project %s", len(cids), projectId)
		return &cids
	}
	log.WithField("CycleID", cycleID).Errorf("Could not fetch DAGCids for project %s after max retries.", projectId)
	return nil
}

func (c *caching) UpdatePrunedStatusToRedis(cycleID string, projectList map[string]*models.ProjectPruneState, projectPruneState *models.ProjectPruneState) {
	lastPrunedStatus := make(map[string]string, len(projectList))
	lastPrunedStatus[projectPruneState.ProjectId] = strconv.Itoa(projectPruneState.LastPrunedHeight)

	for i := 0; i < 3; i++ {
		log.WithField("CycleID", cycleID).Infof("Updating Last Pruned Status at key %s", redisutils.REDIS_KEY_PRUNING_STATUS)
		res := c.redisClient.HSet(context.Background(), redisutils.REDIS_KEY_PRUNING_STATUS, lastPrunedStatus)
		if res.Err() != nil {
			log.WithField("CycleID", cycleID).Warnf("Failed to update Last Pruned Status in redis..Retrying %d", i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.WithField("CycleID", cycleID).Debugf("Updated last Pruned status %+v successfully in redis", projectPruneState.ProjectId)
		return
	}
	log.WithField("CycleID", cycleID).Errorf("Failed to update last Pruned status %+v in redis", projectPruneState.ProjectId)
}

func (c *caching) UpdateDagSegmentStatusToRedis(cycleID string, projectID string, height int, dagSegment *models.ProjectDAGSegment) bool {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_METADATA, projectID)
	for i := 0; i < 3; i++ {
		bytes, err := json.Marshal(dagSegment)
		if err != nil {
			log.WithField("CycleID", cycleID).Fatalf("Failed to marshal dag segment due toe error %+v", err)
			return false
		}
		res := c.redisClient.HSet(context.Background(), key, height, bytes)
		if res.Err() != nil {
			log.WithField("CycleID", cycleID).Warnf("Could not update key %s due to error %+v. Retrying %d.",
				key, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.WithField("CycleID", cycleID).Debugf("Successfully updated archivedDAGSegments to redis for projectId %s with value %+v",
			projectID, *dagSegment)
		return true
	}
	log.WithField("CycleID", cycleID).Errorf("Failed to update DAGSegments %+v for project %s to redis after max retries.", *dagSegment, projectID)
	return false
}

func (c *caching) GetPayloadCidsFromRedis(cycleID, projectId string, startScore int, endScore int) *map[int]string {
	cids := make(map[int]string, endScore-startScore)

	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectId)
	for i := 0; i < 5; i++ {
		res := c.redisClient.ZRangeByScoreWithScores(context.Background(), key, &redis.ZRangeBy{
			Min: strconv.Itoa(startScore),
			Max: strconv.Itoa(endScore),
		})
		if res.Err() != nil {
			log.WithField("CycleID", cycleID).Warnf("Could not fetch payloadCids for project %s due to error %+v. Retrying %d.", projectId, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		for j := range res.Val() {
			cids[int(res.Val()[j].Score)] = fmt.Sprintf("%v", res.Val()[j].Member)
		}
		log.WithField("CycleID", cycleID).Debugf("Fetched %d payload Cids from redis for project %s", len(cids), projectId)
		return &cids
	}
	log.WithField("CycleID", cycleID).Errorf("Could not fetch payloadCids for project %s after max retries.", projectId)
	return nil
}

func (c *caching) PruneProjectInRedis(cycleID, projectId string, startScore int, endScore int) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectId)
	c.PruneZSetInRedis(cycleID, key, startScore, endScore)
	key = fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectId)
	c.PruneZSetInRedis(cycleID, key, startScore, endScore)
}

func (c *caching) PruneZSetInRedis(cycleID, key string, startScore, endScore int) {
	i := 0
	for ; i < 3; i++ {
		res := c.redisClient.ZRemRangeByScore(
			context.Background(), key,
			"-inf", // Always prune from start
			strconv.Itoa(endScore),
		)
		if res.Err() != nil {
			log.WithField("CycleID", cycleID).Warnf("Could not prune redis Zset %s between height %d to %d due to error %+v. Retrying %d.",
				key, startScore, endScore, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}

		log.WithField("CycleID", cycleID).Debugf("Successfully pruned redis Zset %s of %d entries between height %d and %d",
			key, res.Val(), startScore, endScore)
		break
	}
	if i >= 3 {
		log.WithField("CycleID", cycleID).Errorf("Could not prune redis Zset %s between height %d to %d even after max-retries.",
			key, startScore, endScore)
	}
}

func (c *caching) DeleteContentFromLocalCache(projectId, cachePath string, dagCids *map[int]string, payloadCids *map[int]string) int {
	errCount := 0
	for _, cid := range *dagCids {
		fileName := fmt.Sprintf("%s/%s/%s.json", cachePath, projectId, cid)
		err := os.Remove(fileName)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				continue
			}
			log.Errorf("Failed to remove file %s from local cache due to error %+v", fileName, err)
			//TODO: Need to have some sort of pruning files older than 8 days logic to handle failures.
			errCount++
		}
		log.Debugf("Deleted file %s successfully from local cache", fileName)
	}

	for _, cid := range *payloadCids {
		fileName := fmt.Sprintf("%s/%s/%s.json", cachePath, projectId, cid)
		err := os.Remove(fileName)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				continue
			}
			log.Errorf("Failed to remove file %s from local cache due to error %+v", fileName, err)
			errCount++
		}
		log.Debugf("Deleted file %s successfully from local cache", fileName)
	}
	return errCount
}

func (c *caching) UpdatePruningCycleDetailsInRedis(cycleDetails *models.PruningCycleDetails, intervalMins int) {
	cycleDetailsStr, _ := json.Marshal(cycleDetails)
	for i := 0; i < 3; i++ {
		res := c.redisClient.ZAdd(context.Background(), redisutils.REDIS_KEY_PRUNING_CYCLE_DETAILS, &redis.Z{Score: float64(cycleDetails.CycleStartTime), Member: cycleDetailsStr})
		if res.Err() != nil {
			log.Warnf("Failed to update PruningCycleDetails in redis due to error %+v. Retrying %d", res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("Successfully update PruningCycle Details in redis as %+v", cycleDetails)
		res = c.redisClient.ZCard(context.Background(), redisutils.REDIS_KEY_PRUNING_CYCLE_DETAILS)
		if res.Err() == nil {
			zsetLen := res.Val()
			if zsetLen > 20 {
				log.Debugf("Pruning entries from pruningCycleDetails Zset as length is %d", zsetLen)
				endRank := -1*(zsetLen-20) + 1
				res = c.redisClient.ZRemRangeByRank(context.Background(), redisutils.REDIS_KEY_PRUNING_CYCLE_DETAILS, 0, endRank)
				log.Debugf("Pruned %d entries from pruningCycleDetails Zset", res.Val())
			}
		}
		key := fmt.Sprintf(redisutils.REDIS_KEY_PRUNING_CYCLE_PROJECT_DETAILS, cycleDetails.CycleID)
		c.redisClient.Expire(context.Background(), key, time.Duration(25*intervalMins*int(time.Minute)))
		//TODO: Migrate to using slack App.

		if cycleDetails.ProjectsProcessFailedCount > 0 {
			cycleDetails.HostName, _ = os.Hostname()
			report, _ := json.MarshalIndent(cycleDetails, "", "\t")
			slackutils.NotifySlackWorkflow(string(report), "Low", "PruningService")
			cycleDetails.ErrorInLastcycle = true
		} else {
			if cycleDetails.ErrorInLastcycle {
				cycleDetails.ErrorInLastcycle = false
				//Send clear status
				report, _ := json.MarshalIndent(cycleDetails, "", "\t")
				slackutils.NotifySlackWorkflow(string(report), "Cleared", "PruningService")
			}
		}
		return
	}
	log.Errorf("Failed to update pruningCycleDetails %+v in redis after max retries.", cycleDetails)
}

func (c *caching) UpdatePruningProjectReportInRedis(cycleDetails *models.PruningCycleDetails, projectPruningReport *models.ProjectPruningReport, projectPruneState *models.ProjectPruneState) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PRUNING_CYCLE_PROJECT_DETAILS, cycleDetails.CycleID)
	projectReportStr, _ := json.Marshal(projectPruningReport)
	for i := 0; i < 3; i++ {
		res := c.redisClient.HSet(context.Background(), key, projectPruningReport.ProjectID, projectReportStr)
		if res.Err() != nil {
			log.Warnf("Failed to update projectPruningReport in redis due to error %+v. Retrying %d", res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("Successfully update projectPruningReport Details in redis as %+v", *projectPruningReport)
		//TODO: Migrate to using slack App.
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
		return
	}
	log.Errorf("Failed to update projectPruningReport %+v for cycle %+v in redis after max retries.", *projectPruningReport, cycleDetails)
}

// FetchProjectMetaData fetches the project's dag segments with their prune status from redis
// can be named as FetchProjectDagSegments
func (c *caching) FetchProjectMetaData(cycleID, projectId string) *models.ProjectMetaData {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_METADATA, projectId)

	// sort of retry logic
	// can be improved by using exponential backoff
	for retry := 0; retry < 3; retry++ {
		res := c.redisClient.HGetAll(context.Background(), key)
		if res.Err() != nil {
			if res.Err() == redis.Nil {
				return nil
			}

			log.WithField("CycleID", cycleID).Warnf("Could not fetch key %s due to error %+v. Retrying %d.",
				key, res.Err(), retry)

			time.Sleep(5 * time.Second)

			continue
		}

		log.WithField("CycleID", cycleID).Debugf("Successfully fetched project metaData from redis for projectId %s with value %s",
			projectId, res.Val())

		projectMetaData := new(models.ProjectMetaData)
		projectMetaData.DagChains = res.Val()

		return projectMetaData
	}

	log.WithField("CycleID", cycleID).Errorf("Failed to fetch metaData for project %s from redis after max retries.", projectId)

	return nil
}

// GetOldestIndexedProjectHeight returns the oldest indexed project height
func (c *caching) GetOldestIndexedProjectHeight(cycleID string, projectPruneState *models.ProjectPruneState, settingsObj *settings.SettingsObj) int {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_TAIL_INDEX, projectPruneState.ProjectId, settingsObj.PruningServiceSettings.OldestProjectIndex)

	lastIndexHeight := -1

	res := c.redisClient.Get(context.Background(), key)
	err := res.Err()

	if err != nil && err == redis.Nil {
		log.WithField("CycleID", cycleID).Infof("Key %s does not exist", key)

		// For summary projects hard-code it to curBlockHeight-1000 as of now which gives safe values till 24hrs.
		key = fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_FINALIZED_HEIGHT, projectPruneState.ProjectId)

		res = c.redisClient.Get(context.Background(), key)

		err = res.Err()
		if err != nil {
			if err == redis.Nil {
				log.WithField("CycleID", cycleID).Errorf("Key %s does not exist", key)

				return 0
			}
		}

		projectFinalizedHeight, err := strconv.Atoi(res.Val())
		if err != nil {
			log.WithField("CycleID", cycleID).Fatalf("Unable to convert retrieved projectFinalizedHeight for project %s to int due to error %+v ", projectPruneState.ProjectId, err)

			return -1
		}

		lastIndexHeight = projectFinalizedHeight - settingsObj.PruningServiceSettings.SummaryProjectsPruneHeightBehindHead

		return lastIndexHeight
	}

	lastIndexHeight, err = strconv.Atoi(res.Val())
	if err != nil {
		log.WithField("CycleID", cycleID).Fatalf("Unable to convert retrieved lastIndexHeight for project %s to int due to error %+v ", projectPruneState.ProjectId, err)
		return -1
	}

	log.WithField("CycleID", cycleID).Debugf("Fetched oldest index height %d for project %s from redis ", lastIndexHeight, projectPruneState.ProjectId)

	return lastIndexHeight
}

func (c *caching) GetProjectsListFromRedis() map[string]*models.ProjectPruneState {
	ctx := context.Background()
	key := redisutils.REDIS_KEY_STORED_PROJECTS

	log.Debugf("Fetching stored Projects from redis at key: %s", key)

	for i := 0; ; i++ {
		res := c.redisClient.SMembers(ctx, key)
		if res.Err() != nil {
			if res.Err() == redis.Nil {
				log.Warnf("Stored Projects key doesn't exist..retrying")
				time.Sleep(5 * time.Minute)
				continue
			}
			log.Errorf("Failed to fetch stored projects from redis due to err %+v. Retrying %d", res.Err(), i)
			time.Sleep(5 * time.Second)

			continue
		}

		projectList := make(map[string]*models.ProjectPruneState, len(res.Val()))
		//projectList = make(map[string]*ProjectPruneState, 375)
		for index := range res.Val() {
			projectId := res.Val()[index]
			//if strings.Contains(projectId, "uniswap_V2PairsSummarySnapshot_UNISWAPV2") {
			projectPruneState := &models.ProjectPruneState{ProjectId: projectId}
			projectList[projectId] = projectPruneState
			//	break
			//}
		}

		log.Infof("Retrieved %d storedProjects %+v from redis", len(res.Val()), projectList)

		return projectList
	}
}

// GetLastPrunedStatusFromRedis gets the last pruned height for each project from redis
func (c *caching) GetLastPrunedStatusFromRedis(projectList map[string]*models.ProjectPruneState, cycleDetails *models.PruningCycleDetails) map[string]*models.ProjectPruneState {
	log.WithField("CycleID", cycleDetails.CycleID).Debug("Fetching Last Pruned Status at key:", redisutils.REDIS_KEY_PRUNING_STATUS)

	res := c.redisClient.HGetAll(context.Background(), redisutils.REDIS_KEY_PRUNING_STATUS)

	if len(res.Val()) == 0 {
		log.WithField("CycleID", cycleDetails.CycleID).Info("Failed to fetch Last Pruned Status  from redis for the projects.")
		//Key doesn't exist.
		log.WithField("CycleID", cycleDetails.CycleID).Info("Key doesn't exist..hence proceed from start of the block.")

		return projectList
	}
	err := res.Err()
	if err != nil {
		log.WithField("CycleID", cycleDetails.CycleID).Error("Ideally should not come here, which means there is some other redis error. To debug:", err)

		return projectList
	}

	// TODO: Need to handle dynamic addition of projects.
	for projectId, lastHeight := range res.Val() {
		if project, ok := projectList[projectId]; ok {
			project.LastPrunedHeight, err = strconv.Atoi(lastHeight)
			if err != nil {
				log.WithField("CycleID", cycleDetails.CycleID).Errorf("lastPrunedHeight corrupt for project %s. It will be set to 0", projectId)
				continue
			}
		} else {
			projectList[projectId] = &models.ProjectPruneState{ProjectId: projectId, LastPrunedHeight: 0}
		}
	}

	log.WithField("CycleID", cycleDetails.CycleID).Debugf("Fetched Last Pruned Status from redis %+v", projectList)

	return projectList
}
