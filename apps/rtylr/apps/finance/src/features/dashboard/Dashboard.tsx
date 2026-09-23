import React from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import {
  DollarSign,
  TrendingDown,
  TrendingUp,
  FileText,
  AlertTriangle,
  ArrowUpRight,
  ArrowDownRight,
} from 'lucide-react';
import { BarDataPoint, Transaction, OverdueInvoice } from '../../shared/services/financeApi';
import { useFinanceDashboard } from './useDashboard';
import FinanceKpiWidgets from '../../shared/components/FinanceKpiWidgets';
import OverdueInvoicesWidget from '../../shared/components/OverdueInvoicesWidget';
import { ErrorBoundary } from '@voxire/ui';
import { formatMoney, type Currency } from '../../shared/lib/money';
import { ghostBtn } from '../../shared/lib/tokens';

const fmt = (n: number, currency = 'USD') => formatMoney(n, currency as Currency, 0);

// ── Components ───────────────────────────────────────────────────
const cardStyle: React.CSSProperties = {
  background: 'var(--hub-card)',
  border: '1px solid var(--hub-border)',
  borderRadius: 12,
  padding: 20,
};

interface StatCardProps {
  /** Stable id, used as the React key: the label is a translation and changes with the language. */
  id: string;
  label: string;
  value: string;
  change: string;
  up: boolean | null;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  color: string;
}
const StatCard: React.FC<StatCardProps> = ({ label, value, change, up, icon: Icon, color }) => (
  <div style={cardStyle}>
    <div
      style={{
        display: 'flex',
        alignItems: 'flex-start',
        justifyContent: 'space-between',
        marginBottom: 12,
      }}
    >
      <span style={{ fontSize: 12, color: 'var(--hub-text-muted)', fontWeight: 500 }}>{label}</span>
      <span
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          width: 32,
          height: 32,
          borderRadius: 8,
          background: `${color}18`,
        }}
      >
        <Icon size={15} style={{ color }} />
      </span>
    </div>
    <div style={{ fontSize: 24, fontWeight: 700, color: 'var(--hub-text)', marginBottom: 6 }}>
      {value}
    </div>
    <div style={{ display: 'flex', alignItems: 'center', gap: 4, fontSize: 12 }}>
      {up !== null &&
        (up ? (
          <ArrowUpRight size={13} style={{ color: 'var(--finance-accent)' }} />
        ) : (
          <ArrowDownRight size={13} style={{ color: 'var(--finance-danger)' }} />
        ))}
      <span
        style={{
          color:
            up === null
              ? 'var(--hub-text-muted)'
              : up
                ? 'var(--finance-accent)'
                : 'var(--finance-danger)',
          fontWeight: 500,
        }}
      >
        {change}
      </span>
      {up !== null && <span style={{ color: 'var(--hub-text-dim)' }}>vs last month</span>}
    </div>
  </div>
);

