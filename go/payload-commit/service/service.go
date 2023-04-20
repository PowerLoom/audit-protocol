package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/caching"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/taskmgr"
	"audit-protocol/payload-commit/datamodel"
)

type PayloadCommitService struct {
	settingsObj *settings.SettingsObj
	redisCache  *caching.RedisCache
}

// InitPayloadCommitService initializes the payload commit service
func InitPayloadCommitService() *PayloadCommitService {
	settingsObj, err := gi.Invoke[*settings.SettingsObj]()
	if err != nil {
		log.WithError(err).Fatal("failed to invoke settings object")
	}

	redisCache, err := gi.Invoke[*caching.RedisCache]()
	if err != nil {
		log.WithError(err).Fatal("failed to invoke redis cache")
	}

	pcService := &PayloadCommitService{
		settingsObj: settingsObj,
		redisCache:  redisCache,
	}

	if err := gi.Inject(pcService); err != nil {
		log.WithError(err).Fatal("failed to inject payload commit service")
	}

	return pcService
}

func (s *PayloadCommitService) Run(msgBody []byte, topic string) error {
	if strings.Contains(topic, taskmgr.FinalizedSuffix) {
		payloadCommitMsg := new(datamodel.PayloadCommitMessage)

		err := json.Unmarshal(msgBody, payloadCommitMsg)
		if err != nil {
			log.WithError(err).Error("failed to unmarshal payload commit message")

			return err
		}

		return s.HandlePayloadCommitTask(payloadCommitMsg)
	}

	if strings.Contains(topic, taskmgr.DataSuffix) {
		finalizedEventMsg := new(datamodel.PayloadCommitFinalizedMessage)

		err := json.Unmarshal(msgBody, finalizedEventMsg)
		if err != nil {
			log.WithError(err).Error("failed to unmarshal finalized event message")

			return err
		}
	}

	return nil
}

func (s *PayloadCommitService) HandlePayloadCommitTask(msg *datamodel.PayloadCommitMessage) error {
	// check if project exists
	exists, err := s.redisCache.CheckIfProjectExists(context.Background(), msg.ProjectID)
	if err != nil {
		log.WithError(err).Error("failed to check if project exists")

		return err
	}

	if !exists {
		log.WithField("project_id", msg.ProjectID).Error("project does not exist")

		return fmt.Errorf("project %s does not exist", msg.ProjectID)
	}

	tentativeBlockHeight := 0

	// this can happen in case of summary project
	if msg.EpochEndHeight == 0 {
		lastTentativeBlockHeight := 0

		lastTentativeBlockHeight, err = s.redisCache.GetTentativeBlockHeight(context.Background(), msg.ProjectID)
		if err != nil {
			return err
		}

		tentativeBlockHeight = lastTentativeBlockHeight + 1
	} else {
		tentativeBlockHeight, err = s.calculateTentativeHeight(msg)
	}
}

// calculateTentativeHeight calculates the tentative block height for the payload commit
func (s *PayloadCommitService) calculateTentativeHeight(payloadCommit *datamodel.PayloadCommitMessage) (int, error) {
	tentativeBlockHeight := 0

	// get project epoch size
	epochSize, err := s.redisCache.GetProjectEpochSize(context.Background(), payloadCommit.ProjectID)
	if err != nil {
		return 0, err
	}

	if epochSize == 0 {
		epochSize = payloadCommit.EpochEndHeight - payloadCommit.EpochStartHeight + 1
	}

	firstEpochEndHeight, epochSize := getFirstEpochDetails(payloadCommit)
	if epochSize == 0 || firstEpochEndHeight == 0 {
		tentativeBlockHeight = 0
	} else {
		tentativeBlockHeight = (payloadCommit.SourceChainDetails.EpochEndHeight-firstEpochEndHeight)/epochSize + 1
		log.Debugf("Assigning tentativeBlockHeight as %d for project %s and payloadCommitID %s",
			tentativeBlockHeight, payloadCommit.ProjectId, payloadCommit.CommitId)
	}
	return tentativeBlockHeight
}

func getFirstEpochDetails(payloadCommit *datamodel.PayloadCommitMessage) (int, int) {
	// Record Project epoch size for the first epoch and also the endHeight.
	epochSize := FetchProjectEpochSize(payloadCommit.ProjectID)
	firstEpochEndHeight := 0
	if epochSize == 0 {
		epochSize = payloadCommit.SourceChainDetails.EpochEndHeight - payloadCommit.SourceChainDetails.EpochStartHeight + 1
		status := SetProjectEpochSize(payloadCommit.ProjectId, epochSize)
		if !status {
			return 0, 0
		}
		firstEpochEndHeight := payloadCommit.SourceChainDetails.EpochEndHeight
		status = SetFirstEpochEndHeight(payloadCommit.ProjectId, firstEpochEndHeight)
		if !status {
			return 0, 0
		}
	}
	firstEpochEndHeight = FetchFirstEpochEndHeight(payloadCommit.ProjectId)
	log.Debugf("Fetched firstEpochEndHeight as %d and epochSize as %d from redis for project %s",
		firstEpochEndHeight, epochSize, payloadCommit.ProjectId)
	return firstEpochEndHeight, epochSize
}
