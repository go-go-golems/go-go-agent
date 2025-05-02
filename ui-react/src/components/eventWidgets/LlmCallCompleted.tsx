import React from 'react';
import { EventSummaryWidgetProps, EventTableWidgetProps, EventTabProps } from './types';
import CodeHighlighter from '../SyntaxHighlighter';
import ErrorBoundary from '../ErrorBoundary';
import SimpleCodeFallback from '../SimpleCodeFallback';
import { isEventType } from '../../helpers/eventType';
import { Button } from 'react-bootstrap';

/**
 * Format the event payload in a safe way
 */
const safeDisplayValue = (value: unknown, beautify = true): string => {
  try {
    if (typeof value === 'object' && value !== null) {
      return beautify 
        ? JSON.stringify(value, null, 2) 
        : JSON.stringify(value);
    }
    return String(value);
  } catch (error) {
    return `[Error displaying value: ${error}]`;
  }
};

// Format preview of text with maximum length
const formatPreview = (text: unknown, maxLength: number = 500): string => {
  const safeText = safeDisplayValue(text, false);
  if (!safeText) return '';
  return safeText.length > maxLength ? safeText.substring(0, maxLength) + '...' : safeText;
};

/**
 * Summary widget for LLM call completed events
 */
export const LlmCallCompletedSummary: React.FC<EventSummaryWidgetProps> = ({
  event,
  setActiveTab,
  onNodeClick
}) => {
  if (!isEventType('llm_call_completed')(event)) {
    return <div className="alert alert-warning">Invalid event type for LlmCallCompletedSummary</div>;
  }

  const { 
    agent_class, 
    model, 
    response, 
    duration_seconds,
    token_usage,
    node_id,
    action_name,
    error
  } = event.payload;
  
  return (
    <>
      <div className="card mb-3">
        <div className="card-header bg-light py-2">
          <strong>LLM Response Information</strong>
        </div>
        <div className="card-body">
          <div className="row g-2">
            <div className="col-md-6">
              <p className="mb-1"><strong>Agent Class:</strong> {agent_class}</p>
              <p className="mb-1"><strong>Model:</strong> {model}</p>
              <p className="mb-1"><strong>Duration:</strong> {duration_seconds.toFixed(2)}s</p>
            </div>
            <div className="col-md-6">
              <p className="mb-1"><strong>Action:</strong> {action_name || 'N/A'}</p>
              {node_id && (
                <p className="mb-1">
                  <strong>Node ID:</strong> 
                  {onNodeClick ? (
                    <span 
                      className="node-id-link ms-1"
                      onClick={() => onNodeClick(node_id)}
                      style={{ cursor: 'pointer', textDecoration: 'underline' }}
                    >
                      {node_id}
                    </span>
                  ) : (
                    <span className="ms-1">{node_id}</span>
                  )}
                </p>
              )}
              {token_usage && (
                <p className="mb-1">
                  <strong>Tokens:</strong> {token_usage.prompt_tokens} / {token_usage.completion_tokens}
                </p>
              )}
            </div>
          </div>
          {error && (
            <div className="alert alert-danger mt-2">
              <strong>Error:</strong> {error}
            </div>
          )}
        </div>
      </div>
      <div className="card">
        <div className="card-header bg-light py-2">
          <strong>Response Preview</strong>
        </div>
        <div className="card-body">
          <ErrorBoundary 
            fallback={SimpleCodeFallback}
            contentForFallback={formatPreview(response || '')}
          >
            <CodeHighlighter
              code={formatPreview(response || '')}
              language="markdown"
              maxHeight="250px"
            />
          </ErrorBoundary>
          <div className="text-end mt-2">
            <Button size="sm" variant="outline-primary" onClick={() => setActiveTab && setActiveTab('response')}>
              View Full Response
            </Button>
          </div>
        </div>
      </div>
    </>
  );
};

/**
 * Table widget for LLM call completed events (for the event table row)
 */
export const LlmCallCompletedTable: React.FC<EventTableWidgetProps> = ({
  event,
  className = '',
  showCallIds = false
}) => {
  if (!isEventType('llm_call_completed')(event)) {
    return <span className="text-warning">Invalid event</span>;
  }

  const { 
    agent_class, 
    duration_seconds,
    token_usage,
    call_id
  } = event.payload;
  
  return (
    <small className={className}>
      <strong>Agent:</strong> {agent_class},{' '}
      <strong>Duration:</strong> {duration_seconds.toFixed(2)}s
      {token_usage && (
        <span>, <strong>Tokens:</strong> {token_usage.prompt_tokens}p + {token_usage.completion_tokens}c</span>
      )}
      {showCallIds && call_id && (
        <span>, <strong>Call ID:</strong> {call_id.substring(0, 8)}...</span>
      )}
    </small>
  );
};

/**
 * Specialized tab for displaying the full response
 */
export const LlmCallCompletedResponseTab: React.FC<EventTabProps> = ({ event }) => {
  if (!isEventType('llm_call_completed')(event)) {
    return <div className="alert alert-warning">Invalid event type for response tab</div>;
  }

  const { response } = event.payload;
  
  return (
    <div className="p-3">
      <h5>Full Response</h5>
      <ErrorBoundary 
        fallback={SimpleCodeFallback}
        contentForFallback={response || ''}
      >
        <CodeHighlighter
          code={response || ''}
          language="markdown"
          maxHeight="400px"
        />
      </ErrorBoundary>
    </div>
  );
}; 