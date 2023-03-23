package verifiers

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"

	"audit-protocol/caching"
	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/ipfsutils"
	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/slackutils"
)

type (
	DagVerificationIssueType string

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
		issues     []interface{}
	}

	blocksOutOfOrderIssue struct {
		IssueType          DagVerificationIssueType `json:"issueType"`
		CurrentBlockHeight int64                    `json:"currentBlockHeight"`
		NextBlockHeight    int64                    `json:"nextBlockHeight"`
		ExpectBlockHeight  int64                    `json:"expectBlockHeight"`
	}

	cidsOutOfSyncIssue struct {
		IssueType DagVerificationIssueType `json:"issueType"`
	}

	duplicateDagBlocksIssue struct {
		IssueType   DagVerificationIssueType `json:"issueType"`
		BlockHeight int64                    `json:"height"`
	}

	multipleDagBlocksAtSameHeightIssue struct {
		IssueType   DagVerificationIssueType `json:"issueType"`
		BlockHeight int64                    `json:"height"`
		CIDs        []string                 `json:"cids"`
	}

	DagBlockWithChainRange struct {
		BlockHeight int64                       `json:"blockHeight"`
		ChainRange  *datamodel.ChainHeightRange `json:"chainRange"`
	}

	epochSkippedIssue struct {
		IssueType                DagVerificationIssueType    `json:"issueType"`
		CurrentPayloadChainRange *DagBlockWithChainRange     `json:"currentPayloadChainRange"`
		NextPayloadChainRange    *DagBlockWithChainRange     `json:"nextPayloadChainRange"`
		ExpectedChainRange       *datamodel.ChainHeightRange `json:"expectedChainRange"`
	}

	payloadMissingIssue struct {
		IssueType   DagVerificationIssueType `json:"issueType"`
		BlockCID    string                   `json:"cid"`
		BlockHeight int64                    `json:"blockHeight"`
	}

	internalVerificationErrIssue struct {
		IssueType   DagVerificationIssueType `json:"issueType"`
		Err         string                   `json:"err"`
		Description string                   `json:"description"`
		Method      string                   `json:"method"`
	}
)

