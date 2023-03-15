package worker

import (
	"context"

	"github.com/cenkalti/backoff/v4"
	"github.com/remeh/sizedwaitgroup"
	log "github.com/sirupsen/logrus"

	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/taskmgr"
	worker2 "audit-protocol/goutils/taskmgr/worker"
	"audit-protocol/pruning-archival/service"
)

type worker struct {
	service  *service.PruningService
	taskmgr  taskmgr.TaskMgr
	settings *settings.SettingsObj
}

func (w worker) ConsumeTask() error {
	// create buffered channel to limit the number of concurrent tasks.
	// buffered channel can is used to accept multiple messages and then process them in parallel.
	// as messages are generated at lower pace than they are consumed, we can use unbuffered channel as well.
	// TBD: we can use unbuffered channel as well.
	taskChan := make(chan taskmgr.TaskHandler, w.settings.PruningServiceSettings.Concurrency)
	defer close(taskChan)

	// start consuming messages in separate go routine.
	// messages will be sent back over taskChan.
	go func() {
		err := backoff.Retry(
			func() error {
				errChan := make(chan error)
				err := w.taskmgr.Consume(context.Background(), worker2.TypePruningServiceWorker, taskChan, errChan)
				if err != nil {
					log.Fatalf("Failed to consume the message: %v", err)
				}

				err = <-errChan
				if err != nil {
					log.Fatalf("Failed to consume the message: %v, retrying again", err)

					return err
				}

				return nil
			}, backoff.NewExponentialBackOff())

		if err != nil {
			log.Fatalf("Failed to consume the message: %v", err)
		}
	}()

	// create a wait group to wait for previous the tasks to finish.
	// limit number of concurrent tasks per worker
	swg := sizedwaitgroup.New(w.settings.PruningServiceSettings.Concurrency)

	for {
		swg.Add()

		// as task chan is buffered channel, we can use it as a semaphore to limit the number of concurrent tasks.
		taskHandler := <-taskChan

		go func(taskHandler taskmgr.TaskHandler) {
			msgBody := taskHandler.GetBody()

			log.Debug("Received message: ", string(msgBody))

			err := w.service.Run(msgBody)
			if err != nil {
				log.Errorf("Failed to run the task: %v", err)

				err = backoff.Retry(func() error {
					return taskHandler.Nack(false)
				}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), uint64(*w.settings.RetryCount)))
				if err != nil {
					log.Errorf("Failed to nack the message: %v", err)
				}
			} else {
				err = backoff.Retry(func() error {
					return taskHandler.Ack()
				}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), uint64(*w.settings.RetryCount)))
				if err != nil {
					log.Errorf("Failed to ack the message: %v", err)
				}
			}

			swg.Done()
		}(taskHandler)

		// wait till all the previous tasks are finished.
		swg.Wait()
	}
}

// NewWorker creates a new worker listening for pruning tasks published by service responsible for creating segments
// a single worker has capability to run multiple tasks concurrently using go routines.
// running multiple instances of this whole service will create multiple workers which can horizontally scale the service.
func NewWorker(service *service.PruningService, settings *settings.SettingsObj, mgr taskmgr.TaskMgr) worker2.Worker {
	return &worker{service: service, settings: settings, taskmgr: mgr}
}
