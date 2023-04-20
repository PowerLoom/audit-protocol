package datamodel

import (
	"encoding/json"
	"fmt"
)

type (
	PayloadCommitMessageType          string
	PayloadCommitFinalizedMessageType string
)

const (
	MessageTypeSnapshot  PayloadCommitMessageType = "SNAPSHOT"
	MessageTypeIndex     PayloadCommitMessageType = "INDEX"
	MessageTypeAggregate PayloadCommitMessageType = "AGGREGATE"
)

const (
	SnapshotFinalized  PayloadCommitFinalizedMessageType = "SNAPSHOTFINALIZED"
	IndexFinalized     PayloadCommitFinalizedMessageType = "INDEXFINALIZED"
	AggregateFinalized PayloadCommitFinalizedMessageType = "AGGREGATEFINALIZED"
)

type PayloadCommitMessage struct {
	MessageType      PayloadCommitMessageType `json:"message_type"`
	Message          map[string]interface{}   `json:"message"`
	Web3Storage      bool                     `json:"web3Storage"`
	SourceChainID    int                      `json:"sourceChainId"`
	ProjectID        string                   `json:"projectId"`
	EpochEndHeight   int                      `json:"epochEndHeight"`
	EpochStartHeight int                      `json:"epochStartHeight"`
}

type PowerloomSnapshotFinalizedMessage struct {
	DAGBlockHeight int    `json:"DAGBlockHeight"`
	ProjectID      string `json:"projectId"`
	SnapshotCID    string `json:"snapshotCid"`
	BroadcastID    string `json:"broadcast_id"`
	Timestamp      int    `json:"timestamp"`
}

type PowerloomIndexFinalizedMessage struct {
	DAGBlockHeight                  int    `json:"DAGBlockHeight"`
	ProjectID                       string `json:"projectId"`
	IndexTailDAGBlockHeight         int    `json:"indexTailDAGBlockHeight"`
	TailBlockEpochSourceChainHeight int    `json:"tailBlockEpochSourceChainHeight"`
	IndexIdentifierHash             string `json:"indexIdentifierHash"`
	BroadcastID                     string `json:"broadcast_id"`
	Timestamp                       int    `json:"timestamp"`
}

type PowerloomAggregateFinalizedMessage struct {
	EpochEnd     int    `json:"epochEnd"`
	ProjectID    string `json:"projectId"`
	AggregateCID string `json:"aggregateCid"`
	BroadcastID  string `json:"broadcast_id"`
	Timestamp    int    `json:"timestamp"`
}

type PayloadCommitFinalizedMessage struct {
	MessageType   PayloadCommitFinalizedMessageType `json:"message_type"`
	Message       json.RawMessage                   `json:"message"`
	Web3Storage   bool                              `json:"web3Storage"`
	SourceChainID int                               `json:"sourceChainId"`
}

func (m *PayloadCommitFinalizedMessage) UnmarshalMessage() (interface{}, error) {
	var msg interface{}

	switch m.MessageType {
	case SnapshotFinalized:
		msg = new(PowerloomSnapshotFinalizedMessage)
	case IndexFinalized:
		msg = new(PowerloomIndexFinalizedMessage)
	case AggregateFinalized:
		msg = new(PowerloomAggregateFinalizedMessage)
	default:
		return nil, fmt.Errorf("unknown message_type: %s", m.MessageType)
	}

	if err := json.Unmarshal(m.Message, msg); err != nil {
		return nil, err
	}

	return msg, nil
}
