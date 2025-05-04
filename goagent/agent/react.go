package agent

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-go-golems/geppetto/pkg/conversation"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-agent/goagent/llm"
	"github.com/go-go-golems/go-go-agent/goagent/memory"
	"github.com/go-go-golems/go-go-agent/goagent/tools"
	"github.com/go-go-golems/go-go-agent/goagent/types"
	"github.com/go-go-golems/go-go-agent/pkg/eventbus"
	events "github.com/go-go-golems/go-go-agent/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type ReactAgentFactory struct {
	LLM llm.LLM
}

func (f *ReactAgentFactory) NewAgent(ctx context.Context, cmd Command, parsedLayers *layers.ParsedLayers, baseModel llm.LLM) (Agent, error) {
	return NewReActAgent(WithLLM(f.LLM))
}

func (f *ReactAgentFactory) CreateLayers() ([]layers.ParameterLayer, error) {
	agentLayer, err := layers.NewParameterLayer(
		ReactAgentType,
		"React agent configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"max-iterations",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of ReAct iterations"),
				parameters.WithDefault(10),
			),
			parameters.NewParameterDefinition(
				"max-tool-calls",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of tool calls per iteration"),
				parameters.WithDefault(5),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	return []layers.ParameterLayer{agentLayer}, nil
}

// ReActAgent implements the ReAct (Reason + Act) agent loop.
// See: https://react-lm.github.io/
type ReActAgent struct {
	*BaseAgent       // Embed BaseAgent for tools and memory
	LLM              llm.LLM
	SystemPrompt     string
	MaxIterations    int
	MaxToolCalls     int // Maximum tool calls per iteration
	currentIteration int
	// Optional event bus
	eventBus *eventbus.EventBus
	runID    *string
	// Removed scratchpad *helpers.RingBuffer[string]
}

const ReactAgentType = "react"

// ReActAgentOption defines functional options for configuring ReActAgent.
type ReActAgentOption func(*ReActAgent)

// WithLLM sets the language model for the agent.
func WithLLM(llm llm.LLM) ReActAgentOption {
	return func(a *ReActAgent) {
		a.LLM = llm
	}
}

// WithSystemPrompt sets the initial system prompt.
func WithSystemPrompt(prompt string) ReActAgentOption {
	return func(a *ReActAgent) {
		a.SystemPrompt = prompt
	}
}

// WithTools sets the available tools for the agent.
func WithTools(toolExecutor tools.ToolExecutor) ReActAgentOption { // Use interface type
	return func(a *ReActAgent) {
		if a.BaseAgent == nil {
			a.BaseAgent = &BaseAgent{}
		}
		a.tools = &toolExecutor // Set on BaseAgent
	}
}

// WithMemory sets the memory system for the agent.
func WithMemory(mem memory.Memory) ReActAgentOption {
	return func(a *ReActAgent) {
		if a.BaseAgent == nil {
			a.BaseAgent = &BaseAgent{}
		}
		a.memory = mem // Set on BaseAgent
	}
}

// WithMaxIterations sets the maximum number of ReAct iterations.
func WithMaxIterations(maxIter int) ReActAgentOption {
	return func(a *ReActAgent) {
		a.MaxIterations = maxIter
	}
}

// WithMaxToolCalls sets the maximum number of tool calls per iteration.
func WithMaxToolCalls(maxCalls int) ReActAgentOption {
	return func(a *ReActAgent) {
		a.MaxToolCalls = maxCalls
	}
}

// WithEventBus configures the agent with an event bus.
func WithEventBus(eb *eventbus.EventBus, runID string) ReActAgentOption {
	return func(a *ReActAgent) {
		a.eventBus = eb
		rid := runID
		a.runID = &rid
	}
}

// NewReActAgent creates a new ReActAgent with the given options.
func NewReActAgent(options ...ReActAgentOption) (*ReActAgent, error) {
	a := &ReActAgent{
		MaxIterations: 10,
		MaxToolCalls:  5,
		// Initialize BaseAgent (sensible defaults)
		BaseAgent: &BaseAgent{
			tools:  tools.NewToolExecutor(), // Default empty executor (assuming this constructor exists)
			memory: memory.NewMockMemory(),  // Use MockMemory as set by user
		},
		// Removed scratchpad initialization
	}
	for _, option := range options {
		option(a)
	}

	if a.LLM == nil {
		return nil, errors.New("LLM must be provided")
	}
	// Ensure BaseAgent fields are initialized if options didn't provide them
	if a.tools == nil {
		a.tools = tools.NewToolExecutor() // Assuming this constructor exists
	}
	if a.memory == nil {
		a.memory = memory.NewMockMemory() // Use MockMemory as set by user
	}

	return a, nil
}

