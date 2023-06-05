package mock

import (
	"context"

	"audit-protocol/goutils/datamodel"
)

type RedisMock struct {
	GetUnfinalizedSnapshotAtEpochIDMock func(ctx context.Context, projectID string, epochId int) (*datamodel.UnfinalizedSnapshot, error)
	GetStoredProjectsMock               func(ctx context.Context) ([]string, error)
	CheckIfProjectExistsMock            func(ctx context.Context, projectID string) (bool, error)
	StoreProjectsMock                   func(background context.Context, projects []string) error
	AddUnfinalizedSnapshotCIDMock       func(ctx context.Context, msg *datamodel.PayloadCommitMessage, ttl int64) error
	AddSnapshotterStatusReportMock      func(ctx context.Context, epochId int, projectId string, report *datamodel.SnapshotterStatusReport) error
	StoreLastFinalizedEpochMock         func(ctx context.Context, projectID string, epochId int) error
	StoreFinalizedSnapshotMock          func(ctx context.Context, msg *datamodel.PowerloomSnapshotFinalizedMessage) error
	GetFinalizedSnapshotAtEpochIDMock   func(ctx context.Context, projectID string, epochId int) (*datamodel.PowerloomSnapshotFinalizedMessage, error)
}

func (m RedisMock) GetUnfinalizedSnapshotAtEpochID(ctx context.Context, projectID string, epochId int) (*datamodel.UnfinalizedSnapshot, error) {
	return m.GetUnfinalizedSnapshotAtEpochIDMock(ctx, projectID, epochId)
}

func (m RedisMock) GetStoredProjects(ctx context.Context) ([]string, error) {
	return m.GetStoredProjectsMock(ctx)
}

func (m RedisMock) CheckIfProjectExists(ctx context.Context, projectID string) (bool, error) {
	return m.CheckIfProjectExistsMock(ctx, projectID)
}

func (m RedisMock) StoreProjects(background context.Context, projects []string) error {
	return m.StoreProjectsMock(background, projects)
}

func (m RedisMock) AddUnfinalizedSnapshotCID(ctx context.Context, msg *datamodel.PayloadCommitMessage, ttl int64) error {
	return m.AddUnfinalizedSnapshotCIDMock(ctx, msg, ttl)
}

func (m RedisMock) AddSnapshotterStatusReport(ctx context.Context, epochId int, projectId string, report *datamodel.SnapshotterStatusReport) error {
	return m.AddSnapshotterStatusReportMock(ctx, epochId, projectId, report)
}

func (m RedisMock) StoreLastFinalizedEpoch(ctx context.Context, projectID string, epochId int) error {
	return m.StoreLastFinalizedEpochMock(ctx, projectID, epochId)
}

func (m RedisMock) StoreFinalizedSnapshot(ctx context.Context, msg *datamodel.PowerloomSnapshotFinalizedMessage) error {
	return m.StoreFinalizedSnapshotMock(ctx, msg)
}

func (m RedisMock) GetFinalizedSnapshotAtEpochID(ctx context.Context, projectID string, epochId int) (*datamodel.PowerloomSnapshotFinalizedMessage, error) {
	return m.GetFinalizedSnapshotAtEpochIDMock(ctx, projectID, epochId)
}
