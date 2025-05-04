package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/geppetto/pkg/conversation"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-agent/goagent/llm"
	"github.com/go-go-golems/go-go-agent/goagent/types"
	"github.com/go-go-golems/go-go-agent/pkg/eventbus"
	events "github.com/go-go-golems/go-go-agent/proto"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// PlanAndExecuteAgentFactory creates PlanAndExecuteAgent instances.
type PlanAndExecuteAgentFactory struct {
	planningModel  llm.LLM
	executionModel llm.LLM
}

func NewPlanAndExecuteAgentFactory(
	planningModel llm.LLM,
	executionModel llm.LLM,
) *PlanAndExecuteAgentFactory {
	return &PlanAndExecuteAgentFactory{
		planningModel:  planningModel,
		executionModel: executionModel,
	}
}

const PlanAndExecuteAgentType = "plan-execute" // Define the type constant

// PlanAndExecuteAgentSettings holds configuration for the PlanAndExecuteAgent.
type PlanAndExecuteAgentSettings struct {
	MaxIterations int `glazed.parameter:"max-iterations"`
	// TODO(manuel): Add separate parameters for planner and executor LLMs if needed
}

// NewAgent creates a new PlanAndExecuteAgent.
func (f *PlanAndExecuteAgentFactory) NewAgent(
	ctx context.Context,
	cmd Command,
	parsedLayers *layers.ParsedLayers,
	baseModel llm.LLM,
) (Agent, error) {
	var settings PlanAndExecuteAgentSettings
	err := parsedLayers.InitializeStruct(PlanAndExecuteAgentType, &settings)
	if err != nil {
		return nil, err
	}

	agentOptions, err := cmd.RenderAgentOptions(parsedLayers.GetDataMap(), nil)
	if err != nil {
		return nil, err
	}

	// Example of using agent options from command
	if cmdMaxIter, ok := agentOptions["max-iterations"]; ok {
		if maxIter, ok := cmdMaxIter.(int); ok {
			settings.MaxIterations = maxIter
		}
	}

	// Use the provided models from factory
	planningModel := f.planningModel
	executionModel := f.executionModel

	// If no dedicated models are set in the factory, use the base model for both
	if planningModel == nil {
		planningModel = baseModel
	}
	if executionModel == nil {
		executionModel = baseModel
	}

	return NewPlanAndExecuteAgent(
		WithExecutorLLM(planningModel),
		WithExecutorLLM(executionModel),
		WithExecutorMaxPlanningLoops(settings.MaxIterations),
	)
}

// CreateLayers defines the Glazed parameter layers for the PlanAndExecuteAgent.
func (f *PlanAndExecuteAgentFactory) CreateLayers() ([]layers.ParameterLayer, error) {
	agentLayer, err := layers.NewParameterLayer(
		PlanAndExecuteAgentType,
		"Plan-and-Execute agent configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"max-iterations",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of planning/execution iterations"),
				parameters.WithDefault(5), // Lower default for plan-execute?
			),
			// Add parameters for planner/executor models here if needed
		),
	)
	if err != nil {
		return nil, err
	}
	return []layers.ParameterLayer{agentLayer}, nil
}

// Plan represents a single step in the execution plan.
type Plan struct {
	ID          string `json:"id"`          // Unique identifier for the step
	Thought     string `json:"thought"`     // Reasoning behind the step
	Action      string `json:"action"`      // Tool or action to perform (e.g., Search, FinalAnswer)
	ActionInput string `json:"actionInput"` // Input for the action/tool
	Result      string `json:"result"`      // Result of executing the action (populated later)
	IsFinal     bool   `json:"isFinal"`     // Indicates if this is the final step
}

// PlanAndExecuteAgent implements a plan-and-execute agent.
// It first creates a plan using an LLM and then executes each step.
type PlanAndExecuteAgent struct {
	*BaseAgent
	LLM              llm.LLM
	PlannerPrompt    string // Template for the planning prompt
	ExecutorPrompt   string // Template for the executor prompt (if needed, maybe just use tool results)
	MaxPlanningLoops int    // Max attempts to create a valid plan
	currentStep      int
	// Optional event bus
	eventBus *eventbus.EventBus
	runID    *string
}

// PlanAndExecuteAgentOption defines functional options.
type PlanAndExecuteAgentOption func(*PlanAndExecuteAgent)

