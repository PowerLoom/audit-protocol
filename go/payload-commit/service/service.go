package service

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

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
	"audit-protocol/goutils/taskmgr/worker"
	w3storage "audit-protocol/goutils/w3s"
	"audit-protocol/payload-commit/signer"
)

const ServiceName = "payload-commit"

type PayloadCommitService struct {
	settingsObj   *settings.SettingsObj
	redisCache    *caching.RedisCache
	ethClient     *ethclient.Client
	contractAPI   *contractApi.ContractApi
	ipfsClient    *ipfsutils.IpfsClient
	web3sClient   *w3storage.W3S
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

	web3sClient, err := gi.Invoke[*w3storage.W3S]()
	if err != nil {
		log.WithError(err).Fatal("failed to invoke web3s client")
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
		web3sClient:   web3sClient,
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
	log.WithField("project_id", msg.ProjectID).Info("handling payload commit task")

	// check if payload cid is already present at the epochId, for the given project in redis
	// if yes, then skip the task
	unfinalizedSnapshot, err := s.redisCache.GetUnfinalizedSnapshotAtEpochID(context.Background(), msg.ProjectID, msg.EpochID)
	if unfinalizedSnapshot != nil {
		log.WithField("epochId", msg.EpochID).
			WithField("messageId", msg.ProjectID).
			WithField("snapshotCid", unfinalizedSnapshot.SnapshotCID).
			WithError(err).Debug("payload commit message already processed for the given epochId and project, skipping task")

		return nil
	}

	// upload payload commit msg to ipfs and web3 storage
	err = s.uploadToIPFSandW3s(msg)
	if err != nil {
		log.WithError(err).Error("failed to upload snapshot to ipfs")

		errMsg := "failed to upload snapshot to ipfs"
		if msg.Web3Storage {
			errMsg = "failed to upload snapshot to ipfs and web3 storage"
		}

		go s.issueReporter.Report(reporting.PayloadCommitInternalIssue, msg.ProjectID, strconv.Itoa(msg.EpochID), map[string]interface{}{
			"issueDetails": "Error: " + err.Error(),
			"msg":          errMsg,
		})

		go s.redisCache.AddSnapshotterStatusReport(context.Background(), msg.EpochID, msg.ProjectID, &datamodel.SnapshotterStatusReport{
			SubmittedSnapshotCid: "",
			State:                datamodel.MissedSnapshotSubmission,
			Reason:               errMsg,
		}, false)
	}

	// publish snapshot submitted event
	go s.publishSnapshotSubmittedEvent(msg)

	// sign payload commit message (eip712 signature)
	signerData, signature, err := s.signPayload(msg.SnapshotCID, msg.ProjectID, int64(msg.EpochID))
	if err != nil {
		go s.redisCache.AddSnapshotterStatusReport(context.Background(), msg.EpochID, msg.ProjectID, &datamodel.SnapshotterStatusReport{
			SubmittedSnapshotCid: msg.SnapshotCID,
			State:                datamodel.MissedSnapshotSubmission,
			Reason:               "failed to sign payload commit message",
		}, false)

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
			}, false)

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
				}, false)

				return err
			}
		} else {
			log.WithField("project_id", msg.ProjectID).Info("skipping relayer submission for project")
		}
		go s.redisCache.UpdateEpochProcessingStatus(context.Background(), msg.ProjectID, msg.EpochID, "success", "")
	}

	// store unfinalized payload cid in redis
	err = s.redisCache.AddUnfinalizedSnapshotCID(context.Background(), msg)
	if err != nil {
		log.WithField("epochId", msg.EpochID).
			WithField("messageId", msg.ProjectID).
			WithField("snapshotCid", msg.SnapshotCID).
			WithError(err).Error("failed to store snapshot cid in redis")

		go s.issueReporter.Report(reporting.PayloadCommitInternalIssue, msg.ProjectID, strconv.Itoa(msg.EpochID), map[string]interface{}{
			"issueDetails": "Error: " + err.Error(),
			"msg":          "failed to store unfinalized snapshot in redis",
		})
	}

	return nil
}

