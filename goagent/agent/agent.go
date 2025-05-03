package agent

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/go-go-agent/goagent/llm"
	"github.com/go-go-golems/go-go-agent/goagent/memory"
	"github.com/go-go-golems/go-go-agent/goagent/tools"
	"github.com/go-go-golems/go-go-agent/goagent/tracing"
)

// Command is the interface that agent command implementations must satisfy
// to be used with the agent factory
type Command interface {
	// Fields needed by agent implementations
	GetCommandDescription() *cmds.CommandDescription
	GetAgentType() string
	GetSystemPrompt() string
	GetPrompt() string
	GetTools() []string
	RenderAgentOptions(parameters map[string]interface{}, tags map[string]interface{}) (map[string]interface{}, error)
}

// Agent is the base interface that all agent implementations must satisfy
type Agent interface {
	// Run executes the agent with a given prompt and returns the result as a string
	Run(ctx context.Context, prompt string) (string, error)

	// AddTool adds a tool to the agent
	AddTool(tool tools.Tool) error

	// SetMemory sets the memory system for the agent
	SetMemory(mem memory.Memory) error
}

// WriterAgent is a marker interface for agents using the standard Run method for output.
type WriterAgent interface {
	Agent
}

// GlazedAgent is an agent that produces structured data output
type GlazedAgent interface {
	Agent
	// RunIntoGlazeProcessor processes the agent's output into a Glaze processor
	RunIntoGlazeProcessor(ctx context.Context, prompt string, gp middlewares.Processor) error
}

// Factory is responsible for creating agent instances
type Factory interface {
	// NewAgent creates a new agent instance
	NewAgent(ctx context.Context, cmd Command, parsedLayers *layers.ParsedLayers, llm llm.LLM) (Agent, error)
	// CreateLayers returns the parameter layers needed by this agent
	CreateLayers() ([]*layers.ParameterLayer, error)
}

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	llm     llm.LLM
	tools   *tools.ToolExecutor
	memory  memory.Memory
	tracer  tracing.Tracer
	maxIter int
}

// NewBaseAgent creates a new BaseAgent
func NewBaseAgent(llmModel llm.LLM, maxIterations int) *BaseAgent {
	return &BaseAgent{
		llm:     llmModel,
		tools:   tools.NewToolExecutor(),
		tracer:  tracing.NewSimpleTracer(),
		maxIter: maxIterations,
	}
}

// AddTool adds a tool to the agent
func (a *BaseAgent) AddTool(tool tools.Tool) error {
	a.tools.AddTool(tool)
	return nil
}

// SetMemory sets the memory system for the agent
func (a *BaseAgent) SetMemory(mem memory.Memory) error {
	a.memory = mem
	return nil
}

// GetTracer returns the tracer
func (a *BaseAgent) GetTracer() tracing.Tracer {
	return a.tracer
}