// WithExecutorLLM sets the LLM for the agent (used for both planning and potentially execution).
func WithExecutorLLM(llm llm.LLM) PlanAndExecuteAgentOption {
	return func(a *PlanAndExecuteAgent) {
		a.LLM = llm
	}
}

// WithExecutorPlannerPrompt sets the planning prompt template.
func WithExecutorPlannerPrompt(prompt string) PlanAndExecuteAgentOption {
	return func(a *PlanAndExecuteAgent) {
		a.PlannerPrompt = prompt
	}
}

// WithExecutorExecutorPrompt sets the executor prompt template (currently unused).
func WithExecutorExecutorPrompt(prompt string) PlanAndExecuteAgentOption {
	return func(a *PlanAndExecuteAgent) {
		a.ExecutorPrompt = prompt // Currently unused
	}
}

// WithExecutorMaxPlanningLoops sets the maximum planning attempts.
func WithExecutorMaxPlanningLoops(maxLoops int) PlanAndExecuteAgentOption {
	return func(a *PlanAndExecuteAgent) {
		a.MaxPlanningLoops = maxLoops
	}
}

// WithExecutorEventBus configures the agent with an event bus.
func WithExecutorEventBus(eb *eventbus.EventBus, runID string) PlanAndExecuteAgentOption {
	return func(a *PlanAndExecuteAgent) {
		a.eventBus = eb
		rid := runID
		a.runID = &rid
	}
}

// NewPlanAndExecuteAgent creates a new PlanAndExecuteAgent.
func NewPlanAndExecuteAgent(options ...PlanAndExecuteAgentOption) (*PlanAndExecuteAgent, error) {
	a := &PlanAndExecuteAgent{
		MaxPlanningLoops: 3, // Default planning attempts
	}
	for _, option := range options {
		option(a)
	}

	if a.LLM == nil {
		return nil, errors.New("LLM must be provided")
	}
	if a.PlannerPrompt == "" {
		// Provide a default planner prompt if none is set
		a.PlannerPrompt = defaultPlannerPromptTemplate()
		log.Info().Msg("Using default planner prompt template")
	}

	return a, nil
}

func defaultPlannerPromptTemplate() string {
	// Simple default template, assumes tools are injected elsewhere or not needed for planning
	return `Create a step-by-step plan to achieve the following goal:
Goal: {{.goal}}

Output the plan as a JSON list of objects, each with "id", "thought", "action", and "actionInput" fields. The final step's action should be "FinalAnswer".

Example:
[
  {
    "id": "1",
    "thought": "I need to search for the weather.",
    "action": "Search",
    "actionInput": "weather in Paris"
  },
  {
    "id": "2",
    "thought": "I have the weather information.",
    "action": "FinalAnswer",
    "actionInput": "The weather in Paris is sunny."
  }
]

Plan:
`
}