// formatHistoryForPrompt retrieves recent memory and formats it for the ReAct prompt.
func (a *ReActAgent) formatHistoryForPrompt(ctx context.Context) (string, error) {
	// TODO(manuel): Define a proper method in the Memory interface (e.g., GetHistory)
	// For now, assume a simple GetMessages or similar exists, or retrieve all.
	// Let's assume GetMessages(limit int) exists for simplicity
	messages, err := a.memory.Get(ctx, []string{})
	if err != nil {
		// If GetMessages doesn't exist or fails, return empty string or handle differently
		log.Warn().Err(err).Msg("Could not retrieve history from memory for prompt")
		// return "", errors.Wrap(err, "failed to get history from memory")
		return "(Could not retrieve history)", nil // Allow continuing without history
	}

	var history strings.Builder
	for i := len(messages) - 1; i >= 0; i-- { // Reverse order for chronological display
		msg := messages[i]
		role := msg.Metadata["role"] // Assuming role is stored in metadata
		content := msg.Content
		// Simple formatting, similar to scratchpad
		switch role {
		case "user":
			history.WriteString(fmt.Sprintf("Prompt: %s\n", content)) // Distinguish initial prompt?
		case "assistant":
			// Check if it was a thought or action or final answer based on content patterns
			if strings.HasPrefix(content, "Thought: ") {
				history.WriteString(fmt.Sprintf("%s\n", content))
			} else if strings.HasPrefix(content, "Action: ") {
				history.WriteString(fmt.Sprintf("%s\n", content))
			} else if strings.HasPrefix(content, "Final Answer: ") {
				history.WriteString(fmt.Sprintf("%s\n", content))
			} else {
				history.WriteString(fmt.Sprintf("LLM Response: %s\n", content))
			}
		case "tool":
			// Extract tool name if possible from content like "Observation [tool]: result"
			history.WriteString(fmt.Sprintf("Observation: %s\n", content)) // Simplify for now
		case "system":
			// Maybe include system messages like errors?
			if strings.HasPrefix(content, "Error parsing response") || strings.HasPrefix(content, "Tool Error") {
				history.WriteString(fmt.Sprintf("%s\n", content))
			}
		default:
			history.WriteString(fmt.Sprintf("%s: %s\n", role, content))
		}
	}
	return history.String(), nil
}

