package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/streadway/amqp"
)

// Conn -
type Conn struct {
	Channel *amqp.Channel
}

// GetConn -
func GetConn(rabbitURL string) (Conn, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return Conn{}, err
	}

	ch, err := conn.Channel()
	return Conn{
		Channel: ch,
	}, err
}

// Publish -
func (conn Conn) Publish(routingKey string, data []byte) error {
	return conn.Channel.Publish(
		// exchange - yours may be different
		"events",
		routingKey,
		// mandatory - we don't care if there I no queue
		false,
		// immediate - we don't care if there is no consumer on the queue
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         data,
			DeliveryMode: amqp.Persistent,
		})
}

// StartConsumer -
func (conn Conn) StartConsumer(
	queueName,
	exchange,
	routingKey string,
	handler func(d amqp.Delivery) bool,
	concurrency int) error {

	// create the queue if it doesn't already exist
	_, err := conn.Channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	// bind the queue to the routing key
	err = conn.Channel.QueueBind(queueName, routingKey, exchange, false, nil)
	if err != nil {
		return err
	}

	// prefetch 4x as many messages as we can handle at once
	prefetchCount := concurrency
	err = conn.Channel.Qos(prefetchCount, 0, false)
	if err != nil {
		return err
	}

	msgs, err := conn.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return err
	}

	// create a goroutine for the number of concurrent threads requested
	for i := 0; i < concurrency; i++ {
		log.Infof("Processing messages on go-routine %v\n", i)
		go func() {
			for msg := range msgs {
				// if tha handler returns true then ACK, else NACK
				// the message back into the rabbit queue for
				// another round of processing
				if handler(msg) {
					err := msg.Ack(false)
					if err != nil {
						log.Fatalf("CRITICAL! Ack failed for message %+v with error %+v", msg, err)
					}
				} else {
					err := msg.Nack(false, true)
					if err != nil {
						log.Fatalf("CRITICAL!Ack failed for message %+v with error %+v", msg, err)
					}
				}
			}
			log.Fatalf("Rabbit consumer closed - critical Error")
			os.Exit(1)
		}()
	}
	return nil
}