// HandleFinalizedPayloadCommitTask handles finalized payload commit task.
func (s *PayloadCommitService) HandleFinalizedPayloadCommitTask(msg *datamodel.PayloadCommitFinalizedMessage) error {
	log.Debug("handling finalized payload commit task")

	// storing current finalized snapshot in redis
	go s.redisCache.StoreFinalizedSnapshot(context.Background(), msg.Message)

	prevEpochId := msg.Message.EpochID - 1

	// fetch previous finalized snapshot from redis.
	// fetching previous finalized snapshot as current snapshotter might not have submitted the snapshot before it got finalized by other snapshotter(s)
	prevSnapshot, err := s.redisCache.GetFinalizedSnapshotAtEpochID(context.Background(), msg.Message.ProjectID, prevEpochId)
	if err != nil || prevSnapshot == nil {
		log.WithField("epochId", msg.Message.EpochID-1).WithError(err).Error("failed to get finalized snapshot cid from redis")

		return err
	}

	// check if payload is already in cache
	unfinalizedSnapshot, err := s.redisCache.GetUnfinalizedSnapshotAtEpochID(context.Background(), msg.Message.ProjectID, prevEpochId)
	if err != nil {
		log.WithError(err).Error("failed to get snapshot cid from redis")
	}

	var report *datamodel.SnapshotterStatusReport

	// if snapshot cid is not found in redis snapshot was missed
	if unfinalizedSnapshot == nil {
		log.Debug("snapshot was missed, fetching snapshot from ipfs")

		dirPath := filepath.Join(s.settingsObj.LocalCachePath, msg.Message.ProjectID, "snapshots")
		filePath := filepath.Join(dirPath, prevSnapshot.SnapshotCID+".json")

		// create file, if it does not exist
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			log.WithError(err).Error("failed to create file")
		}

		// get snapshot from ipfs and store it in output path
		err = s.ipfsClient.GetSnapshotFromIPFS(prevSnapshot.SnapshotCID, filePath)
		if err != nil {
			log.WithError(err).Error("failed to get snapshot from ipfs")

			go s.issueReporter.Report(
				reporting.PayloadCommitInternalIssue,
				msg.Message.ProjectID,
				strconv.Itoa(prevEpochId),
				map[string]interface{}{
					"issueDetails": "Error: " + err.Error(),
					"msg":          "failed to get snapshot from ipfs",
				})
		}

		report = &datamodel.SnapshotterStatusReport{
			FinalizedSnapshotCid: prevSnapshot.SnapshotCID,
			State:                datamodel.MissedSnapshotSubmission,
			Reason:               "INTERNAL_ERROR: snapshot was missed due to internal error",
		}
	} else if unfinalizedSnapshot.SnapshotCID != prevSnapshot.SnapshotCID {
		// if stored snapshot cid does not match with finalized snapshot cid, fetch snapshot from ipfs and store in local disk.
		log.Debug("cached snapshot cid does not match with finalized snapshot cid, fetching snapshot commit message from ipfs")

		go s.issueReporter.Report(
			reporting.SubmittedIncorrectSnapshotIssue,
			msg.Message.ProjectID,
			strconv.Itoa(prevEpochId),
			map[string]interface{}{
				"issueDetails":         "Error: " + "submitted snapshot cid does not match with finalized snapshot cid",
				"submittedSnapshotCID": unfinalizedSnapshot.SnapshotCID,
				"finalizedSnapshotCID": prevSnapshot.SnapshotCID,
			})

		dirPath := filepath.Join(s.settingsObj.LocalCachePath, msg.Message.ProjectID, "snapshots")
		filePath := filepath.Join(dirPath, prevSnapshot.SnapshotCID+".json")

		// create file, if it does not exist
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			log.WithError(err).Error("failed to create file")
		}

		finalizedSnapshot := make(map[string]interface{})

		// get snapshot from ipfs and store it in output path
		err = s.ipfsClient.GetSnapshotFromIPFS(prevSnapshot.SnapshotCID, filePath)
		if err != nil {
			log.WithError(err).Error("failed to get snapshot from ipfs")

			go s.issueReporter.Report(
				reporting.PayloadCommitInternalIssue,
				msg.Message.ProjectID,
				strconv.Itoa(prevEpochId),
				map[string]interface{}{
					"issueDetails": "Error: " + err.Error(),
					"msg":          "failed to get snapshot from ipfs",
				})
		} else {
			snapshotDataBytes, _ := os.ReadFile(filePath)
			_ = json.Unmarshal(snapshotDataBytes, &finalizedSnapshot)
		}

		report = &datamodel.SnapshotterStatusReport{
			SubmittedSnapshotCid: unfinalizedSnapshot.SnapshotCID,
			SubmittedSnapshot:    unfinalizedSnapshot.Snapshot,
			FinalizedSnapshotCid: prevSnapshot.SnapshotCID,
			FinalizedSnapshot:    finalizedSnapshot,
			State:                datamodel.IncorrectSnapshotSubmission,
			Reason:               "INTERNAL_ERROR: submitted snapshot cid does not match with finalized snapshot cid",
		}

		// unpin unfinalized snapshot cid from ipfs
		err = s.ipfsClient.Unpin(unfinalizedSnapshot.SnapshotCID)
		if err != nil {
			log.WithError(err).WithField("cid", unfinalizedSnapshot.SnapshotCID).Error("failed to unpin snapshot cid from ipfs")
		}
	} else {
		outputPath := filepath.Join(s.settingsObj.LocalCachePath, msg.Message.ProjectID, "snapshots", prevSnapshot.SnapshotCID+".json")

		data, err := json.Marshal(unfinalizedSnapshot.Snapshot)
		if err != nil {
			log.WithError(err).Error("failed to marshal payload commit message")
		}

		err = s.diskCache.Write(outputPath, data)
		if err != nil {
			log.WithError(err).Error("failed to write payload commit message to disk cache")
		}

		// if snapshot cid matches with finalized snapshot cid just increment the snapshot success count
		report = nil
	}

	// generate report and store in redis
	err = s.redisCache.AddSnapshotterStatusReport(context.Background(), prevEpochId, msg.Message.ProjectID, report, true)
	if err != nil {
		log.WithError(err).Error("failed to add snapshotter status report to redis")
	}

	err = s.redisCache.StoreLastFinalizedEpoch(context.Background(), msg.Message.ProjectID, msg.Message.EpochID)
	if err != nil {
		log.WithError(err).Error("failed to store last finalized epoch")
	}

	return nil
}

