package caching

import (
	"context"
	"errors"
)

// DbCache is responsible for data caching in db stores like redis, memcache etc.
// for disk caching use DiskCache interface
type DbCache interface {
	GetPayloadCidAtEpochID(ctx context.Context, projectID string, dagHeight int) (string, error)
	GetStoredProjects(ctx context.Context) ([]string, error)
	CheckIfProjectExists(ctx context.Context, projectID string) (bool, error)
	StoreProjects(background context.Context, projects []string) error
	AddPayloadCID(ctx context.Context, projectID, payloadCID string, height float64) error
	RemovePayloadCIDAtEpochID(ctx context.Context, projectID string, dagHeight int) error
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
