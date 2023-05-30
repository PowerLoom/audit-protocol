package mock

import (
	"context"

	"audit-protocol/goutils/datamodel"
)

type RedisMock struct {
	GetSnapshotAtEpochIDMock       func(ctx context.Context, projectID string, epochId int) (*datamodel.UnfinalizedSnapshot, error)
	GetStoredProjectsMock          func(ctx context.Context) ([]string, error)
	CheckIfProjectExistsMock       func(ctx context.Context, projectID string) (bool, error)
	StoreProjectsMock              func(background context.Context, projects []string) error
	AddUnfinalizedSnapshotCIDMock  func(ctx context.Context, msg *datamodel.PayloadCommitMessage, ttl int64) error
	AddSnapshotterStatusReportMock func(ctx context.Context, epochId int, projectId string, report *datamodel.SnapshotterStatusReport) error
}

func (m RedisMock) GetSnapshotAtEpochID(ctx context.Context, projectID string, epochId int) (*datamodel.UnfinalizedSnapshot, error) {
	return m.GetSnapshotAtEpochIDMock(ctx, projectID, epochId)
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
