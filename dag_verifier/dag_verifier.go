package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

var ctx = context.Background()

type DagVerifier struct {
	redisClient                      *redis.Client
	projects                         []string
	periodicRetrievalInterval        time.Duration
	lastVerifiedDagBlockHeights      map[string]string
	lastVerifiedDagBlockHeightsMutex *sync.RWMutex
}

//TODO: Migrate to env or settings.
const NAMESPACE string = "UNISWAPV2"
const PAIR_TRADEVOLUME_PROJECTID string = "projectID:uniswap_pairContract_trade_volume_%s_%s"
const PAIR_TOTALRESERVE_PROJECTID string = "projectID:uniswap_pairContract_pair_total_reserves_%s_%s"

const DAG_CHAIN_ISSUE_DUPLICATE_HEIGHT string = "DUPLICATE_HEIGHT_IN_CHAIN"
const DAG_CHAIN_ISSUE_GAP_IN_CHAIN string = "GAP_IN_CHAIN"

func (verifier *DagVerifier) Initialize(settings SettingsObj, pairContractAddresses *[]string) {
	verifier.InitIPFSClient(settings)
	verifier.InitRedisClient(settings)
	verifier.projects = make([]string, 0, 50)

	verifier.PopulateProjects(pairContractAddresses)
	verifier.periodicRetrievalInterval = 300 * time.Second
	//Fetch DagChain verification status from redis for all projects.
	verifier.FetchLastVerificationStatusFromRedis()
	verifier.lastVerifiedDagBlockHeightsMutex = &sync.RWMutex{}
}

func (verifier *DagVerifier) PopulateProjects(pairContractAddresses *[]string) {
	pairAddresses := *pairContractAddresses
	//For now as we are aware there are 2 types of projects for uniswap, we can hardcode the same.
	for i := range *pairContractAddresses {
		pairTradeVolumeProjectId := fmt.Sprintf(PAIR_TRADEVOLUME_PROJECTID, pairAddresses[i], NAMESPACE)
		pairTotalReserveProjectId := fmt.Sprintf(PAIR_TOTALRESERVE_PROJECTID, pairAddresses[i], NAMESPACE)
		verifier.projects = append(verifier.projects, pairTotalReserveProjectId)
		verifier.projects = append(verifier.projects, pairTradeVolumeProjectId)
	}
}

func (verifier *DagVerifier) FetchLastVerificationStatusFromRedis() {
	key := fmt.Sprintf(REDIS_KEY_DAG_VERIFICATION_STATUS, NAMESPACE)
	log.Debug("Fetching LastVerificationStatusFromRedis at key:", key)

	res := verifier.redisClient.HGetAll(ctx, key)
	//res := verifier.redisClient.HGetAll(key)
	//log.Debug("Res:", res, "\n\n")
	if len(res.Val()) == 0 {
		log.Info("Failed to fetch LastverificationStatus from redis for the projects.")
		//Key doesn't exist.
		log.Info("Key doesn't exist..hence proceed from start of the block.")
		verifier.lastVerifiedDagBlockHeights = make(map[string]string)
		for i := range verifier.projects {
			//TODO: do we need to change this to be based on some timeDuration like last 24h etc instead of starting from 1??
			verifier.lastVerifiedDagBlockHeights[verifier.projects[i]] = "1"
		}
		return
	}
	if res.Err() != nil {
		log.Error("Ideally should not come here, which means there is some other redis error. To debug:", res.Err())
	}
	//TODO: Need to handle dynamic addition of projects.
	verifier.lastVerifiedDagBlockHeights = res.Val()
	log.Debugf("Fetched LastVerificationStatus from redis %+v", verifier.lastVerifiedDagBlockHeights)
}

func (verifier *DagVerifier) UpdateLastStatusToRedis() {
	key := fmt.Sprintf(REDIS_KEY_DAG_VERIFICATION_STATUS, NAMESPACE)
	log.Info("Updating LastVerificationStatusFromRedis at key:", key)
	res := verifier.redisClient.HMSet(ctx, key, verifier.lastVerifiedDagBlockHeights)
	if res.Err() != nil {
		log.Error("Failed to update lastVerifiedDagBlockHeights in redis..Retry in next run.")
	}
}

