# Go-Go-Agent TUI Event Viewer

The Terminal UI (TUI) Event Viewer is a command-line application for monitoring and inspecting events from the Go-Go-Agent system. It provides real-time event streaming from Redis as well as historical event viewing from a SQLite database.

## Features

- Real-time event streaming and display
- Loading historical events from a SQLite database
- Filtering events by run ID, time range, or count
- Interactive filtering by event type and free-text search
- Customizable event rendering with type-specific views
- Interactive navigation and event inspection
- Expandable/collapsible detail sections
- Auto-scrolling for live monitoring

## Installation

The viewer is built as part of the Go-Go-Agent system:

```bash
make build
```

This will compile the TUI viewer along with other components of the system.

## Usage

### Starting the Viewer

To start the viewer with default settings (streaming events from Redis):

```bash
./go-go-agent tui-viewer
```

### Command-line Options

The TUI viewer supports various command-line options:

```
Options:
  --max-events int         Maximum number of events to keep in memory (0 for unlimited) (default 1000)
  --db-path string         Path to SQLite database file (for loading events) (default "./writehere.db")
  --use-db                 Load events from database instead of streaming
  --run-id string          Load events for a specific run ID
  --latest-run             Load events from the latest run
  --limit int              Maximum number of events to load from database (default 100)
  --start-time string      Start time for loading events (ISO format)
  --end-time string        End time for loading events (ISO format)
  --streaming-enabled      Enable streaming of new events (even when loading from DB) (default true)
```

### Example Commands

**Streaming live events**:
```bash
./go-go-agent tui-viewer
```

**Loading events from the latest run in the database**:
```bash
./go-go-agent tui-viewer --use-db --latest-run
```

**Loading events for a specific run ID**:
```bash
./go-go-agent tui-viewer --use-db --run-id="run_1234567890"
```

**Loading the last 500 events from the database**:
```bash
./go-go-agent tui-viewer --use-db --limit=500
```

**Loading events from a time range**:
```bash
./go-go-agent tui-viewer --use-db --start-time="2025-05-01T00:00:00Z" --end-time="2025-05-02T00:00:00Z"
```

### Interactive Controls

Once the viewer is running, you can interact with it using these keyboard commands:

- `Enter`: Select an event to view its details
- `Esc`: Return from event details view to event list
- `a`: Toggle auto-scrolling (follows new events when enabled)
- `f`: Open event filter view
- `F`: Apply current filters and return to event list
- `e`: Expand/collapse sections in detail view
- `?`: Toggle help display
- `q` or `Ctrl+C`: Quit the application

## Architecture

### Component Overview

The TUI viewer is built using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework, which provides a reactive terminal UI paradigm. The application consists of these main components:

1. **Command & Settings** (`TUIViewerCmd`, `TUIViewerSettings`): Handles command-line arguments and configuration
2. **Event Loading** (`db.GetLatestEvents`, `db.GetRunEvents`, etc.): Loads events from the SQLite database
3. **Event Streaming** (`setupRouter`): Configures Redis for streaming new events
4. **UI Model** (`ViewModel`): Manages the application's state and UI components
5. **Registry System** (`views.Registry`): Provides extensible view rendering for different event types
6. **Filter System** (`filters.FilterView`): Implements interactive event filtering
7. **Event Display** (`EventView` implementations): Type-specific rendering of events

### Data Flow

1. Events are either loaded from the database (`internal/db/events.go`, `internal/db/manager.go`) or streamed from Redis
2. Events are sent to an event channel (`eventChan`)
3. The TUI model (`ViewModel`) processes these events and updates the UI
4. If filters are active, events pass through the filter system before being displayed
5. Events are rendered using type-specific views from the registry
6. User interactions trigger state changes in the model, which can toggle between list view, detail view, and filter view

### Key Types

#### `TUIViewerCmd`
Main command type that implements the `cmds.BareCommand` interface. Responsible for setting up the application based on command-line arguments.

#### `TUIViewerSettings`
Stores the configuration options parsed from command-line flags:
- `MaxEvents`: Maximum events to keep in memory
- `DBPath`: Path to the SQLite database
- `UseDB`: Whether to load events from the database
- `RunID`, `LatestRun`, `Limit`, `StartTime`, `EndTime`: Database query parameters
- `StreamingEnabled`: Whether to stream new events

#### `ViewModel`
The main UI model that implements the Bubble Tea `tea.Model` interface. Contains:
- `List`: List of events (using `list.Model`)
- `Viewport`: Detail view for a selected event (using `viewport.Model`)
- `Help`: Help display (using `help.Model`)
- `Events`: Slice of loaded events
- `FilteredEvents`: Events after filtering has been applied
- `ShouldAutoScroll`: Auto-scroll state
- `ShowingDetail`: Whether the detail view is active
- `ShowingFilter`: Whether the filter view is active
- `FilterView`: The filter UI component
- `ViewRegistry`: Registry of custom event type formatters
- `Expanded`: Tracking which sections are expanded in detail view

