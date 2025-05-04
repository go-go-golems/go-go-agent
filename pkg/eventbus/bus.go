package eventbus

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	events "github.com/go-go-golems/go-go-agent/proto"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// EventBus provides methods for publishing agent events.
type EventBus struct {
	publisher message.Publisher
	topic     string
	encoder   func(event *events.Event) ([]byte, error)
}

// EventBusOption defines options for configuring the EventBus.
type EventBusOption func(*EventBus)

// WithPublisher sets the Watermill publisher.
func WithPublisher(publisher message.Publisher) EventBusOption {
	return func(eb *EventBus) {
		eb.publisher = publisher
	}
}

// WithTopic sets the topic to publish events to.
func WithTopic(topic string) EventBusOption {
	return func(eb *EventBus) {
		eb.topic = topic
	}
}

// WithEncoder sets the function used to encode events before publishing.
// Defaults to protojson encoding.
func WithEncoder(encoder func(event *events.Event) ([]byte, error)) EventBusOption {
	return func(eb *EventBus) {
		eb.encoder = encoder
	}
}

// DefaultJSONEncoder encodes events using protojson.
func DefaultJSONEncoder(event *events.Event) ([]byte, error) {
	opts := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}
	b, err := opts.Marshal(event)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal event to JSON")
	}
	return b, nil
}

// DefaultProtoEncoder encodes events using protobuf binary format.
func DefaultProtoEncoder(event *events.Event) ([]byte, error) {
	b, err := proto.Marshal(event)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal event to proto binary")
	}
	return b, nil
}

// NewEventBus creates a new EventBus.
func NewEventBus(options ...EventBusOption) (*EventBus, error) {
	eb := &EventBus{
		encoder: DefaultJSONEncoder, // Default to JSON encoding
	}
	for _, option := range options {
		option(eb)
	}

	if eb.publisher == nil {
		return nil, errors.New("publisher is required")
	}
	if eb.topic == "" {
		return nil, errors.New("topic is required")
	}

	return eb, nil
}

// Publish encodes and sends an event to the configured topic.
func (eb *EventBus) Publish(ctx context.Context, event *events.Event) error {
	if eb.publisher == nil {
		return errors.New("cannot publish event, publisher is not configured")
	}
	if event == nil {
		return errors.New("cannot publish nil event")
	}

	payloadBytes, err := eb.encoder(event)
	if err != nil {
		return errors.Wrap(err, "failed to encode event")
	}

	msg := message.NewMessage(watermill.NewUUID(), payloadBytes)
	// TODO(manuel): Add metadata? Maybe event_type?
	// msg.Metadata.Set("event_type", event.EventType.String())

	err = eb.publisher.Publish(eb.topic, msg)
	if err != nil {
		return errors.Wrapf(err, "failed to publish event (id: %s, type: %s) to topic %s",
			event.EventId, event.EventType.String(), eb.topic)
	}

	return nil
}

// Close closes the underlying publisher.
func (eb *EventBus) Close() error {
	if eb.publisher != nil {
		return eb.publisher.Close()
	}
	return nil
}

// --- Helper methods for specific event types --- //

// Emit sends a pre-constructed event.
func (eb *EventBus) Emit(ctx context.Context, event *events.Event, runID *string) error {
	if runID != nil {
		event.RunId = runID
	}
	return eb.Publish(ctx, event)
}

func (eb *EventBus) emitEvent(ctx context.Context, eventType events.EventType, payload interface{}, runID *string) error {
	e, err := events.NewEvent(eventType, payload)
	if err != nil {
		return errors.Wrapf(err, "failed to create event %s", eventType.String())
	}
	return eb.Emit(ctx, e, runID)
}

func (eb *EventBus) EmitStepStarted(ctx context.Context, payload *events.StepStartedPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_STEP_STARTED, payload, runID)
}

func (eb *EventBus) EmitStepFinished(ctx context.Context, payload *events.StepFinishedPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_STEP_FINISHED, payload, runID)
}

func (eb *EventBus) EmitNodeStatusChanged(ctx context.Context, payload *events.NodeStatusChangePayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_NODE_STATUS_CHANGED, payload, runID)
}

func (eb *EventBus) EmitLlmCallStarted(ctx context.Context, payload *events.LlmCallStartedPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_LLM_CALL_STARTED, payload, runID)
}

func (eb *EventBus) EmitLlmCallCompleted(ctx context.Context, payload *events.LlmCallCompletedPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_LLM_CALL_COMPLETED, payload, runID)
}

func (eb *EventBus) EmitToolInvoked(ctx context.Context, payload *events.ToolInvokedPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_TOOL_INVOKED, payload, runID)
}

func (eb *EventBus) EmitToolReturned(ctx context.Context, payload *events.ToolReturnedPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_TOOL_RETURNED, payload, runID)
}

func (eb *EventBus) EmitNodeCreated(ctx context.Context, payload *events.NodeCreatedPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_NODE_CREATED, payload, runID)
}

func (eb *EventBus) EmitPlanReceived(ctx context.Context, payload *events.PlanReceivedPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_PLAN_RECEIVED, payload, runID)
}

func (eb *EventBus) EmitNodeAdded(ctx context.Context, payload *events.NodeAddedPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_NODE_ADDED, payload, runID)
}

func (eb *EventBus) EmitEdgeAdded(ctx context.Context, payload *events.EdgeAddedPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_EDGE_ADDED, payload, runID)
}

func (eb *EventBus) EmitInnerGraphBuilt(ctx context.Context, payload *events.InnerGraphBuiltPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_INNER_GRAPH_BUILT, payload, runID)
}

func (eb *EventBus) EmitNodeResultAvailable(ctx context.Context, payload *events.NodeResultAvailablePayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_NODE_RESULT_AVAILABLE, payload, runID)
}

func (eb *EventBus) EmitRunStarted(ctx context.Context, payload *events.RunStartedPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_RUN_STARTED, payload, runID)
}

func (eb *EventBus) EmitRunFinished(ctx context.Context, payload *events.RunFinishedPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_RUN_FINISHED, payload, runID)
}

func (eb *EventBus) EmitRunError(ctx context.Context, payload *events.RunErrorPayload, runID *string) error {
	return eb.emitEvent(ctx, events.EventType_EVENT_TYPE_RUN_ERROR, payload, runID)
}
