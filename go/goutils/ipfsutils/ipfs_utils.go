package ipfsutils

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/settings"
)

type IpfsClient struct {
	ipfsClient            *shell.Shell
	ipfsClientRateLimiter *rate.Limiter
}

// InitClient initializes the IPFS client.
// Init functions should not be treated as methods, they are just functions that return a pointer to the struct.
func InitClient(url string, poolSize int, rateLimiter *settings.RateLimiter, timeoutSecs int) *IpfsClient {
	// no need to use underscore for _url, it is just a local variable
	// _url := url
	url = ParseMultiAddrUrl(url)

	transport := http.Transport{
		MaxIdleConns:        poolSize,
		MaxConnsPerHost:     poolSize,
		MaxIdleConnsPerHost: poolSize,
		IdleConnTimeout:     0,
		DisableCompression:  true,
	}

	ipfsHTTPClient := http.Client{
		Timeout:   time.Duration(timeoutSecs * int(time.Second)),
		Transport: &transport,
	}

	log.Debug("Initializing the IPFS client with IPFS Daemon URL:", url)

	client := new(IpfsClient)
	client.ipfsClient = shell.NewShellWithClient(url, &ipfsHTTPClient)
	timeout := time.Duration(timeoutSecs * int(time.Second))
	client.ipfsClient.SetTimeout(timeout)

	log.Debugf("Setting IPFS timeout of %f seconds", timeout.Seconds())

	tps := rate.Limit(10) // 10 TPS
	burst := 10

	if rateLimiter != nil {
		burst = rateLimiter.Burst

		if rateLimiter.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(rateLimiter.RequestsPerSec)
		}
	}

	log.Infof("Rate Limit configured for IPFS Client at %v TPS with a burst of %d", tps, burst)
	client.ipfsClientRateLimiter = rate.NewLimiter(tps, burst)

	return client
}

func ParseMultiAddrUrl(url string) string {
	if _, err := ma.NewMultiaddr(url); err == nil {
		url = strings.Split(url, "/")[2] + ":" + strings.Split(url, "/")[4]
	}

	return url
}

