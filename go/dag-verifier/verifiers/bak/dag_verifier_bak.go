package verifiers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/swagftw/gi"

	"audit-protocol/caching"
	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/slackutils"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type DagVerifier struct {
	redisCache                       *caching.RedisCache
	diskCache                        *caching.LocalDiskCache
	ipfsClient                       *ipfsutils.IpfsClient
	projects                         []string
	SummaryProjects                  []string
	settings                         *settings.SettingsObj
	lastVerifiedDagBlockHeights      map[string]string
	ProjectsIndexedState             map[string]*datamodel.ProjectIndexedState
	lastVerifiedDagBlockHeightsMutex *sync.RWMutex
	dagChainHasIssues                bool
	dagCacheIssues                   int
	dagChainIssues                   map[string][]*datamodel.DagChainIssue
	previousCycleDagChainHeight      map[string]int64
	noOfCyclesSinceChainStuck        map[string]int
	lastNotifyTime                   int64
}

const (
	DAG_CHAIN_ISSUE_DUPLICATE_HEIGHT string = "DUPLICATE_HEIGHT_IN_CHAIN"
	DAG_CHAIN_ISSUE_GAP_IN_CHAIN     string = "GAP_IN_CHAIN"
	DAG_CHAIN_REPORT_SEVERITY_HIGH   string = "High"
	DAG_CHAIN_REPORT_SEVERITY_MEDIUM string = "Medium"
	DAG_CHAIN_REPORT_SEVERITY_LOW    string = "Low"
	DAG_CHAIN_REPORT_SEVERITY_CLEAR  string = "Cleared"
)

var (
	ErrInconsistentHeightInCache = errors.New("inconsistent height in cache chain")
	ErrInconsistentCacheChain    = errors.New("inconsistent cache chain")
)

func InitializeDagVerifier() (*DagVerifier, error) {
	verifier := new(DagVerifier)

	var err error

	verifier.ipfsClient, err = gi.Invoke[*ipfsutils.IpfsClient]()
	if err != nil {
		return nil, err
	}

	verifier.redisCache, err = gi.Invoke[*caching.RedisCache]()
	if err != nil {
		return nil, err
	}

	verifier.dagChainIssues = make(map[string][]*datamodel.DagChainIssue)

	err = verifier.PopulateProjects()
	if err != nil {
		return nil, err
	}

	verifier.lastVerifiedDagBlockHeightsMutex = &sync.RWMutex{}

	err = gi.Inject(verifier)
	if err != nil {
		log.WithError(err).Error("failed to inject dag verifier")
		return nil, err
	}

	return verifier, nil
}

// Run starts the dag verifier
// this is called on event when new blocks are inserted in chain
func (verifier *DagVerifier) Run() {
	periodicRetrievalInterval := time.Duration(verifier.settings.DagVerifierSettings.RunIntervalSecs) * time.Second

	for {
		if len(verifier.projects) == 0 {
			log.Info("No projects to be verified. Have to check in next run.")
			return nil
		}

		// get last verified dag block height for all projects
		err := verifier.GetLastVerificationStatus()
		if err != nil {
			return err
		}

		// get index status for all the projects from cache
		// index status is map of startSourceChainHeight and currentSourceChainHeight
		err = verifier.GetLastProjectIndexedStatus()
		if err != nil {
			return err
		}

		verifier.VerifyAllProjects() //Projects are pairContracts
		verifier.SummarizeDAGIssuesAndNotify()
		verifier.UpdateLastStatusToRedis()

		log.Info("Sleeping for " + periodicRetrievalInterval.String())
		time.Sleep(periodicRetrievalInterval)
	}
}

// GetLastProjectIndexedStatus gets index status for all the projects from cache
func (verifier *DagVerifier) GetLastProjectIndexedStatus() error {
	indexedStateMap, err := verifier.redisCache.GetLastProjectIndexedState(context.Background())
	if err != nil && err != caching.ErrNotFound {
		return err
	}

	if len(indexedStateMap) == 0 {
		// if last project indexed state is not found, fetch starting index for all projects
		var wg sync.WaitGroup

		for _, projectID := range verifier.projects {
			wg.Add(1)

			go func(projectID string) {
				defer wg.Done()

				startIndex, err := verifier.FetchStartIndex(projectID)
				if err != nil {
					log.Errorf("Could not fetch Start index for project %s due to error %+v", projectID, err)
					return
				}

				log.Infof("Fetched startIndex for project %s as %d", projectID, startIndex)

				indexedStateMap[projectID] = &datamodel.ProjectIndexedState{
					StartSourceChainHeight:   startIndex,
					CurrentSourceChainHeight: startIndex,
				}
			}(projectID)
		}

		wg.Wait()
	}

	verifier.ProjectsIndexedState = indexedStateMap

	return nil
}

