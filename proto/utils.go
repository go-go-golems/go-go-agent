package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NewEvent creates a new Event with the given type and payload
func NewEvent(eventType EventType, payload interface{}) (*Event, error) {
	event := &Event{
		EventId:   uuid.New().String(),
		Timestamp: timestamppb.Now(),
		EventType: eventType,
	}

	switch eventType {
	case EventType_EVENT_TYPE_UNSPECIFIED:
		return nil, fmt.Errorf("cannot create event with unspecified event type")
	case EventType_EVENT_TYPE_STEP_STARTED:
		if p, ok := payload.(*StepStartedPayload); ok {
			event.Payload = &Event_StepStarted{StepStarted: p}
		} else {
			return nil, fmt.Errorf("payload must be *StepStartedPayload for EVENT_TYPE_STEP_STARTED")
		}
	case EventType_EVENT_TYPE_STEP_FINISHED:
		if p, ok := payload.(*StepFinishedPayload); ok {
			event.Payload = &Event_StepFinished{StepFinished: p}
		} else {
			return nil, fmt.Errorf("payload must be *StepFinishedPayload for EVENT_TYPE_STEP_FINISHED")
		}
	case EventType_EVENT_TYPE_NODE_STATUS_CHANGED:
		if p, ok := payload.(*NodeStatusChangePayload); ok {
			event.Payload = &Event_NodeStatusChanged{NodeStatusChanged: p}
		} else {
			return nil, fmt.Errorf("payload must be *NodeStatusChangePayload for EVENT_TYPE_NODE_STATUS_CHANGED")
		}
	case EventType_EVENT_TYPE_LLM_CALL_STARTED:
		if p, ok := payload.(*LlmCallStartedPayload); ok {
			event.Payload = &Event_LlmCallStarted{LlmCallStarted: p}
		} else {
			return nil, fmt.Errorf("payload must be *LlmCallStartedPayload for EVENT_TYPE_LLM_CALL_STARTED")
		}
	case EventType_EVENT_TYPE_LLM_CALL_COMPLETED:
		if p, ok := payload.(*LlmCallCompletedPayload); ok {
			event.Payload = &Event_LlmCallCompleted{LlmCallCompleted: p}
		} else {
			return nil, fmt.Errorf("payload must be *LlmCallCompletedPayload for EVENT_TYPE_LLM_CALL_COMPLETED")
		}
	case EventType_EVENT_TYPE_TOOL_INVOKED:
		if p, ok := payload.(*ToolInvokedPayload); ok {
			event.Payload = &Event_ToolInvoked{ToolInvoked: p}
		} else {
			return nil, fmt.Errorf("payload must be *ToolInvokedPayload for EVENT_TYPE_TOOL_INVOKED")
		}
	case EventType_EVENT_TYPE_TOOL_RETURNED:
		if p, ok := payload.(*ToolReturnedPayload); ok {
			event.Payload = &Event_ToolReturned{ToolReturned: p}
		} else {
			return nil, fmt.Errorf("payload must be *ToolReturnedPayload for EVENT_TYPE_TOOL_RETURNED")
		}
	case EventType_EVENT_TYPE_NODE_CREATED:
		if p, ok := payload.(*NodeCreatedPayload); ok {
			event.Payload = &Event_NodeCreated{NodeCreated: p}
		} else {
			return nil, fmt.Errorf("payload must be *NodeCreatedPayload for EVENT_TYPE_NODE_CREATED")
		}
	case EventType_EVENT_TYPE_PLAN_RECEIVED:
		if p, ok := payload.(*PlanReceivedPayload); ok {
			event.Payload = &Event_PlanReceived{PlanReceived: p}
		} else {
			return nil, fmt.Errorf("payload must be *PlanReceivedPayload for EVENT_TYPE_PLAN_RECEIVED")
		}
	case EventType_EVENT_TYPE_NODE_ADDED:
		if p, ok := payload.(*NodeAddedPayload); ok {
			event.Payload = &Event_NodeAdded{NodeAdded: p}
		} else {
			return nil, fmt.Errorf("payload must be *NodeAddedPayload for EVENT_TYPE_NODE_ADDED")
		}
	case EventType_EVENT_TYPE_EDGE_ADDED:
		if p, ok := payload.(*EdgeAddedPayload); ok {
			event.Payload = &Event_EdgeAdded{EdgeAdded: p}
		} else {
			return nil, fmt.Errorf("payload must be *EdgeAddedPayload for EVENT_TYPE_EDGE_ADDED")
		}
	case EventType_EVENT_TYPE_INNER_GRAPH_BUILT:
		if p, ok := payload.(*InnerGraphBuiltPayload); ok {
			event.Payload = &Event_InnerGraphBuilt{InnerGraphBuilt: p}
		} else {
			return nil, fmt.Errorf("payload must be *InnerGraphBuiltPayload for EVENT_TYPE_INNER_GRAPH_BUILT")
		}
	case EventType_EVENT_TYPE_NODE_RESULT_AVAILABLE:
		if p, ok := payload.(*NodeResultAvailablePayload); ok {
			event.Payload = &Event_NodeResultAvailable{NodeResultAvailable: p}
		} else {
			return nil, fmt.Errorf("payload must be *NodeResultAvailablePayload for EVENT_TYPE_NODE_RESULT_AVAILABLE")
		}
	case EventType_EVENT_TYPE_RUN_STARTED:
		if p, ok := payload.(*RunStartedPayload); ok {
			event.Payload = &Event_RunStarted{RunStarted: p}
		} else {
			return nil, fmt.Errorf("payload must be *RunStartedPayload for EVENT_TYPE_RUN_STARTED")
		}
	case EventType_EVENT_TYPE_RUN_FINISHED:
		if p, ok := payload.(*RunFinishedPayload); ok {
			event.Payload = &Event_RunFinished{RunFinished: p}
		} else {
			return nil, fmt.Errorf("payload must be *RunFinishedPayload for EVENT_TYPE_RUN_FINISHED")
		}
	case EventType_EVENT_TYPE_RUN_ERROR:
		if p, ok := payload.(*RunErrorPayload); ok {
			event.Payload = &Event_RunError{RunError: p}
		} else {
			return nil, fmt.Errorf("payload must be *RunErrorPayload for EVENT_TYPE_RUN_ERROR")
		}
	default:
		// For unknown event types, try to convert the payload to a structpb.Struct
		s, err := ToStruct(payload)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert unknown payload to struct")
		}
		event.Payload = &Event_UnknownPayload{UnknownPayload: s}
	}

	return event, nil
}

