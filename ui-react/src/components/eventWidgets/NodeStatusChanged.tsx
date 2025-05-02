import React from 'react';
import { EventSummaryWidgetProps, EventTableWidgetProps } from './types';
import { isEventType } from '../../helpers/eventType';
import { RenderClickableNodeId } from '../../helpers/formatters';
import { statusColorMap } from '../../helpers/eventConstants';
import { ArrowRight } from 'lucide-react';

/**
 * Summary widget for node_status_changed events
 */
export const NodeStatusChangedSummary: React.FC<EventSummaryWidgetProps> = ({
  event,
  onNodeClick
}) => {
  if (!isEventType('node_status_changed')(event)) {
    return <div className="alert alert-warning">Invalid event type for NodeStatusChangedSummary</div>;
  }

  const { node_id, node_goal, old_status, new_status } = event.payload;
  
  return (
    <div className="card">
      <div className="card-header bg-light py-2">
        <strong>Status Change Details</strong>
      </div>
      <div className="card-body">
        <div className="row g-2">
          <div className="col-md-6">
            <p className="mb-1">
              <strong>Node ID:</strong> 
              <RenderClickableNodeId nodeId={node_id} onNodeClick={onNodeClick} />
            </p>
            <p className="mb-1"><strong>Node Goal:</strong> {node_goal}</p>
          </div>
          <div className="col-md-6">
            <p className="mb-1">
              <strong>Status Change:</strong>
              <span className={statusColorMap[old_status] || ''}> {old_status}</span>
              <ArrowRight size={16} className="mx-1" />
              <span className={statusColorMap[new_status] || ''}>{new_status}</span>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

/**
 * Table widget for node_status_changed events (for the event table row)
 */
export const NodeStatusChangedTable: React.FC<EventTableWidgetProps> = ({
  event,
  className = ''
}) => {
  if (!isEventType('node_status_changed')(event)) {
    return <span className="text-warning">Invalid event</span>;
  }
  
  const { old_status, new_status } = event.payload;
  const oldStatusClass = statusColorMap[old_status] || 'text-dark';
  const newStatusClass = statusColorMap[new_status] || 'text-dark';
  
  return (
    <div className={`d-flex align-items-center ${className}`}>
      <span className={oldStatusClass}>{old_status}</span>
      <ArrowRight size={14} className="mx-2 text-muted" />
      <span className={newStatusClass}>{new_status}</span>
    </div>
  );
}; 