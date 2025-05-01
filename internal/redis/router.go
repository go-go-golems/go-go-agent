package redis

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// RouterConfig holds configuration for the message router
type RouterConfig struct {
	// Redis connection options
	RedisURL         string
	RedisPassword    string
	RedisDB          int
	RedisMaxRetries  int
	RedisDialTimeout time.Duration

	// Transport type selection
	TransportType TransportType

	// Stream subscription options (used when TransportType is TransportStream)
	StreamName        string
	ConsumerGroup     string
	ConsumerName      string
	ClaimMinIdleTime  time.Duration
	BlockTime         time.Duration
	MaxIdleTime       time.Duration
	NackResendSleep   time.Duration
	CommitOffsetAfter time.Duration

	// Pub/Sub subscription options (used when TransportType is TransportPubSub)
	TopicPattern string // Pattern to subscribe to (e.g., "agent_events:*")

	// Processing options
	AckWait time.Duration

	// Internal fields
	redisClient redis.UniversalClient
}

// DefaultRouterConfig returns a RouterConfig with reasonable defaults
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		RedisURL:          "localhost:6379",
		RedisPassword:     "",
		RedisDB:           0,
		RedisMaxRetries:   3,
		RedisDialTimeout:  time.Second * 5,
		TransportType:     TransportStream, // Default to Stream for backward compatibility
		StreamName:        "agent_events",
		ConsumerGroup:     "go_server_group",
		ConsumerName:      "go_server_consumer",
		ClaimMinIdleTime:  time.Minute * 1,
		BlockTime:         time.Second * 1,
		MaxIdleTime:       time.Minute * 5,
		NackResendSleep:   time.Second * 2,
		CommitOffsetAfter: time.Second * 10,
		TopicPattern:      "agent_events:*",
		AckWait:           time.Second * 30,
	}
}

// WatermillLogger is an adapter to use zerolog with Watermill
type WatermillLogger struct {
	logger zerolog.Logger
}

// NewWatermillLogger creates a new WatermillLogger
func NewWatermillLogger(logger zerolog.Logger) *WatermillLogger {
	return &WatermillLogger{
		logger: logger.With().Str("component", "watermill").Logger(),
	}
}

// Error logs an error message
func (l *WatermillLogger) Error(msg string, err error, fields watermill.LogFields) {
	event := l.logger.Error().Err(err)
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Info logs an info message
func (l *WatermillLogger) Info(msg string, fields watermill.LogFields) {
	event := l.logger.Info()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Debug logs a debug message
func (l *WatermillLogger) Debug(msg string, fields watermill.LogFields) {
	event := l.logger.Debug()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// Trace logs a trace message
func (l *WatermillLogger) Trace(msg string, fields watermill.LogFields) {
	event := l.logger.Trace()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	event.Msg(msg)
}

// With returns a new WatermillLogger with the given fields appended
func (l *WatermillLogger) With(fields watermill.LogFields) watermill.LoggerAdapter {
	loggerContext := l.logger.With()
	for k, v := range fields {
		loggerContext = loggerContext.Interface(k, v)
	}
	return &WatermillLogger{logger: loggerContext.Logger()}
}

// MessageHandler is a function that processes a message and returns an error if processing fails
type MessageHandler func(msg *message.Message) error

// NewRouter creates a new message router connected to Redis
func NewRouter(
	ctx context.Context,
	config RouterConfig,
	logger zerolog.Logger, // Expect main application logger
	handlers ...MessageHandler,
) (*message.Router, error) {
	// Watermill's router still needs its adapter
	watermillLogger := NewWatermillLogger(logger)

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:        config.RedisURL,
		Password:    config.RedisPassword,
		DB:          config.RedisDB,
		MaxRetries:  config.RedisMaxRetries,
		DialTimeout: config.RedisDialTimeout,
	})

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, errors.Wrap(err, "failed to connect to Redis")
	}

	// Store Redis client in config for use by transport
	config.redisClient = redisClient

	// Create appropriate transport based on config
	var subscriber message.Subscriber
	var transport Transport
	switch config.TransportType {
	case TransportStream:
		streamTransport := NewStreamTransport()
		sub, err := streamTransport.CreateSubscriber(config, watermillLogger) // StreamTransport expects adapter
		if err != nil {
			redisClient.Close()
			return nil, errors.Wrap(err, "failed to create Stream subscriber")
		}
		subscriber = sub
		transport = streamTransport
	case TransportPubSub:
		pubsubTransport := NewPubSubTransport()
		sub, err := pubsubTransport.CreateSubscriber(config, watermillLogger) // PubSubTransport now expects watermill.LoggerAdapter
		if err != nil {
			redisClient.Close()
			return nil, errors.Wrap(err, "failed to create Pub/Sub subscriber")
		}
		subscriber = sub
		transport = pubsubTransport
	default:
		redisClient.Close() // Close client before returning error
		return nil, errors.Errorf("unsupported transport type: %s", config.TransportType)
	}

	// Create router
	router, err := message.NewRouter(message.RouterConfig{
		CloseTimeout: time.Second * 30,
	}, watermillLogger) // Router uses the adapter
	if err != nil {
		redisClient.Close()
		return nil, errors.Wrap(err, "failed to create message router")
	}

	// Add router plugins and middlewares
	router.AddPlugin(plugin.SignalsHandler)
	router.AddMiddleware(middleware.Recoverer)
	router.AddMiddleware(middleware.CorrelationID)
	router.AddMiddleware(middleware.Timeout(config.AckWait))

	// Get topic name from transport
	topic := transport.GetTopicName(config)

	// Register all handlers to the topic
	for i, handler := range handlers {
		handlerIndex := i // Capture for closure
		handlerFunc := handler

		router.AddHandler(
			// Handler name must be unique
			watermill.NewUUID(),
			// Topic to subscribe
			topic,
			// Subscriber
			subscriber,
			// No publishing needed, just consume
			"",
			nil,
			// Message handler function
			func(msg *message.Message) ([]*message.Message, error) {
				// This handler function adapts the MessageHandler signature to Watermill's expected format
				err := handlerFunc(msg)
				if err != nil {
					// Log using the main application logger passed into NewRouter
					logger.Error().
						Err(err).
						Int("handler_index", handlerIndex).
						Str("message_uuid", msg.UUID).
						Msg("Handler returned error, will NACK message")
					return nil, err // NACK
				}
				return nil, nil // ACK
			},
		)
	}

	return router, nil
}
