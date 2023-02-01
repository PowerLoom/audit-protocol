package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"

	"github.com/go-redis/redis/v8"
	"github.com/powerloom/goutils/logger"
	"github.com/powerloom/goutils/redisutils"
	"github.com/powerloom/goutils/settings"
)

var settingsObj *settings.SettingsObj
var redisClient *redis.Client
var pairContracts []string
var tokenList map[string]*TokenData
var apHttpClient http.Client
var lastSnapshotBlockHeight int64
var auditProtocolBaseURL string
var tokenPairTokenMapping map[string]TokenDataRefs
var snapshotCallbackProjectLocks map[string]*sync.Mutex

type TokenDataRefs struct {
	token0Ref *TokenData
	token1Ref *TokenData
}

const settingsFile string = "../settings.json"
const pairContractListFile string = "../static/cached_pair_addresses.json"
const TOKENSUMMARY_PROJECTID string = "uniswap_V2TokensSummarySnapshot_%s"
const PAIRSUMMARY_PROJECTID string = "uniswap_V2PairsSummarySnapshot_%s"
const DAILYSTATSSUMMARY_PROJECTID string = "uniswap_V2DailyStatsSnapshot_%s"
const MAX_RETRIES_BEFORE_EXIT int = 10

var retryInterval int
var ctx = context.Background()

func main() {
	// first read config settings
	logger.InitLogger()

	ReadSettings()
	auditProtocolBaseURL = fmt.Sprintf("http://%s:%d", settingsObj.TokenAggregatorSettings.APHost, settingsObj.APBackend.Port)
	var pairContractAddressesFile string
	snapshotCallbackProjectLocks = make(map[string]*sync.Mutex)

	http.HandleFunc("/block_height_confirm_callback", BlockHeightConfirmCallback)
	port := settingsObj.TokenAggregatorSettings.Port
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Infof("Starting HTTP server on port %d in a go routine.", port)
		http.ListenAndServe(fmt.Sprint(":", port), nil)
	}()

	if len(os.Args) == 3 {
		pairContractAddressesFile = os.Args[2]
	}

	RegisterAggregatorCallbackKey()
	SetupRedisClient()
	InitAuditProtocolClient()
	tokenList = make(map[string]*TokenData)
	tokenPairTokenMapping = make(map[string]TokenDataRefs)
	Run(pairContractAddressesFile)
	wg.Wait()
}

func Run(pairContractAddress string) {
	PopulatePairContractList(pairContractAddress)
	var pairsSummaryBlockHeight int64
	for {
		log.Info("Waiting for first Pairs Summary snapshot to be formed...")
		pairsSummaryBlockHeight = FetchPairsSummaryLatestBlockHeight()
		if pairsSummaryBlockHeight != 0 {
			log.Infof("PairsSummary snapshot has been created at height %d", pairsSummaryBlockHeight)
			break
		}
		time.Sleep(time.Duration(settingsObj.TokenAggregatorSettings.RunIntervalSecs) * time.Second)
	}

	FetchTokensMetaData()

	for {
		PrepareAndSubmitTokenSummarySnapshot()

		log.Infof("Sleeping for %d secs", settingsObj.TokenAggregatorSettings.RunIntervalSecs)
		time.Sleep(time.Duration(settingsObj.TokenAggregatorSettings.RunIntervalSecs) * time.Second)
	}
}