// Run executes the ReAct agent loop for a given initial prompt.
func (a *ReActAgent) Run(ctx context.Context, goal string) (string, error) {
	a.currentIteration = 0
	// Removed a.scratchpad.Add(fmt.Sprintf("Prompt: %s", prompt))
	if err := a.memory.Add(ctx, types.MemoryEntry{ // Add to main memory as well
		ID:      fmt.Sprintf("react-prompt-%d", a.currentIteration),
		Content: goal, // Store the goal/prompt directly
		Metadata: map[string]string{
			"role": "user",
		},
	}); err != nil {
		log.Warn().Err(err).Msg("Failed to add memory entry for initial prompt")
	}

	// Build initial messages using conversation package
	messages := []*conversation.Message{
		conversation.NewChatMessage(conversation.RoleSystem, a.SystemPrompt),
		conversation.NewChatMessage(conversation.RoleUser, goal),
	}

	for i := 0; i < a.MaxIterations; i++ {
		a.currentIteration = i + 1
		log.Info().Ctx(ctx).Int("iteration", a.currentIteration).Msg("Starting ReAct iteration")

		// --- Emit StepStarted Event ---
		if a.eventBus != nil {
			stepPayload := &events.StepStartedPayload{
				Step:     int32(a.currentIteration),
				NodeId:   "react-agent",
				NodeGoal: goal, // Use initial prompt as goal for now
				RootId:   *a.runID,
			}
			err := a.eventBus.EmitStepStarted(ctx, stepPayload, a.runID)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to emit StepStarted event")
			}
		}
		startTime := time.Now()

		// Add history context from memory to messages
		historyContent, err := a.formatHistoryForPrompt(ctx)
		if err != nil {
			// Decide how to handle: stop or continue without history?
			return "", errors.Wrap(err, "failed to format history for prompt")
		}
		if historyContent != "" {
			// Add history as a user message, simulating the scratchpad context
			messages = append(messages, conversation.NewChatMessage(conversation.RoleUser, "Current History:\n"+historyContent))
		}

		// Reason + Act - Use Generate
		// Assuming LLM interface has Generate(ctx, []*conversation.Message) (*conversation.Message, error)
		response, err := a.LLM.Generate(ctx, messages)
		if err != nil {
			// Log error to memory?
			if err := a.memory.Add(ctx, types.MemoryEntry{
				ID:       fmt.Sprintf("react-llm-error-%d", a.currentIteration),
				Content:  fmt.Sprintf("LLM Error: %v", err),
				Metadata: map[string]string{"role": "system"},
			}); err != nil {
				log.Warn().Err(err).Msg("Failed to add memory entry for LLM error")
			}
			return "", errors.Wrapf(err, "LLM generate failed in iteration %d", a.currentIteration)
		}
		// Assuming response is *conversation.Message and Content is a *conversation.Content
		// Need to handle potential multiple content parts if necessary
		responseContent := ""
		if response != nil && response.Content != nil {
			responseContent = response.Content.String() // Assuming String() method exists
		} else {
			log.Warn().Msg("LLM returned nil response or content")
			// Handle empty response? Add to memory and continue?
			if err := a.memory.Add(ctx, types.MemoryEntry{
				ID:       fmt.Sprintf("react-llm-empty-%d", a.currentIteration),
				Content:  "LLM returned empty response",
				Metadata: map[string]string{"role": "system"},
			}); err != nil {
				log.Warn().Err(err).Msg("Failed to add memory entry for empty response")
			}
			continue
		}

		// Add assistant response to messages for next turn
		messages = append(messages, response)
		// Removed a.scratchpad.Add(fmt.Sprintf("LLM Response (%d): %s", a.currentIteration, responseContent))
		if err := a.memory.Add(ctx, types.MemoryEntry{
			ID:      fmt.Sprintf("react-llm-response-%d", a.currentIteration),
			Content: responseContent, // Store the raw response content
			Metadata: map[string]string{
				"role": string(conversation.RoleAssistant), // Use conversation role
			},
		}); err != nil {
			log.Warn().Err(err).Msg("Failed to add memory entry for LLM response")
		}

		// Check for final answer or tool calls
		thought, action, actionInput, isFinal, err := a.parseResponse(responseContent)
		if err != nil {
			// This error from parseResponse might now be less critical if thought-only is allowed
			log.Warn().Ctx(ctx).Err(err).Str("response", responseContent).Msg("Failed to parse LLM response, continuing loop")
			parseErrorMsg := fmt.Sprintf("Error parsing response (%d): %v", a.currentIteration, err)
			// Removed a.scratchpad.Add(parseErrorMsg)
			if err := a.memory.Add(ctx, types.MemoryEntry{
				ID:      fmt.Sprintf("react-parse-error-%d", a.currentIteration),
				Content: parseErrorMsg,
				Metadata: map[string]string{
					"role": "system",
				},
			}); err != nil {
				log.Warn().Err(err).Msg("Failed to add memory entry for parse error")
			}
			// Potentially add the raw response back as an observation? Or just continue?
			// messages = append(messages, conversation.NewChatMessage(conversation.RoleTool, fmt.Sprintf("Parsing Error: %v", err))) // Example
			continue // Try again in the next iteration
		}

		if thought != "" && thought != "(No specific thought found in response)" { // Only log meaningful thoughts
			// Removed a.scratchpad.Add(thoughtMsg)
			if err := a.memory.Add(ctx, types.MemoryEntry{
				ID:      fmt.Sprintf("react-thought-%d", a.currentIteration),
				Content: fmt.Sprintf("Thought: %s", thought),
				Metadata: map[string]string{
					"role": string(conversation.RoleAssistant), // LLM's thought
				},
			}); err != nil {
				log.Warn().Err(err).Msg("Failed to add memory entry for thought")
			}
		}

		if isFinal {
			log.Info().Ctx(ctx).Str("finalAnswer", actionInput).Int("iteration", a.currentIteration).Msg("ReAct agent finished")
			finalMsg := fmt.Sprintf("Final Answer: %s", actionInput)
			// Removed a.scratchpad.Add(finalMsg)
			if err := a.memory.Add(ctx, types.MemoryEntry{
				ID:      fmt.Sprintf("react-final-answer-%d", a.currentIteration),
				Content: finalMsg,
				Metadata: map[string]string{
					"role": string(conversation.RoleAssistant),
				},
			}); err != nil {
				log.Warn().Err(err).Msg("Failed to add memory entry for final answer")
			}
			// --- Emit StepFinished Event (Final) ---
			if a.eventBus != nil {
				stepFinishedPayload := &events.StepFinishedPayload{
					Step:            int32(a.currentIteration),
					NodeId:          "react-agent",
					ActionName:      "FinalAnswer",
					StatusAfter:     "FINISH",
					DurationSeconds: time.Since(startTime).Seconds(),
				}
				err := a.eventBus.EmitStepFinished(ctx, stepFinishedPayload, a.runID)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to emit StepFinished event")
				}
			}
			return actionInput, nil
		}

		// If no action and not final, it might be a thought-only response.
		// The loop continues, and the thought is already in memory and messages.
		if action == "" {
			log.Debug().Msg("No action found in response, continuing loop with thought.")
			// --- Emit StepFinished Event (Thought Only) ---
			if a.eventBus != nil {
				stepFinishedPayload := &events.StepFinishedPayload{
					Step:            int32(a.currentIteration),
					NodeId:          "react-agent",
					ActionName:      "Thought", // Indicate thought step
					StatusAfter:     "THOUGHT_RECEIVED",
					DurationSeconds: time.Since(startTime).Seconds(),
				}
				err := a.eventBus.EmitStepFinished(ctx, stepFinishedPayload, a.runID)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to emit StepFinished event for thought step")
				}
			}
			continue
		}

		// Execute Action (Tool Call)
		// Removed a.scratchpad.Add(actionMsg)
		if err := a.memory.Add(ctx, types.MemoryEntry{
			ID:      fmt.Sprintf("react-action-%d", a.currentIteration),
			Content: fmt.Sprintf("Action: %s(%s)", action, actionInput),
			Metadata: map[string]string{
				"role": string(conversation.RoleAssistant), // LLM's requested action
			},
		}); err != nil {
			log.Warn().Err(err).Msg("Failed to add memory entry for action")
		}

		toolResult, toolErr := a.executeTool(ctx, action, actionInput)
		stepStatus := "TOOL_EXECUTED"
		observationMsg := "" // This will be the content for the tool message
		if toolErr != nil {
			log.Warn().Ctx(ctx).Err(toolErr).Str("tool", action).Msg("Tool execution failed")
			if err := a.memory.Add(ctx, types.MemoryEntry{
				ID:      fmt.Sprintf("react-tool-error-%d", a.currentIteration),
				Content: fmt.Sprintf("Tool Error [%s]: %v", action, toolErr), // More detailed in memory
				Metadata: map[string]string{
					"role": string(conversation.RoleTool),
				},
			}); err != nil {
				log.Warn().Err(err).Msg("Failed to add memory entry for tool error")
			}
			stepStatus = "TOOL_ERROR"
			observationMsg = fmt.Sprintf("Error: %v", toolErr) // Content for the tool message back to LLM
		} else {
			log.Info().Ctx(ctx).Str("tool", action).Str("result", toolResult).Msg("Tool executed successfully")
			observationMsg = toolResult // Use raw tool result for observation message
			// Removed a.scratchpad.Add(observationMsg)
			if err := a.memory.Add(ctx, types.MemoryEntry{
				ID:      fmt.Sprintf("react-observation-%d", a.currentIteration),
				Content: fmt.Sprintf("Observation [%s]: %s", action, toolResult), // More detailed in memory
				Metadata: map[string]string{
					"role": string(conversation.RoleTool),
				},
			}); err != nil {
				log.Warn().Err(err).Msg("Failed to add memory entry for observation")
			}
		}

		// Add tool result/error back to LLM context using conversation.Message
		// We need a tool call ID that matches the assistant's request if the LLM uses tool calling.
		// Since we parse action/input manually, we generate a new ID.
		// The role should be RoleTool.
		messages = append(messages, conversation.NewChatMessage(conversation.RoleTool, observationMsg))

		// --- Emit StepFinished Event (Iteration) ---
		if a.eventBus != nil {
			stepFinishedPayload := &events.StepFinishedPayload{
				Step:            int32(a.currentIteration),
				NodeId:          "react-agent",
				ActionName:      action,
				StatusAfter:     stepStatus,
				DurationSeconds: time.Since(startTime).Seconds(),
				// TODO(manuel): Add observation/result summary?
			}
			err := a.eventBus.EmitStepFinished(ctx, stepFinishedPayload, a.runID)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to emit StepFinished event")
			}
		}
	}

	return "", errors.Errorf("agent stopped after %d iterations", a.MaxIterations)
}

