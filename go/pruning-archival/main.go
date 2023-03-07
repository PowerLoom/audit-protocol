package main

import (
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/logger"
	"audit-protocol/goutils/redisutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/slackutils"
	ps "audit-protocol/pruning-archival/service"
)

// using global context can be awful.
// context passed to functions is can be cancelled due to error or timeout.
// using same context again may be problematic and cause all sorts of issues.
//var ctx = context.Background()

// avoid using global variables
/*var redisClient *redis.Client
var ipfsClient ipfsutils.IpfsClient
var settingsObj *settings.SettingsObj
var ipfsHTTPURL string

var ipfsHttpClient http.Client
var w3sHttpClient http.Client

var web3StorageClientRateLimiter *rate.Limiter
var cycleDetails PruningCycleDetails
var projectList map[string]*models.ProjectPruneState*/

// move this to constants/constants.go file
/*const DAG_CHAIN_STORAGE_TYPE_COLD string = "COLD"
const DAG_CHAIN_STORAGE_TYPE_PRUNED string = "PRUNED"
const DAG_CHAIN_STORAGE_TYPE_PENDING string = "pending"*/

func main() {
	logger.InitLogger()

	// load settings
	settingsObj := settings.ParseSettings()
	settingsObj.PruningServiceSettings = settingsObj.GetDefaultPruneConfig()

	ipfsUrl := settingsObj.IpfsConfig.ReaderURL
	if ipfsUrl == "" {
		ipfsUrl = settingsObj.IpfsConfig.URL
	}

	ipfsClient := ipfsutils.InitClient(
		ipfsUrl,
		settingsObj.PruningServiceSettings.Concurrency,
		settingsObj.PruningServiceSettings.IPFSRateLimiter,
		settingsObj.PruningServiceSettings.IpfsTimeout,
	)

	redisClient := redisutils.InitRedisClient(
		settingsObj.Redis.Host,
		settingsObj.Redis.Port,
		settingsObj.Redis.Db,
		settingsObj.DagVerifierSettings.RedisPoolSize,
		settingsObj.Redis.Password,
	)

	pruningService := ps.InitPruningService(settingsObj, redisClient, ipfsClient)

	slackutils.InitSlackWorkFlowClient(settingsObj.DagVerifierSettings.SlackNotifyURL)

	pruningService.Run()
}

// improvements needed

/*
1. find a way to passaround cycleID
*/

// moved this function to goutils/settings/settings.go
/*func SetDefaultPruneConfig() {
	if settingsObj.PruningServiceSettings == nil {
		defaultSettings := settings.PruningServiceSettings{
			RunIntervalMins:                      600,
			Concurrency:                          5,
			CARStoragePath:                       "/tmp/",
			PerformArchival:                      true,
			PerformIPFSUnPin:                     true,
			PruneRedisZsets:                      true,
			BackUpRedisZSets:                     false,
			OldestProjectIndex:                   "7d",
			PruningHeightBehindOldestIndex:       100,
			IpfsTimeout:                          300,
			IPFSRateLimiter:                      &settings.RateLimiter{Burst: 20, RequestsPerSec: 20},
			SummaryProjectsPruneHeightBehindHead: 1000,
		}
		defaultSettings.Web3Storage.TimeoutSecs = 600
		defaultSettings.Web3Storage.RateLimit = &settings.RateLimiter{Burst: 1, RequestsPerSec: 1}
		defaultSettings.Web3Storage.UploadChunkSizeMB = 50
		settingsObj.PruningServiceSettings = &defaultSettings
	}
}*/

/*func Run(settingsObj *settings.SettingsObj) {
	// GetProjectsListFromRedis()
	for {
		// db/cache query in infinite for loop is not a good practice.
		// But as this is waiting for another cycle to complete, it is fine.
		GetProjectsListFromRedis()
		if !settingsObj.PruningServiceSettings.PerformArchival && !settingsObj.PruningServiceSettings.PerformIPFSUnPin &&
			!settingsObj.PruningServiceSettings.PruneRedisZsets {
			log.Infof("None of the pruning features enabled. Not doing anything in current cycle")
			time.Sleep(time.Duration(settingsObj.PruningServiceSettings.RunIntervalMins) * time.Minute)
			continue
		}

		GetLastPrunedStatusFromRedis()

		cycleDetails := &PruningCycleDetails{}
		cycleDetails.CycleID = uuid.New().String()

		log.WithField("CycleID", cycleDetails.CycleID).Infof("Running Pruning Cycle")
		VerifyAndPruneDAGChains()
		log.WithField("CycleID", cycleDetails.CycleID).Infof("Completed cycle")
		//TODO: Cleanup storage path if it has old files.
		time.Sleep(time.Duration(settingsObj.PruningServiceSettings.RunIntervalMins) * time.Minute)
	}
}*/

