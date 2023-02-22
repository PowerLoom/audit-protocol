package datamodel

import "encoding/json"

type retryType int64

const (
	NO_RETRY_SUCCESS retryType = iota
	RETRY_IMMEDIATE            //TOD be used in timeout scenarios or non server returned error scenarios.
	RETRY_WITH_DELAY           //TO be used when immediate error is returned so that server is not overloaded.
	NO_RETRY_FAILURE           //This is to be used for unexpected conditions which are not recoverable and hence no retry
)

type SlackNotifyReq struct {
	DAGsummary    string `json:"dagChainSummary"`
	IssueSeverity string `json:"severity"`
}

type SummaryProjectVerificationStatus struct {
	ProjectId     string `json:"projectId"`
	ProjectHeight string `json:"chainHeight"`
}

type DagChainReport struct {
	InstanceId                  string                             `json:"instanceid"`
	HostName                    string                             `json:"hostname"`
	Severity                    string                             `json:"severity"`                               //HIGH,MEDIUM, LOW, CLEAR
	ProjectsWithCacheIssueCount int                                `json:"projectsWithCacheIssuesCount,omitempty"` //Projects that have only issues in the cached data.
	ProjectsTrackedCount        int                                `json:"projectsTrackedCount,omitempty"`
	ProjectsWithIssuesCount     int                                `json:"projectsWithIssuesCount,omitempty"`
	ProjectsWithStuckChainCount int                                `json:"projectsWithStuckChainCount,omitempty"`
	CurrentMinChainHeight       int64                              `json:"currentMinChainHeight,omitempty"`
	OverallIssueCount           int                                `json:"overallIssueCount,omitempty"`
	OverallDAGChainGaps         int                                `json:"overallDAGChainGaps,omitempty"`
	OverallDAGChainDuplicates   int                                `json:"overallDAGChainDuplicates,omitempty"`
	SummaryProjectsStuckDetails []SummaryProjectVerificationStatus `json:"summaryProjectsStuck,omitempty"`
	SummaryProjectsRecovered    []SummaryProjectVerificationStatus `json:"summaryProjectsRecovered,omitempty"`
	IssurResolvedMessage        string                             `json:"issueResolvedMessage,omitempty"`
}

type SlackResp struct {
	Error            string `json:"error"`
	Ok               bool   `json:"ok"`
	ResponseMetadata struct {
		Messages []string `json:"messages"`
	} `json:"response_metadata"`
}

type DagChainIssue struct {
	IssueType string `json:"issueType"`
	//In case of missing blocks in chain or Gap
	MissingBlockHeightStart int64 `json:"missingBlockHeightStart"`
	MissingBlockHeightEnd   int64 `json:"missingBlockHeightEnd"`
	TimestampIdentified     int64 `json:"timestampIdentified"`
	DAGBlockHeight          int64 `json:"dagBlockHeight"`
}

type DagPayload struct {
	PayloadCid     string `json:"payloadCid"`
	DagChainHeight int64  `json:"dagChainHeight"`
	Data           DagPayloadChainHeightRange
}

type DagPayloadChainHeightRange struct {
	ChainHeightRange struct {
		Begin int64 `json:"begin"`
		End   int64 `json:"end"`
	} `json:"chainHeightRange"`
}

type IPLDLink struct {
	LinkData string `json:"/"`
}

type DagBlock struct {
	Data struct {
		Cid IPLDLink `json:"cid"`
	} `json:"data"`
	Height     int64      `json:"height"`
	PrevCid    IPLDLink   `json:"prevCid"`
	Timestamp  int64      `json:"timestamp"`
	TxHash     string     `json:"txHash"`
	Payload    DagPayload `json:"payload"`
	CurrentCid string
}

type IssueReport struct {
	Instanceid       string  `json:"instanceID"`
	Namespace        string  `json:"namespace,omitempty"`
	Severity         string  `json:"severity"`
	IssueType        string  `json:"issueType"`
	ProjectID        string  `json:"projectID"`
	Epochs           []int64 `json:"epochs,omitempty"`
	NoOfEpochsBehind int64   `json:"noOfEpochsBehind,omitempty"`
	Service          string  `json:"serviceName"`
}

type RecordTxEventData struct {
	TxHash               string  `json:"txHash"`
	ProjectId            string  `json:"projectId"`
	ApiKeyHash           string  `json:"apiKeyHash"`
	Timestamp            float64 `json:"timestamp"`
	PayloadCommitId      string  `json:"payloadCommitId"`
	SnapshotCid          string  `json:"snapshotCid"`
	TentativeBlockHeight int     `json:"tentativeBlockHeight"`
	SkipAnchorProof      bool    `json:"skipAnchorProof"`
}

