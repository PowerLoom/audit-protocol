package taskmgr

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"audit-protocol/goutils/settings"
	"audit-protocol/goutils/taskmgr"
	"audit-protocol/goutils/taskmgr/worker"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
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

func (r RabbitmqTaskMgr) Consume(ctx context.Context, workerType worker.Type, msgChan chan taskmgr.TaskHandler) error {
	channel, err := r.conn.Channel()
	if err != nil {
		log.Errorf("Failed to open a channel on rabbitmq: %v", err)

		return taskmgr.ErrConsumerInitFailed
	}

	defer func(channel *amqp.Channel) {
		err = channel.Close()
		if err != nil {
			log.Errorf("Failed to close channel on rabbitmq: %v", err)
		}
	}(channel)

	exchange := r.getExchange(workerType)
	err = channel.ExchangeDeclare(exchange, "direct", true, false, false, false, nil)
	if err != nil {
		log.Errorf("Failed to declare an exchange on rabbitmq: %v", err)

		return taskmgr.ErrConsumerInitFailed
	}

	// declare the queue
	queue, err := channel.QueueDeclare(r.getQueue(workerType), true, false, false, false, nil)
	if err != nil {
		log.Errorf("Failed to declare a queue on rabbitmq: %v", err)

		return taskmgr.ErrConsumerInitFailed
	}

	err = channel.QueueBind(queue.Name, r.getRoutingKey(workerType), exchange, false, nil)
	if err != nil {
		log.Errorf("Failed to bind a queue on rabbitmq: %v", err)

		return taskmgr.ErrConsumerInitFailed
	}

	// consume messages
	msgs, err := channel.Consume(
		queue.Name,
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

	log.Infof("RabbitmqTaskMgr: consuming messages from queue %s", queue.Name)

	forever := make(chan bool)

	for msg := range msgs {
		log.Infof("received new message")

		task := taskmgr.Task{Msg: msg}

		msgChan <- task
	}

	<-forever

	return nil
}

func Dial(config *settings.SettingsObj) *amqp.Connection {
	rabbitmqConfig := config.Rabbitmq

	url := fmt.Sprintf("amqp://%s:%s@%s/", rabbitmqConfig.User, rabbitmqConfig.Password, net.JoinHostPort(rabbitmqConfig.Host, strconv.Itoa(rabbitmqConfig.Port)))

	conn, err := amqp.Dial(url)
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

func (r RabbitmqTaskMgr) getQueue(workerType worker.Type) string {
	switch workerType {
	case worker.TypePruningServiceWorker:
		return r.settings.Rabbitmq.Setup.Queues.DagPruning.QueueName
	default:
		return ""
	}
}

func (r RabbitmqTaskMgr) getRoutingKey(workerType worker.Type) string {
	switch workerType {
	case worker.TypePruningServiceWorker:
		return r.settings.Rabbitmq.Setup.Queues.DagPruning.RoutingKey
	default:
		return ""
	}
}
