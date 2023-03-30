package verifiers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/ipfs/go-cid"
	"github.com/remeh/sizedwaitgroup"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/caching"
	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/slackutils"
	w3storage "audit-protocol/goutils/w3s"
)

type (
	NewBlocksAddedEvent struct {
		ProjectID   string
		StartHeight string // can be "-inf"
		EndHeight   string // can be "+inf"
	}

	// evenMessages for a project
	eventMessages struct {
		mutex    sync.Mutex
		messages []*NewBlocksAddedEvent
	}

	DagVerifier struct {
		redisCache *caching.RedisCache
		diskCache  *caching.LocalDiskCache
		settings   *settings.SettingsObj
		ipfsClient *ipfsutils.IpfsClient
		issues     map[string][]*datamodel.DagVerifierStatus
		w3s        *w3storage.W3S

		// in-memory event queue and status map
		genesisRunStatusMap map[string]bool
		eventMessagesMap    map[string]*eventMessages
	}

	blocksOutOfOrderIssue struct {
		CurrentBlockHeight int64 `json:"currentBlockHeight"`
		NextBlockHeight    int64 `json:"nextBlockHeight"`
		ExpectBlockHeight  int64 `json:"expectBlockHeight"`
	}

	cidHeightsOutOfSyncIssue struct {
		ChainHeightRange *datamodel.DagBlocksHeightRange `json:"chainHeightRange"`
	}

	multipleDagBlocksAtSameHeightIssue struct {
		CIDs []string `json:"cids"`
	}

	DagBlockWithChainRange struct {
		BlockHeight      int64                       `json:"blockHeight"`
		ChainHeightRange *datamodel.ChainHeightRange `json:"chainHeightRange"`
	}

	epochSkippedIssue struct {
		CurrentPayloadChainRange *DagBlockWithChainRange     `json:"currentPayloadChainRange"`
		NextPayloadChainRange    *DagBlockWithChainRange     `json:"nextPayloadChainRange"`
		ExpectedChainRange       *datamodel.ChainHeightRange `json:"expectedChainRange"`
	}

	payloadMissingIssue struct {
		BlockCID string `json:"cid"`
	}

	countMismatchIssue struct {
		CIDCount         int                             `json:"cidCount"`
		PayloadCIDCount  int                             `json:"payloadCIDCount"`
		ChainHeightRange *datamodel.DagBlocksHeightRange `json:"chainHeightRange"`
	}

	internalVerificationErrIssue struct {
		Err         string `json:"err"`
		Description string `json:"description"`
	}
)

const (
	BlocksOutOfOrder              datamodel.DagVerificationIssueType = "BLOCKS_OUT_OF_ORDER"
	EpochSkipped                  datamodel.DagVerificationIssueType = "EPOCH_SKIPPED"
	PayloadMissing                datamodel.DagVerificationIssueType = "PAYLOAD_MISSING"
	DuplicateDagBlock             datamodel.DagVerificationIssueType = "DUPLICATE_DAG_BLOCK"
	PayloadCorrupted              datamodel.DagVerificationIssueType = "PAYLOAD_CORRUPTED"
	CidHeightsOutOfSyncInCache    datamodel.DagVerificationIssueType = "CID_HEIGHTS_OUT_OF_SYNC_IN_CACHE"
	CIDAndPayloadCIDCountMismatch datamodel.DagVerificationIssueType = "CID_AND_PAYLOAD_CID_COUNT_MISMATCH"
	MultipleBlocksAtSameHeight    datamodel.DagVerificationIssueType = "MULTIPLE_BLOCKS_AT_SAME_HEIGHT"
	InternalVerificationErrIssue  datamodel.DagVerificationIssueType = "INTERNAL_VERIFICATION_ERR"
)

