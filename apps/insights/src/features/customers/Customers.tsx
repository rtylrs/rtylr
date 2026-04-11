import React from 'react';
import { motion } from 'motion/react';
import { Users, UserPlus, UserMinus, Star } from 'lucide-react';
import { useCustomerAnalytics, useTopCustomers, useCustomerChurn, useCustomerLtv } from './useCustomers';
import { formatCurrency, formatNumber, formatPercent } from '../../shared/lib/format';
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

const Customers: React.FC = () => {
  const { data: analytics, isLoading: analyticsLoading, isError, refetch } = useCustomerAnalytics();
  const { data: topCustomers, isLoading: topLoading } = useTopCustomers(10);
  const { data: churn, isLoading: churnLoading } = useCustomerChurn();
  const { data: ltv, isLoading: ltvLoading } = useCustomerLtv();

  const isLoading = analyticsLoading || topLoading || churnLoading || ltvLoading;

  if (isError) {
    return <ErrorState message="Failed to load customer data." onRetry={refetch} />;
  }

  const newCustomers = analytics?.newCustomers ?? 0;
  const returningCustomers = analytics?.returningCustomers ?? 0;
  const totalCustomers = analytics?.totalCustomers ?? 0;
  const churnRate = churn?.rate ?? 0;
  const avgLtv = ltv?.average ?? 0;

  const topList: any[] = topCustomers?.data ?? topCustomers ?? [];

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <motion.div variants={itemVariants}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          Customers
        </h2>
        <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
          Customer analytics and retention metrics
        </p>
      </motion.div>

      {isLoading ? (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => <CardSkeleton key={i} />)}
        </div>
      ) : (
        <motion.div variants={containerVariants} className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { label: 'Total Customers', value: formatNumber(totalCustomers), icon: Users, color: ACCENT },
            { label: 'New Customers', value: formatNumber(newCustomers), icon: UserPlus, color: '#10b981' },
            { label: 'Returning', value: formatNumber(returningCustomers), icon: Star, color: '#f59e0b' },
            { label: 'Churn Rate', value: `${churnRate.toFixed(1)}%`, icon: UserMinus, color: '#f87171' },
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

      {/* Average LTV */}
      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-2" style={{ color: 'var(--hub-text)' }}>
          Customer Lifetime Value
        </h3>
        <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>
          {ltvLoading ? '...' : formatCurrency(avgLtv)}
        </p>
        <p className="text-xs mt-1" style={{ color: 'var(--hub-text-dim)' }}>Average per customer</p>
      </motion.div>

      {/* Top Customers */}
      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          Top Customers by Spend
        </h3>
        {topLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => <SkeletonRow key={i} cols={3} />)}
          </div>
        ) : topList.length === 0 ? (
          <EmptyState title="No customer data" description="Customer data will appear here once transactions are recorded." />
        ) : (
          <div className="space-y-1">
            {topList.map((c: any, i: number) => (
              <div
                key={c.id ?? i}
                className="grid grid-cols-3 items-center px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <div className="flex items-center gap-2">
                  <span className="text-[10px] w-5" style={{ color: 'var(--hub-text-dim)' }}>#{i + 1}</span>
                  <span className="text-xs font-medium truncate" style={{ color: 'var(--hub-text)' }}>
                    {c.name ?? c.email ?? 'Unknown'}
                  </span>
                </div>
                <span className="text-xs text-center" style={{ color: 'var(--hub-text-muted)' }}>
                  {formatNumber(c.orders ?? c.orderCount ?? 0)} orders
                </span>
                <span className="text-xs font-semibold text-right" style={{ color: 'var(--hub-text)' }}>
                  {formatCurrency(c.totalSpend ?? c.revenue ?? 0)}
                </span>
              </div>
            ))}
          </div>
        )}
      </motion.div>
    </motion.div>
  );
};

export default Customers;
