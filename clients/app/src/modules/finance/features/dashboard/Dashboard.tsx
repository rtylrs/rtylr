import React from 'react';
import { useTranslation } from 'react-i18next';
import { pageTitle } from '../../shared/lib/pageTitle';
import { HelmetContainer } from '@voxire/components';
import { Link } from '@tanstack/react-router';
import { motion } from 'motion/react';
import { AlertTriangle, ArrowDownLeft, ArrowUpRight, Clock } from 'lucide-react';
import { useFinanceDashboard } from './useDashboard';
import type { DashboardActivity, DashboardMonth } from '../../shared/schemas/dashboard';
import OverdueInvoicesWidget from '../../shared/components/OverdueInvoicesWidget';
import { ErrorBoundary } from '@voxire/ui';
import { formatMoney, type Currency } from '../../shared/lib/money';
import { formatDate, formatDateShort } from '../../shared/lib/date';
import { useSettings } from '../../shared/lib/settings';
import { cardStyle, ghostBtn } from '../../shared/lib/tokens';

const card: React.CSSProperties = { ...cardStyle, padding: 20 };

const StatCard: React.FC<{ label: string; value: string; hint?: string; tone?: string }> = ({
  label,
  value,
  hint,
  tone,
}) => (
  <div style={card}>
    <div style={{ fontSize: 12, color: 'var(--hub-text-muted)', fontWeight: 500 }}>{label}</div>
    <div
      style={{
        marginTop: 8,
        fontSize: 24,
        fontWeight: 700,
        color: tone ?? 'var(--hub-text)',
        fontVariantNumeric: 'tabular-nums',
      }}
    >
      {value}
    </div>
    {hint && <div style={{ marginTop: 4, fontSize: 12, color: 'var(--hub-text-dim)' }}>{hint}</div>}
  </div>
);

const Notice: React.FC<{
  tone: 'warning' | 'danger';
  icon: React.ReactNode;
  children: React.ReactNode;
  to: string;
  linkLabel: string;
}> = ({ tone, icon, children, to, linkLabel }) => (
  <div
    role="status"
    style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: 12,
      flexWrap: 'wrap',
      padding: '12px 16px',
      borderRadius: 10,
      background: `var(--finance-${tone}-soft)`,
      border: `1px solid var(--finance-${tone}-border)`,
      fontSize: 13,
      color: 'var(--hub-text)',
    }}
  >
    <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
      {icon}
      {children}
    </span>
    <Link to={to} style={{ fontWeight: 600, color: 'var(--finance-accent)' }}>
      {linkLabel}
    </Link>
  </div>
);