func InitDagVerifier() *DagVerifier {
	settingsObj, err := gi.Invoke[*settings.SettingsObj]()
	if err != nil {
		log.WithError(err).Fatal("failed to initialize DAG verifier")
	}

	redisCache, err := gi.Invoke[*caching.RedisCache]()
	if err != nil {
		log.WithError(err).Fatal("failed to initialize DAG verifier")
	}

	diskCache, err := gi.Invoke[*caching.LocalDiskCache]()
	if err != nil {
		log.WithError(err).Fatal("failed to initialize DAG verifier")
	}

	ipfsClient, err := gi.Invoke[*ipfsutils.IpfsClient]()
	if err != nil {
		log.WithError(err).Fatal("failed to initialize DAG verifier")
	}

	w3storageClient, err := gi.Invoke[*w3storage.W3S]()
	if err != nil {
		log.WithError(err).Fatal("failed to initialize DAG verifier")
	}

	verifier := &DagVerifier{
		redisCache:          redisCache,
		diskCache:           diskCache,
		settings:            settingsObj,
		ipfsClient:          ipfsClient,
		issues:              make(map[string][]*datamodel.DagVerifierStatus),
		w3s:                 w3storageClient,
		genesisRunStatusMap: make(map[string]bool),
		eventMessagesMap:    make(map[string]*eventMessages),
	}

	err = gi.Inject(verifier)
	if err != nil {
		log.WithError(err).Fatal("failed to initialize DAG verifier")
	}

	return verifier
}

// Run starts the dag verifier
// this is called on event when new blocks are inserted in chain
// startHeight and endHeight are the range of blocks to verify can be "-inf" to "+inf"
// Run acts as a worker and only cares about the given inputs and acts accordingly
func (d *DagVerifier) Run(newBlocksAddedEvent *NewBlocksAddedEvent, tillGenesis bool) {
	projectID := newBlocksAddedEvent.ProjectID
	startHeight := newBlocksAddedEvent.StartHeight
	endHeight := newBlocksAddedEvent.EndHeight

	l := log.WithField("projectID", projectID).
		WithField("startHeight", startHeight).
		WithField("endHeight", endHeight)

	// check if genesis run is running
	if _, ok := d.genesisRunStatusMap[projectID]; ok && !tillGenesis {
		l.Debug("genesis run is running, skipping dag verification")

		if events, ok := d.eventMessagesMap[projectID]; ok {
			events.mutex.Lock()
			events.messages = append(events.messages, newBlocksAddedEvent)
			events.mutex.Unlock()
		} else {
			d.eventMessagesMap[projectID] = &eventMessages{
				messages: []*NewBlocksAddedEvent{newBlocksAddedEvent},
				mutex:    sync.Mutex{},
			}
		}

		return
	}

	l.Debug("dag verifier started")

	// get last verified dag height from cache
	lastReportedDagHeight := 0
	var err error

	err = backoff.Retry(func() error {
		lastReportedDagHeight, err = d.redisCache.GetLastReportedDagHeight(context.Background(), projectID)
		if err != nil {
			return err
		}

		return nil
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), uint64(*d.settings.RetryCount)))
	if err != nil {
		l.WithError(err).Fatal("failed to get last reported dag height")

		return
	}

	// if lastReportedDagHeight is 0, genesis run should take care of it
	if lastReportedDagHeight == 0 && !tillGenesis {
		l.Debug("last reported dag height is 0, skipping dag verification")
	}

	// if lastReportedDagHeight is greater than endHeight verification is already done for this ranges
	if endHeight != "+inf" {
		endHeightInt, _ := strconv.ParseInt(endHeight, 10, 64)
		if lastReportedDagHeight >= int(endHeightInt) {
			l.Debug("dag verification already done for this range")

			return
		}
	}

	if startHeight != "-inf" {
		startHeight = strconv.Itoa(lastReportedDagHeight)
	}

	dagBlockChain, err := d.redisCache.GetDagChainCIDs(context.Background(), projectID, startHeight, endHeight)
	if err != nil {
		l.WithError(err).Error("failed to get dag chain cids for verification")

		return
	}

	dagChainWithPayloadCIDs, err := d.redisCache.GetPayloadCIDs(context.Background(), projectID, startHeight, endHeight)
	if err != nil {
		l.WithError(err).Error("failed to get payload cids")

		return
	}

	defer func(projectID string, endHeight string) {
		lastHeight, _ := strconv.ParseInt(endHeight, 10, 64)
		d.updateStatusReport(projectID, lastHeight)
	}(projectID, endHeight)

	// this is unlikely to happen but needs to be verified
	if len(dagChainWithPayloadCIDs) != len(dagBlockChain) {
		l.Error("dag cids and payloads cids are out of sync in cache for given lastReportedDagHeight range")

		heightString := strconv.Itoa(int(dagBlockChain[0].Height))

		d.issues[heightString] = append(d.issues[heightString], &datamodel.DagVerifierStatus{
			Timestamp:   time.Now().UnixMicro(),
			BlockHeight: dagBlockChain[0].Height,
			IssueType:   CIDAndPayloadCIDCountMismatch,
			Meta: &countMismatchIssue{
				CIDCount:        len(dagBlockChain),
				PayloadCIDCount: len(dagChainWithPayloadCIDs),
				ChainHeightRange: &datamodel.DagBlocksHeightRange{
					StartHeight: dagBlockChain[0].Height,
					EndHeight:   dagBlockChain[len(dagBlockChain)-1].Height,
				},
			},
		})

		return
	}

	// fill up payload data in dag chain
	// fillPayloadData ideally should not return error because with error in payload data(empty/nil), dag verification is pointless
	err = d.fillPayloadData(projectID, dagBlockChain, dagChainWithPayloadCIDs)
	if err != nil {
		l.WithError(err).Error("error filling up payload data")

		return
	}

	// fills up the dag chain with archived blocks data
	if tillGenesis {
		err = d.traverseArchivedDagSegments(projectID, dagBlockChain)
		if err != nil {
			l.WithError(err).Error("failed to traverse archived dag segments")

			return
		}
	}

	// check for null payload
	d.checkNullPayload(projectID, dagBlockChain)

	// check for blocks out of order
	err = d.checkBlocksOutOfOrder(projectID, dagBlockChain)
	if err != nil {
		l.WithError(err).Error("error checking blocks out of order")

		return
	}

	// check for epoch skipped
	d.checkEpochsOutOfOrder(projectID, dagBlockChain)
}

