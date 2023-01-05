package main

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

	"github.com/go-redis/redis/v8"
	"github.com/powerloom/goutils/filecache"
	"github.com/powerloom/goutils/redisutils"
	"github.com/powerloom/goutils/settings"
	"github.com/powerloom/goutils/slackutils"
	log "github.com/sirupsen/logrus"
)

var ctx = context.Background()

type ProjectIndexedState struct {
	StartSourceChainHeight   int64 `json:"startSourceChainHeight"`
	CurrentSourceChainHeight int64 `json:"currentSourceChainHeight"`
}

type DagVerifier struct {
	redisClient                      *redis.Client
	projects                         []string
	SummaryProjects                  []string
	settings                         *settings.SettingsObj
	lastVerifiedDagBlockHeights      map[string]string
	ProjectsIndexedState             map[string]*ProjectIndexedState
	lastVerifiedDagBlockHeightsMutex *sync.RWMutex
	dagChainHasIssues                bool
	dagCacheIssues                   int
	dagChainIssues                   map[string][]DagChainIssue
	previousCycleDagChainHeight      map[string]int64
	noOfCyclesSinceChainStuck        map[string]int
	lastNotifyTime                   int64
}

const DAG_CHAIN_ISSUE_DUPLICATE_HEIGHT string = "DUPLICATE_HEIGHT_IN_CHAIN"
const DAG_CHAIN_ISSUE_GAP_IN_CHAIN string = "GAP_IN_CHAIN"

const DAG_CHAIN_REPORT_SEVERITY_HIGH = "High"
const DAG_CHAIN_REPORT_SEVERITY_MEDIUM = "Medium"
const DAG_CHAIN_REPORT_SEVERITY_LOW = "Low"
const DAG_CHAIN_REPORT_SEVERITY_CLEAR = "Cleared"

func (verifier *DagVerifier) Initialize(settings *settings.SettingsObj) {
	verifier.settings = settings
	verifier.InitIPFSClient()
	verifier.InitRedisClient()
	verifier.InitSlackClient()

	verifier.dagChainIssues = make(map[string][]DagChainIssue)

	verifier.PopulateProjects()

	//Fetch DagChain verification status from redis for all projects.
	verifier.FetchLastVerificationStatusFromRedis()
	verifier.FetchLastProjectIndexedStatusFromRedis()
	verifier.lastVerifiedDagBlockHeightsMutex = &sync.RWMutex{}
}

// TODO: REuse function from pruningVerifier
func (verifier *DagVerifier) PopulateProjects() {
	for {
		log.Debugf("Fetching stored Projects from redis at key: %s", redisutils.REDIS_KEY_STORED_PROJECTS)
		res := verifier.redisClient.SMembers(ctx, redisutils.REDIS_KEY_STORED_PROJECTS)
		if res.Err() != nil {
			if res.Err() == redis.Nil {
				log.Warnf("Stored Projects key doesn't exist..retrying")
				time.Sleep(30 * time.Minute)
				continue
			}
			log.Errorf("Failed to fetch stored projects from redis due to err %+v. Retrying", res.Err())
			time.Sleep(5 * time.Second)
			continue
		}
		if len(res.Val()) == 0 {
			log.Infof("No projects registered. Will wait initializing until projects are registered")
			time.Sleep(30 * time.Second)
			continue
		}
		verifier.projects = make([]string, 0, len(res.Val()))

		for i := range res.Val() {
			projectId := res.Val()[i]
			skipAdd := false
			//TODO: SummaryPRoject tracking to be added back by finding out from projectState
			for j := range verifier.settings.DagVerifierSettings.SummaryProjectsToTrack {
				if strings.Contains(projectId, verifier.settings.DagVerifierSettings.SummaryProjectsToTrack[j]) {
					log.Infof("Removing summary Project %s from tracking", projectId)
					skipAdd = true
					break
				}
			}
			if !skipAdd {
				verifier.projects = append(verifier.projects, projectId)
			}
		}
		break
	}
	verifier.noOfCyclesSinceChainStuck = make(map[string]int, len(verifier.projects))
	verifier.previousCycleDagChainHeight = make(map[string]int64, len(verifier.projects))
	log.Infof("Retrieved %d storedProjects %+v from redis", len(verifier.projects), verifier.projects)
}