func (client *IpfsClient) GetDagBlock(dagCid string) (datamodel.DagBlock, error) {
	var dagBlock datamodel.DagBlock
	var err error
	var i int
	for i = 0; i < 3; i++ {

		err = client.ipfsClientRateLimiter.Wait(context.Background())
		if err != nil {
			log.Errorf("IPFSClient Rate Limiter wait timeout with error %+v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		err = client.ipfsClient.DagGet(dagCid, &dagBlock)
		if err != nil {
			log.Error("Error in getting Dag block from IPFS, Dag-CID:", dagCid, ", error:", err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
	if i >= 3 {
		log.Errorf("Failed to fetch CID %s even after retrying.", dagCid)
		return dagBlock, err
	}
	log.Tracef("Fetched the dag Block with CID %s, BlockInfo:%+v", dagCid, dagBlock)
	return dagBlock, nil
}

func (client *IpfsClient) GetPayloadChainHeightRang(
	payloadCid string,
	retryCount int) (datamodel.DagPayloadChainHeightRange, error) {
	var payload datamodel.DagPayloadChainHeightRange
	var err error
	var i int
	for i = 0; i < retryCount; i++ {
		err := client.ipfsClientRateLimiter.Wait(context.Background())
		if err != nil {
			log.Errorf("IPFSClient Rate Limiter wait timeout with error %+v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		data, err := client.ipfsClient.Cat(payloadCid)
		if err != nil {
			log.Error("Failed to fetch Payload from IPFS, CID:", payloadCid, ", error:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(data)

		err = json.Unmarshal(buf.Bytes(), &payload)
		if err != nil {
			log.Error("Failed to Unmarshal Json Payload from IPFS, CID:", payloadCid, ", bytes:", buf, ", error:", err)
			return payload, err
		}
		break
	}
	if i >= retryCount {
		log.Error("Failed to fetch even after retrying.")
		return payload, err
	}
	log.Tracef("Fetched Payload with CID %s from IPFS: %+v", payloadCid, payload)
	return payload, nil
}

func (client *IpfsClient) UploadSnapshotToIPFS(payloadCommit *datamodel.PayloadCommit,
	retryIntervalSecs int, retryCountConfig int) bool {

	for retryCount := 0; ; {

		err := client.ipfsClientRateLimiter.Wait(context.Background())
		if err != nil {
			log.Errorf("IPFSClient Rate Limiter wait timeout with error %+v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		snapshotCid, err := client.ipfsClient.Add(bytes.NewReader(payloadCommit.Payload), shell.CidVersion(1))

		if err != nil {
			if retryCount == retryCountConfig {
				log.Errorf("IPFS Add failed for message %+v after max-retry of %d, with err %v", payloadCommit, retryCount, err)
				return false
			}
			time.Sleep(time.Duration(retryIntervalSecs) * time.Second)
			retryCount++
			log.Errorf("IPFS Add failed for message %v, with err %v..retryCount %d .", payloadCommit, err, retryCount)
			continue
		}
		log.Debugf("IPFS add Successful. Snapshot CID is %s for project %s with commitId %s at tentativeBlockHeight %d",
			snapshotCid, payloadCommit.ProjectId, payloadCommit.CommitId, payloadCommit.TentativeBlockHeight)
		payloadCommit.SnapshotCID = snapshotCid
		break
	}
	return true
}

func (client *IpfsClient) GetPayloadFromIPFS(payloadCid string, retryIntervalSecs int, retryCount int) (*datamodel.PayloadData, error) {
	var payload datamodel.PayloadData
	for i := 0; ; {
		log.Debugf("Fetching payloadCid %s from IPFS", payloadCid)
		data, err := client.ipfsClient.Cat(payloadCid)
		if err != nil {
			if i >= retryCount {
				log.Errorf("Failed to fetch Payload with CID %s from IPFS even after max retries due to error %+v.", payloadCid, err)
				return &payload, err
			}
			log.Warnf("Failed to fetch Payload from IPFS, CID %s due to error %+v", payloadCid, err)
			time.Sleep(time.Duration(retryIntervalSecs) * time.Second)
			i++
			continue
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(data)

		err = json.Unmarshal(buf.Bytes(), &payload)
		if err != nil {
			log.Errorf("Failed to Unmarshal Json Payload from IPFS, CID %s, bytes: %+v due to error %+v ",
				payloadCid, buf, err)
			return nil, err
		}
		break
	}

	log.Debugf("Fetched Payload with CID %s from IPFS: %+v", payloadCid, payload)
	return &payload, nil
}

func (client *IpfsClient) UnPinCidsFromIPFS(projectID string, cids map[int]string) error {
	for height, cid := range cids {
		err := client.ipfsClientRateLimiter.Wait(context.Background())
		if err != nil {
			log.Warnf("IPFSClient Rate Limiter wait timeout with error %+v", err)

			continue
		}

		log.Debugf("Unpinning CID %s at height %d from IPFS for project %s", cid, height, projectID)

		err = client.ipfsClient.Unpin(cid)
		if err != nil {
			// CID has already been unpinned.
			if err.Error() == "pin/rm: not pinned or pinned indirectly" || err.Error() == "pin/rm: pin is not part of the pinset" {
				log.Debugf("CID %s for project %s at height %d could not be unpinned from IPFS as it was not pinned on the IPFS node.", cid, projectID, height)

				continue
			}

			log.Warnf("Failed to unpin CID %s from ipfs for project %s at height %d due to error %+v", cid, projectID, height, err)

			return err
		}

		log.Debugf("Unpinned CID %s at height %d from IPFS successfully for project %s", cid, height, projectID)
	}

	return nil
}

// testing and simulation methods

// AddFileToIPFS adds a file to IPFS and returns the CID of the file
func (client *IpfsClient) AddFileToIPFS(data []byte) (string, error) {
	cid, err := client.ipfsClient.Add(bytes.NewReader(data), shell.CidVersion(1))
	if err != nil {
		return "", err
	}

	return cid, nil
}
