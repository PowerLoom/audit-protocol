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
	"time"

	"github.com/alanshaw/go-carbites"
	"github.com/cenkalti/backoff/v4"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/commonutils"
	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/pruning-archival/constants"
	"audit-protocol/pruning-archival/models"
)

type PruningService struct {
	settingsObj *settings.SettingsObj
	caching     *caching
	ipfsClient  *ipfsutils.IpfsClient
}

// PruningTask is the task that is sent task manager
type PruningTask struct {
	ProjectID string `json:"projectID"`
}

func InitPruningService(settingsObj *settings.SettingsObj, redisClient *redis.Client, ipfsClient *ipfsutils.IpfsClient) *PruningService {
	return &PruningService{
		settingsObj: settingsObj,
		caching:     &caching{redisClient: redisClient},
		ipfsClient:  ipfsClient,
	}
}

// Run runs the pruning task
func (p *PruningService) Run(msgBody []byte) error {
	taskDetails := new(models.PruningTaskDetails)
	taskDetails.StartTime = time.Now().UnixMilli()
	taskDetails.TaskID = uuid.New().String()

	log.WithField("TaskID", taskDetails.TaskID).WithField("task", string(msgBody)).Infof("received task")

	task := new(PruningTask)

	err := json.Unmarshal(msgBody, task)
	if err != nil {
		log.Errorf("Failed to unmarshal task: %v", err)

		return err
	}

	if task.ProjectID == "" {
		log.Errorf("ProjectID is empty")

		return errors.New("ProjectID is empty")
	}

	if !p.settingsObj.PruningServiceSettings.PerformArchival && !p.settingsObj.PruningServiceSettings.PerformIPFSUnPin &&
		!p.settingsObj.PruningServiceSettings.PruneRedisZsets {
		log.Infof("None of the pruning features enabled. Not doing anything")
		time.Sleep(time.Duration(p.settingsObj.PruningServiceSettings.RunIntervalMins) * time.Minute)

		return nil
	}

	// get the last pruned height for each project from redis
	lastPrunedHeight, err := p.caching.GetLastPrunedHeightOfProjectFromRedis(task.ProjectID, taskDetails.TaskID)
	if err != nil {
		return err
	}

	err = p.pruneProjectSegments(task.ProjectID, lastPrunedHeight, taskDetails)
	if err != nil {
		return err
	}

	log.WithField("TaskID", taskDetails.TaskID).Infof("Completed task")

	// TODO: Cleanup storage path if it has old files

	return nil
}

