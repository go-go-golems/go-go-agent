import { AgentEvent, KnownEventType } from '../../features/events/eventsApi';
import { v4 as uuidv4 } from 'uuid';

let eventCounter = 0;

export const mockAgentEvent = (
    eventType: KnownEventType | string = 'unknown_event_type',
    payload: Record<string, any> = { default: 'payload' },
    runId: string = 'run-storybook',
    nodeId?: string
): AgentEvent => {
    eventCounter++;
    const baseEvent = {
        event_id: uuidv4(),
        timestamp: new Date().toISOString(),
        event_type: eventType,
        payload: { ...payload }, // Ensure payload is copied
        run_id: runId,
    };

    // Add common contextual fields often present in payloads
    if (nodeId) {
        baseEvent.payload.node_id = nodeId;
    }
    if (!baseEvent.payload.step) {
        baseEvent.payload.step = eventCounter; // Simple step counter for mock
    }

    // Type assertion based on known types
    switch (eventType) {
        case 'step_started':
            return { 
                ...baseEvent, 
                event_type: 'step_started', 
                payload: { 
                    step: eventCounter, 
                    node_id: nodeId || `node-${eventCounter}`, 
                    node_goal: 'Mock Goal', 
                    root_id: 'root-0', 
                    ...payload 
                } 
            } as AgentEvent;
        case 'llm_call_completed':
            return { 
                ...baseEvent, 
                event_type: 'llm_call_completed', 
                payload: { 
                    agent_class: 'MockAgent', 
                    model: 'gpt-mock', 
                    duration_seconds: 1.23, 
                    response: 'Mock LLM Response', 
                    token_usage: { prompt_tokens: 10, completion_tokens: 20 }, 
                    node_id: nodeId || `node-${eventCounter}`, 
                    ...payload 
                } 
            } as AgentEvent;
        case 'node_created':
            return {
                ...baseEvent,
                event_type: 'node_created',
                payload: {
                    node_id: nodeId || `node-${eventCounter}`,
                    node_type: 'task',
                    node_goal: 'Mock Node Goal',
                    ...payload
                }
            } as AgentEvent;
        case 'node_status_changed':
            return {
                ...baseEvent,
                event_type: 'node_status_changed',
                payload: {
                    node_id: nodeId || `node-${eventCounter}`,
                    old_status: 'pending',
                    new_status: 'running',
                    ...payload
                }
            } as AgentEvent;
        default:
            return baseEvent as AgentEvent;
    }
};

// Example specific mocks
export const mockStepStartedEvent = (nodeId = 'node-1') => 
    mockAgentEvent('step_started', { node_goal: 'Start the process' }, 'run-123', nodeId);

export const mockLlmCompletedEvent = (nodeId = 'node-2') => 
    mockAgentEvent('llm_call_completed', { 
        response: 'This is the generated text.',
        action_name: 'generate' 
    }, 'run-123', nodeId);

export const mockNodeCreatedEvent = (nodeId = 'node-3') =>
    mockAgentEvent('node_created', {
        node_type: 'task',
        node_goal: 'Accomplish something'
    }, 'run-123', nodeId);

export const generateMockEventSequence = (count: number = 10): AgentEvent[] => {
    const events: AgentEvent[] = [];
    const eventTypes: KnownEventType[] = ['step_started', 'llm_call_completed', 'node_created', 'node_status_changed'];
    
    for (let i = 0; i < count; i++) {
        const eventType = eventTypes[i % eventTypes.length] as KnownEventType;
        events.push(mockAgentEvent(eventType));
    }
    
    return events;
}; 