func (*DagVerifier) InitIPFSClient(settingsObj SettingsObj) {

	//Initialize and do a basic test to see if IPFS client is connected to IPFS server and is able to fetch.
	ipfsClient.Init(settingsObj)
	//TODO: Add  a way to verify IPFS client initialization is sucess and connection to IPFS node?
	/*dagCid := "bafyreidiweqijqgiaaitzyktovv3zpiuqh7sbk5rmbrjxupgg7dhfcehvu"

	dagBlock, err := ipfsClient.DagGet(dagCid)
	if err != nil {
		return
	}
	//log.Debug("Got dag Block", dagBlock, ", for CID:", dagCid)
	log.Debugf("Got dag Block %+v for CID:%s", dagBlock, dagCid)
	dagPayload, err := ipfsClient.GetPayload(dagBlock.Data.Cid)
	if err != nil {
		return
	}
	log.Debugf("Read Data CId from IPFS: %+v", dagPayload)*/
}

func (verifier *DagVerifier) Run() {

	for {
		//TODO: Invoke PopulatePairContractList(pairContractAddress) to fetch updated file.

		verifier.VerifyAllProjects() //Projects are pairContracts
		verifier.UpdateLastStatusToRedis()
		log.Info("Sleeping for " + verifier.periodicRetrievalInterval.String() + " secs")
		time.Sleep(verifier.periodicRetrievalInterval)
	}

}

func (verifier *DagVerifier) VerifyAllProjects() {
	var wg sync.WaitGroup

	for i := range verifier.projects {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()
			verifier.VerifyDagChain(verifier.projects[index])
		}(i)

	}
	wg.Wait()
}

func (verifier *DagVerifier) VerifyDagChain(projectId string) error {
	//Fetch the DAGChain cached in redis and then corresponding payloads at chainHeight.
	// For now only the Chain that is stored in redis is used as a reference to verify.
	//TODO: Need to validate the original dag chain from what is stored in IPFS. Is this required??
	var dagChain []DagChainBlock
	verifier.lastVerifiedDagBlockHeightsMutex.RLock()
	startScore := verifier.lastVerifiedDagBlockHeights[projectId]
	verifier.lastVerifiedDagBlockHeightsMutex.RUnlock()
	dagChain, err := verifier.GetDagChainCidsFromRedis(projectId, startScore)
	if err != nil {
		//Raise an alarm in future for this
		log.Error("Failed to fetch DAG Chain CIDS for projectID from redis:", projectId)
		return err
	}
	if len(dagChain) == 0 {
		log.Info("No new blocks to verify in the chain from previous height for projectId", projectId)
		return nil
	}
	//Get ZSet from redis for payloadCids and start verifying if there are any gaps.
	dagChain, err = verifier.GetPayloadCidsFromRedis(projectId, startScore, dagChain)
	if err != nil {
		//Raise an alarm in future for this
		log.Errorf("Failed to fetch payload CIDS for projectID %s from redis with error %s.", projectId, err)
		return err
	}

	for i := range dagChain {
		payload, err := ipfsClient.GetPayload(dagChain[i].Payload.PayloadCid, 3)
		if err != nil {
			//If we are unable to fetch a CID from IPFS, retry
			log.Error("Failed to get PayloadCID from IPFS. Either cache is corrupt or there is an actual issue.CID:", dagChain[i].Payload.PayloadCid)
			//TODO: Either cache is corrupt or there is an actual issue.
			//Check and fix cache corruption by getting Dagchain from IPFS.
			return err
		}
		dagChain[i].Payload.Data = payload
		//Fetch payload from IPFS and check gaps in chainHeight.\
		log.Debugf("Index: %d ,payload: %+v", i, payload)
	}
	log.Infof("Verifying Dagchain for ProjectId %s , from block %d to %d", projectId, dagChain[0].Height, dagChain[(len(dagChain)-1)].Height)
	issuesPresent, chainIssues := verifier.verifyDagForIssues(&dagChain)
	if !issuesPresent {
		log.Info("Dag chain has issues for projectID:", projectId)
		verifier.updateDagGapsinRedis(projectId, chainIssues)
	}
	//Store last verified blockHeight so that in next run, we just need to verify from the same.
	//Use single hash key in redis to store the same against contractAddress.
	verifier.lastVerifiedDagBlockHeightsMutex.Lock()
	verifier.lastVerifiedDagBlockHeights[projectId] = strconv.FormatInt(dagChain[len(dagChain)-1].Height, 10)
	verifier.lastVerifiedDagBlockHeightsMutex.Unlock()
	return nil
}

