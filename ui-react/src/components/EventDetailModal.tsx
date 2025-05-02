import React, { useState, useEffect } from 'react';
import { Modal, Tab, Nav, Button, Badge } from 'react-bootstrap';
import { AgentEvent, LlmMessage } from '../features/events/eventsApi';
import { isEventType } from '../helpers/eventType';
import { ArrowRight, ArrowLeft, ChevronLeft, ChevronRight } from 'lucide-react';
import CodeHighlighter from './SyntaxHighlighter';
import { statusColorMap, eventTypeBadgeVariant } from '../helpers/eventConstants.ts';
import { formatTimestamp, RenderClickableNodeId } from '../helpers/formatters.tsx';
import ErrorBoundary from './ErrorBoundary';
import SimpleCodeFallback from './SimpleCodeFallback';
import { getSummaryWidget, getExtraTabs } from './eventWidgets/widgetRegistry';

interface EventDetailModalProps {
  show: boolean;
  onHide: () => void;
  event: AgentEvent | null;
  onNodeClick?: (nodeId: string) => void;
  hasPrevious?: boolean;
  onBack?: () => void;
  onNext?: () => void;
  onPrevious?: () => void;
  hasPreviousEvent?: boolean;
  hasNextEvent?: boolean;
}

/**
 * Safely converts a value to a string for display
 * @param value Value that needs to be displayed safely
 * @param prettyJson Whether to format as pretty JSON
 */
const safeDisplayValue = (value: unknown, prettyJson = true): string => {
  if (value === undefined || value === null) {
    return '';
  }
  
  if (typeof value === 'string') {
    return value;
  }
  
  try {
    return prettyJson ? JSON.stringify(value, null, 2) : JSON.stringify(value);
  } catch (e) {
    console.error('Failed to stringify value:', e);
    return '[Error: Unable to display value]';
  }
};

