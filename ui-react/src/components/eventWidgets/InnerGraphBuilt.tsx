import React from 'react';
import { EventSummaryWidgetProps, EventTableWidgetProps } from './types';
import { isEventType } from '../../helpers/eventType';
import { RenderClickableNodeId } from '../../helpers/formatters';
import { Badge } from 'react-bootstrap';

/**
 * Summary widget for inner_graph_built events
 */
export const InnerGraphBuiltSummary: React.FC<EventSummaryWidgetProps> = ({
  event,
  onNodeClick
}) => {
  if (!isEventType('inner_graph_built')(event)) {
    return <div className="alert alert-warning">Invalid event type for InnerGraphBuiltSummary</div>;
  }

  const { 
    node_id, 
    node_count,
    edge_count,
    node_ids,
    task_type,
    task_goal
  } = event.payload;
  
  return (
    <div className="card">
      <div className="card-header bg-light py-2">
        <strong>Inner Graph Built</strong>
      </div>
      <div className="card-body">
        <div className="row g-2 mb-3">
          <div className="col-md-6">
            <p className="mb-1">
              <strong>Owner Node ID:</strong> 
              <RenderClickableNodeId nodeId={node_id} onNodeClick={onNodeClick} />
            </p>
            {task_type && (
              <p className="mb-1">
                <strong>Task Type:</strong> 
                <Badge bg="secondary" className="ms-1">{task_type}</Badge>
              </p>
            )}
          </div>
          <div className="col-md-6">
            <p className="mb-1">
              <strong>Graph Size:</strong> {node_count} nodes, {edge_count} edges
            </p>
          </div>
        </div>
        
        {task_goal && (
          <div className="mb-3">
            <strong>Goal:</strong>
            <p className="mb-0 mt-1 p-2 bg-light rounded">{task_goal}</p>
          </div>
        )}
        
        {node_ids && node_ids.length > 0 && (
          <div>
            <strong>Inner Node IDs:</strong>
            <div className="mt-2 p-3 bg-light rounded" style={{ maxHeight: '200px', overflowY: 'auto' }}>
              <ul className="mb-0">
                {node_ids.map((nodeId: string, index: number) => (
                  <li key={index}>
                    <RenderClickableNodeId nodeId={nodeId} onNodeClick={onNodeClick} />
                  </li>
                ))}
              </ul>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

/**
 * Table widget for inner_graph_built events (for the event table row)
 */
export const InnerGraphBuiltTable: React.FC<EventTableWidgetProps> = ({
  event,
  className = ''
}) => {
  if (!isEventType('inner_graph_built')(event)) {
    return <span className="text-warning">Invalid event</span>;
  }
  
  const { node_count, edge_count, task_type } = event.payload;
  
  return (
    <div className={`d-flex align-items-center ${className}`}>
      <span>Built inner graph with {node_count} nodes, {edge_count} edges</span>
      {task_type && <Badge bg="secondary" className="ms-2">{task_type}</Badge>}
    </div>
  );
}; 