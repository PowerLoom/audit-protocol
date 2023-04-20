package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/caching"
	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/token-aggregator/models"
)

const pairContractListFile string = "/static/cached_pair_addresses.json"
const tokenSummaryProjectID string = "uniswap_V2TokensSummarySnapshot_%s"
const pairSummaryProjectID string = "uniswap_V2PairsSummarySnapshot_%s"
const dailyStatsSummaryProjectID string = "uniswap_V2DailyStatsSnapshot_%s"

var lastSnapshotBlockHeight int64
var auditProtocolBaseURL string

type TokenDataRefs struct {
	token0Ref *models.TokenData
	token1Ref *models.TokenData
}

type TokenAggregator struct {
	settingsObj *settings.SettingsObj
	redisCache  *caching.RedisCache

	// defaultHTTPClient is retryable http client
	defaultHTTPClient *retryablehttp.Client
	pairContracts     []string
	tokenList         map[string]*models.TokenData
	tokenListLock     *sync.Mutex

	// projectLocksMap is map of projectID to lock
	projectLocksMap map[string]*sync.Mutex

	// projectLocksMapLock is lock for projectLocksMap
	projectLocksMapLock *sync.Mutex

	// tokenPairMapping is map of token pair to token data references
	tokenPairMapping map[string]*TokenDataRefs

	// tokenPairMappingLock is lock for tokenPairMapping
	tokenPairMappingLock *sync.Mutex
}

// InitTokenAggService initializes the token aggregator service
func InitTokenAggService() *TokenAggregator {
	settingsObj, err := gi.Invoke[*settings.SettingsObj]()
	if err != nil {
		log.WithError(err).Fatal("error while getting settings object")
	}

	redisCache, err := gi.Invoke[*caching.RedisCache]()
	if err != nil {
		log.WithError(err).Fatal("error while getting redis cache")
	}

	tokenAggregator := &TokenAggregator{
		settingsObj:          settingsObj,
		redisCache:           redisCache,
		defaultHTTPClient:    httpclient.GetDefaultHTTPClient(),
		pairContracts:        make([]string, 0),
		tokenList:            make(map[string]*models.TokenData, 0),
		tokenListLock:        new(sync.Mutex),
		projectLocksMap:      make(map[string]*sync.Mutex, 0),
		projectLocksMapLock:  new(sync.Mutex),
		tokenPairMapping:     make(map[string]*TokenDataRefs, 0),
		tokenPairMappingLock: new(sync.Mutex),
	}

	err = gi.Inject(tokenAggregator)
	if err != nil {
		log.WithError(err).Fatal("error while injecting token aggregator")
	}

	// audit protocol base url
	hostPort := net.JoinHostPort(settingsObj.TokenAggregatorSettings.APHost, strconv.Itoa(settingsObj.APBackend.Port))
	auditProtocolBaseURL = fmt.Sprintf("http://%s", hostPort)

	return tokenAggregator
}

// Run starts the token aggregator service
func (s *TokenAggregator) Run(pairContractAddress string) {
	log.Info("starting token aggregator service")

	// register callback keys
	s.RegisterAggregatorCallbackKey()

	// populate pair contract list
	// stops the service if encounters any error in this call
	s.pairContracts = s.PopulatePairContractList(pairContractAddress)

	var pairsSummaryBlockHeight int64

	for {
		log.Info("waiting for first Pairs Summary snapshot to be formed...")

		pairsSummaryBlockHeight = s.redisCache.FetchPairsSummaryLatestBlockHeight(context.Background(), s.settingsObj.PoolerNamespace)
		if pairsSummaryBlockHeight != 0 {
			log.Infof("pairsSummary snapshot has been created at height %d", pairsSummaryBlockHeight)

			break
		}

		log.Infof("pairsSummary snapshot has not been created yet, sleeping for %d secs", s.settingsObj.TokenAggregatorSettings.RunIntervalSecs)

		time.Sleep(time.Duration(s.settingsObj.TokenAggregatorSettings.RunIntervalSecs) * time.Second)
	}

	s.FetchTokensMetaData()

	// run in infinite loop
	for {
		_ = s.PrepareAndSubmitTokenSummarySnapshot()

		log.Infof("Sleeping for %d secs", s.settingsObj.TokenAggregatorSettings.RunIntervalSecs)
		time.Sleep(time.Duration(s.settingsObj.TokenAggregatorSettings.RunIntervalSecs) * time.Second)
	}
}

