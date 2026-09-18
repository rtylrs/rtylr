import React from 'react';

interface ErrorBoundaryProps {
  children: React.ReactNode;
  /**
   * Custom fallback. Receives the captured error and a reset callback that
   * clears the boundary's error state so the children can re-mount.
   */
  fallback?: React.ReactNode | ((args: { error: Error; reset: () => void }) => React.ReactNode);
  /**
   * Side-effect hook for telemetry (Sentry, PostHog, etc.). Called once per
   * caught error.
   */
  onError?: (error: Error, info: React.ErrorInfo) => void;
}

interface ErrorBoundaryState {
  error: Error | null;
}

const DYNAMIC_IMPORT_RELOAD_KEY = 'rtylr:dynamic-import-reload';

export function isDynamicImportError(error: unknown): boolean {
  const message = error instanceof Error ? error.message : String(error ?? '');

  return /Failed to fetch dynamically imported module|Importing a module script failed|Loading chunk \d+ failed|ChunkLoadError|error loading dynamically imported module/i.test(
    message
  );
}

export function recoverFromDynamicImportError(error: unknown): boolean {
  if (!isDynamicImportError(error) || typeof window === 'undefined') return false;

  const reloadKey = `${window.location.origin}${window.location.pathname}${window.location.search}`;

  try {
    if (window.sessionStorage.getItem(DYNAMIC_IMPORT_RELOAD_KEY) === reloadKey) return false;
    window.sessionStorage.setItem(DYNAMIC_IMPORT_RELOAD_KEY, reloadKey);
  } catch {
    return false;
  }

  window.location.reload();
  return true;
}

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { error: null };

  static getDerivedStateFromError(error: unknown): ErrorBoundaryState {
    return { error: error instanceof Error ? error : new Error(String(error)) };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo): void {
    this.props.onError?.(error, info);
    recoverFromDynamicImportError(error);
  }

  reset = (): void => {
    this.setState({ error: null });
  };

  render(): React.ReactNode {
    const { error } = this.state;
    if (!error) return this.props.children;

    if (typeof this.props.fallback === 'function') {
      return this.props.fallback({ error, reset: this.reset });
    }
    if (this.props.fallback !== undefined) return this.props.fallback;

    return <DefaultErrorFallback error={error} reset={this.reset} />;
  }
}

interface FallbackProps {
  error: Error;
  reset: () => void;
}

export function DefaultErrorFallback({ error, reset }: FallbackProps): React.JSX.Element {
  return (
    <div
      role="alert"
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '60vh',
        gap: 12,
        padding: 24,
        textAlign: 'center',
      }}
    >
      <p style={{ fontWeight: 600, fontSize: 16, color: 'var(--text, #111)' }}>
        Something went wrong
      </p>
      <p
        style={{
          fontSize: 13,
          color: 'var(--text-muted, #6b7280)',
          maxWidth: 420,
        }}
      >
        {error.message || 'An unexpected error occurred. Try again or refresh the page.'}
      </p>
      <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
        <button
          onClick={reset}
          style={{
            padding: '8px 16px',
            borderRadius: 8,
            border: '1px solid var(--border, #e5e7eb)',
            background: 'transparent',
            color: 'var(--text, #111)',
            fontSize: 13,
            cursor: 'pointer',
          }}
        >
          Try again
        </button>
        <button
          onClick={() => window.location.reload()}
          style={{
            padding: '8px 16px',
            borderRadius: 8,
            border: 'none',
            background: 'var(--accent, #111)',
            color: 'var(--accent-foreground, #fff)',
            fontSize: 13,
            cursor: 'pointer',
          }}
        >
          Reload
        </button>
      </div>
    </div>
  );
}

export function WidgetErrorFallback({ error, reset }: FallbackProps): React.JSX.Element {
  return (
    <div
      role="alert"
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 6,
        padding: 16,
        minHeight: 120,
        border: '1px dashed var(--border, #e5e7eb)',
        borderRadius: 8,
        textAlign: 'center',
      }}
    >
      <p style={{ fontSize: 12, color: 'var(--text-muted, #6b7280)' }}>
        {error.message || 'Widget failed to load'}
      </p>
      <button
        onClick={reset}
        style={{
          fontSize: 11,
          color: 'var(--text, #111)',
          background: 'transparent',
          border: 'none',
          textDecoration: 'underline',
          cursor: 'pointer',
        }}
      >
        Retry
      </button>
    </div>
  );
}

export default ErrorBoundary;