// pruneProjectSegment prunes safe to prune dag segments in pending state for the given project
func (p *PruningService) pruneProjectSegments(projectID string, lastPrunedHeight int, taskDetails *models.PruningTaskDetails) error {
	l := log.WithField("TaskID", taskDetails.TaskID).WithField("ProjectID", projectID)

	projectReport := &models.ProjectPruningReport{ProjectID: projectID}

	// Fetch project dagSegments from redis
	dagSegments, err := p.caching.FetchProjectDagSegments(projectID, taskDetails.TaskID)
	if err != nil {
		return err
	}

	// finding safe to prune dag height
	heightToPrune := p.FindPruningHeight(lastPrunedHeight, projectID, taskDetails.TaskID)

	if heightToPrune <= lastPrunedHeight {
		l.Debugf("Nothing to Prune for project %s", projectID)

		_ = p.caching.UpdatePruningProjectReportInRedis(taskDetails, projectReport)

		return nil
	}

	l.Debugf("Height to Prune is %d", heightToPrune)

	// gets the dagSegments end height in sorted order
	sortedDagSegments := commonutils.SortKeysAsNumber(dagSegments.DagChains)

	for _, dagSegmentEndHeightStr := range sortedDagSegments {
		dagChainSegment := dagSegments.DagChains[dagSegmentEndHeightStr]

		dagSegmentEndHeight := 0

		dagSegmentEndHeight, err = strconv.Atoi(dagSegmentEndHeightStr)
		if err != nil {
			l.Errorf("dagSegmentEndHeight %s is not an integer.", dagSegmentEndHeightStr)

			return errors.New("dagSegmentEndHeight is not an integer")
		}

		// there can be multiple dag segments in pending state, so we need to process all of them
		// start with the dag segment with the lowest end height first
		if dagSegmentEndHeight < heightToPrune {
			dagSegment := new(models.ProjectDAGSegment)

			err = json.Unmarshal([]byte(dagChainSegment), dagSegment)
			if err != nil {
				log.WithField("TaskID", taskDetails.TaskID).Errorf("Unable to unmarshal dagChainSegment data due to error %+v", err)

				continue
			}

			log.WithField("TaskID", taskDetails.TaskID).Debugf("Processing DAG Segment at height %d for project %s", dagSegment.BeginHeight, projectID)

			// if the dag segment is in pending state, perform archival/pruning
			if dagSegment.StorageType == constants.DAG_CHAIN_STORAGE_TYPE_PENDING {
				err = p.archiveAndPruneSegment(projectID, taskDetails, dagSegment, projectReport)
				if err != nil {
					return err
				}
			}
		}
	}

	if projectReport.DAGSegmentsArchived == 0 {
		log.Infof("No segments to prune for project %s", projectID)
	}

	err = p.caching.UpdatePruningProjectReportInRedis(taskDetails, projectReport)
	if err != nil {
		l.WithError(err).Errorf("Failed to update project report in redis")

		// TODO: send report to slack
	}

	taskDetails.EndTime = time.Now().UnixMilli()

	// ignoring error as it is not critical but can be sent to slack
	_ = p.caching.UpdatePruningCycleDetailsInRedis(taskDetails, p.settingsObj.PruningServiceSettings.RunIntervalMins)

	return nil
}

