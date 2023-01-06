package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	cid "github.com/ipfs/go-cid"
	carv2 "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/blockstore"
	log "github.com/sirupsen/logrus"
	"github.com/web3-storage/go-w3s-client"
	"golang.org/x/time/rate"

	"github.com/powerloom/goutils/commonutils"
	"github.com/powerloom/goutils/redisutils"
	"github.com/powerloom/goutils/settings"
	"github.com/powerloom/goutils/slackutils"
)

type ProjectPruningVerificationStatus struct {
	LastSegmentEndHeight int   `json:"lastSegmentEndHeight"`
	SegmentWithErrors    []int `json:"segmentsWithErrors,omitempty"`
}

type PruningVerifier struct {
	redisClient                  *redis.Client
	w3sClient                    w3s.Client
	w3httpClient                 *http.Client
	web3StorageClientRateLimiter *rate.Limiter
	RunInterval                  int
	Projects                     []string
	ProjectVerificationStatus    map[string]*ProjectPruningVerificationStatus
	verificationReport           PruningVerificationReport
	settingsObj                  *settings.SettingsObj
}

type PruningVerificationReport struct {
	TotalProjects          int
	ArchivalFailedProjects int
	Host                   string
}

type PruningIssueReport struct {
	ProjectID    string          `json:"projectID"`
	SegmentError SegmentError    `json:"error"`
	ChainIssues  []DagChainIssue `json:"dagChainIssues"`
}
type SegmentError struct {
	Error     string `json:"errorData"`
	EndHeight int    `json:"endHeight"`
}

type ProjectDAGSegment struct {
	BeginHeight int    `json:"beginHeight"`
	EndHeight   int    `json:"endHeight"`
	EndDAGCID   string `json:"endDAGCID"`
	StorageType string `json:"storageType"`
}

type Web3StorageErrResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func (verifier *PruningVerifier) Init(settings *settings.SettingsObj) {
	redisURL := settings.Redis.Host + ":" + strconv.Itoa(settings.Redis.Port)
	redisDb := settings.Redis.Db
	poolSize := settings.DagVerifierSettings.RedisPoolSize
	//Verify every half-time of pruningInterval
	verifier.RunInterval = settings.PruningServiceSettings.RunIntervalMins / 2
	verifier.redisClient = redisutils.InitRedisClient(redisURL, redisDb, poolSize)
	verifier.settingsObj = settings
	verifier.w3httpClient = InitW3sHTTPClient(settings)
	verifier.InitW3sClient()
}

func InitW3sHTTPClient(settings *settings.SettingsObj) *http.Client {
	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        settings.Web3Storage.MaxIdleConns,
		MaxConnsPerHost:     settings.Web3Storage.MaxIdleConns,
		MaxIdleConnsPerHost: settings.Web3Storage.MaxIdleConns,
		IdleConnTimeout:     time.Duration(settings.Web3Storage.IdleConnTimeout),
		DisableCompression:  true,
	}

	w3sHttpClient := http.Client{
		Timeout:   time.Duration(settings.PruningServiceSettings.Web3Storage.TimeoutSecs) * time.Second,
		Transport: &t,
	}
	return &w3sHttpClient
}

func (verifier *PruningVerifier) InitW3sClient() {
	for {
		var err error
		verifier.w3sClient, err = w3s.NewClient(w3s.WithToken(verifier.settingsObj.Web3Storage.APIToken))
		if err != nil {
			log.Fatalf("Failed to initialize Web3.Storage client due to error %+v.Retrying after sometime.", err)
			time.Sleep(5 * time.Minute)
			continue
		}
		break
	}
	tps := rate.Limit(1) //1 TPS
	burst := 1
	if verifier.settingsObj.PruningServiceSettings.Web3Storage.RateLimit != nil {
		burst = verifier.settingsObj.PruningServiceSettings.Web3Storage.RateLimit.Burst
		if verifier.settingsObj.PruningServiceSettings.Web3Storage.RateLimit.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(verifier.settingsObj.PruningServiceSettings.Web3Storage.RateLimit.RequestsPerSec)
		}
	}
	log.Infof("Rate Limit configured for web3.storage at %v TPS with a burst of %d", tps, burst)
	verifier.web3StorageClientRateLimiter = rate.NewLimiter(tps, burst)
}

