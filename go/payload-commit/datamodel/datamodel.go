package datamodel

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
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
	MessageType          PayloadCommitMessageType `json:"messageType"`
	Message              map[string]interface{}   `json:"message"`
	Web3Storage          bool                     `json:"web3Storage"`
	SourceChainID        int                      `json:"sourceChainId"`
	ProjectID            string                   `json:"projectId"`
	EpochEndHeight       int                      `json:"epochEndHeight"`
	TentativeBlockHeight int                      `json:"tentativeBlockHeight"`
	SnapshotCID          string                   `json:"snapshotCID"`
}

type PowerloomSnapshotFinalizedMessage struct {
	DAGBlockHeight int    `json:"DAGBlockHeight"`
	ProjectID      string `json:"projectId"`
	SnapshotCID    string `json:"snapshotCid"`
	BroadcastID    string `json:"broadcastId"`
	Timestamp      int    `json:"timestamp"`
}

type PowerloomIndexFinalizedMessage struct {
	DAGBlockHeight                  int    `json:"DAGBlockHeight"`
	ProjectID                       string `json:"projectId"`
	IndexTailDAGBlockHeight         int    `json:"indexTailDAGBlockHeight"`
	TailBlockEpochSourceChainHeight int    `json:"tailBlockEpochSourceChainHeight"`
	IndexIdentifierHash             string `json:"indexIdentifierHash"`
	BroadcastID                     string `json:"broadcastId"`
	Timestamp                       int    `json:"timestamp"`
}

type PowerloomAggregateFinalizedMessage struct {
	EpochEnd     int    `json:"epochEnd"`
	ProjectID    string `json:"projectId"`
	AggregateCID string `json:"aggregateCid"`
	BroadcastID  string `json:"broadcastId"`
	Timestamp    int    `json:"timestamp"`
}

type PayloadCommitFinalizedMessage struct {
	MessageType   PayloadCommitFinalizedMessageType `json:"messageType"`
	Message       json.RawMessage                   `json:"message"`
	Web3Storage   bool                              `json:"web3Storage"`
	SourceChainID int                               `json:"sourceChainId"`
}

type SnapshotRelayerPayload struct {
	ProjectID   string                    `json:"projectId"`
	SnapshotCID string                    `json:"snapshotCid"`
	EpochEnd    int                       `json:"epochEnd"`
	Request     apitypes.TypedDataMessage `json:"request"`
	Signature   string                    `json:"signature"`
}

type IndexRelayerPayload struct {
	ProjectID                       string                    `json:"projectId"`
	DagBlockHeight                  int                       `json:"DAGBlockHeight"`
	IndexTailDagBlockHeight         int                       `json:"indexTailDAGBlockHeight"`
	TailBlockEpochSourceChainHeight int                       `json:"tailBlockEpochSourceChainHeight"`
	IndexIdentifierHash             string                    `json:"indexIdentifierHash"`
	Request                         apitypes.TypedDataMessage `json:"request"`
	Signature                       string                    `json:"signature"`
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
