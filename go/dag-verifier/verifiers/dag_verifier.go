package verifiers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

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
	// cidAndHeight is a struct that holds the CID and height of a dag block
	cidAndHeight struct {
		CID    string `json:"cid"`
		Height int64  `json:"height"`
	}

	DagVerifier struct {
		redisCache *caching.RedisCache
		diskCache  *caching.LocalDiskCache
		settings   *settings.SettingsObj
		ipfsClient *ipfsutils.IpfsClient
		issues     []*datamodel.DagVerifierStatus
		w3s        *w3storage.W3S
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
		BlockHeight int64    `json:"height"`
		CIDs        []string `json:"cids"`
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
	NoIssues                      datamodel.DagVerificationIssueType = "NO_ISSUES"
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

	return &DagVerifier{
		redisCache: redisCache,
		diskCache:  diskCache,
		settings:   settingsObj,
		ipfsClient: ipfsClient,
		w3s:        w3storageClient,
	}
}

// Run starts the dag verifier
// this is called on event when new blocks are inserted in chain
func (d *DagVerifier) Run(blocksInserted *datamodel.DagBlocksInsertedReq) {
	l := log.WithField("req", blocksInserted)

	l.Debug("dag verifier started")

	projectID := blocksInserted.ProjectID
	sortedCIDs := sortCIDsByHeight(blocksInserted.DagHeightCIDMap)

	// get last verified dag height from cache
	height, err := d.redisCache.GetLastVerifiedDagHeight(context.Background(), projectID)
	if err != nil {
		log.WithError(err).Error("failed to get last verification height for project")

		return
	}

	startHeight := strconv.Itoa(height)
	endHeight := strconv.Itoa(int(sortedCIDs[len(sortedCIDs)-1].Height))

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

	if len(dagChainWithPayloadCIDs) != len(dagBlockChain) {
		// this is unlikely to happen but needs to be verified
		l.Error("dag cids and payloads cids are out of sync in cache for given height range")

		d.issues = append(d.issues, &datamodel.DagVerifierStatus{
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
	d.fillPayloadData(projectID, dagBlockChain, dagChainWithPayloadCIDs)

	// check for null payload
	d.checkNullPayload(projectID, dagBlockChain)

	// check for blocks out of order
	err = d.checkBlocksOutOfOrder(projectID, dagBlockChain)
	if err != nil {
		return
	}

	// check for epoch skipped
	d.checkEpochsOutOfOrder(projectID, dagBlockChain)

	defer func(err error) {
		verStatus := &datamodel.DagVerifierStatus{
			Timestamp: time.Now().UnixMicro(),
		}

		if err != nil {
			l.WithError(err).Error("failed to verify dag chain, internal error")
		}

		// update last verified dag height in cache
		err = d.redisCache.UpdateDagVerificationStatus(context.Background(), projectID, verStatus)
		if err != nil {
			l.WithError(err).Error("failed to update last verified dag height in cache")
		}

		data, _ := json.Marshal(verStatus)
		// notify on slack
		err = slackutils.NotifySlackWorkflow(string(data), "High", "DagVerifier")
		if err != nil {
			l.WithError(err).Error("failed to notify slack")
		}
	}(err)
}

// OneTimeRun runs the dag verifier from genesis block to the latest block for all the projects
func (d *DagVerifier) OneTimeRun() {
	log.Debug("running one time run for dag verifier")

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

			dagBlockChain, err := d.redisCache.GetDagChainCIDs(context.Background(), projectID, "-inf", "+inf")
			if err != nil {
				l.WithError(err).Error("failed to get dag chain cids for verification")

				return
			}

			dagChainWithPayloadCIDs, err := d.redisCache.GetPayloadCIDs(context.Background(), projectID, "-inf", "+inf")
			if err != nil {
				l.WithError(err).Error("failed to get payload cids")

				return
			}

			if len(dagChainWithPayloadCIDs) != len(dagBlockChain) {
				l.WithError(err).Error("dag cids and payloads cids are out of sync in cache for given height range, stopping verification")

				d.issues = append(d.issues, &datamodel.DagVerifierStatus{
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
			d.fillPayloadData(projectID, dagBlockChain, dagChainWithPayloadCIDs)

			// fills up the dag chain with archived blocks data
			err = d.traverseArchivedDagSegments(projectID, dagBlockChain)
			if err != nil {
				l.WithError(err).Error("failed to traverse archived dag segments")

				return
			}

			// check for null payload
			d.checkNullPayload(projectID, dagBlockChain)

			// check for blocks out of order
			err = d.checkBlocksOutOfOrder(projectID, dagBlockChain)
			if err != nil {
				return
			}

			// check for epoch skipped
			d.checkEpochsOutOfOrder(projectID, dagBlockChain)

			defer func(err error) {
				verStatus := &datamodel.DagVerifierStatus{
					Timestamp: time.Now().UnixMicro(),
				}

				if err != nil {
					l.WithError(err).Error("failed to verify dag chain, internal error")
				}

				// update last verified dag height in cache
				err = d.redisCache.UpdateDagVerificationStatus(context.Background(), projectID, verStatus)
				if err != nil {
					l.WithError(err).Error("failed to update last verified dag height in cache")
				}

				data, _ := json.Marshal(verStatus)
				// notify on slack
				err = slackutils.NotifySlackWorkflow(string(data), "High", "DagVerifier")
				if err != nil {
					l.WithError(err).Error("failed to notify slack")
				}
			}(err)
		}(projectID)
	}

	swg.Wait()
}

// checkNullPayload checks if the payload cid is null
func (d *DagVerifier) checkNullPayload(projectID string, chain []*datamodel.DagBlock) {
	l := log.WithField("projectID", projectID)

	if len(chain) == 0 {
		log.Debug("dag chain is empty")

		return
	}

	for _, block := range chain {
		if strings.HasPrefix(block.Data.PayloadLink.Cid, "null") {
			l.Error("payload is null")

			d.issues = append(d.issues, &datamodel.DagVerifierStatus{
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

		d.issues = append(d.issues, &datamodel.DagVerifierStatus{
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
			// get the dag blocks at these heights
			dagBlock1, err := d.getDagBlock(projectID, chain[index].CurrentCid)
			if err != nil {
				l.WithError(err).WithField("cid", chain[index].CurrentCid).Error("error to get dag block from ipfs")

				d.issues = append(d.issues, &datamodel.DagVerifierStatus{
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

				d.issues = append(d.issues, &datamodel.DagVerifierStatus{
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

			// check if blocks have same data cid
			if dagBlock1.Data.PayloadLink.Cid == dagBlock2.Data.PayloadLink.Cid {
				d.issues = append(d.issues, &datamodel.DagVerifierStatus{
					Timestamp:   time.Now().UnixMicro(),
					BlockHeight: currentHeight,
					IssueType:   DuplicateDagBlock,
					Meta:        nil,
				})
			} else {
				// else multiple blocks at same height
				d.issues = append(d.issues, &datamodel.DagVerifierStatus{
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

			d.issues = append(d.issues, &datamodel.DagVerifierStatus{
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

	return
}

// ================================
// utility functions
// ================================

// sortCIDsByHeight sorts the given map of cid to height and returns sorted blocks by height
func sortCIDsByHeight(cidToHeightMap map[string]int64) []*cidAndHeight {
	cidAndHeightList := make([]*cidAndHeight, 0, len(cidToHeightMap))

	for cid, height := range cidToHeightMap {
		cidAndHeightList = append(cidAndHeightList, &cidAndHeight{
			CID:    cid,
			Height: height,
		})
	}

	sort.Slice(cidAndHeightList, func(i, j int) bool {
		return cidAndHeightList[i].Height < cidAndHeightList[j].Height
	})

	return cidAndHeightList
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

func (d *DagVerifier) fillPayloadData(projectID string, dagBlockChain, dagChainWithPayloadCIDs []*datamodel.DagBlock) {
	l := log.WithField("projectID", projectID)

	swg := sizedwaitgroup.New(d.settings.DagVerifierSettings.Concurrency)

	for index, block := range dagBlockChain {
		swg.Add()

		go func(index int, block *datamodel.DagBlock) {
			defer swg.Done()
			// height should match
			if block.Height != dagChainWithPayloadCIDs[index].Height {
				l.Error("dag cids and payloads cids are out of sync in cache")

				d.issues = append(d.issues, &datamodel.DagVerifierStatus{
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

					d.issues = append(d.issues, &datamodel.DagVerifierStatus{
						Timestamp:   time.Now().UnixMicro(),
						BlockHeight: dagBlockChain[index].Height,
						IssueType:   InternalVerificationErrIssue,
						Meta: &internalVerificationErrIssue{
							Err:         err.Error(),
							Description: fmt.Sprintf("error getting payload from ipfs for cid %s", payloadCID),
						},
					})

					return
				}
			}

			block.Payload = payload
		}(index, block)
	}

	swg.Wait()
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
