package datamodel

import "github.com/ethereum/go-ethereum/signer/core/apitypes"

type SummaryProjectVerificationStatus struct {
	ProjectId     string `json:"projectId"`
	ProjectHeight string `json:"chainHeight"`
}

type SlackResp struct {
	Error            string `json:"error"`
	Ok               bool   `json:"ok"`
	ResponseMetadata struct {
		Messages []string `json:"messages"`
	} `json:"response_metadata"`
}

type Web3StoragePutResponse struct {
	CID string `json:"cid"`
}

type Web3StorageErrResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type (
	PayloadCommitMessageType          string
	PayloadCommitFinalizedMessageType string
)

type PayloadCommitMessage struct {
	Message       map[string]interface{} `json:"message"`
	Web3Storage   bool                   `json:"web3Storage"`
	SourceChainID int                    `json:"sourceChainId"`
	ProjectID     string                 `json:"projectId"`
	EpochID       int                    `json:"epochId"`
	SnapshotCID   string                 `json:"snapshotCID"`
}

type PowerloomSnapshotFinalizedMessage struct {
	EpochID     int    `json:"epochId"`
	ProjectID   string `json:"projectId"`
	SnapshotCID string `json:"snapshotCid"`
	Timestamp   int    `json:"timestamp"`
}

type PayloadCommitFinalizedMessage struct {
	Message       *PowerloomSnapshotFinalizedMessage `json:"message"`
	Web3Storage   bool                               `json:"web3Storage"`
	SourceChainID int                                `json:"sourceChainId"`
}

type SnapshotRelayerPayload struct {
	ProjectID   string                    `json:"projectId"`
	SnapshotCID string                    `json:"snapshotCid"`
	EpochID     int                       `json:"epochId"`
	Request     apitypes.TypedDataMessage `json:"request"`
	Signature   string                    `json:"signature"`
}

type SnapshotterStatusReport struct {
	SubmittedSnapshotCid string `json:"submittedSnapshotCid"`
	FinalizedSnapshotCid string `json:"finalizedSnapshotCid"`
	Missed               bool   `json:"missed"`
}

type UnfinalizedSnapshot struct {
	SnapshotCID string `json:"snapshotCid"`
	Expiration  int64  `json:"expiration"`
}