// moved this function to service/service.go
/*func GetLastPrunedStatusFromRedis() {
	log.WithField("CycleID", cycleDetails.CycleID).Debug("Fetching Last Pruned Status at key:", redisutils.REDIS_KEY_PRUNING_STATUS)

	res := redisClient.HGetAll(ctx, redisutils.REDIS_KEY_PRUNING_STATUS)

	if len(res.Val()) == 0 {
		log.WithField("CycleID", cycleDetails.CycleID).Info("Failed to fetch Last Pruned Status  from redis for the projects.")
		//Key doesn't exist.
		log.WithField("CycleID", cycleDetails.CycleID).Info("Key doesn't exist..hence proceed from start of the block.")
		return
	}
	err := res.Err()
	if err != nil {
		log.WithField("CycleID", cycleDetails.CycleID).Error("Ideally should not come here, which means there is some other redis error. To debug:", err)
	}
	//TODO: Need to handle dynamic addition of projects.
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
}*/

/*func UpdatePrunedStatusToRedis(projectPruneState *models.ProjectPruneState) {
	lastPrunedStatus := make(map[string]string, len(projectList))
	lastPrunedStatus[projectPruneState.ProjectId] = strconv.Itoa(projectPruneState.LastPrunedHeight)

	for i := 0; i < 3; i++ {
		log.WithField("CycleID", cycleDetails.CycleID).Infof("Updating Last Pruned Status at key %s", redisutils.REDIS_KEY_PRUNING_STATUS)
		res := redisClient.HSet(ctx, redisutils.REDIS_KEY_PRUNING_STATUS, lastPrunedStatus)
		if res.Err() != nil {
			log.WithField("CycleID", cycleDetails.CycleID).Warnf("Failed to update Last Pruned Status in redis..Retrying %d", i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.WithField("CycleID", cycleDetails.CycleID).Debugf("Updated last Pruned status %+v successfully in redis", projectPruneState.ProjectId)
		return
	}
	log.WithField("CycleID", cycleDetails.CycleID).Errorf("Failed to update last Pruned status %+v in redis", projectPruneState.ProjectId)
}*/

// moved this function to service/service.go
/*func VerifyAndPruneDAGChains() {
	cycleDetails.CycleStartTime = time.Now().UnixMilli()

	concurrency := settingsObj.PruningServiceSettings.Concurrency
	totalProjects := len(projectList)
	projectIds := make([]string, totalProjects)
	index := 0
	//TODO: Optimize
	for projectId := range projectList {
		projectIds[index] = projectId
		index++
	}
	cycleDetails.ProjectsCount = uint64(totalProjects)
	noOfProjectsPerRoutine := totalProjects / concurrency
	var wg sync.WaitGroup
	log.WithField("CycleID", cycleDetails.CycleID).Debugf("totalProjects %d, noOfProjectsPerRouting %d concurrency %d \n", totalProjects, noOfProjectsPerRoutine, concurrency)
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
				result := processProject(projectIds[k])
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
	UpdatePruningCycleDetailsInRedis()
}*/

// moved this function to service/service.go
/*func FetchProjectMetaData(projectId string) *ProjectMetaData {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_METADATA, projectId)
	for i := 0; i < 3; i++ {
		res := redisClient.HGetAll(ctx, key)
		if res.Err() != nil {
			if res.Err() == redis.Nil {
				return nil
			}
			log.WithField("CycleID", cycleDetails.CycleID).Warnf("Could not fetch key %s due to error %+v. Retrying %d.",
				key, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.WithField("CycleID", cycleDetails.CycleID).Debugf("Successfully fetched project metaData from redis for projectId %s with value %s",
			projectId, res.Val())
		var projectMetaData ProjectMetaData
		projectMetaData.DagChains = res.Val()
		return &projectMetaData
	}
	log.WithField("CycleID", cycleDetails.CycleID).Errorf("Failed to fetch metaData for project %s from redis after max retries.", projectId)
	return nil
}*/