// GenesisRun runs the dag verifier from genesis block to the latest block for all the projects
func (d *DagVerifier) GenesisRun() {
	// get all the projects
	log.Debug("getting all the projects")

	projects, err := d.redisCache.GetStoredProjects(context.Background())
	if err != nil {
		log.WithError(err).Error("failed to get projects")

		return
	}

	swg := sizedwaitgroup.New(d.settings.DagVerifierSettings.Concurrency)
	for _, projectID := range projects {
		swg.Add()

		go func(projectID string) {
			swg.Add()

			l := log.WithField("projectID", projectID)

			// get last verified dag height from cache
			lastReportedDagHeight := 0

			err = backoff.Retry(func() error {
				lastReportedDagHeight, err = d.redisCache.GetLastReportedDagHeight(context.Background(), projectID)
				if err != nil {
					return err
				}

				return nil
			}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), uint64(*d.settings.RetryCount)))
			if err != nil {
				log.WithError(err).Error("failed to get last reported dag height")

				return
			}

			if lastReportedDagHeight > 0 {
				l.Debug("dag verifier already ran for the project")

				return
			}

			// else add genesis run to cache for the project
			d.genesisRunStatusMap[projectID] = true

			// run till genesis block
			d.Run(&NewBlocksAddedEvent{
				ProjectID:   projectID,
				StartHeight: "-inf",
				EndHeight:   "+inf",
			}, true)
		}(projectID)
	}

	swg.Wait()

	// run status report for cached projects when dag verifier was running
	d.runQueuedProjectEvents(projects)
}

// runQueuedProjectEvents runs the queued events for the projects
func (d *DagVerifier) runQueuedProjectEvents(projects []string) {
	wg := sync.WaitGroup{}

	for _, projectID := range projects {
		wg.Add(1)

		go func(projectID string) {
			defer wg.Done()

			// run queued events one by one for the project
			// mutex is used so that we will always get finalized queued events
			for {
				events, ok := d.eventMessagesMap[projectID]
				if !ok {
					break
				}

				events.mutex.Lock()

				if len(events.messages) == 0 {
					events.mutex.Unlock()

					break
				}

				event := events.messages[0]

				d.Run(event, false)

				events.messages = events.messages[1:]

				events.mutex.Unlock()
			}

			delete(d.genesisRunStatusMap, projectID)
		}(projectID)
	}

	wg.Wait()
}

// checkNullPayload checks if the payload cid is null
func (d *DagVerifier) checkNullPayload(projectID string, chain []*datamodel.DagBlock) {
	l := log.WithField("projectID", projectID)

	if len(chain) == 0 {
		log.Debug("dag chain is empty")

		return
	}

	for _, block := range chain {
		if block.Data == nil {
			continue
		}

		if strings.HasPrefix(block.Data.PayloadLink.Cid, "null") {
			l.Error("payload is null")

			heightString := strconv.Itoa(int(block.Height))

			d.issues[heightString] = append(d.issues[heightString], &datamodel.DagVerifierStatus{
				IssueType:   PayloadMissing,
				Timestamp:   time.Now().UnixMicro(),
				BlockHeight: block.Height,
				Meta: &payloadMissingIssue{
					BlockCID: block.CurrentCid,
				},
			})
		}
	}
}