// Run executes the plan-and-execute loop.
func (a *PlanAndExecuteAgent) Run(ctx context.Context, goal string) (string, error) {
	var plan []*Plan
	var err error

	// 1. Planning Phase
	for i := 0; i < a.MaxPlanningLoops; i++ {
		a.currentStep = i + 1 // Use currentStep for planning phase steps/attempts
		log.Info().Ctx(ctx).Int("attempt", a.currentStep).Str("goal", goal).Msg("Planning phase attempt")
		// --- Emit StepStarted Event (Planning Attempt) ---
		if a.eventBus != nil {
			stepPayload := &events.StepStartedPayload{
				Step:     int32(a.currentStep),
				NodeId:   "plan-execute-planner",
				NodeGoal: fmt.Sprintf("Plan for: %s", goal),
				RootId:   *a.runID,
			}
			errEmit := a.eventBus.EmitStepStarted(ctx, stepPayload, a.runID)
			if errEmit != nil {
				log.Warn().Err(errEmit).Msg("Failed to emit StepStarted event for planning")
			}
		}
		startTime := time.Now()

		plan, err = a.createPlan(ctx, goal)
		stepStatus := "PLAN_CREATED"
		if err != nil {
			log.Warn().Ctx(ctx).Int("attempt", a.currentStep).Err(err).Msg("Failed to create plan, retrying")
			if err := a.memory.Add(ctx, types.MemoryEntry{
				ID:      fmt.Sprintf("plan-execute-planner-error-%d", a.currentStep),
				Content: fmt.Sprintf("Planning Error (%d): %v", a.currentStep, err),
				Metadata: map[string]string{
					"role": "planner",
				},
			}); err != nil {
				log.Warn().Err(err).Msg("Failed to add memory entry for planning error")
			}
			stepStatus = "PLAN_ERROR"
		}

		// --- Emit StepFinished Event (Planning Attempt) ---
		if a.eventBus != nil {
			stepFinishedPayload := &events.StepFinishedPayload{
				Step:            int32(a.currentStep),
				NodeId:          "plan-execute-planner",
				ActionName:      "CreatePlan",
				StatusAfter:     stepStatus,
				DurationSeconds: time.Since(startTime).Seconds(),
			}
			errEmit := a.eventBus.EmitStepFinished(ctx, stepFinishedPayload, a.runID)
			if errEmit != nil {
				log.Warn().Err(errEmit).Msg("Failed to emit StepFinished event for planning")
			}
		}

		if err == nil {
			break // Plan created successfully
		}
	}

	if err != nil {
		return "", errors.Wrapf(err, "failed to create plan after %d attempts", a.MaxPlanningLoops)
	}

	log.Info().Ctx(ctx).Int("steps", len(plan)).Msg("Plan created successfully")
	if err := a.memory.Add(ctx, types.MemoryEntry{
		ID:      fmt.Sprintf("plan-execute-planner-plan-%d", a.currentStep),
		Content: fmt.Sprintf("Plan Created: %v", plan),
		Metadata: map[string]string{
			"role": "planner",
		},
	}); err != nil {
		log.Warn().Err(err).Msg("Failed to add memory entry for plan creation")
	}
	for _, p := range plan {
		if err := a.memory.Add(ctx, types.MemoryEntry{
			ID:      fmt.Sprintf("plan-execute-planner-plan-%d", a.currentStep),
			Content: fmt.Sprintf(" - Step %s: %s(%s) - Thought: %s", p.ID, p.Action, p.ActionInput, p.Thought),
			Metadata: map[string]string{
				"role": "planner",
			},
		}); err != nil {
			log.Warn().Err(err).Msg("Failed to add memory entry for plan step")
		}
	}

	// 2. Execution Phase
	log.Info().Ctx(ctx).Msg("Starting execution phase")
	for idx, step := range plan {
		a.currentStep = idx + 1 // Reset currentStep for execution phase
		log.Info().Ctx(ctx).Int("step", a.currentStep).Str("action", step.Action).Str("id", step.ID).Msg("Executing plan step")

		// --- Emit StepStarted Event (Execution Step) ---
		if a.eventBus != nil {
			stepPayload := &events.StepStartedPayload{
				Step:     int32(a.currentStep),
				NodeId:   fmt.Sprintf("plan-execute-step-%s", step.ID),
				NodeGoal: fmt.Sprintf("%s: %s", step.Action, step.ActionInput),
				RootId:   *a.runID,
			}
			errEmit := a.eventBus.EmitStepStarted(ctx, stepPayload, a.runID)
			if errEmit != nil {
				log.Warn().Err(errEmit).Msg("Failed to emit StepStarted event for execution")
			}
		}
		startTime := time.Now()

		if step.IsFinal || step.Action == "FinalAnswer" {
			log.Info().Ctx(ctx).Int("step", a.currentStep).Str("finalAnswer", step.ActionInput).Msg("Reached final answer")
			if err := a.memory.Add(ctx, types.MemoryEntry{
				ID:      fmt.Sprintf("plan-execute-step-%d", a.currentStep),
				Content: fmt.Sprintf("Execution Step %d: Final Answer: %s", a.currentStep, step.ActionInput),
				Metadata: map[string]string{
					"role": "executor",
				},
			}); err != nil {
				log.Warn().Err(err).Msg("Failed to add memory entry for final answer")
			}
			step.Result = step.ActionInput // Store final answer as result
			// --- Emit StepFinished Event (Final Execution Step) ---
			if a.eventBus != nil {
				stepFinishedPayload := &events.StepFinishedPayload{
					Step:            int32(a.currentStep),
					NodeId:          fmt.Sprintf("plan-execute-step-%s", step.ID),
					ActionName:      "FinalAnswer",
					StatusAfter:     "FINISH",
					DurationSeconds: time.Since(startTime).Seconds(),
				}
				errEmit := a.eventBus.EmitStepFinished(ctx, stepFinishedPayload, a.runID)
				if errEmit != nil {
					log.Warn().Err(errEmit).Msg("Failed to emit StepFinished event for final execution step")
				}
			}
			return step.Result, nil
		}

		// Execute tool
		toolResult, toolErr := a.executeTool(ctx, step.Action, step.ActionInput, step.ID) // Pass step ID
		stepStatus := "TOOL_EXECUTED"
		if toolErr != nil {
			log.Error().Ctx(ctx).Int("step", a.currentStep).Err(toolErr).Str("action", step.Action).Msg("Tool execution failed")
			if err := a.memory.Add(ctx, types.MemoryEntry{
				ID:      fmt.Sprintf("plan-execute-step-%d", a.currentStep),
				Content: fmt.Sprintf("Execution Step %d: Tool Error: %v", a.currentStep, toolErr),
				Metadata: map[string]string{
					"role": "executor",
				},
			}); err != nil {
				log.Warn().Err(err).Msg("Failed to add memory entry for tool error")
			}
			step.Result = fmt.Sprintf("Error: %v", toolErr)
			stepStatus = "TOOL_ERROR"
			// Decide whether to stop execution on tool error
			// For now, we stop and return the error
			// --- Emit StepFinished Event (Error Execution Step) ---
			if a.eventBus != nil {
				stepFinishedPayload := &events.StepFinishedPayload{
					Step:            int32(a.currentStep),
					NodeId:          fmt.Sprintf("plan-execute-step-%s", step.ID),
					ActionName:      step.Action,
					StatusAfter:     stepStatus,
					DurationSeconds: time.Since(startTime).Seconds(),
				}
				errEmit := a.eventBus.EmitStepFinished(ctx, stepFinishedPayload, a.runID)
				if errEmit != nil {
					log.Warn().Err(errEmit).Msg("Failed to emit StepFinished event for error execution step")
				}
			}
			return "", errors.Wrapf(toolErr, "tool execution failed at step %d (%s)", a.currentStep, step.ID)
		} else {
			log.Info().Ctx(ctx).Int("step", a.currentStep).Str("result", toolResult).Msg("Tool executed successfully")
			if err := a.memory.Add(ctx, types.MemoryEntry{
				ID:      fmt.Sprintf("plan-execute-step-%d", a.currentStep),
				Content: fmt.Sprintf("Execution Step %d: Observation: %s", a.currentStep, toolResult),
				Metadata: map[string]string{
					"role": "executor",
				},
			}); err != nil {
				log.Warn().Err(err).Msg("Failed to add memory entry for tool observation")
			}
			step.Result = toolResult
		}

		// --- Emit StepFinished Event (Successful Execution Step) ---
		if a.eventBus != nil {
			stepFinishedPayload := &events.StepFinishedPayload{
				Step:            int32(a.currentStep),
				NodeId:          fmt.Sprintf("plan-execute-step-%s", step.ID),
				ActionName:      step.Action,
				StatusAfter:     stepStatus,
				DurationSeconds: time.Since(startTime).Seconds(),
			}
			errEmit := a.eventBus.EmitStepFinished(ctx, stepFinishedPayload, a.runID)
			if errEmit != nil {
				log.Warn().Err(errEmit).Msg("Failed to emit StepFinished event for execution step")
			}
		}
	}

	// If loop finishes without a FinalAnswer step
	log.Warn().Ctx(ctx).Msg("Agent finished execution loop without a FinalAnswer step")
	// Return the result of the last step?
	if len(plan) > 0 {
		return plan[len(plan)-1].Result, errors.New("agent finished without explicit FinalAnswer")
	}
	return "", errors.New("agent finished without completing any steps or providing a FinalAnswer")
}