// executeTool finds and runs the specified tool using the BaseAgent's tool executor.
func (a *ReActAgent) executeTool(ctx context.Context, toolName, toolInput string) (string, error) {
	if a.tools == nil {
		return "", errors.New("tool executor not initialized in BaseAgent")
	}

	toolCallID := uuid.New().String()
	nodeID := "react-agent" // Keep it simple for now

	// --- Emit ToolInvoked Event ---
	if a.eventBus != nil {
		invokePayload := &events.ToolInvokedPayload{
			ToolName:    toolName,
			ApiName:     "Run", // Assuming Run method for tools used by executor
			ArgsSummary: toolInput,
			NodeId:      ptr(nodeID),
			Step:        ptr(int32(a.currentIteration)),
			AgentClass:  ptr("ReActAgent"),
			ToolCallId:  toolCallID,
		}
		err := a.eventBus.EmitToolInvoked(ctx, invokePayload, a.runID)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to emit ToolInvoked event")
		}
	}
	startTime := time.Now()

	// Execute the tool via the BaseAgent's executor
	result, err := a.tools.ExecuteTool(ctx, toolName, toolInput) // Use BaseAgent's tools
	duration := time.Since(startTime)

	// --- Emit ToolReturned Event ---
	if a.eventBus != nil {
		state := "SUCCESS"
		var errorStr *string
		resultSummary := result // Simple summary for now
		if err != nil {
			state = "ERROR"
			es := err.Error()
			errorStr = &es
			resultSummary = "Error: " + es // Provide error in summary if execution failed
		}

		returnPayload := &events.ToolReturnedPayload{
			ToolName:        toolName,
			ApiName:         "Run", // Assuming Run
			State:           state,
			DurationSeconds: duration.Seconds(),
			ResultSummary:   resultSummary,
			Error:           errorStr,
			NodeId:          ptr(nodeID),
			Step:            ptr(int32(a.currentIteration)),
			AgentClass:      ptr("ReActAgent"),
			ToolCallId:      toolCallID,
		}
		errEmit := a.eventBus.EmitToolReturned(ctx, returnPayload, a.runID)
		if errEmit != nil {
			log.Warn().Err(errEmit).Msg("Failed to emit ToolReturned event")
		}
	}

	// Check for specific tool execution error vs. general error
	if err != nil {
		return "", errors.Wrapf(err, "tool '%s' execution failed", toolName)
	}

	return result, nil
}

