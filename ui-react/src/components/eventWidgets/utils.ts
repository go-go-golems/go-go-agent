/**
 * Format any value in a safe way
 * Handles objects by converting to JSON string, handles primitives
 * and catches errors
 */
export const safeDisplayValue = (value: any, beautify = true): string => {
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
 * Create a truncated preview of text with ellipsis if needed
 */
export const truncateText = (text: string, maxLength: number): string => {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
};

/**
 * Extract a safe preview from a possibly complex value
 */
export const getContentPreview = (
  content: any, 
  maxLength: number = 100, 
  fallbackText: string = '[Complex data]'
): string => {
  if (typeof content === 'string') {
    return truncateText(content, maxLength);
  }
  
  if (typeof content === 'object' && content !== null) {
    try {
      const preview = JSON.stringify(content);
      return truncateText(preview, maxLength);
    } catch (error) {
      return fallbackText;
    }
  }
  
  return String(content);
}; 