func BlockHeightConfirmCallback(w http.ResponseWriter, req *http.Request) {
	log.Infof("Received block height confirm callback %+v : ", *req)
	reqBytes, _ := io.ReadAll(req.Body)
	var reqPayload BlockHeightConfirmationPayload

	err := json.Unmarshal(reqBytes, &reqPayload)
	if err != nil {
		log.Errorf("Error while parsing json body of callback confirmation %s", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := make(map[string]string)
	resp["message"] = "Callback Recieved"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Callback Confirmation: Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)

	go func() {
		FetchAndUpdateStatusOfOlderSnapshots(reqPayload.ProjectId)
	}()
}

func RegisterAggregatorCallbackKey() {
	tokenSummaryProjectId := fmt.Sprintf(TOKENSUMMARY_PROJECTID, settingsObj.PoolerNamespace)
	pairSummaryProjectId := fmt.Sprintf(PAIRSUMMARY_PROJECTID, settingsObj.PoolerNamespace)
	dailyStatsSummaryProjectId := fmt.Sprintf(DAILYSTATSSUMMARY_PROJECTID, settingsObj.PoolerNamespace)

	body, _ := json.Marshal(map[string]string{
		"callbackURL": fmt.Sprintf("http://localhost:%d/block_height_confirm_callback", settingsObj.TokenAggregatorSettings.Port),
	})

	tokenSummaryUrl := fmt.Sprintf("%s/%s/confirmations/callback", auditProtocolBaseURL, tokenSummaryProjectId)
	pairSummaryUrl := fmt.Sprintf("%s/%s/confirmations/callback", auditProtocolBaseURL, pairSummaryProjectId)
	dailyStatsUrl := fmt.Sprintf("%s/%s/confirmations/callback", auditProtocolBaseURL, dailyStatsSummaryProjectId)

	SetCallbackKeyInRedis(tokenSummaryUrl, tokenSummaryProjectId, body)
	SetCallbackKeyInRedis(pairSummaryUrl, pairSummaryProjectId, body)
	SetCallbackKeyInRedis(dailyStatsUrl, dailyStatsSummaryProjectId, body)
}

func SetCallbackKeyInRedis(callbackUrl string, projectId string, payload []byte) {

	for retryCount := 0; retryCount < 3; retryCount++ {
		resp, err := apHttpClient.Post(callbackUrl, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			log.Errorf("Failed to register callback url due to error %+v | url:%s", err.Error(), callbackUrl)
			time.Sleep(3 * time.Second)
			continue
		}
		log.Debugf("Registered callback response: %s | url:%s", resp, callbackUrl)

		// add projectId to callback lock
		var projecIdLock sync.Mutex
		snapshotCallbackProjectLocks[projectId] = &projecIdLock
		log.Debugf("Created lock for callbacks %s", projectId)
		break
	}
}

func FetchTokensMetaData() {
	for i := range pairContracts {
		pairContractAddress := pairContracts[i]
		pairContractAddr := common.HexToAddress(pairContractAddress).Hex()
		FetchAndFillTokenMetaData(pairContractAddr)
	}
}

func FetchAndFillTokenMetaData(pairContractAddr string) {
	redisKey := fmt.Sprintf(redisutils.REDIS_KEY_TOKEN_PAIR_CONTRACT_TOKENS_DATA, settingsObj.PoolerNamespace, pairContractAddr)
	log.Debug("Fetching PairContractTokensData from redis with key:", redisKey)
	var tokenPairMeta map[string]string
	var err error
	for retryCount := 0; ; {
		tokenPairMeta, err = redisClient.HGetAll(ctx, redisKey).Result()
		if err != nil {
			retryCount++
			if retryCount > MAX_RETRIES_BEFORE_EXIT {
				log.Fatalf("Unable to fetch PairContractTokensData for %s key even after max retries %d . Hence exiting...",
					MAX_RETRIES_BEFORE_EXIT, redisKey)
				os.Exit(1)
			}
			log.Errorf("Failed to get tokenPair MetaData from redis for PairContract %s due to error %s, retrying %d",
				pairContractAddr, err.Error(), retryCount)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
		log.Debug("Fetched PairContractTokensData from redis:", tokenPairMeta)
		break
	}

	//Use tokenContractAddress to store tokenData as tokenSymbol is not gauranteed to be unique.
	var tokenContractAddresses map[string]string
	redisKey = fmt.Sprintf(redisutils.REDIS_KEY_PAIR_TOKEN_ADDRESSES, settingsObj.PoolerNamespace, pairContractAddr)
	for retryCount := 0; ; {
		tokenContractAddresses, err = redisClient.HGetAll(ctx, redisKey).Result()
		if err != nil {
			retryCount++
			if retryCount > MAX_RETRIES_BEFORE_EXIT {
				log.Fatalf("Unable to fetch PairContractTokensAddresses for %s key even after max retries %d. Hence exiting...",
					MAX_RETRIES_BEFORE_EXIT, redisKey)
				os.Exit(1)
			}
			log.Errorf("Failed to get PairContractTokensAddresses from redis for PairContract due to error %s, retrying %d",
				pairContractAddr, err.Error(), retryCount)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
		log.Debugf("Fetched PairContractTokensAddresses from redis %+v", tokenContractAddresses)
		break
	}
	token0Addr := tokenContractAddresses["token0Addr"]
	token1Addr := tokenContractAddresses["token1Addr"]
	var tokenRefs TokenDataRefs
	//FIX: TOKEN Symbol and name not getting stored in tokenData.
	if _, ok := tokenList[token0Addr]; !ok {
		var tokenData TokenData
		tokenData.Symbol = tokenPairMeta["token0_symbol"]
		tokenData.Name = tokenPairMeta["token0_name"]
		tokenData.ContractAddress = token0Addr
		tokenList[token0Addr] = &tokenData
		tokenRefs.token0Ref = &tokenData
		log.Debugf("Token0 Data : %+v", tokenData)
	}
	tokenRefs.token0Ref = tokenList[token0Addr]
	if _, ok := tokenList[token1Addr]; !ok {
		var tokenData TokenData
		tokenData.Symbol = tokenPairMeta["token1_symbol"]
		tokenData.Name = tokenPairMeta["token1_name"]
		tokenData.ContractAddress = token1Addr
		tokenList[token1Addr] = &tokenData
		tokenRefs.token1Ref = &tokenData
		log.Debugf("Token1 Data: %+v", tokenData)
	}
	tokenRefs.token1Ref = tokenList[token1Addr]
	tokenPairTokenMapping[pairContractAddr] = tokenRefs
}

func PrepareAndSubmitTokenSummarySnapshot() {

	curBlockHeight := FetchPairsSummaryLatestBlockHeight()
	dagChainProjectId := fmt.Sprintf(TOKENSUMMARY_PROJECTID, settingsObj.PoolerNamespace)

	if curBlockHeight > lastSnapshotBlockHeight {
		var sourceBlockHeight int64
		tokensPairData := FetchPairSummarySnapshot(curBlockHeight)
		if tokensPairData == nil {
			return
		}
		log.Debugf("Collating tokenData at blockHeight %d", curBlockHeight)
		for _, tokenPairProcessedData := range tokensPairData {
			//TODO: Need to remove 0x from contractAddress saved as string.
			token0Data := tokenPairTokenMapping[common.HexToAddress(tokenPairProcessedData.ContractAddress).Hex()].token0Ref
			token1Data := tokenPairTokenMapping[common.HexToAddress(tokenPairProcessedData.ContractAddress).Hex()].token1Ref

			token0Data.Liquidity += tokenPairProcessedData.Token0Liquidity
			token0Data.LiquidityUSD += tokenPairProcessedData.Token0LiquidityUSD

			token1Data.Liquidity += tokenPairProcessedData.Token1Liquidity
			token1Data.LiquidityUSD += tokenPairProcessedData.Token1LiquidityUSD

			token0Data.TradeVolume_24h += tokenPairProcessedData.Token0TradeVolume_24h
			token0Data.TradeVolumeUSD_24h += tokenPairProcessedData.Token0TradeVolumeUSD_24h

			token1Data.TradeVolume_24h += tokenPairProcessedData.Token1TradeVolume_24h
			token1Data.TradeVolumeUSD_24h += tokenPairProcessedData.Token1TradeVolumeUSD_24h

			token0Data.TradeVolume_7d += tokenPairProcessedData.Token0TradeVolume_7d
			token0Data.TradeVolumeUSD_7d += tokenPairProcessedData.Token0TradeVolumeUSD_7d

			token1Data.TradeVolume_7d += tokenPairProcessedData.Token1TradeVolume_7d
			token1Data.TradeVolumeUSD_7d += tokenPairProcessedData.Token1TradeVolumeUSD_7d

			token0Data.Block_height = tokenPairProcessedData.Block_height
			token1Data.Block_height = tokenPairProcessedData.Block_height

			token0Data.Block_timestamp = tokenPairProcessedData.Block_timestamp
			token1Data.Block_timestamp = tokenPairProcessedData.Block_timestamp
			sourceBlockHeight = int64(tokenPairProcessedData.Block_height)
		}

		tm, err := strconv.ParseInt(fmt.Sprint(tokensPairData[0].Block_timestamp), 10, 64)
		if err != nil {
			log.Errorf("Failed to parse current timestamp int %s due to error %s", tokensPairData[0].Block_timestamp, err.Error())
			return
		}
		currentTimestamp := time.Unix(tm, 0)
		toTime := float64(currentTimestamp.Unix())
		//TODO: Make this logic more generic to support diffrent time based indexes.
		time24h := currentTimestamp.AddDate(0, 0, -1)
		fromTime := float64(time24h.Unix())
		log.Debug("TimeStamp for 1 day before is:", fromTime)
		//Fetch lastTokensummaryBlockHeight for the project
		lastTokensummaryBlockHeight := FetchTokenSummaryLatestBlockHeight()
		//Update tokenPrice
		beginBlockHeight24h := 0
		beginTimeStamp24h := 0.0
		for key, tokenData := range tokenList {
			tokenData.Price = FetchTokenPriceAtBlockHeight(tokenData.ContractAddress, int64(tokenData.Block_height))
			if tokenData.Price != 0 {
				//Update TokenPrice in History Zset
				UpdateTokenPriceHistoryRedis(toTime, fromTime, tokenData)

				tokenPrice24hEntry := CalculateAndFillPriceChange(fromTime, tokenData)
				if beginBlockHeight24h == 0 {
					beginBlockHeight24h = tokenPrice24hEntry.BlockHeight
					beginTimeStamp24h = tokenPrice24hEntry.Timestamp
				}
				//tokenList[key] = tokenData
			} else {
				//TODO: Should we create a snapshot if we don't have any tokenPrice at specified height?
				log.Errorf("Price couldn't be retrieved for token %s with name %s at blockHeight %d hence removing token from the list.",
					key, tokenData.Name, tokenData.Block_height)
				//delete(tokenList, key)
			}
		}
		err = CommitTokenSummaryPayload()
		if err != nil {
			log.Errorf("Failed to commit payload at blockHeight %d due to error %s", curBlockHeight, err.Error())
			ResetTokenData()
			return
		}
		tentativeBlockHeight := lastTokensummaryBlockHeight + 1
		tokenSummarySnapshotMeta, err := WaitAndFetchBlockHeightStatus(dagChainProjectId, tentativeBlockHeight, *settingsObj.RetryCount)
		if err != nil {
			log.Errorf("Failed to Fetch payloadCID at blockHeight %d due to error %s", tentativeBlockHeight, err.Error())
			ResetTokenData()
			return
		}
		tokenSummarySnapshotMeta.BeginBlockHeight24h = int64(beginBlockHeight24h)
		tokenSummarySnapshotMeta.BeginBlockheightTimeStamp24h = beginTimeStamp24h
		StoreTokenSummaryCIDInSnapshotsZSet(sourceBlockHeight, tokenSummarySnapshotMeta)
		StoreTokensSummaryPayload(sourceBlockHeight)
		ResetTokenData()
		lastSnapshotBlockHeight = curBlockHeight

		//Prune TokenPrice ZSet as price already fetched for all tokens
		for _, tokenData := range tokenList {
			PruneTokenPriceZSet(tokenData.ContractAddress, int64(tokenData.Block_height))
		}

	} else {
		log.Debugf("PairSummary blockHeight has not moved yet and is still at %d, lastSnapshotBlockHeight is %d. Hence not processing anything.",
			curBlockHeight, lastSnapshotBlockHeight)
	}
}

func FetchAndUpdateStatusOfOlderSnapshots(projectId string) error {
	// Fetch all entries in snapshotZSet
	//Any entry that has a txStatus as TX_CONFIRM_PENDING, query its updated status and update ZSet
	//If txHash changes, store old one in prevTxhash and update the new one in txHash

	projectLock := snapshotCallbackProjectLocks[projectId]

	projectLock.Lock()
	defer projectLock.Unlock()

	var redisAggregatorProjectId string
	switch projectId {
	case fmt.Sprintf(TOKENSUMMARY_PROJECTID, settingsObj.PoolerNamespace):
		redisAggregatorProjectId = fmt.Sprintf(
			redisutils.REDIS_KEY_TOKENS_SUMMARY_SNAPSHOTS_ZSET,
			settingsObj.PoolerNamespace)
	case fmt.Sprintf(PAIRSUMMARY_PROJECTID, settingsObj.PoolerNamespace):
		redisAggregatorProjectId = fmt.Sprintf(
			redisutils.REDIS_KEY_PAIRS_SUMMARY_SNAPSHOTS_ZSET,
			settingsObj.PoolerNamespace)
	case fmt.Sprintf(DAILYSTATSSUMMARY_PROJECTID, settingsObj.PoolerNamespace):
		redisAggregatorProjectId = fmt.Sprintf(
			redisutils.REDIS_KEY_DAILY_STATS_SUMMARY_SNAPSHOTS_ZSET,
			settingsObj.PoolerNamespace)
	}

	key := redisAggregatorProjectId
	log.Debugf("Checking and updating status of older blockHeight entries in snapshotsZset")
	res := redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{Min: "-inf", Max: "+inf"})
	if res.Err() != nil {
		if res.Err() == redis.Nil {
			log.Infof("No entries found in snapshotsZSet")
			return nil
		}
		log.Errorf("Failed to fetch entries from snapshotZSet. Retry in next cycle")
		return res.Err()
	}
	snapshotsMeta := res.Val()
	for i := range snapshotsMeta {
		var snapshotMeta TokenSummarySnapshotMeta
		snapshot := fmt.Sprintf("%v", snapshotsMeta[i].Member)
		err := json.Unmarshal([]byte(snapshot), &snapshotMeta)
		if err != nil {
			log.Errorf("Critical! Unable to unmarshal snapshot meta data")
			return err
		}
		if snapshotMeta.DAGHeight == 0 {
			//skip processing of blockHeight snapshots if DAGheight is not available to fetch status.
			continue
		}
		if snapshotMeta.TxStatus <= TX_CONFIRMATION_PENDING {

			//Fetch updated status.
			snapshotMetaNew, err := WaitAndFetchBlockHeightStatus(projectId, int64(snapshotMeta.DAGHeight), 3)
			if err != nil {
				log.Infof("Could not get blockheight status for TokensSummary at height %d", snapshotMeta.DAGHeight)
				continue
			}

			if snapshotMeta.TxHash != snapshotMetaNew.TxHash {
				snapshotMeta.PrevTxHash = snapshotMeta.TxHash
				snapshotMeta.TxHash = snapshotMetaNew.TxHash
			}
			snapshotMeta.TxStatus = snapshotMetaNew.TxStatus
			snapshotNew, err := json.Marshal(snapshotMeta)
			if err != nil {
				log.Errorf("CRITICAL! Json marshal failed for snapshotMeta %+v with error %+v", snapshotMetaNew, err)
				continue
			}

			// once new snapshot is prepared then only delete the Zset Entry
			res := redisClient.ZRem(ctx, key, snapshot)
			if res.Err() != nil {
				log.Errorf("Failed to remove snapshotsZset entry due to error %+v", res.Err())
				continue
			}

			for j := 0; j < 3; j++ {
				res = redisClient.ZAdd(ctx, key, &redis.Z{
					Score:  snapshotsMeta[i].Score,
					Member: snapshotNew,
				})
				if res.Err() != nil {
					log.Errorf("Failed to Add entry at score %f due to error %+v. Retrying", snapshotsMeta[i].Score, res.Err())
					time.Sleep(time.Duration(retryInterval) * time.Second)
					continue
				}
				break
			}

		}
	}

	log.Debugf("Updated old snapshot txHashs status!")

	return nil
}

func FetchTokenSummaryLatestBlockHeight() int64 {
	key := fmt.Sprintf(redisutils.REDIS_KEY_TOKENS_SUMMARY_TENTATIVE_HEIGHT, settingsObj.PoolerNamespace)
	for retryCount := 0; retryCount < 3; retryCount++ {
		res := redisClient.Get(ctx, key)
		if res.Err() != nil {
			if res.Err() == redis.Nil {
				log.Debugf("Waiting for tentativeblock height key to be created for Token Summary project")
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue
			}
			log.Errorf("Could not fetch tentativeblock height Error %+v", res.Err())
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}

		tentativeHeight, err := strconv.Atoi(res.Val())
		if err != nil {
			log.Errorf("CRITICAL! Unable to extract tentativeHeight from redis result due to err %+v", err)
			return 0
		}

		log.Debugf("Latest tentative block height for TokenSummary project is : %d", tentativeHeight)
		return int64(tentativeHeight)
	}
	return 0
}

func ResetTokenData() {
	for _, tokenData := range tokenList {
		tokenData.Liquidity = 0
		tokenData.LiquidityUSD = 0
		tokenData.TradeVolume_24h = 0
		tokenData.TradeVolumeUSD_24h = 0
		tokenData.TradeVolume_7d = 0
		tokenData.TradeVolumeUSD_7d = 0
	}
}

func CalculateAndFillPriceChange(fromTime float64, tokenData *TokenData) *TokenPriceHistoryEntry {
	curTimeEpoch := float64(time.Now().Unix())
	key := fmt.Sprintf(redisutils.REDIS_KEY_TOKEN_PRICE_HISTORY, settingsObj.PoolerNamespace, tokenData.ContractAddress)

	zRangeByScore := redisClient.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", fromTime),
		Max: fmt.Sprintf("%f", curTimeEpoch),
	})
	if zRangeByScore.Err() != nil {
		log.Error("Could not fetch entries error: ", zRangeByScore.Err().Error(), "fromTime:", fromTime)
		return nil
	}
	//Fetch the oldest Value closest to 24h
	var tokenPriceHistoryEntry TokenPriceHistoryEntry
	err := json.Unmarshal([]byte(zRangeByScore.Val()[0]), &tokenPriceHistoryEntry)
	if err != nil {
		log.Error("Unable to decode value fetched from Zset...something wrong!!")
		return nil
	}
	//TODO: Need to add validation if value is newer than x hours, should we still show as priceChange?
	oldPrice := tokenPriceHistoryEntry.Price
	tokenData.PriceChangePercent_24h = (tokenData.Price - oldPrice) * 100 / tokenData.Price
	return &tokenPriceHistoryEntry
}

