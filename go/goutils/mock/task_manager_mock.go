package mock

import (
	"context"

	"audit-protocol/goutils/taskmgr/worker"
)

type TaskHandlerMock struct {
	GetBodyMock  func() []byte
	GetTopicMock func() string
	AckMock      func() error
	NackMock     func(requeue bool) error
}

func (m TaskHandlerMock) GetBody() []byte {
	return m.GetBodyMock()
}

func (m TaskHandlerMock) GetTopic() string {
	return m.GetTopicMock()
}

func (m TaskHandlerMock) Ack() error {
	return m.AckMock()
}

func (m TaskHandlerMock) Nack(requeue bool) error {
	return m.NackMock(requeue)
}

type TaskManagerMock struct {
	PublishMock  func(ctx context.Context) error
	ConsumeMock  func(ctx context.Context, workerType worker.Type, msgChan chan TaskHandlerMock) error
	ShutdownMock func(ctx context.Context) error
}

func (m TaskManagerMock) Publish(ctx context.Context) error {
	return m.PublishMock(ctx)
}

func (m TaskManagerMock) Consume(ctx context.Context, workerType worker.Type, msgChan chan TaskHandlerMock) error {
	return m.ConsumeMock(ctx, workerType, msgChan)
}

func (m TaskManagerMock) Shutdown(ctx context.Context) error {
	return m.ShutdownMock(ctx)
}
