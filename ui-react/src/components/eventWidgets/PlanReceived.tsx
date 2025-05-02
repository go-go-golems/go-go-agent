import React from 'react';
import { EventSummaryWidgetProps, EventTableWidgetProps } from './types';
import { isEventType } from '../../helpers/eventType';
import { RenderClickableNodeId } from '../../helpers/formatters';
import CodeHighlighter from '../SyntaxHighlighter';
import ErrorBoundary from '../ErrorBoundary';
import SimpleCodeFallback from '../SimpleCodeFallback';
import { Badge } from 'react-bootstrap';

/**
 * Format the plan for display based on type
 */
const formatPlan = (plan: unknown): React.ReactNode => {
  if (plan === null || plan === undefined) {
    return <span className="text-muted">No plan data</span>;
  }
  
  try {
    // If plan is a string, display it directly
    if (typeof plan === 'string') {
      return (
        <ErrorBoundary fallback={SimpleCodeFallback} contentForFallback={plan}>
          <CodeHighlighter 
            code={plan}
            language="markdown"
            maxHeight="250px"
          />
        </ErrorBoundary>
      );
    }
    
    // If plan is an object or array, format as JSON
    if (typeof plan === 'object') {
      const formattedJson = JSON.stringify(plan, null, 2);
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
    return String(plan);
  } catch (error) {
    return <span className="text-danger">Error displaying plan: {String(error)}</span>;
  }
};

/**
 * Summary widget for plan_received events
 */
export const PlanReceivedSummary: React.FC<EventSummaryWidgetProps> = ({
  event,
  onNodeClick
}) => {
  if (!isEventType('plan_received')(event)) {
    return <div className="alert alert-warning">Invalid event type for PlanReceivedSummary</div>;
  }

  const { 
    node_id, 
    raw_plan,
    task_type,
    task_goal
  } = event.payload;
  
  return (
    <div className="card">
      <div className="card-header bg-light py-2">
        <strong>Plan Received</strong>
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
            {task_type && (
              <p className="mb-1">
                <strong>Task Type:</strong> 
                <Badge bg="secondary" className="ms-1">{task_type}</Badge>
              </p>
            )}
          </div>
        </div>
        
        {task_goal && (
          <div className="mb-3">
            <strong>Goal:</strong>
            <p className="mb-0 mt-1 p-2 bg-light rounded">{task_goal}</p>
          </div>
        )}
        
        <div>
          <strong>Plan:</strong>
          <div className="mt-2">
            {formatPlan(raw_plan)}
          </div>
        </div>
      </div>
    </div>
  );
};

/**
 * Table widget for plan_received events (for the event table row)
 */
export const PlanReceivedTable: React.FC<EventTableWidgetProps> = ({
  event,
  className = ''
}) => {
  if (!isEventType('plan_received')(event)) {
    return <span className="text-warning">Invalid event</span>;
  }
  
  const { task_type, raw_plan } = event.payload;
  
  // Try to create a small preview of the plan
  let planPreview = 'Plan received';
  if (typeof raw_plan === 'string') {
    planPreview = raw_plan.length > 50 ? raw_plan.substring(0, 50) + '...' : raw_plan;
  } else if (Array.isArray(raw_plan) && raw_plan.length > 0) {
    planPreview = `${raw_plan.length} step plan`;
  } else if (raw_plan && typeof raw_plan === 'object') {
    planPreview = 'Plan structure received';
  }
  
  return (
    <div className={`d-flex align-items-center ${className}`}>
      {task_type && <Badge bg="secondary" className="me-2">{task_type}</Badge>}
      <span className="text-truncate">{planPreview}</span>
    </div>
  );
}; 