// RegisterAggregatorCallbackKey registers callback key in redis for all summary projects
func (s *TokenAggregator) RegisterAggregatorCallbackKey() {
	log.Debug("registering aggregator callback key")

	tokenSummaryProjectId := fmt.Sprintf(tokenSummaryProjectID, s.settingsObj.PoolerNamespace)
	pairSummaryProjectId := fmt.Sprintf(pairSummaryProjectID, s.settingsObj.PoolerNamespace)
	dailyStatsSummaryProjectId := fmt.Sprintf(dailyStatsSummaryProjectID, s.settingsObj.PoolerNamespace)

	body, _ := json.Marshal(map[string]string{
		"callbackURL": fmt.Sprintf("http://localhost:%d/block_height_confirm_callback", s.settingsObj.TokenAggregatorSettings.Port),
	})

	tokenSummaryURL := fmt.Sprintf("%s/%s/confirmations/callback", auditProtocolBaseURL, tokenSummaryProjectId)
	pairSummaryURL := fmt.Sprintf("%s/%s/confirmations/callback", auditProtocolBaseURL, pairSummaryProjectId)
	dailyStatsURL := fmt.Sprintf("%s/%s/confirmations/callback", auditProtocolBaseURL, dailyStatsSummaryProjectId)

	wg := new(sync.WaitGroup)
	wg.Add(3)

	go s.SetCallbackKeyInRedis(wg, tokenSummaryURL, tokenSummaryProjectId, body)
	go s.SetCallbackKeyInRedis(wg, pairSummaryURL, pairSummaryProjectId, body)
	go s.SetCallbackKeyInRedis(wg, dailyStatsURL, dailyStatsSummaryProjectId, body)

	wg.Wait()
}

// SetCallbackKeyInRedis sets the callback key in redis via the audit protocol backend
func (s *TokenAggregator) SetCallbackKeyInRedis(wg *sync.WaitGroup, callbackUrl string, projectId string, payload []byte) {
	defer wg.Done()
	resp, err := s.defaultHTTPClient.Post(callbackUrl, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		log.WithError(err).
			WithField("callbackUrl", callbackUrl).
			Error("failed to register callback url")
	}

	if resp.StatusCode == http.StatusOK {
		log.WithField("callbackUrl", callbackUrl).
			Info("successfully registered callback url")
	} else {
		log.WithField("callbackUrl", callbackUrl).
			Error("failed to register callback url")
	}

	// lock the map before project lock as this method is called concurrently
	s.projectLocksMapLock.Lock()
	// add projectId to callback lock in map
	s.projectLocksMap[projectId] = &sync.Mutex{}
	s.projectLocksMapLock.Unlock()
}

// FetchTokensMetaData fetches token metadata from redis and fills it in the tokenList,
func (s *TokenAggregator) FetchTokensMetaData() {
	wg := new(sync.WaitGroup)

	for _, contract := range s.pairContracts {
		wg.Add(1)

		pairContractAddr := common.HexToAddress(contract).Hex()

		go func(pairContractAddr string) {
			defer wg.Done()

			err := s.FetchAndFillTokenMetaData(pairContractAddr)
			if err != nil {
				log.WithField("address", pairContractAddr).WithError(err).Error("failed to fetch token metadata")
			}
		}(pairContractAddr)
	}

	wg.Wait()
}