// FetchStartIndex fetches the start index or chain range begin height from the payload
func (verifier *DagVerifier) FetchStartIndex(projectId string) (int64, error) {
	// get payload cid for the first dag block
	payloadCid, err := verifier.redisCache.GetPayloadCidAtDAGHeight(context.Background(), projectId, 1)
	if err != nil {
		// Raise an alarm in future for this.
		log.Error("Failed to fetch DAG Chain CIDS for projectID from redis:", projectId)

		return 0, err
	}

	if payloadCid == "" {
		log.Info("Indexing has not started for projectId", projectId)

		return 0, nil
	}

	payload, err := verifier.GetPayloadFromDiskCache(projectId, payloadCid)
	if err != nil {
		// If not found in disk cache, fetch from IPFS.
		// Fetch CID from IPFS and store the startHeight.
		payload, err = verifier.ipfsClient.GetPayloadChainHeightRang(payloadCid)
		if err != nil {
			log.Errorf("Failed to fetch payloadCID %s from IPFS for project %s", payloadCid, projectId)

			return 0, nil
		}
	}

	return payload.ChainHeightRange.Begin, nil
}

// GetLastVerificationStatus gets last verified dag block height for every project from cache
func (verifier *DagVerifier) GetLastVerificationStatus() error {
	log.Debug("Fetching last verification status from cache")

	lastVerStatusMap, err := verifier.redisCache.GetLastVerificationStatus(context.Background())
	if err != nil {
		return err
	}

	if lastVerStatusMap == nil {
		for i := range verifier.projects {
			// TODO: do we need to change this to be based on some timeDuration like last 24h etc instead of starting from 1??
			verifier.lastVerifiedDagBlockHeights[verifier.projects[i]] = "0"
		}
		for i := range verifier.SummaryProjects {
			verifier.lastVerifiedDagBlockHeights[verifier.SummaryProjects[i]] = "0"
		}
	}

	verifier.lastVerifiedDagBlockHeights = lastVerStatusMap

	return nil
}

func (verifier *DagVerifier) UpdateLastStatusToRedis() {
	//No retry has been added, because in case of a failure, status will get updated in next run.
	key := fmt.Sprintf(redisutils.REDIS_KEY_DAG_VERIFICATION_STATUS)
	log.Info("Updating LastVerificationStatus at key:", key)
	res := verifier.redisClient.HMSet(context.Background(), key, verifier.lastVerifiedDagBlockHeights)
	if res.Err() != nil {
		log.Error("Failed to update lastVerifiedDagBlockHeights in redis..Retry in next run.")
	}
	//Update indexed status to redis
	key = fmt.Sprintf(redisutils.REDIS_KEY_PROJECTS_INDEX_STATUS)
	log.Info("Updating LastIndexedStatus at key:", key)
	projectsIndexedState := make(map[string]string)
	for i := range verifier.projects {
		if val, ok := verifier.ProjectsIndexedState[verifier.projects[i]]; ok {
			marshalledState, err := json.Marshal(val)
			if err != nil {
				log.Fatalf("Failed to marshal json %+v", err)
			}
			projectsIndexedState[verifier.projects[i]] = string(marshalledState)
		}
	}
	res1 := verifier.redisClient.HMSet(context.Background(), key, projectsIndexedState)
	if res1.Err() != nil {
		log.Error("Failed to update LastIndexedStatus in redis..Retry in next run.")
	}
}

// VerifyAllProjects verifies all the projects in parallel
func (verifier *DagVerifier) VerifyAllProjects() {
	var wg sync.WaitGroup
	//TODO: change to batch logic as done in pruning service.
	for _, projectID := range verifier.projects {
		wg.Add(1)

		go func(projectID string) {
			defer wg.Done()
			err := verifier.VerifyDagChain(projectID)
			if err != nil {
				log.
					WithField("projectID", projectID).
					WithError(err).Errorf("failed to verify project")
			}
		}(projectID)
	}

	wg.Wait()
}

