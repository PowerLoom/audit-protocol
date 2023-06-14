package worker

type Worker interface {
	ConsumeTask() error
}

type Type string

const (
	TypePayloadCommitWorker Type = "payload-commit-worker"
	TypeEventDetectorWorker Type = "event-detector-worker"
)
