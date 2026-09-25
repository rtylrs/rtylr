import React from 'react';
import { useTranslation } from 'react-i18next';

interface StepContainerProps {
  title: string;
  description: string;
  children: React.ReactNode;
  onBack?: () => void;
  onNext: () => void;
  isFirst?: boolean;
  isLast?: boolean;
  isSubmitting?: boolean;
}

const StepContainer = ({
  title,
  description,
  children,
  onBack,
  onNext,
  isFirst = false,
  isLast = false,
  isSubmitting = false,
}: StepContainerProps): React.JSX.Element => {
  const { t } = useTranslation();

  return (
    <div
      className="rounded-xl p-6 sm:p-8"
      style={{
        background: 'var(--hr-card)',
        border: '1px solid var(--hr-border)',
      }}
    >
      <div className="mb-6">
        <h2 className="text-lg font-semibold" style={{ color: 'var(--hr-text)' }}>
          {title}
        </h2>
        <p className="mt-1 text-sm" style={{ color: 'var(--hr-text-muted)' }}>
          {description}
        </p>
      </div>

      <div className="mb-8">{children}</div>

      <div className="flex items-center justify-between">
        {!isFirst && onBack ? (
          <button
            type="button"
            onClick={onBack}
            className="rounded-lg px-4 py-2 text-sm font-medium transition-colors"
            style={{
              background: 'var(--hr-hover)',
              color: 'var(--hr-text-muted)',
              border: '1px solid var(--hr-border)',
            }}
          >
            {t('common.back')}
          </button>
        ) : (
          <div />
        )}
        <button
          type="button"
          onClick={onNext}
          disabled={isSubmitting}
          className="rounded-lg px-5 py-2 text-sm font-medium transition-opacity disabled:opacity-50"
          style={{ background: 'var(--accent)', color: 'var(--hr-on-accent)' }}
        >
          {isSubmitting ? t('common.loading') : isLast ? t('setup.finishSetup') : t('common.next')}
        </button>
      </div>
    </div>
  );
};

export default StepContainer;
