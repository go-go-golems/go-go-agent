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
 * Summary widget for LLM call started events
 */
export const LlmCallStartedSummary: React.FC<EventSummaryWidgetProps> = ({ 
  event, 
  setActiveTab,
  onNodeClick
}) => {
  if (!isEventType('llm_call_started')(event)) {
    return <div className="alert alert-warning">Invalid event type for LlmCallStartedSummary</div>;
  }

  const { agent_class, model, prompt_preview, node_id, action_name, prompt } = event.payload;
  
  return (
    <>
      <div className="card mb-3">
        <div className="card-header bg-light py-2">
          <strong>LLM Call Information</strong>
        </div>
        <div className="card-body">
          <div className="row g-2">
            <div className="col-md-6">
              <p className="mb-1"><strong>Agent Class:</strong> {agent_class}</p>
              <p className="mb-1"><strong>Model:</strong> {model}</p>
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
            </div>
          </div>
        </div>
      </div>
      <div className="card">
        <div className="card-header bg-light py-2">
          <strong>Prompt Preview</strong>
        </div>
        <div className="card-body">
          <ErrorBoundary 
            fallback={SimpleCodeFallback}
            contentForFallback={formatPreview(prompt_preview)}
          >
            <CodeHighlighter
              code={formatPreview(prompt_preview || '')}
              language="markdown"
              maxHeight="250px"
            />
          </ErrorBoundary>
          {prompt && (
            <div className="text-end mt-2">
              <Button size="sm" variant="outline-primary" onClick={() => setActiveTab && setActiveTab('prompt')}>
                View Full Prompt
              </Button>
            </div>
          )}
        </div>
      </div>
    </>
  );
};

/**
 * Table widget for LLM call started events (for the event table row)
 */
export const LlmCallStartedTable: React.FC<EventTableWidgetProps> = ({
  event,
  className = '',
  showCallIds = false
}) => {
  if (!isEventType('llm_call_started')(event)) {
    return <span className="text-warning">Invalid event</span>;
  }
  
  const { agent_class, model, call_id } = event.payload;
  
  return (
    <small className={className}>
      <strong>Agent:</strong> {agent_class},{' '}
      <strong>Model:</strong> {model}
      {showCallIds && call_id && (
        <span>, <strong>Call ID:</strong> {call_id.substring(0, 8)}...</span>
      )}
    </small>
  );
};

/**
 * Specialized tab for displaying the full prompt
 */
export const LlmCallStartedPromptTab: React.FC<EventTabProps> = ({ event }) => {
  if (!isEventType('llm_call_started')(event)) {
    return <div className="alert alert-warning">Invalid event type for prompt tab</div>;
  }

  // Extract the prompt from the payload
  const { prompt } = event.payload;
  
  if (Array.isArray(prompt)) {
    return (
      <div className="p-3">
        <h5>Full Prompt</h5>
        {prompt.map((message, index) => {
          if (!message || typeof message !== 'object' || typeof message.content === 'undefined') {
            return <div key={index} className="alert alert-warning">Invalid message format</div>;
          }
          
          return (
            <div key={index} className="mb-3">
              <div className="fw-bold">{message.role}</div>
              <ErrorBoundary 
                fallback={SimpleCodeFallback}
                contentForFallback={typeof message.content === 'string' ? message.content : safeDisplayValue(message.content, false)}
              >
                <CodeHighlighter
                  code={typeof message.content === 'string' ? message.content : safeDisplayValue(message.content)}
                  language="markdown"
                  maxHeight="400px"
                />
              </ErrorBoundary>
            </div>
          );
        })}
      </div>
    );
  }
  
  const content = typeof prompt === 'string'
    ? prompt
    : safeDisplayValue(prompt);
    
  const language = typeof prompt === 'string' ? 'markdown' : 'json';

  return (
    <div>
      <h6 className="mb-3">Full Prompt</h6>
      <ErrorBoundary fallback={SimpleCodeFallback} contentForFallback={String(content)}>
        <CodeHighlighter code={content} language={language} />
      </ErrorBoundary>
    </div>
  );
}; 