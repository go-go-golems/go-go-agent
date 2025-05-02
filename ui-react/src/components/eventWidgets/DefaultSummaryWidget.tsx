import React from 'react';
import { EventSummaryWidgetProps } from './types';
import CodeHighlighter from '../SyntaxHighlighter';
import ErrorBoundary from '../ErrorBoundary';
import SimpleCodeFallback from '../SimpleCodeFallback';

/**
 * Format the event payload in a safe way
 */
const safeDisplayValue = (value: any, beautify = true): string => {
  try {
    if (typeof value === 'object' && value !== null) {
      return beautify 
        ? JSON.stringify(value, null, 2) 
        : JSON.stringify(value);
    }
    return String(value);
  } catch (error) {
    return `[Error displaying value: ${error}]`;
  }
};

/**
 * Default summary widget used as a fallback for any event type without a specific widget
 */
const DefaultSummaryWidget: React.FC<EventSummaryWidgetProps> = ({ event }) => {
  const content = safeDisplayValue(event.payload);
  return (
    <div className="alert alert-secondary small">
      <p><i>Displaying default summary view for event type: <code>{event.event_type}</code></i></p>
      <ErrorBoundary fallback={SimpleCodeFallback} contentForFallback={content}>
        <CodeHighlighter code={content} language="json" maxHeight="200px" />
      </ErrorBoundary>
    </div>
  );
};

export default DefaultSummaryWidget; 