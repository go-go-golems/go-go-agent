package redis

import (
	"context"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// PubSubTransport implements the Transport interface using Redis Pub/Sub
type PubSubTransport struct{}

// NewPubSubTransport creates a new PubSubTransport
func NewPubSubTransport() *PubSubTransport {
	return &PubSubTransport{}
}

// CreateSubscriber creates a new Redis Pub/Sub subscriber
// Note: This implementation expects a zerolog.Logger, not watermill.LoggerAdapter
func (t *PubSubTransport) CreateSubscriber(config RouterConfig, logger watermill.LoggerAdapter) (message.Subscriber, error) {
	return NewPubSubSubscriber(
		PubSubSubscriberConfig{
			Client:       config.redisClient,
			Unmarshaller: CustomEventUnmarshaller{}, // Use the same unmarshaller as Stream
		},
		logger, // Pass the zerolog.Logger directly
	)
}

// CreatePublisher creates a new Redis Pub/Sub publisher
func (t *PubSubTransport) CreatePublisher(config RouterConfig, logger watermill.LoggerAdapter) (message.Publisher, error) {
	// For now, we don't need a publisher
	return nil, nil
}

// GetTopicName returns the topic pattern for Pub/Sub
func (t *PubSubTransport) GetTopicName(config RouterConfig) string {
	return config.TopicPattern
}

// PubSubSubscriber implements message.Subscriber for Redis Pub/Sub
type PubSubSubscriber struct {
	config        PubSubSubscriberConfig
	logger        watermill.LoggerAdapter
	closing       chan struct{}
	subscribersWg sync.WaitGroup
	closed        bool
	closeMutex    sync.Mutex
}

// PubSubSubscriberConfig holds configuration for the Redis Pub/Sub subscriber
type PubSubSubscriberConfig struct {
	Client       redis.UniversalClient
	Unmarshaller CustomEventUnmarshaller // Use the same unmarshaller
}

// NewPubSubSubscriber creates a new PubSubSubscriber
func NewPubSubSubscriber(config PubSubSubscriberConfig, logger watermill.LoggerAdapter) (*PubSubSubscriber, error) {
	return &PubSubSubscriber{
		config:  config,
		logger:  logger,
		closing: make(chan struct{}),
	}, nil
}

// Subscribe subscribes to Redis Pub/Sub messages
func (s *PubSubSubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	s.closeMutex.Lock()
	if s.closed {
		s.closeMutex.Unlock()
		return nil, errors.New("subscriber closed")
	}
	s.closeMutex.Unlock()

	// Create output channel for messages
	output := make(chan *message.Message)

	log.Debug().Str("topic", topic).Msg("Starting Pub/Sub subscription goroutine")
	s.subscribersWg.Add(1)
	go func() {
		defer s.subscribersWg.Done()
		defer close(output)
		s.consume(ctx, topic, output)
		log.Info().Str("topic", topic).Msg("Pub/Sub consume loop finished")
	}()

	return output, nil
}

// consume handles the actual subscription and message forwarding
func (s *PubSubSubscriber) consume(ctx context.Context, topic string, output chan *message.Message) {
	logger := log.With().Str("topic", topic).Logger()
	logger.Info().Msg("Starting Pub/Sub consume loop")

	// Create Pub/Sub subscription using PSubscribe for patterns
	pubsub := s.config.Client.PSubscribe(ctx, topic)
	defer func() {
		logger.Debug().Msg("Closing Redis Pub/Sub subscription")
		_ = pubsub.Close() // Best effort close
	}()

	// Wait for confirmation
	_, err := pubsub.Receive(ctx) // This blocks until subscribed or error
	if err != nil {
		logger.Error().Err(err).Msg("Failed to PSubscribe to Redis Pub/Sub")
		return
	}
	logger.Info().Msg("Successfully PSubscribed to Redis Pub/Sub")

	// Get message channel
	msgChan := pubsub.Channel()

	for {
		select {
		case <-s.closing:
			logger.Debug().Msg("Subscriber closing, exiting consume loop")
			return
		case <-ctx.Done():
			logger.Debug().Msg("Context cancelled, exiting consume loop")
			return
		case msg, ok := <-msgChan:
			if !ok {
				logger.Info().Msg("Redis Pub/Sub channel closed by Redis, exiting consume loop")
				return
			}

			msgLogger := logger.With().
				Str("redis_channel", msg.Channel).
				Str("redis_pattern", msg.Pattern).
				Logger()
			msgLogger.Trace().Msg("&&& Received raw message from Redis Pub/Sub")

			// Convert Redis message to Watermill message
			msgLogger.Trace().Msg("&&& Attempting to unmarshal payload")
			watermillMsg, err := s.config.Unmarshaller.Unmarshal(map[string]interface{}{ // Adapt to Unmarshaller format
				"json_payload": msg.Payload, // Assuming Unmarshaller expects this key
			})
			if err != nil {
				msgLogger.Error().Err(err).Msg("&&& Failed to unmarshal message, skipping")
				continue // Skip malformed messages
			}
			watermillMsg.Metadata.Set("redis_channel", msg.Channel)
			if msg.Pattern != "" {
				watermillMsg.Metadata.Set("redis_pattern", msg.Pattern)
			}
			msgLogger = msgLogger.With().Str("message_uuid", watermillMsg.UUID).Logger()
			msgLogger.Trace().Msg("&&& Successfully unmarshalled message")

			// Send message to output channel (Watermill router)
			msgLogger.Trace().Msg("&&& Sending message to output channel")
			select {
			case output <- watermillMsg:
				msgLogger.Trace().Msg("&&& Message sent to router output channel")

				// In Pub/Sub, messages are fire-and-forget from Redis perspective.
				// We wait here briefly to see if the handler ACKs/NACKs for logging.
				select {
				case <-watermillMsg.Acked():
					msgLogger.Trace().Msg("&&& Message Acked by handler")
				case <-watermillMsg.Nacked():
					msgLogger.Error().Err(errors.New("message nacked by handler")).Msg("&&& Message Nacked by handler")
				case <-s.closing:
					msgLogger.Debug().Msg("Subscriber closing while waiting for ACK/NACK")
					return
				case <-ctx.Done():
					msgLogger.Debug().Msg("Context cancelled while waiting for ACK/NACK")
					return
				case <-time.After(time.Second * 5): // Timeout for logging ACK/NACK
					msgLogger.Trace().Msg("Timeout waiting for ACK/NACK log")
				}

			case <-s.closing:
				msgLogger.Debug().Msg("Subscriber closing before sending to output channel")
				return
			case <-ctx.Done():
				msgLogger.Debug().Msg("Context cancelled before sending to output channel")
				return
			}
		}
	}
}

// Close closes the subscriber
func (s *PubSubSubscriber) Close() error {
	s.closeMutex.Lock()
	defer s.closeMutex.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true
	log.Debug().Msg("Closing Pub/Sub subscriber")

	close(s.closing)
	log.Debug().Msg("Waiting for Pub/Sub consume goroutines to finish")
	s.subscribersWg.Wait()
	log.Info().Msg("Pub/Sub subscriber closed cleanly")

	return nil
}
