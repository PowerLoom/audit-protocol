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

	"github.com/cenkalti/backoff/v4"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
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
	unfinalizedSnapshot, err := s.redisCache.GetSnapshotAtEpochID(context.Background(), msg.ProjectID, msg.EpochID)
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
		log.WithError(err).Error("failed to upload payload commit message to ipfs")

		errMsg := "failed to upload payload commit message to ipfs"
		if msg.Web3Storage {
			errMsg = "failed to upload payload commit message to ipfs and web3 storage"
		}

		go s.issueReporter.Report(reporting.PayloadCommitInternalIssue, msg.ProjectID, strconv.Itoa(msg.EpochID), map[string]interface{}{
			"issueDetails": "Error: " + err.Error(),
			"msg":          errMsg,
		})
	}

	// store unfinalized payload cid in redis
	err = s.redisCache.AddUnfinalizedSnapshotCID(context.Background(), msg)
	if err != nil {
		log.WithField("epochId", msg.EpochID).
			WithField("messageId", msg.ProjectID).
			WithField("snapshotCid", msg.SnapshotCID).
			WithError(err).Error("failed to store snapshot cid in redis")
	}

	// sign payload commit message (eip712 signature)
	signerData, signature, err := s.signPayload(msg.SnapshotCID, msg.ProjectID, int64(msg.EpochID))
	if err != nil {
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

			return err
		}
	} else {
		// send payload commit message with signature to relayer
		err = s.sendSignatureToRelayer(txPayload)
		if err != nil {
			log.WithError(err).Error("failed to send signature to relayer")

			go s.issueReporter.Report(
				reporting.MissedSnapshotIssue,
				msg.ProjectID,
				strconv.Itoa(msg.EpochID),
				map[string]interface{}{
					"issueDetails": "Error: " + err.Error(),
					"msg":          "failed to send signature to relayer",
				})

			return err
		}

		return nil
	}

	return nil
}

// HandleFinalizedPayloadCommitTask handles finalized payload commit task.
func (s *PayloadCommitService) HandleFinalizedPayloadCommitTask(msg *datamodel.PayloadCommitFinalizedMessage) error {
	log.Debug("handling finalized payload commit task")

	// check if payload is already in cache
	unfinalizedSnapshot, err := s.redisCache.GetSnapshotAtEpochID(context.Background(), msg.Message.ProjectID, msg.Message.EpochID)
	if err != nil {
		log.WithError(err).Error("failed to get snapshot cid from redis")
	}

	var report *datamodel.SnapshotterStatusReport

	// if snapshot cid is not found in redis snapshot was missed
	if unfinalizedSnapshot == nil {
		log.Debug("snapshot was missed, fetching snapshot from ipfs")

		dirPath := filepath.Join(s.settingsObj.LocalCachePath, msg.Message.ProjectID, "snapshots")
		filePath := filepath.Join(dirPath, msg.Message.SnapshotCID+".json")

		// create file, if it does not exist
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			log.WithError(err).Error("failed to create file")
		}

		// get snapshot from ipfs and store it in output path
		err = s.ipfsClient.GetSnapshotFromIPFS(msg.Message.SnapshotCID, filePath)
		if err != nil {
			log.WithError(err).Error("failed to get snapshot from ipfs")

			go s.issueReporter.Report(
				reporting.PayloadCommitInternalIssue,
				msg.Message.ProjectID,
				strconv.Itoa(msg.Message.EpochID),
				map[string]interface{}{
					"issueDetails": "Error: " + err.Error(),
					"msg":          "failed to get snapshot from ipfs",
				})

			return err
		}

		report = &datamodel.SnapshotterStatusReport{
			SubmittedSnapshotCid: "",
			FinalizedSnapshotCid: msg.Message.SnapshotCID,
			State:                datamodel.MissedSnapshotSubmission,
		}
	} else if unfinalizedSnapshot.SnapshotCID != msg.Message.SnapshotCID {
		// if stored snapshot cid does not match with finalized snapshot cid, fetch snapshot from ipfs and store in local disk.
		log.Debug("cached snapshot cid does not match with finalized snapshot cid, fetching snapshot commit message from ipfs")

		go s.issueReporter.Report(
			reporting.SubmittedIncorrectSnapshotIssue,
			msg.Message.ProjectID,
			strconv.Itoa(msg.Message.EpochID),
			map[string]interface{}{
				"issueDetails": "Error: " + "submitted snapshot cid does not match with finalized snapshot cid",
			})

		dirPath := filepath.Join(s.settingsObj.LocalCachePath, msg.Message.ProjectID, "snapshots")
		filePath := filepath.Join(dirPath, msg.Message.SnapshotCID+".json")

		// create file, if it does not exist
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			log.WithError(err).Error("failed to create file")
		}

		// get snapshot from ipfs and store it in output path
		err = s.ipfsClient.GetSnapshotFromIPFS(msg.Message.SnapshotCID, filePath)
		if err != nil {
			log.WithError(err).Error("failed to get snapshot from ipfs")

			go s.issueReporter.Report(
				reporting.PayloadCommitInternalIssue,
				msg.Message.ProjectID,
				strconv.Itoa(msg.Message.EpochID),
				map[string]interface{}{
					"issueDetails": "Error: " + err.Error(),
					"msg":          "failed to get snapshot from ipfs",
				})

			return err
		}

		report = &datamodel.SnapshotterStatusReport{
			SubmittedSnapshotCid: unfinalizedSnapshot.SnapshotCID,
			FinalizedSnapshotCid: msg.Message.SnapshotCID,
			State:                datamodel.IncorrectSnapshotSubmission,
		}

		// unpin unfinalized snapshot cid from ipfs
		err = s.ipfsClient.Unpin(unfinalizedSnapshot.SnapshotCID)
		if err != nil {
			log.WithError(err).WithField("cid", unfinalizedSnapshot.SnapshotCID).Error("failed to unpin snapshot cid from ipfs")
		}
	} else {
		err = s.redisCache.StoreLastFinalizedEpoch(context.Background(), msg.Message.ProjectID, msg.Message.EpochID)
		if err != nil {
			log.WithError(err).Error("failed to store last finalized epoch")
		}

		outputPath := filepath.Join(s.settingsObj.LocalCachePath, msg.Message.ProjectID, "snapshots", msg.Message.SnapshotCID+".json")

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
	err = s.redisCache.AddSnapshotterStatusReport(context.Background(), msg.Message.EpochID, msg.Message.ProjectID, report)
	if err != nil {
		log.WithError(err).Error("failed to add snapshotter status report to redis")

		return err
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
		Signature string `json:"signature"`
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
		Signature: "0x" + payload.Signature,
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
