import React from 'react';
import { motion } from 'motion/react';
import {
  DollarSign,
  ShoppingCart,
  Users,
  Briefcase,
  TrendingUp,
  Package,
} from 'lucide-react';
import { useKpis, useDashboardSummary } from './useDashboard';
import { formatCurrency, formatNumber } from '../../shared/lib/format';
import { CardSkeleton } from '../../shared/components/Skeleton';
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

const today = new Date().toLocaleDateString('en-US', {
  weekday: 'long',
  year: 'numeric',
  month: 'long',
  day: 'numeric',
});

const Dashboard: React.FC = () => {
  const { data: kpis, isLoading: kpisLoading, isError: kpisError, refetch: refetchKpis } = useKpis();
  const { data: summary, isLoading: summaryLoading, isError: summaryError, refetch: refetchSummary } = useDashboardSummary();

  const isLoading = kpisLoading || summaryLoading;
  const isError = kpisError && summaryError;

  if (isError) {
    return <ErrorState message="Failed to load dashboard data." onRetry={() => { refetchKpis(); refetchSummary(); }} />;
  }

  const totalRevenue = kpis?.totalRevenue ?? summary?.totalRevenue ?? 0;
  const ordersThisMonth = kpis?.ordersThisMonth ?? summary?.ordersThisMonth ?? 0;
  const customerCount = kpis?.customerCount ?? summary?.customerCount ?? 0;
  const staffCount = kpis?.staffCount ?? summary?.staffCount ?? 0;
  const avgOrderValue = kpis?.avgOrderValue ?? summary?.avgOrderValue ?? 0;
  const topProduct = kpis?.topProduct ?? summary?.topProduct ?? null;

  const stats = [
    {
      label: 'Total Revenue',
      value: formatCurrency(totalRevenue),
      icon: DollarSign,
      subtitle: 'All time',
    },
    {
      label: 'Orders This Month',
      value: formatNumber(ordersThisMonth),
      icon: ShoppingCart,
      subtitle: 'Current month',
    },
    {
      label: 'Customer Count',
      value: formatNumber(customerCount),
      icon: Users,
      subtitle: 'Active customers',
    },
    {
      label: 'Staff Count',
      value: formatNumber(staffCount),
      icon: Briefcase,
      subtitle: 'Active employees',
    },
    {
      label: 'Avg Order Value',
      value: formatCurrency(avgOrderValue),
      icon: TrendingUp,
      subtitle: 'Per transaction',
    },
    {
      label: 'Top Product',
      value: topProduct?.name ?? 'N/A',
      icon: Package,
      subtitle: topProduct?.revenue ? formatCurrency(topProduct.revenue) : 'By revenue',
    },
  ];

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <motion.div variants={itemVariants}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          Business Overview
        </h2>
        <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
          {today}
        </p>
      </motion.div>

      {isLoading ? (
        <div className="grid grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      ) : (
        <motion.div variants={containerVariants} className="grid grid-cols-2 lg:grid-cols-3 gap-4">
          {stats.map((stat) => {
            const Icon = stat.icon;
            return (
              <motion.div
                key={stat.label}
                variants={itemVariants}
                className="rounded-xl p-5"
                style={{
                  background: 'var(--hub-card)',
                  border: '1px solid var(--hub-border)',
                }}
              >
                <div className="flex items-center justify-between mb-3">
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                    {stat.label}
                  </span>
                  <span
                    className="flex items-center justify-center h-7 w-7 rounded-lg"
                    style={{ background: 'rgba(99,102,241,0.1)' }}
                  >
                    <Icon size={14} style={{ color: ACCENT }} />
                  </span>
                </div>
                <p
                  className="text-2xl font-bold truncate"
                  style={{ color: 'var(--hub-text)' }}
                >
                  {stat.value}
                </p>
                <p className="text-xs mt-1" style={{ color: 'var(--hub-text-dim)' }}>
                  {stat.subtitle}
                </p>
              </motion.div>
            );
          })}
        </motion.div>
      )}

      {!isLoading && totalRevenue === 0 && ordersThisMonth === 0 && (
        <EmptyState
          title="No KPI data yet"
          description="KPI data will appear here once your business starts generating activity."
        />
      )}
    </motion.div>
  );
};

export default Dashboard;
