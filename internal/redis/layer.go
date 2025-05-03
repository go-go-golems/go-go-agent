package redis

import (
	"errors"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// RedisSettings holds the Redis configuration parameters
type RedisSettings struct {
	URL         string        `glazed.parameter:"redis-url"`
	Password    string        `glazed.parameter:"redis-password"`
	DB          int           `glazed.parameter:"redis-db"`
	MaxRetries  int           `glazed.parameter:"redis-max-retries"`
	DialTimeout time.Duration `glazed.parameter:"redis-dial-timeout-s"`
}

// NewRedisLayer creates a new parameter layer for Redis configuration
func NewRedisLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"redis",
		"Redis connection configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"redis-url",
				parameters.ParameterTypeString,
				parameters.WithHelp("Redis server URL"),
				parameters.WithDefault("localhost:6379"),
			),
			parameters.NewParameterDefinition(
				"redis-password",
				parameters.ParameterTypeString,
				parameters.WithHelp("Redis server password"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"redis-db",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Redis database number"),
				parameters.WithDefault(0),
			),
			parameters.NewParameterDefinition(
				"redis-max-retries",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of Redis connection retries"),
				parameters.WithDefault(3),
			),
			parameters.NewParameterDefinition(
				"redis-dial-timeout-s",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Redis connection timeout in seconds"),
				parameters.WithDefault(5),
			),
		),
	)
}

// GetRedisSettingsFromParsedLayers extracts Redis settings from parsed layers
func GetRedisSettingsFromParsedLayers(parsedLayers *layers.ParsedLayers) (*RedisSettings, error) {
	s := &RedisSettings{}
	if err := parsedLayers.InitializeStruct("redis", s); err != nil {
		return nil, err
	}
	// Convert seconds to duration
	timeoutSecs, ok := parsedLayers.GetParameter("redis", "redis-dial-timeout-s")
	if !ok {
		return nil, errors.New("redis-dial-timeout-s not found")
	}
	s.DialTimeout = time.Duration(timeoutSecs.Value.(int)) * time.Second
	return s, nil
}
