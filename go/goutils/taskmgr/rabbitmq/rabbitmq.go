package taskmgr

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/taskmgr"
	"audit-protocol/goutils/taskmgr/worker"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type RabbitmqTaskMgr struct {
	conn     *amqp.Connection
	settings *settings.SettingsObj
}

// NewRabbitmqTaskMgr returns a new rabbitmq task manager
func NewRabbitmqTaskMgr(settingsObj *settings.SettingsObj, conn *amqp.Connection) taskmgr.TaskMgr {
	taskMgr := &RabbitmqTaskMgr{
		conn:     conn,
		settings: settingsObj,
	}

	log.Debug("rabbitmq task manager initialized")

	return taskMgr
}

func (r *RabbitmqTaskMgr) Publish(ctx context.Context) error {
	// TODO implement me
	panic("implement me")
}

// getChannel returns a channel from the connection
// this method is also used to create a new channel if channel is closed
func (r *RabbitmqTaskMgr) getChannel(workerType worker.Type) (*amqp.Channel, error) {
	if r.conn == nil || r.conn.IsClosed() {
		log.Debug("rabbitmq connection is closed, reconnecting")
		var err error

		r.conn, err = Dial(r.settings)
		if err != nil {
			return nil, err
		}
	}

	channel, err := r.conn.Channel()
	if err != nil {
		log.Errorf("failed to open a channel on rabbitmq: %v", err)

		return nil, taskmgr.ErrConsumerInitFailed
	}

	exchange := r.getExchange(workerType)

	err = channel.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
	if err != nil {
		log.Errorf("failed to declare an exchange on rabbitmq: %v", err)

		return nil, taskmgr.ErrConsumerInitFailed
	}

	// declare the queue
	routingKeys := r.getRoutingKeys(workerType) // dag-pruning:task

	queue, err := channel.QueueDeclare(r.getQueue(workerType, ""), false, false, false, false, nil)

	if err != nil {
		log.Errorf("failed to declare a queue on rabbitmq: %v", err)

		return nil, taskmgr.ErrConsumerInitFailed
	}

	// bind the queue to the exchange on the routing keys
	for _, routingKey := range routingKeys {
		err = channel.QueueBind(queue.Name, routingKey, exchange, false, nil)
		if err != nil {
			log.WithField("routingKey", routingKey).Errorf("failed to bind a queue on rabbitmq: %v", err)

			return nil, taskmgr.ErrConsumerInitFailed
		}
	}

	return channel, nil
}

// Consume consumes messages from the queue
func (r *RabbitmqTaskMgr) Consume(ctx context.Context, workerType worker.Type, msgChan chan taskmgr.TaskHandler) error {
	defer func() {
		// recover from panic
		if p := recover(); p != nil {
			log.Errorf("recovered from panic: %v", p)
		}
	}()
	channel, err := r.getChannel(workerType)
	if err != nil {
		return err
	}

	defer func(channel *amqp.Channel) {
		err = channel.Close()
		if err != nil && !errors.Is(err, amqp.ErrClosed) {
			log.Errorf("failed to close channel on rabbitmq: %v", err)
		}
	}(channel)

	defer func() {
		err = r.conn.Close()
		if err != nil && !errors.Is(err, amqp.ErrClosed) {
			log.Errorf("failed to close connection on rabbitmq: %v", err)
		}
	}()

	queueName := r.getQueue(workerType, taskmgr.TaskSuffix)
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
		log.Errorf("failed to register a consumer on rabbitmq: %v", err)

		return err
	}

	log.Infof("RabbitmqTaskMgr: consuming messages from queue %s", queueName)

	connCloseChan := make(chan *amqp.Error, 1)
	channelCloseChan := make(chan *amqp.Error, 1)

	connCloseChan = r.conn.NotifyClose(connCloseChan)
	channelCloseChan = channel.NotifyClose(channelCloseChan)

	go func() {
		for msg := range msgs {
			log.Debug(msg.Headers)

			log.Infof("received new message")

			task := &taskmgr.Task{Msg: msg, Topic: msg.RoutingKey}

			msgChan <- task
		}
	}()

	select {
	case err = <-connCloseChan:
		log.WithError(err).Error("connection closed")

		return err
	case err = <-channelCloseChan:
		log.WithError(err).Error("channel closed")

		return err
	}
}

func Dial(config *settings.SettingsObj) (*amqp.Connection, error) {
	rabbitmqConfig := config.Rabbitmq

	url := fmt.Sprintf("amqp://%s:%s@%s/", rabbitmqConfig.User, rabbitmqConfig.Password, net.JoinHostPort(rabbitmqConfig.Host, strconv.Itoa(rabbitmqConfig.Port)))

	conn, err := amqp.Dial(url)
	if err != nil {
		log.WithError(err).Error("failed to connect to RabbitMQ")

		return nil, err
	}

	return conn, nil
}

func (r *RabbitmqTaskMgr) getExchange(workerType worker.Type) string {
	switch workerType {
	case worker.TypePayloadCommitWorker:
		return fmt.Sprintf("%s%s", r.settings.Rabbitmq.Setup.Core.CommitPayloadExchange, r.settings.PoolerNamespace)
	default:
		return ""
	}
}

func (r *RabbitmqTaskMgr) getQueue(workerType worker.Type, suffix string) string {
	switch workerType {
	case worker.TypePayloadCommitWorker:
		return fmt.Sprintf("%s%s:%s", r.settings.Rabbitmq.Setup.PayloadCommit.QueueNamePrefix, r.settings.PoolerNamespace, r.settings.InstanceId)
	default:
		return ""
	}
}

// getRoutingKeys returns the routing key(s) for the given worker type
func (r *RabbitmqTaskMgr) getRoutingKeys(workerType worker.Type) []string {
	switch workerType {
	case worker.TypePayloadCommitWorker:
		return []string{
			fmt.Sprintf("%s%s:%s%s", r.settings.Rabbitmq.Setup.PayloadCommit.RoutingKeyPrefix, r.settings.PoolerNamespace, r.settings.InstanceId, taskmgr.FinalizedSuffix),
			fmt.Sprintf("%s%s:%s%s", r.settings.Rabbitmq.Setup.PayloadCommit.RoutingKeyPrefix, r.settings.PoolerNamespace, r.settings.InstanceId, taskmgr.DataSuffix),
		}
	default:
		return nil
	}
}

func (r *RabbitmqTaskMgr) Shutdown(ctx context.Context) error {
	if r.conn == nil || r.conn.IsClosed() {
		return nil
	}

	if err := r.conn.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
		log.Errorf("failed to close connection on rabbitmq: %v", err)

		return err
	}

	return nil
}
