import React from 'react';
import { Alert } from 'react-bootstrap';
import './styles.css'; // Make sure styles are imported

interface SimpleCodeFallbackProps {
  error: Error | null;
  retry: () => void;
  content?: string;
}

/**
 * A simple fallback component when the syntax highlighter fails.
 * Displays code as plain text in a pre block.
 */
const SimpleCodeFallback: React.FC<SimpleCodeFallbackProps> = ({ error, retry, content }) => {
  if (!content) {
    return (
      <Alert variant="warning">
        <Alert.Heading>Syntax highlighter failed</Alert.Heading>
        <p>No content was provided to display as fallback.</p>
        <div className="d-flex justify-content-end mt-2">
          <button className="btn btn-sm btn-outline-warning" onClick={retry}>
            Try again
          </button>
        </div>
      </Alert>
    );
  }

  return (
    <div className="simple-code-fallback">
      <div className="mb-2 text-muted">
        <small>
          <i className="me-1">Syntax highlighting failed. Showing plain text.</i>
          <button 
            className="btn btn-sm btn-link p-0 ms-2" 
            onClick={retry}
          >
            Retry
          </button>
        </small>
      </div>
      <div className="border rounded bg-light">
        <pre 
          className="m-0 p-3 user-select-all overflow-auto" 
          style={{ 
            maxHeight: '400px',
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
            overflowWrap: 'break-word'
          }}
        >
          {content}
        </pre>
      </div>
    </div>
  );
};

export default SimpleCodeFallback; 