/*func GetOldestIndexedProjectHeight(projectPruneState *ProjectPruneState) int {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_TAIL_INDEX, projectPruneState.ProjectId, settingsObj.PruningServiceSettings.OldestProjectIndex)
	lastIndexHeight := -1
	res := redisClient.Get(ctx, key)
	err := res.Err()
	if err != nil {
		if err == redis.Nil {
			log.WithField("CycleID", cycleDetails.CycleID).Infof("Key %s does not exist", key)
			//For summary projects hard-code it to curBlockHeight-1000 as of now which gives safe values till 24hrs
			key = fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_FINALIZED_HEIGHT, projectPruneState.ProjectId)
			res = redisClient.Get(ctx, key)
			err := res.Err()
			if err != nil {
				if err == redis.Nil {
					log.WithField("CycleID", cycleDetails.CycleID).Errorf("Key %s does not exist", key)
					return 0
				}
			}
			projectFinalizedHeight, err := strconv.Atoi(res.Val())
			if err != nil {
				log.WithField("CycleID", cycleDetails.CycleID).Fatalf("Unable to convert retrieved projectFinalizedHeight for project %s to int due to error %+v ", projectPruneState.ProjectId, err)
				return -1
			}
			lastIndexHeight = projectFinalizedHeight - settingsObj.PruningServiceSettings.SummaryProjectsPruneHeightBehindHead
			return lastIndexHeight
		}
	}
	lastIndexHeight, err = strconv.Atoi(res.Val())
	if err != nil {
		log.WithField("CycleID", cycleDetails.CycleID).Fatalf("Unable to convert retrieved lastIndexHeight for project %s to int due to error %+v ", projectPruneState.ProjectId, err)
		return -1
	}
	log.WithField("CycleID", cycleDetails.CycleID).Debugf("Fetched oldest index height %d for project %s from redis ", lastIndexHeight, projectPruneState.ProjectId)
	return lastIndexHeight
}*/

// moved this function to service/service.go
/*func FindPruningHeight(projectMetaData *ProjectMetaData, projectPruneState *ProjectPruneState) int {
	heightToPrune := projectPruneState.LastPrunedHeight
	//Fetch oldest height used by indexers
	oldestIndexedHeight := GetOldestIndexedProjectHeight(projectPruneState)
	if oldestIndexedHeight != -1 {
		heightToPrune = oldestIndexedHeight - settingsObj.PruningServiceSettings.PruningHeightBehindOldestIndex //Adding a buffer just in case 7d index is just crossed and some heights before it are used in sliding window.
	}
	return heightToPrune
}*/

/*func ArchiveDAG(projectId string, startScore int, endScore int, lastDagCid string) (bool, error) {
	var errToReturn error
	//Export DAG from IPFS
	fileName, opStatus := ExportDAGFromIPFS(projectId, startScore, endScore, lastDagCid)
	if !opStatus {
		log.WithField("CycleID", cycleDetails.CycleID).Errorf("Unable to export DAG for project %s at height %d. Will retry in next cycle", projectId, endScore)
		return opStatus, errors.New("failed to export CAR File from IPFS at height " + strconv.Itoa(endScore))
	}
	//Can be Optimized: Consider batch upload of files if sizes are too small.
	CID, opStatus := UploadFileToWeb3Storage(fileName)
	if opStatus {
		log.WithField("CycleID", cycleDetails.CycleID).Debugf("CID of CAR file %s uploaded to web3.storage is %s", fileName, CID)
	} else {
		log.WithField("CycleID", cycleDetails.CycleID).Debugf("Failed to upload CAR file %s to web3.storage", fileName)
		errToReturn = errors.New("failed to upload CAR File to web3.storage" + fileName)
	}
	//Delete file from local storage.
	err := os.Remove(fileName)
	if err != nil {
		log.WithField("CycleID", cycleDetails.CycleID).Errorf("Failed to delete file %s due to error %+v", fileName, err)
	}
	log.WithField("CycleID", cycleDetails.CycleID).Debugf("Deleted file %s successfully from local storage", fileName)

	return opStatus, errToReturn
}*/

/*func UpdateDagSegmentStatusToRedis(projectID string, height int, dagSegment *ProjectDAGSegment) bool {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_METADATA, projectID)
	for i := 0; i < 3; i++ {
		bytes, err := json.Marshal(dagSegment)
		if err != nil {
			log.WithField("CycleID", cycleDetails.CycleID).Fatalf("Failed to marshal dag segment due toe error %+v", err)
			return false
		}
		res := redisClient.HSet(ctx, key, height, bytes)
		if res.Err() != nil {
			log.WithField("CycleID", cycleDetails.CycleID).Warnf("Could not update key %s due to error %+v. Retrying %d.",
				key, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		log.WithField("CycleID", cycleDetails.CycleID).Debugf("Successfully updated archivedDAGSegments to redis for projectId %s with value %+v",
			projectID, *dagSegment)
		return true
	}
	log.WithField("CycleID", cycleDetails.CycleID).Errorf("Failed to update DAGSegments %+v for project %s to redis after max retries.", *dagSegment, projectID)
	return false
}*/