// checkBlocksOutOfOrder checks if the blocks are out of order
func (d *DagVerifier) checkBlocksOutOfOrder(projectID string, chain []*datamodel.DagBlock) error {
	l := log.WithField("projectID", projectID)

	if len(chain) == 0 {
		return nil
	}
	// check if the blocks are out_of_order/missing/discontinuous
	for index := 0; index < len(chain)-1; index++ {
		currentHeight := chain[index].Height
		nextHeight := chain[index+1].Height

		if currentHeight+1 == nextHeight {
			continue
		}

		l.Infof("blocks out of order at height %d", currentHeight)

		heightString := strconv.Itoa(int(currentHeight))

		d.issues[heightString] = append(d.issues[heightString], &datamodel.DagVerifierStatus{
			Timestamp:   time.Now().UnixMicro(),
			BlockHeight: currentHeight,
			IssueType:   BlocksOutOfOrder,
			Meta: &blocksOutOfOrderIssue{
				CurrentBlockHeight: currentHeight,
				NextBlockHeight:    nextHeight,
				ExpectBlockHeight:  currentHeight + 1,
			},
		})

		// if blocks are not continuous check if there are different entries at same height in cache
		if currentHeight == nextHeight {
			// get the dag blocks at currentHeight & nextHeight heights
			dagBlock1, err := d.getDagBlock(projectID, chain[index].CurrentCid)
			if err != nil {
				l.WithError(err).WithField("cid", chain[index].CurrentCid).Error("error to get dag block from ipfs")

				heightString = strconv.Itoa(int(currentHeight))

				d.issues[heightString] = append(d.issues[heightString], &datamodel.DagVerifierStatus{
					Timestamp:   time.Now().UnixMicro(),
					BlockHeight: currentHeight,
					IssueType:   InternalVerificationErrIssue,
					Meta: &internalVerificationErrIssue{
						Err:         err.Error(),
						Description: fmt.Sprintf("error getting dag block from ipfs for cid %s", chain[index].CurrentCid),
					},
				})

				return err
			}

			dagBlock2, err := d.getDagBlock(projectID, chain[index+1].CurrentCid)
			if err != nil {
				l.WithError(err).WithField("cid", chain[index+1].CurrentCid).Error("error to get dag block from ipfs")

				heightString = strconv.Itoa(int(currentHeight))

				d.issues[heightString] = append(d.issues[heightString], &datamodel.DagVerifierStatus{
					Timestamp:   time.Now().UnixMicro(),
					BlockHeight: currentHeight,
					IssueType:   InternalVerificationErrIssue,
					Meta: &internalVerificationErrIssue{
						Err:         err.Error(),
						Description: fmt.Sprintf("error getting dag block from ipfs for cid %s", chain[index+1].CurrentCid),
					},
				})

				return err
			}

			heightString = strconv.Itoa(int(currentHeight))

			// check if blocks have same data cid
			if dagBlock1.Data.PayloadLink.Cid == dagBlock2.Data.PayloadLink.Cid {
				d.issues[heightString] = append(d.issues[heightString], &datamodel.DagVerifierStatus{
					Timestamp:   time.Now().UnixMicro(),
					BlockHeight: currentHeight,
					IssueType:   DuplicateDagBlock,
					Meta:        nil,
				})
			} else {
				// else multiple blocks at same height
				d.issues[heightString] = append(d.issues[heightString], &datamodel.DagVerifierStatus{
					Timestamp:   time.Now().UnixMicro(),
					BlockHeight: currentHeight,
					IssueType:   MultipleBlocksAtSameHeight,
					Meta: &multipleDagBlocksAtSameHeightIssue{
						CIDs: []string{chain[index].CurrentCid, chain[index+1].CurrentCid},
					},
				})
			}
		}
	}

	return nil
}

