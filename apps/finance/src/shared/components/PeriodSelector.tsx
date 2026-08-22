import React from 'react';
import { ACCENT } from '../lib/tokens';

export type Period = 'month' | 'quarter' | 'year';

interface PeriodSelectorProps {
  period: Period;
  onPeriodChange: (p: Period) => void;
}

const PeriodSelector: React.FC<PeriodSelectorProps> = ({ period, onPeriodChange }) => {
  const options: Period[] = ['month', 'quarter', 'year'];
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
      <div
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
        {options.map((p) => (
          <button
            type="button"
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
              color: period === p ? 'var(--finance-white)' : 'var(--hub-text-muted)',
              textTransform: 'capitalize',
            }}
          >
            {p}
          </button>
        ))}
      </div>
    </div>
  );
};

export default PeriodSelector;
