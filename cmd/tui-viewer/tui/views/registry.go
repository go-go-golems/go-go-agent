package views

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-agent/pkg/model"
)

// EventView defines the interface for custom event views
type EventView interface {
	// Render formats an event for display
	Render(event model.Event, expanded map[string]bool) (string, error)
	// ExpandableFields returns a list of field names that can be expanded/collapsed
	ExpandableFields() []string
}

// Default view implements EventView for generic events
type DefaultView struct{}

func (d *DefaultView) Render(event model.Event, expanded map[string]bool) (string, error) {
	sb := &strings.Builder{}

	fmt.Fprintf(sb, "Event ID: %s\n", event.EventID)
	fmt.Fprintf(sb, "Timestamp: %s\n", event.Timestamp)
	fmt.Fprintf(sb, "Event Type: %s\n", event.EventType)
	fmt.Fprintf(sb, "Run ID: %s\n\n", event.RunID)
	fmt.Fprintf(sb, "--- Payload ---\n")

	// Pretty-print the payload
	var payload interface{}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return "", err
	}

	// Convert to pretty JSON
	prettyJSON, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}

	fmt.Fprintf(sb, "%s", string(prettyJSON))

	return sb.String(), nil
}

func (d *DefaultView) ExpandableFields() []string {
	return []string{} // No expandable fields in default view
}

// Registry maintains a mapping from event types to their custom views
type Registry struct {
	views       map[string]EventView
	defaultView EventView
}

// NewRegistry creates a new view registry with the default view
func NewRegistry() *Registry {
	return &Registry{
		views:       make(map[string]EventView),
		defaultView: &DefaultView{},
	}
}

// RegisterView adds a custom view for a specific event type
func (r *Registry) RegisterView(eventType string, view EventView) {
	r.views[eventType] = view
}

// GetView returns the appropriate view for an event type
func (r *Registry) GetView(eventType string) EventView {
	view, exists := r.views[eventType]
	if !exists {
		return r.defaultView
	}
	return view
}

// FormatEvent renders an event using the appropriate view
func (r *Registry) FormatEvent(event model.Event, expanded map[string]bool) (string, error) {
	view := r.GetView(event.EventType)
	return view.Render(event, expanded)
}