func (verifier *DagVerifier) FetchLastProjectIndexedStatusFromRedis() {
	key := redisutils.REDIS_KEY_PROJECTS_INDEX_STATUS
	log.Debug("Fetching Projects Index Status at key:", key)

	res := verifier.redisClient.HGetAll(ctx, key)
	if res.Err() != nil {
		log.Error("Ideally should not come here, which means there is some other redis error. To debug:", res.Err())
	}
	verifier.ProjectsIndexedState = make(map[string]*ProjectIndexedState)
	if len(res.Val()) == 0 {
		log.Info("Failed to fetch Projects Index Status from redis.")
		//Key doesn't exist.
		log.Info("Key doesn't exist..hence find the starting index range for projects.")
		for i := range verifier.projects {
			startIndex, err := verifier.FetchStartIndex(verifier.projects[i])
			if err != nil {
				log.Errorf("Could not fetch Start index for project %s due to error %+v", verifier.projects[i], err)
			}
			log.Infof("Fetched startIndex for project %s as %d", verifier.projects[i], startIndex)
			projectIndexedState := ProjectIndexedState{
				StartSourceChainHeight:   startIndex,
				CurrentSourceChainHeight: startIndex,
			}
			verifier.ProjectsIndexedState[verifier.projects[i]] = &projectIndexedState
		}
		return
	}
	//TODO: Update when handling dynamic addition of projects.
	for i := range verifier.projects {
		var projectIndexedState ProjectIndexedState
		projectId := verifier.projects[i]
		if val, ok := res.Val()[projectId]; ok {
			err := json.Unmarshal([]byte(val), &projectIndexedState)
			if err != nil {
				log.Errorf("Could not fetch previous indexed state for project %s due to error %+v", verifier.projects[i], err)
			}
		}
		verifier.ProjectsIndexedState[verifier.projects[i]] = &projectIndexedState
		log.Infof("Updating startIndex for project %s as %d", projectId, projectIndexedState.StartSourceChainHeight)
	}
	log.Debugf("Fetched LastVerificationStatus from redis %+v", verifier.lastVerifiedDagBlockHeights)
}

func (verifier *DagVerifier) FetchStartIndex(projectId string) (int64, error) {
	payloadCid, err := verifier.GetPayloadCidAtDAGHeightFromRedis(projectId, "1")
	if err != nil {
		//Raise an alarm in future for this
		log.Error("Failed to fetch DAG Chain CIDS for projectID from redis:", projectId)
		return 0, err
	}
	if payloadCid == "" {
		log.Info("Indexing has not started for projectId", projectId)
		return 0, nil
	}
	payload, err := verifier.GetPayloadFromCache(projectId, payloadCid)
	if err != nil {
		//Fetch CID from IPFS and store the startHeight.
		payload, err = ipfsClient.GetPayload(payloadCid, 3)
		if err != nil {
			log.Errorf("Failed to fetch payloadCID %s from IPFS for project %s", payloadCid, projectId)
			return 0, nil
		}
	}
	return payload.ChainHeightRange.Begin, nil
}

