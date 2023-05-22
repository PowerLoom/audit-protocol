package worker

type Worker interface {
	ConsumeTask() error
	ShutdownWorker() error
}

type Type string

const (
	TypePayloadCommitWorker Type = "payload-commit-worker"
)