// FetchAndFillTokenMetaData fetches token metadata from redis and fills it in the tokenList
func (s *TokenAggregator) FetchAndFillTokenMetaData(pairContractAddr string) error {
	pairContractMetadata, err := s.redisCache.FetchPairTokenMetadata(context.Background(), s.settingsObj.PoolerNamespace, pairContractAddr)
	if err != nil {
		log.WithError(err).Error("failed to fetch token metadata")

		return err
	}

	pairContractAddresses, err := s.redisCache.FetchPairTokenAddresses(context.Background(), s.settingsObj.PoolerNamespace, pairContractAddr)
	if err != nil {
		log.WithError(err).Error("failed to fetch token addresses")

		return err
	}

	token0Addr := pairContractAddresses.Token0Address
	token1Addr := pairContractAddresses.Token1Address

	tokenRefs := new(TokenDataRefs)

	// FIX: TOKEN Symbol and name not getting stored in tokenData.
	s.tokenListLock.Lock()
	if _, ok := s.tokenList[token0Addr]; !ok {
		tokenData := new(models.TokenData)

		tokenData.Symbol = pairContractMetadata.Token0Symbol
		tokenData.Name = pairContractMetadata.Token0Name
		tokenData.ContractAddress = token0Addr

		s.tokenList[token0Addr] = tokenData

		tokenRefs.token0Ref = tokenData
		log.WithField("token0", tokenData).Debug("token0 Data")
	}

	tokenRefs.token0Ref = s.tokenList[token0Addr]

	if _, ok := s.tokenList[token1Addr]; !ok {
		tokenData := new(models.TokenData)

		tokenData.Symbol = pairContractMetadata.Token1Symbol
		tokenData.Name = pairContractMetadata.Token1Name
		tokenData.ContractAddress = token1Addr

		s.tokenList[token1Addr] = tokenData

		tokenRefs.token1Ref = tokenData
		log.WithField("token1", tokenData).Debug("token1 Data")
	}

	tokenRefs.token1Ref = s.tokenList[token1Addr]
	s.tokenListLock.Unlock()

	s.tokenPairMappingLock.Lock()
	s.tokenPairMapping[pairContractAddr] = tokenRefs
	s.tokenPairMappingLock.Unlock()

	return nil
}

