package taskmgr

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/cenkalti/backoff/v4"
	"github.com/swagftw/gi"

	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/taskmgr"
	"audit-protocol/goutils/taskmgr/worker"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type RabbitmqTaskMgr struct {
	consumeConn *amqp.Connection
	publishConn *amqp.Connection
	settings    *settings.SettingsObj
}

type ConnectionType string

const (
	Consumer  ConnectionType = "consumer"
	Publisher ConnectionType = "publisher"
)

var _ taskmgr.TaskMgr = &RabbitmqTaskMgr{}

// NewRabbitmqTaskMgr returns a new rabbitmq task manager
func NewRabbitmqTaskMgr() *RabbitmqTaskMgr {
	settingsObj, err := gi.Invoke[*settings.SettingsObj]()
	if err != nil {
		log.WithError(err).Fatalf("failed to invoke settingsObj object")
	}

	consumeConn := new(amqp.Connection)
	publishConn := new(amqp.Connection)

	err = backoff.Retry(func() error {
		consumeConn, err = Dial(settingsObj)

		return err
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))
	if err != nil {
		log.WithError(err).Fatalf("failed to connect to rabbitmq")
	}

	err = backoff.Retry(func() error {
		publishConn, err = Dial(settingsObj)

		return err
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5))
	if err != nil {
		log.WithError(err).Fatalf("failed to connect to rabbitmq")
	}

	taskMgr := &RabbitmqTaskMgr{
		consumeConn: consumeConn,
		publishConn: publishConn,
		settings:    settingsObj,
	}

	if err := gi.Inject(taskMgr); err != nil {
		log.WithError(err).Fatalf("failed to inject dependencies")
	}

	log.Debug("rabbitmq task manager initialized")

	return taskMgr
}

// TODO: improve publishing logic with channel pooling
func (r *RabbitmqTaskMgr) Publish(ctx context.Context, workerType worker.Type, msg []byte) error {
	defer func() {
		// recover from panic
		if p := recover(); p != nil {
			log.Errorf("recovered from panic: %v", p)
		}
	}()

	channel, err := r.getChannel(workerType, Publisher)
	if err != nil {
		return err
	}

	defer func(channel *amqp.Channel) {
		err = channel.Close()
		if err != nil {
			log.WithError(err).Error("failed to close rabbitmq channel")
		}
	}(channel)

	exchange := r.getExchange(workerType)
	routingKey := r.getRoutingKeys(workerType)

	err = channel.Publish(
		exchange,
		routingKey[0],
		true,  // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         msg,
		},
	)
	if err != nil {
		log.WithError(err).
			WithField("exchange", exchange).
			WithField("routingKey", routingKey).
			Error("failed to publish rabbitmq message")
	}

	return err
}

// getChannel returns a channel from the connection
// this method is also used to create a new channel if channel is closed
func (r *RabbitmqTaskMgr) getChannel(workerType worker.Type, connectionType ConnectionType) (*amqp.Channel, error) {
	var conn *amqp.Connection

	switch connectionType {
	case Consumer:
		conn = r.consumeConn
	case Publisher:
		conn = r.publishConn
	}

	if conn == nil || conn.IsClosed() {
		log.Debug("rabbitmq connection is closed, reconnecting")
		var err error

		conn, err = Dial(r.settings)
		if err != nil {
			return nil, err
		}

		switch connectionType {
		case Consumer:
			r.consumeConn = conn
		case Publisher:
			r.publishConn = conn
		}
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Errorf("failed to open a channel on rabbitmq: %v", err)

		return nil, err
	}

	exchange := r.getExchange(workerType)

	err = channel.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
	if err != nil {
		log.Errorf("failed to declare an exchange on rabbitmq: %v", err)

		return nil, err
	}

	// declare the queue
	routingKeys := r.getRoutingKeys(workerType) // dag-pruning:task

	queue, err := channel.QueueDeclare(r.getQueue(workerType, ""), false, false, false, false, nil)

	if err != nil {
		log.Errorf("failed to declare a queue on rabbitmq: %v", err)

		return nil, err
	}

	// bind the queue to the exchange on the routing keys
	for _, routingKey := range routingKeys {
		err = channel.QueueBind(queue.Name, routingKey, exchange, false, nil)
		if err != nil {
			log.WithField("routingKey", routingKey).Errorf("failed to bind a queue on rabbitmq: %v", err)

			return nil, err
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

	channel, err := r.getChannel(workerType, Consumer)
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
		err = r.consumeConn.Close()
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

	connCloseChan = r.consumeConn.NotifyClose(connCloseChan)
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
	case worker.TypeEventDetectorWorker:
		return fmt.Sprintf("%s%s", r.settings.Rabbitmq.Setup.Core.EventDetectorExchange, r.settings.PoolerNamespace)
	default:
		return ""
	}
}

func (r *RabbitmqTaskMgr) getQueue(workerType worker.Type, suffix string) string {
	switch workerType {
	case worker.TypePayloadCommitWorker:
		return fmt.Sprintf("%s%s:%s", r.settings.Rabbitmq.Setup.PayloadCommit.QueueNamePrefix, r.settings.PoolerNamespace, r.settings.InstanceId)
	case worker.TypeEventDetectorWorker:
		return fmt.Sprintf("%s%s:%s", r.settings.Rabbitmq.Setup.EventDetector.QueueNamePrefix, r.settings.PoolerNamespace, r.settings.InstanceId)
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
	case worker.TypeEventDetectorWorker:
		return []string{
			fmt.Sprintf("%s%s:%s%s", r.settings.Rabbitmq.Setup.EventDetector.RoutingKeyPrefix, r.settings.PoolerNamespace, r.settings.InstanceId, taskmgr.SnapshotSubmitted),
		}
	default:
		return nil
	}
}

func (r *RabbitmqTaskMgr) Shutdown(ctx context.Context) error {
	if r.consumeConn == nil || r.consumeConn.IsClosed() {
		return nil
	}

	if err := r.consumeConn.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
		log.Errorf("failed to close connection on rabbitmq: %v", err)

		return err
	}

	if r.publishConn == nil || r.publishConn.IsClosed() {
		return nil
	}

	if err := r.publishConn.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
		log.WithError(err).Error("failed to close connection on rabbitmq")

		return err
	}

	return nil
}
