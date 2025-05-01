package redis

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
)

// StreamTransport implements the Transport interface using Redis Streams
type StreamTransport struct{}

// NewStreamTransport creates a new StreamTransport
func NewStreamTransport() *StreamTransport {
	return &StreamTransport{}
}

// CreateSubscriber creates a new Redis Stream subscriber
func (t *StreamTransport) CreateSubscriber(config RouterConfig, logger watermill.LoggerAdapter) (message.Subscriber, error) {
	subscriber, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:          config.redisClient,
			Unmarshaller:    CustomEventUnmarshaller{},
			ConsumerGroup:   config.ConsumerGroup,
			Consumer:        config.ConsumerName,
			BlockTime:       config.BlockTime,
			MaxIdleTime:     config.MaxIdleTime,
			NackResendSleep: config.NackResendSleep,
		},
		logger,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Redis Stream subscriber")
	}
	return subscriber, nil
}

// CreatePublisher creates a new Redis Stream publisher
func (t *StreamTransport) CreatePublisher(config RouterConfig, logger watermill.LoggerAdapter) (message.Publisher, error) {
	// For now, we don't need a publisher
	return nil, nil
}

// GetTopicName returns the stream name as the topic
func (t *StreamTransport) GetTopicName(config RouterConfig) string {
	return config.StreamName
}
