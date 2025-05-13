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

func (v *LLMCallStartedView) FormatListTitle(event model.Event) string {
	var payload model.LLMCallStartedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return event.EventType
	}
	return fmt.Sprintf("📝 LLM Call [%s]", payload.Model)
}

func (v *LLMCallStartedView) FormatListDescription(event model.Event) string {
	var payload model.LLMCallStartedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Sprintf("ID: %s | Time: %s", event.EventID, event.Timestamp)
	}

	// Prepare a summary of the prompt preview
	preview := payload.PromptPreview
	if len(preview) > 60 {
		preview = preview[:57] + "..."
	}

	return fmt.Sprintf("%s | %s",
		payload.AgentClass,
		strings.ReplaceAll(preview, "\n", " "))
}

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

func (v *LLMCallCompletedView) FormatListTitle(event model.Event) string {
	var payload model.LLMCallCompletedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return event.EventType
	}

	// Add an icon to show status (error or success)
	prefix := "💬" // Speech bubble for success
	if payload.Error != nil {
		prefix = "❌" // Red X for error
	}

	// Extract token usage if available
	tokenInfo := ""
	if len(payload.TokenUsage) > 0 {
		var tokenUsage map[string]interface{}
		if err := json.Unmarshal(payload.TokenUsage, &tokenUsage); err == nil {
			// Try to extract common token count fields
			var totalTokens int
			var inputTokens, outputTokens int

			// Handle different token usage formats
			if total, ok := tokenUsage["total_tokens"].(float64); ok {
				totalTokens = int(total)
			}
			if prompt, ok := tokenUsage["prompt_tokens"].(float64); ok {
				inputTokens = int(prompt)
			} else if input, ok := tokenUsage["input_tokens"].(float64); ok {
				inputTokens = int(input)
			}
			if completion, ok := tokenUsage["completion_tokens"].(float64); ok {
				outputTokens = int(completion)
			} else if output, ok := tokenUsage["output_tokens"].(float64); ok {
				outputTokens = int(output)
			}

			// Format token information
			if totalTokens > 0 {
				tokenInfo = fmt.Sprintf(" [%d tokens]", totalTokens)
			} else if inputTokens > 0 || outputTokens > 0 {
				tokenInfo = fmt.Sprintf(" [%d→%d tokens]", inputTokens, outputTokens)
			}
		}
	}

	return fmt.Sprintf("%s LLM Response [%s] %.1fs%s",
		prefix, payload.Model, payload.DurationSeconds, tokenInfo)
}

func (v *LLMCallCompletedView) FormatListDescription(event model.Event) string {
	var payload model.LLMCallCompletedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Sprintf("ID: %s | Time: %s", event.EventID, event.Timestamp)
	}

	// Show error message if present
	if payload.Error != nil {
		return fmt.Sprintf("ERROR: %s", *payload.Error)
	}

	// Show summary or start of response
	if payload.ResultSummary != "" {
		summary := payload.ResultSummary
		if len(summary) > 60 {
			summary = summary[:57] + "..."
		}
		return strings.ReplaceAll(summary, "\n", " ")
	} else if payload.Response != "" {
		resp := payload.Response
		if len(resp) > 60 {
			resp = resp[:57] + "..."
		}
		return strings.ReplaceAll(resp, "\n", " ")
	}

	return fmt.Sprintf("%s | Complete", payload.AgentClass)
}

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
		// First extract and show a token summary
		var tokenUsage map[string]interface{}
		if err := json.Unmarshal(payload.TokenUsage, &tokenUsage); err == nil {
			// Try to extract common token count fields
			var totalTokens int
			var inputTokens, outputTokens int

			// Handle different token usage formats
			if total, ok := tokenUsage["total_tokens"].(float64); ok {
				totalTokens = int(total)
			}
			if prompt, ok := tokenUsage["prompt_tokens"].(float64); ok {
				inputTokens = int(prompt)
			} else if input, ok := tokenUsage["input_tokens"].(float64); ok {
				inputTokens = int(input)
			}
			if completion, ok := tokenUsage["completion_tokens"].(float64); ok {
				outputTokens = int(completion)
			} else if output, ok := tokenUsage["output_tokens"].(float64); ok {
				outputTokens = int(output)
			}

			// Display token summary
			if totalTokens > 0 {
				fmt.Fprintf(sb, "\n%s: %d\n", fieldStyle.Render("Total Tokens"), totalTokens)
			} else if inputTokens > 0 || outputTokens > 0 {
				fmt.Fprintf(sb, "\n%s: %d %s %d = %d\n",
					fieldStyle.Render("Tokens"),
					inputTokens,
					lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render("→"),
					outputTokens,
					inputTokens+outputTokens)
			}
		}

		expandHeader := "[+] Token Usage Details (press 'e' to expand)"
		if isTokenUsageExpanded {
			expandHeader = "[-] Token Usage Details (press 'e' to collapse)"

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
