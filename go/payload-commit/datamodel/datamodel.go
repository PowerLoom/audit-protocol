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
	MessageType   PayloadCommitMessageType `json:"messageType"`
	Message       map[string]interface{}   `json:"message"`
	Web3Storage   bool                     `json:"web3Storage"`
	SourceChainID int                      `json:"sourceChainId"`
	ProjectID     string                   `json:"projectId"`
	EpochID       int                      `json:"epochId"`
	EpochEnd      int                      `json:"epochEnd"`
	SnapshotCID   string                   `json:"snapshotCID"`
}

type PowerloomSnapshotFinalizedMessage struct {
	EpochID     int    `json:"epochId"`
	EpochEnd    int    `json:"epochEnd"`
	ProjectID   string `json:"projectId"`
	SnapshotCID string `json:"snapshotCid"`
	Timestamp   int    `json:"timestamp"`
}

type PowerloomIndexFinalizedMessage struct {
	EpochID                         int    `json:"epochId"`
	ProjectID                       string `json:"projectId"`
	EpochEnd                        int    `json:"epochEnd"`
	IndexTailDAGBlockHeight         int    `json:"indexTailDAGBlockHeight"`
	TailBlockEpochSourceChainHeight int    `json:"tailBlockEpochSourceChainHeight"`
	IndexIdentifierHash             string `json:"indexIdentifierHash"`
	Timestamp                       int    `json:"timestamp"`
}

type PowerloomAggregateFinalizedMessage struct {
	EpochEnd     int    `json:"epochEnd"`
	EpochID      int    `json:"epochId"`
	ProjectID    string `json:"projectId"`
	AggregateCID string `json:"aggregateCid"`
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
	EpochID     int                       `json:"epochId"`
	Request     apitypes.TypedDataMessage `json:"request"`
	Signature   string                    `json:"signature"`
}

type IndexRelayerPayload struct {
	ProjectID                       string                    `json:"projectId"`
	EpochId                         int                       `json:"epochId"`
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
