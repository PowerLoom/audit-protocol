package main

import "encoding/json"

type RecordTxEventData struct {
	TxHash               string  `json:"txHash"`
	ProjectId            string  `json:"projectId"`
	ApiKeyHash           string  `json:"apiKeyHash"`
	Timestamp            float64 `json:"timestamp"`
	PayloadCommitId      string  `json:"payloadCommitId"`
	SnapshotCid          string  `json:"snapshotCid"`
	TentativeBlockHeight int     `json:"tentativeBlockHeight"`
}

type PendingTransaction struct {
	TxHash           string            `json:"txHash"`
	LastTouchedBlock int               `json:"lastTouchedBlock"`
	EventData        RecordTxEventData `json:"event_data"`
}

type PayloadCommit struct {
	ProjectId string `json:"projectId"`
	CommitId  string `json:"commitId"`
	Payload   json.RawMessage
	// following two can be used to substitute for not supplying the payload but the CID and hash itself
	SnapshotCID          string `json:"snapshotCID"`
	ApiKeyHash           string `json:"apiKeyHash"`
	TentativeBlockHeight int    `json:"tentativeBlockHeight"`
	Resubmitted          bool   `json:"resubmitted"`
	ResubmissionBlock    int    `json:"resubmissionBlock"` // corresponds to lastTouchedBlock in PendingTransaction model
}

type Snapshot struct {
	Cid  string `json:"cid"`
	Type string `json:"type"`
}
