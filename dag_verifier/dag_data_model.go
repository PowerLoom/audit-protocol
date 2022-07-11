package main

type retryType int64

const (
	NO_RETRY_SUCCESS retryType = iota
	RETRY_IMMEDIATE            //TOD be used in timeout scenarios or non server returned error scenarios.
	RETRY_WITH_DELAY           //TO be used when immediate error is returned so that server is not overloaded.
	NO_RETRY_FAILURE           //This is to be used for unexpected conditions which are not recoverable and hence no retry
)

type SlackNotifyReq struct {
	DAGsummary string `json:"dagChainSummary"`
}

type DagChainSummary struct {
	Namespace                   string `json:"namespace"`
	ProjectsTrackedCount        int    `json:"projectsTrackedCount"`
	ProjectsWithIssuesCount     int    `json:"projectsWithIssuesCount"`
	ProjectsWithStuckChainCount int    `json:"projectsWithStuckChainCount"`
	CurrentMinChainHeight       int64  `json:"currentMinChainHeight"`
	OverallIssueCount           int    `json:"overallIssueCount"`
	OverallDAGChainGaps         int    `json:"overallDAGChainGaps"`
	OverallDAGChainDuplicates   int    `json:"overallDAGChainDuplicates"`
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
	Data           DagPayloadData
}

type DagPayloadData struct {
	Contract string `json:"contract"`
	/* Commenting out payload Data, to keep it generic.
	Token0Reserves map[string]float64  `json:"token0Reserves"`
	Token1Reserves map[string]float64 `json:"token1Reserves"`*/
	ChainHeightRange struct {
		Begin int64 `json:"begin"`
		End   int64 `json:"end"`
	} `json:"chainHeightRange"`
	BroadcastID string  `json:"broadcast_id"`
	Timestamp   float64 `json:"timestamp"`
}

type DagChainBlock struct {
	Data struct {
		Cid string `json:"cid"`
	} `json:"data"`
	Height    int64      `json:"height"`
	PrevCid   string     `json:"prevCid"`
	Timestamp int64      `json:"timestamp"`
	TxHash    string     `json:"txHash"`
	Payload   DagPayload `json:"payload"`
}
