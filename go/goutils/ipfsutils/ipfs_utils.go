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
	"github.com/swagftw/gi"
	"golang.org/x/time/rate"

	"audit-protocol/goutils/datamodel"
	"audit-protocol/goutils/httpclient"
	"audit-protocol/goutils/settings"
)

type IpfsClient struct {
	ipfsClient            *shell.Shell
	ipfsClientRateLimiter *rate.Limiter
}

// InitClient initializes the IPFS client.
func InitClient(settingsObj *settings.SettingsObj) *IpfsClient {
	url := settingsObj.IpfsConfig.URL

	url = ParseMultiAddrURL(url)

	ipfsHTTPClient := httpclient.GetIPFSHTTPClient(settingsObj)

	log.Debug("initializing the IPFS client with IPFS Daemon URL:", url)

	client := new(IpfsClient)
	client.ipfsClient = shell.NewShellWithClient(url, ipfsHTTPClient)
	timeout := time.Duration(settingsObj.IpfsConfig.Timeout * int(time.Second))
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

	// exit if injection fails
	if err := gi.Inject(client); err != nil {
		log.Fatalln("Failed to inject dependencies", err)
	}

	// check if ipfs connection is successful
	version, commit, err := client.ipfsClient.Version()
	if err != nil {
		log.WithError(err).Fatal("failed to connect to IPFS daemon")
	}

	log.Infof("connected to IPFS daemon with version %s and commit %s", version, commit)

	return client
}

func ParseMultiAddrURL(url string) string {
	if _, err := ma.NewMultiaddr(url); err == nil {
		url = strings.Split(url, "/")[2] + ":" + strings.Split(url, "/")[4]
	}

	if strings.Contains(url, "localhost") {
		return url
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	return url
}

func (client *IpfsClient) UploadSnapshotToIPFS(payloadCommit *datamodel.PayloadCommitMessage) error {
	err := client.ipfsClientRateLimiter.Wait(context.Background())
	if err != nil {
		log.WithError(err).Error("ipfs rate limiter errored")

		return err
	}

	msg, err := json.Marshal(payloadCommit.Message)
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
		WithField("epochId", payloadCommit.EpochID).
		Debug("ipfs add Successful")

	payloadCommit.SnapshotCID = snapshotCid

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

	log.WithField("cid", snapshotCID).Debug("successfully fetched snapshot message from ipfs and wrote in local disk")

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
