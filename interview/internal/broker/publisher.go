package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type Publisher interface {
	PublishResumeMessage(ctx context.Context, msg ResumeMessage) error
	Close()
}

type RabbitMQPublisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	queue    string
}

func NewRabbitMQPublisher(url, exchange, queue string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Объявляем exchange
	if err := ch.ExchangeDeclare(
		exchange, // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Объявляем очередь
	if _, err := ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Привязываем очередь к exchange
	if err := ch.QueueBind(
		queue,    // queue name
		queue,    // routing key
		exchange, // exchange
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &RabbitMQPublisher{
		conn:     conn,
		channel:  ch,
		exchange: exchange,
		queue:    queue,
	}, nil
}

func (p *RabbitMQPublisher) PublishResumeMessage(ctx context.Context, msg ResumeMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return p.channel.Publish(
		p.exchange, // exchange
		p.queue,    // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // make message persistent
			Timestamp:    time.Now(),
			MessageId:    msg.ID,
		},
	)
}

func (p *RabbitMQPublisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

// NullPublisher для тестов или когда RabbitMQ недоступен
type NullPublisher struct{}

func NewNullPublisher() *NullPublisher {
	return &NullPublisher{}
}

func (p *NullPublisher) PublishResumeMessage(ctx context.Context, msg ResumeMessage) error {
	log.Printf("NullPublisher: would publish message for resume %s", msg.ID)
	return nil
}

func (p *NullPublisher) Close() {
	// nothing to close
}
