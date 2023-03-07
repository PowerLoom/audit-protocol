package service

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alanshaw/go-carbites"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/commonutils"
	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/pruning-archival/constants"
	"audit-protocol/pruning-archival/models"
)

type PruningService struct {
	settingsObj *settings.SettingsObj
	caching     *caching
	ipfsClient  *ipfsutils.IpfsClient
}

func InitPruningService(settingsObj *settings.SettingsObj, redisClient *redis.Client, ipfsClient *ipfsutils.IpfsClient) *PruningService {
	return &PruningService{
		settingsObj: settingsObj,
		caching:     &caching{redisClient: redisClient},
		ipfsClient:  ipfsClient,
	}
}

func (p *PruningService) Run() {
	// not sure why this is getting called twice.
	// GetProjectsListFromRedis()
	for {
		// db/cache query in infinite for loop is not a good practice.
		// But as this is waiting for another cycle to complete, it is fine.
		// gets the list of projects from redis
		projectList := p.caching.GetProjectsListFromRedis()

		if !p.settingsObj.PruningServiceSettings.PerformArchival && !p.settingsObj.PruningServiceSettings.PerformIPFSUnPin &&
			!p.settingsObj.PruningServiceSettings.PruneRedisZsets {
			log.Infof("None of the pruning features enabled. Not doing anything in current cycle")
			time.Sleep(time.Duration(p.settingsObj.PruningServiceSettings.RunIntervalMins) * time.Minute)
			continue
		}

		cycleDetails := new(models.PruningCycleDetails)
		cycleDetails.CycleID = uuid.New().String()

		// get the last pruned height for each project from redis
		projectList = p.caching.GetLastPrunedStatusFromRedis(projectList, cycleDetails)

		log.WithField("CycleID", cycleDetails.CycleID).Infof("Running Pruning Cycle")

		p.verifyAndPruneDAGChains(projectList, cycleDetails)

		log.WithField("CycleID", cycleDetails.CycleID).Infof("Completed cycle")
		//TODO: Cleanup storage path if it has old files.
		time.Sleep(time.Duration(p.settingsObj.PruningServiceSettings.RunIntervalMins) * time.Minute)
	}
}

func (c *caching) GetProjectsListFromRedis() map[string]*models.ProjectPruneState {
	ctx := context.Background()
	key := redisutils.REDIS_KEY_STORED_PROJECTS

	log.Debugf("Fetching stored Projects from redis at key: %s", key)

	for i := 0; ; i++ {
		res := c.redisClient.SMembers(ctx, key)
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

		projectList := make(map[string]*models.ProjectPruneState, len(res.Val()))
		//projectList = make(map[string]*ProjectPruneState, 375)
		for index := range res.Val() {
			projectId := res.Val()[index]
			//if strings.Contains(projectId, "uniswap_V2PairsSummarySnapshot_UNISWAPV2") {
			projectPruneState := &models.ProjectPruneState{ProjectId: projectId}
			projectList[projectId] = projectPruneState
			//	break
			//}
		}

		log.Infof("Retrieved %d storedProjects %+v from redis", len(res.Val()), projectList)

		return projectList
	}
}