func (verifier *DagVerifier) GetProjectDAGBlockHeightFromRedis(projectId string) string {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_BLOCK_HEIGHT, projectId)
	for i := 0; i < 3; i++ {
		res := verifier.redisClient.Get(context.Background(), key)
		if res.Err() == redis.Nil {
			log.Errorf("No blockHeight key for the project %s is present in redis", projectId)
			return "0"
		}
		if res.Err() != nil {
			log.Errorf("Failed to fetch blockHeight for project %s from redis due to error %+v", projectId, res.Err())
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("Retrieved BlockHeight for project %s from redis is %s", projectId, res.Val())
		return res.Val()
	}
	return "0"
}

func (verifier *DagVerifier) GetPayloadFromDiskCache(projectId string, payloadCid string) (*datamodel.DagPayloadChainHeightRange, error) {
	payload := new(datamodel.DagPayloadChainHeightRange)
	fileName := fmt.Sprintf("%s/%s/%s.json", verifier.settings.PayloadCachePath, projectId, payloadCid)

	bytes, err := verifier.diskCache.Read(fileName)
	if err != nil {
		log.Errorf("Failed to fetch payload with cid %s for project %s from cache due to error %+v", payloadCid, projectId, err)

		return payload, err
	}

	err = json.Unmarshal(bytes, payload)
	if err != nil {
		log.Error("Failed to Unmarshal Json Payload from IPFS, CID:", payloadCid, ", bytes:", string(bytes), ", error:", err)

		return payload, err
	}

	return payload, nil
}

// VerifyDagChain verifies the dag chain for a given project
func (verifier *DagVerifier) VerifyDagChain(projectID string) error {
	// Fetch the DAGChain cached in redis and then corresponding payloads at chainHeight.
	// For now only the Chain that is stored in redis is used as a reference to verify.
	// TODO: Need to validate the original dag chain from what is stored in IPFS. Is this required??
	verifier.lastVerifiedDagBlockHeightsMutex.RLock()
	startScore := verifier.lastVerifiedDagBlockHeights[projectID] // last verified block height
	verifier.lastVerifiedDagBlockHeightsMutex.RUnlock()

	l := log.WithField("projectID", projectID)

	startHeight, err := strconv.Atoi(startScore)
	if err != nil {
		l.WithError(err).Error("Failed to parse startScore to int")

		return err
	}

	// return dag blocks with cid and respective height
	dagChainBlocks, err := verifier.redisCache.GetDagChainCIDs(context.Background(), projectID, startHeight, -1)
	if err != nil {
		return err
	}

	if len(dagChainBlocks) == 1 && dagChainBlocks[0].Height == int64(startHeight) {
		// This means no new dag blocks in the chain from last verified height.
		l.Info("No new blocks to verify in the chain from previous height")

		return nil
	}

	// Get ZSet from cache for payloadCids and start verifying if there are any gaps.
	dagChainBlocksWithPayloadData, err := verifier.redisCache.GetPayloadCIDs(context.Background(), projectID, startHeight, int(dagChainBlocks[len(dagChainBlocks)-1].Height))
	if err != nil {
		// Raise an alarm in future for this.
		l.WithError(err).Errorf("Failed to fetch payload CIDS")

		return err
	}

	if len(dagChainBlocks) != len(dagChainBlocksWithPayloadData) {
		l.WithField("dagBlockCount", len(dagChainBlocks)).
			WithField("payloadCount", len(dagChainBlocksWithPayloadData)).
			Error("dag block count and payload count are not matching in cache")

		return ErrInconsistentHeightInCache
	}

	for index, block := range dagChainBlocks {
		if block.Height != dagChainBlocksWithPayloadData[index].Height {
			log.WithField("projectID", projectID).
				WithField("dagBlockHeight", block.Height).
				WithField("payloadHeight", dagChainBlocksWithPayloadData[index].Height).
				Error("Dag block height and payload height are not matching")

			return ErrInconsistentCacheChain
		}

		block.Payload = datamodel.DagPayload{
			PayloadCid: dagChainBlocksWithPayloadData[index].Payload.PayloadCid,
		}
	}

	// check of duplicate block heights
	// invalid entries in cache are entries with same height but different cid and null payload cid
	err = verifier.checkInvalidEntriesInCache(dagChainBlocks, projectID)
	if err != nil {
		return err
	}

	// get payload data from cache or ipfs
	for _, block := range dagChainBlocks {
		payload, err := verifier.GetPayloadFromDiskCache(projectID, block.Payload.PayloadCid)
		if err != nil {
			payload, err = verifier.ipfsClient.GetPayloadChainHeightRang(block.Payload.PayloadCid)
			if err != nil {
				log.Error("Failed to get PayloadLink from IPFS. Either cache is corrupt or there is an actual issue.CID:", block.Payload.PayloadCid)
				// TODO: Either cache is corrupt or there is an actual issue.
				// Check and fix cache corruption by getting Dagchain from IPFS.
				return err
			}
		}
		block.Payload.Data = payload
	}

	startHeight = int(dagChainBlocks[0].Height)
	endHeight := int(dagChainBlocks[len(dagChainBlocks)-1].Height)

	log.Infof("Verifying Dagchain for ProjectId %s , from block %d to %d", projectID, startHeight, endHeight)

	issuesPresent, chainIssues := verifier.verifyDagChainForIssues(dagChainBlocks)
	if issuesPresent {
		log.Infof("Dag chain has issues for projectID %s. issues are: %+v", projectID, chainIssues)

		err = verifier.redisCache.UpdateDAGChainIssues(context.Background(), projectID, chainIssues)
		if err != nil {
			log.WithError(err).Errorf("Failed to update dag chain issues in cache")
		}

		verifier.dagChainHasIssues = true
		verifier.lastVerifiedDagBlockHeightsMutex.Lock()
		verifier.dagChainIssues[projectID] = chainIssues
		verifier.lastVerifiedDagBlockHeightsMutex.Unlock()
	}

	// Store last verified blockHeight so that in next run, we just need to verify from the same.
	// Use single hash key in redis to store the same against contractAddress.
	verifier.lastVerifiedDagBlockHeightsMutex.Lock()
	verifier.lastVerifiedDagBlockHeights[projectID] = strconv.Itoa(endHeight)
	verifier.lastVerifiedDagBlockHeightsMutex.Unlock()

	// update the start and current chain height
	if verifier.ProjectsIndexedState[projectID].StartSourceChainHeight == 0 {
		verifier.ProjectsIndexedState[projectID].StartSourceChainHeight, err = verifier.FetchStartIndex(projectID)
		if err != nil {
			log.Errorf("Failed to fetch startIndex for project %s due to error %+v", projectID, err)
		}
		log.Infof("Updating startIndex for project %s as %d", projectID, verifier.ProjectsIndexedState[projectID].StartSourceChainHeight)
	}

	verifier.ProjectsIndexedState[projectID].CurrentSourceChainHeight = dagChainBlocks[len(dagChainBlocks)-1].Payload.Data.ChainHeightRange.End

	log.Infof("Updating currentIndex for project %s as %d", projectID, verifier.ProjectsIndexedState[projectID].CurrentSourceChainHeight)

	return nil
}