const BarChart: React.FC<{ barData: BarDataPoint[] }> = ({ barData }) => {
  const { t } = useTranslation();
  const chartH = 140;
  const barW = 28;
  const gap = 16;
  const groupW = barW * 2 + gap;
  const groupGap = 24;
  const maxVal = Math.max(...barData.flatMap((d) => [d.revenue, d.expenses]), 1);
  const totalW = barData.length * (groupW + groupGap) - groupGap;

  return (
    <svg viewBox={`0 0 ${totalW + 40} ${chartH + 32}`} width="100%" style={{ overflow: 'visible' }}>
      {/* Grid lines */}
      {[0, 0.25, 0.5, 0.75, 1].map((f) => (
        <line
          key={f}
          x1={0}
          y1={chartH * (1 - f)}
          x2={totalW + 40}
          y2={chartH * (1 - f)}
          stroke="var(--hub-border)"
          strokeWidth={1}
        />
      ))}

      {barData.map((d, i) => {
        const x = i * (groupW + groupGap);
        const revH = (d.revenue / maxVal) * chartH;
        const expH = (d.expenses / maxVal) * chartH;
        return (
          <g key={d.month}>
            {/* Revenue bar */}
            <rect
              x={x}
              y={chartH - revH}
              width={barW}
              height={revH}
              rx={4}
              fill="var(--finance-accent)"
              opacity={0.85}
            />
            {/* Expenses bar */}
            <rect
              x={x + barW + gap}
              y={chartH - expH}
              width={barW}
              height={expH}
              rx={4}
              fill="var(--finance-danger)"
              opacity={0.7}
            />
            {/* Month label */}
            <text
              x={x + groupW / 2}
              y={chartH + 18}
              textAnchor="middle"
              fontSize={10}
              fill="var(--hub-text-dim)"
              fontFamily="inherit"
            >
              {d.month}
            </text>
          </g>
        );
      })}

      {/* Legend */}
      <g transform={`translate(${totalW - 100}, ${chartH + 28})`}>
        <rect
          x={0}
          y={-6}
          width={10}
          height={10}
          rx={2}
          fill="var(--finance-accent)"
          opacity={0.85}
        />
        <text x={14} y={4} fontSize={9} fill="var(--hub-text-muted)" fontFamily="inherit">
          {t('finance.dashboard.revenue')}
        </text>
        <rect
          x={60}
          y={-6}
          width={10}
          height={10}
          rx={2}
          fill="var(--finance-danger)"
          opacity={0.7}
        />
        <text x={74} y={4} fontSize={9} fill="var(--hub-text-muted)" fontFamily="inherit">
          {t('finance.dashboard.expenses')}
        </text>
      </g>
    </svg>
  );
};

