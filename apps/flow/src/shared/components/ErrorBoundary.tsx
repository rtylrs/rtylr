import React from 'react';

interface State {
  hasError: boolean;
  message: string;
}

export class ErrorBoundary extends React.Component<{ children: React.ReactNode }, State> {
  state: State = { hasError: false, message: '' };

  static getDerivedStateFromError(error: unknown): State {
    const message = error instanceof Error ? error.message : String(error);
    return { hasError: true, message };
  }

  componentDidCatch(error: unknown, info: React.ErrorInfo) {
    console.error('[Flow ErrorBoundary]', error, info.componentStack);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            height: '100%',
            gap: 12,
            padding: 24,
          }}
        >
          <p style={{ fontWeight: 600, fontSize: 14, color: 'var(--flow-text, #111)' }}>
            Something went wrong
          </p>
          <p
            style={{
              fontSize: 12,
              color: 'var(--flow-text-dim, #999)',
              maxWidth: 360,
              textAlign: 'center',
            }}
          >
            {this.state.message || 'An unexpected error occurred.'}
          </p>
          <button
            onClick={() => this.setState({ hasError: false, message: '' })}
            style={{
              marginTop: 8,
              padding: '6px 16px',
              borderRadius: 8,
              border: '1px solid var(--flow-border, #e5e5e5)',
              background: 'transparent',
              color: 'var(--flow-text-muted, #555)',
              fontSize: 12,
              cursor: 'pointer',
            }}
          >
            Try again
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}

export default ErrorBoundary;
