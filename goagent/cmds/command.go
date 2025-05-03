package cmds

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/go-go-golems/geppetto/pkg/helpers"
	"github.com/go-go-golems/geppetto/pkg/steps/ai/settings"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	glazed_settings "github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/go-emrichen/pkg/emrichen"
	"github.com/go-go-golems/go-go-agent/goagent/agent"
	"github.com/go-go-golems/go-go-agent/goagent/llm"
	"github.com/go-go-golems/go-go-agent/goagent/types"
	"github.com/go-go-golems/go-go-agent/pkg/eventbus"
	events "github.com/go-go-golems/go-go-agent/proto"
	pinocchio_cmds "github.com/go-go-golems/pinocchio/pkg/cmds"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// AgentCommand is a command that encapsulates agent execution configuration.
// It serves as the base for WriterAgentCommand and GlazedAgentCommand.
type AgentCommand struct {
	*cmds.CommandDescription `yaml:",inline"`

	// Agent-specific fields
	AgentType    string
	SystemPrompt string
	Prompt       string // Template string for the initial prompt
	Tools        []string
	AgentOptions *types.RawNode
}

// Ensure AgentCommand implements the types.AgentCommandDescription interface
var _ agent.Command = &AgentCommand{}

// WriterAgentCommand is an AgentCommand designed to output plain text results.
type WriterAgentCommand struct {
	*AgentCommand
}

// Ensure WriterAgentCommand implements the WriterCommand interface
var _ cmds.WriterCommand = &WriterAgentCommand{}

// GlazedAgentCommand is an AgentCommand designed to output structured data.
type GlazedAgentCommand struct {
	*AgentCommand
}

// Ensure GlazedAgentCommand implements the GlazeCommand interface
var _ cmds.GlazeCommand = &GlazedAgentCommand{}

// AgentCommandOption is a functional option for configuring an AgentCommand
type AgentCommandOption func(*AgentCommand)

// WithAgentType sets the agent type
func WithAgentType(agentType string) AgentCommandOption {
	return func(a *AgentCommand) {
		a.AgentType = agentType
	}
}

// WithSystemPrompt sets the system prompt
func WithSystemPrompt(systemPrompt string) AgentCommandOption {
	return func(a *AgentCommand) {
		a.SystemPrompt = systemPrompt
	}
}

// WithPrompt sets the prompt template
func WithPrompt(prompt string) AgentCommandOption {
	return func(a *AgentCommand) {
		a.Prompt = prompt
	}
}

// WithTools sets the tools to use
func WithTools(tools []string) AgentCommandOption {
	return func(a *AgentCommand) {
		a.Tools = tools
	}
}

// WithAgentOptions sets additional agent options
func WithAgentOptions(options *types.RawNode) AgentCommandOption {
	return func(a *AgentCommand) {
		a.AgentOptions = options
	}
}

// NewAgentCommand creates a new base AgentCommand configuration.
// It's typically used internally by NewWriterAgentCommand and NewGlazedAgentCommand.
func NewAgentCommand(
	description *cmds.CommandDescription,
	options ...AgentCommandOption,
) (*AgentCommand, error) {
	// Add Geppetto layers for LLM configuration
	tempSettings, err := settings.NewStepSettings()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temporary step settings")
	}
	geppettoLayers, err := pinocchio_cmds.CreateGeppettoLayers(tempSettings)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Geppetto layers")
	}
	description.Layers.AppendLayers(geppettoLayers...)

	ret := &AgentCommand{
		CommandDescription: description,
	}

	for _, option := range options {
		option(ret)
	}

	// Ensure AgentType is set before proceeding
	if ret.AgentType == "" {
		return nil, errors.New("agent type must be specified using WithAgentType option")
	}

	// Get the factory for the specified agent type
	factory, err := agent.GetAgentFactory(ret.AgentType)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get agent factory for type '%s'", ret.AgentType)
	}

	// Get agent-specific layers from the factory
	agentLayers, err := factory.CreateLayers()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create layers for agent type '%s'", ret.AgentType)
	}

	// Append agent-specific layers to the command description
	description.Layers.AppendLayers(agentLayers...) // TODO(manuel): Remove this

	return ret, nil
}

// NewWriterAgentCommand creates a new AgentCommand configured for text output.
func NewWriterAgentCommand(
	description *cmds.CommandDescription,
	options ...AgentCommandOption,
) (*WriterAgentCommand, error) {
	agentCmd, err := NewAgentCommand(description, options...)
	if err != nil {
		return nil, err
	}
	return &WriterAgentCommand{AgentCommand: agentCmd}, nil
}

