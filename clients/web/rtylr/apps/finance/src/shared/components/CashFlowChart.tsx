import React from 'react';
import { useTranslation } from 'react-i18next';
import { ACCENT } from '../lib/tokens';

export interface CashFlowPoint {
  month: string;
  inflow: number;
  outflow: number;
}

interface CashFlowChartProps {
  data: CashFlowPoint[];
  height?: number;
}

const CashFlowChart: React.FC<CashFlowChartProps> = ({ data, height = 200 }) => {
  const { t } = useTranslation();
  const max = Math.max(...data.flatMap((d) => [d.inflow, d.outflow]), 1);
  const width = Math.max(data.length * 70, 320);
  const pad = 30;
  const chartH = height - pad * 2;
  const step = (width - pad * 2) / Math.max(data.length - 1, 1);

  const inflowPath = data
    .map(
      (d, i) =>
        `${i === 0 ? 'M' : 'L'} ${pad + i * step},${pad + chartH - (d.inflow / max) * chartH}`
    )
    .join(' ');

  const outflowPath = data
    .map(
      (d, i) =>
        `${i === 0 ? 'M' : 'L'} ${pad + i * step},${pad + chartH - (d.outflow / max) * chartH}`
    )
    .join(' ');

  return (
    <svg viewBox={`0 0 ${width} ${height}`} width="100%" style={{ overflow: 'visible' }}>
      {[0, 0.25, 0.5, 0.75, 1].map((f) => (
        <line
          key={f}
          x1={pad}
          y1={pad + chartH * f}
          x2={width - pad}
          y2={pad + chartH * f}
          stroke="var(--hub-border)"
          strokeDasharray="3 3"
        />
      ))}
      <path d={inflowPath} stroke={ACCENT} strokeWidth={2} fill="none" />
      <path d={outflowPath} stroke="var(--finance-danger)" strokeWidth={2} fill="none" />
      {data.map((d, i) => {
        const x = pad + i * step;
        const y1 = pad + chartH - (d.inflow / max) * chartH;
        const y2 = pad + chartH - (d.outflow / max) * chartH;
        return (
          <g key={i}>
            <circle cx={x} cy={y1} r={4} fill={ACCENT} />
            <circle cx={x} cy={y2} r={4} fill="var(--finance-danger)" />
            <text x={x} y={height - 6} textAnchor="middle" fontSize={10} fill="var(--hub-text-dim)">
              {d.month}
            </text>
          </g>
        );
      })}
      <g>
        <circle cx={pad} cy={pad - 12} r={4} fill={ACCENT} />
        <text x={pad + 8} y={pad - 8} fontSize={10} fill="var(--hub-text-muted)">
          {t('finance.cashFlow.inflow')}
        </text>
        <circle cx={pad + 64} cy={pad - 12} r={4} fill="var(--finance-danger)" />
        <text x={pad + 72} y={pad - 8} fontSize={10} fill="var(--hub-text-muted)">
          {t('finance.cashFlow.outflow')}
        </text>
      </g>
    </svg>
  );
};

export default CashFlowChart;
