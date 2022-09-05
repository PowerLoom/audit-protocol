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

	"github.com/go-redis/redis/v8"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/powerloom/goutils/settings"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
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
	ProjectID string `json:"projectId"`
	DagChains []struct {
		BeginHeight     int    `json:"beginHeight"`
		EndHeight       int    `json:"endHeight"`
		EndDAGCID       string `json:"endDAGCID"`
		StorageType     string `json:"storageType"`
		ReadyForPruning bool   `json:"readyForPruning"`
	} `json:"dagChains"`
}

type ProjectState struct {
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

var projectList map[string]*ProjectState

const REDIS_KEY_STORED_PROJECTS string = "storedProjectIds"
const REDIS_KEY_PROJECT_PAYLOAD_CIDS string = "projectID:%s:payloadCids"
const REDIS_KEY_PROJECT_CIDS string = "projectID:%s:Cids"
const REDIS_KEY_PRUNING_STATUS string = "projects:pruningStatus"
const REDIS_KEY_PROJECT_METADATA string = "projectID:%s:projectMetaData"

func InitLogger() {
	log.SetOutput(ioutil.Discard) // Send all logs to nowhere by default

	log.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})
	log.AddHook(&writer.Hook{ // Send info and debug logs to stdout
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.TraceLevel,
			log.InfoLevel,
			log.DebugLevel,
		},
	})
	if len(os.Args) < 2 {
		fmt.Println("Pass loglevel as an argument if you don't want default(INFO) to be set.")
		fmt.Println("Values to be passed for logLevel: ERROR(2),INFO(4),DEBUG(5)")
		log.SetLevel(log.DebugLevel)
	} else {
		logLevel, err := strconv.ParseUint(os.Args[1], 10, 32)
		if err != nil || logLevel > 6 {
			log.SetLevel(log.DebugLevel) //TODO: Change default level to error
		} else {
			//TODO: Need to come up with approach to dynamically update logLevel.
			log.SetLevel(log.Level(logLevel))
		}
	}
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {

	InitLogger()
	settingsObj = settings.ParseSettings("../settings.json")
	InitIPFSClient()
	InitRedisClient()
	InitIPFSHTTPClient()
	InitW3sClient()
	Run()
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
		UpdatePrunedStatusToRedis()
		log.Infof("Completed cycle")
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
		//TODO: Read this from project metaData
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
			projectList[projectId] = &ProjectState{ProjectId: projectId, LastPrunedHeight: 0}
		}
	}
	log.Debugf("Fetched Last Pruned Status from redis %+v", projectList)
}

