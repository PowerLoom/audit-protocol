package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/powerloom/goutils/settings"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type IpfsClient struct {
	ipfsClient            *shell.Shell
	ipfsClientRateLimiter *rate.Limiter
}

func (client *IpfsClient) Init(settingsObj *settings.SettingsObj) {
	_url := settingsObj.IpfsConfig.ReaderURL
	if _url == "" {
		_url = settingsObj.IpfsConfig.URL
	}
	url := ""
	_, err := ma.NewMultiaddr(_url)
	if err == nil {
		// Convert the URL from /ip4/<IPAddress>/tcp/<Port> to IP:Port format.
		url = strings.Split(_url, "/")[2] + ":" + strings.Split(_url, "/")[4]
	}

	t := http.Transport{
		//TLSClientConfig:    &tls.Config{KeyLogWriter: kl, InsecureSkipVerify: true},
		MaxIdleConns:        settingsObj.DagVerifierSettings.Concurrency,
		MaxConnsPerHost:     settingsObj.DagVerifierSettings.Concurrency,
		MaxIdleConnsPerHost: settingsObj.DagVerifierSettings.Concurrency,
		IdleConnTimeout:     0,
		DisableCompression:  true,
	}

	ipfsHttpClient := http.Client{
		Timeout:   time.Duration(5 * time.Minute),
		Transport: &t,
	}
	log.Debug("Initializing the IPFS client with IPFS Daemon URL:", url)
	client.ipfsClient = shell.NewShellWithClient(url, &ipfsHttpClient)
	timeout := time.Duration(5 * time.Minute)
	client.ipfsClient.SetTimeout(timeout)
	log.Debugf("Setting IPFS timeout of %d seconds", timeout.Seconds())
	tps := rate.Limit(10) //10 TPS
	burst := 10
	if settingsObj.DagVerifierSettings.IPFSRateLimiter != nil {
		burst = settingsObj.DagVerifierSettings.IPFSRateLimiter.Burst
		if settingsObj.DagVerifierSettings.IPFSRateLimiter.RequestsPerSec == -1 {
			tps = rate.Inf
			burst = 0
		} else {
			tps = rate.Limit(settingsObj.DagVerifierSettings.IPFSRateLimiter.RequestsPerSec)
		}
	}
	log.Infof("Rate Limit configured for IPFS Client at %v TPS with a burst of %d", tps, burst)
	client.ipfsClientRateLimiter = rate.NewLimiter(tps, burst)
}

func (client *IpfsClient) DagGet(dagCid string) (DagChainBlock, error) {
	var dagBlock DagChainBlock
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

func (client *IpfsClient) GetPayload(payloadCid string, retryCount int) (DagPayloadData, error) {
	var payload DagPayloadData
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