// WithRunID sets the run ID for the event
func (e *Event) WithRunID(runID string) *Event {
	e.RunId = &runID
	return e
}

// ToJSON converts an Event to a JSON string
func (e *Event) ToJSON() (string, error) {
	opts := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}
	bytes, err := opts.Marshal(e)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal event to JSON")
	}
	return string(bytes), nil
}

// EventFromJSON parses a JSON string into an Event
func EventFromJSON(jsonStr string) (*Event, error) {
	event := &Event{}
	opts := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	err := opts.Unmarshal([]byte(jsonStr), event)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON to event")
	}
	return event, nil
}

// ToStruct converts an arbitrary value to a structpb.Struct
func ToStruct(v interface{}) (*structpb.Struct, error) {
	if v == nil {
		return structpb.NewStruct(map[string]interface{}{})
	}

	// First convert to JSON bytes
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal value to JSON")
	}

	// Then unmarshal into a structpb.Struct
	s := &structpb.Struct{}
	err = protojson.Unmarshal(bytes, s)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON to struct")
	}

	return s, nil
}

// GetStep returns the step number from an event payload if available
func (e *Event) GetStep() (int32, bool) {
	switch p := e.Payload.(type) {
	case *Event_StepStarted:
		return p.StepStarted.Step, true
	case *Event_StepFinished:
		return p.StepFinished.Step, true
	case *Event_NodeStatusChanged:
		if p.NodeStatusChanged.Step != nil {
			return *p.NodeStatusChanged.Step, true
		}
	case *Event_LlmCallStarted:
		if p.LlmCallStarted.Step != nil {
			return *p.LlmCallStarted.Step, true
		}
	case *Event_LlmCallCompleted:
		if p.LlmCallCompleted.Step != nil {
			return *p.LlmCallCompleted.Step, true
		}
	case *Event_ToolInvoked:
		if p.ToolInvoked.Step != nil {
			return *p.ToolInvoked.Step, true
		}
	case *Event_ToolReturned:
		if p.ToolReturned.Step != nil {
			return *p.ToolReturned.Step, true
		}
	case *Event_NodeCreated:
		if p.NodeCreated.Step != nil {
			return *p.NodeCreated.Step, true
		}
	case *Event_PlanReceived:
		if p.PlanReceived.Step != nil {
			return *p.PlanReceived.Step, true
		}
	case *Event_NodeAdded:
		if p.NodeAdded.Step != nil {
			return *p.NodeAdded.Step, true
		}
	case *Event_EdgeAdded:
		if p.EdgeAdded.Step != nil {
			return *p.EdgeAdded.Step, true
		}
	case *Event_InnerGraphBuilt:
		if p.InnerGraphBuilt.Step != nil {
			return *p.InnerGraphBuilt.Step, true
		}
	case *Event_NodeResultAvailable:
		if p.NodeResultAvailable.Step != nil {
			return *p.NodeResultAvailable.Step, true
		}
	case *Event_RunError:
		if p.RunError.Step != nil {
			return *p.RunError.Step, true
		}
	}
	return 0, false
}

