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
	"strings"
	"sync"
	"time"

	"github.com/alanshaw/go-carbites"
	"github.com/go-redis/redis/v8"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/powerloom/goutils/logger"
	"github.com/powerloom/goutils/settings"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

var ctx = context.Background()

var redisClient *redis.Client
var ipfsClient *shell.Shell
var settingsObj *settings.SettingsObj
var ipfsClientRateLimiter *rate.Limiter
var ipfsHTTPURL string

var ipfsHttpClient http.Client
var w3sHttpClient http.Client

var web3StorageClientRateLimiter *rate.Limiter

type ProjectMetaData struct {
	DagChains map[string]string
}

type ProjectDAGSegment struct {
	BeginHeight int    `json:"beginHeight"`
	EndHeight   int    `json:"endHeight"`
	EndDAGCID   string `json:"endDAGCID"`
	StorageType string `json:"storageType"`
}

type ProjectPruneState struct {
	ProjectId        string
	LastPrunedHeight int
}

type Web3StoragePostResponse struct {
	CID string `json:"cid"`
}

type Web3StorageErrResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

var projectList map[string]*ProjectPruneState

const REDIS_KEY_STORED_PROJECTS string = "storedProjectIds"
const REDIS_KEY_PROJECT_PAYLOAD_CIDS string = "projectID:%s:payloadCids"
const REDIS_KEY_PROJECT_CIDS string = "projectID:%s:Cids"
const REDIS_KEY_PROJECT_FINALIZED_HEIGHT string = "projectID:%s:blockHeight"

const REDIS_KEY_PRUNING_STATUS string = "projects:pruningStatus"
const REDIS_KEY_PROJECT_METADATA string = "projectID:%s:dagSegments"

const REDIS_KEY_PROJECT_TAIL_INDEX string = "projectID:%s:slidingCache:%s:tail"

const DAG_CHAIN_STORAGE_TYPE_COLD string = "COLD"

func main() {

	logger.InitLogger()
	settingsObj = settings.ParseSettings("../settings.json")
	SetDefaultPruneConfig()
	InitIPFSClient()
	InitRedisClient()
	InitIPFSHTTPClient()
	InitW3sClient()
	Run()
}

func SetDefaultPruneConfig() {
	if settingsObj.PruningServiceSettings == nil {
		defaultSettings := settings.PruningServiceSettings_{
			RunIntervalMins:    600,
			Concurrency:        5,
			CARStoragePath:     "/tmp/",
			PerformArchival:    true,
			PerformIPFSUnPin:   true,
			PruneRedisZsets:    true,
			BackUpRedisZSets:   false,
			OldestProjectIndex: "7d",
			IpfsTimeout:        300,
			IPFSRateLimiter:    &settings.RateLimiter_{Burst: 20, RequestsPerSec: 20},
		}
		settingsObj.PruningServiceSettings = &defaultSettings
	}
}

func Run() {
	GetProjectsListFromRedis()

	for {
		if !settingsObj.PruningServiceSettings.PerformArchival && !settingsObj.PruningServiceSettings.PerformIPFSUnPin &&
			!settingsObj.PruningServiceSettings.PruneRedisZsets {
			log.Infof("None of the pruning features enabled. Not doing anything in current cycle")
			time.Sleep(time.Duration(settingsObj.PruningServiceSettings.RunIntervalMins) * time.Minute)
			continue
		}
		GetLastPrunedStatusFromRedis()
		log.Infof("Running Pruning Cycle")
		VerifyAndPruneDAGChains()
		log.Infof("Completed cycle")
		//TODO: Cleanup storage path if it has old files.
		time.Sleep(time.Duration(settingsObj.PruningServiceSettings.RunIntervalMins) * time.Minute)
	}
}