func (verifier *DagVerifier) GetPayloadCidAtDAGHeightFromRedis(projectId string, startScore string) (string, error) {
	//key := projectId + ":payloadCids"
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectId)
	payloadCid := ""

	log.Debug("Fetching PayloadCid from redis at key:", key, ",with startScore: ", startScore)
	zRangeByScore := verifier.redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: startScore,
		Max: startScore,
	})

	err := zRangeByScore.Err()
	log.Debug("Result for ZRangeByScoreWithScores : ", zRangeByScore)
	if err != nil {
		log.Error("Could not fetch entries error: ", err, "Query:", zRangeByScore)
		return "", err
	}
	res := zRangeByScore.Val()
	//dagPayloadsInfo = make([]DagPayload, len(res))
	log.Debugf("Fetched %d Payload CIDs for key %s", len(res), key)
	if len(res) == 1 {
		payloadCid = fmt.Sprintf("%v", res[0].Member)
	}
	return payloadCid, nil
}

func (verifier *DagVerifier) FetchLastVerificationStatusFromRedis() {
	key := fmt.Sprintf(redisutils.REDIS_KEY_DAG_VERIFICATION_STATUS)
	log.Debug("Fetching LastVerificationStatusFromRedis at key:", key)

	res := verifier.redisClient.HGetAll(ctx, key)

	if len(res.Val()) == 0 {
		log.Info("Failed to fetch LastverificationStatus from redis for the projects.")
		//Key doesn't exist.
		log.Info("Key doesn't exist..hence proceed from start of the block.")
		verifier.lastVerifiedDagBlockHeights = make(map[string]string)
		for i := range verifier.projects {
			//TODO: do we need to change this to be based on some timeDuration like last 24h etc instead of starting from 1??
			verifier.lastVerifiedDagBlockHeights[verifier.projects[i]] = "0"
		}
		for i := range verifier.SummaryProjects {
			verifier.lastVerifiedDagBlockHeights[verifier.SummaryProjects[i]] = "0"
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
	//No retry has been added, because in case of a failure, status will get updated in next run.
	key := fmt.Sprintf(redisutils.REDIS_KEY_DAG_VERIFICATION_STATUS)
	log.Info("Updating LastVerificationStatus at key:", key)
	res := verifier.redisClient.HMSet(ctx, key, verifier.lastVerifiedDagBlockHeights)
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
	res1 := verifier.redisClient.HMSet(ctx, key, projectsIndexedState)
	if res1.Err() != nil {
		log.Error("Failed to update LastIndexedStatus in redis..Retry in next run.")
	}
}

func (verifier *DagVerifier) InitIPFSClient() {

	//Initialize and do a basic test to see if IPFS client is connected to IPFS server and is able to fetch.
	ipfsClient.Init(verifier.settings)
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
	periodicRetrievalInterval := time.Duration(verifier.settings.DagVerifierSettings.RunIntervalSecs) * time.Second

	for {
		if len(verifier.projects) > 0 {
			verifier.FetchLastProjectIndexedStatusFromRedis()
			verifier.FetchLastVerificationStatusFromRedis()
			verifier.VerifyAllProjects() //Projects are pairContracts
			verifier.SummarizeDAGIssuesAndNotifySlack()
			verifier.UpdateLastStatusToRedis()
		} else {
			log.Info("No projects to be verified. Have to check in next run.")
		}
		log.Info("Sleeping for " + periodicRetrievalInterval.String())
		time.Sleep(periodicRetrievalInterval)
	}
}

func (verifier *DagVerifier) VerifyAllProjects() {
	var wg sync.WaitGroup
	//TODO: change to batch logic as done in pruning service.
	for i := range verifier.projects {
		wg.Add(1)

		go func(index int) {
			defer wg.Done()
			verifier.VerifyDagChain(verifier.projects[index])
		}(i)

	}

	wg.Wait()
}

func (verifier *DagVerifier) GetProjectDAGBlockHeightFromRedis(projectId string) string {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_BLOCK_HEIGHT, projectId)
	for i := 0; i < 3; i++ {
		res := verifier.redisClient.Get(ctx, key)
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

func (verifier *DagVerifier) GetPayloadFromCache(projectId string, payloadCid string) (DagPayloadData, error) {
	var payload DagPayloadData
	//Remove projectID prefix, for now hard-coding it
	projectId = projectId[10:]
	bytes, err := filecache.ReadFromCache(verifier.settings.PayloadCachePath, projectId, payloadCid)
	if err != nil {
		if !strings.Contains(err.Error(), "no such file or directory") {
			log.Errorf("Failed to fetch payload with cid %s for project %s from cache due to error %+v", payloadCid, projectId, err)
		}
		return payload, err
	}
	err = json.Unmarshal(bytes, &payload)
	if err != nil {
		log.Error("Failed to Unmarshal Json Payload from IPFS, CID:", payloadCid, ", bytes:", bytes, ", error:", err)
		return payload, err
	}
	return payload, nil
}

func (verifier *DagVerifier) VerifyDagChain(projectId string) error {
	//Fetch the DAGChain cached in redis and then corresponding payloads at chainHeight.
	// For now only the Chain that is stored in redis is used as a reference to verify.
	//TODO: Need to validate the original dag chain from what is stored in IPFS. Is this required??
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
		payload, err := verifier.GetPayloadFromCache(projectId, dagChain[i].Payload.PayloadCid)
		if err != nil {
			payload, err = ipfsClient.GetPayload(dagChain[i].Payload.PayloadCid, 3)
			if err != nil {
				//If we are unable to fetch a CID from IPFS, retry
				log.Error("Failed to get PayloadCID from IPFS. Either cache is corrupt or there is an actual issue.CID:", dagChain[i].Payload.PayloadCid)
				//TODO: Either cache is corrupt or there is an actual issue.
				//Check and fix cache corruption by getting Dagchain from IPFS.
				return err
			}
		}
		dagChain[i].Payload.Data = payload
		//Fetch payload from IPFS and check gaps in chainHeight.\
		log.Debugf("Index: %d ,payload: %+v", i, payload)
	}
	log.Infof("Verifying Dagchain for ProjectId %s , from block %d to %d", projectId, dagChain[0].Height, dagChain[(len(dagChain)-1)].Height)
	issuesPresent, chainIssues := verifier.verifyDagForIssues(&dagChain)
	if !issuesPresent {
		log.Infof("Dag chain has issues for projectID %s. Issues are: %+v", projectId, chainIssues)
		verifier.updateDagIssuesInRedis(projectId, chainIssues)
		verifier.dagChainHasIssues = true
		verifier.lastVerifiedDagBlockHeightsMutex.Lock()
		verifier.dagChainIssues[projectId] = chainIssues
		verifier.lastVerifiedDagBlockHeightsMutex.Unlock()
	}
	//Store last verified blockHeight so that in next run, we just need to verify from the same.
	//Use single hash key in redis to store the same against contractAddress.
	verifier.lastVerifiedDagBlockHeightsMutex.Lock()
	verifier.lastVerifiedDagBlockHeights[projectId] = strconv.FormatInt(dagChain[len(dagChain)-1].Height, 10)
	if verifier.ProjectsIndexedState[projectId].StartSourceChainHeight == 0 {
		verifier.ProjectsIndexedState[projectId].StartSourceChainHeight, err = verifier.FetchStartIndex(projectId)
		if err != nil {
			log.Errorf("Failed to fetch startIndex for project %s due to error %+v", projectId, err)
		}
		log.Infof("Updating startIndex for project %s as %d", projectId,
			verifier.ProjectsIndexedState[projectId].StartSourceChainHeight)
	}
	verifier.ProjectsIndexedState[projectId].CurrentSourceChainHeight = dagChain[len(dagChain)-1].Payload.Data.ChainHeightRange.End
	log.Infof("Updating currentIndex for project %s as %d", projectId,
		verifier.ProjectsIndexedState[projectId].CurrentSourceChainHeight)
	verifier.lastVerifiedDagBlockHeightsMutex.Unlock()
	return nil
}

func (verifier *DagVerifier) updateDagIssuesInRedis(projectId string, chainGaps []DagChainIssue) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_DAG_CHAIN_GAPS, projectId)
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
	for j := 0; j < 3; j++ {
		res := verifier.redisClient.ZAdd(ctx, key, gaps...)
		if res.Err() != nil {
			log.Error("Failed to update dagChainGaps into redis for projectID:", projectId, ", GapData:", chainGaps)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Infof("Added %d DagGaps data successfully in redis for project: %s", len(chainGaps), projectId)
		break
	}
	//TODO: Need to prune older gaps.
}

func (verifier *DagVerifier) SummarizeDAGIssuesAndNotifySlack() {
	var dagSummary DagChainReport
	dagSummary.InstanceId = verifier.settings.InstanceId
	dagSummary.HostName, _ = os.Hostname()
	currentCycleDAGchainHeight := make(map[string]int64, len(verifier.projects))
	var currentMinChainHeight int64
	currentMinChainHeight, _ = strconv.ParseInt(verifier.lastVerifiedDagBlockHeights[verifier.projects[0]], 10, 64)
	isDagchainStuckForAnyProject := 0
	summaryProjectsMovingAheadAfterStuck := false
	//Check if dag chain is stuck for any project.
	for _, projectId := range verifier.projects {
		var err error
		//Should we instead find min-Height for all projects??
		currentCycleDAGchainHeight[projectId], err = strconv.ParseInt(verifier.lastVerifiedDagBlockHeights[projectId], 10, 64)
		if err != nil {
			log.Fatalf("LastVerifierDAGBlockHeight for project %s is not int and value is %s",
				projectId, verifier.lastVerifiedDagBlockHeights[projectId])
			return
		}
		if currentCycleDAGchainHeight[projectId] < currentMinChainHeight {
			currentMinChainHeight = currentCycleDAGchainHeight[projectId]
		}
		if verifier.previousCycleDagChainHeight[projectId] == currentCycleDAGchainHeight[projectId] {
			verifier.noOfCyclesSinceChainStuck[projectId]++
		} else {
			verifier.noOfCyclesSinceChainStuck[projectId] = 0
		}
		if verifier.noOfCyclesSinceChainStuck[projectId] > 3 {
			isDagchainStuckForAnyProject++
			verifier.dagChainHasIssues = true
			log.Infof("Dag chain is stuck for project %s at DAG height %d from past 3 cycles of run.",
				projectId, currentCycleDAGchainHeight[projectId])
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
				var summaryProject SummaryProjectState
				summaryProject.ProjectHeight = currentDagHeight
				summaryProject.ProjectId = projectId
				dagSummary.Severity = DAG_CHAIN_REPORT_SEVERITY_HIGH
				dagSummary.SummaryProjectsStuckDetails = append(dagSummary.SummaryProjectsStuckDetails, summaryProject)
			}
			summaryProjectsMovingAheadAfterStuck = false
		} else {
			if verifier.noOfCyclesSinceChainStuck[projectId] > 2 {
				summaryProjectsMovingAheadAfterStuck = true
				var summaryProject SummaryProjectState
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
		dagSummary.ProjectsWithStuckChainCount = isDagchainStuckForAnyProject
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
		var dagSummary DagChainReport
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
	if verifier.dagChainHasIssues || isDagchainStuckForAnyProject > 0 || isDagchainStuckForSummaryProject > 0 {
		if time.Now().Unix()-verifier.lastNotifyTime > verifier.settings.DagVerifierSettings.SuppressNotificationTimeSecs {
			verifier.NotifySlack(&dagSummary)
			verifier.lastNotifyTime = time.Now().Unix()
		}
	}
	if summaryProjectsMovingAheadAfterStuck {
		verifier.NotifySlack(&dagSummary)
	}
	//Cleanup reported issues, because either they auto-recover or a manual recovery is needed.
	verifier.dagChainHasIssues = false
	verifier.dagChainIssues = make(map[string][]DagChainIssue)

	for _, projectId := range verifier.projects {
		verifier.previousCycleDagChainHeight[projectId] = currentCycleDAGchainHeight[projectId]
	}

}

func (verifier *DagVerifier) NotifySlack(dagSummary *DagChainReport) error {
	dagSummaryStr, _ := json.MarshalIndent(dagSummary, "", "\t")
	//err := verifier.NotifySlackOfDAGSummary(dagSummary)
	err := slackutils.NotifySlackWorkflow(string(dagSummaryStr), dagSummary.Severity, "DAGVerifier")
	if err != nil {
		log.Errorf("Slack Notify failed with error %+v", err)
	}
	return err
}

// Need to handle Dagchain reorg event and reset the lastVerifiedBlockHeight to the same.
/*func (verifier *DagVerifier) HandleChainReOrg(){

}*/

func (verifier *DagVerifier) GetDagChainCidsFromRedis(projectId string, startScore string) ([]DagChainBlock, error) {
	var dagChainCids []DagChainBlock

	//key := projectId + ":payloadCids"
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectId)

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
		dagChainCids[i].CurrentCid = fmt.Sprintf("%v", res[i].Member)
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
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectId)

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
					dagBlock, err := ipfsClient.DagGet(dagChain[i].CurrentCid)
					if err != nil {
						log.Infof("Failed to fetch DAGblock %s from IPFS due to error %+v..retrying in next cycle",
							dagChain[i].CurrentCid, err)
						return nil, errors.New("failed to fetch DAGblock from IPFS due to error")
					}
					dagBlock1, err := ipfsClient.DagGet(dagChain[i-1].CurrentCid)
					if err != nil {
						log.Infof("Failed to fetch DAGblock %s from IPFS due to error %+v..retrying in next cycle",
							dagChain[i-1].CurrentCid, err)
						return nil, errors.New("failed to fetch DAGblock from IPFS due to error")
					}
					if dagBlock.Data.Cid == dagBlock1.Data.Cid {
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
			if dagChain[i].Payload.PayloadCid == "None" {
				//Case where DAG Chain is self-healed because no consensus was achieved.
				//Remove the DAG entry so that it would be reported as gap in chain.
				log.Warnf("Empty payload found in redis cache at DAGChain Height %d for Project %s. Consensus has not been achieved at this blockHeight",
					dagChain[i].Height, projectId)
				copy(dagChain[i:], dagChain[i+1:])
				dagChain = dagChain[:len(dagChain)-1]
			}
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
		//TODO: Add logic of out of order identification

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
			if curBlockStart != 0 && curBlockStart != prevDagBlockEnd+1 {
				log.Debug("Gap identified at ChainHeight:", dagChain[i].Height, ",PayloadCID:", dagChain[i].Payload.PayloadCid, ", between height:", dagChain[i-1].Height, " and ", dagChain[i].Height)
				log.Debug("Missing blocks from(not including): ", prevDagBlockEnd, " to(not including): ", curBlockStart)
				dagIssues = append(dagIssues, DagChainIssue{IssueType: DAG_CHAIN_ISSUE_GAP_IN_CHAIN, MissingBlockHeightStart: prevDagBlockEnd + 1,
					MissingBlockHeightEnd: curBlockStart - 1, // How can -1 be recorded??
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

func (verifier *DagVerifier) InitRedisClient() {
	redisURL := verifier.settings.Redis.Host + ":" + strconv.Itoa(verifier.settings.Redis.Port)
	redisDb := verifier.settings.Redis.Db
	verifier.redisClient = redisutils.InitRedisClient(redisURL, redisDb, verifier.settings.DagVerifierSettings.RedisPoolSize)
}

func (verifier *DagVerifier) InitSlackClient() {

	slackutils.InitSlackWorkFlowClient(verifier.settings.DagVerifierSettings.SlackNotifyURL)
}
