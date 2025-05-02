import React from 'react';
import { EventSummaryWidgetProps, EventTableWidgetProps } from './types';
import { isEventType } from '../../helpers/eventType';
import { RenderClickableNodeId } from '../../helpers/formatters';
import { Badge } from 'react-bootstrap';

/**
 * Summary widget for node_added events
 */
export const NodeAddedSummary: React.FC<EventSummaryWidgetProps> = ({
  event,
  onNodeClick
}) => {
  if (!isEventType('node_added')(event)) {
    return <div className="alert alert-warning">Invalid event type for NodeAddedSummary</div>;
  }

  const { 
    graph_owner_node_id, 
    added_node_id,
    added_node_nid,
    task_type,
    task_goal
  } = event.payload;
  
  return (
    <div className="card">
      <div className="card-header bg-light py-2">
        <strong>Node Added to Graph</strong>
      </div>
      <div className="card-body">
        <div className="row g-2">
          <div className="col-md-6">
            <p className="mb-1">
              <strong>Added Node ID:</strong> 
              <RenderClickableNodeId nodeId={added_node_id} onNodeClick={onNodeClick} />
            </p>
            <p className="mb-1"><strong>Added Node NID:</strong> {added_node_nid}</p>
            {task_type && (
              <p className="mb-1">
                <strong>Task Type:</strong> 
                <Badge bg="secondary" className="ms-1">{task_type}</Badge>
              </p>
            )}
          </div>
          <div className="col-md-6">
            <p className="mb-1">
              <strong>Graph Owner Node:</strong>
              <RenderClickableNodeId nodeId={graph_owner_node_id} onNodeClick={onNodeClick} />
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
 * Table widget for node_added events (for the event table row)
 */
export const NodeAddedTable: React.FC<EventTableWidgetProps> = ({
  event,
  className = ''
}) => {
  if (!isEventType('node_added')(event)) {
    return <span className="text-warning">Invalid event</span>;
  }
  
  const { added_node_nid, task_type, graph_owner_node_id } = event.payload;
  
  return (
    <div className={`d-flex align-items-center ${className}`}>
      <span>Added node <code>{added_node_nid}</code> to graph </span>
      <RenderClickableNodeId 
        nodeId={graph_owner_node_id} 
        truncate={true} 
        className="ms-1" 
      />
      {task_type && <Badge bg="secondary" className="ms-2">{task_type}</Badge>}
    </div>
  );
}; 