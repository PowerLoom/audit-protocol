package datamodel

import "github.com/ethereum/go-ethereum/signer/core/apitypes"

type SnapshotSubmissionState string

const (
	MissedSnapshotSubmission    SnapshotSubmissionState = "MISSED_SNAPSHOT"
	IncorrectSnapshotSubmission SnapshotSubmissionState = "SUBMITTED_INCORRECT_SNAPSHOT"
)

type IssueType string

const (
	PayloadCommitInternalIssue      IssueType = "PAYLOAD_COMMIT_INTERNAL_ISSUE" // generic issue type for internal errors
	MissedSnapshotIssue             IssueType = "MISSED_SNAPSHOT"               // when a snapshot was missed
	SubmittedIncorrectSnapshotIssue IssueType = "SUBMITTED_INCORRECT_SNAPSHOT"  // when a snapshot was submitted but it was incorrect
)

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
	Message       map[string]interface{} `json:"message" validate:"required"`
	Web3Storage   bool                   `json:"web3Storage"`
	SourceChainID int                    `json:"sourceChainId"`
	ProjectID     string                 `json:"projectId" validate:"required"`
	EpochID       int                    `json:"epochId" validate:"required"`
	SnapshotCID   string                 `json:"snapshotCID"`
}

type PowerloomSnapshotFinalizedMessage struct {
	EpochID     int    `json:"epochId"`
	ProjectID   string `json:"projectId"`
	SnapshotCID string `json:"snapshotCid"`
	Timestamp   int    `json:"timestamp"`
	Expiry      int    `json:"expiry"` // for redis cleanup
}

type PayloadCommitFinalizedMessage struct {
	Message       *PowerloomSnapshotFinalizedMessage `json:"message" validate:"required"`
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
	SubmittedSnapshotCid string                  `json:"submittedSnapshotCid"`
	FinalizedSnapshotCid string                  `json:"finalizedSnapshotCid"`
	State                SnapshotSubmissionState `json:"state"`
}

type UnfinalizedSnapshot struct {
	SnapshotCID string                 `json:"snapshotCid"`
	Snapshot    map[string]interface{} `json:"snapshot"`
	TTL         int64                  `json:"ttl"`
}

type SnapshotterIssue struct {
	InstanceID      string `json:"instanceID"`
	IssueType       string `json:"issueType"`
	ProjectID       string `json:"projectID"`
	EpochID         string `json:"epochId"`
	TimeOfReporting string `json:"timeOfReporting"`
	Extra           string `json:"extra"`
}

type RelayerRequest struct {
	ProjectID       string   `json:"projectId"`
	SnapshotCID     string   `json:"snapshotCid"`
	EpochID         int      `json:"epochId"`
	Signature       string   `json:"signature"`
	Request         *Request `json:"request"`
	ContractAddress string   `json:"contractAddress"`
}

type Request struct {
	Deadline uint64 `json:"deadline"`
}
