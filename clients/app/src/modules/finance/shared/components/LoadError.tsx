import React from 'react';
import { useTranslation } from 'react-i18next';
import { AlertTriangle, RotateCcw } from 'lucide-react';
import { cardStyle, ghostBtn } from '../lib/tokens';

interface LoadErrorProps {
  title: string;
  message: string;
  onRetry: () => void;
  retrying?: boolean;
}

/**
 * A request for the page's data failed. It names what failed in a heading, so
 * the failure reads as one instead of passing for an empty or half-drawn page.
 */
const LoadError: React.FC<LoadErrorProps> = ({ title, message, onRetry, retrying }) => {
  const { t } = useTranslation('financeCommon');
  return (
    <div
      role="alert"
      style={{ ...cardStyle, padding: 20, display: 'flex', flexDirection: 'column', gap: 10 }}
    >
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <AlertTriangle size={16} style={{ color: 'var(--finance-danger)', flexShrink: 0 }} />
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 600, color: 'var(--hub-text)' }}>
          {title}
        </h3>
      </div>
      <p style={{ margin: 0, fontSize: 13, color: 'var(--hub-text-muted)', lineHeight: 1.5 }}>
        {message}
      </p>
      <button
        type="button"
        style={{ ...ghostBtn, alignSelf: 'flex-start' }}
        disabled={retrying}
        onClick={onRetry}
      >
        <RotateCcw size={13} /> {t('financeCommon.refresh')}
      </button>
    </div>
  );
};

export default LoadError;