// createPlan uses the LLM to generate a plan based on the goal.
func (a *PlanAndExecuteAgent) createPlan(ctx context.Context, goal string) ([]*Plan, error) {
	// Simple prompt rendering (replace with proper templating if needed)
	prompt := strings.ReplaceAll(a.PlannerPrompt, "{{.goal}}", goal)

	messages := []*conversation.Message{
		conversation.NewChatMessage(conversation.RoleSystem, prompt),
		conversation.NewChatMessage(conversation.RoleUser, goal),
	}

	if err := a.memory.Add(ctx, types.MemoryEntry{
		ID:      fmt.Sprintf("plan-execute-planner-prompt-%d", a.currentStep),
		Content: fmt.Sprintf("Planning Prompt: %s", prompt),
		Metadata: map[string]string{
			"role": "planner",
		},
	}); err != nil {
		log.Warn().Err(err).Msg("Failed to add memory entry for planning prompt")
	}

	response, err := a.llm.Generate(ctx, messages)
	if err != nil {
		return nil, errors.Wrap(err, "LLM call failed during planning")
	}

	planContent := response.Content.String()

	// Parse the plan from the response (assuming JSON list)
	var plan []*Plan
	// Try to extract JSON block if LLM adds surrounding text
	startIndex := strings.Index(planContent, "[")
	endIndex := strings.LastIndex(planContent, "]")
	if startIndex == -1 || endIndex == -1 || endIndex < startIndex {
		return nil, errors.Errorf("failed to find JSON list in LLM response: %s", planContent)
	}
	jsonPlan := planContent[startIndex : endIndex+1]

	err = json.Unmarshal([]byte(jsonPlan), &plan)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse JSON plan: %s", jsonPlan)
	}

	// Validate the plan (basic checks)
	if len(plan) == 0 {
		return nil, errors.New("LLM generated an empty plan")
	}
	for _, step := range plan {
		if step.ID == "" || step.Action == "" {
			return nil, errors.Errorf("invalid plan step: missing ID or Action in %+v", step)
		}
		// Check if action is valid? Requires tool knowledge here.
		if step.Action != "FinalAnswer" {
			_, err := a.tools.ExecuteTool(ctx, step.Action, step.ActionInput)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to execute tool %s with input %s", step.Action, step.ActionInput)
			}
		}
	}

	// Mark the last step as final if not already identified
	foundFinal := false
	for _, step := range plan {
		if step.Action == "FinalAnswer" {
			step.IsFinal = true
			foundFinal = true
			break
		}
	}
	if !foundFinal && len(plan) > 0 {
		log.Warn().Msg("Plan does not explicitly contain 'FinalAnswer' action. Marking last step as final.")
		plan[len(plan)-1].IsFinal = true
	}

	return plan, nil
}

