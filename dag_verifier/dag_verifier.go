package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

var ctx = context.Background()

type DagVerifier struct {
	redisClient                 *redis.Client
	projects                    []string
	periodicRetrievalInterval   time.Duration
	lastVerifiedDagBlockHeights map[string]string
}

//TODO: Migrate to env or settings.
const NAMESPACE string = "UNISWAPV2"

func (verifier *DagVerifier) Initialize(settings SettingsObj, pairContractAddresses *[]string) {
	verifier.InitIPFSClient(settings)
	verifier.InitRedisClient(settings)
	verifier.projects = make([]string, 0, 50)

	verifier.PopulateProjects(pairContractAddresses)
	verifier.periodicRetrievalInterval = 300 * time.Second
	//Fetch DagChain verification status from redis for all projects.
	verifier.FetchLastVerificationStatusFromRedis()
}

func (verifier *DagVerifier) PopulateProjects(pairContractAddresses *[]string) {
	pairAddresses := *pairContractAddresses
	//For now as we are aware there are 2 types of projects for uniswap, we can hardcode the same.
	for i := range *pairContractAddresses {
		pairTradeVolumeProjectId := "projectID:uniswap_pairContract_trade_volume_" + pairAddresses[i] + "_" + NAMESPACE
		pairTotalReserveProjectId := "projectID:uniswap_pairContract_pair_total_reserves_" + pairAddresses[i] + "_" + NAMESPACE
		verifier.projects = append(verifier.projects, pairTotalReserveProjectId)
		verifier.projects = append(verifier.projects, pairTradeVolumeProjectId)
	}
}

func (verifier *DagVerifier) FetchLastVerificationStatusFromRedis() {
	key := "projects:" + NAMESPACE + ":dagVerificationStatus"
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
	key := "projects:" + NAMESPACE + ":dagVerificationStatus"
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
	//For now only the PayloadChain that is stored in redis is used as a reference to verify.
	//TODO: Need to validate the original dag chain either from redis/from what is stored in IPFS.
	var chain []DagPayload
	//Get ZSet from redis for payloadCids and start verifying if there are any gaps.
	chain, err := verifier.GetPayloadCidsFromRedis(projectId, verifier.lastVerifiedDagBlockHeights[projectId])
	if err != nil {
		//Raise an alarm in future for this
		log.Error("Failed to fetch payload CIDS for projectID from redis:", projectId)
		return err
	}
	if len(chain) == 0 {
		log.Info("No new blocks to verify in the chain from previous height for projectId", projectId)
		return nil
	}
	for i := range chain {
		payload, err := ipfsClient.GetPayload(chain[i].PayloadCid, 3)
		if err != nil {
			//If we are unable to fetch a CID from IPFS, retry
			log.Error("Failed to get PayloadCID from IPFS. Either cache is corrupt or there is an actual issue.CID:", chain[i].PayloadCid)
			//TODO: Either cache is corrupt or there is an actual issue.
			//Check and fix cache corruption by getting Dagchain from IPFS.
			return err
		}
		chain[i].Data = payload
		//Fetch payload from IPFS and check gaps in chainHeight.\
		log.Debugf("Index: %d ,payload: %+v", i, payload)
	}
	log.Info("Verifying Dagchain for ProjectId %s , from block %d to %d", projectId, chain[0].DagChainHeight, chain[(len(chain)-1)].DagChainHeight)
	gapsPresent, chainGaps := verifier.verifyDagForGaps(&chain)
	if !gapsPresent {
		log.Info("Dag chain has gaps for projectID:", projectId)
		verifier.updateDagGapsinRedis(projectId, chainGaps)
	}
	//Store last verified blockHeight so that in next run, we just need to verify from the same.
	//Use single hash key in redis to store the same against contractAddress.
	verifier.lastVerifiedDagBlockHeights[projectId] = strconv.FormatInt(chain[len(chain)-1].DagChainHeight, 10)
	return nil
}

