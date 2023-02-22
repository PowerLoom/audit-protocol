package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/slackutils"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type ProjectPruningReport struct {
	HostName                  string `json:"Host"`
	ProjectID                 string `json:"projectID"`
	DAGSegmentsProcessed      int    `json:"DAGSegmentsProcessed"`
	DAGSegmentsArchived       int    `json:"DAGSegmentsArchived"`
	DAGSegmentsArchivalFailed int    `json:"DAGSegmentsArchivalFailed,omitempty"`
	ArchivalFailureCause      string `json:"failureCause,omitempty"`
	CIDsUnPinned              int    `json:"CIDsUnPinned"`
	UnPinFailed               int    `json:"unPinFailed,omitempty"`
	LocalCacheDeletionsFailed int    `json:"localCacheDeletionsFailed,omitempty"`
}

type PruningCycleDetails struct {
	CycleID                     string `json:"pruningCycleID"`
	CycleStartTime              int64  `json:"cycleStartTime"`
	CycleEndTime                int64  `json:"cycleEndTime"`
	ProjectsCount               uint64 `json:"projectsCount"`
	ProjectsProcessSuccessCount uint64 `json:"projectsProcessSuccessCount"`
	ProjectsProcessFailedCount  uint64 `json:"projectsProcessFailedCount"`
	ProjectsNotProcessedCount   uint64 `json:"projectsNotProcessedCount"`
	HostName                    string `json:"hostName"`
	ErrorInLastcycle            bool   `json:"-"`
}

func UpdatePruningCycleDetailsInRedis() {
	cycleDetailsStr, _ := json.Marshal(cycleDetails)
	for i := 0; i < 3; i++ {
		res := redisClient.ZAdd(ctx, redisutils.REDIS_KEY_PRUNING_CYCLE_DETAILS, &redis.Z{Score: float64(cycleDetails.CycleStartTime), Member: cycleDetailsStr})
		if res.Err() != nil {
			log.Warnf("Failed to update PruningCycleDetails in redis due to error %+v. Retrying %d", res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("Successfully update PruningCycle Details in redis as %+v", cycleDetails)
		res = redisClient.ZCard(ctx, redisutils.REDIS_KEY_PRUNING_CYCLE_DETAILS)
		if res.Err() == nil {
			zsetLen := res.Val()
			if zsetLen > 20 {
				log.Debugf("Pruning entries from pruningCycleDetails Zset as length is %d", zsetLen)
				endRank := -1*(zsetLen-20) + 1
				res = redisClient.ZRemRangeByRank(ctx, redisutils.REDIS_KEY_PRUNING_CYCLE_DETAILS, 0, endRank)
				log.Debugf("Pruned %d entries from pruningCycleDetails Zset", res.Val())
			}
		}
		key := fmt.Sprintf(redisutils.REDIS_KEY_PRUNING_CYCLE_PROJECT_DETAILS, cycleDetails.CycleID)
		redisClient.Expire(ctx, key, time.Duration(25*settingsObj.PruningServiceSettings.RunIntervalMins*int(time.Minute)))
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

func UpdatePruningProjectReportInRedis(projectPruningReport *ProjectPruningReport, projectPruneState *ProjectPruneState) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PRUNING_CYCLE_PROJECT_DETAILS, cycleDetails.CycleID)
	projectReportStr, _ := json.Marshal(projectPruningReport)
	for i := 0; i < 3; i++ {
		res := redisClient.HSet(ctx, key, projectPruningReport.ProjectID, projectReportStr)
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
		   			projectPruneState.ErrorInLastcycle = true
		   		} else {
		   			if projectPruneState.ErrorInLastcycle {
		   				projectPruningReport.HostName, _ = os.Hostname()
		   				//Send clear status
		   				report, _ := json.MarshalIndent(projectPruningReport, "", "\t")
		   				slackutils.NotifySlackWorkflow(string(report), "Cleared")
		   			}
		   			projectPruneState.ErrorInLastcycle = false
		   		} */
		return
	}
	log.Errorf("Failed to update projectPruningReport %+v for cycle %+v in redis after max retries.", *projectPruningReport, cycleDetails)
}
