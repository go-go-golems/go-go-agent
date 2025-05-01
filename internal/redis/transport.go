package redis

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

// TransportType defines the type of Redis transport to use
type TransportType string

const (
	// TransportStream uses Redis Streams
	TransportStream TransportType = "stream"
	// TransportPubSub uses Redis Pub/Sub
	TransportPubSub TransportType = "pubsub"
)

// Transport defines the interface for message transport mechanisms
type Transport interface {
	// CreateSubscriber creates a new message subscriber
	CreateSubscriber(config RouterConfig, logger watermill.LoggerAdapter) (message.Subscriber, error)
	// CreatePublisher creates a new message publisher (if needed)
	CreatePublisher(config RouterConfig, logger watermill.LoggerAdapter) (message.Publisher, error)
	// GetTopicName returns the appropriate topic name for the transport
	GetTopicName(config RouterConfig) string
}