// NewGlazedAgentCommand creates a new AgentCommand configured for structured output.
func NewGlazedAgentCommand(
	description *cmds.CommandDescription,
	options ...AgentCommandOption,
) (*GlazedAgentCommand, error) {
	agentCmd, err := NewAgentCommand(description, options...)
	if err != nil {
		return nil, err
	}

	glazedLayer, err := glazed_settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create glazed parameter layers")
	}
	agentCmd.Layers.AppendLayers(glazedLayer)

	return &GlazedAgentCommand{AgentCommand: agentCmd}, nil
}

// RunMode determines how the agent command should execute and output results.
// This affects whether the event bus and router are used.
type RunMode int

const (
	RunModeWriter RunMode = iota // Output plain text, uses event bus + stdout handler
	RunModeGlazed                // Output structured data, does not use event bus
)

// prepareLlmAndEventBus prepares the LLM model, event bus, and related settings.
func (a *AgentCommand) prepareLlmAndEventBus(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	runMode RunMode,
	runID string, // Pass runID for event association
) (llm.LLM, *settings.StepSettings, *eventbus.EventBus, *message.Router, string, error) {
	// Create StepSettings from parsed layers for LLM creation
	stepSettings, err := settings.NewStepSettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return nil, nil, nil, nil, "", errors.Wrap(err, "failed to create step settings from parsed layers")
	}

	llmOptions := []llm.GeppettoLLMOption{
		llm.WithRunID(runID),
	}
	var eb *eventbus.EventBus
	var router *message.Router
	var topicID string

	// Setup EventBus and Router only for Writer Mode
	if runMode == RunModeWriter {
		// Create a simple GoChannel pub/sub for local event handling
		logger := helpers.NewWatermill(log.Logger)
		pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

		topicID = fmt.Sprintf("%s-agent-events-%s", a.Name, runID)

		// Create the EventBus
		eb, err = eventbus.NewEventBus(
			eventbus.WithPublisher(pubSub),
			eventbus.WithTopic(topicID),
			// Use DefaultJSONEncoder for human-readable stdout printing
			eventbus.WithEncoder(eventbus.DefaultJSONEncoder),
		)
		if err != nil {
			return nil, nil, nil, nil, "", errors.Wrap(err, "failed to create event bus")
		}

		// Create and configure the router
		router, err = message.NewRouter(message.RouterConfig{}, logger)
		if err != nil {
			_ = eb.Close() // Attempt to close event bus on error
			return nil, nil, nil, nil, "", errors.Wrap(err, "failed to create message router")
		}

		// Add the stdout printing handler
		handlerName := "stdout-event-printer-" + runID
		log.Info().Str("handler", handlerName).Str("topic", topicID).Msg("Registering stdout event handler")
		router.AddHandler(
			handlerName,
			topicID,
			pubSub, // Subscribe to the same pub/sub
			topicID,
			pubSub,             // Publish ACKs/NACKs to the same pub/sub (for potential future use)
			StdoutEventHandler, // Use the new handler
		)

		// Configure LLM to use the event bus
		llmOptions = append(llmOptions, llm.WithEventBus(eb))
	}

	// Prepare LLM model using Geppetto LLM
	llmModel, err := llm.NewGeppettoLLM(
		stepSettings,
		llmOptions...,
	)
	if err != nil {
		if eb != nil {
			_ = eb.Close()
		}
		if router != nil {
			_ = router.Close() // Attempt to close router on error
		}
		return nil, nil, nil, nil, "", errors.Wrap(err, "failed to create Geppetto LLM")
	}

	return llmModel, stepSettings, eb, router, topicID, nil
}

// StdoutEventHandler is a Watermill handler that prints received events to stdout.
func StdoutEventHandler(msg *message.Message) ([]*message.Message, error) {
	// Decode the message payload into an Event
	event := &events.Event{}
	opts := protojson.UnmarshalOptions{
		DiscardUnknown: true, // Be lenient with unknown fields
	}
	err := opts.Unmarshal(msg.Payload, event)
	if err != nil {
		log.Error().Err(err).Str("msg_uuid", msg.UUID).Msg("Failed to unmarshal event from message payload")
		// Nack the message? Or Ack and log? For stdout printing, Ack is probably fine.
		return nil, err // Don't requeue
	}

	// Pretty print the event to stdout
	outputStr, err := prettyPrintEvent(event)
	if err != nil {
		log.Error().Err(err).Str("event_id", event.EventId).Msg("Failed to pretty print event")
		if _, err := fmt.Fprintf(os.Stdout, "---\\nEvent ID: %s\\nType: %s\\nTimestamp: %s\\nRaw Payload: %s\\nError Pretty Printing: %v\\n---\\n",
			event.EventId, event.EventType, event.GetTimestamp().AsTime().Format(time.RFC3339Nano), string(msg.Payload), err); err != nil {
			log.Error().Err(err).Msg("Failed to write error message to stdout")
		}
	} else {
		if _, err := fmt.Fprintln(os.Stdout, outputStr); err != nil {
			log.Error().Err(err).Msg("Failed to write output to stdout")
		}
	}

	// Mark the message as processed successfully
	msg.Ack()
	return nil, nil
}