func (s *TokenAggregator) PrepareAndSubmitTokenSummarySnapshot() error {
	defer func() {
		if err := recover(); err != nil {
			log.WithField("recover", err).Error("recovered from panic")
		}
	}()

	curBlockHeight := s.redisCache.FetchPairsSummaryLatestBlockHeight(context.Background(), s.settingsObj.PoolerNamespace)
	dagChainProjectID := fmt.Sprintf(tokenSummaryProjectID, s.settingsObj.PoolerNamespace)

	if curBlockHeight <= lastSnapshotBlockHeight {
		log.
			Debugf("pairSummary blockHeight has not moved yet and is still at %d, lastSnapshotBlockHeight is %d. Hence not processing anything.",
				curBlockHeight, lastSnapshotBlockHeight)

		return nil
	}

	var sourceBlockHeight int64

	tokensPairData, err := s.redisCache.FetchPairSummarySnapshot(context.Background(), curBlockHeight, s.settingsObj.PoolerNamespace)
	if err != nil {
		log.WithError(err).Error("failed to fetch pairSummary snapshot")
		return err
	}

	if tokensPairData == nil {
		log.WithField("blockHeight", curBlockHeight).Info("no tokenData found")

		return nil
	}

	log.Debugf("collating tokenData at blockHeight %d", curBlockHeight)

	for _, tokenPairProcessedData := range tokensPairData {
		if tokenPairProcessedData == nil || tokenPairProcessedData.ContractAddress == "" {
			continue
		}

		// TODO: Need to remove 0x from contractAddress saved as string.
		ref, ok := s.tokenPairMapping[common.HexToAddress(tokenPairProcessedData.ContractAddress).Hex()]
		if !ok {
			continue
		}

		token0Data := ref.token0Ref
		token1Data := ref.token1Ref

		token0Data.Liquidity += tokenPairProcessedData.Token0Liquidity
		token0Data.LiquidityUSD += tokenPairProcessedData.Token0LiquidityUSD

		token1Data.Liquidity += tokenPairProcessedData.Token1Liquidity
		token1Data.LiquidityUSD += tokenPairProcessedData.Token1LiquidityUSD

		token0Data.TradeVolume24h += tokenPairProcessedData.Token0TradeVolume24h
		token0Data.TradeVolumeUSD24h += tokenPairProcessedData.Token0TradeVolumeUSD24h

		token1Data.TradeVolume24h += tokenPairProcessedData.Token1TradeVolume24h
		token1Data.TradeVolumeUSD24h += tokenPairProcessedData.Token1TradeVolumeUSD24h

		token0Data.TradeVolume7d += tokenPairProcessedData.Token0TradeVolume7d
		token0Data.TradeVolumeUSD7d += tokenPairProcessedData.Token0TradeVolumeUSD7d

		token1Data.TradeVolume7d += tokenPairProcessedData.Token1TradeVolume7d
		token1Data.TradeVolumeUSD7d += tokenPairProcessedData.Token1TradeVolumeUSD7d

		token0Data.BlockHeight = tokenPairProcessedData.BlockHeight
		token1Data.BlockHeight = tokenPairProcessedData.BlockHeight

		token0Data.BlockTimestamp = tokenPairProcessedData.BlockTimestamp
		token1Data.BlockTimestamp = tokenPairProcessedData.BlockTimestamp
		sourceBlockHeight = int64(tokenPairProcessedData.BlockHeight)
	}

	tm, err := strconv.ParseInt(fmt.Sprint(tokensPairData[0].BlockTimestamp), 10, 64)
	if err != nil {
		log.Errorf("failed to parse current timestamp int %s due to error %s", tokensPairData[0].BlockTimestamp, err.Error())

		return err
	}

	currentTimestamp := time.Unix(tm, 0)
	toTime := float64(currentTimestamp.Unix())

	// TODO: Make this logic more generic to support diffrent time based indexes.
	time24hPast := currentTimestamp.AddDate(0, 0, -1)
	fromTime := float64(time24hPast.Unix())

	log.Debug("timeStamp for 1 day before is: ", fromTime)

	// update tokenPrice
	beginBlockHeight24h := 0
	beginTimeStamp24h := 0.0

	s.tokenListLock.Lock()
	for key, tokenData := range s.tokenList {
		tokenData.Price, err = s.redisCache.FetchTokenPriceAtBlockHeight(context.Background(), tokenData.ContractAddress, int64(tokenData.BlockHeight), s.settingsObj.PoolerNamespace)
		if err != nil {
			delete(s.tokenList, key)

			continue
		}

		if tokenData.Price != 0 {
			// update token price in history zset
			err = s.redisCache.UpdateTokenPriceHistoryInRedis(context.Background(), toTime, fromTime, tokenData, s.settingsObj.PoolerNamespace)
			if err != nil {
				continue
			}

			curTimeEpoch := float64(time.Now().Unix())

			priceHistory, err := s.redisCache.FetchTokenPriceHistoryInRedis(context.Background(), fromTime, curTimeEpoch, tokenData.ContractAddress, s.settingsObj.PoolerNamespace)
			if err != nil {
				s.tokenListLock.Unlock()
				return err
			}

			// TODO: Need to add validation if value is newer than x hours, should we still show as priceChange?
			oldPrice := priceHistory.Price
			tokenData.PriceChangePercent24h = (tokenData.Price - oldPrice) * 100 / tokenData.Price

			if beginBlockHeight24h == 0 {
				beginBlockHeight24h = priceHistory.BlockHeight
				beginTimeStamp24h = priceHistory.Timestamp
			}
		} else {
			// TODO: Should we create a snapshot if we don't have any tokenPrice at specified height?
			log.Errorf("Price couldn't be retrieved for token %s with name %s at blockHeight %d hence removing token from the list.",
				key, tokenData.Name, tokenData.BlockHeight)
		}
	}
	s.tokenListLock.Unlock()

	err = s.CommitTokenSummaryPayload()
	if err != nil {
		log.WithField("blockHeight", curBlockHeight).
			WithError(err).
			Errorf("failed to commit payload")

		s.ResetTokenData()

		return err
	}

	// fetch lastTokenSummaryBlockHeight for the project
	lastTokenSummaryBlockHeight, err := s.redisCache.FetchTokenSummaryLatestBlockHeight(context.Background(), s.settingsObj.PoolerNamespace)
	if err != nil {
		return err
	}

	tentativeBlockHeight := lastTokenSummaryBlockHeight + 1

	tokenSummarySnapshotMeta := new(models.TokenSummarySnapshotMeta)

	err = backoff.Retry(func() error {
		tokenSummarySnapshotMeta, err = s.WaitAndFetchBlockHeightStatus(dagChainProjectID, tentativeBlockHeight)

		return err
	}, backoff.NewExponentialBackOff())
	if err != nil {
		log.WithError(err).Errorf("failed to fetch payloadCID at blockHeight %d", tentativeBlockHeight)
		s.ResetTokenData()

		return err
	}

	tokenSummarySnapshotMeta.BeginBlockHeight24h = int64(beginBlockHeight24h)
	tokenSummarySnapshotMeta.BeginBlockheightTimeStamp24h = beginTimeStamp24h

	err = s.redisCache.StoreTokenSummaryCIDInSnapshotsZSet(context.Background(), sourceBlockHeight, s.settingsObj.PoolerNamespace, tokenSummarySnapshotMeta)
	if err != nil {
		log.WithError(err).Errorf("failed to store token summary CID at blockHeight %d", sourceBlockHeight)
		s.ResetTokenData()

		return err
	}

	err = s.redisCache.StoreTokensSummaryPayload(context.Background(), sourceBlockHeight, s.settingsObj.PoolerNamespace, s.tokenList)
	if err != nil {
		log.WithError(err).Errorf("failed to store token summary payload at blockHeight %d", sourceBlockHeight)
		s.ResetTokenData()

		return err
	}

	s.ResetTokenData()

	lastSnapshotBlockHeight = curBlockHeight

	// prune TokenPrice ZSet as price already fetched for all tokens
	s.tokenListLock.Lock()
	for _, tokenData := range s.tokenList {
		err = s.redisCache.PruneTokenPriceZSet(context.Background(), tokenData.ContractAddress, int64(tokenData.BlockHeight), s.settingsObj.PoolerNamespace)
		if err != nil {
			log.WithError(err).Error("failed to prune price zset")
		}
	}
	s.tokenListLock.Unlock()

	return nil
}