// archiveAndPruneSegment performs archival and pruning for the given dag segment
func (p *PruningService) archiveAndPruneSegment(projectID string, taskDetails *models.PruningTaskDetails, dagSegment *models.ProjectDAGSegment, projectReport *models.ProjectPruningReport) error {
	l := log.WithField("TaskID", taskDetails.TaskID).WithField("ProjectID", projectID)
	dagSegmentEndHeight := dagSegment.EndHeight

	l.Infof("Performing Archival for segment with endHeight %d", dagSegment.EndHeight)

	projectReport.DAGSegmentsProcessed++

	// if archival is enabled, perform archival and after success update the storage type to cold
	// else prune the dag segment
	if p.settingsObj.PruningServiceSettings.PerformArchival {
		err := p.ArchiveDAG(projectID, taskDetails.TaskID, dagSegment.BeginHeight,
			dagSegment.EndHeight, dagSegment.EndDAGCID)

		if err == nil {
			dagSegment.StorageType = constants.DAG_CHAIN_STORAGE_TYPE_COLD
		} else {
			l.Errorf("Failed to Archive DAG for project %s at height %d due to error.", projectID, dagSegmentEndHeight)
			projectReport.DAGSegmentsArchivalFailed++
			projectReport.ArchivalFailureCause = err.Error()
			_ = p.caching.UpdatePruningProjectReportInRedis(taskDetails, projectReport)

			return err
		}
	} else {
		l.Infof("Archival disabled, hence proceeding with pruning")
		dagSegment.StorageType = constants.DAG_CHAIN_STORAGE_TYPE_PRUNED
	}

	dagSegmentStartHeight := dagSegment.BeginHeight

	// get all cid's for the project from range of startScore and dagSegmentEndHeight
	payloadCids, err := p.caching.GetPayloadCidsFromRedis(projectID, taskDetails.TaskID, dagSegmentStartHeight, dagSegmentEndHeight)
	if err != nil {
		projectReport.UnPinFailed += dagSegment.EndHeight - dagSegment.BeginHeight
		projectReport.ArchivalFailureCause = "Failed to fetch payloadCids from Redis"
		_ = p.caching.UpdatePruningProjectReportInRedis(taskDetails, projectReport)

		return err
	}

	dagCids, err := p.caching.GetDAGCidsFromRedis(projectID, taskDetails.TaskID, dagSegmentStartHeight, dagSegmentEndHeight)
	if dagCids == nil {
		projectReport.UnPinFailed += dagSegment.EndHeight - dagSegment.BeginHeight
		projectReport.ArchivalFailureCause = "Failed to fetch DAGCids from Redis"
		_ = p.caching.UpdatePruningProjectReportInRedis(taskDetails, projectReport)

		return err
	}

	log.WithField("TaskID", taskDetails.TaskID).Infof("Unpinning DAG CIDS from IPFS for project %s segment with endheight %d", projectID, dagSegmentEndHeight)

	bkoff := backoff.WithMaxRetries(backoff.NewConstantBackOff(1*time.Second), uint64(*p.settingsObj.RetryCount))

	// remove cids from ipfs AKA unpin
	// TBD: if unpin fails (not because it was not pinned), task should be marked as failed
	err = backoff.Retry(func() error { return p.ipfsClient.UnPinCidsFromIPFS(projectID, dagCids) }, bkoff)
	if err != nil {
		return err
	}

	log.WithField("TaskID", taskDetails.TaskID).Infof("Unpinning payload CIDS from IPFS for project %s segment with endheight %d", projectID, dagSegmentEndHeight)

	err = backoff.Retry(func() error { return p.ipfsClient.UnPinCidsFromIPFS(projectID, payloadCids) }, bkoff)
	if err != nil {
		return err
	}

	err = p.caching.UpdateDagSegmentStatusToRedis(taskDetails.TaskID, projectID, dagSegmentEndHeight, dagSegment)
	if err != nil {
		return err
	}

	if p.settingsObj.PruningServiceSettings.PruneRedisZsets {
		if p.settingsObj.PruningServiceSettings.BackUpRedisZSets {
			err = p.BackupZsetsToFile(projectID, taskDetails.TaskID, dagSegmentStartHeight, dagSegmentEndHeight, payloadCids, dagCids)
			if err != nil {
				return err
			}
		}

		log.WithField("TaskID", taskDetails.TaskID).Infof("Pruning redis Zsets from IPFS for project %s segment with endheight %d", projectID, dagSegmentEndHeight)

		err = p.caching.PruneProjectCIDsInRedis(taskDetails.TaskID, projectID, dagSegmentStartHeight, dagSegmentEndHeight)
		if err != nil {
			return err
		}
	}

	err = p.caching.UpdatePrunedStatusToRedis(projectID, taskDetails.TaskID, dagSegmentEndHeight)
	if err != nil {
		return err
	}

	err = p.caching.DeleteContentFromLocalDrive(projectID, p.settingsObj.PayloadCachePath, dagCids, payloadCids)
	if err != nil {
		return err
	}

	projectReport.DAGSegmentsArchived++
	projectReport.CIDsUnPinned += len(payloadCids) + len(dagCids)

	return nil
}

// BackupZsetsToFile backs up the zsets from redis to a file
func (p *PruningService) BackupZsetsToFile(projectID, taskID string, startScore int, endScore int, payloadCids, dagCids map[int]string) error {
	path := p.settingsObj.PruningServiceSettings.CARStoragePath
	fileName := fmt.Sprintf("%s%s__%d_%d.json", path, projectID, startScore, endScore)

	file, err := os.Create(fileName)
	if err != nil {
		log.WithField("TaskID", taskID).Errorf("Unable to create file %s in specified path due to errro %+v", fileName, err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.WithField("TaskID", taskID).Errorf("Failed to close file %s due to error %+v", fileName, err)
		}
	}(file)

	zSets := models.ZSets{PayloadCids: payloadCids, DagCids: dagCids}

	zSetsJson, err := json.Marshal(zSets)
	if err != nil {
		log.WithField("TaskID", taskID).Fatalf("Failed to marshal payloadCids map to json due to error %+v", err)
	}

	bytesWritten, err := file.Write(zSetsJson)
	if err != nil {
		log.WithField("TaskID", taskID).Errorf("Failed to write payloadCidsJson to file %s due to error %+v", fileName, err)
		return err
	}

	log.WithField("TaskID", taskID).Debugf("Wrote %d bytes of payloadCids successfully to file %s.", bytesWritten, fileName)
	_ = file.Sync()

	return nil
}