//func (verifier *DagVerifier) updateDagIssuesInRedis(projectId string, chainGaps []*datamodel.DagChainIssue) {
//	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_DAG_CHAIN_GAPS, projectId)
//	var gaps []*redis.Z
//	for i := range chainGaps {
//		gapStr, err := json.Marshal(chainGaps[i])
//		if err != nil {
//			log.Error("Serious issue if json marshal fails..Can't do anything else than continue and log error:", err)
//			continue
//		}
//		gaps = append(gaps, &redis.Z{Score: float64(chainGaps[i].DAGBlockHeight),
//			Member: gapStr,
//		})
//	}
//	for j := 0; j < 3; j++ {
//		res := verifier.redisClient.ZAdd(context.Background(), key, gaps...)
//		if res.Err() != nil {
//			log.Error("Failed to update dagChainGaps into redis for projectID:", projectId, ", GapData:", chainGaps)
//			time.Sleep(5 * time.Second)
//			continue
//		}
//		log.Infof("Added %d DagGaps data successfully in redis for project: %s", len(chainGaps), projectId)
//		break
//	}
//	//TODO: Need to prune older gaps.
//}

func (verifier *DagVerifier) SummarizeDAGIssuesAndNotify() error {
	dagSummary := new(datamodel.DagChainReport)
	dagSummary.InstanceId = verifier.settings.InstanceId
	dagSummary.HostName, _ = os.Hostname()

	var currentMinChainHeight int64
	currentCycleDAGChainHeight := make(map[string]int64, len(verifier.projects))
	currentMinChainHeight, _ = strconv.ParseInt(verifier.lastVerifiedDagBlockHeights[verifier.projects[0]], 10, 64)
	isDagChainStuckForAnyProject := 0
	summaryProjectsMovingAheadAfterStuck := false

	// Check if dag chain is stuck for any project.
	for _, projectId := range verifier.projects {
		var err error
		// Should we instead find min-Height for all projects??
		currentCycleDAGChainHeight[projectId], err = strconv.ParseInt(verifier.lastVerifiedDagBlockHeights[projectId], 10, 64)
		if err != nil {
			log.Errorf("LastVerifierDAGBlockHeight for project %s is not int and value is %s", projectId, verifier.lastVerifiedDagBlockHeights[projectId])

			return err
		}

		if currentCycleDAGChainHeight[projectId] != 0 {
			if currentCycleDAGChainHeight[projectId] < currentMinChainHeight {
				currentMinChainHeight = currentCycleDAGChainHeight[projectId]
			}
			if verifier.previousCycleDagChainHeight[projectId] == currentCycleDAGChainHeight[projectId] {
				verifier.noOfCyclesSinceChainStuck[projectId]++
			} else {
				verifier.noOfCyclesSinceChainStuck[projectId] = 0
			}
			if verifier.noOfCyclesSinceChainStuck[projectId] > 3 {
				isDagChainStuckForAnyProject++
				verifier.dagChainHasIssues = true
				log.Infof("Dag chain is stuck for project %s at DAG height %d from past 3 cycles of run.",
					projectId, currentCycleDAGChainHeight[projectId])
			}
		}
	}
	isDagchainStuckForSummaryProject := 0
	for j := range verifier.SummaryProjects {
		projectId := verifier.SummaryProjects[j]
		currentDagHeight := verifier.GetProjectDAGBlockHeightFromRedis(projectId)
		if currentDagHeight == "0" {
			log.Debugf("Project's %s height is 0 and not moved ahead. Skipping check for stuck", projectId)
			continue
		}
		if currentDagHeight == verifier.lastVerifiedDagBlockHeights[projectId] {
			verifier.noOfCyclesSinceChainStuck[projectId]++
			if verifier.noOfCyclesSinceChainStuck[projectId] > 2 {
				log.Errorf("DAG Chain stuck for summary project %s at height %s", projectId, currentDagHeight)
				isDagchainStuckForSummaryProject++
				var summaryProject datamodel.SummaryProjectVerificationStatus
				summaryProject.ProjectHeight = currentDagHeight
				summaryProject.ProjectId = projectId
				dagSummary.Severity = DAG_CHAIN_REPORT_SEVERITY_HIGH
				dagSummary.SummaryProjectsStuckDetails = append(dagSummary.SummaryProjectsStuckDetails, summaryProject)
			}
			summaryProjectsMovingAheadAfterStuck = false
		} else {
			if verifier.noOfCyclesSinceChainStuck[projectId] > 2 {
				summaryProjectsMovingAheadAfterStuck = true
				var summaryProject datamodel.SummaryProjectVerificationStatus
				summaryProject.ProjectId = projectId
				summaryProject.ProjectHeight = currentDagHeight
				dagSummary.Severity = DAG_CHAIN_REPORT_SEVERITY_CLEAR
				dagSummary.SummaryProjectsRecovered = append(dagSummary.SummaryProjectsRecovered, summaryProject)
			}
			verifier.noOfCyclesSinceChainStuck[projectId] = 0
		}
		verifier.lastVerifiedDagBlockHeights[projectId] = currentDagHeight
	}
	//Check if dagChain has issues for any project.
	if verifier.dagChainHasIssues {
		dagSummary.ProjectsTrackedCount = len(verifier.projects)
		dagSummary.ProjectsWithIssuesCount = len(verifier.dagChainIssues)
		dagSummary.CurrentMinChainHeight = currentMinChainHeight
		dagSummary.ProjectsWithStuckChainCount = isDagChainStuckForAnyProject
		dagSummary.ProjectsWithCacheIssueCount = verifier.dagCacheIssues

		for _, projectIssues := range verifier.dagChainIssues {
			dagSummary.OverallIssueCount += len(projectIssues)
			for i := range projectIssues {
				if projectIssues[i].IssueType == DAG_CHAIN_ISSUE_DUPLICATE_HEIGHT {
					dagSummary.OverallDAGChainDuplicates += 1
				} else if projectIssues[i].IssueType == DAG_CHAIN_ISSUE_GAP_IN_CHAIN {
					dagSummary.OverallDAGChainGaps += 1
				}
			}
		}
		dagSummary.Severity = DAG_CHAIN_REPORT_SEVERITY_HIGH
	} else if verifier.dagCacheIssues > 0 {
		var dagSummary datamodel.DagChainReport
		dagSummary.ProjectsTrackedCount = len(verifier.projects)
		dagSummary.Severity = DAG_CHAIN_REPORT_SEVERITY_LOW
		dagSummary.ProjectsWithCacheIssueCount = verifier.dagCacheIssues
		dagSummary.CurrentMinChainHeight = currentMinChainHeight
		verifier.NotifySlack(&dagSummary)

		verifier.dagCacheIssues = 0
	}

	//Do not notify if recently notification has been sent.
	//TODO: Rough suppression logic, not very elegant.
	//Better to have a method to clear this notificationTime manually via SIGUR or some other means once problem is addressed.
	//As of now the workaround is to restart dag-verifier once issues are resolved.
	if verifier.dagChainHasIssues || isDagChainStuckForAnyProject > 0 || isDagchainStuckForSummaryProject > 0 {
		if time.Now().Unix()-verifier.lastNotifyTime > verifier.settings.DagVerifierSettings.SuppressNotificationTimeSecs {
			verifier.NotifySlack(dagSummary)
			verifier.lastNotifyTime = time.Now().Unix()
		}
	}
	if summaryProjectsMovingAheadAfterStuck {
		verifier.NotifySlack(dagSummary)
	}
	//Cleanup reported issues, because either they auto-recover or a manual recovery is needed.
	verifier.dagChainHasIssues = false
	verifier.dagChainIssues = make(map[string][]*datamodel.DagChainIssue)

	for _, projectId := range verifier.projects {
		verifier.previousCycleDagChainHeight[projectId] = currentCycleDAGChainHeight[projectId]
	}

	return nil
}

