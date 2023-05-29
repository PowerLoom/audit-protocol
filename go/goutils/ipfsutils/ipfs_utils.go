package ipfsutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
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
	readClient             *shell.Shell
	writeClient            *shell.Shell
	readClientRateLimiter  *rate.Limiter
	writeClientRateLimiter *rate.Limiter
}

// InitClient initializes the IPFS client.
func InitClient(settingsObj *settings.SettingsObj) *IpfsClient {
	writeUrl := settingsObj.IpfsConfig.URL
	writeUrl = ParseMultiAddrURL(writeUrl)

	readUrl := settingsObj.IpfsConfig.ReaderURL
	readUrl = ParseMultiAddrURL(readUrl)

	ipfsReadHTTPClient := httpclient.GetIPFSReadHTTPClient(settingsObj)
	ipfsWriteHTTPClient := httpclient.GetIPFSWriteHTTPClient(settingsObj)

	client := new(IpfsClient)
	timeout := time.Duration(settingsObj.IpfsConfig.Timeout * int(time.Second))

	// init read client
	client.readClient = shell.NewShellWithClient(readUrl, ipfsReadHTTPClient)
	client.readClient.SetTimeout(timeout)

	log.Debugf("setting IPFS read client timeout of %f seconds", timeout.Seconds())

	// init write client
	client.writeClient = shell.NewShellWithClient(writeUrl, ipfsWriteHTTPClient)
	client.writeClient.SetTimeout(timeout)

	tps := rate.Limit(10) // 10 TPS
	burst := 10

	writeRateLimiter := settingsObj.IpfsConfig.WriteRateLimiter

	if writeRateLimiter != nil {
		burst = writeRateLimiter.Burst

		if writeRateLimiter.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(writeRateLimiter.RequestsPerSec)
		}
	}

	client.writeClientRateLimiter = rate.NewLimiter(tps, burst)

	log.Infof("rate Limit configured for writer IPFS client at %v TPS with a burst of %d", tps, burst)

	readRateLimiter := settingsObj.IpfsConfig.ReadRateLimiter

	if readRateLimiter != nil {
		burst = readRateLimiter.Burst

		if readRateLimiter.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(readRateLimiter.RequestsPerSec)
		}
	}

	client.readClientRateLimiter = rate.NewLimiter(tps, burst)

	log.Infof("rate Limit configured for reader IPFS client at %v TPS with a burst of %d", tps, burst)

	// check if ipfs connection is successful
	version, commit, err := client.readClient.Version()
	if err != nil {
		log.WithError(err).Fatal("failed to connect to IPFS daemon")
	}

	log.Infof("connected to reader IPFS with version %s and commit %s", version, commit)

	// check if ipfs connection is successful
	version, commit, err = client.writeClient.Version()
	if err != nil {
		log.WithError(err).Fatal("failed to connect to IPFS daemon")
	}

	log.Infof("connected to writer IPFS with version %s and commit %s", version, commit)

	// exit if injection fails
	if err := gi.Inject(client); err != nil {
		log.Fatalln("Failed to inject dependencies", err)
	}

	return client
}

func ParseMultiAddrURL(url string) string {
	if _, err := ma.NewMultiaddr(url); err == nil {
		urlSplits := strings.Split(url, "/")

		// join host and port
		url = net.JoinHostPort(urlSplits[2], urlSplits[4])

		// check if scheme is present
		if len(urlSplits) >= 6 {
			url = fmt.Sprintf("%s://%s", urlSplits[5], url)
		}
	}

	return url
}

func (client *IpfsClient) UploadSnapshotToIPFS(payloadCommit *datamodel.PayloadCommitMessage) error {
	err := client.writeClientRateLimiter.Wait(context.Background())
	if err != nil {
		log.WithError(err).Error("ipfs rate limiter errored")

		return err
	}

	msg, err := json.Marshal(payloadCommit.Message)
	if err != nil {
		log.WithError(err).Error("failed to marshal payload commit message")

		return err
	}

	snapshotCid, err := client.writeClient.Add(bytes.NewReader(msg), shell.CidVersion(1))
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
	err := client.readClientRateLimiter.Wait(context.Background())
	if err != nil {
		log.WithError(err).Error("ipfs rate limiter errored")

		return err
	}

	err = client.readClient.Get(snapshotCID, outputPath)
	if err != nil {
		log.WithError(err).Error("failed to get snapshot message from ipfs")

		return err
	}

	log.WithField("cid", snapshotCID).Debug("successfully fetched snapshot message from ipfs and wrote in local disk")

	return nil
}

func (client *IpfsClient) Unpin(cid string) error {
	err := client.writeClientRateLimiter.Wait(context.Background())
	if err != nil {
		log.WithError(err).Error("ipfs rate limiter errored")

		return err
	}

	err = client.writeClient.Unpin(cid)
	if err != nil {
		return err
	}

	log.WithField("cid", cid).Debug("successfully unpinned cid")

	return nil
}
