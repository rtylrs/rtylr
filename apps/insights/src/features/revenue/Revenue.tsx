import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { TrendingUp, ArrowUpRight, ArrowDownRight } from 'lucide-react';
import { useRevenue, useRevenueByApp, useRevenueComparison } from './useRevenue';
import { formatCurrency, formatPercent } from '../../shared/lib/format';
import { CardSkeleton, SkeletonRow } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';
import EmptyState from '../../shared/components/EmptyState';

const ACCENT = '#6366f1';

const periods = [
  { label: 'Daily', value: 'daily' },
  { label: 'Weekly', value: 'weekly' },
  { label: 'Monthly', value: 'monthly' },
];

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Revenue: React.FC = () => {
  const [period, setPeriod] = useState('monthly');
  const { data: revenue, isLoading, isError, refetch } = useRevenue(period);
  const { data: byApp, isLoading: byAppLoading } = useRevenueByApp(period);
  const { data: comparison, isLoading: compLoading } = useRevenueComparison(period);

  if (isError) {
    return <ErrorState message="Failed to load revenue data." onRetry={refetch} />;
  }

  const revenueData: any[] = revenue?.data ?? revenue ?? [];
  const appBreakdown: any[] = byApp?.data ?? byApp ?? [];
  const compData = comparison ?? {};

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
            Revenue
          </h2>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            Track revenue across all channels
          </p>
        </div>
        <div className="flex gap-1 rounded-lg p-0.5" style={{ background: 'var(--hub-hover)' }}>
          {periods.map((p) => (
            <button
              key={p.value}
              onClick={() => setPeriod(p.value)}
              className="px-3 py-1.5 rounded-md text-xs font-medium transition-colors"
              style={{
                background: period === p.value ? ACCENT : 'transparent',
                color: period === p.value ? '#fff' : 'var(--hub-text-muted)',
              }}
            >
              {p.label}
            </button>
          ))}
        </div>
      </motion.div>

      {/* Comparison Cards */}
      {compLoading ? (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          {Array.from({ length: 3 }).map((_, i) => <CardSkeleton key={i} />)}
        </div>
      ) : (
        <motion.div variants={containerVariants} className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          <motion.div
            variants={itemVariants}
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <p className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>Current Period</p>
            <p className="text-2xl font-bold mt-1" style={{ color: 'var(--hub-text)' }}>
              {formatCurrency(compData.current ?? 0)}
            </p>
            <div className="flex items-center gap-1 mt-1">
              {(compData.changePercent ?? 0) >= 0
                ? <ArrowUpRight size={12} style={{ color: '#10b981' }} />
                : <ArrowDownRight size={12} style={{ color: '#f87171' }} />}
              <span className="text-xs" style={{ color: (compData.changePercent ?? 0) >= 0 ? '#10b981' : '#f87171' }}>
                {formatPercent(compData.changePercent ?? 0)} vs previous
              </span>
            </div>
          </motion.div>
          <motion.div
            variants={itemVariants}
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <p className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>Previous Period</p>
            <p className="text-2xl font-bold mt-1" style={{ color: 'var(--hub-text)' }}>
              {formatCurrency(compData.previous ?? 0)}
            </p>
          </motion.div>
          <motion.div
            variants={itemVariants}
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <p className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>Difference</p>
            <p className="text-2xl font-bold mt-1" style={{ color: 'var(--hub-text)' }}>
              {formatCurrency(compData.difference ?? 0)}
            </p>
          </motion.div>
        </motion.div>
      )}

      {/* Revenue Chart Data */}
      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center gap-2 mb-4">
          <TrendingUp size={14} style={{ color: ACCENT }} />
          <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
            Revenue Trend
          </h3>
        </div>
        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => <SkeletonRow key={i} cols={3} />)}
          </div>
        ) : revenueData.length === 0 ? (
          <EmptyState title="No revenue data" description="Revenue data will appear here once transactions are recorded." />
        ) : (
          <div className="space-y-1">
            {revenueData.map((item: any, i: number) => (
              <div
                key={i}
                className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                  {item.date ?? item.period ?? `Period ${i + 1}`}
                </span>
                <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                  {formatCurrency(item.revenue ?? item.amount ?? 0)}
                </span>
              </div>
            ))}
          </div>
        )}
      </motion.div>

      {/* By App Breakdown */}
      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          Revenue by App
        </h3>
        {byAppLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 3 }).map((_, i) => <SkeletonRow key={i} cols={2} />)}
          </div>
        ) : appBreakdown.length === 0 ? (
          <EmptyState title="No app breakdown" description="Breakdown by app will appear here." />
        ) : (
          <div className="space-y-1">
            {appBreakdown.map((item: any, i: number) => (
              <div
                key={i}
                className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                  {item.app ?? item.name ?? 'Unknown'}
                </span>
                <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                  {formatCurrency(item.revenue ?? item.amount ?? 0)}
                </span>
              </div>
            ))}
          </div>
        )}
      </motion.div>
    </motion.div>
  );
};

export default Revenue;
