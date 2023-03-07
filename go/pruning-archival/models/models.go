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

type PruningCycleDetails struct {
	CycleID                     string `json:"pruningCycleID"`
	CycleStartTime              int64  `json:"cycleStartTime"`
	CycleEndTime                int64  `json:"cycleEndTime"`
	ProjectsCount               uint64 `json:"projectsCount"`
	ProjectsProcessSuccessCount uint64 `json:"projectsProcessSuccessCount"`
	ProjectsProcessFailedCount  uint64 `json:"projectsProcessFailedCount"`
	ProjectsNotProcessedCount   uint64 `json:"projectsNotProcessedCount"`
	HostName                    string `json:"hostName"`
	ErrorInLastcycle            bool   `json:"-"`
}
