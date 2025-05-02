import { AgentEvent, KnownEventType } from '../../features/events/eventsApi';
import { v4 as uuidv4 } from 'uuid';

/**
 * Create a mock agent event with the given parameters
 */
export const mockAgentEvent = (
  eventType: KnownEventType | string = 'unknown_event_type',
  payload: Record<string, any> = {},
  runId: string = 'run-mock-1',
  timestamp: string = new Date().toISOString()
): AgentEvent => {
  return {
    event_id: uuidv4(),
    timestamp,
    event_type: eventType,
    payload: { ...payload },
    run_id: runId,
  };
};

/**
 * Generate a sequence of mock events for testing the event table
 */
export const generateMockEventSequence = (count: number = 10): AgentEvent[] => {
  const events: AgentEvent[] = [];
  const baseTime = new Date();
  
  // Create a run structure with a variety of event types
  for (let i = 0; i < count; i++) {
    // Calculate timestamp (going backward in time)
    const eventTime = new Date(baseTime.getTime() - (i * 1000));
    const timestamp = eventTime.toISOString();
    
    // Create different event types based on position in sequence
    switch (i % 8) {
      case 0:
        events.push(mockStepStartedEvent('node-' + i, timestamp));
        break;
      case 1:
        events.push(mockLlmCallStartedEvent('node-' + i, timestamp));
        break;
      case 2:
        events.push(mockLlmCallCompletedEvent('node-' + i, timestamp));
        break;
      case 3:
        events.push(mockNodeCreatedEvent('node-' + i, timestamp));
        break;
      case 4:
        events.push(mockToolInvokedEvent('node-' + i, timestamp));
        break;
      case 5:
        events.push(mockToolReturnedEvent('node-' + i, timestamp));
        break;
      case 6:
        events.push(mockStepFinishedEvent('node-' + i, timestamp));
        break;
      case 7:
        events.push(mockNodeStatusChangedEvent('node-' + i, timestamp));
        break;
    }
  }
  
  return events;
};

// Specific mock event generators
export const mockStepStartedEvent = (
  nodeId = 'node-123', 
  timestamp?: string
) => mockAgentEvent(
  'step_started', 
  { 
    node_id: nodeId,
    node_goal: 'Gather information about the user request',
  }, 
  'run-mock-1',
  timestamp
);

export const mockStepFinishedEvent = (
  nodeId = 'node-123', 
  timestamp?: string
) => mockAgentEvent(
  'step_finished', 
  { 
    node_id: nodeId,
    status: 'succeeded',
    duration_seconds: 1.23,
    action: 'research',
  }, 
  'run-mock-1',
  timestamp
);

export const mockNodeCreatedEvent = (
  nodeId = 'node-123', 
  timestamp?: string
) => mockAgentEvent(
  'node_created', 
  { 
    node_id: nodeId,
    node_type: 'task',
    goal: 'Research information about the topic',
  }, 
  'run-mock-1',
  timestamp
);

export const mockNodeStatusChangedEvent = (
  nodeId = 'node-123', 
  timestamp?: string
) => mockAgentEvent(
  'node_status_changed', 
  { 
    node_id: nodeId,
    old_status: 'in_progress',
    new_status: 'succeeded',
  }, 
  'run-mock-1',
  timestamp
);

export const mockNodeResultAvailableEvent = (
  nodeId = 'node-123', 
  timestamp?: string
) => mockAgentEvent(
  'node_result_available', 
  { 
    node_id: nodeId,
    result_type: 'text',
    result: 'The research has been completed successfully',
  }, 
  'run-mock-1',
  timestamp
);

export const mockLlmCallStartedEvent = (
  nodeId = 'node-123', 
  timestamp?: string,
  customPayload?: Record<string, any>
) => mockAgentEvent(
  'llm_call_started', 
  { 
    node_id: nodeId,
    agent_class: 'ResearchAgent',
    model: 'gpt-4',
    prompt: 'Research the following topic and provide a detailed summary: Artificial Intelligence',
    ...customPayload
  }, 
  'run-mock-1',
  timestamp
);

export const mockLlmCallCompletedEvent = (
  nodeId = 'node-123', 
  timestamp?: string,
  customPayload?: Record<string, any>
) => mockAgentEvent(
  'llm_call_completed', 
  { 
    node_id: nodeId,
    agent_class: 'ResearchAgent',
    model: 'gpt-4',
    duration_seconds: 2.5,
    token_usage: {
      prompt_tokens: 150,
      completion_tokens: 320,
      total_tokens: 470
    },
    response: 'Artificial Intelligence (AI) refers to the simulation of human intelligence in machines...',
    ...customPayload
  }, 
  'run-mock-1',
  timestamp
);

export const mockToolInvokedEvent = (
  nodeId = 'node-123', 
  timestamp?: string
) => mockAgentEvent(
  'tool_invoked', 
  { 
    node_id: nodeId,
    tool_name: 'web_search',
    tool_api: 'mento-tools',
    tool_args: {
      query: 'latest developments in artificial intelligence',
      max_results: 5
    },
  }, 
  'run-mock-1',
  timestamp
);

export const mockToolReturnedEvent = (
  nodeId = 'node-123', 
  timestamp?: string
) => mockAgentEvent(
  'tool_returned', 
  { 
    node_id: nodeId,
    tool_name: 'web_search',
    status: 'success',
    duration_seconds: 1.8,
    result_summary: 'Found 5 relevant articles about AI advancements',
    result: [
      { title: 'Recent advances in AI research', url: 'https://example.com/ai-research' },
      { title: 'New AI models outperform humans', url: 'https://example.com/ai-performance' }
    ]
  }, 
  'run-mock-1',
  timestamp
);

export const mockPlanReceivedEvent = (
  nodeId = 'node-123', 
  timestamp?: string
) => mockAgentEvent(
  'plan_received', 
  { 
    node_id: nodeId,
    plan_items: [
      { id: 'item-1', description: 'Research the topic' },
      { id: 'item-2', description: 'Analyze findings' },
      { id: 'item-3', description: 'Generate report' }
    ],
    raw_plan: 'I will approach this by first researching the topic thoroughly, then analyzing the findings, and finally generating a comprehensive report.'
  }, 
  'run-mock-1',
  timestamp
);

export const mockEdgeAddedEvent = (
  parentNodeId = 'node-123',
  childNodeId = 'node-456', 
  timestamp?: string
) => mockAgentEvent(
  'edge_added', 
  { 
    parent_node_id: parentNodeId,
    child_node_id: childNodeId,
    owner_node_id: 'node-root',
  }, 
  'run-mock-1',
  timestamp
); 