func (verifier *DagVerifier) updateDagGapsinRedis(projectId string, chainGaps []DagChainIssue) {
	key := fmt.Sprintf(REDIS_KEY_PROJECT_DAG_CHAIN_GAPS, projectId)
	var gaps []*redis.Z
	for i := range chainGaps {
		gapStr, err := json.Marshal(chainGaps[i])
		if err != nil {
			log.Error("Serious issue if json marshal fails..Can't do anything else than continue and log error:", err)
			continue
		}
		gaps = append(gaps, &redis.Z{Score: float64(chainGaps[i].DAGBlockHeight),
			Member: gapStr,
		})
	}

	res := verifier.redisClient.ZAdd(ctx, key, gaps...)
	if res.Err() != nil {
		//TODO:Add retry logic later.
		log.Error("Failed to update dagChainGaps into redis for projectID:", projectId, ", GapData:", chainGaps)
	}
	log.Infof("Added %d DagGaps data successfully in redis for project: %s", len(chainGaps), projectId)
	//TODO: Need to prune older gaps.
}

//TODO: Need to handle Dagchain reorg event and reset the lastVerifiedBlockHeight to the same.
/*func (verifier *DagVerifier) HandleChainReOrg(){

}*/

func (verifier *DagVerifier) GetDagChainCidsFromRedis(projectId string, startScore string) ([]DagChainBlock, error) {
	var dagChainCids []DagChainBlock

	//key := projectId + ":payloadCids"
	key := fmt.Sprintf(REDIS_KEY_PROJECT_CIDS, projectId)

	log.Debug("Fetching DAG Chain Cids from redis at key:", key, ",with startScore: ", startScore)
	zRangeByScore := verifier.redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: startScore,
		Max: "+inf",
	})

	err := zRangeByScore.Err()
	log.Debug("Result for ZRangeByScoreWithScores : ", zRangeByScore)
	if err != nil {
		log.Error("Could not fetch entries error: ", err, "Query:", zRangeByScore)
		return nil, err
	}
	res := zRangeByScore.Val()
	dagChainCids = make([]DagChainBlock, len(res))
	log.Debugf("Fetched %d DAG Chain CIDs for key %s", len(res), key)
	for i := range res {
		//Safe to convert as we know height will always be int.
		dagChainCids[i].Height = int64(res[i].Score)
		dagChainCids[i].Data.Cid = fmt.Sprintf("%v", res[i].Member)
	}
	lastVerifiedStatus, err := strconv.ParseInt(startScore, 10, 64)
	if err != nil && len(dagChainCids) == 1 && dagChainCids[0].Height == lastVerifiedStatus {
		//This means no new dag blocks in the chain from last verified height.
		dagChainCids = nil
	}
	return dagChainCids, nil
}

func (verifier *DagVerifier) GetPayloadCidsFromRedis(projectId string, startScore string, dagChain []DagChainBlock) ([]DagChainBlock, error) {
	//var dagPayloadsInfo []DagPayload

	//key := projectId + ":payloadCids"
	key := fmt.Sprintf(REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectId)

	log.Debug("Fetching PayloadCids from redis at key:", key, ",with startScore: ", startScore)
	zRangeByScore := verifier.redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: startScore,
		//Max: "+inf",
		Max: fmt.Sprintf("%d", dagChain[len(dagChain)-1].Height),
	})

	err := zRangeByScore.Err()
	log.Debug("Result for ZRangeByScoreWithScores : ", zRangeByScore)
	if err != nil {
		log.Error("Could not fetch entries error: ", err, "Query:", zRangeByScore)
		return dagChain, err
	}
	res := zRangeByScore.Val()
	//dagPayloadsInfo = make([]DagPayload, len(res))
	log.Debugf("Fetched %d Payload CIDs for key %s", len(res), key)
	for i := range res {
		if len(dagChain) > i {
			if dagChain[i].Height != int64(res[i].Score) {
				return dagChain, fmt.Errorf("CRITICAL:Inconsistency between DAG Chain and Payloads stored in redis for Project:%s", projectId)
			}
			dagChain[i].Payload.PayloadCid = fmt.Sprintf("%v", res[i].Member)
		} else {
			log.Debugf("DAGChain for project %s is little behind payloadCids chainHeight.", projectId, len(dagChain))
			return nil, errors.New("chain construction seems to be in progress. Have to retry next time")
		}
	}
	return dagChain, nil
}

