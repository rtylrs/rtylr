import React from 'react';
import { useTranslation } from 'react-i18next';

interface ErrorBoundaryState {
  hasError: boolean;
}

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

class ErrorBoundaryClass extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true };
  }

  render(): React.ReactNode {
    if (this.state.hasError) {
      return this.props.fallback ?? <DefaultErrorFallback />;
    }
    return this.props.children;
  }
}

function DefaultErrorFallback(): React.JSX.Element {
  const { t } = useTranslation();

  return (
    <div
      className="flex min-h-screen flex-col items-center justify-center gap-4"
      style={{ background: 'var(--hr-bg)', color: 'var(--hr-text)' }}
    >
      <p className="text-lg font-semibold">{t('errors.generic')}</p>
      <button
        onClick={() => window.location.reload()}
        className="rounded-lg px-4 py-2 text-sm font-medium"
        style={{ background: 'var(--accent)', color: '#fff' }}
      >
        {t('common.close')}
      </button>
    </div>
  );
}

export default ErrorBoundaryClass;