// checkEpochsOutOfOrder checks if the epochs are out_of_order/missing/skipped/discontinuous
// this won't apply for summary projects
func (d *DagVerifier) checkEpochsOutOfOrder(projectID string, chain []*datamodel.DagBlock) {
	l := log.WithField("projectID", projectID).WithField("method", "checkBlocksOutOfOrder")

	if len(chain) == 0 {
		return
	}

	// check if the epochs are out_of_order/missing/skipped/discontinuous
	for index := 0; index < len(chain)-1; index++ {
		if chain[index].Data == nil {
			continue
		}

		// check if the payload cid is null
		if strings.HasPrefix(chain[index].Data.PayloadLink.Cid, "null") {
			continue
		}

		// check if the blocks are continuous, then only it makes sense to check for epochs continuity
		if chain[index].Height+1 != chain[index+1].Height {
			continue
		}

		// check if the chain height range is nil
		if chain[index].Payload.ChainHeightRange == nil {
			continue
		}

		currentIpfsBlock := chain[index]
		nextIpfsBlock := chain[index+1]

		// check if the epochs are out of order
		currentBlockChainRangeEnd := currentIpfsBlock.Payload.ChainHeightRange.End
		nextBlockChainRangeStart := nextIpfsBlock.Payload.ChainHeightRange.Begin

		if currentBlockChainRangeEnd+1 != nextBlockChainRangeStart {
			l.WithField("blockHeight", currentIpfsBlock.Height).Info("epochs out of order")

			heightString := strconv.Itoa(int(currentIpfsBlock.Height))

			d.issues[heightString] = append(d.issues[heightString], &datamodel.DagVerifierStatus{
				Timestamp:   time.Now().UnixMicro(),
				BlockHeight: currentIpfsBlock.Height,
				IssueType:   EpochSkipped,
				Meta: &epochSkippedIssue{
					CurrentPayloadChainRange: &DagBlockWithChainRange{
						BlockHeight:      currentIpfsBlock.Height,
						ChainHeightRange: currentIpfsBlock.Payload.ChainHeightRange,
					},
					NextPayloadChainRange: &DagBlockWithChainRange{
						BlockHeight:      nextIpfsBlock.Height,
						ChainHeightRange: nextIpfsBlock.Payload.ChainHeightRange,
					},
					ExpectedChainRange: &datamodel.ChainHeightRange{
						Begin: currentBlockChainRangeEnd + 1,
						End: (currentBlockChainRangeEnd + 1) +
							(currentIpfsBlock.Payload.ChainHeightRange.End - currentIpfsBlock.Payload.ChainHeightRange.Begin), // need some config for epoch size
					},
				},
			})
		}
	}
}

// getPayloadFromDiskCache gets the blocks from disk cache
func (d *DagVerifier) getPayloadFromDiskCache(projectID string, payloadCid string) (*datamodel.DagPayload, error) {
	l := log.WithField("projectID", projectID).WithField("payloadCid", payloadCid)

	payload := new(datamodel.DagPayload)
	fileName := fmt.Sprintf("%s%s/%s.json", d.settings.PayloadCachePath, projectID, payloadCid)

	bytes, err := d.diskCache.Read(fileName)
	if err != nil {
		l.WithError(err).Errorf("failed to fetch payload for given payload cid")

		return payload, err
	}

	err = json.Unmarshal(bytes, payload)
	if err != nil {
		l.WithError(err).Error("failed to Unmarshal json payload from IPFS")

		return payload, err
	}

	return payload, nil
}

// getDagBlockFromDiskCache gets the block from disk cache
func (d *DagVerifier) getDagBlockFromDiskCache(projectID, cid string) (*datamodel.DagBlock, error) {
	l := log.WithField("projectID", projectID).WithField("payloadCid", cid)

	dagBlock := new(datamodel.DagBlock)
	fileName := fmt.Sprintf("%s/%s/%s.json", d.settings.PayloadCachePath, projectID, cid)

	bytes, err := d.diskCache.Read(fileName)
	if err != nil {
		l.WithError(err).Errorf("fdailed to fetch dagBlock for given dagBlock cid")

		return dagBlock, err
	}

	err = json.Unmarshal(bytes, dagBlock)
	if err != nil {
		l.WithError(err).Error("failed to unmarshal json dagBlock from IPFS")

		return dagBlock, err
	}

	return dagBlock, nil
}