// FindPruningHeight  finds next height of dag chain to prune.
func (p *PruningService) FindPruningHeight(lastPrunedHeight int, projectID, taskID string) int {
	heightToPrune := lastPrunedHeight

	// Fetch oldest height used by indexers.
	oldestIndexedHeight := p.caching.GetOldestIndexedProjectHeight(projectID, taskID, p.settingsObj)

	if oldestIndexedHeight != -1 {
		// Adding a buffer just in case 7d index is just crossed and some heights before it are used in sliding window.
		heightToPrune = oldestIndexedHeight - p.settingsObj.PruningServiceSettings.PruningHeightBehindOldestIndex
	}

	return heightToPrune
}

// ArchiveDAG archives the DAG from IPFS and uploads it to Web3 Storage.
func (p *PruningService) ArchiveDAG(projectID, cycleID string, startScore int, endScore int, lastDagCid string) error {
	var errToReturn error

	// Export DAG from IPFS.
	fileName := ""
	bkoff := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), uint64(*p.settingsObj.RetryCount))

	err := backoff.Retry(func() error {
		var err error

		fileName, err = p.ExportDAGFromIPFS(projectID, cycleID, startScore, endScore, lastDagCid)
		if err != nil {
			return err
		}

		return err
	}, bkoff)
	if err != nil {
		log.WithField("TaskID", cycleID).Errorf("Unable to export DAG for project %s at height %d. Will retry in next cycle", projectID, endScore)

		return errors.New("failed to export CAR File from IPFS at height " + strconv.Itoa(endScore))
	}

	CID := ""
	// Can be Optimized: Consider batch upload of files if sizes are too small.
	err = backoff.Retry(func() error {
		CID, err = p.UploadFileToWeb3Storage(fileName, cycleID)
		if err != nil {
			return err
		}

		return nil
	}, bkoff)

	if err != nil {
		log.WithField("TaskID", cycleID).Debugf("Failed to upload CAR file %s to web3.storage", fileName)
		errToReturn = errors.New("failed to upload CAR File to web3.storage" + fileName)
	} else {
		log.WithField("TaskID", cycleID).Debugf("CID of CAR file %s uploaded to web3.storage is %s", fileName, CID)
	}

	// Delete file from local storage.
	if err := os.Remove(fileName); err != nil {
		log.WithField("TaskID", cycleID).Errorf("Failed to delete file %s due to error %+v", fileName, err)
	}

	log.WithField("TaskID", cycleID).Debugf("Deleted file %s successfully from local storage", fileName)

	return errToReturn
}