func UpdateTokenPriceHistoryRedis(toTime float64, fromTime float64, tokenData *TokenData) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_TOKEN_PRICE_HISTORY, settingsObj.PoolerNamespace, tokenData.ContractAddress)
	var priceHistoryEntry TokenPriceHistoryEntry = TokenPriceHistoryEntry{toTime, tokenData.Price, tokenData.Block_height}
	val, err := json.Marshal(priceHistoryEntry)
	if err != nil {
		log.Error("Couldn't marshal json..something is really wrong with data.curTime:", toTime, " TokenData:", tokenData)
		return
	}
	err = redisClient.ZAdd(ctx, key, &redis.Z{
		Score:  float64(toTime),
		Member: string(val),
	}).Err()
	if err != nil {
		log.Error("Failed to add to redis ZSet, err:", err, " key :", key, ", Value:", val)
	}
	log.Debug("Updated TokenPriceHistory at Zset:", key, " with score:", toTime, ",val:", priceHistoryEntry)

	PrunePriceHistoryInRedis(key, fromTime)
}

func PrunePriceHistoryInRedis(key string, fromTime float64) {
	//Remove any entries older than 1 hour from fromTime.
	res := redisClient.ZRemRangeByScore(ctx, key, fmt.Sprintf("%f", 0.0),
		fmt.Sprintf("%f", fromTime-60*60))
	if res.Err() != nil {
		log.Error("Pruning entries at key:", key, "failed with error:", res.Err().Error())
	}
	log.Debug("Pruning: Removed ", res.Val(), " entries in redis Zset at key:", key)
}

