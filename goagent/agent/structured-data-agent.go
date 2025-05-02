package agent

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/go-go-golems/geppetto/pkg/conversation"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	glazed_types "github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-agent/goagent/llm"
	"github.com/go-go-golems/go-go-agent/goagent/types"
	"github.com/pkg/errors"
)

// StructuredDataAgentFactory creates StructuredDataAgent instances.
type StructuredDataAgentFactory struct{}

var _ AgentFactory = &StructuredDataAgentFactory{}

const StructuredDataAgentType = "structured-data" // Define the type constant

// StructuredDataAgentSettings holds configuration for the StructuredDataAgent.
type StructuredDataAgentSettings struct {
	MaxIterations int    `glazed.parameter:"max-iterations"`
	TagNames      string `glazed.parameter:"tag-names"`
	OutputAsXML   bool   `glazed.parameter:"output-as-xml"`
}

// NewAgent creates a new StructuredDataAgent.
func (f *StructuredDataAgentFactory) NewAgent(ctx context.Context, parsedLayers *layers.ParsedLayers, llmModel llm.LLM) (Agent, error) {
	var settings StructuredDataAgentSettings
	err := parsedLayers.InitializeStruct(StructuredDataAgentType, &settings)
	if err != nil {
		return nil, err
	}

	// Convert comma-separated tag names to slice
	tagNames := []string{"results"} // Default to "results" tag
	if settings.TagNames != "" {
		tagNames = strings.Split(settings.TagNames, ",")
		// Trim whitespace from each tag name
		for i, tag := range tagNames {
			tagNames[i] = strings.TrimSpace(tag)
		}
	}

	return NewStructuredDataAgent(llmModel, &settings, tagNames), nil
}

// CreateLayers defines the Glazed parameter layers for the StructuredDataAgent.
func (f *StructuredDataAgentFactory) CreateLayers() ([]layers.ParameterLayer, error) {
	agentLayer, err := layers.NewParameterLayer(
		StructuredDataAgentType,
		"Structured Data agent configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"max-iterations",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of generation iterations"),
				parameters.WithDefault(1),
			),
			parameters.NewParameterDefinition(
				"tag-names",
				parameters.ParameterTypeString,
				parameters.WithHelp("Comma-separated list of XML tag names to extract (default: 'results')"),
				parameters.WithDefault("results"),
			),
			parameters.NewParameterDefinition(
				"output-as-xml",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Output the extracted data in XML format"),
				parameters.WithDefault(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}
	return []layers.ParameterLayer{agentLayer}, nil
}

// StructuredDataAgent implements an agent that extracts structured XML data
// from an LLM response.
type StructuredDataAgent struct {
	*BaseAgent                                 // Embed BaseAgent from goagent/agent
	tagNames      []string                     // The XML tag names to extract
	extractedData map[string]string            // Maps tag names to their extracted content
	settings      *StructuredDataAgentSettings // Added settings field
}

// NewStructuredDataAgent creates a new StructuredDataAgent.
func NewStructuredDataAgent(llmModel llm.LLM, settings *StructuredDataAgentSettings, tagNames []string) *StructuredDataAgent {
	return &StructuredDataAgent{
		BaseAgent:     NewBaseAgent(llmModel, settings.MaxIterations),
		tagNames:      tagNames,
		extractedData: make(map[string]string),
		settings:      settings,
	}
}

// GetExtractedData returns the extracted structured data collected during the last Run.
func (a *StructuredDataAgent) GetExtractedData() map[string]string {
	return a.extractedData
}

// buildSystemPrompt builds the system prompt for the StructuredDataAgent.
// It instructs the LLM to use specific XML tags for structured data.
func (a *StructuredDataAgent) buildSystemPrompt() string {
	// List all tags that the agent will look for
	tagList := strings.Join(a.tagNames, ", ")

	return fmt.Sprintf(`You are an AI assistant that analyzes input and responds with structured data in XML format.

When asked to analyze data, you MUST follow these instructions:
1. Thoroughly analyze the input according to the instructions provided.
2. Structure your findings using XML tags. You SHOULD use the following tags: %s.
3. Be thorough and detailed in your analysis, ensuring all required information is included.
4. Format your response as valid, well-formed XML.
5. Do not include any additional text or explanations outside of the XML structure.
6. Follow any specific output structure requested in the user prompt.

For example, if asked to extract information about programming languages, you might respond:
<results>
  <languages>
    <language>
      <name>Python</name>
      <features>
        <feature>Easy to learn</feature>
        <feature>Versatile</feature>
      </features>
    </language>
    <!-- More structured data here -->
  </languages>
</results>

Remember that your response should be thorough, accurate, and properly structured within the requested XML format.`, tagList)
}

