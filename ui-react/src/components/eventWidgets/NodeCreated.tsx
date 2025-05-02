import React from 'react';
import { EventSummaryWidgetProps, EventTableWidgetProps } from './types';
import { isEventType } from '../../helpers/eventType';
import { RenderClickableNodeId } from '../../helpers/formatters';
import { Badge } from 'react-bootstrap';

/**
 * Summary widget for node_created events
 */
export const NodeCreatedSummary: React.FC<EventSummaryWidgetProps> = ({
  event,
  onNodeClick
}) => {
  if (!isEventType('node_created')(event)) {
    return <div className="alert alert-warning">Invalid event type for NodeCreatedSummary</div>;
  }

  const { 
    node_id, 
    node_nid, 
    node_type, 
    task_type, 
    task_goal, 
    layer, 
    outer_node_id, 
    root_node_id 
  } = event.payload;
  
  return (
    <div className="card">
      <div className="card-header bg-light py-2">
        <strong>Node Created</strong>
      </div>
      <div className="card-body">
        <div className="row g-2">
          <div className="col-md-6">
            <p className="mb-1">
              <strong>Node ID:</strong> 
              <RenderClickableNodeId nodeId={node_id} onNodeClick={onNodeClick} />
            </p>
            <p className="mb-1"><strong>NID:</strong> {node_nid}</p>
            <p className="mb-1">
              <strong>Type:</strong> 
              <Badge bg="info" className="ms-1">{node_type}</Badge>
              {task_type && <Badge bg="secondary" className="ms-1">{task_type}</Badge>}
            </p>
            <p className="mb-1"><strong>Layer:</strong> {layer}</p>
          </div>
          <div className="col-md-6">
            {outer_node_id && (
              <p className="mb-1">
                <strong>Outer Node:</strong>
                <RenderClickableNodeId nodeId={outer_node_id} onNodeClick={onNodeClick} />
              </p>
            )}
            <p className="mb-1">
              <strong>Root Node:</strong>
              <RenderClickableNodeId nodeId={root_node_id} onNodeClick={onNodeClick} />
            </p>
          </div>
        </div>
        {task_goal && (
          <div className="mt-3">
            <strong>Goal:</strong>
            <p className="mb-0 mt-1 p-2 bg-light rounded">{task_goal}</p>
          </div>
        )}
      </div>
    </div>
  );
};

/**
 * Table widget for node_created events (for the event table row)
 */
export const NodeCreatedTable: React.FC<EventTableWidgetProps> = ({
  event,
  className = ''
}) => {
  if (!isEventType('node_created')(event)) {
    return <span className="text-warning">Invalid event</span>;
  }
  
  const { node_type, task_type, task_goal } = event.payload;
  
  return (
    <div className={`d-flex align-items-center ${className}`}>
      <Badge bg="info" className="me-1">{node_type}</Badge>
      {task_type && <Badge bg="secondary" className="me-2">{task_type}</Badge>}
      <span className="text-truncate">{task_goal || 'No goal specified'}</span>
    </div>
  );
}; 