func (d *DagVerifier) fillPayloadData(projectID string, dagBlockChain, dagChainWithPayloadCIDs []*datamodel.DagBlock) error {
	l := log.WithField("projectID", projectID)

	swg := sizedwaitgroup.New(d.settings.DagVerifierSettings.Concurrency)

	// getting payload data happens in separate goroutine of each block
	// but when there is error getting payload data for any of the blocks everything should stop
	// as verification can't be done successfully without the payload information
	// to stop go routines errChan is used, when error is sent to errChan, all go routines should stop and error is returned.
	errChan := make(chan error, 2)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	for index, block := range dagBlockChain {
		swg.Add()

		go func(index int, block *datamodel.DagBlock) {
			defer swg.Done()

			select {
			case <-ctx.Done():
				return
			default:
				// non-blocking
			}

			// height should match
			if block.Height != dagChainWithPayloadCIDs[index].Height {
				l.Error("dag cids and payloads cids are out of sync in cache")

				heightString := strconv.Itoa(int(block.Height))

				d.issues[heightString] = append(d.issues[heightString], &datamodel.DagVerifierStatus{
					Timestamp:   time.Now().UnixMicro(),
					BlockHeight: block.Height,
					IssueType:   CidHeightsOutOfSyncInCache,
					Meta: &cidHeightsOutOfSyncIssue{
						ChainHeightRange: &datamodel.DagBlocksHeightRange{
							StartHeight: dagBlockChain[0].Height,
							EndHeight:   dagBlockChain[len(dagBlockChain)-1].Height,
						},
					},
				})

				return
			}

			block.Data = dagChainWithPayloadCIDs[index].Data

			payload, err := d.getPayloadFromDiskCache(projectID, block.Data.PayloadLink.Cid)
			if err != nil {
				l.WithError(err).Error("failed to get payload from disk cache")

				// get payload from ipfs
				payloadCID := dagChainWithPayloadCIDs[index].Data.PayloadLink.Cid

				payload, err = d.ipfsClient.GetPayloadFromIPFS(payloadCID)
				if err != nil {
					l.WithError(err).Error("failed to get payload from ipfs")

					heightString := strconv.Itoa(int(dagBlockChain[index].Height))

					d.issues[heightString] = append(d.issues[heightString], &datamodel.DagVerifierStatus{
						Timestamp:   time.Now().UnixMicro(),
						BlockHeight: dagBlockChain[index].Height,
						IssueType:   InternalVerificationErrIssue,
						Meta: &internalVerificationErrIssue{
							Err:         err.Error(),
							Description: fmt.Sprintf("error getting payload from ipfs for cid %s", payloadCID),
						},
					})

					cancel()

					if len(errChan) != 0 {
						return
					}

					errChan <- err

					return
				}
			}

			block.Payload = payload
		}(index, block)
	}

	swg.Wait()

	if len(errChan) != 0 {
		err := <-errChan

		return err
	}

	return nil
}

// traverseArchivedDagSegments traverses the archived dag segments and verifies the data
func (d *DagVerifier) traverseArchivedDagSegments(projectID string, dagBlockChain []*datamodel.DagBlock) error {
	l := log.WithField("projectID", projectID)
	fullChain := make([]*datamodel.DagBlock, dagBlockChain[len(dagBlockChain)-1].Height)

	firstBlock := dagBlockChain[0]

	if firstBlock.Height == 0 {
		l.Info("dag block height is already 0, no need to traverse")

		return nil
	}

	// get dag block ipfs
	currentBlock, err := d.getDagBlock(projectID, firstBlock.CurrentCid)
	if err != nil {
		l.WithError(err).Error("failed to get dag block from ipfs")

		return err
	}
	currentBlock.CurrentCid = firstBlock.CurrentCid

	// check if block has prevRoot field which is used for get last dag segment
	// as this is the first block of the segment it must have prevRoot field otherwise something went wrong in archival service
	if currentBlock.PrevRoot == "" {
		l.WithField("blockHeight", firstBlock.Height).Error("first block of the segment does not have prevRoot field")

		return errors.New("missing prevRoot field in first block of segment")
	}

	// start traversing
	l.Debug("starting to traverse dag segments")

	for {
		// first block of the dag chain
		if currentBlock.Height == 1 {
			break
		}

		prevBlock := new(datamodel.DagBlock)
		if currentBlock.PrevRoot != "" {
			prevBlock, err = d.getDagBlock(projectID, currentBlock.PrevRoot)
			if err != nil {
				l.WithError(err).Error("failed to get dag block")

				return err
			}
		} else {
			prevBlock, err = d.getDagBlock(projectID, currentBlock.PrevLink.Cid)
			if err != nil {
				l.WithError(err).Error("failed to get dag block")

				return err
			}
		}

		payload, err := d.getPayloadFromDiskCache(projectID, prevBlock.Data.PayloadLink.Cid)
		if err != nil {
			l.WithError(err).Error("failed to get payload from disk cache")

			// get payload from ipfs
			payloadCID := prevBlock.Data.PayloadLink.Cid

			payload, err = d.ipfsClient.GetPayloadFromIPFS(payloadCID)
			if err != nil {
				l.WithError(err).Error("failed to get payload from ipfs")

				return err
			}
		}

		prevBlock.Payload = payload
		fullChain[prevBlock.Height-1] = prevBlock
		currentBlock = prevBlock
	}

	fullChain = append(fullChain, dagBlockChain...)

	return nil
}