func (c *caching) GetLastPrunedStatusFromRedis(projectList map[string]*models.ProjectPruneState, cycleDetails *models.PruningCycleDetails) map[string]*models.ProjectPruneState {
	log.WithField("CycleID", cycleDetails.CycleID).Debug("Fetching Last Pruned Status at key:", redisutils.REDIS_KEY_PRUNING_STATUS)

	res := c.redisClient.HGetAll(context.Background(), redisutils.REDIS_KEY_PRUNING_STATUS)

	if len(res.Val()) == 0 {
		log.WithField("CycleID", cycleDetails.CycleID).Info("Failed to fetch Last Pruned Status  from redis for the projects.")
		//Key doesn't exist.
		log.WithField("CycleID", cycleDetails.CycleID).Info("Key doesn't exist..hence proceed from start of the block.")

		return projectList
	}
	err := res.Err()
	if err != nil {
		log.WithField("CycleID", cycleDetails.CycleID).Error("Ideally should not come here, which means there is some other redis error. To debug:", err)

		return projectList
	}

	// TODO: Need to handle dynamic addition of projects.
	for projectId, lastHeight := range res.Val() {
		if project, ok := projectList[projectId]; ok {
			project.LastPrunedHeight, err = strconv.Atoi(lastHeight)
			if err != nil {
				log.WithField("CycleID", cycleDetails.CycleID).Errorf("lastPrunedHeight corrupt for project %s. It will be set to 0", projectId)
				continue
			}
		} else {
			projectList[projectId] = &models.ProjectPruneState{ProjectId: projectId, LastPrunedHeight: 0}
		}
	}

	log.WithField("CycleID", cycleDetails.CycleID).Debugf("Fetched Last Pruned Status from redis %+v", projectList)

	return projectList
}

func (p *PruningService) verifyAndPruneDAGChains(projectList map[string]*models.ProjectPruneState, cycleDetails *models.PruningCycleDetails) {
	cycleDetails.CycleStartTime = time.Now().UnixMilli()
	concurrency := p.settingsObj.PruningServiceSettings.Concurrency
	totalProjects := len(projectList)
	projectIds := make([]string, totalProjects)
	index := 0

	//TODO: Optimize.
	// creating projectIds array to be used in go-routines
	for projectId := range projectList {
		projectIds[index] = projectId
		index++
	}

	cycleDetails.ProjectsCount = uint64(totalProjects)
	noOfProjectsPerRoutine := totalProjects / concurrency
	var wg sync.WaitGroup

	log.WithField("CycleID", cycleDetails.CycleID).Debugf("totalProjects %d, noOfProjectsPerRoutine %d concurrency %d \n", totalProjects, noOfProjectsPerRoutine, concurrency)

	// process n(numberOfProjectsPerRoutine) number of projects in each go-routine
	for startIndex := 0; startIndex < totalProjects; startIndex = startIndex + noOfProjectsPerRoutine + 1 {
		endIndex := startIndex + noOfProjectsPerRoutine
		if endIndex >= totalProjects {
			endIndex = totalProjects - 1
		}

		wg.Add(1)

		log.WithField("CycleID", cycleDetails.CycleID).Debugf("Go-Routine start %d, end %d \n", startIndex, endIndex)
		go func(start int, end int, limit int) {
			defer wg.Done()
			for k := start; k <= end; k++ {
				result := p.processProject(projectIds[k], projectList, cycleDetails)
				switch result {
				case 0:
					atomic.AddUint64(&cycleDetails.ProjectsProcessSuccessCount, 1)
				case 1:
					atomic.AddUint64(&cycleDetails.ProjectsNotProcessedCount, 1)
				default:
					atomic.AddUint64(&cycleDetails.ProjectsProcessFailedCount, 1)
				}
			}
		}(startIndex, endIndex, totalProjects)
	}
	wg.Wait()
	log.WithField("CycleID", cycleDetails.CycleID).Debugf("Finished all go-routines")
	cycleDetails.CycleEndTime = time.Now().UnixMilli()
	p.caching.UpdatePruningCycleDetailsInRedis(cycleDetails, p.settingsObj.PruningServiceSettings.RunIntervalMins)
}

