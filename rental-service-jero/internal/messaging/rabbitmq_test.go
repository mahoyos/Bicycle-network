package messaging

import (
	"testing"
)

func TestNewRabbitMQManager(t *testing.T) {
	m := NewRabbitMQManager("amqp://guest:guest@localhost:5672/")
	if m.url != "amqp://guest:guest@localhost:5672/" {
		t.Errorf("unexpected url: %s", m.url)
	}
	if m.conn != nil {
		t.Error("connection should be nil before Connect()")
	}
	if m.channel != nil {
		t.Error("channel should be nil before Connect()")
	}
}

func TestIsConnected_BeforeConnect(t *testing.T) {
	m := NewRabbitMQManager("amqp://localhost:5672/")
	if m.IsConnected() {
		t.Error("should not be connected before Connect()")
	}
}

func TestClose_BeforeConnect(t *testing.T) {
	m := NewRabbitMQManager("amqp://localhost:5672/")
	// Should not panic when closing without connecting
	m.Close()
}

func TestChannel_BeforeConnect(t *testing.T) {
	m := NewRabbitMQManager("amqp://localhost:5672/")
	if m.Channel() != nil {
		t.Error("channel should be nil before Connect()")
	}
}

func TestNewConsumer_Constructor(t *testing.T) {
	repo := &mockBikeRepo{}
	c := NewConsumer(nil, repo)
	if c.bikeRepo == nil {
		t.Error("bikeRepo should be set")
	}
	if c.channel != nil {
		t.Error("channel should be nil when passed nil")
	}
}