type PendingTransaction struct {
	TxHash           string            `json:"txHash"`
	RequestID        string            `json:"requestID"`
	LastTouchedBlock int               `json:"lastTouchedBlock"`
	EventData        RecordTxEventData `json:"event_data"`
}

type SourceChainDetails_ struct {
	ChainID          int `json:"chainID"`
	EpochStartHeight int `json:"epochStartHeight"`
	EpochEndHeight   int `json:"epochEndHeight"`
}

type PayloadCommit struct {
	ProjectId          string              `json:"projectId"`
	CommitId           string              `json:"commitId"`
	SourceChainDetails SourceChainDetails_ `json:"sourceChainDetails"`
	Payload            json.RawMessage
	RequestID          string `json:"requestID,omitempty"`
	// following two can be used to substitute for not supplying the payload but the CID and hash itself
	SnapshotCID           string `json:"snapshotCID"`
	ApiKeyHash            string `json:"apiKeyHash"`
	TentativeBlockHeight  int    `json:"tentativeBlockHeight"`
	Resubmitted           bool   `json:"resubmitted"`
	ResubmissionBlock     int    `json:"resubmissionBlock"` // corresponds to lastTouchedBlock in PendingTransaction model
	Web3Storage           bool   `json:"web3Storage"`       //This flag indicates to store the payload in web3.storage instead of IPFS.
	SkipAnchorProof       bool   `json:"skipAnchorProof"`
	ConsensusSubmissionTs int64  `json:"-"`
	IsSummaryProject      bool   `json:"-"`
}

type _ChainHeightRange_ struct {
	Begin int64 `json:"begin"`
	End   int64 `json:"end"`
}

type PayloadData struct {
	ChainHeightRange *_ChainHeightRange_ `json:"chainHeightRange"`
}

type Snapshot struct {
	Cid string `json:"cid"`
}

type CommonTxRequestParams struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type AuditContractCommitParams struct {
	RequestID            string `json:"requestID"`
	PayloadCommitId      string `json:"payloadCommitId"`
	SnapshotCid          string `json:"snapshotCid"`
	ApiKeyHash           string `json:"apiKeyHash"`
	ProjectId            string `json:"projectId"`
	TentativeBlockHeight int    `json:"tentativeBlockHeight"`
}

type AuditContractCommitResp struct {
	Success bool                          `json:"success"`
	Data    []AuditContractCommitRespData `json:"data"`
	Error   AuditContractErrResp          `json:"error"`
}
type AuditContractCommitRespData struct {
	TxHash    string `json:"txHash"`
	RequestID string `json:"requestID"`
}

type AuditContractErrResp struct {
	Message string `json:"message"`
	Error   struct {
		Message string `json:"message"`
		Details struct {
			BriefMessage string `json:"briefMessage"`
			FullMessage  string `json:"fullMessage"`
			Data         []struct {
				Contract       string          `json:"contract"`
				Method         string          `json:"method"`
				Params         json.RawMessage `json:"params"`
				EncodingErrors struct {
					APIKeyHash string `json:"apiKeyHash"`
				} `json:"encodingErrors"`
			} `json:"data"`
		} `json:"details"`
	} `json:"error"`
}

type Web3StoragePutResponse struct {
	CID string `json:"cid"`
}

type Web3StorageErrResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

// Note that this is a simulated request and hence eventData structure has been hardcoded.
type AuditContractSimWebhookCallbackRequest struct {
	Type             string `json:"type"`
	RequestID        string `json:"requestID"`
	TxHash           string `json:"txHash"`
	LogIndex         int64  `json:"logIndex,omitempty"`
	BlockNumber      int64  `json:"blockNumber,omitempty"`
	TransactionIndex int64  `json:"transactionIndex,omitempty"`
	Contract         string `json:"contract"`
	EventName        string `json:"event_name"`
	EventData        struct {
		PayloadCommitId      string `json:"payloadCommitId"`
		SnapshotCid          string `json:"snapshotCid"`
		ApiKeyHash           string `json:"apiKeyHash"`
		ProjectId            string `json:"projectId"`
		TentativeBlockHeight int    `json:"tentativeBlockHeight"`
		Timestamp            int64  `json:"timestamp"`
	} `json:"event_data"`
	ProstvigilEventID int64 `json:"prostvigil_event_id,omitempty"`
	Ctime             int64 `json:"ctime"`
}
