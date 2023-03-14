package models

type ProjectMetaData struct {
	DagChains map[string]string
}

type ProjectDAGSegment struct {
	BeginHeight int    `json:"beginHeight"`
	EndHeight   int    `json:"endHeight"`
	EndDAGCID   string `json:"endDAGCID"`
	StorageType string `json:"storageType"`
}

type ProjectPruneState struct {
	ProjectId        string
	LastPrunedHeight int
	ErrorInLastCycle bool
}

type Web3StoragePostResponse struct {
	CID string `json:"cid"`
}

type Web3StorageErrResponse struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type ProjectPruningReport struct {
	HostName                  string `json:"Host"`
	ProjectID                 string `json:"projectID"`
	DAGSegmentsProcessed      int    `json:"DAGSegmentsProcessed"`
	DAGSegmentsArchived       int    `json:"DAGSegmentsArchived"`
	DAGSegmentsArchivalFailed int    `json:"DAGSegmentsArchivalFailed,omitempty"`
	ArchivalFailureCause      string `json:"failureCause,omitempty"`
	CIDsUnPinned              int    `json:"CIDsUnPinned"`
	UnPinFailed               int    `json:"unPinFailed,omitempty"`
	LocalCacheDeletionsFailed int    `json:"localCacheDeletionsFailed,omitempty"`
}

type PruningTaskDetails struct {
	TaskID    string `json:"pruningTaskID"`
	StartTime int64  `json:"cycleStartTime"`
	EndTime   int64  `json:"cycleEndTime"`
	HostName  string `json:"hostName"`
}

type ZSets struct {
	PayloadCids map[int]string `json:"payloadCids"`
	DagCids     map[int]string `json:"dagCids"`
}