func (p *PruningService) processProject(projectId string, projectList map[string]*models.ProjectPruneState, cycleDetails *models.PruningCycleDetails) int {
	log.WithField("CycleID", cycleDetails.CycleID).Debugf("Processing Project %s", projectId)

	projectReport := new(models.ProjectPruningReport)

	projectReport.ProjectID = projectId
	// Fetch Project metaData from redis
	projectMetaData := p.caching.FetchProjectMetaData(cycleDetails.CycleID, projectId)
	if projectMetaData == nil {
		log.WithField("CycleID", cycleDetails.CycleID).Debugf("No state metaData available for project %s, skipping this cycle.", projectId)
		return 1
	}

	projectPruneState := projectList[projectId]
	startScore := projectPruneState.LastPrunedHeight
	heightToPrune := p.FindPruningHeight(cycleDetails.CycleID, projectMetaData, projectPruneState)

	if heightToPrune <= startScore {
		log.WithField("CycleID", cycleDetails.CycleID).Debugf("Nothing to Prune for project %s", projectId)
		p.caching.UpdatePruningProjectReportInRedis(cycleDetails, projectReport, projectPruneState)
		return 1
	}

	log.WithField("CycleID", cycleDetails.CycleID).Debugf("Height to Prune is %d for project %s", heightToPrune, projectId)
	projectProcessed := false
	dagSegments := commonutils.SortKeysAsNumber(&projectMetaData.DagChains)
	//Sort DAGSegments by their height and then process.
	/* 	dagSegments := make([]string, 0, len(projectMetaData.DagChains))
	   	for k := range projectMetaData.DagChains {
	   		dagSegments = append(dagSegments, k)
	   		//log.Debugf("Key %s", k)
	   	}
	   	sort.Sort(asInt(dagSegments)) */

	//for dagSegmentEndHeightStr, dagChainSegment := range projectMetaData.DagChains {
	for _, dagSegmentEndHeightStr := range *dagSegments {
		dagChainSegment := projectMetaData.DagChains[dagSegmentEndHeightStr]
		dagSegmentEndHeight, err := strconv.Atoi(dagSegmentEndHeightStr)
		if err != nil {
			log.WithField("CycleID", cycleDetails.CycleID).Errorf("dagSegmentEndHeight %s is not an integer.", dagSegmentEndHeightStr)
			return -1
		}
		if dagSegmentEndHeight < heightToPrune {
			var dagSegment models.ProjectDAGSegment
			err = json.Unmarshal([]byte(dagChainSegment), &dagSegment)
			if err != nil {
				log.WithField("CycleID", cycleDetails.CycleID).Errorf("Unable to unmarshal dagChainSegment data due to error %+v", err)
				continue
			}
			log.WithField("CycleID", cycleDetails.CycleID).Debugf("Processing DAG Segment at height %d for project %s", dagSegment.BeginHeight, projectId)

			if dagSegment.StorageType == constants.DAG_CHAIN_STORAGE_TYPE_PENDING {
				log.WithField("CycleID", cycleDetails.CycleID).Infof("Performing Archival for project %s segment with endHeight %d", projectId, dagSegmentEndHeight)
				projectReport.DAGSegmentsProcessed++
				if p.settingsObj.PruningServiceSettings.PerformArchival {
					opStatus, err := p.ArchiveDAG(projectId, cycleDetails.CycleID, dagSegment.BeginHeight,
						dagSegment.EndHeight, dagSegment.EndDAGCID)
					if opStatus {
						dagSegment.StorageType = constants.DAG_CHAIN_STORAGE_TYPE_COLD
					} else {
						log.WithField("CycleID", cycleDetails.CycleID).Errorf("Failed to Archive DAG for project %s at height %d due to error.", projectId, dagSegmentEndHeight)
						projectReport.DAGSegmentsArchivalFailed++
						projectReport.ArchivalFailureCause = err.Error()
						p.caching.UpdatePruningProjectReportInRedis(cycleDetails, projectReport, projectPruneState)
						return -2
					}
				} else {
					log.Infof("Archival disabled, hence proceeding with pruning")
					dagSegment.StorageType = constants.DAG_CHAIN_STORAGE_TYPE_PRUNED
				}
				startScore = dagSegment.BeginHeight
				payloadCids := p.caching.GetPayloadCidsFromRedis(cycleDetails.CycleID, projectId, startScore, dagSegmentEndHeight)
				if payloadCids == nil {
					projectReport.UnPinFailed += dagSegment.EndHeight - dagSegment.BeginHeight
					projectReport.ArchivalFailureCause = "Failed to fetch payloadCids from Redis"
					p.caching.UpdatePruningProjectReportInRedis(cycleDetails, projectReport, projectPruneState)
					return -2
				}
				dagCids := p.caching.GetDAGCidsFromRedis(cycleDetails.CycleID, projectId, startScore, dagSegmentEndHeight)
				if dagCids == nil {
					projectReport.UnPinFailed += dagSegment.EndHeight - dagSegment.BeginHeight
					projectReport.ArchivalFailureCause = "Failed to fetch DAGCids from Redis"
					p.caching.UpdatePruningProjectReportInRedis(cycleDetails, projectReport, projectPruneState)
					return -2
				}
				log.WithField("CycleID", cycleDetails.CycleID).Infof("Unpinning DAG CIDS from IPFS for project %s segment with endheight %d", projectId, dagSegmentEndHeight)
				errCount := p.ipfsClient.UnPinCidsFromIPFS(projectId, dagCids)
				projectReport.UnPinFailed += errCount
				log.WithField("CycleID", cycleDetails.CycleID).Infof("Unpinning payload CIDS from IPFS for project %s segment with endheight %d", projectId, dagSegmentEndHeight)
				errCount = p.ipfsClient.UnPinCidsFromIPFS(projectId, payloadCids)
				projectReport.UnPinFailed += errCount
				p.caching.UpdateDagSegmentStatusToRedis(cycleDetails.CycleID, projectId, dagSegmentEndHeight, &dagSegment)

				projectPruneState.LastPrunedHeight = dagSegmentEndHeight
				if p.settingsObj.PruningServiceSettings.PruneRedisZsets {
					if p.settingsObj.PruningServiceSettings.BackUpRedisZSets {
						p.BackupZsetsToFile(projectId, cycleDetails.CycleID, startScore, dagSegmentEndHeight, payloadCids, dagCids)
					}
					log.WithField("CycleID", cycleDetails.CycleID).Infof("Pruning redis Zsets from IPFS for project %s segment with endheight %d", projectId, dagSegmentEndHeight)
					p.caching.PruneProjectInRedis(cycleDetails.CycleID, projectId, startScore, dagSegmentEndHeight)
				}

				p.caching.UpdatePrunedStatusToRedis(cycleDetails.CycleID, projectList, projectPruneState)
				errCount = p.caching.DeleteContentFromLocalCache(projectId, p.settingsObj.PayloadCachePath, dagCids, payloadCids)
				projectReport.LocalCacheDeletionsFailed += errCount
				projectReport.DAGSegmentsArchived++
				projectReport.CIDsUnPinned += len(*payloadCids) + len(*dagCids)
				projectProcessed = true
			}
		}
	}
	if projectProcessed {
		p.caching.UpdatePruningProjectReportInRedis(cycleDetails, projectReport, projectPruneState)
		return 0
	}
	return 1
}

