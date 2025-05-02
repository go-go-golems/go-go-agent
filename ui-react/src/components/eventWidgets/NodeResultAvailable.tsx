import React from 'react';
import { EventSummaryWidgetProps, EventTableWidgetProps } from './types';
import { isEventType } from '../../helpers/eventType';
import { RenderClickableNodeId } from '../../helpers/formatters';
import CodeHighlighter from '../SyntaxHighlighter';
import ErrorBoundary from '../ErrorBoundary';
import SimpleCodeFallback from '../SimpleCodeFallback';
import { Badge } from 'react-bootstrap';

/**
 * Format the result for display based on type
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
 * Summary widget for node_result_available events
 */
export const NodeResultAvailableSummary: React.FC<EventSummaryWidgetProps> = ({
  event,
  onNodeClick
}) => {
  if (!isEventType('node_result_available')(event)) {
    return <div className="alert alert-warning">Invalid event type for NodeResultAvailableSummary</div>;
  }

  const { 
    node_id, 
    action_name,
    result_summary,
    result
  } = event.payload;
  
  return (
    <div className="card">
      <div className="card-header bg-light py-2">
        <strong>Node Result Available</strong>
      </div>
      <div className="card-body">
        <div className="row g-2 mb-3">
          <div className="col-md-6">
            <p className="mb-1">
              <strong>Node ID:</strong> 
              <RenderClickableNodeId nodeId={node_id} onNodeClick={onNodeClick} />
            </p>
          </div>
          <div className="col-md-6">
            {action_name && (
              <p className="mb-1">
                <strong>Action:</strong> 
                <Badge bg="info" className="ms-1">{action_name}</Badge>
              </p>
            )}
          </div>
        </div>
        
        {result_summary && (
          <div className="mb-3">
            <strong>Summary:</strong>
            <p className="mb-0 mt-1 p-2 bg-light rounded">{result_summary}</p>
          </div>
        )}
        
        <div>
          <strong>Result:</strong>
          <div className="mt-2">
            {formatResult(result)}
          </div>
        </div>
      </div>
    </div>
  );
};

/**
 * Table widget for node_result_available events (for the event table row)
 */
export const NodeResultAvailableTable: React.FC<EventTableWidgetProps> = ({
  event,
  className = ''
}) => {
  if (!isEventType('node_result_available')(event)) {
    return <span className="text-warning">Invalid event</span>;
  }
  
  const { action_name, result_summary } = event.payload;
  
  return (
    <div className={`d-flex align-items-center ${className}`}>
      {action_name && <Badge bg="info" className="me-2">{action_name}</Badge>}
      <span className="text-truncate">{result_summary || 'Result available'}</span>
    </div>
  );
}; 