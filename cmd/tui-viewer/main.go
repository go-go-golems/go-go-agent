package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	tea "github.com/charmbracelet/bubbletea"
	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	ghelp "github.com/go-go-golems/glazed/pkg/help"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-agent/cmd/tui-viewer/tui"
	"github.com/go-go-golems/go-go-agent/internal/db"
	"github.com/go-go-golems/go-go-agent/internal/redis"
	"github.com/go-go-golems/go-go-agent/pkg/model"
)

type TUIViewerCmd struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = (*TUIViewerCmd)(nil)

type TUIViewerSettings struct {
	MaxEvents        int    `glazed.parameter:"max-events"`
	DBPath           string `glazed.parameter:"db-path"`
	UseDB            bool   `glazed.parameter:"use-db"`
	RunID            string `glazed.parameter:"run-id"`
	LatestRun        bool   `glazed.parameter:"latest-run"`
	Limit            int    `glazed.parameter:"limit"`
	StartTime        string `glazed.parameter:"start-time"`
	EndTime          string `glazed.parameter:"end-time"`
	StreamingEnabled bool   `glazed.parameter:"streaming-enabled"`
}

func (c *TUIViewerCmd) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	// Get settings
	tuiSettings := &TUIViewerSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, tuiSettings); err != nil {
		return err
	}

	// Use the global logger
	logger := log.Logger

	// Create a channel to receive events
	eventChan := make(chan model.Event, 100)

	// Create a place to store a router for cleanup
	var router *message.Router

	// If we're loading from database, set it up
	if tuiSettings.UseDB {
		dbManager, err := db.NewDatabaseManager(tuiSettings.DBPath)
		if err != nil {
			return errors.Wrap(err, "failed to initialize database manager")
		}
		defer func() {
			if err := dbManager.Close(); err != nil {
				logger.Error().Err(err).Msg("Error closing database manager")
			}
		}()

		// Load events based on the provided options
		var loadedEvents []model.Event

		if tuiSettings.RunID != "" {
			// Load events for a specific run
			eventData, err := dbManager.GetRunEvents(ctx, tuiSettings.RunID)
			if err != nil {
				return errors.Wrap(err, "failed to load events for run")
			}

			// Convert raw events to model.Event
			loadedEvents, err = convertRawEventsToEvents(eventData.Events)
			if err != nil {
				return errors.Wrap(err, "failed to convert events")
			}

		} else if tuiSettings.LatestRun {
			// Load events for the latest run
			eventData, err := dbManager.GetLatestRunEvents(ctx)
			if err != nil {
				return errors.Wrap(err, "failed to load latest run events")
			}

			// Convert raw events to model.Event
			loadedEvents, err = convertRawEventsToEvents(eventData.Events)
			if err != nil {
				return errors.Wrap(err, "failed to convert events")
			}

		} else if tuiSettings.StartTime != "" && tuiSettings.EndTime != "" {
			// Load events in a time range
			eventData, err := dbManager.GetEventsInTimeRange(ctx, tuiSettings.StartTime, tuiSettings.EndTime)
			if err != nil {
				return errors.Wrap(err, "failed to load events in time range")
			}

			// Convert raw events to model.Event
			loadedEvents, err = convertRawEventsToEvents(eventData.Events)
			if err != nil {
				return errors.Wrap(err, "failed to convert events")
			}

		} else {
			// Load the latest N events
			eventData, err := dbManager.GetLatestEvents(ctx, tuiSettings.Limit)
			if err != nil {
				return errors.Wrap(err, "failed to load latest events")
			}

			// Convert raw events to model.Event
			loadedEvents, err = convertRawEventsToEvents(eventData.Events)
			if err != nil {
				return errors.Wrap(err, "failed to convert events")
			}
		}

		logger.Info().Int("loaded_events", len(loadedEvents)).Msg("Loaded events from database")

		// Pre-load the events channel
		for _, event := range loadedEvents {
			eventChan <- event
		}

		// If streaming is enabled, also set up the router
		if tuiSettings.StreamingEnabled {
			router, err = setupRouter(ctx, parsedLayers, logger, eventChan)
			if err != nil {
				return err
			}

			// Start the router in a goroutine
			go func() {
				if err := router.Run(ctx); err != nil {
					logger.Error().Err(err).Msg("Router error")
				}
			}()
		}
	} else {
		// Just use streaming
		var err error
		router, err = setupRouter(ctx, parsedLayers, logger, eventChan)
		if err != nil {
			return err
		}

		// Start the router in a goroutine
		go func() {
			if err := router.Run(ctx); err != nil {
				logger.Error().Err(err).Msg("Router error")
			}
		}()
	}

	// Start bubbletea app
	m := tui.NewViewModel(tuiSettings.MaxEvents, eventChan)
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Clean up when the program exits
	defer func() {
		if router != nil {
			if err := router.Close(); err != nil {
				logger.Error().Err(err).Msg("Failed to close router")
			}
		}
		close(eventChan)
	}()

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

// Helper function to convert raw JSON events to model.Event structs
func convertRawEventsToEvents(rawEvents []json.RawMessage) ([]model.Event, error) {
	events := make([]model.Event, 0, len(rawEvents))

	for _, rawEvent := range rawEvents {
		var event model.Event
		if err := json.Unmarshal(rawEvent, &event); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal event")
		}
		events = append(events, event)
	}

	return events, nil
}