func UpdatePrunedStatusToRedis() {
	//No retry has been added, because in case of a failure, status will get updated in next run.
	lastPrunedStatus := make(map[string]string, len(projectList))
	//Prepare the status
	for projectId, projectState := range projectList {
		lastPrunedStatus[projectId] = strconv.Itoa(projectState.LastPrunedHeight)
	}

	for i := 0; i < 3; i++ {
		log.Info("Updating Last Pruned Status at key:", REDIS_KEY_PRUNING_STATUS)
		res := redisClient.HSet(ctx, REDIS_KEY_PRUNING_STATUS, lastPrunedStatus)
		if res.Err() != nil {
			log.Error("Failed to update Last Pruned Status in redis..Retrying %d", i)
			time.Sleep(5 * time.Second)
			continue
		}
		return
	}
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
		res := redisClient.Get(ctx, key)
		if res.Err() != nil {
			log.Errorf("Could not fetch key %s due to error %+v. Retrying %d.",
				key, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debugf("Successfully fetched project metaData from redis for projectId %s with value %s",
			projectId, res.Val())
		var projectMetaData ProjectMetaData
		err := json.Unmarshal([]byte(res.Val()), &projectMetaData)
		if err != nil {
			log.Fatalf("Failed to Unmarshal project metaData fetched from redis for project %s due to error %+v",
				projectId, err)
			return nil
		}
		return &projectMetaData
	}
	log.Errorf("Failed to fetch metaData for project %s from redis after max retries.", projectId)
	return nil
}

func FindPruningHeight(projectMetaData *ProjectMetaData, project *ProjectState) int {
	heightToPrune := project.LastPrunedHeight + 100
	for i := range projectMetaData.DagChains {
		if projectMetaData.DagChains[i].EndHeight < project.LastPrunedHeight {
			continue
		} else {
			if projectMetaData.DagChains[i].ReadyForPruning {
				heightToPrune = projectMetaData.DagChains[i].EndHeight
			}
		}
	}
	return heightToPrune
}

func ProcessProject(projectId string) {
	log.Debugf("Processing Project %s", projectId)
	//TODO: Fetch Project metaData from redis
	projectMetaData := FetchProjectMetaData(projectId)
	if projectMetaData == nil {
		return
	}
	project := projectList[projectId]
	startScore := project.LastPrunedHeight
	endScore := FindPruningHeight(projectMetaData, project)
	if endScore == startScore {
		log.Debugf("Nothing to Prune for project %s", projectId)
		return
	}
	payloadCids := GetPayloadCidsFromRedis(projectId, startScore, endScore)
	dagCids, lastDagCid := GetDAGCidsFromRedis(projectId, startScore, endScore)

	if settingsObj.PruningServiceSettings.PerformArchival {
		//Export DAG from IPFS
		fileName, err := ExportDAGFromIPFS(projectId, startScore, endScore, lastDagCid)
		if err {
			log.Errorf("Unable to export DAG for project %s at height %d. Will retry in next cycle", projectId, endScore)
			return
		}
		//Can be Optimized: Consider batch upload of files if sizes are too small.
		CID, opStatus := UploadToWeb3Storage(fileName)
		if opStatus {
			log.Debugf("CID of CAR file %s uploaded to web3.storage is %s", fileName, CID)
		} else {
			//TODO: Need to handle failure
		}
		//TODO: UpdateProject metadata about archival status of DAG Chains.
	}
	if settingsObj.PruningServiceSettings.PerformIPFSUnPin {
		UnPinFromIPFS(projectId, dagCids)
		UnPinFromIPFS(projectId, payloadCids)
	}
	project.LastPrunedHeight = endScore
	if settingsObj.PruningServiceSettings.PruneRedisZsets {
		//Backup ZSets before pruning
		BackupZsetsToFile(projectId, startScore, endScore, payloadCids, dagCids)
		//Prune redis ZSets
		PruneProjectInRedis(projectId, startScore, endScore)
	}
}

func BackupZsetsToFile(projectId string, startScore int, endScore int, payloadCids *map[int]string, dagCids *map[int]string) {
	path := settingsObj.PruningServiceSettings.CARStoragePath
	fileName := fmt.Sprintf("%s%s_%d_%d.json", path, projectId, startScore, endScore)
	file, err := os.Create(fileName)
	if err != nil {
		log.Errorf("Unable to create file %s in specified path due to errro %+v", fileName, err)
	}
	defer file.Close()
	payloadCidsJson, err := json.Marshal(payloadCids)
	if err != nil {
		log.Fatalf("Failed to marshal payloadCids map to json due to error %+v", err)
	}
	dagCidsJson, err := json.Marshal(dagCids)
	if err != nil {
		log.Fatalf("Failed to marshal dagCids map to json due to error %+v", err)
	}

	bytesWritten, err := file.Write(payloadCidsJson)
	if err != nil {
		log.Errorf("Failed to write payloadCidsJson to file %s due to error %+v", fileName, err)
	} else {
		log.Debugf("Wrote %d bytes of payloadCids successfully to file %s.", bytesWritten, fileName)
		file.Sync()
	}
	bytesWritten, err = file.Write(dagCidsJson)
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
			strconv.Itoa(startScore),
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
			log.Debugf("Received 200 OK from IPFS for project %s at height %d with Content-Length: %s",
				projectId, fromHeight, res.ContentLength)
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
				log.Errorf("Failed to write to %s due to error %+v", err)
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

func UploadToWeb3Storage(fileName string) (string, bool) {

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
	if fileStat.Size() > 100*1024*1024 {
		log.Infof("File size greater than 100MB..needs chunking. Chunking logic not yet in place!!")
		return "", false
		//TODO: Need to chunk CAR files more than 100MB as web3.storage has size limit right now.
		//Use code from carbites mentioned here https://web3.storage/docs/how-tos/work-with-car-files/
	}
	fileReader := bufio.NewReader(file)
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
	for _, cid := range *cids {
		i := 0
		for ; i < 3; i++ {
			err := ipfsClientRateLimiter.Wait(context.Background())
			if err != nil {
				log.Errorf("IPFSClient Rate Limiter wait timeout with error %+v", err)
				time.Sleep(1 * time.Second)
				continue
			}
			log.Debugf("Unpinning CID %s from IPFS for project %s", cid, projectId)
			err = ipfsClient.Unpin(cid)
			if err != nil {
				if err.Error() == "pin/rm: not pinned or pinned indirectly" {
					log.Errorf("CID %s for project %s could not be unpinned from IPFS as it was not pinned on the IPFS node.", cid, projectId)
					break
				}
				log.Errorf("Failed to unpin CID %s from ipfs for project %s due to error %+v. Retrying %d", cid, projectId, err, i)
				time.Sleep(5 * time.Second)
				continue
			}
			log.Debugf("Unpinned CID %s from IPFS successfully for project %s", cid, projectId)
			break
		}
		if i == 3 {
			log.Errorf("Failed to unpin CID %s from ipfs for project %s after max retries", cid, projectId)
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

func GetDAGCidsFromRedis(projectId string, startScore int, endScore int) (*map[int]string, string) {
	cids := make(map[int]string, endScore-startScore)
	var lastDagCid string
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
			if res.Val()[j].Score == float64(endScore) {
				lastDagCid = fmt.Sprintf("%v", res.Val()[j].Member)
			}
		}
		log.Debugf("Fetched %d DAG Cids from redis for project %s", len(cids), projectId)
		break
	}
	return &cids, lastDagCid
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
		//projectList = make(map[string]*ProjectState, len(res.Val()))
		//TODO: Remove hardcoding.
		projectList = make(map[string]*ProjectState, 375)
		for i := range res.Val() {
			projectId := res.Val()[i]
			//if strings.Contains(projectId, "UNISWAPV2") {
			projectState := ProjectState{projectId, 0}
			projectList[projectId] = &projectState
			//break
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
		Timeout:   time.Duration(settingsObj.Web3Storage.TimeoutSecs) * time.Second,
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
	tps := rate.Limit(3) //3 TPS
	burst := 3
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
	tps := rate.Limit(200) //50 TPS
	burst := 100
	if settingsObj.IPFSRateLimiter != nil {
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
