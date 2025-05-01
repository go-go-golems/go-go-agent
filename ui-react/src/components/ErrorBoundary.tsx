import React, { Component, ErrorInfo, ReactNode } from 'react';
import Alert from 'react-bootstrap/Alert';

// Add a custom Error interface with the _fromFallback flag
interface ErrorWithFallbackFlag extends Error {
  _fromFallback?: boolean;
}

interface Props {
  children: ReactNode;
  fallback?: React.ComponentType<{ error: Error | null; retry: () => void; content?: string }>;
  contentForFallback?: string; // Original content as string for simple display
}

interface State {
  hasError: boolean;
  error: ErrorWithFallbackFlag | null;
  fallbackHasError: boolean; // Track if the fallback component itself fails
}

class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
    fallbackHasError: false
  };

  public static getDerivedStateFromError(error: Error): Partial<State> {
    // Cast to our custom interface
    const errorWithFlag = error as ErrorWithFallbackFlag;
    
    // If fallback is already failing, mark that state too
    if (errorWithFlag._fromFallback) {
      return {
        hasError: true,
        fallbackHasError: true,
        error: errorWithFlag
      };
    }
    
    // Otherwise, just the main component failed
    return {
      hasError: true,
      error: errorWithFlag
    };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    // Add a flag to the error if it came from the fallback
    // Cast to our custom interface
    const errorWithFlag = error as ErrorWithFallbackFlag;
    
    if (this.state.hasError && !this.state.fallbackHasError) {
      errorWithFlag._fromFallback = true;
    }
    
    console.error('Error caught by error boundary:', error);
    console.error('Error info:', errorInfo);
    console.error('Component stack:', errorInfo.componentStack);
  }

  private handleRetry = () => {
    this.setState({ 
      hasError: false, 
      error: null,
      fallbackHasError: false
    });
  };

  public render() {
    const { children, fallback: FallbackComponent, contentForFallback } = this.props;
    
    // No error, render children normally
    if (!this.state.hasError) {
      return children;
    }
    
    // Main component failed, but we have a custom fallback and it hasn't failed
    if (FallbackComponent && !this.state.fallbackHasError) {
      try {
        return (
          <FallbackComponent 
            error={this.state.error} 
            retry={this.handleRetry} 
            content={contentForFallback} 
          />
        );
      } catch (fallbackError) {
        // If rendering the fallback throws synchronously, catch it here
        console.error('Fallback component failed:', fallbackError);
        this.setState({ fallbackHasError: true });
        // Continue to default error UI
      }
    }
    
    // Last resort - both main component and custom fallback failed (or no fallback provided)
    return (
      <Alert variant="danger">
        <Alert.Heading>Something went wrong</Alert.Heading>
        <p>
          {this.state.error?.message || 'An unexpected error occurred'}
        </p>
        {contentForFallback && (
          <div className="mt-2 border p-2 bg-light">
            <small className="text-muted">Original content:</small>
            <pre className="mt-1 user-select-all overflow-auto" style={{ maxHeight: '200px' }}>
              {contentForFallback}
            </pre>
          </div>
        )}
        <hr />
        <div className="d-flex justify-content-end">
          <button
            className="btn btn-outline-danger"
            onClick={this.handleRetry}
          >
            Try again
          </button>
        </div>
      </Alert>
    );
  }
}

export default ErrorBoundary; 