import React from 'react';
import { EventSummaryWidgetProps, EventTableWidgetProps } from './types';
import { isEventType } from '../../helpers/eventType';
import { RenderClickableNodeId } from '../../helpers/formatters';
import { Badge } from 'react-bootstrap';
import { ArrowRight } from 'lucide-react';

/**
 * Summary widget for edge_added events
 */
export const EdgeAddedSummary: React.FC<EventSummaryWidgetProps> = ({
  event,
  onNodeClick
}) => {
  if (!isEventType('edge_added')(event)) {
    return <div className="alert alert-warning">Invalid event type for EdgeAddedSummary</div>;
  }

  const { 
    graph_owner_node_id, 
    parent_node_id,
    child_node_id,
    parent_node_nid,
    child_node_nid,
    task_type,
    task_goal
  } = event.payload;
  
  return (
    <div className="card">
      <div className="card-header bg-light py-2">
        <strong>Edge Added to Graph</strong>
      </div>
      <div className="card-body">
        <div className="row g-2 mb-3">
          <div className="col-12">
            <p className="mb-1">
              <strong>Graph Owner:</strong> 
              <RenderClickableNodeId nodeId={graph_owner_node_id} onNodeClick={onNodeClick} />
            </p>
          </div>
        </div>
        
        <div className="mb-3">
          <strong>Edge:</strong>
          <div className="d-flex align-items-center gap-2 mt-1 p-2 bg-light rounded">
            <div>
              <p className="mb-0"><strong>Parent Node:</strong></p>
              <RenderClickableNodeId nodeId={parent_node_id} onNodeClick={onNodeClick} />
              <div className="small text-muted">NID: {parent_node_nid}</div>
            </div>
            <ArrowRight size={20} className="mx-3" />
            <div>
              <p className="mb-0"><strong>Child Node:</strong></p>
              <RenderClickableNodeId nodeId={child_node_id} onNodeClick={onNodeClick} />
              <div className="small text-muted">NID: {child_node_nid}</div>
            </div>
          </div>
        </div>
        
        {task_type && (
          <p className="mb-2">
            <strong>Task Type:</strong> 
            <Badge bg="secondary" className="ms-1">{task_type}</Badge>
          </p>
        )}
        
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
 * Table widget for edge_added events (for the event table row)
 */
export const EdgeAddedTable: React.FC<EventTableWidgetProps> = ({
  event,
  className = '',
  onNodeClick
}) => {
  if (!isEventType('edge_added')(event)) {
    return <span className="text-warning">Invalid event</span>;
  }
  
  const { parent_node_nid, child_node_nid, task_type, parent_node_id, child_node_id } = event.payload;
  
  return (
    <div className={`d-flex align-items-center ${className}`}>
      <span className="text-nowrap">
        {onNodeClick && parent_node_id ? (
          <span 
            className="node-id-link"
            onClick={(e) => {
              e.stopPropagation();
              onNodeClick(parent_node_id);
            }}
            style={{ cursor: 'pointer', textDecoration: 'underline' }}
          >
            {parent_node_nid}
          </span>
        ) : (
          parent_node_nid
        )}
      </span>
      <ArrowRight size={14} className="mx-2" />
      <span className="text-nowrap">
        {onNodeClick && child_node_id ? (
          <span 
            className="node-id-link"
            onClick={(e) => {
              e.stopPropagation();
              onNodeClick(child_node_id);
            }}
            style={{ cursor: 'pointer', textDecoration: 'underline' }}
          >
            {child_node_nid}
          </span>
        ) : (
          child_node_nid
        )}
      </span>
      {task_type && <Badge bg="secondary" className="ms-2">{task_type}</Badge>}
    </div>
  );
}; 