const Dashboard: React.FC = () => {
  const { t } = useTranslation();
  const { data: dashData, isPending, isError, isFetching, refetch } = useFinanceDashboard();
  const loading = isPending || (!dashData && !isError);

  const expensesMtd = dashData?.totalExpensesMtd ?? 0;
  const netProfitMtd = dashData?.netProfitMtd ?? 0;
  const revenueMtd = dashData?.totalRevenueMtd ?? 0;

  const statsItems: StatCardProps[] = [
    {
      id: 'totalRevenueMtd',

      label: t('finance.dashboard.totalRevenueMtd'),
      value: loading ? '…' : fmt(revenueMtd),
      change: dashData?.totalRevenueMtdChange ?? '+0%',
      up: true,
      icon: DollarSign,
      color: 'var(--finance-accent)',
    },
    {
      id: 'totalExpensesMtd',

      label: t('finance.dashboard.totalExpensesMtd'),
      value: loading ? '…' : fmt(expensesMtd),
      change: dashData?.totalExpensesMtdChange ?? '+0%',
      up: false,
      icon: TrendingDown,
      color: 'var(--finance-danger)',
    },
    {
      id: 'netProfitMtd',

      label: t('finance.dashboard.netProfitMtd'),
      value: loading ? '…' : fmt(netProfitMtd),
      change: dashData?.netProfitChange ?? '+0%',
      up: true,
      icon: TrendingUp,
      color: 'var(--finance-accent)',
    },
    {
      id: 'outstandingInvoices',

      label: t('finance.dashboard.outstandingInvoices'),
      value: loading ? '…' : fmt(dashData?.totalReceivable ?? 0),
      change: dashData?.outstandingInvoicesCount
        ? `${dashData.outstandingInvoicesCount} invoices`
        : `${fmt(dashData?.overdueAmount ?? 0)} overdue`,
      up: null,
      icon: FileText,
      color: 'var(--finance-warning)',
    },
  ];

  const barData: BarDataPoint[] = dashData?.barData ?? [];
  const transactions: Transaction[] = dashData?.recentTransactions ?? [];
  const overdueInvoices: OverdueInvoice[] = dashData?.overdueInvoices ?? [];

  if (isError) {
    return (
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.25 }}
        style={{ display: 'flex', flexDirection: 'column', gap: 24 }}
      >
        <HelmetContainer title={pageTitle(t('nav.dashboard'))} />
        <div
          role="alert"
          style={{ ...cardStyle, padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <AlertTriangle size={16} style={{ color: 'var(--finance-danger)' }} />
            <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
              {t('errors.generic')}
            </span>
          </div>
          <button
            type="button"
            style={ghostBtn}
            disabled={isFetching}
            onClick={() => void refetch()}
          >
            {t('common.refresh')}
          </button>
        </div>
      </motion.div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.25 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 24 }}
    >
      <HelmetContainer title={pageTitle(t('nav.dashboard'))} />
      {/* Overdue alert */}
      {overdueInvoices.length > 0 && (
        <div
          style={{
            display: 'flex',
            alignItems: 'flex-start',
            gap: 12,
            padding: '14px 16px',
            borderRadius: 10,
            background: 'var(--finance-danger-subtle)',
            border: '1px solid var(--finance-danger-strong-border)',
          }}
        >
          <AlertTriangle
            size={16}
            style={{ color: 'var(--finance-danger)', flexShrink: 0, marginTop: 1 }}
          />
          <div>
            <p
              style={{
                margin: 0,
                fontSize: 13,
                fontWeight: 600,
                color: 'var(--finance-danger)',
                marginBottom: 4,
              }}
            >
              {overdueInvoices.length} Overdue Invoice{overdueInvoices.length > 1 ? 's' : ''}
            </p>
            <p style={{ margin: 0, fontSize: 12, color: 'var(--hub-text-muted)' }}>
              {overdueInvoices
                .map(
                  (inv) => `${inv.id} – ${inv.client} (${inv.amount}, ${inv.daysOverdue}d overdue)`
                )
                .join(' · ')}
            </p>
          </div>
        </div>
      )}

      {/* Stats grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 16 }}>
        {statsItems.map((s) => (
          <StatCard key={s.id} {...s} />
        ))}
      </div>

      {/* KPI Widgets */}
      <ErrorBoundary>
        <FinanceKpiWidgets />
      </ErrorBoundary>

      {/* Overdue widget + chart grid */}
      <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 16 }}>
        <ErrorBoundary>
          {barData.length > 0 ? (
            <div style={{ ...cardStyle, padding: 20 }}>
              <h2
                style={{
                  margin: '0 0 20px',
                  fontSize: 14,
                  fontWeight: 600,
                  color: 'var(--hub-text)',
                }}
              >
                {t('finance.dashboard.revenueVsExpenses')}
              </h2>
              <BarChart barData={barData} />
            </div>
          ) : (
            <div
              style={{
                ...cardStyle,
                padding: 20,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'var(--hub-text-dim)',
                fontSize: 13,
                minHeight: 180,
              }}
            >
              {loading ? 'Loading chart...' : 'No chart data available'}
            </div>
          )}
        </ErrorBoundary>
        <ErrorBoundary>
          <OverdueInvoicesWidget
            items={overdueInvoices.map((i) => ({
              id: i.id,
              client: i.client,
              amount: i.amount,
              daysOverdue: i.daysOverdue,
            }))}
          />
        </ErrorBoundary>
      </div>

      {/* Recent transactions */}
      <ErrorBoundary>
        <div style={cardStyle}>
          <h2
            style={{ margin: '0 0 16px', fontSize: 14, fontWeight: 600, color: 'var(--hub-text)' }}
          >
            {t('finance.dashboard.recentTransactions')}
          </h2>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 0 }}>
            {transactions.map((t, i) => (
              <div
                key={t.id}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '10px 0',
                  borderBottom:
                    i < transactions.length - 1 ? '1px solid var(--hub-border)' : 'none',
                }}
              >
                <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                  <span style={{ fontSize: 13, color: 'var(--hub-text)', fontWeight: 500 }}>
                    {t.description}
                  </span>
                  <span style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>
                    {t.id} · {t.date}
                  </span>
                </div>
                <span
                  style={{
                    fontSize: 14,
                    fontWeight: 600,
                    color: t.type === 'credit' ? 'var(--finance-accent)' : 'var(--finance-danger)',
                    fontVariantNumeric: 'tabular-nums',
                  }}
                >
                  {t.amount}
                </span>
              </div>
            ))}
          </div>
        </div>
      </ErrorBoundary>
    </motion.div>
  );
};

export default Dashboard;