// prettyPrintEvent formats an event for readable stdout logging.
func prettyPrintEvent(event *events.Event) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("---\nEvent ID: %s\nRun ID:   %s\nType:     %s\nTime:     %s\nPayload:\n",
		event.EventId,
		*event.RunId, // Assume RunId is always set when handled here
		event.EventType.String(),
		event.GetTimestamp().AsTime().Format(time.RFC3339Nano),
	))

	// Use protojson marshaller for readable payload output
	opts := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		EmitUnpopulated: false, // Don't show empty optional fields
		UseProtoNames:   true,
	}

	var payloadProto proto.Message
	switch p := event.Payload.(type) {
	case *events.Event_StepStarted:
		payloadProto = p.StepStarted
	case *events.Event_StepFinished:
		payloadProto = p.StepFinished
	case *events.Event_NodeStatusChanged:
		payloadProto = p.NodeStatusChanged
	case *events.Event_LlmCallStarted:
		payloadProto = p.LlmCallStarted
	case *events.Event_LlmCallCompleted:
		payloadProto = p.LlmCallCompleted
	case *events.Event_ToolInvoked:
		payloadProto = p.ToolInvoked
	case *events.Event_ToolReturned:
		payloadProto = p.ToolReturned
	case *events.Event_NodeCreated:
		payloadProto = p.NodeCreated
	case *events.Event_PlanReceived:
		payloadProto = p.PlanReceived
	case *events.Event_NodeAdded:
		payloadProto = p.NodeAdded
	case *events.Event_EdgeAdded:
		payloadProto = p.EdgeAdded
	case *events.Event_InnerGraphBuilt:
		payloadProto = p.InnerGraphBuilt
	case *events.Event_NodeResultAvailable:
		payloadProto = p.NodeResultAvailable
	case *events.Event_RunStarted:
		payloadProto = p.RunStarted
	case *events.Event_RunFinished:
		payloadProto = p.RunFinished
	case *events.Event_RunError:
		payloadProto = p.RunError
	case *events.Event_UnknownPayload:
		payloadProto = p.UnknownPayload
	default:
		return "", fmt.Errorf("unknown payload type: %T", event.Payload)
	}

	if payloadProto != nil {
		payloadBytes, err := opts.Marshal(payloadProto)
		if err != nil {
			sb.WriteString(fmt.Sprintf("  Error marshalling payload: %v\n", err))
		} else {
			// Indent the payload JSON
			scanner := bufio.NewScanner(bytes.NewReader(payloadBytes))
			for scanner.Scan() {
				sb.WriteString("  ") // Add indentation
				sb.WriteString(scanner.Text())
				sb.WriteString("\n")
			}
			if err := scanner.Err(); err != nil {
				sb.WriteString(fmt.Sprintf("  Error reading marshalled payload: %v\n", err))
			}
		}
	} else {
		sb.WriteString("  <nil payload>\n")
	}

	sb.WriteString("---")
	return sb.String(), nil
}

