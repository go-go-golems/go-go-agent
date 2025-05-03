package agent

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io" // Added for reading template file if needed (though Glazed might handle it)
	"os"
	"strings"

	// Added for template rendering
	"github.com/go-go-golems/geppetto/pkg/conversation"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
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
	MaxIterations  int    `glazed.parameter:"max-iterations"`
	OutputTemplate string `glazed.parameter:"output-template,string-to-file"` // Template file path
	Render         bool   `glazed.parameter:"render"`
	Raw            bool   `glazed.parameter:"raw"`
}

// NewAgent creates a new StructuredDataAgent.
func (f *StructuredDataAgentFactory) NewAgent(
	ctx context.Context,
	cmd Command,
	parsedLayers *layers.ParsedLayers,
	llmModel llm.LLM,
) (Agent, error) {
	var settings StructuredDataAgentSettings
	// InitializeStruct should handle loading the file content for ParameterTypeStringToFile
	// It will primarily use settings defined in the layer.
	err := parsedLayers.InitializeStruct(StructuredDataAgentType, &settings)
	if err != nil {
		return nil, err
	}

	agentOptions, err := cmd.RenderAgentOptions(parsedLayers.GetDataMap(), nil)
	if err != nil {
		return nil, err
	}

	if maxIter, ok := agentOptions["max-iterations"].(int); ok {
		settings.MaxIterations = maxIter
	}
	if template, ok := agentOptions["template"].(string); ok {
		settings.OutputTemplate = template
	}

	// The settings.Template field now contains the content of the file if loaded
	templateContent := settings.OutputTemplate

	// Default system prompt - this is what buildSystemPrompt used to return
	systemPrompt := `You are an AI assistant that analyzes input and responds with structured data in XML format.`

	// If a system prompt is explicitly set on the command, use it.
	if cmd.GetSystemPrompt() != "" {
		systemPrompt = cmd.GetSystemPrompt()
	}

	// Create the agent, passing the potentially overridden system prompt
	return NewStructuredDataAgent(llmModel, &settings, templateContent, systemPrompt), nil
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
				"output-template",
				parameters.ParameterTypeStringFromFile, // Use StringFromFile for template input
				parameters.WithHelp("Path to the Go template file for rendering the extracted XML data"),
				parameters.WithRequired(false), // Template is optional
			),
			parameters.NewParameterDefinition(
				"render",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Whether to render the output using the template"),
				parameters.WithDefault(true),
			),
			parameters.NewParameterDefinition(
				"raw",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Whether to return the raw LLM response"),
				parameters.WithDefault(false),
			),
		),
	)
	if err != nil {
		return nil, err
	}
	return []layers.ParameterLayer{agentLayer}, nil
}

// Node represents a generic XML node for parsing.
type Node struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Content  []byte     `xml:",innerxml"` // Capture raw inner XML
	Children []Node     `xml:",any"`      // Capture children nodes
}

// StructuredDataAgent implements an agent that extracts structured XML data
// from an LLM response and optionally renders it using a template.
type StructuredDataAgent struct {
	*BaseAgent
	extractedData   map[string]interface{}       // Holds the parsed XML data (likely map[string]interface{})
	settings        *StructuredDataAgentSettings // Agent settings
	templateContent string                       // Content of the template file
	systemPrompt    string                       // System prompt to use
}

// NewStructuredDataAgent creates a new StructuredDataAgent.
func NewStructuredDataAgent(llmModel llm.LLM, settings *StructuredDataAgentSettings, templateContent string, systemPrompt string) *StructuredDataAgent {
	return &StructuredDataAgent{
		BaseAgent:       NewBaseAgent(llmModel, settings.MaxIterations),
		extractedData:   nil, // Initialize as nil
		settings:        settings,
		templateContent: templateContent,
		systemPrompt:    systemPrompt,
	}
}

// GetExtractedData returns the parsed XML data collected during the last Run.
func (a *StructuredDataAgent) GetExtractedData() interface{} {
	return a.extractedData
}