// moved to service/service.go.
//func processProject(projectId string) int {
//	log.WithField("CycleID", cycleDetails.CycleID).Debugf("Processing Project %s", projectId)
//	var projectReport ProjectPruningReport
//	projectReport.ProjectID = projectId
//	// Fetch Project metaData from redis
//	projectMetaData := FetchProjectMetaData(projectId)
//	if projectMetaData == nil {
//		log.WithField("CycleID", cycleDetails.CycleID).Debugf("No state metaData available for project %s, skipping this cycle.", projectId)
//		return 1
//	}
//	projectPruneState := projectList[projectId]
//	startScore := projectPruneState.LastPrunedHeight
//	heightToPrune := FindPruningHeight(projectMetaData, projectPruneState)
//	if heightToPrune <= startScore {
//		log.WithField("CycleID", cycleDetails.CycleID).Debugf("Nothing to Prune for project %s", projectId)
//		UpdatePruningProjectReportInRedis(&projectReport, projectPruneState)
//		return 1
//	}
//	log.WithField("CycleID", cycleDetails.CycleID).Debugf("Height to Prune is %d for project %s", heightToPrune, projectId)
//	projectProcessed := false
//	dagSegments := commonutils.SortKeysAsNumber(&projectMetaData.DagChains)
//	//Sort DAGSegments by their height and then process.
//	/* 	dagSegments := make([]string, 0, len(projectMetaData.DagChains))
//	   	for k := range projectMetaData.DagChains {
//	   		dagSegments = append(dagSegments, k)
//	   		//log.Debugf("Key %s", k)
//	   	}
//	   	sort.Sort(asInt(dagSegments)) */
//
//	//for dagSegmentEndHeightStr, dagChainSegment := range projectMetaData.DagChains {
//	for _, dagSegmentEndHeightStr := range *dagSegments {
//		dagChainSegment := projectMetaData.DagChains[dagSegmentEndHeightStr]
//		dagSegmentEndHeight, err := strconv.Atoi(dagSegmentEndHeightStr)
//		if err != nil {
//			log.WithField("CycleID", cycleDetails.CycleID).Errorf("dagSegmentEndHeight %s is not an integer.", dagSegmentEndHeightStr)
//			return -1
//		}
//		if dagSegmentEndHeight < heightToPrune {
//			var dagSegment ProjectDAGSegment
//			err = json.Unmarshal([]byte(dagChainSegment), &dagSegment)
//			if err != nil {
//				log.WithField("CycleID", cycleDetails.CycleID).Errorf("Unable to unmarshal dagChainSegment data due to error %+v", err)
//				continue
//			}
//			log.WithField("CycleID", cycleDetails.CycleID).Debugf("Processing DAG Segment at height %d for project %s", dagSegment.BeginHeight, projectId)
//
//			if dagSegment.StorageType == DAG_CHAIN_STORAGE_TYPE_PENDING {
//				log.WithField("CycleID", cycleDetails.CycleID).Infof("Performing Archival for project %s segment with endHeight %d", projectId, dagSegmentEndHeight)
//				projectReport.DAGSegmentsProcessed++
//				if settingsObj.PruningServiceSettings.PerformArchival {
//					opStatus, err := ArchiveDAG(projectId, dagSegment.BeginHeight,
//						dagSegment.EndHeight, dagSegment.EndDAGCID)
//					if opStatus {
//						dagSegment.StorageType = DAG_CHAIN_STORAGE_TYPE_COLD
//					} else {
//						log.WithField("CycleID", cycleDetails.CycleID).Errorf("Failed to Archive DAG for project %s at height %d due to error.", projectId, dagSegmentEndHeight)
//						projectReport.DAGSegmentsArchivalFailed++
//						projectReport.ArchivalFailureCause = err.Error()
//						UpdatePruningProjectReportInRedis(&projectReport, projectPruneState)
//						return -2
//					}
//				} else {
//					log.Infof("Archival disabled, hence proceeding with pruning")
//					dagSegment.StorageType = DAG_CHAIN_STORAGE_TYPE_PRUNED
//				}
//				startScore = dagSegment.BeginHeight
//				payloadCids := GetPayloadCidsFromRedis(projectId, startScore, dagSegmentEndHeight)
//				if payloadCids == nil {
//					projectReport.UnPinFailed += dagSegment.EndHeight - dagSegment.BeginHeight
//					projectReport.ArchivalFailureCause = "Failed to fetch payloadCids from Redis"
//					UpdatePruningProjectReportInRedis(&projectReport, projectPruneState)
//					return -2
//				}
//				dagCids := GetDAGCidsFromRedis(projectId, startScore, dagSegmentEndHeight)
//				if dagCids == nil {
//					projectReport.UnPinFailed += dagSegment.EndHeight - dagSegment.BeginHeight
//					projectReport.ArchivalFailureCause = "Failed to fetch DAGCids from Redis"
//					UpdatePruningProjectReportInRedis(&projectReport, projectPruneState)
//					return -2
//				}
//				log.WithField("CycleID", cycleDetails.CycleID).Infof("Unpinning DAG CIDS from IPFS for project %s segment with endheight %d", projectId, dagSegmentEndHeight)
//				errCount := ipfsClient.UnPinCidsFromIPFS(projectId, dagCids)
//				projectReport.UnPinFailed += errCount
//				log.WithField("CycleID", cycleDetails.CycleID).Infof("Unpinning payload CIDS from IPFS for project %s segment with endheight %d", projectId, dagSegmentEndHeight)
//				errCount = ipfsClient.UnPinCidsFromIPFS(projectId, payloadCids)
//				projectReport.UnPinFailed += errCount
//				UpdateDagSegmentStatusToRedis(projectId, dagSegmentEndHeight, &dagSegment)
//
//				projectPruneState.LastPrunedHeight = dagSegmentEndHeight
//				if settingsObj.PruningServiceSettings.PruneRedisZsets {
//					if settingsObj.PruningServiceSettings.BackUpRedisZSets {
//						BackupZsetsToFile(projectId, startScore, dagSegmentEndHeight, payloadCids, dagCids)
//					}
//					log.WithField("CycleID", cycleDetails.CycleID).Infof("Pruning redis Zsets from IPFS for project %s segment with endheight %d", projectId, dagSegmentEndHeight)
//					PruneProjectInRedis(projectId, startScore, dagSegmentEndHeight)
//				}
//				UpdatePrunedStatusToRedis(projectPruneState)
//				errCount = DeleteContentFromLocalCache(projectId, dagCids, payloadCids)
//				projectReport.LocalCacheDeletionsFailed += errCount
//				projectReport.DAGSegmentsArchived++
//				projectReport.CIDsUnPinned += len(*payloadCids) + len(*dagCids)
//				projectProcessed = true
//			}
//		}
//	}
//	if projectProcessed {
//		UpdatePruningProjectReportInRedis(&projectReport, projectPruneState)
//		return 0
//	}
//	return 1
//}

