import React from 'react';
import { TrendingUp, TrendingDown, Calendar, Clock } from 'lucide-react';
import MoneyDisplay from './MoneyDisplay';
import { cardStyle, ACCENT } from '../lib/tokens';

interface Kpi {
  label: string;
  value: React.ReactNode;
  hint?: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  color: string;
}

// Default KPI values shown before real dashboard data arrives.
const DEFAULT_KPIS = {
  mrr: 82400,
  mrrChange: '+8.4%',
  burn: 64200,
  runway: 18,
  overdue: 12420,
  overdueCount: 3,
};

const FinanceKpiWidgets: React.FC<{ data?: Partial<typeof DEFAULT_KPIS> }> = ({ data }) => {
  const d = { ...DEFAULT_KPIS, ...data };
  const kpis: Kpi[] = [
    { label: 'MRR', value: <MoneyDisplay amount={d.mrr} fractionDigits={0} />, hint: d.mrrChange, icon: TrendingUp, color: ACCENT },
    { label: 'Burn', value: <MoneyDisplay amount={d.burn} fractionDigits={0} />, hint: 'per month', icon: TrendingDown, color: '#ef4444' },
    { label: 'Runway', value: `${d.runway} mo`, hint: 'at current burn', icon: Calendar, color: '#3b82f6' },
    { label: 'Overdue', value: <MoneyDisplay amount={d.overdue} fractionDigits={0} />, hint: `${d.overdueCount} invoices`, icon: Clock, color: '#f59e0b' },
  ];
  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
      {kpis.map((k) => {
        const Icon = k.icon;
        return (
          <div key={k.label} style={{ ...cardStyle, padding: 18 }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 10 }}>
              <span style={{ fontSize: 11, fontWeight: 600, color: 'var(--hub-text-dim)', textTransform: 'uppercase', letterSpacing: '0.08em' }}>
                {k.label}
              </span>
              <span
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  width: 28,
                  height: 28,
                  borderRadius: 8,
                  background: `${k.color}18`,
                  color: k.color,
                }}
              >
                <Icon size={14} />
              </span>
            </div>
            <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--hub-text)' }}>{k.value}</div>
            {k.hint && <div style={{ fontSize: 11, color: 'var(--hub-text-dim)', marginTop: 4 }}>{k.hint}</div>}
          </div>
        );
      })}
    </div>
  );
};

export default FinanceKpiWidgets;