// buildSystemPrompt builds the system prompt for the StructuredDataAgent.
// It instructs the LLM to return well-formed XML.
func (a *StructuredDataAgent) buildSystemPrompt() string {
	// Return the stored system prompt
	return a.systemPrompt
}

// buildSummary generates a string summarizing the outcome.
func (a *StructuredDataAgent) buildSummary() string {
	if a.extractedData == nil {
		return "No structured data was extracted or parsed."
	}

	summary := "Successfully parsed XML data."

	// Try to add more details about the parsed structure
	if len(a.extractedData) > 0 {
		summary += fmt.Sprintf(" Found %d top-level elements: ", len(a.extractedData))
		keys := make([]string, 0, len(a.extractedData))
		for k := range a.extractedData {
			keys = append(keys, k)
		}
		summary += strings.Join(keys, ", ")
	}

	if a.templateContent != "" {
		summary += " Template will be used for rendering."
	}

	return summary
}

// ConvertNodesToMap converts []Node into map[string]interface{} suitable for templates.
// It handles nested structures and attributes.
func ConvertNodesToMap(nodes []Node) map[string]interface{} {
	rootMap := make(map[string]interface{})
	for _, node := range nodes {
		nodeMap := make(map[string]interface{})
		// Add attributes, prefixing with '@'
		if len(node.Attrs) > 0 {
			attrsMap := make(map[string]string)
			for _, attr := range node.Attrs {
				// Prefix attribute names to avoid collision with element names
				attrsMap["@"+attr.Name.Local] = attr.Value
			}
			nodeMap["_attrs"] = attrsMap
		}

		// Process children nodes if present
		if len(node.Children) > 0 {
			// Recursively convert children nodes
			nodeMap["_children"] = ConvertNodesToMap(node.Children)
		}

		// Process inner content if present and no children were found
		// This helps with mixed content or text-only nodes
		innerXML := strings.TrimSpace(string(node.Content))
		if len(innerXML) > 0 {
			// If we have children nodes, the content is likely mixed
			// Check if inner content looks like XML that wasn't parsed as children
			if len(node.Children) == 0 && strings.HasPrefix(innerXML, "<") && strings.HasSuffix(innerXML, ">") {
				// If inner content looks like XML but wasn't parsed as children nodes,
				// attempt to parse it manually as a fallback
				innerNodes := []Node{}
				// Use a temporary reader and decoder for the inner content
				tempDecoder := xml.NewDecoder(bytes.NewReader([]byte(innerXML)))
				tempDecoder.Strict = false // Maintain relaxed parsing
				tempDecoder.AutoClose = xml.HTMLAutoClose
				tempDecoder.Entity = xml.HTMLEntity
				for {
					var n Node
					err := tempDecoder.Decode(&n)
					if err == io.EOF {
						break
					}
					// Ignore errors in recursive parsing for simplicity, log if needed
					if err == nil {
						innerNodes = append(innerNodes, n)
					} else {
						// If decoding fails, treat the whole inner content as text
						nodeMap["_text"] = innerXML
						innerNodes = nil // Ensure we don't process partial nodes
						break
					}
				}
				if len(innerNodes) > 0 {
					// Recursively convert fallback parsed nodes
					nodeMap["_fallback_nodes"] = ConvertNodesToMap(innerNodes)
				} else if _, textExists := nodeMap["_text"]; !textExists {
					// If parsing resulted in no nodes but we didn't already store as text, store the raw innerXML
					nodeMap["_text"] = innerXML
				}
			} else {
				// If inner content doesn't look like XML or we already have children nodes,
				// store as text, but only if we don't have children nodes
				// (to avoid duplicate content)
				if len(node.Children) == 0 {
					nodeMap["_text"] = innerXML
				}
			}
		}

		// Add this nodeMap to the rootMap, handling potential duplicate keys by creating slices
		key := node.XMLName.Local
		if existing, ok := rootMap[key]; ok {
			// Key exists, convert to/append to slice
			if slice, isSlice := existing.([]interface{}); isSlice {
				rootMap[key] = append(slice, nodeMap)
			} else {
				rootMap[key] = []interface{}{existing, nodeMap}
			}
		} else {
			// Key doesn't exist, add as single item
			rootMap[key] = nodeMap
		}
	}
	return rootMap
}

