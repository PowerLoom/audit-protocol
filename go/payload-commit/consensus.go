package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"audit-protocol/goutils/datamodel"

	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var consensusClientRateLimiter *rate.Limiter
var consensusHttpClient http.Client

type Epoch_ struct {
	Begin int `json:"begin"`
	End   int `json:"end"`
}
type SubmitSnapshotRequest struct {
	Epoch       Epoch_ `json:"epoch"`
	ProjectID   string `json:"projectID"`
	InstanceID  string `json:"instanceID"`
	SnapshotCID string `json:"snapshotCID"`
}

const SNAPSHOT_CONSENSUS_STATUS_ACCEPTED string = "ACCEPTED"
const SNAPSHOT_CONSENSUS_STATUS_FINALIZED string = "FINALIZED"

type SubmitSnapshotResponse struct {
	Status               string `json:"status"`
	DelayedSubmission    bool   `json:"delayedSubmission"`
	FinalizedSnapshotCID string `json:"finalizedSnapshotCID"`
}

func InitConsensusClient() {
	//TODO: Set default config
	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        settingsObj.ConsensusConfig.MaxIdleConns,
		MaxConnsPerHost:     settingsObj.ConsensusConfig.MaxIdleConns,
		MaxIdleConnsPerHost: settingsObj.ConsensusConfig.MaxIdleConns,
		IdleConnTimeout:     time.Duration(settingsObj.ConsensusConfig.IdleConnTimeout),
		DisableCompression:  true,
	}

	consensusHttpClient = http.Client{
		Timeout:   time.Duration(settingsObj.ConsensusConfig.TimeoutSecs) * time.Second,
		Transport: &t,
	}

	//Default values
	tps := rate.Limit(3) //3 TPS
	burst := 3
	if settingsObj.ConsensusConfig.RateLimiter != nil {
		burst = settingsObj.ConsensusConfig.RateLimiter.Burst
		if settingsObj.ConsensusConfig.RateLimiter.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(settingsObj.ConsensusConfig.RateLimiter.RequestsPerSec)
		}
	}
	log.Infof("Rate Limit configured for ConsensusClient at %v TPS with a burst of %d", tps, burst)
	consensusClientRateLimiter = rate.NewLimiter(tps, burst)
}

func PollConsensusForConfirmations() {
	for {
		if len(WaitQueueForConsensus) > 0 {
			//TODO: Scale this based on number of items in queue.
			log.Debugf("%d entries are waiting for consensus", len(WaitQueueForConsensus))
			for snapshotCID, value := range WaitQueueForConsensus {
				timeNow := time.Now().UnixMicro()
				if timeNow > value.ConsensusSubmissionTs+settingsObj.ConsensusConfig.FinalizationWaiTime*1000000 {
					log.Warnf("Finalization threshold crossed for %s project at epoch %d with snapshot %s at tentativeHeight %d.",
						value.ProjectId, value.SourceChainDetails.EpochEndHeight, value.SnapshotCID, value.TentativeBlockHeight)
					QueueLock.Lock()
					delete(WaitQueueForConsensus, snapshotCID)
					QueueLock.Unlock()
					continue
				}
				ProcessPendingSnapshot(snapshotCID, value)
			}
		}
		time.Sleep(time.Duration(settingsObj.ConsensusConfig.PollingIntervalSecs) * time.Second)
	}
}

func ProcessPendingSnapshot(snapshotCID string, payload *datamodel.PayloadCommit) {
	//Fetch status of snapshot
	status, err := SendRequestToConsensusService(payload, http.MethodPost, 3, "/checkForSnapshotConfirmation")
	if err != nil {
		log.Errorf("Failed to fetch snapshot status for payload Snapshot %s at tentativeHeight %d for project %s",
			payload.SnapshotCID, payload.TentativeBlockHeight, payload.ProjectId)
		return
	}
	if status == SNAPSHOT_CONSENSUS_STATUS_FINALIZED {
		QueueLock.Lock()
		delete(WaitQueueForConsensus, snapshotCID)
		QueueLock.Unlock()
		opStatus := AddToPendingTxns(payload, payload.ApiKeyHash, payload.RequestID)
		if !opStatus {
			log.Errorf("Failed to invoke webhook listener callback")
			return
		}
		log.Debugf("Processed finalized snapshot %s at tentativeHeight %d for project %s",
			payload.SnapshotCID, payload.TentativeBlockHeight, payload.ProjectId)
	} else {
		log.Debugf("Snapshot %s at tentativeHeight %d for project %s is not yet finalized and is in status %s",
			payload.SnapshotCID, payload.TentativeBlockHeight, payload.ProjectId, status)
		return
	}
}

func SubmitSnapshotForConsensus(payload *datamodel.PayloadCommit) (string, error) {
	payload.ConsensusSubmissionTs = time.Now().UnixMicro()
	return SendRequestToConsensusService(payload, http.MethodPost, *settingsObj.RetryCount, "/submitSnapshot")
}

