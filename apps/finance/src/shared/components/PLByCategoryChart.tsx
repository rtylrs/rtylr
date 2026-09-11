import React from 'react';
import { useTranslation } from 'react-i18next';
import MoneyDisplay from './MoneyDisplay';

interface CategoryRow {
  category: string;
  amount: number;
  color: string;
}

interface PLByCategoryChartProps {
  data: CategoryRow[];
  title?: string;
}

const PLByCategoryChart: React.FC<PLByCategoryChartProps> = ({ data, title }) => {
  const { t } = useTranslation();
  const max = Math.max(...data.map((d) => d.amount), 1);
  return (
    <div>
      <div
        style={{
          fontSize: 11,
          color: 'var(--hub-text-dim)',
          textTransform: 'uppercase',
          letterSpacing: '0.08em',
          marginBottom: 12,
        }}
      >
        {title ?? t('finance.plChart.title')}
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        {data.map((row) => {
          const pct = (row.amount / max) * 100;
          return (
            <div key={row.category}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  fontSize: 12,
                  color: 'var(--hub-text)',
                  marginBottom: 4,
                }}
              >
                <span style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <span style={{ width: 8, height: 8, borderRadius: 4, background: row.color }} />
                  {row.category}
                </span>
                <MoneyDisplay amount={row.amount} />
              </div>
              <div
                style={{
                  height: 6,
                  background: 'var(--hub-hover)',
                  borderRadius: 3,
                  overflow: 'hidden',
                }}
              >
                <div
                  style={{
                    width: `${pct}%`,
                    height: '100%',
                    background: row.color,
                    transition: 'width 0.3s',
                  }}
                />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default PLByCategoryChart;