const (
	BlocksOutOfOrder             DagVerificationIssueType = "BLOCKS_OUT_OF_ORDER"
	EpochSkipped                 DagVerificationIssueType = "EPOCH_SKIPPED"
	PayloadMissing               DagVerificationIssueType = "PAYLOAD_MISSING"
	DuplicateDagBlock            DagVerificationIssueType = "DUPLICATE_DAG_BLOCK"
	PayloadCorrupted             DagVerificationIssueType = "PAYLOAD_CORRUPTED"
	CidsOutOfSyncInCache         DagVerificationIssueType = "CIDS_OUT_OF_SYNC_IN_CACHE"
	MultipleBlocksAtSameHeight   DagVerificationIssueType = "MULTIPLE_BLOCKS_AT_SAME_HEIGHT"
	InternalVerificationErrIssue DagVerificationIssueType = "INTERNAL_VERIFICATION_ERR"
	NoIssues                     DagVerificationIssueType = "NO_ISSUES"
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

	return &DagVerifier{
		redisCache: redisCache,
		diskCache:  diskCache,
		settings:   settingsObj,
		ipfsClient: ipfsClient,
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

		d.issues = append(d.issues, &cidsOutOfSyncIssue{
			IssueType: CidsOutOfSyncInCache,
		})

		return
	}

	var wg sync.WaitGroup
	for index, block := range dagBlockChain {
		wg.Add(1)

		go func(index int, block *datamodel.DagBlock) {
			defer wg.Done()
			// height should match
			if block.Height != dagChainWithPayloadCIDs[index].Height {
				l.Error("dag cids and payloads cids are out of sync in cache for given height range")

				d.issues = append(d.issues, &cidsOutOfSyncIssue{
					IssueType: CidsOutOfSyncInCache,
				})

				return
			}

			block.Data.PayloadLink.Cid = dagChainWithPayloadCIDs[index].Data.PayloadLink.Cid

			payload, err := d.getPayloadFromDiskCache(projectID, block.Data.PayloadLink.Cid)
			if err != nil {
				l.WithError(err).Error("failed to get payload from disk cache")

				// get payload from ipfs
				payload, err = d.ipfsClient.GetPayloadFromIPFS(dagChainWithPayloadCIDs[index].Data.PayloadLink.Cid)
				if err != nil {
					l.WithError(err).Error("failed to get payload from ipfs")

					return
				}
			}

			block.Payload.ChainHeightRange = payload.ChainHeightRange
		}(index, block)
	}

	wg.Wait()

	// check for null payload
	err = d.checkNullPayload(projectID, dagBlockChain)
	if err != nil {
		return
	}

	// check for blocks out of order
	err = d.checkBlocksOutOfOrder(projectID, dagBlockChain)
	if err != nil {
		return
	}

	// check for epoch skipped
	err = d.checkEpochsOutOfOrder(projectID, dagBlockChain)
	if err != nil {
		return
	}

	defer func(err error) {
		verStatus := &datamodel.DagVerifierStatus{
			EventPayload: blocksInserted,
			Timestamp:    time.Now().UnixMicro(),
			NoIssues:     true,
			DagBlocksHeightRange: &datamodel.DagBlocksHeightRange{
				StartHeight: dagBlockChain[0].Height,
				EndHeight:   dagBlockChain[len(dagBlockChain)-1].Height,
			},
			Issues: make([]interface{}, 0), // empty array rather than null
		}

		if err != nil {
			l.WithError(err).Error("failed to verify dag chain, internal error")
			verStatus.NoIssues = false
			verStatus.Issues = d.issues
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

// checkNullPayload checks if the payload cid is null
func (d *DagVerifier) checkNullPayload(projectID string, chain []*datamodel.DagBlock) error {
	l := log.WithField("projectID", projectID)

	if len(chain) == 0 {
		return nil
	}

	for _, block := range chain {
		if strings.HasPrefix(block.Data.PayloadLink.Cid, "null") {
			l.Error("payload is null")

			d.issues = append(d.issues, &payloadMissingIssue{
				IssueType:   PayloadMissing,
				BlockCID:    block.CurrentCid,
				BlockHeight: block.Height,
			})
		}
	}

	return nil
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

		d.issues = append(d.issues, &blocksOutOfOrderIssue{
			IssueType:          BlocksOutOfOrder,
			CurrentBlockHeight: currentHeight,
			NextBlockHeight:    nextHeight,
			ExpectBlockHeight:  currentHeight + 1,
		})

		// if blocks are not continuous check if there are different entries at same height in cache
		if currentHeight == nextHeight {
			// get the dag blocks at these heights
			dagBlock1, err := d.ipfsClient.GetDagBlock(chain[index].CurrentCid)
			if err != nil {
				l.WithError(err).WithField("cid", chain[index].CurrentCid).Error("error to get dag block from ipfs")

				d.issues = append(d.issues, &internalVerificationErrIssue{
					IssueType:   InternalVerificationErrIssue,
					Err:         err.Error(),
					Description: fmt.Sprintf("error getting dag block from ipfs for cid %s", chain[index].CurrentCid),
				})
				return err
			}

			dagBlock2, err := d.ipfsClient.GetDagBlock(chain[index+1].CurrentCid)
			if err != nil {
				l.WithError(err).WithField("cid", chain[index+1].CurrentCid).Error("error to get dag block from ipfs")

				d.issues = append(d.issues, &internalVerificationErrIssue{
					IssueType:   InternalVerificationErrIssue,
					Err:         err.Error(),
					Method:      "checkBlocksOutOfOrder",
					Description: fmt.Sprintf("error getting dag block from ipfs for cid %s", chain[index+1].CurrentCid),
				})

				return err
			}

			// check if blocks have same data cid
			if dagBlock1.Data.PayloadLink.Cid == dagBlock2.Data.PayloadLink.Cid {
				d.issues = append(d.issues, &duplicateDagBlocksIssue{
					IssueType:   DuplicateDagBlock,
					BlockHeight: currentHeight,
				})
			} else {
				// else multiple blocks at same height
				d.issues = append(d.issues, &multipleDagBlocksAtSameHeightIssue{
					IssueType:   MultipleBlocksAtSameHeight,
					BlockHeight: currentHeight,
					CIDs:        []string{chain[index].CurrentCid, chain[index+1].CurrentCid},
				})
			}
		}
	}

	return nil
}

// checkEpochsOutOfOrder checks if the epochs are out_of_order/missing/skipped/discontinuous
// this won't apply for summary projects
func (d *DagVerifier) checkEpochsOutOfOrder(projectID string, chain []*datamodel.DagBlock) error {
	l := log.WithField("projectID", projectID).WithField("method", "checkBlocksOutOfOrder")

	if len(chain) == 0 {
		return nil
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

		currentIpfsBlock := chain[index]
		nextIpfsBlock := chain[index+1]

		// check if the epochs are out of order
		currentBlockChainRangeEnd := currentIpfsBlock.Payload.ChainHeightRange.End
		nextBlockChainRangeStart := nextIpfsBlock.Payload.ChainHeightRange.Begin

		if currentBlockChainRangeEnd+1 != nextBlockChainRangeStart {
			l.WithField("blockHeight", currentIpfsBlock.Height).Info("epochs out of order")

			d.issues = append(d.issues, &epochSkippedIssue{
				IssueType: EpochSkipped,
				CurrentPayloadChainRange: &DagBlockWithChainRange{
					BlockHeight: currentIpfsBlock.Height,
					ChainRange:  currentIpfsBlock.Payload.ChainHeightRange,
				},
				NextPayloadChainRange: &DagBlockWithChainRange{
					BlockHeight: nextIpfsBlock.Height,
					ChainRange:  nextIpfsBlock.Payload.ChainHeightRange,
				},
				ExpectedChainRange: &datamodel.ChainHeightRange{
					Begin: currentBlockChainRangeEnd + 1,
					End: (currentBlockChainRangeEnd + 1) +
						(currentIpfsBlock.Payload.ChainHeightRange.End - currentIpfsBlock.Payload.ChainHeightRange.Begin), // need some config for epoch size
				},
			})
		}
	}

	return nil
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
	fileName := fmt.Sprintf("%s/%s/%s.json", d.settings.PayloadCachePath, projectID, payloadCid)

	bytes, err := d.diskCache.Read(fileName)
	if err != nil {
		l.WithError(err).Errorf("fdailed to fetch payload for given payload cid")

		return payload, err
	}

	err = json.Unmarshal(bytes, payload)
	if err != nil {
		l.WithError(err).Error("failed to Unmarshal json payload from IPFS")

		return payload, err
	}

	return payload, nil
}