func (s *PayloadCommitService) uploadToIPFSandW3s(msg *datamodel.PayloadCommitMessage) error {
	log.WithField("msg", msg).Debug("uploading payload commit msg to ipfs and web3 storage")

	wg := sync.WaitGroup{}

	// upload to ipfs
	var ipfsUploadErr error

	ipfsErrChan := make(chan error)

	wg.Add(1)

	go func() {
		defer wg.Done()

		ipfsUploadErr = backoff.Retry(func() error {
			ipfsUploadErr = s.ipfsClient.UploadSnapshotToIPFS(msg)
			if ipfsUploadErr != nil {
				log.WithError(ipfsUploadErr).Error("failed to upload snapshot to ipfs, retrying")

				return ipfsUploadErr
			}

			return nil
		}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))

		if ipfsUploadErr != nil {
			log.WithError(ipfsUploadErr).Error("failed to upload snapshot to ipfs after max retries")
			ipfsErrChan <- ipfsUploadErr

			return
		}

		ipfsErrChan <- nil
	}()

	// upload to web3 storage
	var (
		w3sUploadErr error
		snapshotCid  string
	)

	if msg.Web3Storage {
		wg.Add(1)

		go func() {
			defer wg.Done()

			w3sUploadErr = backoff.Retry(func() error {
				snapshotCid, w3sUploadErr = s.web3sClient.UploadToW3s(msg.Message)
				if w3sUploadErr != nil {
					return w3sUploadErr
				}

				return nil
			}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))

			if w3sUploadErr != nil {
				log.WithError(w3sUploadErr).Error("failed to upload snapshot to web3 storage after max retries")
			}
		}()
	}

	// if ipfs upload fails, wait for web3 storage upload to finish if msg.Web3Storage flag is true
	if err := <-ipfsErrChan; err != nil {
		if msg.Web3Storage {
			wg.Wait()

			if w3sUploadErr != nil || snapshotCid == "" {
				return fmt.Errorf("failed to upload to ipfs and web3 storage")
			}

			msg.SnapshotCID = snapshotCid
		} else {
			return fmt.Errorf("failed to upload to ipfs")
		}
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

// publishSnapshotSubmittedEvent publishes snapshot submitted event message
func (s *PayloadCommitService) publishSnapshotSubmittedEvent(msg *datamodel.PayloadCommitMessage) {
	eventMsg := &datamodel.SnapshotSubmittedEventMessage{
		SnapshotCid: msg.SnapshotCID,
		EpochId:     msg.EpochID,
		ProjectId:   msg.ProjectID,
		BroadcastId: s.uuid.String(),
		Timestamp:   time.Now().Unix(),
	}

	msgBytes, err := json.Marshal(eventMsg)
	if err != nil {
		log.WithError(err).Error("failed to marshal snapshot submitted event message")

		return
	}

	err = s.taskMgr.Publish(context.Background(), worker.TypeEventDetectorWorker, msgBytes)
	if err != nil {
		log.WithField("msg", string(msgBytes)).WithError(err).Error("failed to publish snapshot submitted event message")

		return
	}

	log.Info("published snapshot submitted event message")
}