// Simplified parsing logic for ReAct response (Thought, Action, Action Input, Final Answer).
// Expects format like:
// Thought: I need to find the weather.
// Action: Search[weather in Berlin]
// or
// Thought: I have the final answer.
// Final Answer: The weather is sunny.
var (
	thoughtRegex = regexp.MustCompile(`(?is)Thought:\s*(.*?)(?:\nAction:|\nFinal Answer:|$)`)
	actionRegex  = regexp.MustCompile(`(?is)Action:\s*([a-zA-Z_][a-zA-Z0-9_]*?)\[(.*?)\]`)
	finalRegex   = regexp.MustCompile(`(?is)Final Answer:\s*(.*)`) // Captures everything after Final Answer:
)

func (a *ReActAgent) parseResponse(response string) (string, string, string, bool, error) {
	var thought, action, actionInput string
	var isFinal bool

	thoughtMatch := thoughtRegex.FindStringSubmatch(response)
	if len(thoughtMatch) > 1 {
		thought = strings.TrimSpace(thoughtMatch[1])
	} else {
		thought = "(No specific thought found in response)"
	}

	finalMatch := finalRegex.FindStringSubmatch(response)
	if len(finalMatch) > 1 {
		isFinal = true
		actionInput = strings.TrimSpace(finalMatch[1])
		action = "FinalAnswer"
		return thought, action, actionInput, isFinal, nil
	}

	actionMatch := actionRegex.FindStringSubmatch(response)
	if len(actionMatch) > 2 {
		action = strings.TrimSpace(actionMatch[1])
		actionInput = strings.TrimSpace(actionMatch[2])
		return thought, action, actionInput, isFinal, nil
	}

	// If neither action nor final answer is found, it might be a thought-only response.
	// Return empty action/input, the main loop will handle continuing.
	log.Debug().Str("response", response).Msg("Response did not contain Action or FinalAnswer, treating as thought-only")
	return thought, action, actionInput, isFinal, nil
}

// Helper function to create pointers for optional protobuf fields
func ptr[T any](v T) *T {
	return &v
}

// Ensure ReActAgent implements the Agent interface
var _ Agent = (*ReActAgent)(nil)