// extractXMLData uses xml.Decoder with relaxed settings to parse XML into []Node.
func extractXMLData(xmlString string) ([]Node, error) {
	// Trim leading/trailing whitespace that might interfere with parsing
	xmlString = strings.TrimSpace(xmlString)
	if xmlString == "" {
		return []Node{}, nil // Return empty if nothing to parse
	}

	// Always wrap with a root element for consistent parsing
	// Skip wrapping if already an XML declaration or already has a root wrapper
	if !strings.HasPrefix(xmlString, "<?xml") && !strings.HasPrefix(xmlString, "<root>") {
		xmlString = "<root>" + xmlString + "</root>"
	}

	r := bytes.NewReader([]byte(xmlString))
	dec := xml.NewDecoder(r)
	dec.Strict = false
	dec.AutoClose = xml.HTMLAutoClose
	dec.Entity = xml.HTMLEntity

	var n Node
	err := dec.Decode(&n)
	if err != nil && err != io.EOF {
		return nil, errors.Wrap(err, "failed to parse XML data")
	}

	// If we successfully parsed the wrapped XML, return its children
	if n.XMLName.Local == "root" {
		return n.Children, nil
	}

	// If for some reason we didn't get a root element (e.g., the XML already had a single root),
	// return the parsed node as a single-element slice
	return []Node{n}, nil
}

// XXX How can we chain validators. custom parser error. Or maybe just as part of the result type, so that many agents can share data.
// something that can render to different "types" (for examples a map or a set of files or a string or an image, etc...) Including the raw LLM inference.
// XXX need to know why the stream finished
// XXX caching configurable

// Run executes the core structured data extraction logic.
func (a *StructuredDataAgent) runInternal(ctx context.Context, input string) (map[string]interface{}, error) {
	ctx, span := a.tracer.StartSpan(ctx, "StructuredDataAgent.Run")
	defer span.End()

	// Reset the extracted data for this run
	a.extractedData = nil

	// Initialize conversation
	messages := []*conversation.Message{
		conversation.NewChatMessage(conversation.RoleSystem, a.buildSystemPrompt()),
		conversation.NewChatMessage(conversation.RoleUser, input),
	}

	// Generate response from LLM
	responseMsg, err := a.llm.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	response := responseMsg.Content.String()

	if a.settings.Raw {
		fmt.Println("response", response)
		return nil, nil
	}

	// Extract structured data from the response using the new XML parser
	nodes, err := extractXMLData(response)
	if err != nil && len(nodes) == 0 { // Only return error if parsing totally failed
		a.tracer.LogEvent(ctx, types.Event{
			Type: "data_extraction_error",
			Data: map[string]interface{}{
				"error":    err.Error(),
				"response": response, // Log the problematic response
			},
		})
		return nil, errors.Wrap(err, "failed to parse structured data from LLM response")
	}
	// Log non-fatal parsing errors if nodes were still extracted
	if err != nil && len(nodes) > 0 {
		a.tracer.LogEvent(ctx, types.Event{
			Type: "data_parsing_warning",
			Data: map[string]interface{}{
				"warning":  err.Error(),
				"response": response,
			},
		})
	}

	// Convert parsed nodes to map for easier template usage
	a.extractedData = ConvertNodesToMap(nodes)

	// Log the outcome
	parseSuccess := a.extractedData != nil
	a.tracer.LogEvent(ctx, types.Event{
		Type: "data_extraction_completed",
		Data: map[string]interface{}{
			"parseSuccess": parseSuccess,
			// Add more details? e.g., number of top-level keys
		},
		Timestamp: 0, // Will be set by the tracer
	})

	return a.extractedData, nil
}

func (a *StructuredDataAgent) Run(ctx context.Context, input string) (string, error) {
	_, err := a.runInternal(ctx, input)
	if err != nil {
		return "", err
	}
	return a.buildSummary(), nil
}

