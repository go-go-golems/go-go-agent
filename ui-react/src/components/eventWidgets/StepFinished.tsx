import React from 'react';
import { EventSummaryWidgetProps, EventTableWidgetProps } from './types';
import { isEventType } from '../../helpers/eventType';
import { RenderClickableNodeId } from '../../helpers/formatters';
import { statusColorMap } from '../../helpers/eventConstants';

/**
 * Summary widget for step_finished events
 */
export const StepFinishedSummary: React.FC<EventSummaryWidgetProps> = ({
  event,
  onNodeClick
}) => {
  if (!isEventType('step_finished')(event)) {
    return <div className="alert alert-warning">Invalid event type for StepFinishedSummary</div>;
  }

  const { step, node_id, action_name, status_after, duration_seconds } = event.payload;
  
  return (
    <div className="card">
      <div className="card-header bg-light py-2">
        <strong>Step Details</strong>
      </div>
      <div className="card-body">
        <div className="row g-2">
          <div className="col-md-6">
            <p className="mb-1"><strong>Step:</strong> {step}</p>
            <p className="mb-1">
              <strong>Node ID:</strong> 
              <RenderClickableNodeId nodeId={node_id} onNodeClick={onNodeClick} />
            </p>
            <p className="mb-1"><strong>Action:</strong> {action_name}</p>
          </div>
          <div className="col-md-6">
            <p className="mb-1">
              <strong>Status After:</strong> 
              <span className={statusColorMap[status_after] || ''}> {status_after}</span>
            </p>
            <p className="mb-1"><strong>Duration:</strong> {duration_seconds.toFixed(2)}s</p>
          </div>
        </div>
      </div>
    </div>
  );
};

/**
 * Table widget for step_finished events (for the event table row)
 */
export const StepFinishedTable: React.FC<EventTableWidgetProps> = ({
  event,
  className = ''
}) => {
  if (!isEventType('step_finished')(event)) {
    return <span className="text-warning">Invalid event</span>;
  }
  
  const { action_name, status_after, duration_seconds } = event.payload;
  const statusClass = statusColorMap[status_after] || 'text-dark';
  
  return (
    <div className={className}>
      Action: <span className="fw-medium">{action_name}</span>,
      Status: <span className={`fw-medium ${statusClass}`}>{status_after}</span>,
      Duration: <span className="fw-medium">{duration_seconds?.toFixed(2)}s</span>
    </div>
  );
}; 