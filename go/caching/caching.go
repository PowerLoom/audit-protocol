package caching

import (
	"context"
	"errors"

	"audit-protocol/goutils/datamodel"
)

// DbCache is responsible for data caching in db stores like redis, memcache etc.
// for disk caching use DiskCache interface
type DbCache interface {
	GetSnapshotCidAtEpochID(ctx context.Context, projectID string, epochId int) (string, error)
	GetStoredProjects(ctx context.Context) ([]string, error)
	CheckIfProjectExists(ctx context.Context, projectID string) (bool, error)
	StoreProjects(background context.Context, projects []string) error
	AddUnfinalizedSnapshotCID(ctx context.Context, projectID, snapshotCID string, height float64) error
	AddSnapshotterStatusReport(ctx context.Context, epochId int, projectId string, report *datamodel.SnapshotterStatusReport) error
}

// DiskCache is responsible for data caching in local disk
type DiskCache interface {
	Read(filepath string) ([]byte, error)
	Write(filepath string, data []byte) error
}

type MemCache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}) error
	Delete(key string)
}

var (
	ErrGettingProjects = errors.New("error getting stored projects")
)
