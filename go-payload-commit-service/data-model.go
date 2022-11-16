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
	SkipAnchorProof      bool    `json:"skipAnchorProof"`
}

type PendingTransaction struct {
	TxHash           string            `json:"txHash"`
	RequestID        string            `json:"requestID"`
	LastTouchedBlock int               `json:"lastTouchedBlock"`
	EventData        RecordTxEventData `json:"event_data"`
}

type PayloadCommit struct {
	ProjectId string `json:"projectId"`
	CommitId  string `json:"commitId"`
	RequestID string `json:"requestID,omitempty"`
	Payload   json.RawMessage
	// following two can be used to substitute for not supplying the payload but the CID and hash itself
	SnapshotCID          string `json:"snapshotCID"`
	ApiKeyHash           string `json:"apiKeyHash"`
	TentativeBlockHeight int    `json:"tentativeBlockHeight"`
	Resubmitted          bool   `json:"resubmitted"`
	ResubmissionBlock    int    `json:"resubmissionBlock"` // corresponds to lastTouchedBlock in PendingTransaction model
	Web3Storage          bool   `json:"web3Storage"`       //This flag indicates to store the payload in web3.storage instead of IPFS.
	SkipAnchorProof      bool   `json:"skipAnchorProof"`
}

type _ChainHeightRange_ struct {
	Begin int64 `json:"begin"`
	End   int64 `json:"end"`
}

type PayloadData struct {
	ChainHeightRange  *_ChainHeightRange_ `json:"chainHeightRange"`
	EndBlockTimestamp float64             `json:"timestamp"`
}

type Snapshot struct {
	Cid string `json:"cid"`
}

type CommonTxRequestParams struct {
	Contract          string          `json:"contract"`
	Method            string          `json:"method"`
	Params            json.RawMessage `json:"params"`
	NetworkId         int             `json:"networkid"`
	Proxy             string          `json:"proxy"` //Review type
	HackerMan         bool            `json:"hackerman"`
	IgnoreGasEstimate bool            `json:"ignoreGasEstimate"`
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