/** Revenue and expense bars for each month, with the month under each pair. */
const MonthlyChart: React.FC<{ months: DashboardMonth[]; currency: Currency }> = ({
  months,
  currency,
}) => {
  const { t } = useTranslation('financeCommon');
  const chartH = 140;
  const barW = 22;
  const pairGap = 6;
  const groupW = barW * 2 + pairGap;
  const groupGap = 28;
  const max = Math.max(...months.flatMap((m) => [m.revenue, m.expenses]), 1);
  const totalW = months.length * (groupW + groupGap) - groupGap;
  const label = (key: string): string => {
    const [y, m] = key.split('-').map(Number);
    return formatDate(new Date(y, m - 1, 1), { month: 'short' });
  };
  return (
    <div>
      <svg
        viewBox={`0 0 ${totalW} ${chartH + 24}`}
        width="100%"
        role="img"
        aria-label={t('finance.dashboard.revenueVsExpenses')}
      >
        {[0, 0.5, 1].map((f) => (
          <line
            key={f}
            x1={0}
            y1={chartH * (1 - f)}
            x2={totalW}
            y2={chartH * (1 - f)}
            stroke="var(--hub-border)"
            strokeWidth={1}
          />
        ))}
        {months.map((m, i) => {
          const x = i * (groupW + groupGap);
          const revH = (m.revenue / max) * chartH;
          const expH = (m.expenses / max) * chartH;
          return (
            <g key={m.month}>
              <rect
                x={x}
                y={chartH - revH}
                width={barW}
                height={revH}
                rx={4}
                fill="var(--finance-success)"
              >
                <title>{`${label(m.month)} · ${t('finance.dashboard.revenue')}: ${formatMoney(m.revenue, currency)}`}</title>
              </rect>
              <rect
                x={x + barW + pairGap}
                y={chartH - expH}
                width={barW}
                height={expH}
                rx={4}
                fill="var(--finance-danger)"
                opacity={0.75}
              >
                <title>{`${label(m.month)} · ${t('finance.dashboard.expenses')}: ${formatMoney(m.expenses, currency)}`}</title>
              </rect>
              <text
                x={x + groupW / 2}
                y={chartH + 16}
                textAnchor="middle"
                fontSize={10}
                fill="var(--hub-text-dim)"
                fontFamily="inherit"
              >
                {label(m.month)}
              </text>
            </g>
          );
        })}
      </svg>
      <div
        style={{
          display: 'flex',
          gap: 16,
          marginTop: 10,
          fontSize: 12,
          color: 'var(--hub-text-muted)',
        }}
      >
        {[
          ['var(--finance-success)', t('finance.dashboard.revenue')],
          ['var(--finance-danger)', t('finance.dashboard.expenses')],
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

const EmptyNote: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <p
    style={{
      margin: 0,
      padding: '16px 0',
      fontSize: 13,
      color: 'var(--hub-text-muted)',
      lineHeight: 1.5,
    }}
  >
    {children}
  </p>
);

const ActivityRow: React.FC<{ item: DashboardActivity; currency: Currency; last: boolean }> = ({
  item,
  currency,
  last,
}) => {
  const { t } = useTranslation('financeCommon');
  const income = item.kind === 'payment';
  const to = income ? '/finance/invoices/$id' : '/finance/expenses/$id';
  return (
    <Link
      to={to}
      params={{ id: item.id }}
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        padding: '10px 0',
        borderBottom: last ? 'none' : '1px solid var(--hub-border)',
        textDecoration: 'none',
      }}
    >
      <span
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          width: 30,
          height: 30,
          borderRadius: 8,
          flexShrink: 0,
          background: income ? 'var(--finance-success-soft)' : 'var(--finance-danger-soft)',
          color: income ? 'var(--finance-success)' : 'var(--finance-danger)',
        }}
      >
        {income ? <ArrowDownLeft size={14} /> : <ArrowUpRight size={14} />}
      </span>
      <span style={{ display: 'flex', flexDirection: 'column', minWidth: 0, flex: 1 }}>
        <span
          style={{
            fontSize: 13,
            fontWeight: 500,
            color: 'var(--hub-text)',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
          }}
        >
          {item.description}
        </span>
        <span style={{ fontSize: 11, color: 'var(--hub-text-dim)' }}>
          {formatDateShort(item.date)} ·{' '}
          {income
            ? t('finance.dashboard.invoicePaid')
            : t(`financeExpenses.status.${item.status}`, { defaultValue: item.status })}
        </span>
      </span>
      <span
        style={{
          fontSize: 13,
          fontWeight: 600,
          fontVariantNumeric: 'tabular-nums',
          color: income ? 'var(--finance-success)' : 'var(--hub-text)',
        }}
      >
        {income ? '+' : ''}
        {formatMoney(item.amount, currency)}
      </span>
    </Link>
  );
};

