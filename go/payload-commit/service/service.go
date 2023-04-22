package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/cenkalti/backoff/v4"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/caching"
	datamodel2 "audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/smartcontract/api"
	"audit-protocol/goutils/taskmgr"
	w3storage "audit-protocol/goutils/w3s"
	"audit-protocol/payload-commit/datamodel"
	"audit-protocol/payload-commit/signer"
)

type PayloadCommitService struct {
	settingsObj *settings.SettingsObj
	redisCache  *caching.RedisCache
	ethClient   *ethclient.Client
	contractApi *contractApi.ContractApi
	ipfsClient  *ipfsutils.IpfsClient
	web3sClient *w3storage.W3S
	diskCache   *caching.LocalDiskCache
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

	pcService := &PayloadCommitService{
		settingsObj: settingsObj,
		redisCache:  redisCache,
		ethClient:   ethClient,
		contractApi: contractAPI,
		ipfsClient:  ipfsClient,
		web3sClient: web3sClient,
		diskCache:   diskCache,
	}

	_ = pcService.initLocalCachedData()

	if err := gi.Inject(pcService); err != nil {
		log.WithError(err).Fatal("failed to inject payload commit service")
	}

	return pcService
}

func (s *PayloadCommitService) Run(msgBody []byte, topic string) error {
	if strings.Contains(topic, taskmgr.FinalizedSuffix) {
		finalizedEventMsg := new(datamodel.PayloadCommitFinalizedMessage)

		err := json.Unmarshal(msgBody, finalizedEventMsg)
		if err != nil {
			log.WithError(err).Error("failed to unmarshal finalized payload commit message")

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

			return err
		}

		err = s.HandlePayloadCommitTask(payloadCommitMsg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PayloadCommitService) HandlePayloadCommitTask(msg *datamodel.PayloadCommitMessage) error {
	// check if project exists
	projects, err := s.contractApi.GetProjects(&bind.CallOpts{})

	if err != nil {
		log.WithError(err).Error("failed to get projects from contract")
		return err
	}

	projectExists := false
	for _, project := range projects {
		if project == msg.ProjectID {
			projectExists = true
			break
		}
	}

	if !projectExists {
		log.WithField("project_id", msg.ProjectID).Error("project does not exist, skipping payload commit")

		return nil
	}

	tentativeHeight, err := s.contractApi.GetTentativeHeight(&bind.CallOpts{})
	if err != nil {
		log.WithError(err).Error("failed to get tentative height from contract")

		return err
	}

	msg.TentativeBlockHeight = tentativeHeight

	// upload payload commit msg to ipfs and web3 storage
	if msg.MessageType == datamodel.MessageTypeAggregate || msg.MessageType == datamodel.MessageTypeSnapshot {
		err = s.uploadToIPFSandW3s(msg)
		if err != nil {
			log.WithError(err).Error("failed to upload payload commit message to ipfs")
		}

		go func() {
			_ = s.storePayloadToCache(msg)
			_ = s.redisCache.AddPayloadCID(context.Background(), msg.ProjectID, msg.SnapshotCID, float64(msg.TentativeBlockHeight))
		}()
	}

	signerData, signature, err := s.signPayload()
	if err != nil {
		return err
	}

	if msg.MessageType == datamodel.MessageTypeAggregate || msg.MessageType == datamodel.MessageTypeSnapshot {
		payload := &datamodel.SnapshotRelayerPayload{
			ProjectID:   msg.ProjectID,
			SnapshotCID: msg.SnapshotCID,
			EpochEnd:    msg.EpochEndHeight,
			Request:     signerData.Message,
			Signature:   string(signature),
		}

		s.sendSignatureToRelayer(payload)
	} else {
		payload := &datamodel.IndexRelayerPayload{
			ProjectID:                       msg.ProjectID,
			DagBlockHeight:                  msg.TentativeBlockHeight,
			IndexTailDagBlockHeight:         0,
			TailBlockEpochSourceChainHeight: 0,
			IndexIdentifierHash:             "",
			Request:                         signerData.Message,
			Signature:                       string(signature),
		}

		s.sendSignatureToRelayer(payload)
	}

	// adding tx to pending txs is taken care by state builder skipping that part

	return nil
}

func (s *PayloadCommitService) HandleFinalizedPayloadCommitTask(msg *datamodel.PayloadCommitFinalizedMessage) error {
	indexFinalizedPayload := new(datamodel.PowerloomIndexFinalizedMessage)
	aggregateFinalizedPayload := new(datamodel.PowerloomAggregateFinalizedMessage)
	snapshotFinalizedPayload := new(datamodel.PowerloomSnapshotFinalizedMessage)

	if msg.MessageType == datamodel.IndexFinalized {
		m, err := msg.UnmarshalMessage()
		if err != nil {
			log.WithError(err).Error("failed to unmarshal index finalized payload message")

			return err
		}

		indexFinalizedPayload = m.(*datamodel.PowerloomIndexFinalizedMessage)
	} else if msg.MessageType == datamodel.SnapshotFinalized {
		m, err := msg.UnmarshalMessage()
		if err != nil {
			log.WithError(err).Error("failed to unmarshal snapshot finalized payload message")

			return err
		}

		snapshotFinalizedPayload = m.(*datamodel.PowerloomSnapshotFinalizedMessage)

		// check if payload is already in cache
		payloadCid, err := s.redisCache.GetPayloadCidAtDAGHeight(context.Background(), snapshotFinalizedPayload.ProjectID, snapshotFinalizedPayload.DAGBlockHeight)
		if err == nil {
			return err
		}

		if payloadCid != snapshotFinalizedPayload.SnapshotCID {
			err = s.redisCache.RemovePayloadCIDAtHeight(context.Background(), snapshotFinalizedPayload.ProjectID, snapshotFinalizedPayload.DAGBlockHeight)
			if err != nil {
				log.WithError(err).Error("failed to remove payload cid from cache")

				return err
			}

			err = s.redisCache.AddPayloadCID(context.Background(), snapshotFinalizedPayload.ProjectID, snapshotFinalizedPayload.SnapshotCID, float64(snapshotFinalizedPayload.DAGBlockHeight))
			if err != nil {
				log.WithError(err).Error("failed to add payload cid to cache")

				return err
			}
		}

	} else if msg.MessageType == datamodel.AggregateFinalized {
		m, err := msg.UnmarshalMessage()
		if err != nil {
			log.WithError(err).Error("failed to unmarshal aggregate finalized payload message")

			return err
		}

		aggregateFinalizedPayload = m.(*datamodel.PowerloomAggregateFinalizedMessage)

		// check if payload is already in cache
		payloadCid, err := s.redisCache.GetPayloadCidAtDAGHeight(context.Background(), aggregateFinalizedPayload.ProjectID, aggregateFinalizedPayload.)
		if err == nil {
			return err
		}

		if payloadCid != snapshotFinalizedPayload.SnapshotCID {
			err = s.redisCache.RemovePayloadCIDAtHeight(context.Background(), snapshotFinalizedPayload.ProjectID, snapshotFinalizedPayload.DAGBlockHeight)
			if err != nil {
				log.WithError(err).Error("failed to remove payload cid from cache")

				return err
			}

			err = s.redisCache.AddPayloadCID(context.Background(), snapshotFinalizedPayload.ProjectID, snapshotFinalizedPayload.SnapshotCID, float64(snapshotFinalizedPayload.DAGBlockHeight))
			if err != nil {
				log.WithError(err).Error("failed to add payload cid to cache")

				return err
			}
		}
	}

}

func (s *PayloadCommitService) uploadToIPFSandW3s(msg *datamodel.PayloadCommitMessage) error {
	log.WithField("msg", msg).Debug("uploading payload commit msg to ipfs and web3 storage")
	wg := sync.WaitGroup{}

	wg.Add(2)

	// upload to ipfs
	var ipfsUploadErr error
	go func() {
		defer wg.Done()
		ipfsUploadErr = backoff.Retry(func() error {
			ipfsUploadErr = s.ipfsClient.UploadSnapshotToIPFS(msg)
			if ipfsUploadErr != nil {
				log.WithError(ipfsUploadErr).Error("failed to upload snapshot to ipfs, retrying")

				return ipfsUploadErr
			}

			return nil
		}, backoff.NewExponentialBackOff())

		if ipfsUploadErr != nil {
			log.WithError(ipfsUploadErr).Error("failed to upload snapshot to ipfs after max retries")
		}
	}()

	// upload to web3 storage
	var w3sUploadErr error
	var snapshotCid string
	if msg.Web3Storage {
		go func() {
			wg.Done()
			w3sUploadErr = backoff.Retry(func() error {
				snapshotCid, w3sUploadErr = s.uploadToW3s(msg)
				if w3sUploadErr != nil {
					return w3sUploadErr
				}

				return nil
			}, backoff.NewExponentialBackOff())

			if w3sUploadErr != nil {
				log.WithError(w3sUploadErr).Error("failed to upload snapshot to web3 storage after max retries")
			}
		}()
	}

	wg.Wait()

	if ipfsUploadErr != nil && !msg.Web3Storage {
		return fmt.Errorf("failed to upload to ipfs")
	}

	if ipfsUploadErr != nil && w3sUploadErr != nil {
		return fmt.Errorf("failed to upload to ipfs and web3 storage")
	}

	if ipfsUploadErr != nil && w3sUploadErr == nil {
		msg.SnapshotCID = snapshotCid

		return nil
	}

	return nil
}

func (s *PayloadCommitService) uploadToW3s(msg *datamodel.PayloadCommitMessage) (string, error) {
	reqURL := s.settingsObj.Web3Storage.URL + s.settingsObj.Web3Storage.UploadURLSuffix

	defaultHTTPClient := httpclient.GetDefaultHTTPClient()

	payloadCommit, err := json.Marshal(msg.Message)
	if err != nil {
		log.WithError(err).Error("failed to marshal payload commit message")

		return "", err
	}

	req, err := retryablehttp.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(payloadCommit))
	if err != nil {
		log.WithError(err).Error("failed to create request to web3.storage")

		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+s.settingsObj.Web3Storage.APIToken)
	req.Header.Add("accept", "application/json")

	err = s.web3sClient.Limiter.Wait(context.Background())
	if err != nil {
		log.Errorf("web3 storage rate limiter wait errored")

		return "", err
	}

	log.WithField("msg", string(payloadCommit)).Debug("sending request to web3.storage")

	res, err := defaultHTTPClient.Do(req)
	if err != nil {
		log.WithError(err).Error("failed to send request to web3.storage")

		return "", err
	}

	defer res.Body.Close()

	web3resp := new(datamodel2.Web3StoragePutResponse)

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).Error("failed to read response body from web3.storage")

		return "", err
	}

	if res.StatusCode == http.StatusOK {
		err = json.Unmarshal(respBody, web3resp)
		if err != nil {
			log.WithError(err).Error("failed to unmarshal response from web3.storage")

			return "", err
		}

		log.WithField("cid", web3resp.CID).Info("successfully uploaded payload commit to web3.storage")

		return web3resp.CID, nil
	} else {
		resp := new(datamodel2.Web3StorageErrResponse)

		err = json.Unmarshal(respBody, resp)
		if err != nil {
			log.WithError(err).Error("failed to unmarshal error response from web3.storage")
		} else {
			log.WithField("payloadCommit", string(payloadCommit)).WithField("error", resp.Message).Error("web3.storage upload error")
		}

		return "", errors.New(resp.Message)
	}
}