func PruneTokenPriceZSet(tokenContractAddr string, blockHeight int64) {
	redisKey := fmt.Sprintf(redisutils.REDIS_KEY_TOKEN_BLOCK_HEIGHT_PRICE, settingsObj.PoolerNamespace, tokenContractAddr)
	res := redisClient.ZRemRangeByScore(ctx,
		redisKey,
		"-inf",
		fmt.Sprintf("%d", blockHeight))
	if res.Err() != nil {
		log.Error("Pruning entries at key:", redisKey, "failed with error:", res.Err().Error())
	}
	log.Debug("Pruning: Removed ", res.Val(), " entries in redis Zset at key:", redisKey)
}

func FetchTokenPriceAtBlockHeight(tokenContractAddr string, blockHeight int64) float64 {

	redisKey := fmt.Sprintf(redisutils.REDIS_KEY_TOKEN_BLOCK_HEIGHT_PRICE, settingsObj.PoolerNamespace, tokenContractAddr)
	type tokenPriceAtBlockHeight struct {
		BlockHeight int     `json:"blockHeight"`
		Price       float64 `json:"price"`
	}
	var tokenPriceAtHeight tokenPriceAtBlockHeight
	tokenPriceAtHeight.Price = 0
	for retryCount := 0; retryCount < 3; retryCount++ {
		zRangeByScore := redisClient.ZRangeByScore(ctx, redisKey, &redis.ZRangeBy{
			Min: fmt.Sprintf("%d", blockHeight),
			Max: fmt.Sprintf("%d", blockHeight),
		})
		if zRangeByScore.Err() != nil {
			log.Errorf("Failed to fetch tokenPrice for contract %s at blockHeight %d due to error %s, retrying %d",
				tokenContractAddr, zRangeByScore.Err().Error(), blockHeight, retryCount)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
		if len(zRangeByScore.Val()) == 0 {
			log.Error("Could not fetch tokenPrice for contract ", tokenContractAddr, " at BlockHeight:", blockHeight, " and hence will be set to 0")
			return tokenPriceAtHeight.Price
		}

		err := json.Unmarshal([]byte(zRangeByScore.Val()[0]), &tokenPriceAtHeight)
		if err != nil {
			log.Fatalf("Unable to parse tokenPrice retrieved from redis key %s error is %+v", redisKey, err)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
	}
	log.Debugf("Fetched tokenPrice %f for tokenContract %s at blockHeight %d", tokenPriceAtHeight.Price, tokenContractAddr, blockHeight)
	return tokenPriceAtHeight.Price
}

func StoreTokensSummaryPayload(blockHeight int64) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_TOKENS_SUMMARY_SNAPSHOT_AT_BLOCKHEIGHT, settingsObj.PoolerNamespace, blockHeight)
	payload := make([]*TokenData, len(tokenList))
	var i int
	for _, tokenData := range tokenList {
		payload[i] = tokenData
		i += 1
	}
	tokenSummaryJson, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Json marshal error %+v", err)
		return
	}
	for retryCount := 0; retryCount < 3; retryCount++ {
		res := redisClient.Set(ctx, key, string(tokenSummaryJson), 60*time.Minute) //TODO: Move to settings
		if res.Err() != nil {
			log.Errorf("Failed to add payload at blockHeight %d due to error %s, retrying %d", blockHeight, res.Err().Error(), retryCount)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
		log.Debugf("Added payload at key %s", key)
		break
	}
}