// executeTool finds and runs the specified tool.
func (a *PlanAndExecuteAgent) executeTool(ctx context.Context, toolName, toolInput, stepID string) (string, error) {
	toolCallID := uuid.New().String()
	nodeID := fmt.Sprintf("plan-execute-step-%s", stepID)
	// --- Emit ToolInvoked Event ---
	if a.eventBus != nil {
		invokePayload := &events.ToolInvokedPayload{
			ToolName:    toolName,
			ApiName:     "Run",
			ArgsSummary: toolInput,
			NodeId:      ptr(nodeID), // Associate with execution step node ID
			Step:        ptr(int32(a.currentStep)),
			AgentClass:  ptr("PlanAndExecuteAgent"),
			ToolCallId:  toolCallID,
		}
		err := a.eventBus.EmitToolInvoked(ctx, invokePayload, a.runID)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to emit ToolInvoked event")
		}
	}
	startTime := time.Now()

	// Execute the tool
	result, err := a.tools.ExecuteTool(ctx, toolName, toolInput)
	if err != nil {
		return "", errors.Wrapf(err, "failed to execute tool %s with input %s", toolName, toolInput)
	}

	duration := time.Since(startTime)

	// --- Emit ToolReturned Event ---
	if a.eventBus != nil {
		state := "SUCCESS"
		var errorStr *string
		resultSummary := result
		if err != nil {
			state = "ERROR"
			es := err.Error()
			errorStr = &es
			resultSummary = "Error: " + es
		}

		returnPayload := &events.ToolReturnedPayload{
			ToolName:        toolName,
			ApiName:         "Run",
			State:           state,
			DurationSeconds: duration.Seconds(),
			ResultSummary:   resultSummary,
			Error:           errorStr,
			NodeId:          ptr(nodeID),
			Step:            ptr(int32(a.currentStep)),
			AgentClass:      ptr("PlanAndExecuteAgent"),
			ToolCallId:      toolCallID,
		}
		errEmit := a.eventBus.EmitToolReturned(ctx, returnPayload, a.runID)
		if errEmit != nil {
			log.Warn().Err(errEmit).Msg("Failed to emit ToolReturned event")
		}
	}

	return result, err
}

// Ensure PlanAndExecuteAgent implements the Agent interface
var _ Agent = (*PlanAndExecuteAgent)(nil)