// Helper function to set up the redis router
func setupRouter(ctx context.Context, parsedLayers *layers.ParsedLayers, logger zerolog.Logger, eventChan chan model.Event) (*message.Router, error) {
	redisSettings, err := redis.GetRedisSettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return nil, err
	}

	streamSettings, err := redis.GetStreamSettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return nil, err
	}

	// Validate transport type
	selectedTransportType := redis.TransportType(streamSettings.TransportType)
	if selectedTransportType != redis.TransportStream && selectedTransportType != redis.TransportPubSub {
		return nil, fmt.Errorf("invalid transport type: %s, must be %s or %s",
			streamSettings.TransportType, redis.TransportStream, redis.TransportPubSub)
	}

	// Setup router configuration
	routerConfig := redis.DefaultRouterConfig()
	routerConfig.RedisURL = redisSettings.URL
	routerConfig.RedisPassword = redisSettings.Password
	routerConfig.RedisDB = redisSettings.DB
	routerConfig.RedisMaxRetries = redisSettings.MaxRetries
	routerConfig.RedisDialTimeout = redisSettings.DialTimeout
	routerConfig.AckWait = streamSettings.AckWait
	routerConfig.TransportType = selectedTransportType

	// Set transport-specific config
	switch selectedTransportType {
	case redis.TransportStream:
		routerConfig.StreamName = streamSettings.StreamName
		routerConfig.ConsumerGroup = streamSettings.ConsumerGroup
		routerConfig.ConsumerName = streamSettings.ConsumerName
		routerConfig.ClaimMinIdleTime = streamSettings.ClaimMinIdleTime
		routerConfig.BlockTime = streamSettings.BlockTime
		routerConfig.MaxIdleTime = streamSettings.MaxIdleTime
		routerConfig.NackResendSleep = streamSettings.NackResendSleep
		routerConfig.CommitOffsetAfter = streamSettings.CommitOffsetAfter
	case redis.TransportPubSub:
		routerConfig.TopicPattern = streamSettings.TopicPattern
	}

	// Create a message handler to forward events to our channel
	messageHandler := func(msg *message.Message) error {
		var event model.Event
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			logger.Error().Err(err).Msg("Failed to parse event")
			return nil // ACK anyway
		}

		// Send to our channel
		selectCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		select {
		case eventChan <- event:
			// Successfully sent
		case <-selectCtx.Done():
			// Channel is full or context done, log but continue
			logger.Warn().Msg("Event channel full or blocked, dropping event")
		}

		return nil // ACK
	}

	// Create the router
	return redis.NewRouter(
		ctx,
		routerConfig,
		logger,
		messageHandler,
	)
}

func NewTUIViewerCmd() (*TUIViewerCmd, error) {
	// Create Redis layer
	redisLayer, err := redis.NewRedisLayer()
	if err != nil {
		return nil, err
	}

	// Create stream layer
	streamLayer, err := redis.NewStreamLayer()
	if err != nil {
		return nil, err
	}

	// Create command description with all parameters
	cmdDesc := cmds.NewCommandDescription(
		"tui-viewer",
		cmds.WithShort("TUI viewer for WriteHERE events"),
		cmds.WithLong(`A terminal UI application that displays WriteHERE events in real-time or loads them from a SQLite database.`),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"max-events",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of events to keep in memory (0 for unlimited)"),
				parameters.WithDefault(1000),
			),
			parameters.NewParameterDefinition(
				"db-path",
				parameters.ParameterTypeString,
				parameters.WithHelp("Path to SQLite database file (for loading events)"),
				parameters.WithDefault("./writehere.db"),
			),
			parameters.NewParameterDefinition(
				"use-db",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Load events from database instead of streaming"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"run-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Load events for a specific run ID"),
			),
			parameters.NewParameterDefinition(
				"latest-run",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Load events from the latest run"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of events to load from database"),
				parameters.WithDefault(100),
			),
			parameters.NewParameterDefinition(
				"start-time",
				parameters.ParameterTypeString,
				parameters.WithHelp("Start time for loading events (ISO format)"),
			),
			parameters.NewParameterDefinition(
				"end-time",
				parameters.ParameterTypeString,
				parameters.WithHelp("End time for loading events (ISO format)"),
			),
			parameters.NewParameterDefinition(
				"streaming-enabled",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable streaming of new events (even when loading from DB)"),
				parameters.WithDefault(true),
			),
		),
		cmds.WithLayersList(redisLayer, streamLayer),
	)

	return &TUIViewerCmd{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "tui-viewer",
		Short: "TUI viewer for WriteHERE events",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Setup debug logging
			return logging.InitLoggerFromViper()
		},
	}

	// Initialize Viper
	err := clay.InitViper("go-go-agent", rootCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing viper: %v\n", err)
		os.Exit(1)
	}

	// Setup zerolog
	err = logging.InitLoggerFromViper()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}

	// Initialize help system
	helpSystem := ghelp.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Create tui-viewer command
	tuiCmd, err := NewTUIViewerCmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tui-viewer command: %v\n", err)
		os.Exit(1)
	}

	// Convert to Cobra command
	cobraCmd, err := cli.BuildCobraCommandFromCommand(tuiCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building cobra command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	log.Info().Msg("Starting go-go-agent tui-viewer")
	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
