package ipfsutils

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/settings"
)

type Service interface {
	UploadSnapshotToIPFS(snapshot *datamodel.PayloadCommitMessage) error
	GetSnapshotFromIPFS(snapshotCID string, outputPath string) error
	Unpin(cid string) error
}

type IpfsClient struct {
	ipfsClient            *shell.Shell
	ipfsClientRateLimiter *rate.Limiter
}

// InitService initializes the IPFS client.
func InitService(url string, settingsObj *settings.SettingsObj, timeoutSecs int) Service {
	url = ParseMultiAddrURL(url)

	ipfsHTTPClient := httpclient.GetDefaultHTTPClient(settingsObj)

	log.Debug("initializing the IPFS client with IPFS Daemon URL:", url)

	client := new(IpfsClient)
	client.ipfsClient = shell.NewShellWithClient(url, ipfsHTTPClient.HTTPClient)
	timeout := time.Duration(timeoutSecs * int(time.Second))
	client.ipfsClient.SetTimeout(timeout)

	log.Debugf("setting IPFS timeout of %f seconds", timeout.Seconds())

	tps := rate.Limit(10) // 10 TPS
	burst := 10

	rateLimiter := settingsObj.IpfsConfig.IPFSRateLimiter
	if rateLimiter != nil {
		burst = rateLimiter.Burst

		if rateLimiter.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(rateLimiter.RequestsPerSec)
		}
	}

	log.Infof("rate Limit configured for IPFS Client at %v TPS with a burst of %d", tps, burst)
	client.ipfsClientRateLimiter = rate.NewLimiter(tps, burst)

	return client
}

func ParseMultiAddrURL(url string) string {
	if _, err := ma.NewMultiaddr(url); err == nil {
		url = strings.Split(url, "/")[2] + ":" + strings.Split(url, "/")[4]
	}

	return url
}

func (client *IpfsClient) UploadSnapshotToIPFS(message *datamodel.PayloadCommitMessage) error {
	err := client.ipfsClientRateLimiter.Wait(context.Background())
	if err != nil {
		log.WithError(err).Error("ipfs rate limiter errored")

		return err
	}

	msg, err := json.Marshal(message.Message)
	if err != nil {
		log.WithError(err).Error("failed to marshal payload commit message")

		return err
	}

	snapshotCid, err := client.ipfsClient.Add(bytes.NewReader(msg), shell.CidVersion(1))
	if err != nil {
		log.WithError(err).Error("failed to add snapshot to ipfs")

		return err
	}

	log.WithField("snapshotCID", snapshotCid).
		WithField("epochId", message.EpochID).
		Debug("ipfs add Successful")

	message.SnapshotCID = snapshotCid

	return nil
}

// GetSnapshotFromIPFS returns the snapshot from IPFS.
func (client *IpfsClient) GetSnapshotFromIPFS(snapshotCID string, outputPath string) error {
	err := client.ipfsClientRateLimiter.Wait(context.Background())
	if err != nil {
		log.WithError(err).Error("ipfs rate limiter errored")

		return err
	}

	err = client.ipfsClient.Get(snapshotCID, outputPath)
	if err != nil {
		log.WithError(err).Error("failed to get snapshot message from ipfs")

		return err
	}

	log.Debug("successfully fetched snapshot message from ipfs and wrote in local disk")

	return nil
}

func (client *IpfsClient) Unpin(cid string) error {
	err := client.ipfsClientRateLimiter.Wait(context.Background())
	if err != nil {
		log.WithError(err).Error("ipfs rate limiter errored")

		return err
	}

	err = client.ipfsClient.Unpin(cid)
	if err != nil {
		return err
	}

	log.WithField("cid", cid).Debug("successfully unpinned cid")

	return nil
}
