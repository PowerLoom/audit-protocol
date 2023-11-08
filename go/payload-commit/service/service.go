package service

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cenkalti/backoff/v4"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/caching"
	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/reporting"
	"audit-protocol/goutils/settings"
	contractApi "audit-protocol/goutils/smartcontract/api"
	"audit-protocol/goutils/smartcontract/transactions"
	"audit-protocol/goutils/taskmgr"
	rabbitmqMgr "audit-protocol/goutils/taskmgr/rabbitmq"
	"audit-protocol/payload-commit/signer"
)

const ServiceName = "payload-commit"

type PayloadCommitService struct {
	settingsObj   *settings.SettingsObj
	redisCache    *caching.RedisCache
	ethClient     *ethclient.Client
	contractAPI   *contractApi.ContractApi
	ipfsClient    *ipfsutils.IpfsClient
	diskCache     *caching.LocalDiskCache
	txManager     *transactions.TxManager
	privKey       *ecdsa.PrivateKey
	issueReporter *reporting.IssueReporter
	httpClient    *retryablehttp.Client
	taskMgr       *rabbitmqMgr.RabbitmqTaskMgr
	uuid          uuid.UUID
}

// InitPayloadCommitService initializes the payload commit service.
func InitPayloadCommitService(reporter *reporting.IssueReporter) *PayloadCommitService {
	settingsObj, err := gi.Invoke[*settings.SettingsObj]()
	if err != nil {
		log.WithError(err).Fatal("failed to invoke settings object")
	}

	redisCache, err := gi.Invoke[*caching.RedisCache]()
	if err != nil {
		log.WithError(err).Fatal("failed to invoke redis cache")
	}

	ethClient, err := gi.Invoke[*ethclient.Client]()
	if err != nil {
		log.WithError(err).Fatal("failed to invoke eth client")
	}

	contractAPI, err := gi.Invoke[*contractApi.ContractApi]()
	if err != nil {
		log.WithError(err).Fatal("failed to invoke contract api")
	}

	ipfsClient, err := gi.Invoke[*ipfsutils.IpfsClient]()
	if err != nil {
		log.WithError(err).Fatal("failed to invoke ipfs client")
	}

	diskCache, err := gi.Invoke[*caching.LocalDiskCache]()
	if err != nil {
		log.WithError(err).Fatal("failed to invoke disk cache")
	}

	privKey, err := signer.GetPrivateKey(settingsObj.Signer.PrivateKey)
	if err != nil {
		log.WithError(err).Fatal("failed to get private key")
	}

	taskMgr, err := gi.Invoke[*rabbitmqMgr.RabbitmqTaskMgr]()
	if err != nil {
		log.WithError(err).Fatal("failed to invoke task manager")
	}

	pcService := &PayloadCommitService{
		settingsObj:   settingsObj,
		redisCache:    redisCache,
		ethClient:     ethClient,
		contractAPI:   contractAPI,
		ipfsClient:    ipfsClient,
		diskCache:     diskCache,
		txManager:     transactions.NewNonceManager(),
		privKey:       privKey,
		issueReporter: reporter,
		httpClient:    httpclient.GetDefaultHTTPClient(settingsObj.HttpClient.ConnectionTimeout),
		taskMgr:       taskMgr,
		uuid:          uuid.New(),
	}

	_ = pcService.initLocalCachedData()

	if err := gi.Inject(pcService); err != nil {
		log.WithError(err).Fatal("failed to inject payload commit service")
	}

	return pcService
}

func (s *PayloadCommitService) Run(msgBody []byte, topic string) error {
	log.WithField("msg", string(msgBody)).Debug("running new task")

	if strings.Contains(topic, taskmgr.FinalizedSuffix) {
		finalizedEventMsg := new(datamodel.PayloadCommitFinalizedMessage)

		err := json.Unmarshal(msgBody, finalizedEventMsg)
		if err != nil {
			log.WithError(err).Error("failed to unmarshal finalized payload commit message")

			// not able to report this issue as data for the reporting is not available
			return err
		}

		err = s.HandleFinalizedPayloadCommitTask(finalizedEventMsg)
		if err != nil {
			return err
		}
	}

	if strings.Contains(topic, taskmgr.DataSuffix) {
		payloadCommitMsg := new(datamodel.PayloadCommitMessage)

		err := json.Unmarshal(msgBody, payloadCommitMsg)
		if err != nil {
			log.WithError(err).Error("failed to unmarshal payload commit message")

			// not able to report this issue as data for the reporting is not available
			return err
		}

		err = s.HandlePayloadCommitTask(payloadCommitMsg)
		if err != nil {
			return err
		}
	}

	return nil
}

