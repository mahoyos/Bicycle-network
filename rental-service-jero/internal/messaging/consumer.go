package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/bicycle-network/rental-service/internal/repository"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	BikeLifecycleExchange = "bike_lifecycle_events"
	RentalBikeQueue       = "rental_bike_lifecycle"
)

type BikeLifecycleEvent struct {
	BikeID string `json:"bike_id"`
	Action string `json:"action"`
}

type Consumer struct {
	channel  *amqp.Channel
	bikeRepo repository.BikeRepository
}

func NewConsumer(channel *amqp.Channel, bikeRepo repository.BikeRepository) *Consumer {
	return &Consumer{
		channel:  channel,
		bikeRepo: bikeRepo,
	}
}

func (c *Consumer) Setup() error {
	err := c.channel.ExchangeDeclare(
		BikeLifecycleExchange,
		"fanout",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	_, err = c.channel.QueueDeclare(
		RentalBikeQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	err = c.channel.QueueBind(
		RentalBikeQueue,
		"",                    // routing key (ignored for fanout)
		BikeLifecycleExchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Printf("Consumer setup: queue '%s' bound to exchange '%s'", RentalBikeQueue, BikeLifecycleExchange)
	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		RentalBikeQueue,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	log.Printf("Started consuming from queue '%s'", RentalBikeQueue)

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer stopped: context cancelled")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Consumer channel closed")
				return nil
			}
			c.processMessage(ctx, msg)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	if err := c.HandleEvent(ctx, msg.Body); err != nil {
		log.Printf("Error handling event: %v", err)
		msg.Nack(false, false)
		return
	}
	msg.Ack(false)
}

func (c *Consumer) HandleEvent(ctx context.Context, body []byte) error {
	var event BikeLifecycleEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}

	bikeID, err := uuid.Parse(event.BikeID)
	if err != nil {
		return fmt.Errorf("invalid bike_id '%s': %w", event.BikeID, err)
	}

	switch event.Action {
	case "CREATED":
		if err := c.bikeRepo.Upsert(ctx, bikeID); err != nil {
			return fmt.Errorf("failed to upsert bike %s: %w", bikeID, err)
		}
		log.Printf("Bike registered: %s", bikeID)
	case "DELETED":
		if err := c.bikeRepo.Delete(ctx, bikeID); err != nil {
			return fmt.Errorf("failed to delete bike %s: %w", bikeID, err)
		}
		log.Printf("Bike removed: %s", bikeID)
	default:
		log.Printf("Unknown action '%s' for bike %s, ignoring", event.Action, bikeID)
	}

	return nil
}