const EventDetailModal: React.FC<EventDetailModalProps> = ({ 
  show, 
  onHide, 
  event, 
  onNodeClick,
  hasPrevious = false,
  onBack,
  onNext,
  onPrevious,
  hasPreviousEvent = false,
  hasNextEvent = false
}) => {
  const [activeTab, setActiveTab] = useState('summary');
  
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!show) return;
      
      switch (e.key) {
        case 'ArrowLeft':
          if (onPrevious && hasPreviousEvent) {
            e.preventDefault();
            onPrevious();
          }
          break;
        case 'ArrowRight':
          if (onNext && hasNextEvent) {
            e.preventDefault();
            onNext();
          }
          break;
        case 'Escape':
          onHide();
          break;
      }
    };
    
    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [show, onPrevious, onNext, onHide, hasPreviousEvent, hasNextEvent]);
  
  if (!event) return null;

  const copyJsonToClipboard = () => {
    navigator.clipboard.writeText(safeDisplayValue(event, true))
      .then(() => {
        alert('JSON copied to clipboard');
      })
      .catch(err => {
        console.error('Failed to copy JSON: ', err);
      });
  };

  const getBadgeVariant = (eventType: string): string => {
    return eventTypeBadgeVariant[eventType] || eventTypeBadgeVariant.default;
  };

  // Get widgets and tabs from registry
  const SummaryWidget = getSummaryWidget(event.event_type);
  const extraTabs = getExtraTabs(event.event_type);

  const handleBackClick = () => {
    if (onBack) {
      onBack();
    }
  };

  // Event navigation handlers
  const handlePreviousClick = () => {
    if (onPrevious && hasPreviousEvent) {
      onPrevious();
    }
  };

  const handleNextClick = () => {
    if (onNext && hasNextEvent) {
      onNext();
    }
  };

  return (
    <Modal show={show} onHide={onHide} size="lg" aria-labelledby="event-detail-modal" centered>
      <Modal.Header closeButton>
        <div className="d-flex align-items-center w-100">
          {hasPrevious && (
            <Button variant="link" className="p-0 me-2" onClick={handleBackClick}>
              <ArrowLeft size={20} />
            </Button>
          )}
          <Modal.Title className="flex-grow-1">
            <Badge bg={getBadgeVariant(event.event_type)} className="me-2">
              {event.event_type}
            </Badge>
            <small className="text-muted">{formatTimestamp(event.timestamp)}</small>
          </Modal.Title>
          <div className="d-flex">
            <Button 
              variant="outline-secondary" 
              size="sm" 
              className="me-2" 
              onClick={handlePreviousClick}
              disabled={!hasPreviousEvent}
              title="Previous event (Left Arrow)"
            >
              <ChevronLeft size={16} />
            </Button>
            <Button 
              variant="outline-secondary" 
              size="sm" 
              onClick={handleNextClick}
              disabled={!hasNextEvent}
              title="Next event (Right Arrow)"
            >
              <ChevronRight size={16} />
            </Button>
          </div>
        </div>
      </Modal.Header>
      <Modal.Body className="p-0">
        <Tab.Container activeKey={activeTab} onSelect={(k) => setActiveTab(k || 'summary')}>
          <Nav variant="tabs" className="px-3 pt-3">
            <Nav.Item>
              <Nav.Link eventKey="summary">Summary</Nav.Link>
            </Nav.Item>
            {/* Dynamically render extra tabs */}
            {extraTabs.map(tab => (
              <Nav.Item key={tab.key}>
                <Nav.Link eventKey={tab.key}>{tab.title}</Nav.Link>
              </Nav.Item>
            ))}
            <Nav.Item>
              <Nav.Link eventKey="json">JSON</Nav.Link>
            </Nav.Item>
            <Nav.Item>
              <Nav.Link eventKey="metadata">Metadata</Nav.Link>
            </Nav.Item>
          </Nav>
          <Tab.Content className="p-3">
            <Tab.Pane eventKey="summary">
              <ErrorBoundary>
                <SummaryWidget 
                  event={event} 
                  onNodeClick={onNodeClick} 
                  setActiveTab={setActiveTab} 
                />
              </ErrorBoundary>
            </Tab.Pane>
            
            {/* Dynamically render extra tab panes */}
            {extraTabs.map(tab => {
              const TabComponent = tab.component;
              return (
                <Tab.Pane key={tab.key} eventKey={tab.key}>
                  {activeTab === tab.key && (
                    <ErrorBoundary>
                      <TabComponent
                        event={event}
                        onNodeClick={onNodeClick}
                        tabKey={tab.key} // Pass needed props
                      />
                    </ErrorBoundary>
                  )}
                </Tab.Pane>
              );
            })}
            
            <Tab.Pane eventKey="json">
              <div className="d-flex justify-content-end mb-2">
                <Button size="sm" variant="outline-secondary" onClick={copyJsonToClipboard}>
                  Copy JSON
                </Button>
              </div>
              <ErrorBoundary>
                <CodeHighlighter
                  code={safeDisplayValue(event)}
                  language="json"
                  maxHeight="400px"
                  showLineNumbers={true}
                />
              </ErrorBoundary>
            </Tab.Pane>
            <Tab.Pane eventKey="metadata">
              <div className="card">
                <div className="card-header bg-light py-2">
                  <strong>Event Metadata</strong>
                </div>
                <div className="card-body">
                  <table className="table table-sm table-hover">
                    <tbody>
                      <tr>
                        <th style={{ width: '180px' }}>Event ID</th>
                        <td>{event.event_id}</td>
                      </tr>
                      <tr>
                        <th>Event Type</th>
                        <td>{event.event_type}</td>
                      </tr>
                      <tr>
                        <th>Timestamp</th>
                        <td>{formatTimestamp(event.timestamp)}</td>
                      </tr>
                      <tr>
                        <th>Run ID</th>
                        <td>{event.run_id || 'N/A'}</td>
                      </tr>
                      {typeof event.payload === 'object' && event.payload !== null && 'node_id' in event.payload && (
                        <tr>
                          <th>Node ID</th>
                          <td><RenderClickableNodeId nodeId={(event.payload as { node_id: string }).node_id} onNodeClick={onNodeClick} truncate={false} /></td>
                        </tr>
                      )}
                      {(isEventType("llm_call_started")(event) || isEventType("llm_call_completed")(event)) && event.payload.call_id && (
                        <tr>
                          <th>LLM Call ID</th>
                          <td>{event.payload.call_id}</td>
                        </tr>
                      )}
                      {(isEventType("tool_invoked")(event) || isEventType("tool_returned")(event)) && event.payload.tool_call_id && (
                        <tr>
                          <th>Tool Call ID</th>
                          <td>{event.payload.tool_call_id}</td>
                        </tr>
                      )}
                      {isEventType('step_started')(event) && (
                        <tr>
                          <th>Root Node ID</th>
                          <td><RenderClickableNodeId nodeId={event.payload.root_id} onNodeClick={onNodeClick} /></td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </Tab.Pane>
          </Tab.Content>
        </Tab.Container>
      </Modal.Body>
    </Modal>
  );
};

export default EventDetailModal; 