// HandlePayloadCommitTask handles the payload commit task.
func (s *PayloadCommitService) HandlePayloadCommitTask(msg *datamodel.PayloadCommitMessage) error {
	log.WithFields(log.Fields{
		"project ID":   msg.ProjectID,
		"epoch ID":     msg.EpochID,
		"snapshot CID": msg.SnapshotCID,
	}).Info("handling payload commit task")

	// sign payload commit message (eip712 signature)
	signerData, signature, err := s.signPayload(msg.SnapshotCID, msg.ProjectID, int64(msg.EpochID))
	if err != nil {
		go s.redisCache.AddSnapshotterStatusReport(context.Background(), msg.EpochID, msg.ProjectID, &datamodel.SnapshotterStatusReport{
			SubmittedSnapshotCid: msg.SnapshotCID,
			State:                datamodel.MissedSnapshotSubmission,
			Reason:               "failed to sign payload commit message",
		})

		return err
	}

	// create transaction payload
	txPayload := &datamodel.SnapshotRelayerPayload{
		ProjectID:   msg.ProjectID,
		EpochID:     msg.EpochID,
		SnapshotCID: msg.SnapshotCID,
		Request:     signerData.Message,
		Signature:   hex.EncodeToString(signature),
	}

	// if relayer host is not set, submit snapshot directly to contract
	if *s.settingsObj.Relayer.Host == "" {
		err = backoff.Retry(func() error {
			err = s.txManager.SubmitSnapshot(s.contractAPI, s.privKey, signerData, txPayload, signature)

			return err
		}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))
		if err != nil {
			log.WithError(err).Error("failed to submit snapshot to contract")

			go s.issueReporter.Report(
				reporting.MissedSnapshotIssue,
				msg.ProjectID,
				strconv.Itoa(msg.EpochID),
				map[string]interface{}{
					"issueDetails": "Error: " + err.Error(),
					"msg":          "failed to submit snapshot to contract",
				})
			go s.redisCache.UpdateEpochProcessingStatus(context.Background(), msg.ProjectID, msg.EpochID, "failed", err.Error())

			go s.redisCache.AddSnapshotterStatusReport(context.Background(), msg.EpochID, msg.ProjectID, &datamodel.SnapshotterStatusReport{
				SubmittedSnapshotCid: msg.SnapshotCID,
				State:                datamodel.MissedSnapshotSubmission,
				Reason:               "failed to submit snapshot to contract",
			})

			return err
		}
		go s.redisCache.UpdateEpochProcessingStatus(context.Background(), msg.ProjectID, msg.EpochID, "success", "")
	} else {
		// send payload commit message with signature to relayer
		if strings.Contains(msg.ProjectID, "aggregate_24h_top_pairs_lite") ||
			strings.Contains(msg.ProjectID, "aggregate_24h_top_tokens_lite") ||
			strings.Contains(msg.ProjectID, "aggregate_7d_top_pairs_lite") ||
			strings.Contains(msg.ProjectID, "aggregate_24h_stats_lite") {

			err = s.sendSignatureToRelayer(txPayload)

			if err != nil {
				log.WithError(err).Error("failed to submit snapshot to relayer")

				go s.issueReporter.Report(
					reporting.MissedSnapshotIssue,
					msg.ProjectID,
					strconv.Itoa(msg.EpochID),
					map[string]interface{}{
						"issueDetails": "Error: " + err.Error(),
						"msg":          "failed to submit snapshot to relayer",
					})
				go s.redisCache.UpdateEpochProcessingStatus(context.Background(), msg.ProjectID, msg.EpochID, "failed", err.Error())
				go s.redisCache.AddSnapshotterStatusReport(context.Background(), msg.EpochID, msg.ProjectID, &datamodel.SnapshotterStatusReport{
					SubmittedSnapshotCid: msg.SnapshotCID,
					State:                datamodel.MissedSnapshotSubmission,
					Reason:               "failed to submit snapshot to relayer",
				})

				return err
			}
		} else {
			log.WithField("project_id", msg.ProjectID).Info("skipping relayer submission for project")
		}
		go s.redisCache.UpdateEpochProcessingStatus(context.Background(), msg.ProjectID, msg.EpochID, "success", "")

	}

	return nil
}