// GetNodeID returns the node ID from an event payload if available
func (e *Event) GetNodeID() (string, bool) {
	switch p := e.Payload.(type) {
	case *Event_StepStarted:
		return p.StepStarted.NodeId, true
	case *Event_StepFinished:
		return p.StepFinished.NodeId, true
	case *Event_NodeStatusChanged:
		return p.NodeStatusChanged.NodeId, true
	case *Event_LlmCallStarted:
		if p.LlmCallStarted.NodeId != nil {
			return *p.LlmCallStarted.NodeId, true
		}
	case *Event_LlmCallCompleted:
		if p.LlmCallCompleted.NodeId != nil {
			return *p.LlmCallCompleted.NodeId, true
		}
	case *Event_ToolInvoked:
		if p.ToolInvoked.NodeId != nil {
			return *p.ToolInvoked.NodeId, true
		}
	case *Event_ToolReturned:
		if p.ToolReturned.NodeId != nil {
			return *p.ToolReturned.NodeId, true
		}
	case *Event_NodeCreated:
		return p.NodeCreated.NodeId, true
	case *Event_PlanReceived:
		return p.PlanReceived.NodeId, true
	case *Event_NodeAdded:
		return p.NodeAdded.AddedNodeId, true
	case *Event_EdgeAdded:
		return p.EdgeAdded.ChildNodeId, true
	case *Event_InnerGraphBuilt:
		return p.InnerGraphBuilt.NodeId, true
	case *Event_NodeResultAvailable:
		return p.NodeResultAvailable.NodeId, true
	case *Event_RunError:
		if p.RunError.NodeId != nil {
			return *p.RunError.NodeId, true
		}
	}
	return "", false
}

// GetTimeStamp returns the timestamp of the event as a time.Time
func (e *Event) GetTimeStamp() time.Time {
	if e.Timestamp != nil {
		return e.Timestamp.AsTime()
	}
	return time.Time{}
}

// EventsResponse helper functions

// NewEventsResponse creates a new EventsResponse with the given status and events
func NewEventsResponse(status ConnectionStatus, events []*Event) *EventsResponse {
	return &EventsResponse{
		Status: status,
		Events: events,
	}
}

// ToJSON converts an EventsResponse to a JSON string
func (er *EventsResponse) ToJSON() (string, error) {
	opts := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}
	bytes, err := opts.Marshal(er)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal events response to JSON")
	}
	return string(bytes), nil
}

// EventsResponseFromJSON parses a JSON string into an EventsResponse
func EventsResponseFromJSON(jsonStr string) (*EventsResponse, error) {
	resp := &EventsResponse{}
	opts := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	err := opts.Unmarshal([]byte(jsonStr), resp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal JSON to events response")
	}
	return resp, nil
}

// Helper functions for event creation

// NewStepStartedEvent creates a new step_started event
func NewStepStartedEvent(step int32, nodeID, nodeGoal, rootID string) (*Event, error) {
	payload := &StepStartedPayload{
		Step:     step,
		NodeId:   nodeID,
		NodeGoal: nodeGoal,
		RootId:   rootID,
	}
	return NewEvent(EventType_EVENT_TYPE_STEP_STARTED, payload)
}

// NewStepFinishedEvent creates a new step_finished event
func NewStepFinishedEvent(step int32, nodeID, actionName, statusAfter string, durationSeconds float64) (*Event, error) {
	payload := &StepFinishedPayload{
		Step:            step,
		NodeId:          nodeID,
		ActionName:      actionName,
		StatusAfter:     statusAfter,
		DurationSeconds: durationSeconds,
	}
	return NewEvent(EventType_EVENT_TYPE_STEP_FINISHED, payload)
}

// NewNodeStatusChangedEvent creates a new node_status_changed event
func NewNodeStatusChangedEvent(nodeID, nodeGoal, oldStatus, newStatus string, step *int32) (*Event, error) {
	payload := &NodeStatusChangePayload{
		NodeId:    nodeID,
		NodeGoal:  nodeGoal,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Step:      step,
	}
	return NewEvent(EventType_EVENT_TYPE_NODE_STATUS_CHANGED, payload)
}

// NewLlmCallStartedEvent creates a new llm_call_started event
func NewLlmCallStartedEvent(agentClass, model string, prompt []*LlmMessage, promptPreview string, step *int32, nodeID, actionName *string, callID string) (*Event, error) {
	payload := &LlmCallStartedPayload{
		AgentClass:    agentClass,
		Model:         model,
		Prompt:        prompt,
		PromptPreview: promptPreview,
		Step:          step,
		NodeId:        nodeID,
		ActionName:    actionName,
		CallId:        callID,
	}
	return NewEvent(EventType_EVENT_TYPE_LLM_CALL_STARTED, payload)
}

// Additional helper functions can be added for other event types as needed