func SendRequestToConsensusService(payload *datamodel.PayloadCommit, method string, maxRetries int, urlSuffix string) (string, error) {
	reqURL := settingsObj.ConsensusConfig.ServiceURL + urlSuffix
	req := SubmitSnapshotRequest{
		Epoch:       Epoch_{Begin: payload.SourceChainDetails.EpochStartHeight, End: payload.SourceChainDetails.EpochEndHeight},
		ProjectID:   payload.ProjectId,
		InstanceID:  settingsObj.InstanceId,
		SnapshotCID: payload.SnapshotCID,
	}
	reqBytes, _ := json.Marshal(req)
	for retryCount := 0; ; {
		if retryCount == maxRetries {
			log.Errorf("failed to send snapshot to consensus service for snapshot %s project %s with commitId %s after max-retry of %d",
				payload.SnapshotCID, payload.ProjectId, payload.CommitId, *settingsObj.RetryCount)
			return "", errors.New("max retires reached while sending to consensus client")
		}
		req, err := http.NewRequest(method, reqURL, bytes.NewBuffer(reqBytes))
		if err != nil {
			log.Errorf("Failed to create new HTTP Req with URL %s for snapshot %s project %s with commitId %s with error %+v",
				reqURL, payload.SnapshotCID, payload.ProjectId, payload.CommitId, err)
			return "", err
		}
		//req.Header.Add("Authorization", "Bearer "+settingsObj.ConsensusConfig.APIToken)
		req.Header.Add("accept", "application/json")

		err = consensusClientRateLimiter.Wait(context.Background())
		if err != nil {
			log.Errorf("ConsensusClient Rate Limiter wait timeout with error %+v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.Debugf("Sending Req %+v to consensus service URL %s for project %s with snapshotCID %s commitId %s ", req,
			reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId)
		res, err := consensusHttpClient.Do(req)
		if err != nil {
			retryCount++
			log.Errorf("Failed to send request %+v towards consensus service URL %s for project %s with snapshotCID %s commitId %s with error %+v.  Retrying %d",
				req, reqURL, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
			continue
		}
		defer res.Body.Close()
		var resp SubmitSnapshotResponse
		respBody, err := io.ReadAll(res.Body)
		if err != nil {
			retryCount++
			log.Errorf("Failed to read response body for project %s with snapshotCID %s commitId %s from consensus service with error %+v. Retrying %d",
				payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
			continue
		}
		if res.StatusCode == http.StatusOK {
			err = json.Unmarshal(respBody, &resp)
			if err != nil {
				retryCount++
				log.Errorf("Failed to unmarshal response %+v for project %s with snapshotCID %s commitId %s towards consensus service with error %+v. Retrying %d",
					respBody, payload.ProjectId, payload.SnapshotCID, payload.CommitId, err, retryCount)
				time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
				continue
			}
			log.Debugf("Received 200 OK from consensus service for project %s with snapshotCID %s commitId %s ",
				payload.ProjectId, payload.SnapshotCID, payload.CommitId)
			log.Debugf("Snapshot status for snapshot %s at tentativeHeight %d for project %s is %s",
				payload.SnapshotCID, payload.TentativeBlockHeight, payload.ProjectId, resp.Status)

			switch resp.Status {

			case SNAPSHOT_CONSENSUS_STATUS_FINALIZED:
				{
					if payload.SnapshotCID != resp.FinalizedSnapshotCID {
						log.Warnf("Snapshot %s at tentativeHeight %d is not matching finalized snapshot %s for project %s. Hence replacing local snapshot CID with finalized one.",
							payload.SnapshotCID, payload.TentativeBlockHeight, resp.FinalizedSnapshotCID, payload.ProjectId)
						payload.SnapshotCID = resp.FinalizedSnapshotCID
					} else {
						log.Debugf("Snapshot %s at tentativeHeight %d is matching with finalized snapshot %s for project %s. ",
							payload.SnapshotCID, payload.TentativeBlockHeight, resp.FinalizedSnapshotCID, payload.ProjectId)
					}
				}
			}
			return resp.Status, nil
		} else {
			if res.StatusCode == http.StatusForbidden ||
				res.StatusCode == http.StatusUnauthorized ||
				res.StatusCode == http.StatusBadRequest {
				return "", errors.New("not authorized to submit to consensus - check if the instance-ID/UUID is registered with consensus service")
			}
			retryCount++
			log.Errorf("Received Error response %+v from consensus service for project %s at tentativeHeight %d with commitId %s with statusCode %d and status : %s ",
				respBody, payload.ProjectId, payload.TentativeBlockHeight, payload.CommitId, res.StatusCode, res.Status)
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			continue
		}
	}
}
