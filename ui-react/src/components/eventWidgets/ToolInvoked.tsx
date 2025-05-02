import React from 'react';
import { EventSummaryWidgetProps, EventTableWidgetProps } from './types';
import { isEventType } from '../../helpers/eventType';
import { RenderClickableNodeId } from '../../helpers/formatters';
import CodeHighlighter from '../SyntaxHighlighter';
import ErrorBoundary from '../ErrorBoundary';
import SimpleCodeFallback from '../SimpleCodeFallback';
import { Badge } from 'react-bootstrap';

/**
 * Format the arguments for display
 */
const formatArgs = (args: unknown): React.ReactNode => {
  if (args === null || args === undefined) {
    return <span className="text-muted">No arguments</span>;
  }
  
  try {
    // Convert to JSON string with formatting
    const formatted = JSON.stringify(args, null, 2);
    
    return (
      <ErrorBoundary fallback={SimpleCodeFallback} contentForFallback={formatted}>
        <CodeHighlighter 
          code={formatted}
          language="json"
          maxHeight="250px"
        />
      </ErrorBoundary>
    );
  } catch (error) {
    return <span className="text-danger">Error displaying arguments: {String(error)}</span>;
  }
};

/**
 * Summary widget for tool_invoked events
 */
export const ToolInvokedSummary: React.FC<EventSummaryWidgetProps> = ({
  event,
  onNodeClick
}) => {
  if (!isEventType('tool_invoked')(event)) {
    return <div className="alert alert-warning">Invalid event type for ToolInvokedSummary</div>;
  }

  const { 
    node_id, 
    tool_name,
    api_name,
    args_summary,
    args,
    agent_class,
    tool_call_id
  } = event.payload;
  
  return (
    <div className="card">
      <div className="card-header bg-light py-2">
        <strong>Tool Invoked</strong>
      </div>
      <div className="card-body">
        <div className="row g-2 mb-3">
          <div className="col-md-6">
            <p className="mb-1">
              <strong>Node ID:</strong> 
              <RenderClickableNodeId nodeId={node_id} onNodeClick={onNodeClick} />
            </p>
            {agent_class && (
              <p className="mb-1">
                <strong>Agent Class:</strong> {agent_class}
              </p>
            )}
            {tool_call_id && (
              <p className="mb-1">
                <strong>Tool Call ID:</strong> <span className="font-monospace small">{tool_call_id}</span>
              </p>
            )}
          </div>
          <div className="col-md-6">
            <p className="mb-1">
              <strong>Tool:</strong> 
              <Badge bg="info" className="ms-1">{tool_name}</Badge>
            </p>
            {api_name && (
              <p className="mb-1">
                <strong>API:</strong> {api_name}
              </p>
            )}
          </div>
        </div>
        
        {args_summary && (
          <div className="mb-3">
            <strong>Arguments Summary:</strong>
            <p className="mb-0 mt-1 p-2 bg-light rounded">{args_summary}</p>
          </div>
        )}
        
        <div>
          <strong>Arguments:</strong>
          <div className="mt-2">
            {formatArgs(args)}
          </div>
        </div>
      </div>
    </div>
  );
};

/**
 * Table widget for tool_invoked events (for the event table row)
 */
export const ToolInvokedTable: React.FC<EventTableWidgetProps> = ({
  event,
  className = ''
}) => {
  if (!isEventType('tool_invoked')(event)) {
    return <span className="text-warning">Invalid event</span>;
  }
  
  const { tool_name, args_summary } = event.payload;
  
  return (
    <div className={`d-flex align-items-center ${className}`}>
      <Badge bg="info" className="me-2">{tool_name}</Badge>
      <span className="text-truncate">{args_summary || 'Tool invoked'}</span>
    </div>
  );
}; 