func (verifier *DagVerifier) updateDagGapsinRedis(projectId string, chainGaps []DagChainGap) {
	key := projectId + ":dagChainGaps"
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

func (verifier *DagVerifier) GetPayloadCidsFromRedis(projectId string, startScore string) ([]DagPayload, error) {
	var dagPayloadsInfo []DagPayload

	key := projectId + ":payloadCids"
	log.Debug("Fetching PayloadCids from redis at key:", key, ",with startScore: ", startScore)
	zRangeByScore := verifier.redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: startScore,
		Max: "+inf",
	})
	/* 	zRangeByScore := verifier.redisClient.ZRangeByScoreWithScores("projectID:uniswap_pairContract_pair_total_reserves_0x61b62c5d56ccd158a38367ef2f539668a06356ab_UNISWAPV2:payloadCids", redis.ZRangeBy{
		Min: startScore,
		Max: "3",
	}) */
	//verifier.redisClient.Z
	err := zRangeByScore.Err()
	log.Debug("Result for ZRangeByScoreWithScores : ", zRangeByScore)
	if err != nil {
		log.Error("Could not fetch entries error: ", err, "Query:", zRangeByScore)
		return nil, err
	}
	res := zRangeByScore.Val()
	dagPayloadsInfo = make([]DagPayload, len(res))
	log.Debugf("Fetched %d Payload CIDs for key %s", len(res), key)
	for i := range res {
		//Safe to convert as we know height will always be int.
		dagPayloadsInfo[i].DagChainHeight = int64(res[i].Score)
		dagPayloadsInfo[i].PayloadCid = fmt.Sprintf("%v", res[i].Member)
	}
	return dagPayloadsInfo, nil
}

func (verifier *DagVerifier) verifyDagForGaps(chain *[]DagPayload) (bool, []DagChainGap) {
	dagPayloads := *chain
	log.Info("Verifying DAG for gaps. DAG chain length is:", len(dagPayloads))
	//fmt.Printf("%+v\n", dagChain)
	var prevDagBlockEnd, lastBlock, firstBlock, numGaps int64
	firstBlock = dagPayloads[0].Data.ChainHeightRange.End
	lastBlock = dagPayloads[len(dagPayloads)-1].Data.ChainHeightRange.Begin
	var dagGaps []DagChainGap
	for i := range dagPayloads {
		//log.Debug("Processing dag block :", i, "nextDagBlockStart:", nextDagBlockStart)
		if prevDagBlockEnd != 0 {
			if dagPayloads[i].DagChainHeight == dagPayloads[i-1].DagChainHeight {
				//TODO:If there are multiple snapshots observed at same blockHeight, need to take action to cleanup snapshots
				//		which are not required from IPFS based on previous and next blocks.
				log.Errorf("Found Same DagchainHeight %d at 2 levels. DagChain needs to be fixed.", dagPayloads[i].DagChainHeight)
			}
			curBlockStart := dagPayloads[i].Data.ChainHeightRange.End
			//log.Debug("curBlockEnd", curBlockEnd, " nextDagBlockStart", nextDagBlockStart)
			if curBlockStart != prevDagBlockEnd+1 {
				log.Debug("Gap identified at ChainHeight:", dagPayloads[i].DagChainHeight, ",PayloadCID:", dagPayloads[i].PayloadCid, ", between height:", dagPayloads[i-1].DagChainHeight, " and ", dagPayloads[i].DagChainHeight)
				log.Debug("Missing blocks from(not including): ", prevDagBlockEnd, " to(not including): ", curBlockStart)
				dagGaps = append(dagGaps, DagChainGap{MissingBlockHeightStart: prevDagBlockEnd + 1,
					MissingBlockHeightEnd: curBlockStart - 1,
					TimestampIdentified:   time.Now().Unix(),
					DAGBlockHeight:        dagPayloads[i].DagChainHeight})
				numGaps++
			}
		}
		prevDagBlockEnd = dagPayloads[i].Data.ChainHeightRange.End
	}
	log.Info("Block Range is from:", firstBlock, ", to:", lastBlock)
	log.Info("Number of gaps found:", numGaps)
	if numGaps == 0 {
		return true, nil
	} else {
		return false, dagGaps
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
