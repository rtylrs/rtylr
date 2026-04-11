import React from 'react';
import { motion } from 'framer-motion';
import { DollarSign, TrendingUp, TrendingDown, Clock } from 'lucide-react';
import { useProfitLoss, useExpenseBreakdown, useInvoiceAging, useCashFlow } from './useFinancial';
import { formatCurrency, formatNumber } from '../../shared/lib/format';
import { CardSkeleton, SkeletonRow } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';
import EmptyState from '../../shared/components/EmptyState';

const ACCENT = '#6366f1';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Financial: React.FC = () => {
  const { data: pl, isLoading: plLoading, isError, refetch } = useProfitLoss();
  const { data: expenses, isLoading: expLoading } = useExpenseBreakdown();
  const { data: aging, isLoading: agingLoading } = useInvoiceAging();
  const { data: cashFlow, isLoading: cfLoading } = useCashFlow('monthly');

  if (isError) {
    return <ErrorState message="Failed to load financial data." onRetry={refetch} />;
  }

  const totalRevenue = pl?.revenue ?? pl?.totalRevenue ?? 0;
  const totalExpenses = pl?.expenses ?? pl?.totalExpenses ?? 0;
  const netProfit = pl?.netProfit ?? (totalRevenue - totalExpenses);
  const profitMargin = totalRevenue > 0 ? (netProfit / totalRevenue) * 100 : 0;

  const expenseList: any[] = expenses?.data ?? expenses?.categories ?? expenses ?? [];
  const agingBuckets: any[] = aging?.data ?? aging?.buckets ?? aging ?? [];
  const cashFlowList: any[] = cashFlow?.data ?? cashFlow ?? [];

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <motion.div variants={itemVariants}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>Financial</h2>
        <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
          P&L summary, expenses, invoicing, and cash flow
        </p>
      </motion.div>

      {/* P&L Summary */}
      {plLoading ? (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => <CardSkeleton key={i} />)}
        </div>
      ) : (
        <motion.div variants={containerVariants} className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { label: 'Revenue', value: formatCurrency(totalRevenue), icon: TrendingUp, color: '#10b981' },
            { label: 'Expenses', value: formatCurrency(totalExpenses), icon: TrendingDown, color: '#f87171' },
            { label: 'Net Profit', value: formatCurrency(netProfit), icon: DollarSign, color: netProfit >= 0 ? '#10b981' : '#f87171' },
            { label: 'Profit Margin', value: `${profitMargin.toFixed(1)}%`, icon: DollarSign, color: ACCENT },
          ].map((stat) => {
            const Icon = stat.icon;
            return (
              <motion.div
                key={stat.label}
                variants={itemVariants}
                className="rounded-xl p-5"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              >
                <div className="flex items-center justify-between mb-3">
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>{stat.label}</span>
                  <span className="flex items-center justify-center h-7 w-7 rounded-lg" style={{ background: `${stat.color}18` }}>
                    <Icon size={14} style={{ color: stat.color }} />
                  </span>
                </div>
                <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>{stat.value}</p>
              </motion.div>
            );
          })}
        </motion.div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Expense Breakdown */}
        <motion.div
          variants={itemVariants}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>Expense Breakdown</h3>
          {expLoading ? (
            <div className="space-y-2">{Array.from({ length: 5 }).map((_, i) => <SkeletonRow key={i} cols={2} />)}</div>
          ) : expenseList.length === 0 ? (
            <EmptyState title="No expense data" description="Expense breakdown will appear here." />
          ) : (
            <div className="space-y-1">
              {expenseList.map((e: any, i: number) => (
                <div key={e.id ?? i} className="flex items-center justify-between px-3 py-2.5 rounded-lg" style={{ background: 'var(--hub-hover)' }}>
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>{e.category ?? e.name ?? 'Other'}</span>
                  <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>{formatCurrency(e.amount ?? 0)}</span>
                </div>
              ))}
            </div>
          )}
        </motion.div>

        {/* Invoice Aging */}
        <motion.div
          variants={itemVariants}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center gap-2 mb-4">
            <Clock size={14} style={{ color: '#f59e0b' }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>Invoice Aging</h3>
          </div>
          {agingLoading ? (
            <div className="space-y-2">{Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} cols={2} />)}</div>
          ) : agingBuckets.length === 0 ? (
            <EmptyState title="No invoice data" description="Invoice aging will appear here." />
          ) : (
            <div className="space-y-1">
              {agingBuckets.map((b: any, i: number) => (
                <div key={i} className="flex items-center justify-between px-3 py-2.5 rounded-lg" style={{ background: 'var(--hub-hover)' }}>
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>{b.bucket ?? b.range ?? `${i * 30}-${(i + 1) * 30} days`}</span>
                  <div className="flex items-center gap-3">
                    <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>{formatNumber(b.count ?? 0)} invoices</span>
                    <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>{formatCurrency(b.amount ?? b.total ?? 0)}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </motion.div>
      </div>

      {/* Cash Flow Trend */}
      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>Cash Flow Trend</h3>
        {cfLoading ? (
          <div className="space-y-2">{Array.from({ length: 5 }).map((_, i) => <SkeletonRow key={i} cols={3} />)}</div>
        ) : cashFlowList.length === 0 ? (
          <EmptyState title="No cash flow data" description="Cash flow data will appear here." />
        ) : (
          <div className="space-y-1">
            {cashFlowList.map((c: any, i: number) => (
              <div key={i} className="grid grid-cols-4 items-center px-3 py-2.5 rounded-lg" style={{ background: 'var(--hub-hover)' }}>
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>{c.month ?? c.period ?? `Month ${i + 1}`}</span>
                <span className="text-xs text-center" style={{ color: '#10b981' }}>{formatCurrency(c.inflow ?? c.revenue ?? 0)}</span>
                <span className="text-xs text-center" style={{ color: '#f87171' }}>{formatCurrency(c.outflow ?? c.expenses ?? 0)}</span>
                <span className="text-xs font-semibold text-right" style={{ color: 'var(--hub-text)' }}>{formatCurrency(c.net ?? (c.inflow ?? 0) - (c.outflow ?? 0))}</span>
              </div>
            ))}
          </div>
        )}
      </motion.div>
    </motion.div>
  );
};

export default Financial;
