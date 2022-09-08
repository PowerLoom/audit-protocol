package main

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

type SummaryProjectState struct {
	ProjectId     string `json:"projectId"`
	ProjectHeight string `json:"chainHeight"`
}

type DagChainReport struct {
	Namespace                   string                `json:"namespace"`
	Severity                    string                `json:"severity"`                               //HIGH,MEDIUM, LOW, CLEAR
	ProjectsWithCacheIssueCount int                   `json:"projectsWithCacheIssuesCount,omitempty"` //Projects that have only issues in the cached data.
	ProjectsTrackedCount        int                   `json:"projectsTrackedCount,omitempty"`
	ProjectsWithIssuesCount     int                   `json:"projectsWithIssuesCount,omitempty"`
	ProjectsWithStuckChainCount int                   `json:"projectsWithStuckChainCount,omitempty"`
	CurrentMinChainHeight       int64                 `json:"currentMinChainHeight,omitempty"`
	OverallIssueCount           int                   `json:"overallIssueCount,omitempty"`
	OverallDAGChainGaps         int                   `json:"overallDAGChainGaps,omitempty"`
	OverallDAGChainDuplicates   int                   `json:"overallDAGChainDuplicates,omitempty"`
	SummaryProjectsStuckDetails []SummaryProjectState `json:"summaryProjectsStuck,omitempty"`
	SummaryProjectsRecovered    []SummaryProjectState `json:"summaryProjectsRecovered,omitempty"`
	IssurResolvedMessage        string                `json:"issueResolvedMessage,omitempty"`
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

type IPLDLink struct {
	LinkData string `json:"/"`
}

type DagChainBlock struct {
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
