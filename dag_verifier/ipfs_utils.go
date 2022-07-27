package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	log "github.com/sirupsen/logrus"
)

type IpfsClient struct {
	ipfsClient *shell.Shell
}

func (client *IpfsClient) Init(settingsObj *SettingsObj) {
	url := settingsObj.IpfsURL
	// Convert the URL from /ip4/172.31.16.206/tcp/5001 to IP:Port format.
	//TODO: this is a very dirty way of doing it, better to take the IP and port from settings directly by adding new fields.
	connectUrl := strings.Split(url, "/")[2] + ":" + strings.Split(url, "/")[4]
	log.Debug("Initializing the IPFS client with IPFS Daemon URL:", connectUrl)
	client.ipfsClient = shell.NewShell(connectUrl)
	timeout := time.Duration(settingsObj.IpfsTimeout * 1000000000)
	client.ipfsClient.SetTimeout(timeout)
	log.Debugf("Setting IPFS timeout of %d seconds", timeout.Seconds())
}

func (client *IpfsClient) DagGet(dagCid string) (DagChainBlock, error) {
	var dagBlock DagChainBlock
	err := client.ipfsClient.DagGet(dagCid, &dagBlock)
	if err != nil {
		log.Error("Error in getting Dag block from IPFS, Dag-CID:", dagCid, ", error:", err)
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
		data, err := client.ipfsClient.Cat(payloadCid)
		if err != nil {
			log.Error("Failed to fetch Payload from IPFS, CID:", payloadCid, ", error:", err)
			//TODO": fin-grain handling needed that in case of network error, retry with exponential backoff
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
