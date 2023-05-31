package worker

import (
	"context"

	"github.com/cenkalti/backoff/v4"
	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/taskmgr"
	workerInterface "audit-protocol/goutils/taskmgr/worker"
	pcService "audit-protocol/payload-commit/service"
)

type Worker struct {
	service  pcService.Service
	taskmgr  taskmgr.TaskMgr
	settings *settings.SettingsObj
}

func (w *Worker) ConsumeTask() error {
	// create buffered channel to limit the number of concurrent tasks.
	// buffered channel can is used to accept multiple messages and then process them in parallel.
	// as messages are generated at lower pace than they are consumed, we can use unbuffered channel as well.
	// TBD: we can use unbuffered channel as well.
	taskChan := make(chan taskmgr.TaskHandler, w.settings.WorkerConcurrency)
	defer close(taskChan)

	// start consuming messages in separate go routine.
	// messages will be sent back over taskChan.
	go func() {
		err := backoff.Retry(func() error {
			err := w.taskmgr.Consume(context.Background(), workerInterface.TypePayloadCommitWorker, taskChan)
			if err != nil {
				log.WithError(err).Error("failed to consume the message, retrying")

				return err
			}

			return nil
		}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), uint64(w.settings.RetryCount)))

		if err != nil {
			log.WithError(err).Fatal("failed to consume the messages after max retries")
		}
	}()

	for {
		// as task chan is buffered channel, we can use it as a semaphore to limit the number of concurrent tasks.
		taskHandler := <-taskChan

		go func(taskHandler taskmgr.TaskHandler) {
			msgBody := taskHandler.GetBody()

			log.Debug("received new rabbitmq message")

			err := w.service.Run(msgBody, taskHandler.GetTopic())
			if err != nil {
				log.WithError(err).Error("failed to run the task")

				err = backoff.Retry(func() error {
					return taskHandler.Nack(false)
				}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))
				if err != nil {
					log.WithError(err).Errorf("failed to nack the message")
				}
			} else {
				err = backoff.Retry(func() error {
					return taskHandler.Ack()
				}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))
				if err != nil {
					log.WithError(err).Error("failed to ack the message")
				}
			}
		}(taskHandler)
	}
}

// NewWorker creates a new *Worker listening for pruning tasks published by service responsible for creating segments
// a single Worker has capability to run multiple tasks concurrently using go routines.
// running multiple instances of this whole service will create multiple workers which can horizontally scale the service.
func NewWorker(settingsObj *settings.SettingsObj, pcService pcService.Service, mgr taskmgr.TaskMgr) workerInterface.Worker {
	return &Worker{service: pcService, settings: settingsObj, taskmgr: mgr}
}

func (w *Worker) ShutdownWorker() error {
	err := w.taskmgr.Shutdown(context.Background())
	if err != nil {
		log.WithError(err).Error("failed to shutdown the worker")
	}

	return err
}