func StoreTokenSummaryCIDInSnapshotsZSet(blockHeight int64, tokenSummarySnapshotMeta *TokenSummarySnapshotMeta) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_TOKENS_SUMMARY_SNAPSHOTS_ZSET, settingsObj.PoolerNamespace)
	ZsetMemberJson, err := json.Marshal(tokenSummarySnapshotMeta)
	if err != nil {
		log.Fatalf("Json marshal error %+v", err)
		return
	}
	for retryCount := 0; retryCount < 3; retryCount++ {
		err := redisClient.ZAdd(ctx, key, &redis.Z{
			Score:  float64(blockHeight),
			Member: ZsetMemberJson,
		}).Err()
		if err != nil {
			log.Errorf("Failed to add payloadCID %s at blockHeight %d due to error %+v, retrying %d", tokenSummarySnapshotMeta.Cid, blockHeight, err, retryCount)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
		log.Debugf("Added payloadCID %s at blockHeight %d successfully at key %s", tokenSummarySnapshotMeta.Cid, blockHeight, key)
		break
	}
	PruneTokenSummarySnapshotsZSet()
}

func PruneTokenSummarySnapshotsZSet() {
	redisKey := fmt.Sprintf(redisutils.REDIS_KEY_TOKENS_SUMMARY_SNAPSHOTS_ZSET, settingsObj.PoolerNamespace)
	res := redisClient.ZCard(ctx, redisKey)
	zsetLen := res.Val()
	log.Debugf("ZSet length is %d", zsetLen)
	if zsetLen > 20 {
		for retryCount := 0; retryCount < 3; retryCount++ {
			endRank := -1*(zsetLen-20) + 1
			log.Debugf("Removing entries in ZSet from rank %d to rank %d", 0, endRank)
			res = redisClient.ZRemRangeByRank(ctx, redisKey, 0, endRank)
			if res.Err() != nil {
				log.Error("Pruning entries at key:", redisKey, "failed with error:", res.Err().Error(), " , retrying ", retryCount)
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue
			}
			log.Debug("Pruning: Removed ", res.Val(), " entries in redis Zset at key:", redisKey)
			break
		}
	}
}

