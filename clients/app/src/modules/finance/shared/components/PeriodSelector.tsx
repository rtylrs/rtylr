import React from 'react';
import { useTranslation } from 'react-i18next';
import { ACCENT } from '../lib/tokens';

export type Period = 'month' | 'quarter' | 'year';

interface PeriodSelectorProps {
  period: Period;
  onPeriodChange: (p: Period) => void;
}

const OPTIONS: Period[] = ['month', 'quarter', 'year'];

const PeriodSelector: React.FC<PeriodSelectorProps> = ({ period, onPeriodChange }) => {
  const { t } = useTranslation('financeCommon');
  return (
    <div
      role="tablist"
      aria-label={t('finance.cashFlow.periodLabel')}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 4,
        padding: 4,
        background: 'var(--hub-surface)',
        border: '1px solid var(--hub-border)',
        borderRadius: 8,
      }}
    >
      {OPTIONS.map((p) => (
        <button
          type="button"
          role="tab"
          aria-selected={period === p}
          key={p}
          onClick={() => onPeriodChange(p)}
          style={{
            padding: '5px 12px',
            borderRadius: 6,
            border: 'none',
            fontSize: 12,
            fontWeight: 500,
            cursor: 'pointer',
            background: period === p ? ACCENT : 'transparent',
            color: period === p ? 'var(--finance-on-accent)' : 'var(--hub-text-muted)',
          }}
        >
          {t(`finance.cashFlow.periods.${p}`)}
        </button>
      ))}
    </div>
  );
};

export default PeriodSelector;