// buildSummary generates a string summarizing the outcome.
func (a *StructuredDataAgent) buildSummary() string {
	if len(a.extractedData) == 0 {
		return "No structured data was extracted."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Extracted data from %d XML tag(s):\n\n", len(a.extractedData)))

	if a.settings.OutputAsXML {
		// Output the raw XML format
		for tagName, content := range a.extractedData {
			sb.WriteString(fmt.Sprintf("<%s>\n%s\n</%s>\n\n", tagName, content, tagName))
		}
	} else {
		// Output a more readable summary
		for tagName, content := range a.extractedData {
			// Calculate the size of the content
			lines := strings.Count(content, "\n") + 1
			sb.WriteString(fmt.Sprintf("- %s (%d lines, %d bytes)\n", tagName, lines, len(content)))
		}

		sb.WriteString("\nUse the returned data object or access the GetExtractedData() method to retrieve the full content.")
	}

	return sb.String()
}

// Run executes the core structured data extraction logic.
func (a *StructuredDataAgent) Run(ctx context.Context, input string) (string, error) {
	ctx, span := a.tracer.StartSpan(ctx, "StructuredDataAgent.Run")
	defer span.End()

	// Reset the extracted data map for this run
	a.extractedData = make(map[string]string)

	// Initialize conversation
	messages := []*conversation.Message{
		conversation.NewChatMessage(conversation.RoleSystem, a.buildSystemPrompt()),
		conversation.NewChatMessage(conversation.RoleUser, input),
	}

	// Generate response from LLM
	responseMsg, err := a.llm.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed: %w", err)
	}

	response := responseMsg.Content.String()

	// Extract structured data from the response
	a.extractedData = ExtractMultipleData(response, a.tagNames)

	// Log the extracted data
	tagCount := len(a.extractedData)
	a.tracer.LogEvent(ctx, types.Event{
		Type: "data_extraction_completed",
		Data: map[string]interface{}{
			"tagCount":  tagCount,
			"tagNames":  a.tagNames,
			"extracted": len(a.extractedData) > 0,
		},
		Timestamp: 0, // Will be set by the tracer
	})

	// Return a summary of the extracted data
	return a.buildSummary(), nil
}

// RunIntoWriter executes the agent and writes the summary string to the writer.
func (a *StructuredDataAgent) RunIntoWriter(ctx context.Context, input string, w io.Writer) error {
	ctx, span := a.tracer.StartSpan(ctx, "StructuredDataAgent.RunIntoWriter")
	defer span.End()

	summary, runErr := a.Run(ctx, input) // Call Run to get the summary and potential error

	_, writeErr := fmt.Fprintln(w, summary)
	if writeErr != nil {
		// Prioritize returning the write error if it occurs
		return errors.Wrap(writeErr, "failed to write summary")
	}

	// Return the error from the Run execution, if any
	return runErr
}

// RunIntoGlazeProcessor executes the agent and streams structured data as rows
// into the Glazed processor.
func (a *StructuredDataAgent) RunIntoGlazeProcessor(
	ctx context.Context,
	input string,
	gp middlewares.Processor,
) error {
	ctx, span := a.tracer.StartSpan(ctx, "StructuredDataAgent.RunIntoGlazeProcessor")
	defer span.End()

	// Run the agent to extract data
	_, runErr := a.Run(ctx, input)
	if runErr != nil {
		a.tracer.LogEvent(ctx, types.Event{
			Type: "run_error_in_glazed_processor",
			Data: runErr.Error(),
		})
		return errors.Wrap(runErr, "agent run failed")
	}

	// Process the extracted data into the Glaze processor
	for tagName, content := range a.extractedData {
		row := glazed_types.NewRow(
			glazed_types.MRP("tag_name", tagName),
			glazed_types.MRP("content", content),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return errors.Wrapf(err, "failed to add row for tag '%s'", tagName)
		}
	}

	return nil
}

// Ensure StructuredDataAgent implements the relevant agent interfaces
var _ Agent = (*StructuredDataAgent)(nil)       // Base interface
var _ WriterAgent = (*StructuredDataAgent)(nil) // For RunIntoWriter
var _ GlazedAgent = (*StructuredDataAgent)(nil) // For RunIntoGlazeProcessor