// ExportDAGFromIPFS exports the DAG from IPFS as CAR file.
func (p *PruningService) ExportDAGFromIPFS(projectID, cycleID string, fromHeight int, toHeight int, dagCID string) (string, error) {
	log.WithField("TaskID", cycleID).Debugf("Exporting DAG for project %s from height %d to height %d with last DAG CID %s", projectID, fromHeight, toHeight, dagCID)

	dagExportSuffix := "/api/v0/dag/export"
	host := ipfsutils.ParseMultiAddrUrl(p.settingsObj.IpfsConfig.ReaderURL)

	reqURL := "http://" + host + dagExportSuffix + "?arg=" + dagCID + "&encoding=json&stream-channels=true&progress=false"
	log.WithField("TaskID", cycleID).Debugf("Sending request to URL %s", reqURL)

	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		log.WithField("TaskID", cycleID).Fatalf("Failed to create new HTTP Req with URL %s with error %+v",
			reqURL, err)

		return "", err
	}

	log.WithField("TaskID", cycleID).Debugf("Sending Req to IPFS URL %s for project %s at height %d ",
		reqURL, projectID, fromHeight)

	ipfsHTTPClient := httpclient.GetIPFSHTTPClient(p.settingsObj)

	res, err := ipfsHTTPClient.Do(req)
	if err != nil {
		log.WithField("TaskID", cycleID).Warnf("Failed to send request %+v towards IPFS URL %s for project %s at height %d with error %+v", req, reqURL, projectID, fromHeight, err)

		return "", err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.WithField("TaskID", cycleID).Errorf("Failed to close response body for project %s at height %d from IPFS with error %+v",
				projectID, fromHeight, err)
		}
	}(res.Body)

	if res.StatusCode == http.StatusOK {
		log.WithField("TaskID", cycleID).Debugf("Received 200 OK from IPFS for project %s at height %d",
			projectID, fromHeight)

		path := p.settingsObj.PruningServiceSettings.CARStoragePath
		fileName := fmt.Sprintf("%s%s_%d_%d.car", path, projectID, fromHeight, toHeight)

		file, err := os.Create(fileName)
		if err != nil {
			log.WithField("TaskID", cycleID).Errorf("Unable to create file %s in specified path due to errro %+v", fileName, err)

			return "", err
		}

		fileWriter := bufio.NewWriter(file)

		// this reads 32kb at a time
		bytesWritten, err := io.Copy(fileWriter, res.Body)
		if err != nil {
			log.WithField("TaskID", cycleID).Warnf("Failed to write to %s due to error %+v", fileName, err)

			return "", err
		}

		err = file.Close()
		if err != nil {
			log.WithField("TaskID", cycleID).Errorf("Failed to close file %s due to error %+v", fileName, err)
		}

		err = fileWriter.Flush()
		if err != nil {
			log.WithField("TaskID", cycleID).Errorf("Failed to flush file %s due to error %+v", fileName, err)
		}

		log.WithField("TaskID", cycleID).Debugf("Wrote %d bytes CAR to local file %s successfully", bytesWritten, fileName)

		return fileName, nil
	}

	log.WithField("TaskID", cycleID).Warnf("Received Error response from IPFS for project %s at height %d with statusCode %d and status : %s ",
		projectID, fromHeight, res.StatusCode, res.Status)

	return "", err
}

// UploadFileToWeb3Storage uploads the file to web3 storage.
func (p *PruningService) UploadFileToWeb3Storage(fileName, cycleID string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		log.WithField("TaskID", cycleID).Errorf("Unable to open file %s due to error %+v", fileName, err)

		return "", err
	}

	fileStat, err := file.Stat()
	if err != nil {
		log.WithField("TaskID", cycleID).Errorf("Unable to stat file %s due to error %+v", fileName, err)

		return "", err
	}

	targetSize := p.settingsObj.PruningServiceSettings.Web3Storage.UploadChunkSizeMB * 1024 * 1024 // 100MiB chunks
	fileReader := bufio.NewReader(file)

	if fileStat.Size() > int64(targetSize) {
		log.WithField("TaskID", cycleID).Infof("File size greater than targetSize %d bytes..doing chunking", targetSize)
		var lastCID string

		// Need to chunk CAR files more than 100MB as web3.storage has size limit right now.
		// Use code from carbites mentioned here https://web3.storage/docs/how-tos/work-with-car-files/
		strategy := carbites.Treewalk
		spltr, _ := carbites.Split(file, targetSize, strategy)

		for i := 1; ; i++ {
			car, err := spltr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}

				log.WithField("TaskID", cycleID).Fatalf("Failed to split car file %s due to error %+v", fileName, err)

				return "", err
			}

			lastCID, err = p.UploadChunkToWeb3Storage(fileName, cycleID, car)
			if err != nil {
				log.WithField("TaskID", cycleID).Errorf("Failed to upload chunk %d for file %s. aborting complete file.", i, fileName)

				return "", err
			}

			log.WithField("TaskID", cycleID).Debugf("Uploaded chunk %d of file %s to web3.storage successfully", i, fileName)
		}

		return lastCID, nil
	}

	return p.UploadChunkToWeb3Storage(fileName, cycleID, fileReader)
}