func (p *PruningService) BackupZsetsToFile(projectId, cycleID string, startScore int, endScore int, payloadCids *map[int]string, dagCids *map[int]string) {
	path := p.settingsObj.PruningServiceSettings.CARStoragePath
	fileName := fmt.Sprintf("%s%s__%d_%d.json", path, projectId, startScore, endScore)
	file, err := os.Create(fileName)
	if err != nil {
		log.WithField("CycleID", cycleID).Errorf("Unable to create file %s in specified path due to errro %+v", fileName, err)
	}
	defer file.Close()

	type ZSets struct {
		PayloadCids *map[int]string `json:"payloadCids"`
		DagCids     *map[int]string `json:"dagCids"`
	}

	zSets := ZSets{PayloadCids: payloadCids, DagCids: dagCids}

	zSetsJson, err := json.Marshal(zSets)
	if err != nil {
		log.WithField("CycleID", cycleID).Fatalf("Failed to marshal payloadCids map to json due to error %+v", err)
	}

	bytesWritten, err := file.Write(zSetsJson)
	if err != nil {
		log.WithField("CycleID", cycleID).Errorf("Failed to write payloadCidsJson to file %s due to error %+v", fileName, err)
	} else {
		log.WithField("CycleID", cycleID).Debugf("Wrote %d bytes of payloadCids successfully to file %s.", bytesWritten, fileName)
		file.Sync()
	}
}