// RunIntoGlazeProcessor implements the GlazeCommand interface for GlazedAgentCommand
func (gac *GlazedAgentCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	runID := uuid.New().String()
	// 1. Prepare LLM (no event bus/router for Glazed mode)
	llmModel, _, _, _, _, err := gac.AgentCommand.prepareLlmAndEventBus(ctx, parsedLayers, RunModeGlazed, runID)
	if err != nil {
		return errors.Wrap(err, "failed to prepare LLM")
	}

	// 2. Get Agent Factory
	factory, err := agent.GetAgentFactory(gac.AgentCommand.AgentType)
	if err != nil {
		return errors.Wrapf(err, "failed to get agent factory for type '%s'", gac.AgentCommand.AgentType)
	}

	// 3. Create Agent Instance using Factory - pass the AgentCommand directly
	agentInstance, err := factory.NewAgent(ctx, gac.AgentCommand, parsedLayers, llmModel)
	if err != nil {
		return errors.Wrap(err, "failed to create agent instance")
	}

	// 4. Render the initial prompt using parameters
	initialPrompt, err := gac.AgentCommand.renderInitialPrompt(parsedLayers)
	if err != nil {
		return errors.Wrap(err, "failed to render initial prompt")
	}

	// 5. Type assert the agent to GlazedAgent
	glazedAgent, ok := agentInstance.(agent.GlazedAgent)
	if !ok {
		return errors.Errorf("agent type '%s' (%T) does not support Glazed output (does not implement GlazedAgent interface)", gac.AgentCommand.AgentType, agentInstance)
	}

	// 6. Run the agent's specific Glazed processor method
	log.Info().Str("agentType", gac.AgentCommand.AgentType).Str("runID", runID).Msg("Running GlazedAgent logic")
	err = glazedAgent.RunIntoGlazeProcessor(ctx, initialPrompt, gp)
	if err != nil {
		log.Error().Err(err).Str("agentType", gac.AgentCommand.AgentType).Str("runID", runID).Msg("GlazedAgent RunIntoGlazeProcessor failed")
	}

	return err // Return the error from the agent run
}

// RunIntoWriter implements the WriterCommand interface for WriterAgentCommand
func (wac *WriterAgentCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer, // Target writer for the final agent output
) error {
	// 1. Setup context, errgroup, and run ID
	runID := uuid.New().String()
	ctx, cancel := context.WithCancel(ctx)
	eg, ctx := errgroup.WithContext(ctx)
	defer cancel()

	// 2. Prepare LLM, EventBus, and Router for Writer Mode
	llmModel, _, eb, router, _, err := wac.AgentCommand.prepareLlmAndEventBus(ctx, parsedLayers, RunModeWriter, runID)
	if err != nil {
		return errors.Wrap(err, "failed to prepare LLM and event bus")
	}
	// Ensure event bus and router are closed eventually
	defer func() {
		if eb != nil {
			_ = eb.Close()
		}
		if router != nil {
			_ = router.Close()
		}
	}()

	// 3. Get Agent Factory
	factory, err := agent.GetAgentFactory(wac.AgentCommand.AgentType)
	if err != nil {
		return errors.Wrapf(err, "failed to get agent factory for type '%s'", wac.AgentCommand.AgentType)
	}

	// 4. Create Agent Instance using Factory
	agentInstance, err := factory.NewAgent(ctx, wac.AgentCommand, parsedLayers, llmModel)
	if err != nil {
		return errors.Wrap(err, "failed to create agent instance")
	}

	// 5. Render the initial prompt
	initialPrompt, err := wac.AgentCommand.renderInitialPrompt(parsedLayers)
	if err != nil {
		return errors.Wrap(err, "failed to render initial prompt")
	}

	// Start the router in a background goroutine
	eg.Go(func() error {
		log.Info().Str("runID", runID).Msg("Starting event router for WriterAgent")
		defer func() {
			log.Info().Str("runID", runID).Msg("Event router closing")
		}()
		runErr := router.Run(ctx) // Use the cancellable context
		log.Info().Err(runErr).Str("runID", runID).Msg("Event router stopped")
		if runErr != nil && !errors.Is(runErr, context.Canceled) && !errors.Is(runErr, context.DeadlineExceeded) {
			return runErr // Return actual router errors
		}
		return nil
	})

	// Run the agent logic in a separate goroutine
	eg.Go(func() error {
		defer cancel() // Ensure cancellation propagates if agent finishes/errors first
		// Wait for the router to be running
		select {
		case <-router.Running():
			log.Info().Str("agentType", wac.AgentCommand.AgentType).Str("runID", runID).Msg("Event router is running, proceeding with agent run")
		case <-ctx.Done():
			log.Warn().Str("agentType", wac.AgentCommand.AgentType).Str("runID", runID).Msg("Context cancelled before router started")
			return ctx.Err()
		case <-time.After(5 * time.Second): // Timeout for router startup
			log.Error().Str("runID", runID).Msg("Timeout waiting for event router to start")
			return errors.New("timeout waiting for event router to start")
		}

		// Run the agent's standard Run method
		log.Info().Str("agentType", wac.AgentCommand.AgentType).Str("runID", runID).Msg("Running WriterAgent logic")
		resultStr, agentErr := agentInstance.Run(ctx, initialPrompt)
		if agentErr != nil {
			log.Error().Err(agentErr).Str("agentType", wac.AgentCommand.AgentType).Str("runID", runID).Msg("WriterAgent Run failed")
			// Emit a run error event if possible
			if eb != nil {
				errPayload := &events.RunErrorPayload{
					ErrorType:    fmt.Sprintf("%T", errors.Cause(agentErr)),
					ErrorMessage: agentErr.Error(),
					// Stack trace might be too verbose for event?
				}
				_ = eb.EmitRunError(context.Background(), errPayload, &runID) // Use background context for last attempt
			}
			return errors.Wrap(agentErr, "failed to run agent")
		} else {
			log.Info().Str("agentType", wac.AgentCommand.AgentType).Str("runID", runID).Msg("WriterAgent logic finished")
		}

		// Write the final result string to the provided writer
		_, writeErr := fmt.Fprintln(w, resultStr)
		if writeErr != nil {
			log.Error().Err(writeErr).Str("agentType", wac.AgentCommand.AgentType).Str("runID", runID).Msg("Failed to write agent result")
			return errors.Wrap(writeErr, "failed to write agent result")
		}

		// Emit run finished event if possible
		if eb != nil {
			// TODO(manuel): Gather actual stats for RunFinishedPayload
			finPayload := &events.RunFinishedPayload{
				// Populate with actual stats later
			}
			_ = eb.EmitRunFinished(context.Background(), finPayload, &runID)
		}

		return nil
	})

	// Wait for all goroutines (router and agent runner)
	log.Info().Str("runID", runID).Msg("Waiting for agent and router shutdown (WriterAgent)")
	waitErr := eg.Wait()
	if waitErr != nil {
		if errors.Is(waitErr, context.Canceled) || errors.Is(waitErr, context.DeadlineExceeded) {
			log.Info().Str("runID", runID).Msg("Agent/Router execution cancelled or timed out (WriterAgent)")
			return nil
		}
		log.Error().Err(waitErr).Str("runID", runID).Msg("Agent execution or router shutdown failed (WriterAgent)")
		return errors.Wrap(waitErr, "agent execution or router shutdown failed")
	}

	log.Info().Str("runID", runID).Msg("Agent and router shut down successfully (WriterAgent)")
	return nil
}