func (s *TokenAggregator) FetchAndUpdateStatusOfOlderSnapshots(projectId string) error {
	// Fetch all entries in snapshotZSet
	// Any entry that has a txStatus as TX_CONFIRM_PENDING, query its updated status and update ZSet
	// If txHash changes, store old one in prevTxhash and update the new one in txHash

	projectLock := s.projectLocksMap[projectId]

	projectLock.Lock()
	defer projectLock.Unlock()

	var redisAggregatorProjectId string
	poolerNamespace := s.settingsObj.PoolerNamespace

	switch projectId {
	case fmt.Sprintf(tokenSummaryProjectID, poolerNamespace):
		redisAggregatorProjectId = fmt.Sprintf(redisutils.REDIS_KEY_TOKENS_SUMMARY_SNAPSHOTS_ZSET, poolerNamespace)

	case fmt.Sprintf(pairSummaryProjectID, poolerNamespace):
		redisAggregatorProjectId = fmt.Sprintf(redisutils.REDIS_KEY_PAIRS_SUMMARY_SNAPSHOTS_ZSET, poolerNamespace)

	case fmt.Sprintf(dailyStatsSummaryProjectID, poolerNamespace):
		redisAggregatorProjectId = fmt.Sprintf(redisutils.REDIS_KEY_DAILY_STATS_SUMMARY_SNAPSHOTS_ZSET, poolerNamespace)
	}

	key := redisAggregatorProjectId
	log.Debugf("checking and updating status of older blockHeight entries in snapshotsZset")

	// fetch all entries in snapshotZSet
	snapshots, err := s.redisCache.FetchSummaryProjectSnapshots(context.Background(), key, "-inf", "+inf")
	if err != nil {
		log.Errorf("failed to fetch snapshots for project %s", projectId)

		return err
	}

	for _, snapshotMeta := range snapshots {
		if snapshotMeta.DAGHeight == 0 {
			// skip processing of blockHeight snapshots if DAGheight is not available to fetch status.
			continue
		}

		if snapshotMeta.TxStatus <= models.TX_CONFIRMATION_PENDING {
			updatedSnapshotMeta := new(models.TokenSummarySnapshotMeta)

			// fetch updated status.
			err = backoff.Retry(func() error {
				updatedSnapshotMeta, err = s.WaitAndFetchBlockHeightStatus(projectId, int64(snapshotMeta.DAGHeight))

				return err
			}, backoff.NewExponentialBackOff())
			if err != nil {
				log.WithError(err).Errorf("failed to fetch payloadCID at blockHeight %d", snapshotMeta.DAGHeight)

				continue
			}

			if snapshotMeta.TxHash != updatedSnapshotMeta.TxHash {
				snapshotMeta.PrevTxHash = snapshotMeta.TxHash
				snapshotMeta.TxHash = updatedSnapshotMeta.TxHash
			}

			snapshotMeta.TxStatus = updatedSnapshotMeta.TxStatus

			// once new snapshot is prepared then only delete the Zset Entry
			err = s.redisCache.RemoveOlderSnapshot(context.Background(), key, snapshotMeta)
			if err != nil {
				log.WithError(err).Errorf("failed to remove older snapshot from Zset for projectId %s", projectId)

				continue
			}

			err = s.redisCache.AddSnapshot(context.Background(), key, snapshotMeta.DAGHeight, updatedSnapshotMeta)
			if err != nil {
				log.WithError(err).Errorf("failed to add new snapshot to Zset for projectId %s", projectId)

				continue
			}
		}
	}

	log.Debugf("Updated old snapshot txHashs status!")

	return nil
}