// UploadChunkToWeb3Storage uploads the chunk to web3 storage.
func (p *PruningService) UploadChunkToWeb3Storage(fileName, cycleID string, fileReader io.Reader) (string, error) {
	reqURL := p.settingsObj.Web3Storage.URL + "/car"

	req, err := http.NewRequest(http.MethodPost, reqURL, fileReader)
	if err != nil {
		log.WithField("TaskID", cycleID).Fatalf("Failed to create new HTTP Req with URL %s for message %+v with error %+v",
			reqURL, err)

		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+p.settingsObj.Web3Storage.APIToken)
	req.Header.Add("accept", "application/vnd.ipld.car")

	w3sHTTPClient, web3StorageClientRateLimiter := httpclient.GetW3sHTTPClient(p.settingsObj)

	err = web3StorageClientRateLimiter.Wait(context.Background())
	if err != nil {
		log.WithField("TaskID", cycleID).Warnf("Web3Storage Rate Limiter wait timeout with error %+v", err)

		return "", err
	}

	log.WithField("TaskID", cycleID).Debugf("Sending Req to web3.storage URL %s for file %s",
		reqURL, fileName)

	res, err := w3sHTTPClient.Do(req)
	if err != nil {
		log.WithField("TaskID", cycleID).Warnf("Failed to send request %+v towards web3.storage URL %s for fileName %s with error %+v.  Retrying %d",
			req, reqURL)

		return "", err
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.WithField("TaskID", cycleID).Errorf("Failed to close response body for fileName %s from web3.storage with error %+v",
				fileName, err)
		}
	}(res.Body)

	resp := new(models.Web3StoragePostResponse)

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.WithError(err).WithField("TaskID", cycleID).Warnf("Failed to read response body for fileName %s from web3.storage", fileName)

		return "", err
	}

	if res.StatusCode == http.StatusOK {
		err = json.Unmarshal(respBody, resp)
		if err != nil {
			log.WithError(err).WithField("TaskID", cycleID).Warnf("Failed to unmarshal response %+v for fileName %s towards web3.storage", respBody, fileName)

			return "", err
		}

		log.WithField("TaskID", cycleID).Debugf("Received 200 OK from web3.storage for fileName %s with CID %s ",
			fileName, resp.CID)

		return resp.CID, nil
	}
	if res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusForbidden ||
		res.StatusCode == http.StatusUnauthorized {
		log.WithField("TaskID", cycleID).Warnf("Failed to upload to web3.storage due to error %+v with statusCode %d", resp, res.StatusCode)

		return "", errors.New("failed to upload to web3.storage")
	}

	errResp := new(models.Web3StorageErrResponse)

	err = json.Unmarshal(respBody, errResp)
	if err != nil {
		log.
			WithError(err).
			WithField("TaskID", cycleID).
			Errorf("Failed to unmarshal error response %+v for fileName %s towards web3.storage", respBody, fileName)

		return "", err
	}

	log.WithField("TaskID", cycleID).Warnf("Received Error response %+v from web3.storage for fileName %s with statusCode %d and status : %s ",
		errResp, fileName, res.StatusCode, res.Status)

	log.WithField("TaskID", cycleID).Errorf("Failed to upload file %s to web3.storage after max retries", fileName)

	return "", errors.New("failed to upload chunk to web3.storage")
}