func GetLastPrunedStatusFromRedis() {
	log.Debug("Fetching Last Pruned Status at key:", REDIS_KEY_PRUNING_STATUS)

	res := redisClient.HGetAll(ctx, REDIS_KEY_PRUNING_STATUS)

	if len(res.Val()) == 0 {
		log.Info("Failed to fetch Last Pruned Status  from redis for the projects.")
		//Key doesn't exist.
		log.Info("Key doesn't exist..hence proceed from start of the block.")
		return
	}
	err := res.Err()
	if err != nil {
		log.Error("Ideally should not come here, which means there is some other redis error. To debug:", err)
	}
	//TODO: Need to handle dynamic addition of projects.
	for projectId, lastHeight := range res.Val() {
		if project, ok := projectList[projectId]; ok {
			project.LastPrunedHeight, err = strconv.Atoi(lastHeight)
			if err != nil {
				log.Errorf("lastPrunedHeight corrupt for project %s. It will be set to 0", projectId)
				continue
			}
		} else {
			projectList[projectId] = &ProjectPruneState{ProjectId: projectId, LastPrunedHeight: 0}
		}
	}
	log.Debugf("Fetched Last Pruned Status from redis %+v", projectList)
}

func UpdatePrunedStatusToRedis(projectPruneState *ProjectPruneState) {
	lastPrunedStatus := make(map[string]string, len(projectList))
	lastPrunedStatus[projectPruneState.ProjectId] = strconv.Itoa(projectPruneState.LastPrunedHeight)

	for i := 0; i < 3; i++ {
		log.Info("Updating Last Pruned Status at key:", REDIS_KEY_PRUNING_STATUS)
		res := redisClient.HSet(ctx, REDIS_KEY_PRUNING_STATUS, lastPrunedStatus)
		if res.Err() != nil {
			log.Error("Failed to update Last Pruned Status in redis..Retrying %d", i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("Updated last Pruned status %+v successfully in redis", projectPruneState.ProjectId)
		return
	}
	log.Errorf("Failed to update last Pruned status %+v in redis", projectPruneState.ProjectId)
}

func VerifyAndPruneDAGChains() {
	concurrency := settingsObj.PruningServiceSettings.Concurrency
	totalProjects := len(projectList)
	projectIds := make([]string, totalProjects)
	index := 0
	//TODO: Optimize
	for projectId := range projectList {
		projectIds[index] = projectId
		index++
	}
	noOfProjectsPerRoutine := totalProjects / concurrency
	var wg sync.WaitGroup
	log.Debugf("totalProjects %d, noOfProjectsPerRouting %d concurrency %d \n", totalProjects, noOfProjectsPerRoutine, concurrency)
	for startIndex := 0; startIndex < totalProjects; startIndex = startIndex + noOfProjectsPerRoutine + 1 {
		endIndex := startIndex + noOfProjectsPerRoutine
		if endIndex >= totalProjects {
			endIndex = totalProjects - 1
		}
		wg.Add(1)
		log.Debugf("Go-Routine start %d, end %d \n", startIndex, endIndex)
		go func(start int, end int, limit int) {
			defer wg.Done()
			for k := start; k <= end; k++ {
				ProcessProject(projectIds[k])
			}
		}(startIndex, endIndex, totalProjects)
	}
	wg.Wait()
	log.Debugf("Finished all go-routines")
}

func FetchProjectMetaData(projectId string) *ProjectMetaData {
	key := fmt.Sprintf(REDIS_KEY_PROJECT_METADATA, projectId)
	for i := 0; i < 3; i++ {
		res := redisClient.HGetAll(ctx, key)
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
		var projectMetaData ProjectMetaData
		projectMetaData.DagChains = res.Val()
		return &projectMetaData
	}
	log.Errorf("Failed to fetch metaData for project %s from redis after max retries.", projectId)
	return nil
}

func GetOldestIndexedProjectHeight(projectPruneState *ProjectPruneState) int {
	key := fmt.Sprintf(REDIS_KEY_PROJECT_TAIL_INDEX, projectPruneState.ProjectId, settingsObj.PruningServiceSettings.OldestProjectIndex)
	lastIndexHeight := -1
	res := redisClient.Get(ctx, key)
	err := res.Err()
	if err != nil {
		if err == redis.Nil {
			log.Infof("Key %s does not exist", key)
			//For summary projects hard-code it to curBlockHeight-1000 as of now which gives safe values till 24hrs
			key = fmt.Sprintf(REDIS_KEY_PROJECT_FINALIZED_HEIGHT, projectPruneState.ProjectId)
			res = redisClient.Get(ctx, key)
			err := res.Err()
			if err != nil {
				if err == redis.Nil {
					log.Errorf("Key %s does not exist", key)
					return 0
				}
			}
			projectFinalizedHeight, err := strconv.Atoi(res.Val())
			if err != nil {
				log.Fatalf("Unable to convert retrieved projectFinalizedHeight for project %s to int due to error %+v ", projectPruneState.ProjectId, err)
				return -1
			}
			lastIndexHeight = projectFinalizedHeight - 1000
			return lastIndexHeight
		}
	}
	lastIndexHeight, err = strconv.Atoi(res.Val())
	if err != nil {
		log.Fatalf("Unable to convert retrieved lastIndexHeight for project %s to int due to error %+v ", projectPruneState.ProjectId, err)
		return -1
	}
	log.Debugf("Fetched oldest index height %d for project %s from redis ", lastIndexHeight, projectPruneState.ProjectId)
	return lastIndexHeight
}

func FindPruningHeight(projectMetaData *ProjectMetaData, projectPruneState *ProjectPruneState) int {
	heightToPrune := projectPruneState.LastPrunedHeight
	//Fetch oldest height used by indexers
	oldestIndexedHeight := GetOldestIndexedProjectHeight(projectPruneState)
	if oldestIndexedHeight != -1 {
		heightToPrune = oldestIndexedHeight - 100 //Adding a buffer just in case 7d index is just crossed and some heights before it are used in sliding window.
	}
	return heightToPrune
}

func ArchiveDAG(projectId string, startScore int, endScore int, lastDagCid string) bool {
	//Export DAG from IPFS
	fileName, opStatus := ExportDAGFromIPFS(projectId, startScore, endScore, lastDagCid)
	if opStatus {
		log.Errorf("Unable to export DAG for project %s at height %d. Will retry in next cycle", projectId, endScore)
		return false
	}
	//Can be Optimized: Consider batch upload of files if sizes are too small.
	CID, opStatus := UploadFileToWeb3Storage(fileName)
	if opStatus {
		log.Debugf("CID of CAR file %s uploaded to web3.storage is %s", fileName, CID)
	} else {
		log.Debugf("Failed to upload CAR file %s to web3.storage", fileName)
	}
	//Delete file from local storage.
	err := os.Remove(fileName)
	if err != nil {
		log.Errorf("Failed to delete file %s due to error %+v", fileName, err)
	}
	log.Debugf("Deleted file %s successfully from local storage", fileName)

	return opStatus
}

func UpdateDagSegmentStatusToRedis(projectID string, height int, dagSegment *ProjectDAGSegment) bool {
	key := fmt.Sprintf(REDIS_KEY_PROJECT_METADATA, projectID)
	for i := 0; i < 3; i++ {
		bytes, err := json.Marshal(dagSegment)
		if err != nil {
			log.Fatalf("Failed to marshal dag segment due toe error %+v", err)
			return false
		}
		res := redisClient.HSet(ctx, key, height, bytes)
		if res.Err() != nil {
			log.Errorf("Could not update key %s due to error %+v. Retrying %d.",
				key, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("Successfully updated archivedDAGSegments to redis for projectId %s with value %+v",
			projectID, *dagSegment)
		return true
	}
	log.Errorf("Failed to update DAGSegments %+v for project %s to redis after max retries.", *dagSegment, projectID)
	return false
}

func ProcessProject(projectId string) {
	log.Debugf("Processing Project %s", projectId)
	// Fetch Project metaData from redis
	projectMetaData := FetchProjectMetaData(projectId)
	if projectMetaData == nil {
		log.Debugf("No state metaData available for project %s, skipping this cycle.", projectId)
		return
	}
	projectPruneState := projectList[projectId]
	startScore := projectPruneState.LastPrunedHeight
	heightToPrune := FindPruningHeight(projectMetaData, projectPruneState)
	if heightToPrune <= startScore {
		log.Debugf("Nothing to Prune for project %s", projectId)
		return
	}
	log.Debugf("Height to Prune is %d for project %s", heightToPrune, projectId)

	for dagSegmentEndHeightStr, dagChainSegment := range projectMetaData.DagChains {
		dagSegmentEndHeight, err := strconv.Atoi(dagSegmentEndHeightStr)
		if err != nil {
			log.Errorf("dagSegmentEndHeight is not an integer.")
			return
		}
		if dagSegmentEndHeight < heightToPrune {
			var dagSegment ProjectDAGSegment
			err = json.Unmarshal([]byte(dagChainSegment), &dagSegment)
			if err != nil {
				log.Errorf("Unable to unmarshal dagChainSegment data due to error %+v", err)
				continue
			}
			log.Debugf("Processing DAG Segment at height %d for project %s", dagSegment.BeginHeight, projectId)

			if dagSegment.StorageType != DAG_CHAIN_STORAGE_TYPE_COLD {
				log.Infof("Performing Archival for project %s segment with endHeight %d", projectId, dagSegmentEndHeight)

				opStatus := ArchiveDAG(projectId, dagSegment.BeginHeight,
					dagSegment.EndHeight, dagSegment.EndDAGCID)
				if opStatus {
					dagSegment.StorageType = DAG_CHAIN_STORAGE_TYPE_COLD
				} else {
					log.Errorf("Failed to Archive DAG for project %s at height %d due to error.", projectId, dagSegmentEndHeight)
					return
				}
				UpdateDagSegmentStatusToRedis(projectId, dagSegmentEndHeight, &dagSegment)

				payloadCids := GetPayloadCidsFromRedis(projectId, startScore, dagSegmentEndHeight)
				dagCids := GetDAGCidsFromRedis(projectId, startScore, dagSegmentEndHeight)
				log.Infof("Unpinning DAG CIDS from IPFS for project %s segment with endheight %d", projectId, dagSegmentEndHeight)
				UnPinFromIPFS(projectId, dagCids)

				log.Infof("Unpinning payload CIDS from IPFS for project %s segment with endheight %d", projectId, dagSegmentEndHeight)
				UnPinFromIPFS(projectId, payloadCids)
				projectPruneState.LastPrunedHeight = dagSegmentEndHeight
				if settingsObj.PruningServiceSettings.PruneRedisZsets {
					if settingsObj.PruningServiceSettings.BackUpRedisZSets {
						BackupZsetsToFile(projectId, startScore, heightToPrune, payloadCids, dagCids)
					}
					log.Infof("Pruning redis Zsets from IPFS for project %s segment with endheight %d", projectId, dagSegmentEndHeight)
					PruneProjectInRedis(projectId, startScore, heightToPrune)
				}
				UpdatePrunedStatusToRedis(projectPruneState)
			}
		}
	}
}

func BackupZsetsToFile(projectId string, startScore int, endScore int, payloadCids *map[int]string, dagCids *map[int]string) {
	path := settingsObj.PruningServiceSettings.CARStoragePath
	fileName := fmt.Sprintf("%s%s__%d_%d.json", path, projectId, startScore, endScore)
	file, err := os.Create(fileName)
	if err != nil {
		log.Errorf("Unable to create file %s in specified path due to errro %+v", fileName, err)
	}
	defer file.Close()

	type ZSets struct {
		PayloadCids *map[int]string `json:"payloadCids"`
		DagCids     *map[int]string `json:"dagCids"`
	}

	zSets := ZSets{PayloadCids: payloadCids, DagCids: dagCids}

	zSetsJson, err := json.Marshal(zSets)
	if err != nil {
		log.Fatalf("Failed to marshal payloadCids map to json due to error %+v", err)
	}

	bytesWritten, err := file.Write(zSetsJson)
	if err != nil {
		log.Errorf("Failed to write payloadCidsJson to file %s due to error %+v", fileName, err)
	} else {
		log.Debugf("Wrote %d bytes of payloadCids successfully to file %s.", bytesWritten, fileName)
		file.Sync()
	}
}

func PruneProjectInRedis(projectId string, startScore int, endScore int) {
	key := fmt.Sprintf(REDIS_KEY_PROJECT_CIDS, projectId)
	PruneZSetInRedis(key, startScore, endScore)
	key = fmt.Sprintf(REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectId)
	PruneZSetInRedis(key, startScore, endScore)
}

func PruneZSetInRedis(key string, startScore int, endScore int) {
	for i := 0; i < 3; i++ {
		res := redisClient.ZRemRangeByScore(
			ctx, key,
			"-inf", //Always prune from start
			strconv.Itoa(endScore),
		)
		if res.Err() != nil {
			log.Errorf("Could not prune redis Zset %s between height %d to %d due to error %+v. Retrying %d.",
				key, startScore, endScore, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Debugf("Successfully pruned redis Zset %s of %d entries between height %d and %d",
			key, res.Val(), startScore, endScore)
		break
	}
}

func ExportDAGFromIPFS(projectId string, fromHeight int, toHeight int, dagCID string) (string, bool) {
	log.Debugf("Exporting DAG for project %s from height %d to height %d with last DAG CID %s", projectId, fromHeight, toHeight, dagCID)
	dagExportSuffix := "/api/v0/dag/export"
	reqURL := "http://" + ipfsHTTPURL + dagExportSuffix + "?arg=" + dagCID + "&encoding=json&stream-channels=true&progress=false"
	log.Debugf("Sending request to URL %s", reqURL)
	for retryCount := 0; ; {
		if retryCount == *settingsObj.RetryCount {
			log.Errorf("CAR export failed for project %s at height %d after max-retry of %d",
				projectId, fromHeight, settingsObj.RetryCount)
			return "", true
		}
		req, err := http.NewRequest(http.MethodPost, reqURL, nil)
		if err != nil {
			log.Fatalf("Failed to create new HTTP Req with URL %s with error %+v",
				reqURL, err)
			return "", true
		}

		log.Debugf("Sending Req to IPFS URL %s for project %s a height %d ",
			reqURL, projectId, fromHeight)
		res, err := ipfsHttpClient.Do(req)
		if err != nil {
			retryCount++
			log.Errorf("Failed to send request %+v towards IPFS URL %s for project %s at height %d with error %+v.  Retrying %d",
				req, reqURL, projectId, fromHeight, err, retryCount)
			continue
		}
		defer res.Body.Close()

		if err != nil {
			retryCount++
			log.Errorf("Failed to read response body for project %s at height %d from IPFS with error %+v. Retrying %d",
				projectId, fromHeight, err, retryCount)
			continue
		}
		if res.StatusCode == http.StatusOK {
			log.Debugf("Received 200 OK from IPFS for project %s at height %d",
				projectId, fromHeight)
			path := settingsObj.PruningServiceSettings.CARStoragePath
			fileName := fmt.Sprintf("%s%s_%d_%d.car", path, projectId, fromHeight, toHeight)
			file, err := os.Create(fileName)
			if err != nil {
				log.Errorf("Unable to create file %s in specified path due to errro %+v", fileName, err)
			}
			defer file.Close()
			fileWriter := bufio.NewWriter(file)
			//TODO: optimize for larger files.
			bytesWritten, err := io.Copy(fileWriter, res.Body)
			if err != nil {
				log.Errorf("Failed to write to %s due to error %+v", fileName, err)
			}
			fileWriter.Flush()
			log.Debugf("Wrote %d bytes CAR to local file %s successfully", bytesWritten, fileName)
			return fileName, false
		} else {
			retryCount++
			log.Errorf("Received Error response from IPFS for project %s at height %d with statusCode %d and status : %s ",
				projectId, fromHeight, res.StatusCode, res.Status)
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			continue
		}
	}
}

func UploadFileToWeb3Storage(fileName string) (string, bool) {

	file, err := os.Open(fileName)
	if err != nil {
		log.Errorf("Unable to open file %s due to error %+v", fileName, err)
		return "", false
	}
	fileStat, err := file.Stat()
	if err != nil {
		log.Errorf("Unable to stat file %s due to error %+v", fileName, err)
		return "", false
	}
	targetSize := 100 * 1024 * 1024 // 100MiB chunks

	fileReader := bufio.NewReader(file)

	if fileStat.Size() > int64(targetSize) {
		log.Infof("File size greater than targetSize %d bytes..doing chunking", targetSize)
		var lastCID string
		var opStatus bool
		// Need to chunk CAR files more than 100MB as web3.storage has size limit right now.
		//Use code from carbites mentioned here https://web3.storage/docs/how-tos/work-with-car-files/
		strategy := carbites.Treewalk
		spltr, _ := carbites.Split(file, targetSize, strategy)
		for i := 1; ; i++ {
			car, err := spltr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatalf("Failed to split car file %s due to error %+v", fileName, err)
				return "", false
			}
			lastCID, opStatus = UploadChunkToWeb3Storage(fileName, car)
			if !opStatus {
				log.Errorf("Failed to upload chunk %d for file %s. aborting complete file.", i, fileName)
				return "", false
			}
			log.Debugf("Uploaded chunk %d of file %s to web3.storage successfully", i, fileName)
		}
		return lastCID, true
	} else {
		return UploadChunkToWeb3Storage(fileName, fileReader)
	}
}

func UploadChunkToWeb3Storage(fileName string, fileReader io.Reader) (string, bool) {

	reqURL := settingsObj.Web3Storage.URL + "/car"
	for retryCount := 0; ; {
		if retryCount == *settingsObj.RetryCount {
			log.Errorf("web3.storage upload failed for file %s after max-retry of %d",
				fileName, *settingsObj.RetryCount)
			return "", false
		}
		req, err := http.NewRequest(http.MethodPost, reqURL, fileReader)
		if err != nil {
			log.Fatalf("Failed to create new HTTP Req with URL %s for message %+v with error %+v",
				reqURL, err)
			return "", false
		}
		req.Header.Add("Authorization", "Bearer "+settingsObj.Web3Storage.APIToken)
		req.Header.Add("accept", "application/vnd.ipld.car")

		err = web3StorageClientRateLimiter.Wait(context.Background())
		if err != nil {
			log.Errorf("Web3Storage Rate Limiter wait timeout with error %+v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.Debugf("Sending Req to web3.storage URL %s for file %s",
			reqURL, fileName)
		res, err := w3sHttpClient.Do(req)
		if err != nil {
			retryCount++
			log.Errorf("Failed to send request %+v towards web3.storage URL %s for fileName %s with error %+v.  Retrying %d",
				req, reqURL)
			continue
		}
		defer res.Body.Close()
		var resp Web3StoragePostResponse
		respBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			retryCount++
			log.Errorf("Failed to read response body for fileName %s from web3.storage with error %+v. Retrying %d",
				err, retryCount)
			continue
		}
		if res.StatusCode == http.StatusOK {
			err = json.Unmarshal(respBody, &resp)
			if err != nil {
				retryCount++
				log.Errorf("Failed to unmarshal response %+v for fileName %s towards web3.storage with error %+v. Retrying %d",
					respBody, fileName, err, retryCount)
				time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
				continue
			}
			log.Debugf("Received 200 OK from web3.storage for fileName %s with CID %s ",
				fileName, resp.CID)
			return resp.CID, true
		} else {
			if res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusForbidden ||
				res.StatusCode == http.StatusUnauthorized {
				log.Errorf("Failed to upload to web3.storage due to error %+v with statusCode %d", resp, res.StatusCode)
				return "", false
			}
			retryCount++
			var resp Web3StorageErrResponse
			err = json.Unmarshal(respBody, &resp)
			if err != nil {
				log.Errorf("Failed to unmarshal error response %+v for fileName %s towards web3.storage with error %+v. Retrying %d",
					respBody, fileName, err, retryCount)
			} else {
				log.Errorf("Received Error response %+v from web3.storage for fileName %s with statusCode %d and status : %s ",
					resp, fileName, res.StatusCode, res.Status)
			}
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			continue
		}
	}
}

func UnPinFromIPFS(projectId string, cids *map[int]string) {
	for height, cid := range *cids {
		i := 0
		for ; i < 3; i++ {
			err := ipfsClientRateLimiter.Wait(context.Background())
			if err != nil {
				log.Errorf("IPFSClient Rate Limiter wait timeout with error %+v", err)
				time.Sleep(1 * time.Second)
				continue
			}
			log.Debugf("Unpinning CID %s at height %d from IPFS for project %s", cid, height, projectId)
			err = ipfsClient.Unpin(cid)
			if err != nil {
				if err.Error() == "pin/rm: not pinned or pinned indirectly" || err.Error() == "pin/rm: pin is not part of the pinset" {
					log.Debugf("CID %s for project %s at height %d could not be unpinned from IPFS as it was not pinned on the IPFS node.", cid, projectId, height)
					break
				}
				log.Errorf("Failed to unpin CID %s from ipfs for project %s at height %d due to error %+v. Retrying %d", cid, projectId, height, err, i)
				time.Sleep(5 * time.Second)
				continue
			}
			log.Debugf("Unpinned CID %s at height %d from IPFS successfully for project %s", cid, height, projectId)
			break
		}
		if i == 3 {
			log.Errorf("Failed to unpin CID %s at height %d from ipfs for project %s after max retries", cid, height, projectId)
			//TODO: Add to some failed queue to Unpin this CID at a later stage.
			continue
		}
	}
}

func GetPayloadCidsFromRedis(projectId string, startScore int, endScore int) *map[int]string {
	cids := make(map[int]string, endScore-startScore)

	key := fmt.Sprintf(REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectId)
	for i := 0; ; i++ {
		res := redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
			Min: strconv.Itoa(startScore),
			Max: strconv.Itoa(endScore),
		})
		if res.Err() != nil {
			log.Errorf("Could not fetch payloadCids for project %s due to error %+v. Retrying %d.", projectId, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		for j := range res.Val() {
			cids[int(res.Val()[j].Score)] = fmt.Sprintf("%v", res.Val()[j].Member)
		}
		log.Debugf("Fetched %d payload Cids from redis for project %s", len(cids), projectId)
		break
	}
	return &cids
}

func GetDAGCidsFromRedis(projectId string, startScore int, endScore int) *map[int]string {
	cids := make(map[int]string, endScore-startScore)
	key := fmt.Sprintf(REDIS_KEY_PROJECT_CIDS, projectId)
	for i := 0; i < 3; i++ {
		res := redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
			Min: strconv.Itoa(startScore),
			Max: strconv.Itoa(endScore),
		})
		if res.Err() != nil {
			log.Errorf("Could not fetch payloadCids for project %s due to error %+v. Retrying %d.", projectId, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		for j := range res.Val() {
			cids[int(res.Val()[j].Score)] = fmt.Sprintf("%v", res.Val()[j].Member)
		}
		log.Debugf("Fetched %d DAG Cids from redis for project %s", len(cids), projectId)
		break
	}
	return &cids
}

func GetProjectsListFromRedis() {
	key := REDIS_KEY_STORED_PROJECTS
	log.Debugf("Fetching stored Projects from redis at key: %s", key)
	for i := 0; i < 3; i++ {
		res := redisClient.SMembers(ctx, key)
		if res.Err() != nil {
			if res.Err() == redis.Nil {
				log.Errorf("Stored Projects key doesn't exist..retrying")
				time.Sleep(5 * time.Minute)
				continue
			}
			log.Errorf("Failed to fetch stored projects from redis due to err %+v. Retrying %d", res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		projectList = make(map[string]*ProjectPruneState, len(res.Val()))
		//projectList = make(map[string]*ProjectPruneState, 375)
		for i := range res.Val() {
			projectId := res.Val()[i]
			//if strings.Contains(projectId, "uniswap_V2PairsSummarySnapshot_UNISWAPV2") {
			projectPruneState := ProjectPruneState{projectId, 0}
			projectList[projectId] = &projectPruneState
			//	break
			//}
		}
		log.Infof("Retrieved %d storedProjects %+v from redis", len(res.Val()), projectList)
		return
	}
}

func InitIPFSHTTPClient() {
	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        settingsObj.Web3Storage.MaxIdleConns,
		MaxConnsPerHost:     settingsObj.Web3Storage.MaxIdleConns,
		MaxIdleConnsPerHost: settingsObj.Web3Storage.MaxIdleConns,
		IdleConnTimeout:     time.Duration(settingsObj.Web3Storage.IdleConnTimeout),
		DisableCompression:  true,
	}

	ipfsHttpClient = http.Client{
		Timeout:   time.Duration(settingsObj.PruningServiceSettings.IpfsTimeout) * time.Second,
		Transport: &t,
	}
}

func InitW3sClient() {
	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        settingsObj.Web3Storage.MaxIdleConns,
		MaxConnsPerHost:     settingsObj.Web3Storage.MaxIdleConns,
		MaxIdleConnsPerHost: settingsObj.Web3Storage.MaxIdleConns,
		IdleConnTimeout:     time.Duration(settingsObj.Web3Storage.IdleConnTimeout),
		DisableCompression:  true,
	}

	w3sHttpClient = http.Client{
		Timeout:   time.Duration(settingsObj.Web3Storage.TimeoutSecs) * time.Second,
		Transport: &t,
	}

	//Default values
	tps := rate.Limit(1) //3 TPS
	burst := 1
	if settingsObj.Web3Storage.RateLimiter != nil {
		burst = settingsObj.Web3Storage.RateLimiter.Burst
		if settingsObj.Web3Storage.RateLimiter.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(settingsObj.Web3Storage.RateLimiter.RequestsPerSec)
		}
	}
	log.Infof("Rate Limit configured for web3.storage at %v TPS with a burst of %d", tps, burst)
	web3StorageClientRateLimiter = rate.NewLimiter(tps, burst)
}

func InitIPFSClient() {
	url := settingsObj.IpfsURL
	// Convert the URL from /ip4/<IPAddress>/tcp/<Port> to IP:Port format.
	connectUrl := strings.Split(url, "/")[2] + ":" + strings.Split(url, "/")[4]
	ipfsHTTPURL = connectUrl
	log.Infof("Initializing the IPFS client with IPFS Daemon URL %s.", connectUrl)
	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        settingsObj.PruningServiceSettings.Concurrency,
		MaxConnsPerHost:     settingsObj.PruningServiceSettings.Concurrency,
		MaxIdleConnsPerHost: settingsObj.PruningServiceSettings.Concurrency,
		IdleConnTimeout:     0,
		DisableCompression:  true,
	}

	ipfsHttpClient := http.Client{
		Timeout:   time.Duration(settingsObj.IpfsTimeout * 1000000000),
		Transport: &t,
	}
	log.Debugf("Setting IPFS HTTP client timeout as %f seconds", ipfsHttpClient.Timeout.Seconds())
	ipfsClient = shell.NewShellWithClient(connectUrl, &ipfsHttpClient)

	//Default values
	tps := rate.Limit(20) //50 TPS
	burst := 20
	if settingsObj.PruningServiceSettings.IPFSRateLimiter != nil {
		burst = settingsObj.PruningServiceSettings.IPFSRateLimiter.Burst
		if settingsObj.PruningServiceSettings.IPFSRateLimiter.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(settingsObj.PruningServiceSettings.IPFSRateLimiter.RequestsPerSec)
		}
	}
	log.Infof("Rate Limit configured for IPFS Client at %v TPS with a burst of %d", tps, burst)
	ipfsClientRateLimiter = rate.NewLimiter(tps, burst)

}

func InitRedisClient() {
	redisURL := fmt.Sprintf("%s:%d", settingsObj.Redis.Host, settingsObj.Redis.Port)
	redisDb := settingsObj.Redis.Db
	log.Infof("Connecting to redis DB %d at %s", redisDb, redisURL)
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: settingsObj.Redis.Password,
		DB:       redisDb,
		PoolSize: settingsObj.DagVerifierSettings.RedisPoolSize,
	})
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Errorf("Unable to connect to redis at %s with error %+v", redisURL, err)
	}
	log.Info("Connected successfully to Redis and received ", pong, " back")

}