#### `EventItem`
Adaptor type that wraps `model.Event` for display in the list. Implements the `list.Item` interface and
uses the view registry to format events for display.

#### `FilterView`
Implements the filter UI component with:
- Checkboxes for each event type
- Category toggles for groups of related event types
- Free-text search input
- State tracking for expanded/collapsed sections

#### `FilterState`
Holds the current filter configuration:
- `EventTypeFilters`: Map of event type to whether it's enabled
- `SearchTerm`: Free-text search term

### Key Functions

#### `(c *TUIViewerCmd) Run()`
Main entry point that sets up the database connection, event loading, and streaming based on provided settings.

#### `convertRawEventsToEvents()`
Converts raw JSON events from the database to `model.Event` structs.

#### `setupRouter()`
Configures and returns a Watermill message router for Redis event streaming.

#### `NewViewModel()`
Creates and initializes a new TUI model with the provided settings, including the view registry and filter view.

#### `(m ViewModel) Init()`, `Update()`, `View()`
Implement the Bubble Tea model interface for UI rendering and state updates. The `Update()` method handles toggling between list, detail, and filter views.

#### `updateListWithEvents()`
Updates the list view with either all events or filtered events.

#### `ApplyFilters()`
Applies the current filter settings to the events list, handling both event type filtering and free-text search.

#### `ViewRegistry` system
The view registry system provides extensible, type-specific rendering for events:
- `FormatEvent()`: Renders detailed view of an event
- `FormatListTitle()`: Formats the title line for an event in the list
- `FormatListDescription()`: Formats the description line for an event in the list

### Event Processing

Events flow through the system as follows:

1. **Loading Events**: When `--use-db` is specified, the appropriate database function is called:
   - `db.GetRunEvents()`: For a specific run ID
   - `db.GetLatestRunEvents()`: For the latest run
   - `db.GetEventsInTimeRange()`: For events within a time range
   - `db.GetLatestEvents()`: For the latest N events

2. **Streaming Events**: A message handler is set up that forwards Redis events to the event channel.

3. **Filtering**: When filters are active, the `ApplyFilters()` method applies both type filters and text search:
   - Events are filtered by type using the `EventTypeFilters` map
   - Text search is applied across event ID, run ID, type, and payload
   - Filtered events are stored in the `FilteredEvents` slice

4. **UI Updates**: The `Update()` method processes events and updates the view:
   - List view shows either all events or filtered events
   - When an event is selected, its custom view renders detailed information
   - Expandable sections can be toggled for better readability

## Customization

### Adding New Event Types

To add support for a new event type:

1. Define the new event type constant in `pkg/model/event.go`
2. Add a corresponding struct for the event payload
3. Create a custom view for the event type by implementing the `EventView` interface
4. Register the view in the registry via `viewRegistry.RegisterView()`
5. The filter system will automatically include the new event type

### Extending the UI

The viewer uses the Bubble Tea framework, which makes it relatively straightforward to extend the UI:

1. Add new key bindings in the `KeyMap` struct
2. Add handling for the new keys in the `Update()` method
3. Update the UI rendering in the `View()` method

### Adding Custom Event Views

To implement a custom view for a specific event type:

1. Create a new type that implements the `EventView` interface in the views package:
   - `Render(event, expanded)`: Render the event details for the detailed view
   - `FormatListTitle(event)`: Format the event title for the list view
   - `FormatListDescription(event)`: Format the event description for the list view
   - `ExpandableFields()`: Return which fields can be expanded/collapsed

2. Register your custom view in the registry when initializing the model:
   ```go
   viewRegistry.RegisterView(model.EventTypeYourEvent, &views.YourEventView{})
   ```

3. The view registry will automatically use your view when rendering events of the registered type

### Customizing Filters

The filter system can be extended to add new categories or filter groups:

1. Modify the `filters.FilterView` to include additional category toggles
2. Update the `updateCategoryFilters()` and `updateCategoryTogglesFromIndividuals()` methods
3. Add logic for new filter categories in the toggle handler

## Troubleshooting

### Event Display Issues

If events don't display correctly, the viewer shows the raw payload to help diagnose the issue. This is particularly useful for complex event types or when the payload structure changes.

### Connection Issues

If you encounter issues connecting to Redis or the database:

1. Verify the connection parameters in your configuration
2. Check that the services are running and accessible
3. Review the log output for specific error messages

## Related Components

The TUI viewer works closely with these other components:

- `internal/db/manager.go`: Database management and event storage
- `internal/db/events.go`: Event retrieval functions
- `internal/redis/router.go`: Redis connectivity and message routing
- `pkg/model/event.go`: Event type definitions
- `pkg/eventbus/bus.go`: Event bus for publishing events
- `cmd/tui-viewer/tui/views/registry.go`: Event rendering registry system
- `cmd/tui-viewer/tui/views/llm_views.go`: Custom views for LLM events
- `cmd/tui-viewer/tui/filters/filter_view.go`: Event filtering system