/*func DeleteContentFromLocalCache(projectId string, dagCids *map[int]string, payloadCids *map[int]string) int {
	path := settingsObj.PayloadCachePath
	errCount := 0
	for _, cid := range *dagCids {
		fileName := fmt.Sprintf("%s/%s/%s.json", path, projectId, cid)
		err := os.Remove(fileName)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				continue
			}
			log.Errorf("Failed to remove file %s from local cache due to error %+v", fileName, err)
			//TODO: Need to have some sort of pruning files older than 8 days logic to handle failures.
			errCount++
		}
		log.Debugf("Deleted file %s successfully from local cache", fileName)
	}

	for _, cid := range *payloadCids {
		fileName := fmt.Sprintf("%s/%s/%s.json", path, projectId, cid)
		err := os.Remove(fileName)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				continue
			}
			log.Errorf("Failed to remove file %s from local cache due to error %+v", fileName, err)
			errCount++
		}
		log.Debugf("Deleted file %s successfully from local cache", fileName)
	}
	return errCount
}*/

/*func BackupZsetsToFile(projectId string, startScore int, endScore int, payloadCids *map[int]string, dagCids *map[int]string) {
	path := settingsObj.PruningServiceSettings.CARStoragePath
	fileName := fmt.Sprintf("%s%s__%d_%d.json", path, projectId, startScore, endScore)
	file, err := os.Create(fileName)
	if err != nil {
		log.WithField("CycleID", cycleDetails.CycleID).Errorf("Unable to create file %s in specified path due to errro %+v", fileName, err)
	}
	defer file.Close()

	type ZSets struct {
		PayloadCids *map[int]string `json:"payloadCids"`
		DagCids     *map[int]string `json:"dagCids"`
	}

	zSets := ZSets{PayloadCids: payloadCids, DagCids: dagCids}

	zSetsJson, err := json.Marshal(zSets)
	if err != nil {
		log.WithField("CycleID", cycleDetails.CycleID).Fatalf("Failed to marshal payloadCids map to json due to error %+v", err)
	}

	bytesWritten, err := file.Write(zSetsJson)
	if err != nil {
		log.WithField("CycleID", cycleDetails.CycleID).Errorf("Failed to write payloadCidsJson to file %s due to error %+v", fileName, err)
	} else {
		log.WithField("CycleID", cycleDetails.CycleID).Debugf("Wrote %d bytes of payloadCids successfully to file %s.", bytesWritten, fileName)
		file.Sync()
	}
}*/

/*func PruneProjectInRedis(projectId string, startScore int, endScore int) {
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectId)
	PruneZSetInRedis(key, startScore, endScore)
	key = fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectId)
	PruneZSetInRedis(key, startScore, endScore)
}*/

