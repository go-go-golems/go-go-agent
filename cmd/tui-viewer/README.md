# Go-Go-Agent TUI Event Viewer

The Terminal UI (TUI) Event Viewer is a command-line application for monitoring and inspecting events from the Go-Go-Agent system. It provides real-time event streaming from Redis as well as historical event viewing from a SQLite database.

## Features

- Real-time event streaming and display
- Loading historical events from a SQLite database
- Filtering events by run ID, time range, or count
- Interactive navigation and event inspection
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
- `?`: Toggle help display
- `q` or `Ctrl+C`: Quit the application

## Architecture

### Component Overview

The TUI viewer is built using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework, which provides a reactive terminal UI paradigm. The application consists of these main components:

1. **Command & Settings** (`TUIViewerCmd`, `TUIViewerSettings`): Handles command-line arguments and configuration
2. **Event Loading** (`db.GetLatestEvents`, `db.GetRunEvents`, etc.): Loads events from the SQLite database
3. **Event Streaming** (`setupRouter`): Configures Redis for streaming new events
4. **UI Model** (`viewModel`): Manages the application's state and UI components
5. **Event Display** (`formatEventDetail`): Renders events for display

### Data Flow

1. Events are either loaded from the database (`internal/db/events.go`, `internal/db/manager.go`) or streamed from Redis
2. Events are sent to an event channel (`eventChan`)
3. The TUI model (`viewModel`) processes these events and updates the UI
4. User interactions trigger state changes in the model

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

#### `viewModel`
The main UI model that implements the Bubble Tea `tea.Model` interface. Contains:
- `list`: List of events (using `list.Model`)
- `viewport`: Detail view for a selected event (using `viewport.Model`)
- `help`: Help display (using `help.Model`)
- `events`: Slice of loaded events
- `shouldAutoScroll`: Auto-scroll state
- `showingDetail`: Whether the detail view is active

#### `eventItem`
Adaptor type that wraps `model.Event` for display in the list. Implements the `list.Item` interface.

### Key Functions

#### `(c *TUIViewerCmd) Run()`
Main entry point that sets up the database connection, event loading, and streaming based on provided settings.

#### `convertRawEventsToEvents()`
Converts raw JSON events from the database to `model.Event` structs.

#### `setupRouter()`
Configures and returns a Watermill message router for Redis event streaming.

#### `newModel()`
Creates and initializes a new TUI model with the provided settings.

#### `(m viewModel) Init()`, `Update()`, `View()`
Implement the Bubble Tea model interface for UI rendering and state updates.

#### `formatEventDetail()`
Formats an event for detailed display, showing both the raw payload and structured data.

### Event Processing

Events flow through the system as follows:

1. **Loading Events**: When `--use-db` is specified, the appropriate database function is called:
   - `db.GetRunEvents()`: For a specific run ID
   - `db.GetLatestRunEvents()`: For the latest run
   - `db.GetEventsInTimeRange()`: For events within a time range
   - `db.GetLatestEvents()`: For the latest N events

2. **Streaming Events**: A message handler is set up that forwards Redis events to the event channel.

3. **UI Updates**: The `Update()` method processes events and updates the list model. When an event is selected, `formatEventDetail()` generates a detailed view.

## Customization

### Adding New Event Types

To add support for a new event type:

1. Define the new event type constant in `pkg/model/event.go`
2. Add a corresponding struct for the event payload
3. Update the `formatEventDetail()` function in `cmd/tui-viewer/main.go` to handle the new event type

### Extending the UI

The viewer uses the Bubble Tea framework, which makes it relatively straightforward to extend the UI:

1. Add new key bindings in the `keyMap` struct
2. Add handling for the new keys in the `Update()` method
3. Update the UI rendering in the `View()` method

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