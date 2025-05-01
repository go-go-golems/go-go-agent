import React from 'react';
import { AgentEvent } from '../../features/events/eventsApi';
import { Table } from 'react-bootstrap';

export interface DefaultTableWidgetProps {
  event: AgentEvent;
  showCallIds?: boolean;
  compact?: boolean;
  onNodeClick?: (nodeId: string) => void;
}

const DefaultTableWidget: React.FC<DefaultTableWidgetProps> = ({
  event,
  showCallIds = true,
  compact = false,
  onNodeClick,
}) => {
  const handleNodeClick = (nodeId: string) => (e: React.MouseEvent) => {
    e.preventDefault();
    onNodeClick?.(nodeId);
  };

  const renderValue = (value: any): React.ReactNode => {
    if (value === null || value === undefined) return '-';
    if (typeof value === 'object') return JSON.stringify(value);
    return String(value);
  };

  const renderNodeId = (nodeId: string | undefined) => {
    if (!nodeId) return '-';
    return onNodeClick ? (
      <a href="#" onClick={handleNodeClick(nodeId)} className="text-primary">
        {nodeId}
      </a>
    ) : (
      nodeId
    );
  };

  return (
    <Table striped bordered hover size={compact ? 'sm' : undefined} className="mb-0">
      <tbody>
        <tr>
          <td className="fw-bold" style={{ width: '30%' }}>Event Type</td>
          <td>{event.event_type}</td>
        </tr>
        {showCallIds && (
          <>
            <tr>
              <td className="fw-bold">Event ID</td>
              <td>{event.event_id}</td>
            </tr>
            <tr>
              <td className="fw-bold">Run ID</td>
              <td>{event.run_id}</td>
            </tr>
          </>
        )}
        {event.payload.node_id && (
          <tr>
            <td className="fw-bold">Node ID</td>
            <td>{renderNodeId(event.payload.node_id)}</td>
          </tr>
        )}
        {Object.entries(event.payload)
          .filter(([key]) => !['node_id'].includes(key))
          .map(([key, value]) => (
            <tr key={key}>
              <td className="fw-bold">{key.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase())}</td>
              <td>{renderValue(value)}</td>
            </tr>
          ))}
      </tbody>
    </Table>
  );
};

export default DefaultTableWidget; 