func (s *TokenAggregator) ResetTokenData() {
	s.tokenListLock.Lock()
	for _, tokenData := range s.tokenList {
		tokenData.Liquidity = 0
		tokenData.LiquidityUSD = 0
		tokenData.TradeVolume24h = 0
		tokenData.TradeVolumeUSD24h = 0
		tokenData.TradeVolume7d = 0
		tokenData.TradeVolumeUSD7d = 0
	}
	s.tokenListLock.Unlock()
}

// CommitTokenSummaryPayload commits the token summary payload to the audit-protocol.
func (s *TokenAggregator) CommitTokenSummaryPayload() error {
	url := auditProtocolBaseURL + "/commit_payload"
	request := new(models.AuditProtocolCommitPayloadReq)

	request.ProjectId = fmt.Sprintf(tokenSummaryProjectID, s.settingsObj.PoolerNamespace)
	request.Payload.TokensData = make([]*models.TokenData, len(s.tokenList))
	request.Web3Storage = true // Always store TokenData snapshot in web3.storage.
	request.SkipAnchorProof = s.settingsObj.ContractCallBackend.SkipSummaryProjectProof

	var index int

	s.tokenListLock.Lock()
	for _, tokenData := range s.tokenList {
		request.Payload.TokensData[index] = tokenData

		index += 1
	}
	s.tokenListLock.Unlock()

	body, err := json.Marshal(request)
	if err != nil {
		log.WithError(err).
			WithField("request", request).
			Error("failed to marshal request")

		return err
	}

	resp, err := s.defaultHTTPClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.WithError(err).
			Error("failed not send commit-payload request to audit-protocol")
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.WithError(err).Error("failed to close response body")
		}
	}(resp.Body)

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Error("unable to read HTTP resp from audit-protocol for commit-payload")
	}

	log.WithField("respBody", string(body)).Info("received response")

	if resp.StatusCode != http.StatusOK {
		errorResp := new(models.AuditProtocolErrorResp)

		if err = json.Unmarshal(body, errorResp); err != nil {
			log.WithError(err).
				Error("failed unmarshal error JSON response received from audit-protocol")

			return err
		}

		log.WithField("statusCode", resp.Status).
			WithField("resp", errorResp).
			Error("failed to commit payload to audit-protocol")

		return nil
	}

	apCommitResp := new(models.AuditProtocolCommitPayloadResp)

	if err = json.Unmarshal(body, apCommitResp); err != nil {
		log.WithError(err).Error("failed to unmarshal JSON response received from audit-protocol")
	}

	log.WithField("tentativeHeight", apCommitResp.TentativeHeight).
		WithField("commitID", apCommitResp.CommitID).
		Debug("successfully committed payload to audit-protocol")

	return nil
}

