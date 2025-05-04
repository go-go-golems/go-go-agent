package redis

import (
	"errors"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// StreamSettings holds the stream configuration parameters
type StreamSettings struct {
	TransportType     string        `glazed.parameter:"transport-type"`
	StreamName        string        `glazed.parameter:"stream-name"`
	ConsumerGroup     string        `glazed.parameter:"consumer-group"`
	ConsumerName      string        `glazed.parameter:"consumer-name"`
	ClaimMinIdleTime  time.Duration `glazed.parameter:"claim-min-idle-time-s"`
	BlockTime         time.Duration `glazed.parameter:"block-time-s"`
	MaxIdleTime       time.Duration `glazed.parameter:"max-idle-time-s"`
	NackResendSleep   time.Duration `glazed.parameter:"nack-resend-sleep-s"`
	CommitOffsetAfter time.Duration `glazed.parameter:"commit-offset-after-s"`
	AckWait           time.Duration `glazed.parameter:"ack-wait-s"`
	TopicPattern      string        `glazed.parameter:"topic-pattern"`
}

// NewStreamLayer creates a new parameter layer for stream configuration
func NewStreamLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"stream",
		"Stream configuration options",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"transport-type",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Redis transport type (stream or pubsub)"),
				parameters.WithDefault("stream"),
				parameters.WithChoices("stream", "pubsub"),
			),
			parameters.NewParameterDefinition(
				"stream-name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Redis stream name (used with transport-type=stream)"),
				parameters.WithDefault("agent_events"),
			),
			parameters.NewParameterDefinition(
				"consumer-group",
				parameters.ParameterTypeString,
				parameters.WithHelp("Redis consumer group name (used with transport-type=stream)"),
				parameters.WithDefault("go_server_group"),
			),
			parameters.NewParameterDefinition(
				"consumer-name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Redis consumer name (used with transport-type=stream)"),
				parameters.WithDefault("go_server_consumer"),
			),
			parameters.NewParameterDefinition(
				"claim-min-idle-time-s",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Minimum idle time before claiming messages in seconds (used with transport-type=stream)"),
				parameters.WithDefault(60), // 1 minute
			),
			parameters.NewParameterDefinition(
				"block-time-s",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Block time for Redis XREADGROUP in seconds (used with transport-type=stream)"),
				parameters.WithDefault(1),
			),
			parameters.NewParameterDefinition(
				"max-idle-time-s",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum idle time for consumer in seconds (used with transport-type=stream)"),
				parameters.WithDefault(300), // 5 minutes
			),
			parameters.NewParameterDefinition(
				"nack-resend-sleep-s",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Sleep time before retrying NACK'd message in seconds (used with transport-type=stream)"),
				parameters.WithDefault(2),
			),
			parameters.NewParameterDefinition(
				"commit-offset-after-s",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Interval between committing offsets in seconds (used with transport-type=stream)"),
				parameters.WithDefault(10),
			),
			parameters.NewParameterDefinition(
				"ack-wait-s",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Time to wait for ACK before timing out in seconds"),
				parameters.WithDefault(30),
			),
			parameters.NewParameterDefinition(
				"topic-pattern",
				parameters.ParameterTypeString,
				parameters.WithHelp("Redis Pub/Sub topic pattern (used with transport-type=pubsub)"),
				parameters.WithDefault("agent_events:*"),
			),
		),
	)
}

// GetStreamSettingsFromParsedLayers extracts stream settings from parsed layers
func GetStreamSettingsFromParsedLayers(parsedLayers *layers.ParsedLayers) (*StreamSettings, error) {
	s := &StreamSettings{}
	if err := parsedLayers.InitializeStruct("stream", s); err != nil {
		return nil, err
	}

	// Convert seconds to durations
	durationParams := []struct {
		name   string
		target *time.Duration
	}{
		{"claim-min-idle-time-s", &s.ClaimMinIdleTime},
		{"block-time-s", &s.BlockTime},
		{"max-idle-time-s", &s.MaxIdleTime},
		{"nack-resend-sleep-s", &s.NackResendSleep},
		{"commit-offset-after-s", &s.CommitOffsetAfter},
		{"ack-wait-s", &s.AckWait},
	}

	for _, param := range durationParams {
		secs, ok := parsedLayers.GetParameter("stream", param.name)
		if !ok {
			return nil, errors.New("stream parameter not found")
		}
		*param.target = time.Duration(secs.Value.(int)) * time.Second
	}

	return s, nil
}
