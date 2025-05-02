import React, { ComponentType } from 'react';
import { AgentEvent, KnownEventType } from '../../features/events/eventsApi';

// Base interface for all event widgets
export interface EventWidgetProps {
  event: AgentEvent;
  onNodeClick?: (nodeId: string) => void;
}

// For widgets in the modal's summary tab
export interface EventSummaryWidgetProps extends EventWidgetProps {
  setActiveTab?: (tabKey: string) => void;
}

// For widgets in the event table rows
export interface EventTableWidgetProps extends EventWidgetProps {
  showCallIds?: boolean; // Existing prop from EventPayloadDetails
  compact?: boolean; // To control density if needed
}

// For specialized tab content in modals
export interface EventTabProps extends EventWidgetProps {
  tabKey: string;
}

export interface TabDefinition {
  key: string;
  title: string;
  component: ComponentType<EventTabProps>;
}

// Registry configuration interface
export interface EventWidgetRegistration {
  eventType: KnownEventType;
  summaryWidget: ComponentType<EventSummaryWidgetProps>;
  tableWidget: ComponentType<EventTableWidgetProps>;
  extraTabs?: Array<TabDefinition>;
} 