// RunIntoWriter executes the agent, parses the XML, and renders it using the template (if provided) into the writer.
func (a *StructuredDataAgent) RunIntoWriter(ctx context.Context, input string, w io.Writer) error {
	ctx, span := a.tracer.StartSpan(ctx, "StructuredDataAgent.RunIntoWriter")
	defer span.End()

	// Run the agent to parse data. The summary is ignored here.
	_, runErr := a.Run(ctx, input)
	if runErr != nil {
		// Log the error from Run
		a.tracer.LogEvent(ctx, types.Event{
			Type: "run_error_before_writing",
			Data: runErr.Error(),
		})
		// Write the error message to the writer as a fallback
		_, _ = fmt.Fprintf(w, "Agent Run failed: %v\n", runErr)
		return errors.Wrap(runErr, "agent run failed before writing")
	}

	if a.extractedData == nil {
		return nil
	}

	// Check if data was successfully parsed
	if a.extractedData == nil {
		errMsg := "No data was parsed, cannot render template."
		a.tracer.LogEvent(ctx, types.Event{Type: "render_error", Data: errMsg})
		_, _ = fmt.Fprintln(w, errMsg)
		return errors.New(errMsg)
	}

	// Render with template if provided
	if a.templateContent != "" {
		tmpl := templating.CreateTemplate("structured-data")
		tmpl, err := tmpl.Parse(a.templateContent)
		if err != nil {
			a.tracer.LogEvent(ctx, types.Event{Type: "template_parse_error", Data: err.Error()})
			_, _ = fmt.Fprintf(w, "Failed to parse template: %v\n", err)
			return errors.Wrap(err, "failed to parse template")
		}

		err = tmpl.Execute(w, a.extractedData)
		if err != nil {
			a.tracer.LogEvent(ctx, types.Event{Type: "template_execute_error", Data: err.Error()})
			// Attempt to write the error to the writer itself
			_, _ = fmt.Fprintf(w, "\nTEMPLATE EXECUTION ERROR: %v\n", err)
			return errors.Wrap(err, "failed to execute template")
		}
		// Template execution successful
		return nil
	}

	// No template provided, write the summary string to the writer
	summary := a.buildSummary() // Get the summary again (Run's return value was ignored)
	_, writeErr := fmt.Fprintln(w, summary)
	if writeErr != nil {
		return errors.Wrap(writeErr, "failed to write summary")
	}

	return nil // Return nil as Run succeeded and summary was written
}

// RunIntoGlazeProcessor executes the agent and streams structured data as rows
// into the Glazed processor. This is marked as not fully supported with templates.
func (a *StructuredDataAgent) RunIntoGlazeProcessor(
	ctx context.Context,
	input string,
	gp middlewares.Processor,
) error {
	if a.settings.Render {
		return a.RunIntoWriter(ctx, input, os.Stdout)
	}

	ctx, span := a.tracer.StartSpan(ctx, "StructuredDataAgent.RunIntoGlazeProcessor")
	defer span.End()

	// Run the agent to extract data
	values, runErr := a.runInternal(ctx, input)
	if values == nil {
		return nil
	}
	if runErr != nil {
		a.tracer.LogEvent(ctx, types.Event{
			Type: "run_error_in_glazed_processor",
			Data: runErr.Error(),
		})
		return errors.Wrap(runErr, "agent run failed")
	}

	if len(values) == 1 && values["_children"] != nil {
		values = values["_children"].(map[string]interface{})
	}

	for key, value := range values {
		row := glazed_types.NewRowFromMap(map[string]interface{}{
			key: value,
		})
		if err := gp.AddRow(ctx, row); err != nil {
			return errors.Wrap(err, "failed to add parsed data row")
		}
	}

	return nil
}

// Ensure StructuredDataAgent implements the relevant agent interfaces
var _ Agent = (*StructuredDataAgent)(nil)       // Base interface
var _ WriterAgent = (*StructuredDataAgent)(nil) // For RunIntoWriter
var _ GlazedAgent = (*StructuredDataAgent)(nil) // For RunIntoGlazeProcessor
