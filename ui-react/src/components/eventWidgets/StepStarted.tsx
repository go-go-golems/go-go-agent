import React from 'react';
import { EventSummaryWidgetProps, EventTableWidgetProps } from './types';
import { isEventType } from '../../helpers/eventType';
import { RenderClickableNodeId } from '../../helpers/formatters';

/**
 * Summary widget for step_started events
 */
export const StepStartedSummary: React.FC<EventSummaryWidgetProps> = ({
  event,
  onNodeClick
}) => {
  if (!isEventType('step_started')(event)) {
    return <div className="alert alert-warning">Invalid event type for StepStartedSummary</div>;
  }

  const { step, node_id, root_id, node_goal } = event.payload;
  
  return (
    <>
      <div className="card mb-3">
        <div className="card-header bg-light py-2">
          <strong>Step Information</strong>
        </div>
        <div className="card-body">
          <div className="row g-2">
            <div className="col-md-6">
              <p className="mb-1"><strong>Step:</strong> {step}</p>
              <p className="mb-1">
                <strong>Node ID:</strong> 
                <RenderClickableNodeId nodeId={node_id} onNodeClick={onNodeClick} />
              </p>
            </div>
            <div className="col-md-6">
              <p className="mb-1">
                <strong>Root ID:</strong> 
                <RenderClickableNodeId nodeId={root_id} onNodeClick={onNodeClick} />
              </p>
            </div>
          </div>
        </div>
      </div>
      <div className="card">
        <div className="card-header bg-light py-2">
          <strong>Node Goal</strong>
        </div>
        <div className="card-body">
          <p>{node_goal}</p>
        </div>
      </div>
    </>
  );
};

/**
 * Table widget for step_started events (for the event table row)
 */
export const StepStartedTable: React.FC<EventTableWidgetProps> = ({
  event,
  className = ''
}) => {
  if (!isEventType('step_started')(event)) {
    return <span className="text-warning">Invalid event</span>;
  }
  
  return (
    <span className={`text-muted text-truncate d-inline-block ${className}`} style={{ maxWidth: '500px' }}>
      {event.payload.node_goal}
    </span>
  );
}; 