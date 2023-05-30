package service

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"

	"audit-protocol/caching"
	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/ethclient"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/reporting"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/smartcontract"
	"audit-protocol/goutils/smartcontract/transactions"
	"audit-protocol/goutils/taskmgr"
	w3storage "audit-protocol/goutils/w3s"
	"audit-protocol/payload-commit/signer"
)

const ServiceName = "payload-commit"

type Service interface {
	Run(msgBody []byte, topic string) error
	HandlePayloadCommitTask(msg *datamodel.PayloadCommitMessage) error
	HandleFinalizedPayloadCommitTask(msg *datamodel.PayloadCommitFinalizedMessage) error
}

type PayloadCommitService struct {
	settingsObj        *settings.SettingsObj
	redisCache         caching.DbCache
	ethService         ethclient.Service
	contractAPIService smartcontract.Service
	ipfsService        ipfsutils.Service
	web3sClient        w3storage.Service
	diskCache          caching.DiskCache
	txManager          transactions.Service
	privKey            *ecdsa.PrivateKey
	issueReporter      reporting.Service
	signerService      signer.Service
}

// InitPayloadCommitService initializes the payload commit service.
func InitPayloadCommitService(
	settingsObj *settings.SettingsObj,
	redisCache caching.DbCache,
	ethService ethclient.Service,
	reporter reporting.Service,
	ipfsService ipfsutils.Service,
	diskCache caching.DiskCache,
	web3StorageService w3storage.Service,
	contractApiService smartcontract.Service,
	signerService signer.Service,
	txManagerService transactions.Service,
) Service {
	privKey, err := signerService.GetPrivateKey(settingsObj.Signer.PrivateKey)
	if err != nil {
		log.WithError(err).Fatal("failed to get private key")
	}

	pcService := &PayloadCommitService{
		settingsObj:        settingsObj,
		redisCache:         redisCache,
		ethService:         ethService,
		contractAPIService: contractApiService,
		ipfsService:        ipfsService,
		web3sClient:        web3StorageService,
		diskCache:          diskCache,
		txManager:          txManagerService,
		privKey:            privKey,
		issueReporter:      reporter,
		signerService:      signerService,
	}

	_ = pcService.initLocalCachedData()

	return pcService
}

func (s *PayloadCommitService) Run(msgBody []byte, topic string) error {
	log.WithField("msg", string(msgBody)).Debug("running new task")

	validate := validator.New()
	if strings.Contains(topic, taskmgr.FinalizedSuffix) {
		finalizedEventMsg := new(datamodel.PayloadCommitFinalizedMessage)

		err := json.Unmarshal(msgBody, finalizedEventMsg)
		if err != nil {
			log.WithError(err).Error("failed to unmarshal finalized payload commit message")

			// not able to report this issue as data for the reporting is not available
			return err
		}

		err = validate.Struct(finalizedEventMsg)
		if err != nil {
			log.WithError(err).Error("failed to validate finalized payload commit message")

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

		err = validate.Struct(payloadCommitMsg)
		if err != nil {
			log.WithError(err).Error("failed to validate payload commit message")

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

		go s.issueReporter.Report(datamodel.PayloadCommitInternalIssue, msg.ProjectID, strconv.Itoa(msg.EpochID), map[string]interface{}{
			"issueDetails": "Error: " + err.Error(),
			"msg":          errMsg,
		})
	}

	// store unfinalized payload cid in redis
	err = s.redisCache.AddUnfinalizedSnapshotCID(context.Background(), msg, time.Now().Unix()+3600*24)
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
			err = s.txManager.SubmitSnapshotToContract(s.privKey, signerData, txPayload, signature)

			return err
		}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))
		if err != nil {
			log.WithError(err).Error("failed to submit snapshot to contract")

			go s.issueReporter.Report(
				datamodel.MissedSnapshotIssue,
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
		err = s.txManager.SendSignatureToRelayer(txPayload)
		if err != nil {
			log.WithError(err).Error("failed to send signature to relayer")

			go s.issueReporter.Report(
				datamodel.MissedSnapshotIssue,
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
		err = s.ipfsService.GetSnapshotFromIPFS(msg.Message.SnapshotCID, filePath)
		if err != nil {
			log.WithError(err).Error("failed to get snapshot from ipfs")

			go s.issueReporter.Report(
				datamodel.PayloadCommitInternalIssue,
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
			datamodel.SubmittedIncorrectSnapshotIssue,
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
		err = s.ipfsService.GetSnapshotFromIPFS(msg.Message.SnapshotCID, filePath)
		if err != nil {
			log.WithError(err).Error("failed to get snapshot from ipfs")

			go s.issueReporter.Report(
				datamodel.PayloadCommitInternalIssue,
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
		err = s.ipfsService.Unpin(unfinalizedSnapshot.SnapshotCID)
		if err != nil {
			log.WithError(err).WithField("cid", unfinalizedSnapshot.SnapshotCID).Error("failed to unpin snapshot cid from ipfs")
		}
	} else {
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
			ipfsUploadErr = s.ipfsService.UploadSnapshotToIPFS(msg)
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

	signerData, err := s.signerService.GetSignerData(s.ethService, snapshotCid, projectId, epochId)
	if err != nil {
		log.WithError(err).Error("failed to get signer data")

		go s.issueReporter.Report(
			datamodel.PayloadCommitInternalIssue,
			projectId,
			strconv.Itoa(int(epochId)),
			map[string]interface{}{
				"issueDetails": "Error: " + err.Error(),
				"msg":          "failed to get eip712 struct to create signature",
			})

		return nil, nil, err
	}

	signature, err := s.signerService.SignMessage(s.privKey, signerData)
	if err != nil {
		log.WithError(err).Error("failed to sign message")

		go s.issueReporter.Report(
			datamodel.PayloadCommitInternalIssue,
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
	projects, err = s.contractAPIService.GetProjects()
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
