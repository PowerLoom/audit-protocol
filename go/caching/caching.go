package caching

import (
	"context"
	"errors"

	"audit-protocol/goutils/datamodel"
)

// DbCache is responsible for data caching in db stores like redis, memcache etc.
// for disk caching use DiskCache interface
type DbCache interface {
	GetStoredProjects(ctx context.Context) ([]string, error)
	GetLastProjectIndexedState(ctx context.Context) (map[string]*datamodel.ProjectIndexedState, error)
	GetPayloadCidAtDAGHeight(ctx context.Context, projectID string, dagHeight int) (string, error)
	GetLastVerifiedDagHeight(ctx context.Context, projectID string) (int, error)
	UpdateDagVerificationStatus(ctx context.Context, projectID string, status *datamodel.DagVerifierStatus) error
	GetProjectDAGBlockHeight(ctx context.Context, projectID string) (int, error)
	UpdateDAGChainIssues(ctx context.Context, projectID string, dagChainIssues []*datamodel.DagChainIssue) error
	StorePruningIssueReport(ctx context.Context, report *datamodel.PruningIssueReport) error
	GetPruningVerificationStatus(ctx context.Context) (map[string]*datamodel.ProjectPruningVerificationStatus, error)
	UpdatePruningVerificationStatus(ctx context.Context, projectID string, status *datamodel.ProjectPruningVerificationStatus) error
	GetProjectDagSegments(ctx context.Context, projectID string) (map[string]string, error)
	StoreReportedIssues(ctx context.Context, issue *datamodel.IssueReport) error
	RemoveOlderReportedIssues(ctx context.Context, tillTime int) error

	//GetPayloadCIDs - startHeight and endHeight are string because they can be "-inf" or "+inf"
	// -inf & +inf are just alias for start and end respectively, though the values must be changed according to cache implementation
	GetPayloadCIDs(ctx context.Context, projectID string, startHeight, endHeight string) ([]*datamodel.DagBlock, error)

	//GetDagChainCIDs - startHeight and endHeight are string because they can be "-inf" or "+inf"
	// -inf & +inf are just alias for start and end respectively, though the values must be changed according to cache implementation
	GetDagChainCIDs(ctx context.Context, projectID string, startHeight, endHeight string) ([]*datamodel.DagBlock, error)
}

// DiskCache is responsible for data caching in local disk
type DiskCache interface {
	Read(filepath string) ([]byte, error)
	Write(filepath string, data []byte) error
}

var (
	ErrNotFound                         = errors.New("not found")
	ErrGettingProjects                  = errors.New("error getting stored projects")
	ErrGettingLastDagVerificationStatus = errors.New("error getting last dag verification status")
)
