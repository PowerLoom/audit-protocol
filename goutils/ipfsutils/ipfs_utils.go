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
	"github.com/powerloom/audit-prototol-private/goutils/datamodel"
	"github.com/powerloom/audit-prototol-private/goutils/settings"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type IpfsClient struct {
	ipfsClient            *shell.Shell
	ipfsClientRateLimiter *rate.Limiter
}

func (client *IpfsClient) Init(url string, poolSize int, rateLimiter *settings.RateLimiter_, timeoutSecs int) {
	_url := url
	_, err := ma.NewMultiaddr(_url)
	if err == nil {
		// Convert the URL from /ip4/<IPAddress>/tcp/<Port> to IP:Port format.
		url = strings.Split(_url, "/")[2] + ":" + strings.Split(_url, "/")[4]
	}

	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        poolSize,
		MaxConnsPerHost:     poolSize,
		MaxIdleConnsPerHost: poolSize,
		IdleConnTimeout:     0,
		DisableCompression:  true,
	}

	ipfsHttpClient := http.Client{
		Timeout:   time.Duration(timeoutSecs * int(time.Second)),
		Transport: &t,
	}
	log.Debug("Initializing the IPFS client with IPFS Daemon URL:", url)
	client.ipfsClient = shell.NewShellWithClient(url, &ipfsHttpClient)
	timeout := time.Duration(timeoutSecs * int(time.Second))
	client.ipfsClient.SetTimeout(timeout)
	log.Debugf("Setting IPFS timeout of %d seconds", timeout.Seconds())
	tps := rate.Limit(10) //10 TPS
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

func (client *IpfsClient) UnPinCidsFromIPFS(projectId string, cids *map[int]string) int {
	errorCount := 0
	for height, cid := range *cids {
		i := 0
		for ; i < 3; i++ {
			err := client.ipfsClientRateLimiter.Wait(context.Background())
			if err != nil {
				log.Warnf("IPFSClient Rate Limiter wait timeout with error %+v", err)
				time.Sleep(1 * time.Second)
				continue
			}
			log.Debugf("Unpinning CID %s at height %d from IPFS for project %s", cid, height, projectId)
			err = client.ipfsClient.Unpin(cid)
			if err != nil {
				//CID has already been unpinned
				if err.Error() == "pin/rm: not pinned or pinned indirectly" || err.Error() == "pin/rm: pin is not part of the pinset" {
					log.Debugf("CID %s for project %s at height %d could not be unpinned from IPFS as it was not pinned on the IPFS node.", cid, projectId, height)
					break
				}
				log.Warnf("Failed to unpin CID %s from ipfs for project %s at height %d due to error %+v. Retrying %d", cid, projectId, height, err, i)
				time.Sleep(5 * time.Second)
				continue
			}
			log.Debugf("Unpinned CID %s at height %d from IPFS successfully for project %s", cid, height, projectId)
			break
		}
		if i == 3 {
			log.Errorf("Failed to unpin CID %s at height %d from ipfs for project %s after max retries", cid, height, projectId)
			errorCount++
			continue
		}
	}
	return errorCount
}