func CommitTokenSummaryPayload() error {
	url := auditProtocolBaseURL + "/commit_payload"

	var apCommitResp AuditProtocolCommitPayloadResp
	var request AuditProtocolCommitPayloadReq
	request.ProjectId = fmt.Sprintf(TOKENSUMMARY_PROJECTID, settingsObj.PoolerNamespace)
	request.Payload.TokensData = make([]*TokenData, len(tokenList))
	request.Web3Storage = true //Always store TokenData snapshot in web3.storage.
	request.SkipAnchorProof = settingsObj.ContractCallBackend.SkipSummaryProjectProof
	var i int
	for _, tokenData := range tokenList {
		request.Payload.TokensData[i] = tokenData
		i += 1
	}
	body, err := json.Marshal(request)
	if err != nil {
		log.Fatalf("Failed to marshal request %+v towards Audit-Protocol with error %+v", request, err)
		return err
	}
	log.Debugf("URL %s. Committing Payload %s", url, string(body))
	retryCount := 0
	for ; retryCount < 3; retryCount++ {
		resp, err := apHttpClient.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Errorf("Error: Could not send commit-payload request to audit-protocol %+v due to error:", request, err)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error("Unable to read HTTP resp from Audit-protocol for commit-Payload.", err)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
		log.Trace("Rsp Body", string(body))
		if resp.StatusCode == http.StatusOK {
			if err = json.Unmarshal(body, &apCommitResp); err != nil { // Parse []byte to the go struct pointer
				log.Errorf("Can not unmarshal JSON response received from Audit-protocol due to error %s, retrying %d",
					err, retryCount)
				continue
			}
			log.Debugf("Sucessfully committed payload to Audit-protocol at tentativeHeight %d with commitId %s",
				apCommitResp.TentativeHeight, apCommitResp.CommitID)
		} else {
			var errorResp AuditProtocolErrorResp
			if err = json.Unmarshal(body, &errorResp); err != nil {
				log.Errorf("Can not unmarshal error JSON response received from Audit-protocol due to error %s, retrying %d",
					err, retryCount)
				continue
			}
			log.Errorf("Received %d error on commit-payload with error data %+v, retrying %d",
				resp.StatusCode, errorResp, retryCount)
			continue
		}
		break
	}
	if retryCount >= 3 {
		return errors.New("failed to commit payload after max retries")
	}
	return nil
}

