import React, { useState, useEffect } from 'react';
import { motion } from 'motion/react';

const MONTHS = ['Nov', 'Dec', 'Jan', 'Feb', 'Mar', 'Apr', 'May'];
const REVENUE = [31200, 44800, 28400, 33600, 38900, 41200, 42840];
const EXPENSES = [21800, 30100, 20200, 23400, 26100, 27800, 28310];

const fmt = (n: number): string => (n >= 1000 ? `$${(n / 1000).toFixed(1)}k` : `$${n}`);

interface Metric {
  label: string;
  value: number;
  prev: number;
  prefix: string;
}

const METRICS: Metric[] = [
  { label: 'Revenue', value: 42840, prev: 41200, prefix: '$' },
  { label: 'Expenses', value: 28310, prev: 27800, prefix: '$' },
  { label: 'Net profit', value: 14530, prev: 13400, prefix: '$' },
];

const pctChange = (cur: number, prev: number): string => {
  const d = ((cur - prev) / prev) * 100;
  return `${d >= 0 ? '+' : ''}${d.toFixed(1)}%`;
};

const maxVal = Math.max(...REVENUE);

export const FinanceDemo = ({ isActive }: { isActive: boolean }): React.JSX.Element => {
  const [shown, setShown] = useState(false);

  useEffect(() => {
    if (isActive) {
      const t = setTimeout(() => setShown(true), 100);
      return () => clearTimeout(t);
    } else {
      setShown(false);
    }
  }, [isActive]);

  return (
    <div
      style={{
        position: 'absolute',
        inset: 0,
        background: 'var(--bg-surface-a)',
        fontFamily: 'inherit',
        fontSize: 13,
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
        padding: '20px',
      }}
    >
      {/* Header */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: 20,
        }}
      >
        <div>
          <span style={{ color: 'var(--ink)', fontWeight: 600, fontSize: 14 }}>P&amp;L</span>
          <span style={{ color: 'rgba(19,19,19,0.5)', fontSize: 12, marginLeft: 10 }}>
            May 2026
          </span>
        </div>
        <span
          style={{
            background: 'rgba(34,197,94,0.12)',
            color: '#16a34a',
            borderRadius: 999,
            padding: '3px 10px',
            fontSize: 11,
            fontWeight: 600,
          }}
        >
          Profitable
        </span>
      </div>

      {/* Metric tiles */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(3, 1fr)',
          gap: 10,
          marginBottom: 24,
        }}
      >
        {METRICS.map((m) => (
          <div
            key={m.label}
            style={{
              background: 'var(--bg-surface-b)',
              border: '1px solid var(--hairline)',
              borderRadius: 10,
              padding: '12px 14px',
            }}
          >
            <div style={{ fontSize: 10, color: 'rgba(19,19,19,0.55)', marginBottom: 6 }}>
              {m.label}
            </div>
            <div style={{ fontSize: 18, fontWeight: 700, color: 'var(--ink)', marginBottom: 4 }}>
              {shown ? `$${(m.value / 1000).toFixed(1)}k` : '-'}
            </div>
            <div style={{ fontSize: 11, color: '#16a34a' }}>{pctChange(m.value, m.prev)} MoM</div>
          </div>
        ))}
      </div>

      {/* Bar chart */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
        <div
          style={{
            fontSize: 11,
            color: 'rgba(19,19,19,0.5)',
            marginBottom: 12,
            display: 'flex',
            gap: 16,
          }}
        >
          <span style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
            <span
              style={{
                width: 8,
                height: 8,
                borderRadius: 2,
                background: 'var(--ink)',
                display: 'inline-block',
              }}
            />
            Revenue
          </span>
          <span style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
            <span
              style={{
                width: 8,
                height: 8,
                borderRadius: 2,
                background: 'rgba(19,19,19,0.25)',
                display: 'inline-block',
              }}
            />
            Expenses
          </span>
        </div>
        <div style={{ flex: 1, display: 'flex', alignItems: 'flex-end', gap: 8, minHeight: 0 }}>
          {MONTHS.map((month, i) => (
            <div
              key={month}
              style={{
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: 4,
                height: '100%',
                justifyContent: 'flex-end',
              }}
            >
              <div
                style={{
                  width: '100%',
                  display: 'flex',
                  gap: 3,
                  alignItems: 'flex-end',
                  flex: 1,
                  minHeight: 0,
                }}
              >
                <motion.div
                  animate={{ height: shown ? `${(REVENUE[i] / maxVal) * 100}%` : '0%' }}
                  transition={{ duration: 0.7, delay: i * 0.05, ease: 'easeOut' }}
                  style={{
                    flex: 1,
                    background: i === MONTHS.length - 1 ? 'var(--ink)' : 'rgba(19,19,19,0.3)',
                    borderRadius: '3px 3px 0 0',
                    minHeight: 2,
                  }}
                />
                <motion.div
                  animate={{ height: shown ? `${(EXPENSES[i] / maxVal) * 100}%` : '0%' }}
                  transition={{ duration: 0.7, delay: i * 0.05, ease: 'easeOut' }}
                  style={{
                    flex: 1,
                    background:
                      i === MONTHS.length - 1 ? 'rgba(19,19,19,0.35)' : 'rgba(19,19,19,0.12)',
                    borderRadius: '3px 3px 0 0',
                    minHeight: 2,
                  }}
                />
              </div>
              <div style={{ fontSize: 9, color: 'rgba(19,19,19,0.45)', whiteSpace: 'nowrap' }}>
                {month}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
