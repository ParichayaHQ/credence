import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
  errorInfo?: ErrorInfo;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return {
      hasError: true,
      error,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    
    this.setState({
      error,
      errorInfo,
    });

    // Report error to monitoring service in production
    if (process.env.NODE_ENV === 'production') {
      // TODO: Add error reporting service
    }
  }

  handleReload = () => {
    window.location.reload();
  };

  handleRestart = () => {
    if (window.walletAPI?.window) {
      window.walletAPI.window.close();
    }
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="error-boundary">
          <div className="error-container">
            <div className="error-icon">⚠️</div>
            
            <h1 className="error-title">Something went wrong</h1>
            
            <div className="error-message">
              {this.state.error?.message || 'An unexpected error occurred'}
            </div>

            {process.env.NODE_ENV === 'development' && this.state.errorInfo && (
              <details className="error-details">
                <summary>Error Details (Development)</summary>
                <pre className="error-stack">
                  {this.state.error?.stack}
                  {'\n\n'}
                  {this.state.errorInfo.componentStack}
                </pre>
              </details>
            )}

            <div className="error-actions">
              <button 
                className="error-button primary"
                onClick={this.handleReload}
              >
                Reload App
              </button>
              
              <button 
                className="error-button secondary"
                onClick={this.handleRestart}
              >
                Restart App
              </button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}