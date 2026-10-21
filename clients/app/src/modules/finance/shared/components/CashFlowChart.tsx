import React from 'react';
import { useTranslation } from 'react-i18next';
import { formatMoney, type Currency } from '../lib/money';

export interface CashFlowPoint {
  month: string;
  inflow: number;
  outflow: number;
}

interface CashFlowChartProps {
  data: (CashFlowPoint & { label: string; title: string })[];
  currency: Currency;
  height?: number;
}

/** Money in and out as paired bars, one pair per bucket. */
const CashFlowChart: React.FC<CashFlowChartProps> = ({ data, currency, height = 240 }) => {
  const { t } = useTranslation('financeCommon');
  const max = Math.max(...data.flatMap((d) => [d.inflow, d.outflow]), 1);
  const slot = 32;
  const barW = Math.max(Math.min(slot / 2 - 3, 14), 3);
  // One slot per bucket; the SVG stretches to the card so labels below line up.
  const width = Math.max(data.length, 1) * slot;
  const labelH = 22;
  const chartH = height - labelH;

  return (
    <div>
      <svg
        viewBox={`0 0 ${width} ${height}`}
        width="100%"
        role="img"
        aria-label={t('finance.cashFlow.chartLabel')}
        preserveAspectRatio="none"
        style={{ display: 'block', height }}
      >
        {[0, 0.5, 1].map((f) => (
          <line
            key={f}
            x1={0}
            y1={chartH * (1 - f) + 0.5}
            x2={width}
            y2={chartH * (1 - f) + 0.5}
            stroke="var(--hub-border)"
          />
        ))}
        {data.map((d, i) => {
          const x = i * slot + (slot - barW * 2 - 2) / 2;
          const inH = (d.inflow / max) * chartH;
          const outH = (d.outflow / max) * chartH;
          return (
            <g key={d.month}>
              <title>
                {`${d.title} · ${t('finance.cashFlow.inflow')} ${formatMoney(d.inflow, currency)} · ${t('finance.cashFlow.outflow')} ${formatMoney(d.outflow, currency)}`}
              </title>
              <rect
                x={x}
                y={chartH - inH}
                width={barW}
                height={inH}
                rx={2}
                fill="var(--finance-success)"
              />
              <rect
                x={x + barW + 2}
                y={chartH - outH}
                width={barW}
                height={outH}
                rx={2}
                fill="var(--finance-danger)"
                opacity={0.8}
              />
            </g>
          );
        })}
      </svg>
      <div style={{ display: 'flex', width: '100%', marginTop: 6 }}>
        {data.map((d) => (
          <span
            key={d.month}
            style={{
              flex: 1,
              textAlign: 'center',
              fontSize: 11,
              color: 'var(--hub-text-dim)',
              whiteSpace: 'nowrap',
              overflow: 'visible',
            }}
          >
            {d.label}
          </span>
        ))}
      </div>
      <div
        style={{
          display: 'flex',
          gap: 16,
          marginTop: 12,
          fontSize: 12,
          color: 'var(--hub-text-muted)',
        }}
      >
        {[
          ['var(--finance-success)', t('finance.cashFlow.inflow')],
          ['var(--finance-danger)', t('finance.cashFlow.outflow')],
        ].map(([color, name]) => (
          <span key={name} style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
            <span style={{ width: 10, height: 10, borderRadius: 2, background: color }} />
            {name}
          </span>
        ))}
      </div>
    </div>
  );
};

export default CashFlowChart;
