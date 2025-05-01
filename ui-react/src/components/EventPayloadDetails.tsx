import React from 'react';
import { ArrowRight } from 'lucide-react';
import { AgentEvent } from '../features/events/eventsApi';
import { isEventType } from '../helpers/eventType';
import { statusColorMap } from '../helpers/eventConstants';
import { RenderClickableNodeId } from '../helpers/formatters'; // Import the component
import CodeHighlighter from './SyntaxHighlighter';
import ErrorBoundary from './ErrorBoundary';
import SimpleCodeFallback from './SimpleCodeFallback';

/**
 * Safely converts a value to a string for display
 * @param value Value that needs to be displayed safely
 * @param prettyJson Whether to format as pretty JSON
 */
const safeDisplayValue = (value: unknown, prettyJson = true): string => {
  if (value === undefined || value === null) {
    return '';
  }
  
  if (typeof value === 'string') {
    return value;
  }
  
  try {
    return prettyJson ? JSON.stringify(value, null, 2) : JSON.stringify(value);
  } catch (e) {
    console.error('Failed to stringify value:', e);
    return '[Error: Unable to display value]';
  }
};

interface EventPayloadDetailsProps {
  event: AgentEvent;
  onNodeClick?: (nodeId: string) => void; // Propagate the handler
  className?: string; // Allow passing custom classes
  showCallIds?: boolean; // Optionally show call/tool IDs
}

