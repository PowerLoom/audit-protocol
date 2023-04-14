package taskmgr

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/taskmgr"
	"audit-protocol/goutils/taskmgr/worker"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

const (
	taskSuffix string = "task"
	dlxSuffix  string = "dlx"
)

type Config struct {
	queueName string
}

type RabbitmqTaskMgr struct {
	conn     *amqp.Connection
	settings *settings.SettingsObj
}

func NewRabbitmqTaskMgr(settings *settings.SettingsObj) taskmgr.TaskMgr {
	return &RabbitmqTaskMgr{
		conn:     Dial(settings),
		settings: settings,
	}
}

func (r RabbitmqTaskMgr) Publish(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

// getChannel returns a channel from the connection
// this method is also used to create a new channel if channel is closed
func (r RabbitmqTaskMgr) getChannel(workerType worker.Type) (*amqp.Channel, error) {
	channel, err := r.conn.Channel()
	if err != nil {
		log.Errorf("Failed to open a channel on rabbitmq: %v", err)

		return nil, taskmgr.ErrConsumerInitFailed
	}

	exchange := r.getExchange(workerType)
	err = channel.ExchangeDeclare(exchange, "direct", true, false, false, false, nil)
	if err != nil {
		log.Errorf("Failed to declare an exchange on rabbitmq: %v", err)

		return nil, taskmgr.ErrConsumerInitFailed
	}

	// dead letter exchange
	dlxExchange := r.settings.Rabbitmq.Setup.Core.DLX
	err = channel.ExchangeDeclare(dlxExchange, "direct", true, false, false, false, nil)
	if err != nil {
		log.Errorf("Failed to declare an exchange on rabbitmq: %v", err)

		return nil, taskmgr.ErrConsumerInitFailed
	}

	// declare the queue
	dlxRoutingKey := r.getRoutingKey(workerType, taskSuffix) // dag-pruning:task
	queue, err := channel.QueueDeclare(r.getQueue(workerType, taskSuffix), true, false, false, false, map[string]interface{}{
		"x-dead-letter-exchange":    dlxExchange,
		"x-dead-letter-routing-key": dlxRoutingKey,
	})
	if err != nil {
		log.Errorf("Failed to declare a queue on rabbitmq: %v", err)

		return nil, taskmgr.ErrConsumerInitFailed
	}

	taskRoutingKey := r.getRoutingKey(workerType, taskSuffix)
	err = channel.QueueBind(queue.Name, taskRoutingKey, exchange, false, nil)
	if err != nil {
		log.Errorf("Failed to bind a queue on rabbitmq: %v", err)

		return nil, taskmgr.ErrConsumerInitFailed
	}

	err = channel.QueueBind(queue.Name, dlxRoutingKey, dlxExchange, false, nil)
	if err != nil {
		log.Errorf("Failed to bind a queue on rabbitmq: %v", err)

		return nil, taskmgr.ErrConsumerInitFailed
	}

	return channel, nil
}

func (r RabbitmqTaskMgr) Consume(ctx context.Context, workerType worker.Type, msgChan chan taskmgr.TaskHandler, errChan chan error) error {
	channel, err := r.getChannel(workerType)
	if err != nil {
		return err
	}

	defer func(channel *amqp.Channel) {
		err = channel.Close()
		if err != nil && err != amqp.ErrClosed {
			log.Errorf("Failed to close channel on rabbitmq: %v", err)
		}
	}(channel)

	defer func() {
		err = r.conn.Close()
		if err != nil && err != amqp.ErrClosed {
			log.Errorf("Failed to close connection on rabbitmq: %v", err)
		}
	}()

	queueName := r.getQueue(workerType, taskSuffix)
	// consume messages
	msgs, err := channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("Failed to register a consumer on rabbitmq: %v", err)

		return err
	}

	log.Infof("RabbitmqTaskMgr: consuming messages from queue %s", queueName)

	forever := make(chan *amqp.Error)

	forever = channel.NotifyClose(forever)

	go func() {
		for msg := range msgs {
			log.Debug(msg.Headers)

			log.Infof("received new message")

			task := taskmgr.Task{Msg: msg}

			msgChan <- task
		}
	}()

	err = <-forever
	if err != nil {
		log.Errorf("RabbitmqTaskMgr: connection closed while consuming messages from queue %s: %s", queueName, err)
	}

	// send back error due to rabbitmq channel closed
	errChan <- err

	return nil
}

func Dial(config *settings.SettingsObj) *amqp.Connection {
	rabbitmqConfig := config.Rabbitmq

	url := fmt.Sprintf("amqp://%s:%s@%s/", rabbitmqConfig.User, rabbitmqConfig.Password, net.JoinHostPort(rabbitmqConfig.Host, strconv.Itoa(rabbitmqConfig.Port)))

	// TODO: remove this before committing
	conn, err := amqp.DialConfig(url, amqp.Config{Heartbeat: 10 * time.Hour})
	if err != nil {
		log.Panicf("Failed to connect to RabbitMQ: %v", err)
	}

	return conn
}

func (r RabbitmqTaskMgr) getExchange(workerType worker.Type) string {
	switch workerType {
	case worker.TypePruningServiceWorker:
		return r.settings.Rabbitmq.Setup.Core.Exchange
	default:
		return ""
	}
}

func (r RabbitmqTaskMgr) getQueue(workerType worker.Type, suffix string) string {
	switch workerType {
	case worker.TypePruningServiceWorker:
		return r.settings.Rabbitmq.Setup.Queues.DagPruning.QueueNamePrefix + suffix
	default:
		return ""
	}
}

func (r RabbitmqTaskMgr) getRoutingKey(workerType worker.Type, suffix string) string {
	switch workerType {
	case worker.TypePruningServiceWorker:
		return r.settings.Rabbitmq.Setup.Queues.DagPruning.RoutingKeyPrefix + suffix
	default:
		return ""
	}
}
