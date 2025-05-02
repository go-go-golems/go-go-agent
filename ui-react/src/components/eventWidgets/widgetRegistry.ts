import React, { ComponentType } from 'react';
import {
  EventSummaryWidgetProps,
  EventTableWidgetProps,
  EventWidgetRegistration,
  TabDefinition
} from './types';
import DefaultSummaryWidget from './DefaultSummaryWidget';
import DefaultTableWidget from './DefaultTableWidget';
import { 
  LlmCallStartedSummary, 
  LlmCallStartedTable, 
  LlmCallStartedPromptTab 
} from './LlmCallStarted';
import { 
  LlmCallCompletedSummary, 
  LlmCallCompletedTable, 
  LlmCallCompletedResponseTab 
} from './LlmCallCompleted';
import {
  StepStartedSummary,
  StepStartedTable
} from './StepStarted';
import {
  StepFinishedSummary,
  StepFinishedTable
} from './StepFinished';
import {
  NodeStatusChangedSummary,
  NodeStatusChangedTable
} from './NodeStatusChanged';

// Storage objects for different widget types
const summaryWidgetRegistry: Record<string, ComponentType<EventSummaryWidgetProps>> = {};
const tableWidgetRegistry: Record<string, ComponentType<EventTableWidgetProps>> = {};
const tabsRegistry: Record<string, Array<TabDefinition>> = {};

// Register default widgets immediately to ensure fallbacks are available
summaryWidgetRegistry.default = DefaultSummaryWidget;
tableWidgetRegistry.default = DefaultTableWidget;

/**
 * Register a widget set for a specific event type
 */
export function registerEventWidget({
  eventType,
  summaryWidget,
  tableWidget,
  extraTabs = []
}: EventWidgetRegistration): void {
  summaryWidgetRegistry[eventType] = summaryWidget;
  tableWidgetRegistry[eventType] = tableWidget;
  tabsRegistry[eventType] = extraTabs;
  console.log(`Registered widgets for event type: ${eventType}`); // Helpful for debugging
}

/**
 * Get the appropriate summary widget component for an event type
 */
export function getSummaryWidget(eventType: string): ComponentType<EventSummaryWidgetProps> {
  return summaryWidgetRegistry[eventType] || summaryWidgetRegistry.default;
}

/**
 * Get the appropriate table widget component for an event type
 */
export function getTableWidget(eventType: string): ComponentType<EventTableWidgetProps> {
  return tableWidgetRegistry[eventType] || tableWidgetRegistry.default;
}

/**
 * Get any extra tabs registered for an event type
 */
export function getExtraTabs(eventType: string): Array<TabDefinition> {
  return tabsRegistry[eventType] || [];
}

/**
 * Register all available widgets
 * This function will be called during application initialization
 */
export function registerAllWidgets() {
  console.log("Registering all event widgets...");
  
  // Register Step events
  registerEventWidget({
    eventType: 'step_started',
    summaryWidget: StepStartedSummary,
    tableWidget: StepStartedTable
  });
  
  registerEventWidget({
    eventType: 'step_finished',
    summaryWidget: StepFinishedSummary,
    tableWidget: StepFinishedTable
  });
  
  // Register Node events
  registerEventWidget({
    eventType: 'node_status_changed',
    summaryWidget: NodeStatusChangedSummary,
    tableWidget: NodeStatusChangedTable
  });
  
  // Register LLM call started widgets
  registerEventWidget({
    eventType: 'llm_call_started',
    summaryWidget: LlmCallStartedSummary,
    tableWidget: LlmCallStartedTable,
    extraTabs: [
      {
        key: 'prompt',
        title: 'Prompt',
        component: LlmCallStartedPromptTab
      }
    ]
  });
  
  // Register LLM call completed widgets
  registerEventWidget({
    eventType: 'llm_call_completed',
    summaryWidget: LlmCallCompletedSummary,
    tableWidget: LlmCallCompletedTable,
    extraTabs: [
      {
        key: 'response',
        title: 'Response',
        component: LlmCallCompletedResponseTab
      }
    ]
  });
  
  // Add more widget registrations here as they are implemented
} 