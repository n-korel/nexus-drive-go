package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/n-korel/nexus-drive-go/shared/contracts"
	"github.com/n-korel/nexus-drive-go/shared/retry"
	"github.com/n-korel/nexus-drive-go/shared/tracing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TripExchange = "trip"
	DeadLetterExchange = "dlx"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create channel: %v", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		channel: ch,
	}

	if err := rmq.setupExchangesAndQueues(); err != nil {
		// Close RabbitMQ if setup exchanges ans queues fails
		rmq.Close()
		return nil, fmt.Errorf("failed to setup exchanges and queues: %v", err)
	}

	return rmq, nil
}

type MessageHandler func(context.Context, amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	err := r.channel.Qos(
		1,     // prefetchCount
		0,     // prefetchSize
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %v", err)
	}

	msgs, err := r.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	
	if err != nil {
		return err
	}


	go func() {
		for msg := range msgs {
			if err := tracing.TracedConsumer(msg, func(ctx context.Context, d amqp.Delivery) error {
				
				log.Printf("Received a message: %s", msg.Body)
				
				cfg := retry.DefaultConfig()
				err := retry.WithBackoff(ctx, cfg, func() error {
					return handler(ctx, d)
				})


				if err != nil {
					log.Printf("Message processing failed after %d retries for message ID: %s, err: %v", cfg.MaxRetries, d.MessageId, err)

					// Add failure context before sending to Dead Letter Queue (DLQ)
					headers := amqp.Table{}
					if d.Headers != nil {
						headers = d.Headers
					}

					headers["x-death-reason"] = err.Error()
					headers["x-origin-exchange"] = d.Exchange
					headers["x-original-routing-key"] = d.RoutingKey
					headers["x-retry-count"] = cfg.MaxRetries
					d.Headers = headers

					// Reject without requeue - message will go to Dead Letter Queue (DLQ)
					_ = d.Reject(false)
					return err
				}
				
				// Only Ack if handler succeeds
				if ackErr := msg.Ack(false); ackErr != nil {
					log.Printf("ERROR: Failed to Ack message: %v. Message body: %s", ackErr, msg.Body)
				}

				return nil
			}); err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message contracts.AmqpMessage) error {
	log.Printf("Publishing message with routing key: %s", routingKey)
	
	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType: "application/json",
		Body: jsonMsg,
	}
	
	return tracing.TracedPublisher(ctx, TripExchange, routingKey, msg, r.publish)
}


func (r *RabbitMQ) publish(ctx context.Context, exchange, routingKey string, msg amqp.Publishing) error {
	return r.channel.PublishWithContext(ctx,
			exchange,      // exchange
			routingKey, // routing key
			false,   // mandatory
			false,   // immediate
			msg,	
		)
}

func (r *RabbitMQ) setupDeadLetterExchange() error {
	// Declare Dead Letter Queue (DLQ)
	err := r.channel.ExchangeDeclare(
		DeadLetterExchange,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter exchange: %v", err)
	}

	q, err := r.channel.QueueDeclare(
		DeadLetterQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare dead letter queue: %v", err)
	}

	// Bind the queue to the exchange with wildcard routing key
	err = r.channel.QueueBind(
		q.Name,
		"#", // wildcard routing key to catch all messages
		DeadLetterExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind dead letter queue: %v", err)
	}

	return nil
}


func (r *RabbitMQ) setupExchangesAndQueues() error {
	// Dead Letter Queue (DLQ) exchange and queue
	if err := r.setupDeadLetterExchange(); err != nil {
		return err
	}

	err := r.channel.ExchangeDeclare(
		TripExchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %s: %v", TripExchange, err)
	}

	if err := r.declareAndBindQueue(
		FindAvailableDriversQueue,
		[]string{
			contracts.TripEventCreated, contracts.TripEventDriverNotInterested,
		},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		DriverCmdTripRequestQueue,
		[]string{
			contracts.DriverCmdTripRequest,
		},
		TripExchange,
	); err != nil {
		return err
	}


	if err := r.declareAndBindQueue(
		DriverTripResponseQueue,
		[]string{
			contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline,
		},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyDriverNoDriversFoundQueue,
		[]string{
			contracts.TripEventNoDriversFound,
		},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyDriverAssignQueue,
		[]string{
			contracts.TripEventDriverAssigned,
		},
		TripExchange,
	); err != nil {
		return err
	}

		if err := r.declareAndBindQueue(
		PaymentTripResponseQueue,
		[]string{contracts.PaymentCmdCreateSession},
		TripExchange,
	); err != nil {
		return err
	}

	if err := r.declareAndBindQueue(
		NotifyPaymentSessionCreatedQueue,
		[]string{contracts.PaymentEventSessionCreated},
		TripExchange,
	); err != nil {
		return err
	}
	
	if err := r.declareAndBindQueue(
		NotifyPaymentSuccessQueue,
		[]string{contracts.PaymentEventSuccess},
		TripExchange,
	); err != nil {
		return err
	}
	
	return nil
}

func (r *RabbitMQ) declareAndBindQueue(queueName string, messageTypes []string, exchange string) error {
	// Add dead letter configuration
	args := amqp.Table{
		"x-dead-letter-exchange": DeadLetterExchange,
	}

	q, err := r.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		args,       // arguments with Dead Letter Queue (DLQ) config
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range messageTypes {
		if err := r.channel.QueueBind(
			q.Name,   // queue name
			msg,      // routing key
			exchange, // exchange
			false,
			nil,
		); err != nil {
			return fmt.Errorf("failed to bind queue to %s: %v", queueName, err)
		}
	}

	return nil
}

func (r *RabbitMQ) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
	if r.channel != nil {
		r.channel.Close()
	}
}