// renderInitialPrompt renders the command's Prompt template string
// using the templating library, not emrichen
func (a *AgentCommand) renderInitialPrompt(parsedLayers *layers.ParsedLayers) (string, error) {
	if a.Prompt == "" {
		return "", errors.New("cannot run agent without a prompt template defined in the command")
	}

	// Simple string replace with parameter values
	prompt := a.Prompt
	params := parsedLayers.GetDataMap()

	// Use the glazed templating package with Sprig functions
	tmpl := templating.CreateTemplate("prompt")
	tmpl, err := tmpl.Parse(prompt)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse prompt template")
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, params)
	if err != nil {
		return "", errors.Wrap(err, "failed to execute prompt template")
	}

	return buf.String(), nil
}

// Implementation of the Command interface

// GetCommandDescription returns the command description
func (a *AgentCommand) GetCommandDescription() *cmds.CommandDescription {
	return a.CommandDescription
}

// GetAgentType returns the agent type
func (a *AgentCommand) GetAgentType() string {
	return a.AgentType
}

// GetSystemPrompt returns the system prompt
func (a *AgentCommand) GetSystemPrompt() string {
	return a.SystemPrompt
}

// GetPrompt returns the prompt template
func (a *AgentCommand) GetPrompt() string {
	return a.Prompt
}

// GetTools returns the tools to use
func (a *AgentCommand) GetTools() []string {
	return a.Tools
}

// RenderAgentOptions renders the agent options using emrichen
func (a *AgentCommand) RenderAgentOptions(parameters map[string]interface{}, tags map[string]interface{}) (map[string]interface{}, error) {
	if a.AgentOptions == nil {
		return map[string]interface{}{}, nil
	}

	// Convert tags to emrichen tag funcs
	additionalTags := map[string]emrichen.TagFunc{}
	for name, tag := range tags {
		if tagFunc, ok := tag.(emrichen.TagFunc); ok {
			additionalTags[name] = tagFunc
		}
	}

	// Create an emrichen interpreter
	ei, err := emrichen.NewInterpreter(
		emrichen.WithVars(parameters),
		emrichen.WithAdditionalTags(additionalTags),
	)
	if err != nil {
		return nil, errors.Wrap(err, "could not create emrichen interpreter")
	}

	// Process the raw node
	v, err := ei.Process(a.AgentOptions.GetNode())
	if err != nil {
		return nil, errors.Wrap(err, "could not process emrichen node for agent options")
	}

	// Decode the processed node into a map
	result := map[string]interface{}{}
	err = v.Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode agent options node")
	}

	return result, nil
}