func (verifier *DagVerifier) verifyDagForIssues(chain *[]DagChainBlock) (bool, []DagChainIssue) {
	dagChain := *chain
	log.Info("Verifying DAG for Issues. DAG chain length is:", len(dagChain))
	//fmt.Printf("%+v\n", dagChain)
	var prevDagBlockEnd, lastBlock, firstBlock, numGaps, numDuplicates int64
	firstBlock = dagChain[0].Payload.Data.ChainHeightRange.End
	lastBlock = dagChain[len(dagChain)-1].Payload.Data.ChainHeightRange.Begin
	var dagIssues []DagChainIssue
	for i := range dagChain {
		//log.Debug("Processing dag block :", i, "nextDagBlockStart:", nextDagBlockStart)
		if prevDagBlockEnd != 0 {
			if dagChain[i].Height == dagChain[i-1].Height {
				dagIssues = append(dagIssues, DagChainIssue{IssueType: DAG_CHAIN_ISSUE_DUPLICATE_HEIGHT,
					TimestampIdentified: time.Now().Unix(),
					DAGBlockHeight:      dagChain[i].Height})
				//TODO:If there are multiple snapshots observed at same blockHeight, need to take action to cleanup snapshots
				//		which are not required from IPFS based on previous and next blocks.
				log.Errorf("Found Same DagchainHeight %d at 2 levels. DagChain needs to be fixed.", dagChain[i].Height)
				numDuplicates++
			}
			curBlockStart := dagChain[i].Payload.Data.ChainHeightRange.Begin
			//log.Debug("curBlockEnd", curBlockEnd, " nextDagBlockStart", nextDagBlockStart)
			if curBlockStart != prevDagBlockEnd+1 {
				log.Debug("Gap identified at ChainHeight:", dagChain[i].Height, ",PayloadCID:", dagChain[i].Payload.PayloadCid, ", between height:", dagChain[i-1].Height, " and ", dagChain[i].Height)
				log.Debug("Missing blocks from(not including): ", prevDagBlockEnd, " to(not including): ", curBlockStart)
				dagIssues = append(dagIssues, DagChainIssue{IssueType: DAG_CHAIN_ISSUE_GAP_IN_CHAIN, MissingBlockHeightStart: prevDagBlockEnd + 1,
					MissingBlockHeightEnd: curBlockStart - 1,
					TimestampIdentified:   time.Now().Unix(),
					DAGBlockHeight:        dagChain[i].Height})
				numGaps++
			}
		}
		prevDagBlockEnd = dagChain[i].Payload.Data.ChainHeightRange.End
	}
	log.Info("Block Range is from:", firstBlock, ", to:", lastBlock)
	log.Infof("Number of gaps found:%d. Number of Duplicates found:%d", numGaps, numDuplicates)
	if numGaps == 0 && numDuplicates == 0 {
		return true, nil
	} else {
		return false, dagIssues
	}
}

func (verifier *DagVerifier) InitRedisClient(settingsObj SettingsObj) {
	redisURL := settingsObj.Redis.Host + ":" + strconv.Itoa(settingsObj.Redis.Port)
	redisDb := settingsObj.Redis.Db
	//TODO: Change post testing to fetch from settings.
	//redisURL = "localhost:6379"
	//redisDb = 0
	log.Info("Connecting to redis at:", redisURL)
	verifier.redisClient = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       redisDb,
	})
	pong, err := verifier.redisClient.Ping(ctx).Result()
	//pong, err := verifier.redisClient.Ping().Result()
	if err != nil {
		log.Error("Unable to connect to redis at:")
	}
	log.Info("Connected successfully to Redis and received ", pong, " back")
}