func (s *PayloadCommitService) storePayloadToCache(msg *datamodel.PayloadCommitMessage) error {
	cachePath := s.settingsObj.PayloadCachePath

	cachePath = fmt.Sprintf("%s%s/%s", cachePath, msg.ProjectID, msg.SnapshotCID)

	byteData, err := json.Marshal(msg)
	if err != nil {
		log.WithError(err).Error("failed to marshal payload commit message")

		return err
	}

	err = s.diskCache.Write(cachePath, byteData)
	if err != nil {
		log.WithError(err).Error("failed to write payload commit message to cache")

		return err
	}

	return nil
}

func (s *PayloadCommitService) signPayload() (*apitypes.TypedData, []byte, error) {
	privKey, err := signer.GetPrivateKey(s.settingsObj.Signer.PrivateKey)
	if err != nil {
		log.WithError(err).Error("failed to get ecdsa private key")

		return nil, nil, err
	}

	signerData, err := signer.GetSignerData(s.ethClient)
	if err != nil {
		log.WithError(err).Error("failed to get signer data")

		return nil, nil, err
	}

	signature, err := signer.SignMessage(privKey, signerData)
	if err != nil {
		log.WithError(err).Error("failed to sign message")

		return nil, nil, err
	}

	verified := signer.VerifySignature(signature, signerData)
	if !verified {
		log.Error("failed to verify signature")

		return nil, nil, errors.New("failed to verify signature")
	}

	return signerData, signature, nil
}

func (s *PayloadCommitService) sendSignatureToRelayer(payload interface{}) {
	panic("implement me")
}

// initLocalCachedData fills data in redis cache from smart contract if not present locally
func (s *PayloadCommitService) initLocalCachedData() error {
	projects, err := s.redisCache.GetStoredProjects(context.Background())
	if err != nil {
		log.WithError(err).Error("failed to get stored projects from redis cache")

		return err
	}

	if len(projects) > 0 {
		return nil
	}

	// get projects from smart contract
	projects, err = s.contractApi.GetProjects(&bind.CallOpts{})
	if err != nil {
		log.WithError(err).Error("failed to get projects from smart contract")

		return err
	}

	// store projects in redis cache
	err = s.redisCache.StoreProjects(context.Background(), projects)
	if err != nil {
		log.WithError(err).Error("failed to store projects in redis cache")

		return err
	}

	return nil
}
