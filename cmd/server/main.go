package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/go-go-golems/go-go-agent/internal/db"
	"github.com/go-go-golems/go-go-agent/internal/redis"
	"github.com/go-go-golems/go-go-agent/internal/server"
	"github.com/go-go-golems/go-go-agent/internal/state"
	"github.com/go-go-golems/go-go-agent/pkg/model"
)

// ServerCommand implements the main server command using Glazed
type ServerCommand struct {
	*cmds.CommandDescription
}

// Ensure ServerCommand implements the Command interface
var _ cmds.Command = (*ServerCommand)(nil)

// ServerSettings holds the server-specific settings
type ServerSettings struct {
	DBPath          string `glazed.parameter:"db-path"`
	HTTPListenAddr  string `glazed.parameter:"http-listen-addr"`
	StaticFilesDir  string `glazed.parameter:"static-files-dir"`
	ReloadSession   bool   `glazed.parameter:"reload-session"`
	MaxEventHistory int    `glazed.parameter:"max-event-history"`
	LogLevel        string `glazed.parameter:"log-level"`
}

func (c *ServerCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Get settings from all layers
	serverSettings := &ServerSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, serverSettings); err != nil {
		return err
	}

	redisSettings, err := redis.GetRedisSettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	streamSettings, err := redis.GetStreamSettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	// Setup logger with level from flag
	level, err := zerolog.ParseLevel(serverSettings.LogLevel)
	if err != nil {
		return errors.Wrap(err, "invalid log level")
	}
	zerolog.SetGlobalLevel(level)

	// Setup pretty console logging
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
	logger := zerolog.New(consoleWriter).With().Timestamp().Caller().Logger()
	log.Logger = logger

	// Validate transport type
	selectedTransportType := redis.TransportType(streamSettings.TransportType)
	if selectedTransportType != redis.TransportStream && selectedTransportType != redis.TransportPubSub {
		return fmt.Errorf("invalid transport type: %s, must be %s or %s",
			streamSettings.TransportType, redis.TransportStream, redis.TransportPubSub)
	}

	logger.Info().
		Str("redis_url", redisSettings.URL).
		Str("transport_type", streamSettings.TransportType).
		Str("db_path", serverSettings.DBPath).
		Str("http_listen_addr", serverSettings.HTTPListenAddr).
		Str("log_level", serverSettings.LogLevel).
		Bool("reload_session", serverSettings.ReloadSession).
		Msg("Starting WriteHERE server")

	// Setup graceful shutdown context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup signal handling
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		logger.Info().Str("signal", sig.String()).Msg("Received shutdown signal, cancelling context")
		cancel()
	}()

	// Initialize database manager
	dbManager, err := db.NewDatabaseManager(serverSettings.DBPath)
	if err != nil {
		return errors.Wrap(err, "failed to initialize database manager")
	}
	defer func() {
		if err := dbManager.Close(); err != nil {
			logger.Error().Err(err).Msg("Error closing database manager")
		}
	}()

	// Initialize state managers
	eventManager := state.NewEventManager(logger, serverSettings.MaxEventHistory)
	graphManager := state.NewGraphManager(logger)

	// If reload_session is enabled, load state from the database
	if serverSettings.ReloadSession {
		logger.Info().Msg("Reloading latest session from database...")

		// Load graph state
		if err := state.LoadGraphFromDB(ctx, logger, dbManager, graphManager); err != nil {
			logger.Warn().Err(err).Msg("Failed to load graph state from database")
		}

		// Load events
		if err := state.LoadEventsFromDB(ctx, logger, dbManager, eventManager); err != nil {
			logger.Warn().Err(err).Msg("Failed to load events from database")
		}
	}

	// Setup HTTP server configuration
	httpConfig := server.DefaultHTTPServerConfig()
	httpConfig.ListenAddr = serverSettings.HTTPListenAddr
	httpConfig.StaticFilesDir = serverSettings.StaticFilesDir
	httpConfig.ReloadSession = serverSettings.ReloadSession

	// Initialize HTTP server
	httpServer := server.NewHTTPServer(httpConfig, logger, eventManager, graphManager)

	// Create a custom message handler that updates state managers and broadcasts to WebSocket clients
	messageHandler := func(msg *message.Message) error {
		// Pass to the DB manager for storage
		if err := dbManager.HandleMessage(msg); err != nil {
			return err // NACK if DB storage fails
		}

		// Parse the event from the message payload
		var event model.Event
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			// Log error but ACK, as DB storage succeeded
			logger.Error().Err(err).Str("message_uuid", msg.UUID).Msg("Failed to unmarshal event for state update, but DB storage succeeded. ACKing message.")
			return nil
		}

		// Update in-memory state with the parsed event
		eventManager.AddEvent(event)
		graphManager.ProcessEvent(event)

		// Broadcast the raw event to WebSocket clients
		httpServer.BroadcastEvent(msg.Payload)

		return nil // ACK
	}

	// Setup Redis router configuration
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
		logger.Info().Str("stream", streamSettings.StreamName).Str("group", streamSettings.ConsumerGroup).Msg("Using Redis Stream transport")
	case redis.TransportPubSub:
		routerConfig.TopicPattern = streamSettings.TopicPattern
		logger.Info().Str("topic_pattern", streamSettings.TopicPattern).Msg("Using Redis Pub/Sub transport")
	}

	// Create the Redis router
	router, err := redis.NewRouter(
		ctx,
		routerConfig,
		logger,
		messageHandler, // Pass the combined handler
	)
	if err != nil {
		return errors.Wrap(err, "failed to create Redis router")
	}

	// Use errgroup to manage the main goroutines
	g, gCtx := errgroup.WithContext(ctx)

	// Start the HTTP server (which includes the WebSocket hub)
	httpServer.Start(g, gCtx)

	// Start the Redis router
	g.Go(func() error {
		logger.Info().Msg("Starting Redis router")
		err := router.Run(gCtx)
		if err != nil {
			logger.Error().Err(err).Msg("Redis router run failed")
		}
		return err // Return the error to the errgroup
	})

	// Wait for the first error or context cancellation
	if err := g.Wait(); err != nil {
		logger.Error().Err(err).Msg("Server encountered an error")
	}

	// Context was cancelled (either by signal or error in a goroutine)
	logger.Info().Msg("Server shutting down...")

	// Close the router explicitly (gives Watermill time to finish processing)
	// HTTP server shutdown is handled by its own goroutine within httpServer.Start
	if err := router.Close(); err != nil {
		logger.Error().Err(err).Msg("Failed to close router gracefully")
	} else {
		logger.Info().Msg("Router closed gracefully")
	}

	logger.Info().Msg("Server shut down complete.")
	return err // Return the error that caused the shutdown, or nil if shutdown was graceful
}

