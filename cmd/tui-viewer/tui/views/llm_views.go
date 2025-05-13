package views

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-agent/pkg/model"
	"gopkg.in/yaml.v3"
)

// Styling
var (
	fieldStyle     = lipgloss.NewStyle().Bold(true)
	headerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	expandedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	collapsedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// LLMCallStartedView implements EventView for LLM call started events
type LLMCallStartedView struct{}

func (v *LLMCallStartedView) ExpandableFields() []string {
	return []string{"prompt"}
}

func (v *LLMCallStartedView) Render(event model.Event, expanded map[string]bool) (string, error) {
	sb := &strings.Builder{}

	// Parse the payload
	var payload model.LLMCallStartedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return "", err
	}

	// Write basic information
	fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Event ID"), event.EventID)
	fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Timestamp"), event.Timestamp)
	fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Run ID"), event.RunID)
	fmt.Fprintf(sb, "\n")

	// LLM specific information
	fmt.Fprintf(sb, "%s\n", headerStyle.Render("LLM Call Details"))
	fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Agent Class"), payload.AgentClass)
	fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Model"), payload.Model)
	if payload.NodeID != "" {
		fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Node ID"), payload.NodeID)
	}
	if payload.ActionName != "" {
		fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Action"), payload.ActionName)
	}
	if payload.TaskType != "" {
		fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Task Type"), payload.TaskType)
	}
	if payload.Step > 0 {
		fmt.Fprintf(sb, "%s: %d\n", fieldStyle.Render("Step"), payload.Step)
	}

	// Format the prompt
	isExpanded := expanded["prompt"]

	// Always show preview
	if payload.PromptPreview != "" {
		fmt.Fprintf(sb, "\n%s:\n%s\n", fieldStyle.Render("Prompt Preview"), payload.PromptPreview)
	}

	// Show full prompt if expanded
	if len(payload.Prompt) > 0 {
		expandHeader := "[+] Full Prompt (press 'e' to expand)"
		if isExpanded {
			expandHeader = "[-] Full Prompt (press 'e' to collapse)"

			// Parse the prompt - could be different formats
			var promptData interface{}
			err := json.Unmarshal(payload.Prompt, &promptData)
			if err != nil {
				fmt.Fprintf(sb, "\n%s\nError parsing prompt: %v\n",
					expandedStyle.Render(expandHeader), err)
			} else {
				// Convert to YAML for better readability
				yamlData, err := yaml.Marshal(promptData)
				if err != nil {
					// Fall back to JSON if YAML conversion fails
					jsonData, _ := json.MarshalIndent(promptData, "", "  ")
					fmt.Fprintf(sb, "\n%s\n%s\n",
						expandedStyle.Render(expandHeader), string(jsonData))
				} else {
					fmt.Fprintf(sb, "\n%s\n%s\n",
						expandedStyle.Render(expandHeader), string(yamlData))
				}
			}
		} else {
			fmt.Fprintf(sb, "\n%s\n", collapsedStyle.Render(expandHeader))
		}
	}

	return sb.String(), nil
}

// LLMCallCompletedView implements EventView for LLM call completed events
type LLMCallCompletedView struct{}

func (v *LLMCallCompletedView) ExpandableFields() []string {
	return []string{"response", "token_usage"}
}

func (v *LLMCallCompletedView) Render(event model.Event, expanded map[string]bool) (string, error) {
	sb := &strings.Builder{}

	// Parse the payload
	var payload model.LLMCallCompletedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return "", err
	}

	// Write basic information
	fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Event ID"), event.EventID)
	fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Timestamp"), event.Timestamp)
	fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Run ID"), event.RunID)
	fmt.Fprintf(sb, "\n")

	// LLM specific information
	fmt.Fprintf(sb, "%s\n", headerStyle.Render("LLM Call Results"))
	fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Agent Class"), payload.AgentClass)
	fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Model"), payload.Model)
	fmt.Fprintf(sb, "%s: %.2f seconds\n", fieldStyle.Render("Duration"), payload.DurationSeconds)

	if payload.NodeID != "" {
		fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Node ID"), payload.NodeID)
	}
	if payload.ActionName != "" {
		fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Action"), payload.ActionName)
	}
	if payload.TaskType != "" {
		fmt.Fprintf(sb, "%s: %s\n", fieldStyle.Render("Task Type"), payload.TaskType)
	}
	if payload.Step > 0 {
		fmt.Fprintf(sb, "%s: %d\n", fieldStyle.Render("Step"), payload.Step)
	}

	// Error handling
	if payload.Error != nil {
		fmt.Fprintf(sb, "\n%s: %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render("ERROR"),
			*payload.Error)
	}

	// Always show summary if available
	if payload.ResultSummary != "" {
		fmt.Fprintf(sb, "\n%s:\n%s\n", fieldStyle.Render("Result Summary"), payload.ResultSummary)
	}

	// Full response (expandable)
	isResponseExpanded := expanded["response"]
	if payload.Response != "" {
		expandHeader := "[+] Full Response (press 'e' to expand)"
		if isResponseExpanded {
			expandHeader = "[-] Full Response (press 'e' to collapse)"
			fmt.Fprintf(sb, "\n%s\n%s\n",
				expandedStyle.Render(expandHeader), payload.Response)
		} else {
			fmt.Fprintf(sb, "\n%s\n", collapsedStyle.Render(expandHeader))
		}
	}

	// Token usage (expandable)
	isTokenUsageExpanded := expanded["token_usage"]
	if len(payload.TokenUsage) > 0 {
		expandHeader := "[+] Token Usage (press 'e' to expand)"
		if isTokenUsageExpanded {
			expandHeader = "[-] Token Usage (press 'e' to collapse)"

			// Parse token usage
			var tokenUsage interface{}
			if err := json.Unmarshal(payload.TokenUsage, &tokenUsage); err != nil {
				fmt.Fprintf(sb, "\n%s\nError parsing token usage: %v\n",
					expandedStyle.Render(expandHeader), err)
			} else {
				// Format as YAML
				yamlData, err := yaml.Marshal(tokenUsage)
				if err != nil {
					// Fall back to JSON
					jsonData, _ := json.MarshalIndent(tokenUsage, "", "  ")
					fmt.Fprintf(sb, "\n%s\n%s\n",
						expandedStyle.Render(expandHeader), string(jsonData))
				} else {
					fmt.Fprintf(sb, "\n%s\n%s\n",
						expandedStyle.Render(expandHeader), string(yamlData))
				}
			}
		} else {
			fmt.Fprintf(sb, "\n%s\n", collapsedStyle.Render(expandHeader))
		}
	}

	return sb.String(), nil
}
