package worker

import (
	"context"

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
		err := w.taskmgr.Consume(context.Background(), worker2.TypePruningServiceWorker, taskChan)
		if err != nil {
			log.Fatalf("Failed to consume the message: %v", err)
		}
	}()

	// create a wait group to wait for all the tasks to finish.
	swg := sizedwaitgroup.New(w.settings.PruningServiceSettings.Concurrency)

	for {
		swg.Add()
		go func() {
			// as task chan is buffered channel, we can use it as a semaphore to limit the number of concurrent tasks.
			taskHandler := <-taskChan

			msgBody := taskHandler.GetBody()

			err := w.service.Run(msgBody)
			if err != nil {
				log.Errorf("Failed to run the task: %v", err)

				err = taskHandler.Nack(false)
				if err != nil {
					log.Errorf("Failed to nack the message: %v", err)
				}

				// TODO: put the msg in dead letter queue.
			}

			err = taskHandler.Ack()
			if err != nil {
				log.Errorf("Failed to ack the message: %v", err)
			}

			swg.Done()
		}()

		// wait till all the tasks previous are finished.
		swg.Wait()
	}
}

// NewWorker creates a new worker listening for pruning tasks published by service responsible for creating segments
// a single worker has capability to run multiple tasks concurrently using go routines.
// running multiple instances of this whole service will create multiple workers which can horizontally scale the service.
func NewWorker(service *service.PruningService, settings *settings.SettingsObj, mgr taskmgr.TaskMgr) worker2.Worker {
	return &worker{service: service, settings: settings, taskmgr: mgr}
}

// TODO: create dead letter queue worker