/*func PruneZSetInRedis(key string, startScore int, endScore int) {
	i := 0
	for ; i < 3; i++ {
		res := redisClient.ZRemRangeByScore(
			ctx, key,
			"-inf", //Always prune from start
			strconv.Itoa(endScore),
		)
		if res.Err() != nil {
			log.WithField("CycleID", cycleDetails.CycleID).Warnf("Could not prune redis Zset %s between height %d to %d due to error %+v. Retrying %d.",
				key, startScore, endScore, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}

		log.WithField("CycleID", cycleDetails.CycleID).Debugf("Successfully pruned redis Zset %s of %d entries between height %d and %d",
			key, res.Val(), startScore, endScore)
		break
	}
	if i >= 3 {
		log.WithField("CycleID", cycleDetails.CycleID).Errorf("Could not prune redis Zset %s between height %d to %d even after max-retries.",
			key, startScore, endScore)
	}
}*/

/*func ExportDAGFromIPFS(projectId string, fromHeight int, toHeight int, dagCID string) (string, bool) {
	log.WithField("CycleID", cycleDetails.CycleID).Debugf("Exporting DAG for project %s from height %d to height %d with last DAG CID %s", projectId, fromHeight, toHeight, dagCID)
	dagExportSuffix := "/api/v0/dag/export"
	reqURL := "http://" + ipfsHTTPURL + dagExportSuffix + "?arg=" + dagCID + "&encoding=json&stream-channels=true&progress=false"
	log.WithField("CycleID", cycleDetails.CycleID).Debugf("Sending request to URL %s", reqURL)
	for retryCount := 0; retryCount < 3; {
		if retryCount == *settingsObj.RetryCount {
			log.WithField("CycleID", cycleDetails.CycleID).Errorf("CAR export failed for project %s at height %d after max-retry of %d",
				projectId, fromHeight, settingsObj.RetryCount)
			return "", false
		}
		req, err := http.NewRequest(http.MethodPost, reqURL, nil)
		if err != nil {
			log.WithField("CycleID", cycleDetails.CycleID).Fatalf("Failed to create new HTTP Req with URL %s with error %+v",
				reqURL, err)
			return "", false
		}

		log.WithField("CycleID", cycleDetails.CycleID).Debugf("Sending Req to IPFS URL %s for project %s at height %d ",
			reqURL, projectId, fromHeight)
		res, err := ipfsHttpClient.Do(req)
		if err != nil {
			retryCount++
			log.WithField("CycleID", cycleDetails.CycleID).Warnf("Failed to send request %+v towards IPFS URL %s for project %s at height %d with error %+v.  Retrying %d",
				req, reqURL, projectId, fromHeight, err, retryCount)
			continue
		}
		defer res.Body.Close()

		if err != nil {
			retryCount++
			log.WithField("CycleID", cycleDetails.CycleID).Warnf("Failed to read response body for project %s at height %d from IPFS with error %+v. Retrying %d",
				projectId, fromHeight, err, retryCount)
			continue
		}
		if res.StatusCode == http.StatusOK {
			log.WithField("CycleID", cycleDetails.CycleID).Debugf("Received 200 OK from IPFS for project %s at height %d",
				projectId, fromHeight)
			path := settingsObj.PruningServiceSettings.CARStoragePath
			fileName := fmt.Sprintf("%s%s_%d_%d.car", path, projectId, fromHeight, toHeight)
			file, err := os.Create(fileName)
			if err != nil {
				log.WithField("CycleID", cycleDetails.CycleID).Errorf("Unable to create file %s in specified path due to errro %+v", fileName, err)
				return "", false
			}
			defer file.Close()
			fileWriter := bufio.NewWriter(file)
			//TODO: optimize for larger files.
			bytesWritten, err := io.Copy(fileWriter, res.Body)
			if err != nil {
				retryCount++
				log.WithField("CycleID", cycleDetails.CycleID).Warnf("Failed to write to %s due to error %+v. Retrying %d", fileName, err, retryCount)
				time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Minute)
				continue
			}
			fileWriter.Flush()
			log.WithField("CycleID", cycleDetails.CycleID).Debugf("Wrote %d bytes CAR to local file %s successfully", bytesWritten, fileName)
			return fileName, true
		} else {
			retryCount++
			log.WithField("CycleID", cycleDetails.CycleID).Warnf("Received Error response from IPFS for project %s at height %d with statusCode %d and status : %s ",
				projectId, fromHeight, res.StatusCode, res.Status)
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Minute)
			continue
		}
	}
	log.WithField("CycleID", cycleDetails.CycleID).Debugf("Failed to export DAG from ipfs for project %s at height %d after max retries.", projectId, toHeight)
	return "", false
}*/

