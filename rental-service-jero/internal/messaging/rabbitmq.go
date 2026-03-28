package messaging

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQManager struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

func NewRabbitMQManager(url string) *RabbitMQManager {
	return &RabbitMQManager{url: url}
}

func (m *RabbitMQManager) Connect() error {
	var err error
	for i := 0; i < 5; i++ {
		m.conn, err = amqp.Dial(m.url)
		if err == nil {
			m.channel, err = m.conn.Channel()
			if err == nil {
				log.Println("Connected to RabbitMQ")
				return nil
			}
			m.conn.Close()
		}
		log.Printf("RabbitMQ connection attempt %d/5 failed: %v", i+1, err)
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("failed to connect to RabbitMQ after 5 attempts: %w", err)
}

func (m *RabbitMQManager) Channel() *amqp.Channel {
	return m.channel
}

func (m *RabbitMQManager) IsConnected() bool {
	return m.conn != nil && !m.conn.IsClosed() &&
		m.channel != nil && !m.channel.IsClosed()
}

func (m *RabbitMQManager) Close() {
	if m.channel != nil {
		if err := m.channel.Close(); err != nil {
			log.Printf("Error closing RabbitMQ channel: %v", err)
		}
	}
	if m.conn != nil {
		if err := m.conn.Close(); err != nil {
			log.Printf("Error closing RabbitMQ connection: %v", err)
		}
	}
	log.Println("RabbitMQ connection closed")
}
