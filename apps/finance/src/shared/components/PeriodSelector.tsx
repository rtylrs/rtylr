import React from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { ACCENT } from '../lib/tokens';

export type Period = 'month' | 'quarter' | 'year';

interface PeriodSelectorProps {
  period: Period;
  onPeriodChange: (p: Period) => void;
  date: Date;
  onDateChange: (d: Date) => void;
}

function shift(d: Date, period: Period, dir: 1 | -1): Date {
  const n = new Date(d);
  if (period === 'month') n.setMonth(n.getMonth() + dir);
  if (period === 'quarter') n.setMonth(n.getMonth() + dir * 3);
  if (period === 'year') n.setFullYear(n.getFullYear() + dir);
  return n;
}

function formatLabel(d: Date, period: Period): string {
  if (period === 'month') return d.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });
  if (period === 'quarter') return `Q${Math.floor(d.getMonth() / 3) + 1} ${d.getFullYear()}`;
  return `${d.getFullYear()}`;
}

const PeriodSelector: React.FC<PeriodSelectorProps> = ({
  period,
  onPeriodChange,
  date,
  onDateChange,
}) => {
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
        <button
          onClick={() => onDateChange(shift(date, period, -1))}
          style={{
            padding: 4,
            borderRadius: 6,
            border: 'none',
            background: 'transparent',
            color: 'var(--hub-text-muted)',
            cursor: 'pointer',
          }}
        >
          <ChevronLeft size={14} />
        </button>
        <span
          style={{
            fontSize: 12,
            fontWeight: 600,
            color: 'var(--hub-text)',
            minWidth: 120,
            textAlign: 'center',
          }}
        >
          {formatLabel(date, period)}
        </span>
        <button
          onClick={() => onDateChange(shift(date, period, 1))}
          style={{
            padding: 4,
            borderRadius: 6,
            border: 'none',
            background: 'transparent',
            color: 'var(--hub-text-muted)',
            cursor: 'pointer',
          }}
        >
          <ChevronRight size={14} />
        </button>
      </div>
    </div>
  );
};

export default PeriodSelector;