func (p *PruningService) FindPruningHeight(cycleID string, projectMetaData *models.ProjectMetaData, projectPruneState *models.ProjectPruneState) int {
	heightToPrune := projectPruneState.LastPrunedHeight
	// Fetch oldest height used by indexers.
	oldestIndexedHeight := p.caching.GetOldestIndexedProjectHeight(cycleID, projectPruneState, p.settingsObj)
	if oldestIndexedHeight != -1 {
		heightToPrune = oldestIndexedHeight - p.settingsObj.PruningServiceSettings.PruningHeightBehindOldestIndex //Adding a buffer just in case 7d index is just crossed and some heights before it are used in sliding window.
	}
	return heightToPrune
}

func (p *PruningService) ArchiveDAG(projectId, cycleID string, startScore int, endScore int, lastDagCid string) (bool, error) {
	var errToReturn error
	//Export DAG from IPFS
	fileName, opStatus := p.ExportDAGFromIPFS(projectId, cycleID, startScore, endScore, lastDagCid)
	if !opStatus {
		log.WithField("CycleID", cycleID).Errorf("Unable to export DAG for project %s at height %d. Will retry in next cycle", projectId, endScore)
		return opStatus, errors.New("failed to export CAR File from IPFS at height " + strconv.Itoa(endScore))
	}
	//Can be Optimized: Consider batch upload of files if sizes are too small.
	CID, opStatus := p.UploadFileToWeb3Storage(fileName, cycleID)
	if opStatus {
		log.WithField("CycleID", cycleID).Debugf("CID of CAR file %s uploaded to web3.storage is %s", fileName, CID)
	} else {
		log.WithField("CycleID", cycleID).Debugf("Failed to upload CAR file %s to web3.storage", fileName)
		errToReturn = errors.New("failed to upload CAR File to web3.storage" + fileName)
	}
	//Delete file from local storage.
	err := os.Remove(fileName)
	if err != nil {
		log.WithField("CycleID", cycleID).Errorf("Failed to delete file %s due to error %+v", fileName, err)
	}
	log.WithField("CycleID", cycleID).Debugf("Deleted file %s successfully from local storage", fileName)

	return opStatus, errToReturn
}