const Dashboard: React.FC = () => {
  const { t } = useTranslation('financeCommon');
  const currency = useSettings((s) => s.currency);
  const { data, isPending, isError, isFetching, refetch } = useFinanceDashboard();
  const money = (n: number): string => formatMoney(n, currency);

  if (isError) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
        <HelmetContainer title={pageTitle(t('financeNav.dashboard'))} />
        <div role="alert" style={{ ...card, display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <AlertTriangle size={16} style={{ color: 'var(--finance-danger)' }} />
            <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--hub-text)' }}>
              {t('financeErrors.generic')}
            </span>
          </div>
          <button
            type="button"
            style={{ ...ghostBtn, alignSelf: 'flex-start' }}
            disabled={isFetching}
            onClick={() => void refetch()}
          >
            {t('financeCommon.refresh')}
          </button>
        </div>
      </div>
    );
  }

  const loading = isPending || !data;
  const stat = (n: number | undefined): string => (loading || n === undefined ? '…' : money(n));
  const monthly = data?.monthly ?? [];
  const hasChartData = monthly.some((m) => m.revenue > 0 || m.expenses > 0);
  const activity = data?.recentActivity ?? [];
  const netProfit = data?.netProfitMtd ?? 0;

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.25 }}
      style={{ display: 'flex', flexDirection: 'column', gap: 20 }}
    >
      <HelmetContainer title={pageTitle(t('financeNav.dashboard'))} />

      {data && data.pendingExpenseCount > 0 && (
        <Notice
          tone="warning"
          icon={<Clock size={14} style={{ color: 'var(--finance-warning)', flexShrink: 0 }} />}
          to="/finance/expenses"
          linkLabel={t('finance.profitLoss.reviewExpenses')}
        >
          {t('finance.dashboard.pendingNotice', {
            count: data.pendingExpenseCount,
            amount: money(data.pendingExpenses),
          })}
        </Notice>
      )}
      {data && data.overdueInvoiceCount > 0 && (
        <Notice
          tone="danger"
          icon={
            <AlertTriangle size={14} style={{ color: 'var(--finance-danger)', flexShrink: 0 }} />
          }
          to="/finance/invoices"
          linkLabel={t('finance.dashboard.viewInvoices')}
        >
          {t('finance.dashboard.overdueNotice', {
            count: data.overdueInvoiceCount,
            amount: money(data.overdueAmount),
          })}
        </Notice>
      )}

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
          gap: 14,
        }}
      >
        <StatCard
          label={t('finance.dashboard.revenueThisMonth')}
          value={stat(data?.totalRevenueMtd)}
          hint={t('finance.dashboard.revenueHint')}
        />
        <StatCard
          label={t('finance.dashboard.expensesThisMonth')}
          value={stat(data?.totalExpensesMtd)}
          hint={t('finance.dashboard.expensesHint')}
        />
        <StatCard
          label={t('finance.dashboard.netProfitThisMonth')}
          value={stat(data?.netProfitMtd)}
          tone={
            loading ? undefined : netProfit < 0 ? 'var(--finance-danger)' : 'var(--finance-success)'
          }
        />
        <StatCard
          label={t('finance.dashboard.outstanding')}
          value={stat(data?.totalReceivable)}
          hint={
            data && data.overdueAmount > 0
              ? t('finance.dashboard.overdueHint', { amount: money(data.overdueAmount) })
              : t('finance.dashboard.outstandingHint')
          }
        />
      </div>

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))',
          gap: 16,
        }}
      >
        <ErrorBoundary>
          <div style={card}>
            <h2
              style={{
                margin: '0 0 16px',
                fontSize: 14,
                fontWeight: 600,
                color: 'var(--hub-text)',
              }}
            >
              {t('finance.dashboard.revenueVsExpenses')}
            </h2>
            {loading ? (
              <EmptyNote>{t('financeCommon.loading')}</EmptyNote>
            ) : hasChartData ? (
              <MonthlyChart months={monthly} currency={currency} />
            ) : (
              <EmptyNote>{t('finance.dashboard.chartEmpty')}</EmptyNote>
            )}
          </div>
        </ErrorBoundary>
        <ErrorBoundary>
          <div style={card}>
            <h2
              style={{ margin: '0 0 8px', fontSize: 14, fontWeight: 600, color: 'var(--hub-text)' }}
            >
              {t('finance.dashboard.recentActivity')}
            </h2>
            {loading ? (
              <EmptyNote>{t('financeCommon.loading')}</EmptyNote>
            ) : activity.length === 0 ? (
              <EmptyNote>{t('finance.dashboard.activityEmpty')}</EmptyNote>
            ) : (
              activity.map((item, i) => (
                <ActivityRow
                  key={`${item.kind}-${item.id}`}
                  item={item}
                  currency={currency}
                  last={i === activity.length - 1}
                />
              ))
            )}
          </div>
        </ErrorBoundary>
      </div>

      {data && data.overdueInvoices.length > 0 && (
        <ErrorBoundary>
          <OverdueInvoicesWidget
            items={data.overdueInvoices.map((inv) => ({
              id: inv.id,
              number: inv.invoiceNumber,
              client: inv.clientName,
              amount: inv.amountDue,
              daysOverdue: inv.daysOverdue,
            }))}
          />
        </ErrorBoundary>
      )}
    </motion.div>
  );
};

export default Dashboard;