func (verifier *DagVerifier) NotifySlack(dagSummary *datamodel.DagChainReport) error {
	dagSummaryStr, _ := json.MarshalIndent(dagSummary, "", "\t")
	//err := verifier.NotifySlackOfDAGSummary(dagSummary)
	err := slackutils.NotifySlackWorkflow(string(dagSummaryStr), dagSummary.Severity, "DAGVerifier")
	if err != nil {
		log.Errorf("Slack Notify failed with error %+v", err)
	}
	return err
}

func (verifier *DagVerifier) GetDagChainCidsFromRedis(projectId string, startScore string) ([]datamodel.DagBlock, error) {
	var dagChainCids []datamodel.DagBlock

	//key := projectId + ":payloadCids"
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectId)

	log.Debug("Fetching DAG Chain Cids from redis at key:", key, ",with startScore: ", startScore)
	zRangeByScore := verifier.redisClient.ZRangeByScoreWithScores(context.Background(), key, &redis.ZRangeBy{
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
	dagChainCids = make([]datamodel.DagBlock, len(res))
	log.Debugf("Fetched %d DAG Chain CIDs for key %s", len(res), key)
	for i := range res {
		//Safe to convert as we know height will always be int.
		dagChainCids[i].Height = int64(res[i].Score)
		dagChainCids[i].CurrentCid = fmt.Sprintf("%v", res[i].Member)
	}
	lastVerifiedStatus, err := strconv.ParseInt(startScore, 10, 64)
	if err != nil && len(dagChainCids) == 1 && dagChainCids[0].Height == lastVerifiedStatus {
		//This means no new dag blocks in the chain from last verified height.
		dagChainCids = nil
	}
	return dagChainCids, nil
}

// Need to handle Dagchain reorg event and reset the lastVerifiedBlockHeight to the same.
/*func (verifier *DagVerifier) HandleChainReOrg(){

}*/

func (verifier *DagVerifier) GetPayloadCidsFromRedis(projectId string, startScore string, dagChain []datamodel.DagBlock) error {

	//key := projectId + ":payloadCids"
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectId)

	log.Debug("Fetching PayloadCids from redis at key:", key, ",with startScore: ", startScore)
	zRangeByScore := verifier.redisClient.ZRangeByScoreWithScores(context.Background(), key, &redis.ZRangeBy{
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

	//Hate GOTO's but felt this is the cleanest way to handle this.
RESTART_CID_COMP_LOOP:
	for i := range res {
		if len(dagChain) > i {
			/* TODO: Implement this logic to Check for sequence of payloadCIDS
			if i != len(dagChain)-1 {
				if res[i].Score != res[i+1].Score-1 {
					log.Warnf("PayloadCids are out of sequence in redis cache for project %s. Missing payloadCid at score %d", projectId, res[i].Score+1)
				}
			}*/
			if dagChain[i].Height != int64(res[i].Score) {
				//Handling special case of duplicate entries in redis DAG cache, this doesn't affect original DAGChain that is in IPFS.
				if (i > 0) &&
					(dagChain[i].Height == dagChain[i-1].Height) {
					dagBlock, err := verifier.ipfsClient.GetDagBlock(dagChain[i].CurrentCid)
					if err != nil {
						log.Infof("Failed to fetch DAGblock %s from IPFS due to error %+v..retrying in next cycle",
							dagChain[i].CurrentCid, err)
						return nil, errors.New("failed to fetch DAGblock from IPFS due to error")
					}
					dagBlock1, err := verifier.ipfsClient.GetDagBlock(dagChain[i-1].CurrentCid)
					if err != nil {
						log.Infof("Failed to fetch DAGblock %s from IPFS due to error %+v..retrying in next cycle",
							dagChain[i-1].CurrentCid, err)
						return nil, errors.New("failed to fetch DAGblock from IPFS due to error")
					}
					if dagBlock.Data.PayloadLink == dagBlock1.Data.PayloadLink {
						log.Warnf("Duplicate entry found in redis cache at DAGChain Height %d in cache for Project %s", dagChain[i].Height, projectId)
						//Notify of a minor issue and proceed by removing the duplicate entry so that verification proceed
						copy(dagChain[i:], dagChain[i+1:])
						dagChain = dagChain[:len(dagChain)-1]
						verifier.dagCacheIssues++
						//TODO: Should we auto-correct the cache or let it be?
						goto RESTART_CID_COMP_LOOP
					} else {
						log.Errorf("Payloads at DAG blocks for project %s are not same and are different in DAGs %+v and %+v .",
							projectId, dagBlock, dagBlock1)
					}
				}
				return dagChain, fmt.Errorf("CRITICAL:Inconsistency between DAG Chain and Payloads stored in redis for Project:%s", projectId)
			}
			dagChain[i].Payload.PayloadCid = fmt.Sprintf("%v", res[i].Member)
			if strings.HasPrefix(dagChain[i].Payload.PayloadCid, "null") {
				//Case where DAG Chain is self-healed because no consensus was achieved.
				//Remove the DAG entry so that it would be reported as gap in chain.
				log.Warnf("Empty payload found in redis cache at DAGChain Height %d for Project %s. Consensus has not been achieved at this blockHeight",
					dagChain[i].Height, projectId)
				copy(dagChain[i:], dagChain[i+1:])
				dagChain = dagChain[:len(dagChain)-1]
				copy(res[i:], res[i+1:])
				res = res[:len(res)-1]
				goto RESTART_CID_COMP_LOOP
			}
		} else {
			log.Debugf("DAGChain height is %d for project %s is little behind payloadCids chainHeight %d.", len(dagChain), projectId, len(res))
			return nil, errors.New("chain construction seems to be in progress. Have to retry next time")
		}
	}
	return dagChain, nil
}

func (verifier *DagVerifier) checkInvalidEntriesInCache(chain []*datamodel.DagBlock, projectID string) error {
restart:
	for {
		for index, block := range chain {
			// stop if we are at the penultimate block
			if index == len(chain)-2 {
				break restart
			}

			// check if the height of the current block is equal to the next block
			if block.Height != chain[index+1].Height {
				continue
			}

			// fetch the dag blocks from IPFS
			dagBlock, err := verifier.ipfsClient.GetDagBlock(block.CurrentCid)
			if err != nil {
				log.Infof("Failed to fetch DAGblock %s from IPFS due to error %+v", block.CurrentCid, err)
				return err
			}

			dagBlock1, err := verifier.ipfsClient.GetDagBlock(chain[index+1].CurrentCid)
			if err != nil {
				log.Infof("Failed to fetch DAGblock %s from IPFS due to error %+v", chain[index+1].CurrentCid, err)

				return err
			}

			// check for duplicate entry
			if dagBlock.Data.PayloadLink == dagBlock1.Data.PayloadLink {
				log.Warnf("Duplicate entry found in redis cache at DAGChain Height %d", block.Height)
				copy(chain[index:], chain[index+1:])
				chain = chain[:len(chain)-1]
				verifier.dagCacheIssues++

				continue restart
			} else {
				// if the payloads in dag blocks are not the same, then we have a problem in cache
				log.Errorf("payloads in DAG blocks %v and %v are not same but are at same height", dagBlock, dagBlock1)
			}

			if strings.HasPrefix(chain[index].Payload.PayloadCid, "null") {
				// Case where DAG Chain is self-healed because no consensus was achieved.
				// Remove the DAG entry so that it would be reported as gap in chain.
				log.Warnf("Empty payload found in redis cache at DAGChain Height %d for Project %s. Consensus has not been achieved at this blockHeight",
					chain[index].Height, projectID)

				copy(chain[index:], chain[index+1:])
				chain = chain[:len(chain)-1]

				continue restart
			}
		}

		break
	}

	return nil
}

func (verifier *DagVerifier) verifyDagChainForIssues(dagChain []*datamodel.DagBlock) (bool, []*datamodel.DagChainIssue) {
	log.Info("Verifying DAG for issues. DAG chain length is:", len(dagChain))

	var prevBlockChainRangeEnd, startHeightOfChain, endHeightOfChain, numGaps, numDuplicates int64
	var dagIssues []*datamodel.DagChainIssue

	startHeightOfChain = dagChain[0].Payload.Data.ChainHeightRange.End
	endHeightOfChain = dagChain[len(dagChain)-1].Payload.Data.ChainHeightRange.Begin

	for index, block := range dagChain {
		if index == 0 {
			continue
		}

		if block.Height == dagChain[index-1].Height {
			dagIssues = append(dagIssues, &datamodel.DagChainIssue{
				IssueType:           DAG_CHAIN_ISSUE_DUPLICATE_HEIGHT,
				TimestampIdentified: time.Now().Unix(),
				DAGBlockHeight:      block.Height,
			})
			//TODO:If there are multiple snapshots observed at same blockHeight, need to take action to cleanup snapshots
			//		which are not required from IPFS based on previous and next blocks.
			log.Errorf("Found Same DagchainHeight %d at 2 levels. DagChain needs to be fixed.", block.Height)

			numDuplicates++
		}

		curBlockChainRangeStart := block.Payload.Data.ChainHeightRange.Begin

		if curBlockChainRangeStart != 0 && curBlockChainRangeStart != prevBlockChainRangeEnd+1 {
			log.Debug("Gap identified at ChainHeight:", block.Height, ",PayloadLink:", block.Payload.PayloadCid, ", between height:", dagChain[index-1].Height, " and ", dagChain[i].Height)
			log.Debug("Missing blocks from(not including): ", prevBlockChainRangeEnd, " to(not including): ", curBlockChainRangeStart)
			dagIssues = append(dagIssues, &datamodel.DagChainIssue{
				IssueType:               DAG_CHAIN_ISSUE_GAP_IN_CHAIN,
				MissingBlockHeightStart: prevBlockChainRangeEnd + 1,
				MissingBlockHeightEnd:   curBlockChainRangeStart - 1, // How can -1 be recorded??
				TimestampIdentified:     time.Now().Unix(),
				DAGBlockHeight:          block.Height},
			)
			numGaps++
		}

		// when loop starts again prevBlockChainRangeEnd will become prev block's chain range end
		prevBlockChainRangeEnd = block.Payload.Data.ChainHeightRange.End
	}

	log.Info("Block Range is from:", startHeightOfChain, ", to:", endHeightOfChain)
	log.Infof("Number of gaps found:%d. Number of Duplicates found:%d", numGaps, numDuplicates)
	if numGaps == 0 && numDuplicates == 0 {
		return false, nil
	} else {
		return true, dagIssues
	}
}

func (verifier *DagVerifier) PopulateProjects() error {
	log.Debugf("Fetching stored Projects from redis at key: %s", redisutils.REDIS_KEY_STORED_PROJECTS)

	projects, err := verifier.redisCache.GetStoredProjects(context.Background())
	if err != nil {
		return err
	}

	verifier.projects = make([]string, 0, len(projects))

	for _, projectID := range projects {
		isSummaryProject := false
		for _, summaryProject := range verifier.settings.DagVerifierSettings.SummaryProjectsToTrack {
			if strings.Contains(strings.ToLower(projectID), strings.ToLower(summaryProject)) {
				log.Infof("Removing summary Project %s from tracking", projectID)
				isSummaryProject = true
				break
			}
		}

		if !isSummaryProject {
			verifier.projects = append(verifier.projects, projectID)
		} else {
			// TODO: SummaryProject tracking to be added back by finding out from projectState
			verifier.SummaryProjects = append(verifier.SummaryProjects, projectID)
		}
	}

	verifier.noOfCyclesSinceChainStuck = make(map[string]int, len(verifier.projects))
	verifier.previousCycleDagChainHeight = make(map[string]int64, len(verifier.projects))

	log.Infof("Retrieved %d storedProjects %+v from redis", len(verifier.projects), verifier.projects)

	return nil
}