/*func UploadFileToWeb3Storage(fileName string) (string, bool) {

	file, err := os.Open(fileName)
	if err != nil {
		log.WithField("CycleID", cycleDetails.CycleID).Errorf("Unable to open file %s due to error %+v", fileName, err)
		return "", false
	}
	fileStat, err := file.Stat()
	if err != nil {
		log.WithField("CycleID", cycleDetails.CycleID).Errorf("Unable to stat file %s due to error %+v", fileName, err)
		return "", false
	}
	targetSize := settingsObj.PruningServiceSettings.Web3Storage.UploadChunkSizeMB * 1024 * 1024 // 100MiB chunks

	fileReader := bufio.NewReader(file)

	if fileStat.Size() > int64(targetSize) {
		log.WithField("CycleID", cycleDetails.CycleID).Infof("File size greater than targetSize %d bytes..doing chunking", targetSize)
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
				log.WithField("CycleID", cycleDetails.CycleID).Fatalf("Failed to split car file %s due to error %+v", fileName, err)
				return "", false
			}
			lastCID, opStatus = UploadChunkToWeb3Storage(fileName, car)
			if !opStatus {
				log.WithField("CycleID", cycleDetails.CycleID).Errorf("Failed to upload chunk %d for file %s. aborting complete file.", i, fileName)
				return "", false
			}
			log.WithField("CycleID", cycleDetails.CycleID).Debugf("Uploaded chunk %d of file %s to web3.storage successfully", i, fileName)
		}
		return lastCID, true
	} else {
		return UploadChunkToWeb3Storage(fileName, fileReader)
	}
}

func UploadChunkToWeb3Storage(fileName string, fileReader io.Reader) (string, bool) {

	reqURL := settingsObj.Web3Storage.URL + "/car"
	for retryCount := 0; retryCount < 3; {
		if retryCount == *settingsObj.RetryCount {
			log.WithField("CycleID", cycleDetails.CycleID).Errorf("web3.storage upload failed for file %s after max-retry of %d",
				fileName, *settingsObj.RetryCount)
			return "", false
		}
		req, err := http.NewRequest(http.MethodPost, reqURL, fileReader)
		if err != nil {
			log.WithField("CycleID", cycleDetails.CycleID).Fatalf("Failed to create new HTTP Req with URL %s for message %+v with error %+v",
				reqURL, err)
			return "", false
		}
		req.Header.Add("Authorization", "Bearer "+settingsObj.Web3Storage.APIToken)
		req.Header.Add("accept", "application/vnd.ipld.car")

		err = web3StorageClientRateLimiter.Wait(context.Background())
		if err != nil {
			log.WithField("CycleID", cycleDetails.CycleID).Warnf("Web3Storage Rate Limiter wait timeout with error %+v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.WithField("CycleID", cycleDetails.CycleID).Debugf("Sending Req to web3.storage URL %s for file %s",
			reqURL, fileName)
		res, err := w3sHttpClient.Do(req)
		if err != nil {
			retryCount++
			log.WithField("CycleID", cycleDetails.CycleID).Warnf("Failed to send request %+v towards web3.storage URL %s for fileName %s with error %+v.  Retrying %d",
				req, reqURL)
			continue
		}
		defer res.Body.Close()
		var resp Web3StoragePostResponse
		respBody, err := io.ReadAll(res.Body)
		if err != nil {
			retryCount++
			log.WithField("CycleID", cycleDetails.CycleID).Warnf("Failed to read response body for fileName %s from web3.storage with error %+v. Retrying %d",
				err, retryCount)
			continue
		}
		if res.StatusCode == http.StatusOK {
			err = json.Unmarshal(respBody, &resp)
			if err != nil {
				retryCount++
				log.WithField("CycleID", cycleDetails.CycleID).Warnf("Failed to unmarshal response %+v for fileName %s towards web3.storage with error %+v. Retrying %d",
					respBody, fileName, err, retryCount)
				time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
				continue
			}
			log.WithField("CycleID", cycleDetails.CycleID).Debugf("Received 200 OK from web3.storage for fileName %s with CID %s ",
				fileName, resp.CID)
			return resp.CID, true
		} else {
			if res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusForbidden ||
				res.StatusCode == http.StatusUnauthorized {
				log.WithField("CycleID", cycleDetails.CycleID).Warnf("Failed to upload to web3.storage due to error %+v with statusCode %d", resp, res.StatusCode)
				return "", false
			}
			retryCount++
			var resp Web3StorageErrResponse
			err = json.Unmarshal(respBody, &resp)
			if err != nil {
				log.WithField("CycleID", cycleDetails.CycleID).Errorf("Failed to unmarshal error response %+v for fileName %s towards web3.storage with error %+v. Retrying %d",
					respBody, fileName, err, retryCount)
			} else {
				log.WithField("CycleID", cycleDetails.CycleID).Warnf("Received Error response %+v from web3.storage for fileName %s with statusCode %d and status : %s ",
					resp, fileName, res.StatusCode, res.Status)
			}
			time.Sleep(time.Duration(settingsObj.RetryIntervalSecs) * time.Second)
			continue
		}
	}
	log.WithField("CycleID", cycleDetails.CycleID).Errorf("Failed to upload file %s to web3.storage after max retries", fileName)
	return "", false
}*/