func (d *DagVerifier) getDagBlock(projectID, blockCID string) (*datamodel.DagBlock, error) {
	// first check if dag block is in cache
	path := fmt.Sprintf("%s%s/%s.json", d.settings.PayloadCachePath, projectID, blockCID)

	blockData, err := d.diskCache.Read(path)
	if err == nil {
		block := new(datamodel.DagBlock)

		err = json.Unmarshal(blockData, block)
		if err != nil {
			return nil, err
		}

		return block, nil
	}

	// get dag block from ipfs
	block, err := d.ipfsClient.GetDagBlock(blockCID)
	if err == nil {
		return block, nil
	}

	// if not present in ipfs, get dag block from web3 storage
	parsedCID, err := cid.Parse(blockCID)
	if err != nil {
		log.WithError(err).WithField("cid", blockCID).Error("failed to parse cid")

		return nil, err
	}

	_, err = d.w3s.Client.Status(context.Background(), parsedCID)
	if err != nil {
		l := log.WithError(err).WithField("cid", blockCID)
		if strings.Contains(err.Error(), "404") {
			l.Error("dag block not found in web3 storage")
		}

		l.Error("error getting status for dagBlock in web3 storage")

		return nil, err
	}

	resp, err := d.w3s.Client.Get(context.Background(), parsedCID)
	if err != nil {
		log.WithError(err).WithField("cid", blockCID).Error("failed to get dag block from web3 storage")

		return nil, err
	}

	if resp.StatusCode != 200 {
		log.WithField("cid", blockCID).Error("failed to get dag block from web3 storage")

		return nil, errors.New("failed to get dag block from web3 storage")
	}

	f, _, err := resp.Files()
	if err != nil {
		log.WithError(err).WithField("cid", blockCID).Error("failed to get dag block from web3 storage")

		return nil, err
	}

	blockData, err = io.ReadAll(f)
	if err != nil {
		log.WithError(err).WithField("cid", blockCID).Error("failed to read file data of dag block from web3 storage")

		return nil, err
	}

	block = new(datamodel.DagBlock)

	err = json.Unmarshal(blockData, block)
	if err != nil {
		log.WithError(err).WithField("cid", blockCID).Error("failed to unmarshal dag block from web3 storage")

		return nil, err
	}

	return block, nil
}

func (d *DagVerifier) updateStatusReport(projectID string, lastBlockHeight int64) {
	l := log.WithField("projectID", projectID)

	maxHeight := lastBlockHeight
	if len(d.issues) != 0 {
		err := d.redisCache.UpdateDagVerificationStatus(context.Background(), projectID, d.issues)
		if err != nil {
			l.WithError(err).Error("failed to update last verified dag height in cache")
		}

		// check if any issues are internal service errors
		for _, statuses := range d.issues {
			// check if any issues are internal service errors
			for _, status := range statuses {
				if status.IssueType == InternalVerificationErrIssue {
					if maxHeight > status.BlockHeight {
						maxHeight = status.BlockHeight
					}
				}
			}
		}

		err = d.redisCache.UpdateLastReportedDagHeight(context.Background(), projectID, int(maxHeight))
		if err != nil {
			l.WithError(err).Error("failed to update last verified dag height in cache")
		}
	} else {
		// if no issues found, update last verified dag height in cache
		err := d.redisCache.UpdateLastReportedDagHeight(context.Background(), projectID, int(maxHeight))
		if err != nil {
			l.WithError(err).Error("failed to update last verified dag height in cache")
		}
	}

	data, _ := json.Marshal(d.issues)
	// notify on slack
	err := slackutils.NotifySlackWorkflow(string(data), "High", "DagVerifier")
	if err != nil {
		l.WithError(err).Error("failed to notify slack")
	}
}
