package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-agent/cmd/tui-viewer/tui/filters"
	"github.com/go-go-golems/go-go-agent/cmd/tui-viewer/tui/views"
	"github.com/go-go-golems/go-go-agent/pkg/model"
)

// Event item for the list
type EventItem struct {
	ID        string
	Timestamp string
	EventType string
	RunID     string
	Event     model.Event
	registry  *views.Registry // View registry reference
}

func (i EventItem) Title() string {
	// Use the registry to format the title if available
	if i.registry != nil {
		return i.registry.FormatListTitle(i.Event)
	}
	return i.EventType
}

func (i EventItem) Description() string {
	// Use the registry to format the description if available
	if i.registry != nil {
		return i.registry.FormatListDescription(i.Event)
	}
	return fmt.Sprintf("ID: %s | Time: %s | Run: %s",
		i.ID, i.Timestamp, i.RunID)
}
func (i EventItem) FilterValue() string { return i.EventType }

// Key mappings
type KeyMap struct {
	ToggleAutoScroll key.Binding
	Help             key.Binding
	Quit             key.Binding
	Back             key.Binding
	Select           key.Binding
	ToggleExpand     key.Binding
	ToggleFilter     key.Binding
	ApplyFilter      key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.ToggleAutoScroll, k.ToggleFilter, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ToggleAutoScroll, k.Help, k.Quit},
		{k.Select, k.Back, k.ToggleExpand},
		{k.ToggleFilter, k.ApplyFilter},
	}
}

var Keys = KeyMap{
	ToggleAutoScroll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "toggle auto-scroll"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	ToggleExpand: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "expand/collapse sections"),
	),
	ToggleFilter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "filter events"),
	),
	ApplyFilter: key.NewBinding(
		key.WithKeys("F"),
		key.WithHelp("F", "apply filter"),
	),
}

// Messages
type EventMsg model.Event

// Main model
type ViewModel struct {
	List             list.Model
	Viewport         viewport.Model
	Help             help.Model
	MaxEvents        int
	EventChan        chan model.Event
	Events           []model.Event
	FilteredEvents   []model.Event    // Events after filtering
	ShouldAutoScroll bool
	ShowingDetail    bool
	ShowingFilter    bool             // Whether we're showing the filter UI
	CurrentEvent     *model.Event
	ViewRegistry     *views.Registry
	Expanded         map[string]bool  // Track expanded sections
	FilterView       filters.FilterView
}

func NewViewModel(maxEvents int, eventChan chan model.Event) ViewModel {
	// Setup list
	listDelegate := list.NewDefaultDelegate()
	listItems := []list.Item{}

	listModel := list.New(listItems, listDelegate, 0, 0)
	listModel.Title = "Event Timeline"
	listModel.SetShowStatusBar(true)
	listModel.SetFilteringEnabled(false)
	listModel.SetShowHelp(false)

	// Setup viewport for details
	vp := viewport.New(0, 0)

	// Help model
	hm := help.New()
	hm.ShowAll = false

	// Create and configure the view registry
	viewRegistry := views.NewRegistry()

	// Register custom views
	viewRegistry.RegisterView(model.EventTypeLLMCallStarted, &views.LLMCallStartedView{})
	viewRegistry.RegisterView(model.EventTypeLLMCallCompleted, &views.LLMCallCompletedView{})

	// Create filter view
	filterView := filters.NewFilterView()

	return ViewModel{
		List:             listModel,
		Viewport:         vp,
		Help:             hm,
		MaxEvents:        maxEvents,
		EventChan:        eventChan,
		Events:           []model.Event{},
		FilteredEvents:   []model.Event{},
		ShouldAutoScroll: true,
		ShowingDetail:    false,
		ShowingFilter:    false,
		CurrentEvent:     nil,
		ViewRegistry:     viewRegistry,
		Expanded:         make(map[string]bool),
		FilterView:       filterView,
	}
}

func CheckForEvents(eventChan chan model.Event) tea.Cmd {
	return func() tea.Msg {
		select {
		case evt, ok := <-eventChan:
			if !ok {
				return nil // Channel closed
			}
			return EventMsg(evt)
		default:
			// Non-blocking
			return nil
		}
	}
}