// WaitAndFetchBlockHeightStatus waits for the blockHeight to be committed to the audit-protocol and returns the status.
func (s *TokenAggregator) WaitAndFetchBlockHeightStatus(projectID string, blockHeight int64) (*models.TokenSummarySnapshotMeta, error) {
	url := fmt.Sprintf("%s/%s/payload/%d/status", auditProtocolBaseURL, projectID, blockHeight)

	log.WithField("reqURL", url).Debug("fetching CID at blockHeight")

	resp, err := s.defaultHTTPClient.Get(url)
	if err != nil {
		log.WithError(err).WithField("reqURL", url).Error("failed to send request to audit-protocol")

		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.WithError(err).Error("failed to close response body")
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Error("unable to read HTTP resp from audit-protocol")

		return nil, err
	}

	if resp.StatusCode == http.StatusBadRequest {
		log.Debugf("snapshot for block at height %d not yet ready", blockHeight)

		return nil, errors.New("snapshot not yet ready")
	}

	apResp := new(models.AuditProtocolBlockHeightStatusResp)
	if err = json.Unmarshal(body, apResp); err != nil { // Parse []byte to the go struct pointer
		log.WithError(err).Error("failed to unmarshal JSON response received from audit-protocol")

		return nil, err
	}

	log.WithField("reqUrl", url).WithField("resp", apResp).Debug("successfully received response")

	if apResp.Status < models.TX_CONFIRMATION_PENDING {
		log.Debugf("blockHeight %d status is still pending with status code %d", blockHeight, apResp.Status)

		return nil, errors.New("blockHeight status is still pending")
	}

	log.WithField("payloadCid", apResp.PayloadCid).
		WithField("txHash", apResp.TxHash).
		WithField("blockHeight", blockHeight).
		WithField("projectID", projectID).
		Debug("received CID for given blockHeight")

	tokenSummarySnapshotMeta := &models.TokenSummarySnapshotMeta{
		Cid:       apResp.PayloadCid,
		TxHash:    apResp.TxHash,
		TxStatus:  apResp.Status,
		DAGHeight: apResp.BlockHeight,
	}

	return tokenSummarySnapshotMeta, nil
}

// PopulatePairContractList reads the pair-contracts.json file and populates the pairContracts slice
// service will stop on error
func (s *TokenAggregator) PopulatePairContractList(pairContractAddress string) []string {
	pairContracts := make([]string, 0)

	if pairContractAddress != "" {
		log.Info("skipping reading contract addresses from json.Considering only passed pairContractAddress:", pairContractAddress)

		pairContracts = append(pairContracts, pairContractAddress)
	}

	pairContractsPath := os.Getenv("CONFIG_PATH") + pairContractListFile
	log.Info("reading contracts:", pairContractsPath)

	data, err := os.ReadFile(pairContractsPath)
	if err != nil {
		log.WithError(err).Fatal("unable to read cache_pair_addresses.json file")
	}

	log.Debug("contracts json data is", string(data))

	err = json.Unmarshal(data, &pairContracts)
	if err != nil {
		log.Error("cannot unmarshal the pair-contracts json ", err)
		log.WithError(err).Fatal()
	}

	return pairContracts
}