// NewServerCommand creates a new server command with all parameter layers
func NewServerCommand() (*ServerCommand, error) {
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
		"server",
		cmds.WithShort("WriteHERE server with database manager"),
		cmds.WithLong(`A server that handles WriteHERE events and stores them in a SQLite database.
The server supports both Redis Streams and Pub/Sub for message transport.`),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"db-path",
				parameters.ParameterTypeString,
				parameters.WithHelp("Path to SQLite database file"),
				parameters.WithDefault("./writehere.db"),
			),
			parameters.NewParameterDefinition(
				"http-listen-addr",
				parameters.ParameterTypeString,
				parameters.WithHelp("HTTP server listen address"),
				parameters.WithDefault(":9999"),
			),
			parameters.NewParameterDefinition(
				"static-files-dir",
				parameters.ParameterTypeString,
				parameters.WithHelp("Path to static UI files"),
				parameters.WithDefault("./ui-react/dist"),
			),
			parameters.NewParameterDefinition(
				"reload-session",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Reload state from the latest session in the database"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"max-event-history",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of events to keep in memory"),
				parameters.WithDefault(1000),
			),
			parameters.NewParameterDefinition(
				"log-level",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Log level (trace, debug, info, warn, error, fatal, panic)"),
				parameters.WithDefault("info"),
				parameters.WithChoices("trace", "debug", "info", "warn", "error", "fatal", "panic"),
			),
		),
		cmds.WithLayersList(redisLayer, streamLayer),
	)

	return &ServerCommand{
		CommandDescription: cmdDesc,
	}, nil
}

func main() {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "go-go-agent",
		Short: "WriteHERE agent and server",
	}

	// Initialize help system
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Create server command
	serverCmd, err := NewServerCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating server command: %v\n", err)
		os.Exit(1)
	}

	// Convert to Cobra command
	cobraCmd, err := cli.BuildCobraCommandFromCommand(serverCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building cobra command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
