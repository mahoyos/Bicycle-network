package messaging

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EventHandler is called when a bike lifecycle event is received
type EventHandler func(event BicycleEvent)

// StartConsumer declares a queue for the rental service, binds it to the
// bike_lifecycle_events fanout exchange, and processes incoming messages.
// It runs in a goroutine and calls handler for each DELETED event.
func (r *RabbitMQClient) StartConsumer(handler EventHandler) error {
	// Declare a queue exclusive to the rental service
	q, err := r.channel.QueueDeclare(
		"bike_lifecycle_events.rental", // queue name
		true,                           // durable
		false,                          // auto-delete
		false,                          // exclusive
		false,                          // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	// Bind queue to the fanout exchange
	err = r.channel.QueueBind(
		q.Name,     // queue
		"",         // routing key (ignored for fanout)
		r.exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := r.channel.Consume(
		q.Name, // queue
		"",     // consumer tag
		false,  // auto-ack (false = manual ack for reliability)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		log.Println("RabbitMQ consumer started, listening for bike lifecycle events")
		for msg := range msgs {
			processMessage(msg, handler)
		}
		log.Println("RabbitMQ consumer stopped")
	}()

	return nil
}

func processMessage(msg amqp.Delivery, handler EventHandler) {
	var event BicycleEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		msg.Ack(false) // Ack bad messages to avoid requeue loop
		return
	}

	log.Printf("Received event: action=%s, bike_id=%s", event.Action, event.BikeID)

	if event.Action == "DELETED" {
		handler(event)
	}
	// Ignore CREATED, RETURNED, and any other actions

	msg.Ack(false)
}
