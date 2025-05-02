import React from 'react';
import { EventTableWidgetProps } from './types';

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
 * Default table widget used as a fallback for any event type without a specific widget
 */
const DefaultTableWidget: React.FC<EventTableWidgetProps> = ({ event }) => {
  const preview = safeDisplayValue(event.payload, false);
  const truncated = preview.length > 100 
    ? preview.substring(0, 100) + '...' 
    : preview;
    
  return (
    <span title={preview} className="text-muted">
      {truncated}
    </span>
  );
};

export default DefaultTableWidget; 