// Extracted from EventTable.tsx
const EventPayloadDetails: React.FC<EventPayloadDetailsProps> = ({ event, onNodeClick, className = '', showCallIds = false }) => {
    if (isEventType("step_started")(event)) {
        return <span className={`text-muted text-truncate d-inline-block ${className}`} style={{ maxWidth: '500px' }}>{event.payload.node_goal}</span>;
    }

    if (isEventType("step_finished")(event)) {
        const statusClass = statusColorMap[event.payload.status_after] || 'text-dark';
        return (
            <div className={className}>
                Action: <span className="fw-medium">{event.payload.action_name}</span>,
                Status: <span className={`fw-medium ${statusClass}`}>{event.payload.status_after}</span>,
                Duration: <span className="fw-medium">{event.payload.duration_seconds?.toFixed(2)}s</span>
            </div>
        );
    }

    if (isEventType("node_status_changed")(event)) {
        const oldStatusClass = statusColorMap[event.payload.old_status] || 'text-dark';
        const newStatusClass = statusColorMap[event.payload.new_status] || 'text-dark';
        return (
            <div className={`d-flex align-items-center ${className}`}>
                <span className={oldStatusClass}>{event.payload.old_status}</span>
                <ArrowRight size={14} className="mx-2 text-muted" />
                <span className={newStatusClass}>{event.payload.new_status}</span>
            </div>
        );
    }

    if (isEventType("llm_call_started")(event)) {
        return (
            <small className={className}>
                <strong>Agent:</strong> {event.payload.agent_class},{' '}
                <strong>Model:</strong> {event.payload.model}
                {showCallIds && event.payload.call_id && (
                    <span>, <strong>Call ID:</strong> {event.payload.call_id.substring(0, 8)}...</span>
                )}
            </small>
        );
    }

    if (isEventType("llm_call_completed")(event)) {
        return (
            <small className={className}>
                <strong>Agent:</strong> {event.payload.agent_class},{' '}
                <strong>Duration:</strong> {event.payload.duration_seconds.toFixed(2)}s
                {event.payload.token_usage && (
                    <span>, <strong>Tokens:</strong> {event.payload.token_usage.prompt_tokens}p + {event.payload.token_usage.completion_tokens}c</span>
                )}
                {showCallIds && event.payload.call_id && (
                    <span>, <strong>Call ID:</strong> {event.payload.call_id.substring(0, 8)}...</span>
                )}
            </small>
        );
    }

    if (isEventType("tool_invoked")(event)) {
        return (
            <small className={className}>
                <strong>Tool:</strong> {event.payload.tool_name}.{event.payload.api_name}
                {showCallIds && event.payload.tool_call_id && (
                    <span>, <strong>Call ID:</strong> {event.payload.tool_call_id.substring(0, 8)}...</span>
                )}
            </small>
        );
    }

    if (isEventType("tool_returned")(event)) {
        return (
            <small className={className}>
                <strong>Tool:</strong> {event.payload.tool_name}.{event.payload.api_name},{' '}
                <strong>State:</strong> {event.payload.state},{' '}
                <strong>Duration:</strong> {event.payload.duration_seconds.toFixed(2)}s
                {showCallIds && event.payload.tool_call_id && (
                    <span>, <strong>Call ID:</strong> {event.payload.tool_call_id.substring(0, 8)}...</span>
                )}
            </small>
        );
    }

    // New event type handlers
    if (isEventType("node_created")(event)) {
        return (
            <div className={`text-truncate ${className}`} style={{ maxWidth: '500px' }}>
                NID: <span className="fw-medium">{event.payload.node_nid}</span>,
                Type: <span className="fw-medium">{event.payload.node_type}</span>,
                Task: <span className="fw-medium">{event.payload.task_type}</span>,
                Layer: <span className="fw-medium">{event.payload.layer}</span>
            </div>
        );
    }

    if (isEventType("plan_received")(event)) {
        const planLength = Array.isArray(event.payload.raw_plan) ? event.payload.raw_plan.length : 0;
        return (
            <div className={`text-truncate ${className}`} style={{ maxWidth: '500px' }}>
                Plan with <span className="fw-medium">{planLength}</span> task(s)
            </div>
        );
    }

    if (isEventType("node_added")(event)) {
        return (
            <div className={`text-truncate ${className}`} style={{ maxWidth: '500px' }}>
                Node <RenderClickableNodeId nodeId={event.payload.added_node_id} onNodeClick={onNodeClick} label={event.payload.added_node_nid} truncate={false} /> added to graph owned by <RenderClickableNodeId nodeId={event.payload.graph_owner_node_id} onNodeClick={onNodeClick} />
            </div>
        );
    }

    if (isEventType("edge_added")(event)) {
        return (
            <div className={`text-truncate ${className}`} style={{ maxWidth: '500px' }}>
                Edge: <RenderClickableNodeId nodeId={event.payload.parent_node_id} onNodeClick={onNodeClick} label={event.payload.parent_node_nid} truncate={false} /> → <RenderClickableNodeId nodeId={event.payload.child_node_id} onNodeClick={onNodeClick} label={event.payload.child_node_nid} truncate={false} />
            </div>
        );
    }

    if (isEventType("inner_graph_built")(event)) {
        return (
            <div className={`text-truncate ${className}`} style={{ maxWidth: '500px' }}>
                Graph built with <span className="fw-medium">{event.payload.node_count}</span> nodes and <span className="fw-medium">{event.payload.edge_count}</span> edges
            </div>
        );
    }

    if (isEventType("node_result_available")(event)) {
        // Extract a preview of the result safely
        let resultPreview = "";
        
        try {
            resultPreview = event.payload.result_summary ? 
                (typeof event.payload.result_summary === 'string' ? 
                    event.payload.result_summary.substring(0, 100) + "..." : 
                    safeDisplayValue(event.payload.result_summary, false).substring(0, 100) + "...") : 
                "(empty result)";
        } catch (e) {
            resultPreview = "(error displaying result)";
            console.error("Error formatting result preview:", e);
        }
        
        return (
            <div className={`text-truncate ${className}`} style={{ maxWidth: '500px' }}>
                Action: <span className="fw-medium">{event.payload.action_name}</span>,
                Result: <span className="text-muted small">
                    <ErrorBoundary 
                        fallback={SimpleCodeFallback}
                        contentForFallback={resultPreview}
                    >
                        {resultPreview}
                    </ErrorBoundary>
                </span>
            </div>
        );
    }

    // Fallback for unknown event types
    return (
        <ErrorBoundary 
            fallback={SimpleCodeFallback}
            contentForFallback={safeDisplayValue(event.payload)}
        >
            <pre className={`text-muted small ${className}`} style={{ maxHeight: '100px', overflowY: 'auto', whiteSpace: 'pre-wrap', wordBreak: 'break-all', fontSize: '0.85em', margin: 0 }}>
                {safeDisplayValue(event.payload)}
            </pre>
        </ErrorBoundary>
    );
}

export default EventPayloadDetails; 