func (p *PruningService) ExportDAGFromIPFS(projectId, cycleID string, fromHeight int, toHeight int, dagCID string) (string, bool) {
	log.WithField("CycleID", cycleID).Debugf("Exporting DAG for project %s from height %d to height %d with last DAG CID %s", projectId, fromHeight, toHeight, dagCID)

	dagExportSuffix := "/api/v0/dag/export"

	// TODO: not really sure, where ipfsHTTPURL is coming from. Improve this
	//reqURL := "http://" + ipfsHTTPURL + dagExportSuffix + "?arg=" + dagCID + "&encoding=json&stream-channels=true&progress=false"
	reqURL := "http://" + dagExportSuffix + "?arg=" + dagCID + "&encoding=json&stream-channels=true&progress=false"
	log.WithField("CycleID", cycleID).Debugf("Sending request to URL %s", reqURL)
	for retryCount := 0; retryCount < 3; {
		if retryCount == *p.settingsObj.RetryCount {
			log.WithField("CycleID", cycleID).Errorf("CAR export failed for project %s at height %d after max-retry of %d",
				projectId, fromHeight, p.settingsObj.RetryCount)
			return "", false
		}
		req, err := http.NewRequest(http.MethodPost, reqURL, nil)
		if err != nil {
			log.WithField("CycleID", cycleID).Fatalf("Failed to create new HTTP Req with URL %s with error %+v",
				reqURL, err)
			return "", false
		}

		log.WithField("CycleID", cycleID).Debugf("Sending Req to IPFS URL %s for project %s at height %d ",
			reqURL, projectId, fromHeight)

		ipfsHTTPClient := httpclient.GetIPFSHTTPClient(p.settingsObj)
		res, err := ipfsHTTPClient.Do(req)
		if err != nil {
			retryCount++
			log.WithField("CycleID", cycleID).Warnf("Failed to send request %+v towards IPFS URL %s for project %s at height %d with error %+v.  Retrying %d",
				req, reqURL, projectId, fromHeight, err, retryCount)
			continue
		}
		defer res.Body.Close()

		if err != nil {
			retryCount++
			log.WithField("CycleID", cycleID).Warnf("Failed to read response body for project %s at height %d from IPFS with error %+v. Retrying %d",
				projectId, fromHeight, err, retryCount)
			continue
		}
		if res.StatusCode == http.StatusOK {
			log.WithField("CycleID", cycleID).Debugf("Received 200 OK from IPFS for project %s at height %d",
				projectId, fromHeight)
			path := p.settingsObj.PruningServiceSettings.CARStoragePath
			fileName := fmt.Sprintf("%s%s_%d_%d.car", path, projectId, fromHeight, toHeight)
			file, err := os.Create(fileName)
			if err != nil {
				log.WithField("CycleID", cycleID).Errorf("Unable to create file %s in specified path due to errro %+v", fileName, err)
				return "", false
			}
			defer file.Close()
			fileWriter := bufio.NewWriter(file)
			//TODO: optimize for larger files.
			bytesWritten, err := io.Copy(fileWriter, res.Body)
			if err != nil {
				retryCount++
				log.WithField("CycleID", cycleID).Warnf("Failed to write to %s due to error %+v. Retrying %d", fileName, err, retryCount)
				time.Sleep(time.Duration(p.settingsObj.RetryIntervalSecs) * time.Minute)
				continue
			}
			fileWriter.Flush()
			log.WithField("CycleID", cycleID).Debugf("Wrote %d bytes CAR to local file %s successfully", bytesWritten, fileName)
			return fileName, true
		} else {
			retryCount++
			log.WithField("CycleID", cycleID).Warnf("Received Error response from IPFS for project %s at height %d with statusCode %d and status : %s ",
				projectId, fromHeight, res.StatusCode, res.Status)
			time.Sleep(time.Duration(p.settingsObj.RetryIntervalSecs) * time.Minute)
			continue
		}
	}
	log.WithField("CycleID", cycleID).Debugf("Failed to export DAG from ipfs for project %s at height %d after max retries.", projectId, toHeight)
	return "", false
}

func (p *PruningService) UploadFileToWeb3Storage(fileName, cycleID string) (string, bool) {

	file, err := os.Open(fileName)
	if err != nil {
		log.WithField("CycleID", cycleID).Errorf("Unable to open file %s due to error %+v", fileName, err)
		return "", false
	}
	fileStat, err := file.Stat()
	if err != nil {
		log.WithField("CycleID", cycleID).Errorf("Unable to stat file %s due to error %+v", fileName, err)
		return "", false
	}
	targetSize := p.settingsObj.PruningServiceSettings.Web3Storage.UploadChunkSizeMB * 1024 * 1024 // 100MiB chunks

	fileReader := bufio.NewReader(file)

	if fileStat.Size() > int64(targetSize) {
		log.WithField("CycleID", cycleID).Infof("File size greater than targetSize %d bytes..doing chunking", targetSize)
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
				log.WithField("CycleID", cycleID).Fatalf("Failed to split car file %s due to error %+v", fileName, err)
				return "", false
			}
			lastCID, opStatus = p.UploadChunkToWeb3Storage(fileName, cycleID, car)
			if !opStatus {
				log.WithField("CycleID", cycleID).Errorf("Failed to upload chunk %d for file %s. aborting complete file.", i, fileName)
				return "", false
			}
			log.WithField("CycleID", cycleID).Debugf("Uploaded chunk %d of file %s to web3.storage successfully", i, fileName)
		}
		return lastCID, true
	} else {
		return p.UploadChunkToWeb3Storage(fileName, cycleID, fileReader)
	}
}