/*func GetPayloadCidsFromRedis(projectId string, startScore int, endScore int) *map[int]string {
	cids := make(map[int]string, endScore-startScore)

	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_PAYLOAD_CIDS, projectId)
	for i := 0; i < 5; i++ {
		res := redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
			Min: strconv.Itoa(startScore),
			Max: strconv.Itoa(endScore),
		})
		if res.Err() != nil {
			log.WithField("CycleID", cycleDetails.CycleID).Warnf("Could not fetch payloadCids for project %s due to error %+v. Retrying %d.", projectId, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		for j := range res.Val() {
			cids[int(res.Val()[j].Score)] = fmt.Sprintf("%v", res.Val()[j].Member)
		}
		log.WithField("CycleID", cycleDetails.CycleID).Debugf("Fetched %d payload Cids from redis for project %s", len(cids), projectId)
		return &cids
	}
	log.WithField("CycleID", cycleDetails.CycleID).Errorf("Could not fetch payloadCids for project %s after max retries.", projectId)
	return nil
}*/

/*func GetDAGCidsFromRedis(projectId string, startScore int, endScore int) *map[int]string {
	cids := make(map[int]string, endScore-startScore)
	key := fmt.Sprintf(redisutils.REDIS_KEY_PROJECT_CIDS, projectId)
	for i := 0; i < 5; i++ {
		res := redisClient.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
			Min: strconv.Itoa(startScore),
			Max: strconv.Itoa(endScore),
		})
		if res.Err() != nil {
			log.WithField("CycleID", cycleDetails.CycleID).Warnf("Could not fetch DAGCids for project %s due to error %+v. Retrying %d.", projectId, res.Err(), i)
			time.Sleep(5 * time.Second)
			continue
		}
		for j := range res.Val() {
			cids[int(res.Val()[j].Score)] = fmt.Sprintf("%v", res.Val()[j].Member)
		}
		log.WithField("CycleID", cycleDetails.CycleID).Debugf("Fetched %d DAG Cids from redis for project %s", len(cids), projectId)
		return &cids
	}
	log.WithField("CycleID", cycleDetails.CycleID).Errorf("Could not fetch DAGCids for project %s after max retries.", projectId)
	return nil
}*/

// moved this to service/service.go
/*func GetProjectsListFromRedis(redisClient) {
	key := redisutils.REDIS_KEY_STORED_PROJECTS
	log.Debugf("Fetching stored Projects from redis at key: %s", key)
	for i := 0; ; i++ {
		res := redisClient.SMembers(ctx, key)
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
		projectList = make(map[string]*ProjectPruneState, len(res.Val()))
		//projectList = make(map[string]*ProjectPruneState, 375)
		for i := range res.Val() {
			projectId := res.Val()[i]
			//if strings.Contains(projectId, "uniswap_V2PairsSummarySnapshot_UNISWAPV2") {
			projectPruneState := ProjectPruneState{projectId, 0, false}
			projectList[projectId] = &projectPruneState
			//	break
			//}
		}
		log.Infof("Retrieved %d storedProjects %+v from redis", len(res.Val()), projectList)
		return
	}
}*/

// moved this function to goutils/httpclient/http_client.go
/*func InitIPFSHTTPClient() {
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
}*/

// moved this function to goutils/httpclient/http_client.go
/*func InitW3sClient() {
	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        settingsObj.Web3Storage.MaxIdleConns,
		MaxConnsPerHost:     settingsObj.Web3Storage.MaxIdleConns,
		MaxIdleConnsPerHost: settingsObj.Web3Storage.MaxIdleConns,
		IdleConnTimeout:     time.Duration(settingsObj.Web3Storage.IdleConnTimeout),
		DisableCompression:  true,
	}

	w3sHttpClient = http.Client{
		Timeout:   time.Duration(settingsObj.PruningServiceSettings.Web3Storage.TimeoutSecs) * time.Second,
		Transport: &t,
	}

	//Default values
	tps := rate.Limit(1) //3 TPS
	burst := 1
	if settingsObj.PruningServiceSettings.Web3Storage.RateLimit != nil {
		burst = settingsObj.PruningServiceSettings.Web3Storage.RateLimit.Burst
		if settingsObj.PruningServiceSettings.Web3Storage.RateLimit.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(settingsObj.PruningServiceSettings.Web3Storage.RateLimit.RequestsPerSec)
		}
	}
	log.Infof("Rate Limit configured for web3.storage at %v TPS with a burst of %d", tps, burst)
	web3StorageClientRateLimiter = rate.NewLimiter(tps, burst)
}*/