// HandleFinalizedPayloadCommitTask handles finalized payload commit task.
func (s *PayloadCommitService) HandleFinalizedPayloadCommitTask(msg *datamodel.PayloadCommitFinalizedMessage) error {
	// log.Debug("handling finalized payload commit task")
	log.WithFields(log.Fields{
		"project ID":   msg.Message.ProjectID,
		"epoch ID":     msg.Message.EpochID,
		"snapshot CID": msg.Message.SnapshotCID,
	}).Debug("handling finalized payload commit task")

	// storing current finalized snapshot in redis
	go s.redisCache.StoreFinalizedSnapshot(context.Background(), msg.Message)
	var report *datamodel.SnapshotterStatusReport
	unfinalized, err := s.redisCache.GetUnfinalizedSnapshotAtEpochID(context.Background(), msg.Message.ProjectID, msg.Message.EpochID)
	if unfinalized != nil {
		if unfinalized.SnapshotCID != msg.Message.SnapshotCID {
			log.WithFields(log.Fields{
				"project ID":               msg.Message.ProjectID,
				"epoch ID":                 msg.Message.EpochID,
				"unfinalized snapshot CID": unfinalized.SnapshotCID,
				"finalized snapshot CID":   msg.Message.SnapshotCID,
			}).Debug("cached unfinalized snapshot cid does not match with finalized snapshot cid, fetching snapshot commit message from ipfs")
			go s.issueReporter.Report(
				reporting.SubmittedIncorrectSnapshotIssue,
				msg.Message.ProjectID,
				strconv.Itoa(msg.Message.EpochID),
				map[string]interface{}{
					"issueDetails":         "Error: submitted snapshot cid does not match with finalized snapshot cid",
					"submittedSnapshotCID": unfinalized.SnapshotCID,
					"finalizedSnapshotCID": msg.Message.SnapshotCID,
				})
			report = &datamodel.SnapshotterStatusReport{
				SubmittedSnapshotCid: unfinalized.SnapshotCID,
				FinalizedSnapshotCid: msg.Message.SnapshotCID,
				State:                datamodel.IncorrectSnapshotSubmission,
				Reason:               "INTERNAL_ERROR: submitted snapshot cid does not match with finalized snapshot cid",
			}
			// unpin unfinalized snapshot cid from ipfs
			err = s.ipfsClient.Unpin(unfinalized.SnapshotCID)
			if err != nil {
				log.WithError(err).WithField("cid", unfinalized.SnapshotCID).Error("failed to unpin snapshot cid from ipfs")
			}

		} else {
			report = &datamodel.SnapshotterStatusReport{
				SubmittedSnapshotCid: unfinalized.SnapshotCID,
				FinalizedSnapshotCid: msg.Message.SnapshotCID,
				State:                datamodel.SuccessfulSnapshotSubmission,
				Reason:               "Submission and finalized CID match",
			}
		}

	} else {
		report = &datamodel.SnapshotterStatusReport{
			SubmittedSnapshotCid: "",
			FinalizedSnapshotCid: msg.Message.SnapshotCID,
			State:                datamodel.OnlyFinalizedSnapshotSubmission,
			Reason:               "Only finalized CID received",
		}
	}
	dirPath := filepath.Join(s.settingsObj.LocalCachePath, msg.Message.ProjectID, "snapshots")

	// create file, if it does not exist
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		log.WithError(err).Error("failed to create directory for snapshots")
	} else {
		outputPath := filepath.Join(s.settingsObj.LocalCachePath, msg.Message.ProjectID, "snapshots", msg.Message.SnapshotCID+".json")
		if err != nil {
			log.WithError(err).Error("failed to marshal payload commit message")
		} else {
			// get snapshot from ipfs and store it in output path
			err = s.ipfsClient.GetSnapshotFromIPFS(msg.Message.SnapshotCID, outputPath)
			if err != nil {
				log.WithFields(log.Fields{
					"project ID":   msg.Message.ProjectID,
					"epoch ID":     msg.Message.EpochID,
					"snapshot CID": msg.Message.SnapshotCID,
				}).WithError(err).Error("failed to get finalizedsnapshot from ipfs")

				go s.issueReporter.Report(
					reporting.PayloadCommitInternalIssue,
					msg.Message.ProjectID,
					strconv.Itoa(msg.Message.EpochID),
					map[string]interface{}{
						"issueDetails": "Error: " + err.Error(),
						"msg": "failed to get finalized snapshot from ipfs, epoch id: " +
							strconv.Itoa(msg.Message.EpochID) +
							", snapshot cid: " + msg.Message.SnapshotCID,
					})
			}
		}

	}
	// generate report and store in redis
	err = s.redisCache.AddSnapshotterStatusReport(context.Background(), msg.Message.EpochID, msg.Message.ProjectID, report)
	if err != nil {
		log.WithError(err).Error("failed to add snapshotter status report to redis")
	}

	err = s.redisCache.StoreLastFinalizedEpoch(context.Background(), msg.Message.ProjectID, msg.Message.EpochID)
	if err != nil {
		log.WithError(err).Error("failed to store last finalized epoch")
	}

	return nil
}