func (p *PruningService) UploadChunkToWeb3Storage(fileName, cycleID string, fileReader io.Reader) (string, bool) {
	reqURL := p.settingsObj.Web3Storage.URL + "/car"
	for retryCount := 0; retryCount < 3; {
		if retryCount == *p.settingsObj.RetryCount {
			log.WithField("CycleID", cycleID).Errorf("web3.storage upload failed for file %s after max-retry of %d",
				fileName, *p.settingsObj.RetryCount)
			return "", false
		}
		req, err := http.NewRequest(http.MethodPost, reqURL, fileReader)
		if err != nil {
			log.WithField("CycleID", cycleID).Fatalf("Failed to create new HTTP Req with URL %s for message %+v with error %+v",
				reqURL, err)
			return "", false
		}
		req.Header.Add("Authorization", "Bearer "+p.settingsObj.Web3Storage.APIToken)
		req.Header.Add("accept", "application/vnd.ipld.car")

		w3sHTTPClient, web3StorageClientRateLimiter := httpclient.GetW3sClient(p.settingsObj)

		err = web3StorageClientRateLimiter.Wait(context.Background())
		if err != nil {
			log.WithField("CycleID", cycleID).Warnf("Web3Storage Rate Limiter wait timeout with error %+v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.WithField("CycleID", cycleID).Debugf("Sending Req to web3.storage URL %s for file %s",
			reqURL, fileName)

		res, err := w3sHTTPClient.Do(req)
		if err != nil {
			retryCount++
			log.WithField("CycleID", cycleID).Warnf("Failed to send request %+v towards web3.storage URL %s for fileName %s with error %+v.  Retrying %d",
				req, reqURL)
			continue
		}
		defer res.Body.Close()
		var resp models.Web3StoragePostResponse
		respBody, err := io.ReadAll(res.Body)
		if err != nil {
			retryCount++
			log.WithField("CycleID", cycleID).Warnf("Failed to read response body for fileName %s from web3.storage with error %+v. Retrying %d",
				err, retryCount)
			continue
		}
		if res.StatusCode == http.StatusOK {
			err = json.Unmarshal(respBody, &resp)
			if err != nil {
				retryCount++
				log.WithField("CycleID", cycleID).Warnf("Failed to unmarshal response %+v for fileName %s towards web3.storage with error %+v. Retrying %d",
					respBody, fileName, err, retryCount)
				time.Sleep(time.Duration(p.settingsObj.RetryIntervalSecs) * time.Second)
				continue
			}
			log.WithField("CycleID", cycleID).Debugf("Received 200 OK from web3.storage for fileName %s with CID %s ",
				fileName, resp.CID)
			return resp.CID, true
		} else {
			if res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusForbidden ||
				res.StatusCode == http.StatusUnauthorized {
				log.WithField("CycleID", cycleID).Warnf("Failed to upload to web3.storage due to error %+v with statusCode %d", resp, res.StatusCode)
				return "", false
			}
			retryCount++
			var resp models.Web3StorageErrResponse
			err = json.Unmarshal(respBody, &resp)
			if err != nil {
				log.WithField("CycleID", cycleID).Errorf("Failed to unmarshal error response %+v for fileName %s towards web3.storage with error %+v. Retrying %d",
					respBody, fileName, err, retryCount)
			} else {
				log.WithField("CycleID", cycleID).Warnf("Received Error response %+v from web3.storage for fileName %s with statusCode %d and status : %s ",
					resp, fileName, res.StatusCode, res.Status)
			}
			time.Sleep(time.Duration(p.settingsObj.RetryIntervalSecs) * time.Second)
			continue
		}
	}
	log.WithField("CycleID", cycleID).Errorf("Failed to upload file %s to web3.storage after max retries", fileName)
	return "", false
}