func (verifier *PruningVerifier) FetchProjectList() {
	key := redisutils.REDIS_KEY_STORED_PROJECTS
	log.Debugf("Fetching stored Projects from redis at key: %s", key)
	for i := 0; ; i++ {
		res := verifier.redisClient.SMembers(ctx, key)
		if res.Err() != nil {
			if res.Err() == redis.Nil {
				log.Warnf("Stored Projects key doesn't exist..retrying")
				time.Sleep(5 * time.Minute)
				continue
			}
			log.Errorf("Failed to fetch stored projects from redis due to err %+v. Retrying %d", res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		verifier.Projects = res.Val()
		log.Infof("Retrieved %d storedProjects %+v from redis", len(res.Val()), verifier.Projects)
		return
	}
}

func (verifier *PruningVerifier) Run() {
	//Fetch projects list
	verifier.FetchProjectList()
	pruningVerifierSleepInterval := time.Duration(verifier.RunInterval) * time.Minute
	for {
		verifier.ProjectVerificationStatus = make(map[string]*ProjectPruningVerificationStatus)
		if !verifier.FetchPruningVerificationStatusFromRedis() {
			continue
		}
		verifier.VerifyPruningAndArchival()

		log.Info("PruningVerifier: Sleeping for " + pruningVerifierSleepInterval.String())
		time.Sleep(pruningVerifierSleepInterval)
	}
}

func (verifier *PruningVerifier) VerifyPruningAndArchival() {
	verifier.verificationReport = PruningVerificationReport{}
	verifier.verificationReport.TotalProjects = len(verifier.Projects)
	verifier.verificationReport.Host, _ = os.Hostname()
	for index := range verifier.Projects {
		verifier.VerifyPruningStatus(verifier.Projects[index])
		verifier.UpdatePruningVerificationStatusToRedis(verifier.Projects[index])
	}
	if verifier.verificationReport.ArchivalFailedProjects > 0 {
		//Notify on slack of Failure.
		log.Errorf("Pruning Verification failed and report is %+v", verifier.verificationReport)
		report, _ := json.MarshalIndent(verifier.verificationReport, "", "\t")
		slackutils.NotifySlackWorkflow(string(report), "High", "PruningVerifier")
		//TODO: How to auto-clear this.
	}
}

func (verifier *PruningVerifier) VerifyPruningStatus(projectId string) {
	dagSegments := *(verifier.FetchProjectDagSegments(projectId))
	projectStatus := verifier.ProjectVerificationStatus[projectId]
	log.Debugf("Project %s, lastVerificationStatus %+v", projectId, projectStatus)
	sortedSegments := commonutils.SortKeysAsNumber(&dagSegments)
	for _, endHeightStr := range *sortedSegments {
		dagSegment := dagSegments[endHeightStr]
		var pruningReport PruningIssueReport
		pruningReport.ChainIssues = make([]DagChainIssue, 0)
		pruningReport.ProjectID = projectId
		endHeight, _ := strconv.Atoi(endHeightStr)
		if endHeight > projectStatus.LastSegmentEndHeight {
			log.Debugf("Project %s, verifying DAG Segment %+v", projectId, dagSegment)
			var segment ProjectDAGSegment
			json.Unmarshal([]byte(dagSegment), &segment)
			if segment.StorageType != "COLD" {
				log.Debugf("Segment not yet archived and still in HOT Storage, hence skipping archival verification")
				continue
			}
			archiveStatus, err, issues := verifier.VerifyArchivalStatus(projectId, &segment)
			//NOT doing any action for redis Zset as in next cycle Zset will get trimmed from startScore
			if !archiveStatus {
				error := SegmentError{
					Error:     fmt.Sprintf("endDAG CID %s is not available in web3.storage due to error %s", segment.EndDAGCID, err),
					EndHeight: segment.EndHeight,
				}
				pruningReport.SegmentError = error
				if issues != nil {
					pruningReport.ChainIssues = *issues
				}
				verifier.verificationReport.ArchivalFailedProjects++
				verifier.ProjectVerificationStatus[projectId].SegmentWithErrors = append(verifier.ProjectVerificationStatus[projectId].SegmentWithErrors, segment.EndHeight)
				verifier.AddPruningIssueReport(&pruningReport)
			}
			verifier.ProjectVerificationStatus[projectId].LastSegmentEndHeight = endHeight
		}
	}
}

func (verifier *PruningVerifier) AddPruningIssueReport(report *PruningIssueReport) {
	var member redis.Z
	member.Score = float64(report.SegmentError.EndHeight)
	reportStr, _ := json.Marshal(report)
	member.Member = reportStr
	key := fmt.Sprintf(redisutils.REDIS_KEY_PRUNING_ISSUES, report.ProjectID)
	i := 0
	for ; i < 3; i++ {
		res := verifier.redisClient.ZAddNX(ctx, key, &member)
		if res.Err() != nil {
			log.Warnf("Failed to update pruning issues key for project %s due to error %+v Retrying %d.", report.ProjectID, res.Err(), i)
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}
	if i >= 3 {
		log.Errorf("Failed to update pruning issues key for project %s after max retries.", report.ProjectID)
	}
	log.Debugf("Updated pruning issues key for project %s successfully.", report.ProjectID)
}

func (verifier *PruningVerifier) VerifyArchivalStatus(projectId string, segment *ProjectDAGSegment) (bool, string, *[]DagChainIssue) {
	//verifyStatus := false
	//curl -i -X 'HEAD' 'https://api.web3.storage/car/bafyreiglb6wf4ps4nthoaqml6sxndccxn5utrwv4n46l5uc263vny5nidm' -H 'accept: */*'
	//Response: - 200 OK with
	//expires: Sun, 24 Sep 2023 09:47:03 GMT
	//For now checking if root DAGCid is availale as part of archived data.

	rootCid, err := cid.Parse(segment.EndDAGCID)
	if err != nil {
		return false, err.Error(), nil
	}
	retryCount := 0
	for retryCount < 3 {

		err = verifier.web3StorageClientRateLimiter.Wait(context.Background())
		if err != nil {
			retryCount++
			log.Warnf("Web3Storage Rate Limiter wait timeout with error %+v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		s, err := verifier.w3sClient.Status(context.Background(), rootCid)
		if err != nil {
			log.Errorf("Project %s: Failed to fetch status of CID %s from web3.storage", projectId, segment.EndDAGCID)
			return false, err.Error(), nil
		}
		log.Debugf("Status for CID %s is: %+v", segment.EndDAGCID, s)
		if len(s.Pins) == 0 && len(s.Deals) == 0 {
			log.Errorf("Project %s: No pins or deals found for the DAG Segment with root CID %s in web3.storage", projectId, segment.EndDAGCID)
			return false, fmt.Sprintf("No pins or deals found for the DAG Segment with root CID %s in web3.storage", segment.EndDAGCID), nil
		}
		/* 		err = verifier.web3StorageClientRateLimiter.Wait(context.Background())
		   		if err != nil {
		   			retryCount++
		   			log.Warnf("Web3Storage Rate Limiter wait timeout with error %+v", err)
		   			time.Sleep(1 * time.Second)
		   			continue
		   		}

		   		if ok, err := verifier.FetchAndValidateCAR(segment); !ok {
		   			return false, err
		   		} */
		issues, errStr := verifier.VerifyDAGSegment(projectId, segment)
		if len(*issues) > 0 {
			log.Errorf("Project %s: %d DAGChain Issues found in the DAG Segment with root CID %s in web3.storage", projectId, len(*issues), segment.EndDAGCID)
			return false, fmt.Sprintf("%d DAGChain Issues found in the DAG Segment with root CID %s in web3.storage. PossibleCause: %s", len(*issues), segment.EndDAGCID, errStr), issues
		}
		break
	}
	if retryCount >= 3 {
		log.Warnf("Project %s: Failed to fetch status of CID %s from web3.storage after max retries", projectId, segment.EndDAGCID)
		return false, fmt.Sprintf("Failed to fetch status of CID %s from web3.storage after max retries", segment.EndDAGCID), nil
	}
	return true, "", nil
}

func (verifier *PruningVerifier) VerifyDAGSegment(projectId string, segment *ProjectDAGSegment) (*[]DagChainIssue, string) {
	dagHeight := segment.EndHeight
	chainIssues := make([]DagChainIssue, 0)
	prevBlockChainEndHeight := int64(0)
	for cid := segment.EndDAGCID; dagHeight >= segment.BeginHeight; dagHeight-- {
		log.Debugf("Project %s: DAG CID is %s at height %d", projectId, cid, dagHeight)
		dagBlock, err := ipfsClient.DagGet(cid)
		if err != nil {
			log.Errorf("Project %s: Could not fetch DAG CID %s at DagHeight %d from IPFS due to error %+v and hence can't verify further DAGChain", projectId, cid, dagHeight, err)
			chainIssue := DagChainIssue{IssueType: DAG_CHAIN_ISSUE_GAP_IN_CHAIN, DAGBlockHeight: int64(dagHeight)}
			chainIssues = append(chainIssues, chainIssue)
			return &chainIssues, fmt.Sprintf("Could not fetch DAG CID %s at DagHeight %d from IPFS and hence can't verify further DAGChain", cid, dagHeight)
		}
		payload, err := ipfsClient.GetPayload(dagBlock.Data.Cid.LinkData, 3)
		if err != nil {
			log.Warnf("Project %s: Could not fetch payload CID %s at DAG Height %d from IPFS due to error %+v", projectId, dagBlock.Data.Cid.LinkData, dagHeight, err)
			chainIssue := DagChainIssue{IssueType: DAG_CHAIN_ISSUE_GAP_IN_CHAIN, DAGBlockHeight: int64(dagHeight)}
			chainIssues = append(chainIssues, chainIssue)
			cid = dagBlock.PrevCid.LinkData
			continue
		}
		if prevBlockChainEndHeight != 0 && payload.ChainHeightRange.Begin == prevBlockChainEndHeight+1 {
			log.Warnf("Project %s: Could not fetch payload CID %s at DAG Height %d from IPFS due to error %+v", projectId, dagBlock.Data.Cid.LinkData, dagHeight, err)
			chainIssue := DagChainIssue{IssueType: DAG_CHAIN_ISSUE_GAP_IN_CHAIN, DAGBlockHeight: int64(dagHeight)}
			chainIssues = append(chainIssues, chainIssue)
			cid = dagBlock.PrevCid.LinkData
			continue
		}
		cid = dagBlock.PrevCid.LinkData
	}
	log.Debugf("Project %s: DAG Segment between height %d and %d with rootCID %s is valid", projectId, segment.BeginHeight, segment.EndHeight, segment.EndDAGCID)
	return &chainIssues, ""
}

func (verifier *PruningVerifier) DeleteCARFile(carFile string) {
	//Delete file from local storage.
	err := os.Remove(carFile)
	if err != nil {
		log.Errorf("Failed to delete file %s due to error %+v", carFile, err)
	}
	log.Debugf("Successfully deleted CAR File %s", carFile)
}

func (verifier *PruningVerifier) FetchAndValidateCAR(segment *ProjectDAGSegment) (bool, string) {
	ok, errStr, carFile := verifier.FetchCAR(segment)
	if !ok {
		return ok, errStr
	}
	defer verifier.DeleteCARFile(carFile)
	blockStore, err := blockstore.OpenReadOnly(carFile,
		blockstore.UseWholeCIDs(true),
		carv2.ZeroLengthSectionAsEOF(true),
	)
	if err != nil {
		log.Errorf("Failed to open CAR file %s to read as blockStore due to error %+v", carFile, err)
		return false, fmt.Sprintf("Failed to open CAR file %s to read as blockStore due to error %+v", carFile, err)
	}
	defer blockStore.Close()

	roots, err := blockStore.Roots()
	if err != nil {
		log.Errorf("Failed to fetch rootCIDs from CAR file %s to read as blockStore due to error %+v", carFile, err)
		return false, fmt.Sprintf("Failed to fetch rootCIDs from CAR file %s to read as blockStore due to error %+v", carFile, err)
	}
	log.Debugf("CAR File %s Contains %v root CID(s):", carFile, len(roots))
	isSegmentCIDPresentInCARFile := false
	for _, r := range roots {
		if r.String() == segment.EndDAGCID {
			isSegmentCIDPresentInCARFile = true
			break
		}
	}
	if !isSegmentCIDPresentInCARFile {
		log.Errorf("Failed to find segment rootCID %s in CAR file %s", segment.EndDAGCID, carFile)
		return false, fmt.Sprintf("Failed to find segment rootCID %s in CAR file %s", segment.EndDAGCID, carFile)
	}
	keysChan, _ := blockStore.AllKeysChan(context.Background())
	i := 0
	log.Tracef("Printing CIDS of CAR File %s", carFile)
	for k := range keysChan {
		log.Tracef("CID %s", k)
		i++
	}
	log.Debugf("Total CIDs in the car file %s is %d", carFile, i)

	if i != 2*(segment.EndHeight-segment.BeginHeight+1) {
		log.Errorf("Number of CIDS in CAR file %s is not matching that of segment. CAR File as %d CIDS, segment has %d",
			carFile, i, 2*(segment.EndHeight-segment.BeginHeight+1))
		return false, fmt.Sprintf("Number of CIDS in CAR file %s is not matching that of segment. CAR File as %d CIDS, segment has %d",
			carFile, i, 2*(segment.EndHeight-segment.BeginHeight+1))
	}

	return true, ""
}

func (verifier *PruningVerifier) FetchCAR(segment *ProjectDAGSegment) (bool, string, string) {
	urlString := verifier.settingsObj.Web3Storage.URL + "/car/" + segment.EndDAGCID
	log.Debugf("Fetching CAR File rom web3.storage url %s", urlString)
	errStr := ""
	for retryCount := 0; ; {
		if retryCount == *verifier.settingsObj.RetryCount {
			log.Errorf("web3.storage Fetch failed for segment %+v after max-retry of %d",
				segment, *verifier.settingsObj.RetryCount)
			return false, "Failed to fetch after max retry due to error " + errStr, ""
		}
		res, err := verifier.w3httpClient.Get(urlString)
		if err != nil {
			errStr = err.Error()
			log.Warnf("Failed to fetch DAG Segment %+v from web3.storage due to error %+v. Retrying", segment, err)
			time.Sleep(5 * time.Minute)
			continue
		}
		defer res.Body.Close()
		log.Debugf("Received response with status %d from web3.storage for DAG segment %+v", res.StatusCode, segment)
		if res.StatusCode == http.StatusOK {
			path := verifier.settingsObj.PruningServiceSettings.CARStoragePath
			fileName := fmt.Sprintf("%s%s_%d_%d.car", path, segment.EndDAGCID, segment.BeginHeight, segment.EndHeight)
			file, err := os.Create(fileName)
			if err != nil {
				log.Errorf("Unable to create file %s in specified path due to errro %+v", fileName, err)
				return false, "Failed to create file in path", ""
			}
			fileWriter := bufio.NewWriter(file)
			defer file.Close()

			//TODO: optimize for larger files.
			bytesWritten, err := io.Copy(fileWriter, res.Body)
			if err != nil {
				retryCount++
				log.Warnf("Failed to write to %s due to error %+v. Retrying %d", fileName, err, retryCount)
				time.Sleep(5 * time.Minute)
				continue
			}
			fileWriter.Flush()
			log.Debugf("Wrote %d bytes CAR to local file %s successfully", bytesWritten, fileName)
			return true, "", fileName
		} else {
			var resp Web3StorageErrResponse
			respBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Errorf("Failed to unmarshal error response from web3.storage due to error %+v", err)
			}
			json.Unmarshal(respBody, &resp)
			if res.StatusCode == http.StatusBadRequest {
				return false, "received 400 response from web3.storage due to error " + resp.Message, ""
			}
			retryCount++
			log.Warnf("Received Error response %+v from web3.storage for segment %+v with statusCode %d and status : %s ", resp,
				segment, res.StatusCode, res.Status)
			time.Sleep(5 * time.Minute)
			continue
		}
	}
}

func (verifier *PruningVerifier) FetchPruningVerificationStatusFromRedis() bool {
	for i := 0; i < 3; i++ {
		log.Infof("Fetching PruningVerification Status at key %s", redisutils.REDIS_KEY_PRUNING_VERIFICATION_STATUS)
		res := verifier.redisClient.HGetAll(ctx, redisutils.REDIS_KEY_PRUNING_VERIFICATION_STATUS)
		if res.Err() != nil {
			if res.Err() == redis.Nil {
				for index := range verifier.Projects {
					verifier.ProjectVerificationStatus[verifier.Projects[index]] = &ProjectPruningVerificationStatus{}
				}
				log.Debugf("No PruningVerification key found in redis and hence creating newly")
				return true
			}
			log.Warnf("Failed to fetch PruningVerification in redis..Retrying %d", i)
			time.Sleep(5 * time.Second)
			continue
		}
		if len(res.Val()) == 0 {
			for index := range verifier.Projects {
				verifier.ProjectVerificationStatus[verifier.Projects[index]] = &ProjectPruningVerificationStatus{}
			}
			log.Debugf("No PruningVerification key found in redis and hence creating newly")
			return true
		}

		for projectID, projectLastStatus := range res.Val() {
			projectStatus := ProjectPruningVerificationStatus{}
			err := json.Unmarshal([]byte(projectLastStatus), &projectStatus)
			if err != nil {
				log.Fatalf("Failed to unmarshal last projectStatus for project %s due to error %+v", projectID, err)
				return false
			}
			verifier.ProjectVerificationStatus[projectID] = &projectStatus
		}
		if len(res.Val()) < len(verifier.Projects) {
			for index := range verifier.Projects {
				if _, ok := verifier.ProjectVerificationStatus[verifier.Projects[index]]; !ok {
					log.Debugf("New project %s detected, Adding it to verification status", verifier.Projects[index])
					verifier.ProjectVerificationStatus[verifier.Projects[index]] = &ProjectPruningVerificationStatus{}
				}
			}
		}
		log.Debugf("Fetched PruningVerification status %+v successfully in redis", verifier.ProjectVerificationStatus)
		return true
	}
	log.Errorf("Failed to fetch PruningVerification status %+v in redis", verifier.ProjectVerificationStatus)
	return false
}

func (verifier *PruningVerifier) UpdatePruningVerificationStatusToRedis(projectID string) {

	for i := 0; i < 3; i++ {
		log.Infof("Updating PruningVerification Status at key %s", redisutils.REDIS_KEY_PRUNING_VERIFICATION_STATUS)
		valueStr, err := json.Marshal(verifier.ProjectVerificationStatus[projectID])
		if err != nil {
			log.Fatalf("Unable to marshal ProjectVerificationStatus due to error %+v", err)
			return
		}
		res := verifier.redisClient.HSet(ctx, redisutils.REDIS_KEY_PRUNING_VERIFICATION_STATUS, projectID, valueStr)
		if res.Err() != nil {
			log.Warnf("Failed to update PruningVerification for project %s in redis due to errro %+v..Retrying %d", projectID, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("Updated PruningVerification status %+v for project %s successfully in redis", verifier.ProjectVerificationStatus[projectID], projectID)
		return
	}
	log.Errorf("Failed to update PruningVerification status %+v for project %s in redis", verifier.ProjectVerificationStatus[projectID], projectID)

}

func (verifier *PruningVerifier) FetchProjectDagSegments(projectId string) *map[string]string {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_METADATA, projectId)
	for i := 0; i < 3; i++ {
		res := verifier.redisClient.HGetAll(ctx, key)
		if res.Err() != nil {
			if res.Err() == redis.Nil {
				return nil
			}
			log.Errorf("Could not fetch key %s due to error %+v. Retrying %d.",
				key, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("Successfully fetched project metaData from redis for projectId %s with value %s",
			projectId, res.Val())
		dagSements := make(map[string]string)
		dagSements = res.Val()
		return &dagSements
	}
	log.Errorf("Failed to fetch metaData for project %s from redis after max retries.", projectId)
	return nil
}