// signPayload signs the payload commit message for relayer.
func (s *PayloadCommitService) signPayload(snapshotCid, projectId string, epochId int64) (*apitypes.TypedData, []byte, error) {
	log.Debug("signing payload commit message")

	signerData, err := signer.GetSignerData(s.ethClient, snapshotCid, projectId, epochId)
	if err != nil {
		log.WithError(err).Error("failed to get signer data")

		go s.issueReporter.Report(
			reporting.PayloadCommitInternalIssue,
			projectId,
			strconv.Itoa(int(epochId)),
			map[string]interface{}{
				"issueDetails": "Error: " + err.Error(),
				"msg":          "failed to get eip712 struct to create signature",
			})

		return nil, nil, err
	}

	signature, err := signer.SignMessage(s.privKey, signerData)
	if err != nil {
		log.WithError(err).Error("failed to sign message")

		go s.issueReporter.Report(
			reporting.PayloadCommitInternalIssue,
			projectId,
			strconv.Itoa(int(epochId)),
			map[string]interface{}{
				"issueDetails": "Error: " + err.Error(),
				"msg":          "failed to sign message",
			})

		return nil, nil, err
	}

	return signerData, signature, nil
}

// sendSignatureToRelayer sends the signature to the relayer.
func (s *PayloadCommitService) sendSignatureToRelayer(payload *datamodel.SnapshotRelayerPayload) error {
	type reqBody struct {
		ProjectID   string `json:"projectId"`
		SnapshotCID string `json:"snapshotCid"`
		EpochID     int    `json:"epochId"`
		Request     struct {
			Deadline    uint64 `json:"deadline"`
			SnapshotCid string `json:"snapshotCid"`
			EpochId     int    `json:"epochId"`
			ProjectId   string `json:"projectId"`
		} `json:"request"`
		Signature       string `json:"signature"`
		ContractAddress string `json:"contractAddress"`
	}

	rb := &reqBody{
		ProjectID:   payload.ProjectID,
		SnapshotCID: payload.SnapshotCID,
		EpochID:     payload.EpochID,
		Request: struct {
			Deadline    uint64 `json:"deadline"`
			SnapshotCid string `json:"snapshotCid"`
			EpochId     int    `json:"epochId"`
			ProjectId   string `json:"projectId"`
		}{
			Deadline:    (*big.Int)(payload.Request["deadline"].(*math.HexOrDecimal256)).Uint64(),
			SnapshotCid: payload.SnapshotCID,
			EpochId:     payload.EpochID,
			ProjectId:   payload.ProjectID,
		},
		Signature:       "0x" + payload.Signature,
		ContractAddress: s.settingsObj.Signer.Domain.VerifyingContract,
	}

	// url = "host+port" ; endpoint = "/endpoint"
	url := *s.settingsObj.Relayer.Host + *s.settingsObj.Relayer.Endpoint

	payloadBytes, err := json.Marshal(rb)
	if err != nil {
		log.WithError(err).Error("failed to marshal payload")

		return err
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.WithError(err).Error("failed to create request to relayer")

		return err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := s.httpClient.Do(req)
	if err != nil {
		log.WithError(err).Error("failed to send request to relayer")

		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.WithField("status", res.StatusCode).Error("failed to send request to relayer")

		return errors.New("failed to send request to relayer, status not ok")
	}

	log.Info("successfully sent signature to relayer")

	return nil
}

// initLocalCachedData fills data in redis cache from smart contract if not present locally.
func (s *PayloadCommitService) initLocalCachedData() error {
	log.Debug("initializing local cached data")

	projects, err := s.redisCache.GetStoredProjects(context.Background())
	if err != nil {
		log.WithError(err).Error("failed to get stored projects from redis cache")

		return err
	}

	if len(projects) > 0 {
		return nil
	}

	// get projects from smart contract
	projects, err = s.contractAPI.GetProjects(&bind.CallOpts{})
	if err != nil {
		log.WithError(err).Error("failed to get projects from smart contract")

		return err
	}

	if len(projects) == 0 {
		return nil
	}

	// store projects in redis cache
	err = s.redisCache.StoreProjects(context.Background(), projects)
	if err != nil {
		log.WithError(err).Error("failed to store projects in redis cache")

		return err
	}

	return nil
}
