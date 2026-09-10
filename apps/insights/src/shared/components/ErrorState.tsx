import React from 'react';
import { AlertTriangle } from 'lucide-react';
import { useTranslation } from 'react-i18next';

interface ErrorStateProps {
  message?: string;
  onRetry?: () => void;
}

const ErrorState: React.FC<ErrorStateProps> = ({ message, onRetry }) => {
  const { t } = useTranslation();
  const defaultMessage = t('insights.errors.generic');

  return (
    <div
      role="alert"
      aria-live="assertive"
      className="flex flex-col items-center justify-center py-16 gap-3"
    >
      <AlertTriangle size={28} style={{ color: '#f87171' }} />
      <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
        {message || defaultMessage}
      </p>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="text-xs px-4 py-2 min-h-[44px] rounded-lg font-medium"
          style={{ background: 'var(--hub-hover)', color: 'var(--hub-text)' }}
        >
          {t('common.tryAgain')}
        </button>
      )}
    </div>
  );
};

export default ErrorState;
