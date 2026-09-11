import React from 'react';
import { useTranslation } from 'react-i18next';
import { TrendingUp, TrendingDown, DollarSign, Clock } from 'lucide-react';
import MoneyDisplay from './MoneyDisplay';
import { cardStyle } from '../lib/tokens';
import { useFinanceDashboard } from '../hooks/queries/useFinanceDashboard';

interface Kpi {
  label: string;
  value: React.ReactNode;
  hint?: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  colorVar: string;
}

const FinanceKpiWidgets: React.FC = () => {
  const { t } = useTranslation();
  const { data, isLoading } = useFinanceDashboard();

  const dash = data as
    | {
        totalReceivable: number;
        overdueAmount: number;
        totalExpensesMtd: number;
        netProfitMtd: number;
      }
    | undefined;

  const kpis: Kpi[] = [
    {
      label: t('finance.kpi.receivable'),
      value:
        isLoading || !dash ? (
          '-'
        ) : (
          <MoneyDisplay amount={dash.totalReceivable} fractionDigits={0} />
        ),
      hint: 'outstanding',
      icon: DollarSign,
      colorVar: '--finance-receivable',
    },
    {
      label: t('finance.kpi.expensesMtd'),
      value:
        isLoading || !dash ? (
          '-'
        ) : (
          <MoneyDisplay amount={dash.totalExpensesMtd} fractionDigits={0} />
        ),
      hint: 'month to date',
      icon: TrendingDown,
      colorVar: '--finance-expenses',
    },
    {
      label: t('finance.kpi.netProfitMtd'),
      value:
        isLoading || !dash ? '-' : <MoneyDisplay amount={dash.netProfitMtd} fractionDigits={0} />,
      hint: 'month to date',
      icon: TrendingUp,
      colorVar: '--finance-profit',
    },
    {
      label: t('finance.kpi.overdue'),
      value:
        isLoading || !dash ? '-' : <MoneyDisplay amount={dash.overdueAmount} fractionDigits={0} />,
      hint: 'past due date',
      icon: Clock,
      colorVar: '--finance-overdue',
    },
  ];

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
      {kpis.map((k) => {
        const Icon = k.icon;
        const color = `var(${k.colorVar})`;
        return (
          <div key={k.label} style={{ ...cardStyle, padding: 18 }}>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                marginBottom: 10,
              }}
            >
              <span
                style={{
                  fontSize: 11,
                  fontWeight: 600,
                  color: 'var(--hub-text-dim)',
                  textTransform: 'uppercase',
                  letterSpacing: '0.08em',
                }}
              >
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
                  background: `color-mix(in srgb, var(${k.colorVar}) 12%, transparent)`,
                  color,
                }}
              >
                <Icon size={14} />
              </span>
            </div>
            <div style={{ fontSize: 22, fontWeight: 700, color: 'var(--hub-text)' }}>{k.value}</div>
            {k.hint && (
              <div style={{ fontSize: 11, color: 'var(--hub-text-dim)', marginTop: 4 }}>
                {k.hint}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
};

export default FinanceKpiWidgets;
