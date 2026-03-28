package messaging

import (
	"encoding/json"
	"fmt"
	"log"
)

type BicycleEvent struct {
	BikeID string `json:"bike_id"`
	Action string `json:"action"`
}

func (r *RabbitMQClient) PublishBicycleReturned(bicycleID string) error {
	event := BicycleEvent{
		BikeID: bicycleID,
		Action: "RETURNED",
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := r.Publish(body); err != nil {
		log.Printf("Failed to publish RETURNED event for bike %s: %v", bicycleID, err)
		return err
	}

	log.Printf("Published action=RETURNED, bike_id=%s", bicycleID)
	return nil
}
