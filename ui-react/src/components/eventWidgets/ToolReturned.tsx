import React from 'react';
import { EventSummaryWidgetProps, EventTableWidgetProps } from './types';
import { isEventType } from '../../helpers/eventType';
import { RenderClickableNodeId } from '../../helpers/formatters';
import CodeHighlighter from '../SyntaxHighlighter';
import ErrorBoundary from '../ErrorBoundary';
import SimpleCodeFallback from '../SimpleCodeFallback';
import { Badge } from 'react-bootstrap';

/**
 * Format the result for display
 */
const formatResult = (result: unknown): React.ReactNode => {
  if (result === null || result === undefined) {
    return <span className="text-muted">No result data</span>;
  }
  
  try {
    // If result is a string, display it directly
    if (typeof result === 'string') {
      return (
        <ErrorBoundary fallback={SimpleCodeFallback} contentForFallback={result}>
          <CodeHighlighter 
            code={result}
            language="markdown"
            maxHeight="250px"
          />
        </ErrorBoundary>
      );
    }
    
    // If result is an object or array, format as JSON
    if (typeof result === 'object') {
      const formattedJson = JSON.stringify(result, null, 2);
      return (
        <ErrorBoundary fallback={SimpleCodeFallback} contentForFallback={formattedJson}>
          <CodeHighlighter 
            code={formattedJson}
            language="json"
            maxHeight="250px"
          />
        </ErrorBoundary>
      );
    }
    
    // Fallback for other types
    return String(result);
  } catch (error) {
    return <span className="text-danger">Error displaying result: {String(error)}</span>;
  }
};

/**
 * Get status badge based on state
 */
const getStatusBadge = (state: string): React.ReactNode => {
  let variant = 'secondary';
  
  if (state === 'success') {
    variant = 'success';
  } else if (state === 'error' || state === 'failed') {
    variant = 'danger';
  } else if (state === 'timeout') {
    variant = 'warning';
  }
  
  return <Badge bg={variant}>{state}</Badge>;
};

/**
 * Summary widget for tool_returned events
 */
export const ToolReturnedSummary: React.FC<EventSummaryWidgetProps> = ({
  event,
  onNodeClick
}) => {
  if (!isEventType('tool_returned')(event)) {
    return <div className="alert alert-warning">Invalid event type for ToolReturnedSummary</div>;
  }

  const { 
    node_id, 
    tool_name,
    api_name,
    state,
    duration_seconds,
    result_summary,
    result,
    error,
    agent_class,
    tool_call_id
  } = event.payload;
  
  return (
    <div className="card">
      <div className="card-header bg-light py-2">
        <strong>Tool Returned</strong>
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
            <p className="mb-1">
              <strong>Status:</strong> <span className="ms-1">{getStatusBadge(state)}</span>
            </p>
            {duration_seconds !== undefined && (
              <p className="mb-1">
                <strong>Duration:</strong> {duration_seconds.toFixed(2)}s
              </p>
            )}
          </div>
        </div>
        
        {error && (
          <div className="alert alert-danger mb-3">
            <strong>Error:</strong> {error}
          </div>
        )}
        
        {result_summary && (
          <div className="mb-3">
            <strong>Result Summary:</strong>
            <p className="mb-0 mt-1 p-2 bg-light rounded">{result_summary}</p>
          </div>
        )}
        
        {result !== undefined && (
          <div>
            <strong>Result:</strong>
            <div className="mt-2">
              {formatResult(result)}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

/**
 * Table widget for tool_returned events (for the event table row)
 */
export const ToolReturnedTable: React.FC<EventTableWidgetProps> = ({
  event,
  className = ''
}) => {
  if (!isEventType('tool_returned')(event)) {
    return <span className="text-warning">Invalid event</span>;
  }
  
  const { tool_name, state, result_summary, duration_seconds } = event.payload;
  
  return (
    <div className={`d-flex align-items-center ${className}`}>
      <Badge bg="info" className="me-2">{tool_name}</Badge>
      <span className="me-2">{getStatusBadge(state)}</span>
      {duration_seconds !== undefined && (
        <span className="me-2 text-muted small">{duration_seconds.toFixed(2)}s</span>
      )}
      <span className="text-truncate">{result_summary || 'Tool returned'}</span>
    </div>
  );
}; 