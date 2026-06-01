package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/VarvaraKurakova/fleet-events-api/internal/messaging"
)

const (
	EventsExchangeName = "fleet.events"
	EventCreatedKey    = "event.created"
	AlertsQueueName    = "fleet.events.alerts"
)

type EventPublisher struct {
	connection *amqp.Connection
}

func NewEventPublisher(connection *amqp.Connection) *EventPublisher {
	return &EventPublisher{
		connection: connection,
	}
}

func (p *EventPublisher) PublishCreated(
	ctx context.Context,
	message messaging.EventCreatedMessage,
) error {
	channel, err := p.connection.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer channel.Close()

	if err := declareEventsTopology(channel); err != nil {
		return err
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal event created message: %w", err)
	}

	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = channel.PublishWithContext(
		publishCtx,
		EventsExchangeName,
		EventCreatedKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish event created message: %w", err)
	}

	return nil
}

func declareEventsTopology(channel *amqp.Channel) error {
	if err := channel.ExchangeDeclare(
		EventsExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare events exchange: %w", err)
	}

	if _, err := channel.QueueDeclare(
		AlertsQueueName,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare alerts queue: %w", err)
	}

	if err := channel.QueueBind(
		AlertsQueueName,
		EventCreatedKey,
		EventsExchangeName,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind alerts queue: %w", err)
	}

	return nil
}