func FetchPairSummarySnapshot(blockHeight int64) []TokenPairLiquidityProcessedData {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PAIRS_SUMMARY_SNAPSHOT_BLOCKHEIGHT, settingsObj.PoolerNamespace, blockHeight)
	log.Debugf("Fetching latest PairSummary snapshot from redis key %s", key)
	var pairsSummarySnapshot PairSummarySnapshot

	for retryCount := 0; retryCount < 3; retryCount++ {
		res := redisClient.Get(ctx, key)
		if res.Err() != nil {
			if res.Err() == redis.Nil {
				log.Errorf("Key %s not found in redis", key)
				return nil
			}
			log.Errorf("Error: Could not fetch latest PairSummary snapshot from redis. Error %+v. Retrying %d", res.Err(), retryCount)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
		res.Val()
		log.Tracef("Rsp Body %s", res.Val())
		if err := json.Unmarshal([]byte(res.Val()), &pairsSummarySnapshot); err != nil { // Parse []byte to the go struct pointer
			log.Errorf("Can not unmarshal JSON due to error %+v", err)
			continue
		}
		log.Debugf("Pairs Summary snapshot is : %+v", pairsSummarySnapshot)
		return pairsSummarySnapshot.Data
	}
	return nil
}

func WaitAndFetchBlockHeightStatus(projectID string, blockHeight int64, retries int) (*TokenSummarySnapshotMeta, error) {
	url := fmt.Sprintf("%s/%s/payload/%d/status", auditProtocolBaseURL, projectID, blockHeight)
	log.Debug("Fetching CID at Blockheight URL:", url)

	var apResp AuditProtocolBlockHeightStatusResp
	var retryCount int
	for retryCount = 0; retryCount <= retries; retryCount++ {
		resp, err := apHttpClient.Get(url)
		if err != nil {
			log.Error("Error: Could not fetch block height for pairContract:", projectID, " Error:", err)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Unable to read HTTP resp due to error %+v , retrying %d", err, retryCount)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
		log.Trace("Rsp Body", string(body))
		if resp.StatusCode == http.StatusBadRequest {
			log.Debugf("Snapshot for Block at height %d not yet ready, retrying %d", blockHeight, retryCount)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
		if err = json.Unmarshal(body, &apResp); err != nil { // Parse []byte to the go struct pointer
			log.Errorf("Can not unmarshal JSON due to error %+v, retrying %d", err, retryCount)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}

		log.Debugf("Successfully received response  %+v for CID fetch for URL %s is", apResp, url)
		if apResp.Status < TX_CONFIRMATION_PENDING {
			log.Debugf("BlockHeight %d status is still pending %d. Retrying", blockHeight, apResp.Status)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
		log.Debugf("Got CID %s, txHash %s at Block Height %d for projectID %s", apResp.PayloadCid, apResp.TxHash, blockHeight, projectID)
		tokenSummarySnapshotMeta := TokenSummarySnapshotMeta{apResp.PayloadCid, apResp.TxHash, apResp.Status, "", apResp.BlockHeight, 0, 0, 0, 0}
		return &tokenSummarySnapshotMeta, nil
	}

	log.Errorf("Max retries reached while trying to fetch payloadCID at height %d. Not retrying anymore.", blockHeight)
	return nil, fmt.Errorf("max retries reached to fetch payloadCID at height %d", blockHeight)

}

func FetchPairsSummaryLatestBlockHeight() int64 {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PAIRS_SUMMARY_SNAPSHOTS_ZSET, settingsObj.PoolerNamespace)
	log.Debugf("Fetching latest available PairSummarySnapshot Blockheight from %s", key)

	for retryCount := 0; retryCount < 3; retryCount++ {
		res := redisClient.ZRangeWithScores(ctx, key, -1, -1)
		if res.Err() != nil {
			log.Errorf("Error: Could not latest block height for PairSummarySnapshot. Error: %+v. Retrying %d", res.Err(), retryCount)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			continue
		}
		if len(res.Val()) == 0 {
			log.Debugf("No latest BlockHeight available for PairSummarySnapshot")
			return 0
		}
		blockHeight := int64(res.Val()[0].Score)
		log.Debugf("Latest available snapshot for PairSummarySnapshot is at height: %d", blockHeight)
		return blockHeight
	}
	log.Errorf("Could not retrieve latest available blockHeight for PairSummarySnapshot even after max retries.")
	return 0
}

func PopulatePairContractList(pairContractAddr string) {
	if pairContractAddr != "" {
		log.Info("Skipping reading contract addresses from json.Considering only passed pairContractaddress:", pairContractAddr)
		pairContracts = make([]string, 1)
		pairContracts[0] = pairContractAddr
		return
	}

	log.Info("Reading contracts:", pairContractListFile)
	data, err := os.ReadFile(pairContractListFile)
	if err != nil {
		log.Error("Cannot read the file:", err)
		panic(err)
	}

	log.Debug("Contracts json data is", string(data))
	err = json.Unmarshal(data, &pairContracts)
	if err != nil {
		log.Error("Cannot unmarshal the pair-contracts json ", err)
		panic(err)
	}
}

func ReadSettings() {

	log.Info("Reading Settings:", settingsFile)
	settingsObj = settings.ParseSettings(settingsFile)
	if settingsObj.TokenAggregatorSettings.RunIntervalSecs == 0 {
		settingsObj.TokenAggregatorSettings.RunIntervalSecs = 60
	}
	if settingsObj.TokenAggregatorSettings.Port == 0 {
		settingsObj.TokenAggregatorSettings.Port = 8000
	}
	retryInterval = settingsObj.RetryIntervalSecs
	log.Info("Settings for namespace", settingsObj.PoolerNamespace)
}

func SetupRedisClient() {
	redisURL := settingsObj.Redis.Host + ":" + strconv.Itoa(settingsObj.Redis.Port)

	log.Info("Connecting to redis at:", redisURL)
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       settingsObj.Redis.Db,
	})
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Error("Unable to connect to redis at:")
	}
	log.Info("Connected successfully to Redis and received ", pong, " back")
}

func InitAuditProtocolClient() {
	//TODO: Move these to settings
	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        2,
		MaxConnsPerHost:     2,
		MaxIdleConnsPerHost: 2,
		IdleConnTimeout:     0,
		DisableCompression:  true,
	}

	apHttpClient = http.Client{
		Timeout:   10 * time.Second,
		Transport: &t,
	}
}
