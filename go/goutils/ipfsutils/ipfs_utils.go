package ipfsutils

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	"github.com/swagftw/gi"
	"golang.org/x/time/rate"

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

	writeUrl, err := ParseURL(writeUrl)
	if err != nil {
		log.WithError(err).Fatal("failed to parse IPFS write URL: ", writeUrl)
	}

	readUrl := settingsObj.IpfsConfig.ReaderURL

	readUrl, err = ParseURL(readUrl)
	if err != nil {
		log.WithError(err).Fatal("failed to parse IPFS read URL: ", readUrl)
	}

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

type UnsupportedMultiaddrError struct {
	URL string
}

func (e *UnsupportedMultiaddrError) Error() string {
	return fmt.Sprintf("unsupported multiaddr url pattern: %s", e.URL)
}

// ParseURL tries to parse a multiaddr URL, if the url is not multiaddr it tries to parse http url.
func ParseURL(ipfsUrl string) (string, error) {
	parts := make([]string, 0) // [host,port,scheme]

	if multiaddr, err := ma.NewMultiaddr(ipfsUrl); err == nil {
		addrSplits := ma.Split(multiaddr)

		// host and port are required
		if len(addrSplits) < 2 {
			return "", &UnsupportedMultiaddrError{URL: ipfsUrl}
		}

		for index, addr := range addrSplits {
			component, _ := ma.SplitFirst(addr)
			if index == 1 && component.Protocol().Code != ma.P_TCP {
				return "", &UnsupportedMultiaddrError{URL: ipfsUrl}
			}

			// check if scheme is present
			if index == 2 {
				if component.Protocol().Code != ma.P_HTTP && component.Protocol().Code != ma.P_HTTPS {
					return "", &UnsupportedMultiaddrError{URL: ipfsUrl}
				}

				parts = append(parts, component.Protocol().Name)

				continue
			}

			parts = append(parts, component.Value())
		}

		if len(parts) < 2 {
			return "", &UnsupportedMultiaddrError{URL: ipfsUrl}
		}

		// join host and port
		ipfsUrl = net.JoinHostPort(parts[0], parts[1])

		// add scheme if present
		if len(parts) >= 3 {
			ipfsUrl = fmt.Sprintf("%s://%s", parts[2], ipfsUrl)
		} else {
			ipfsUrl = fmt.Sprintf("http://%s", ipfsUrl) // default to http if scheme is not present
		}
	} else {
		// parse http url
		parsedURL, err := url.ParseRequestURI(ipfsUrl)
		if err != nil {
			return "", err
		}

		// check if scheme is http or https
		if !strings.EqualFold(parsedURL.Scheme, "http") && !strings.EqualFold(parsedURL.Scheme, "https") {
			return "", fmt.Errorf("unsupported scheme: %s", parsedURL.Scheme)
		}

		// check if host is present
		if parsedURL.Host == "" {
			return "", errors.New("host is required in url")
		}
	}

	return ipfsUrl, nil
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