func (m ViewModel) Init() tea.Cmd {
	return CheckForEvents(m.EventChan)
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	eventHandled := false

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		verticalMargin := 4 // For header and footer
		horizontalMargin := 2

		m.List.SetWidth(msg.Width - horizontalMargin)
		m.List.SetHeight(msg.Height - verticalMargin)
		m.Viewport.Width = msg.Width - horizontalMargin
		m.Viewport.Height = msg.Height - verticalMargin
		m.FilterView.Width = msg.Width - horizontalMargin
		m.FilterView.Height = msg.Height - verticalMargin

		m.Help.Width = msg.Width

	case EventMsg:
		// Add the event to our list
		event := model.Event(msg)
		m.Events = append(m.Events, event)

		// Truncate events if exceeding max
		if m.MaxEvents > 0 && len(m.Events) > m.MaxEvents {
			m.Events = m.Events[len(m.Events)-m.MaxEvents:]
		}

		// Check if we need to apply filters
		if len(m.FilteredEvents) > 0 {
			// Re-apply filters to get updated filtered list
			m.FilteredEvents = m.FilterView.ApplyFilters(m.Events)
			updateListWithEvents(&m, m.FilteredEvents)
		} else {
			// No filters, show all events
			updateListWithEvents(&m, m.Events)
		}

		// Auto-scroll if enabled
		if m.ShouldAutoScroll && !m.ShowingDetail && !m.ShowingFilter {
			m.List.Select(len(m.List.Items()) - 1)
		}

		cmds = append(cmds, CheckForEvents(m.EventChan))

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Quit):
			log.Info().Msg("Quitting")
			return m, tea.Quit

		case key.Matches(msg, Keys.ToggleAutoScroll):
			m.ShouldAutoScroll = !m.ShouldAutoScroll
			if m.ShouldAutoScroll && !m.ShowingDetail && !m.ShowingFilter {
				m.List.Select(len(m.Events) - 1)
			}

		case key.Matches(msg, Keys.Help):
			m.Help.ShowAll = !m.Help.ShowAll

		case key.Matches(msg, Keys.ToggleFilter):
			if !m.ShowingDetail {
				// Toggle filter view
				m.ShowingFilter = !m.ShowingFilter
				eventHandled = true
			}

		case key.Matches(msg, Keys.ApplyFilter):
			if m.ShowingFilter {
				// Apply the current filters
				// Get filter state and apply it
				m.FilteredEvents = m.FilterView.ApplyFilters(m.Events)
				
				// Update the list view with filtered events
				updateListWithEvents(&m, m.FilteredEvents)
				
				// Exit filter view
				m.ShowingFilter = false
				eventHandled = true
			}

		case key.Matches(msg, Keys.Back):
			if m.ShowingDetail {
				m.ShowingDetail = false
				m.CurrentEvent = nil
				m.Expanded = make(map[string]bool) // Reset expanded sections
				eventHandled = true
			} else if m.ShowingFilter {
				m.ShowingFilter = false
				eventHandled = true
			}

		case key.Matches(msg, Keys.Select):
			if !m.ShowingDetail {
				selected := m.List.SelectedItem()
				if selected != nil {
					item := selected.(EventItem)
					m.ShowingDetail = true
					m.CurrentEvent = &item.Event

					// Format the event details using the registry
					detailContent, err := m.ViewRegistry.FormatEvent(*m.CurrentEvent, m.Expanded)
					if err == nil {
						m.Viewport.SetContent(detailContent)
					} else {
						m.Viewport.SetContent(fmt.Sprintf("Error formatting event: %v", err))
					}
					m.Viewport.GotoTop()
				}
			}
			eventHandled = true

		case key.Matches(msg, Keys.ToggleExpand):
			if m.ShowingDetail && m.CurrentEvent != nil {
				// Toggle expand based on event type and current focus
				view := m.ViewRegistry.GetView(m.CurrentEvent.EventType)
				expandable := view.ExpandableFields()

				// If we have expandable fields, toggle them
				if len(expandable) > 0 {
					// For simplicity, toggle all fields
					for _, field := range expandable {
						m.Expanded[field] = !m.Expanded[field]
					}

					// Redraw the content
					detailContent, err := m.ViewRegistry.FormatEvent(*m.CurrentEvent, m.Expanded)
					if err == nil {
						m.Viewport.SetContent(detailContent)
					}
				}
			}
			eventHandled = true
		}
	}

	if !eventHandled {
		if m.ShowingFilter {
			// Update filter view
			var cmd tea.Cmd
			m.FilterView, cmd = m.FilterView.Update(msg)
			cmds = append(cmds, cmd)
		} else if m.ShowingDetail {
			// Update detail view
			var cmd tea.Cmd
			m.Viewport, cmd = m.Viewport.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			// Update list view
			var cmd tea.Cmd
			m.List, cmd = m.List.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Always check for new events
	cmds = append(cmds, CheckForEvents(m.EventChan))

	return m, tea.Batch(cmds...)
}

// Helper function to update the list with events (filtered or unfiltered)
func updateListWithEvents(m *ViewModel, events []model.Event) {
	listItems := make([]list.Item, len(events))
	for i, e := range events {
		listItems[i] = EventItem{
			ID:        e.EventID,
			Timestamp: e.Timestamp,
			EventType: e.EventType,
			RunID:     e.RunID,
			Event:     e,
			registry:  m.ViewRegistry,
		}
	}
	m.List.SetItems(listItems)
}

func (m ViewModel) View() string {
	// If in filter mode, show the filter view
	if m.ShowingFilter {
		return m.FilterView.View()
	}

	// If in detail mode, show the detail view
	if m.ShowingDetail {
		// Check if the current event has expandable fields
		expandableHelp := ""
		if m.CurrentEvent != nil {
			view := m.ViewRegistry.GetView(m.CurrentEvent.EventType)
			if len(view.ExpandableFields()) > 0 {
				expandableHelp = "\nPress 'e' to expand/collapse sections"
			}
		}

		s := fmt.Sprintf("%s\n%s\n%s%s",
			lipgloss.NewStyle().Bold(true).Render("Event Details"),
			m.Viewport.View(),
			m.Help.View(Keys),
			expandableHelp,
		)
		return s
	}

	s := m.List.View()

	// Add status line showing auto-scroll status
	autoScrollStatus := "Auto-scroll: "
	if m.ShouldAutoScroll {
		autoScrollStatus += lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("ON")
	} else {
		autoScrollStatus += lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("OFF")
	}

	s += "\n" + m.Help.View(Keys) + "\n" + autoScrollStatus

	return s
}
