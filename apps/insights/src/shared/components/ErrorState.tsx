import React from 'react';
import { AlertTriangle } from 'lucide-react';

interface ErrorStateProps {
  message?: string;
  onRetry?: () => void;
}

const ErrorState: React.FC<ErrorStateProps> = ({
  message = 'Something went wrong while loading data.',
  onRetry,
}) => (
  <div className="flex flex-col items-center justify-center py-16 gap-3">
    <AlertTriangle size={28} style={{ color: '#f87171' }} />
    <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
      {message}
    </p>
    {onRetry && (
      <button
        onClick={onRetry}
        className="text-xs px-3 py-1.5 rounded-lg font-medium"
        style={{ background: 'var(--hub-hover)', color: 'var(--hub-text)' }}
      >
        Try again
      </button>
    )}
  </div>
);

export default ErrorState;
