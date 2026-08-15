import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import {
  DollarSign,
  ShoppingCart,
  Users,
  Percent,
  ArrowUpRight,
  ArrowDownRight,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Bar } from 'react-chartjs-2';
import type { ChartOptions } from 'chart.js';
import { useInsightsDashboard } from '../../shared/hooks/queries/useInsightsDashboard';
import { useRevenueSeries } from '../../shared/hooks/queries/useRevenue';
import { formatMoney } from '../../shared/utils/money';
import { formatNumber, formatPercent } from '../../shared/lib/format';
import { CardSkeleton } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';
import EmptyState from '../../shared/components/EmptyState';
import CaretAskBar from '../../shared/components/CaretAskBar';

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

const sparklineOptions: ChartOptions<'bar'> = {
  responsive: true,
  maintainAspectRatio: false,
  animation: false,
  plugins: { legend: { display: false }, tooltip: { enabled: false } },
  scales: {
    x: { display: false },
    y: { display: false },
  },
};

interface DeltaBadgeProps {
  deltaPct: number;
}

const DeltaBadge = ({ deltaPct }: DeltaBadgeProps): React.JSX.Element => {
  const up = deltaPct >= 0;
  return (
    <div className="flex items-center gap-1 mt-1">
      {up ? (
        <ArrowUpRight size={12} style={{ color: 'var(--delta-up)' }} />
      ) : (
        <ArrowDownRight size={12} style={{ color: 'var(--delta-down)' }} />
      )}
      <span className="text-xs" style={{ color: up ? 'var(--delta-up)' : 'var(--delta-down)' }}>
        {formatPercent(deltaPct)} vs prior
      </span>
    </div>
  );
};

const Dashboard: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data, isLoading, isError, refetch } = useInsightsDashboard();
  const { data: seriesData } = useRevenueSeries();

  if (isError) {
    return <ErrorState message={t('insights.errors.generic')} onRetry={refetch} />;
  }

  const revenueSparkPoints = (seriesData?.series ?? []).slice(-12);
  const sparklineData = {
    labels: revenueSparkPoints.map((p) => p.date),
    datasets: [
      {
        data: revenueSparkPoints.map((p) => p.amountMinorUnits / 100),
        backgroundColor: 'var(--accent-muted)',
        hoverBackgroundColor: 'var(--accent-muted)',
        borderColor: 'var(--accent)',
        borderWidth: 1,
        borderRadius: 3,
      },
    ],
  };

  const stats = [
    {
      label: t('insights.dashboard.cards.revenue'),
      value: data ? formatMoney(data.revenue.value) : '-',
      deltaPct: data?.revenue.deltaPct ?? 0,
      icon: DollarSign,
      sparkline: true,
    },
    {
      label: t('insights.dashboard.cards.customers'),
      value: data ? formatNumber(data.customers.value) : '-',
      deltaPct: data?.customers.deltaPct ?? 0,
      icon: Users,
      sparkline: false,
    },
    {
      label: t('insights.dashboard.cards.orders'),
      value: data ? formatNumber(data.orders.value) : '-',
      deltaPct: data?.orders.deltaPct ?? 0,
      icon: ShoppingCart,
      sparkline: false,
    },
    {
      label: t('insights.dashboard.cards.margin'),
      value: data ? `${data.margin.value.toFixed(1)}%` : '-',
      deltaPct: data?.margin.deltaPct ?? 0,
      icon: Percent,
      sparkline: false,
    },
  ];

  const isEmpty =
    !!data && data.revenue.value === 0 && data.customers.value === 0 && data.orders.value === 0;

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl mx-auto"
    >
      <HelmetContainer title="Dashboard · rtylr" />
      <motion.div variants={itemVariants as any}>
        <CaretAskBar />
      </motion.div>

      <motion.div variants={itemVariants as any}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.dashboard.title')}
        </h2>
        <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
          {today}
        </p>
      </motion.div>

      {isLoading ? (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      ) : (
        <motion.div
          variants={containerVariants as any}
          className="grid grid-cols-2 lg:grid-cols-4 gap-4"
        >
          {stats.map((stat) => {
            const Icon = stat.icon;
            return (
              <motion.div
                key={stat.label}
                variants={itemVariants as any}
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
                    style={{ background: 'var(--accent-muted)' }}
                  >
                    <Icon size={14} style={{ color: 'var(--accent)' }} />
                  </span>
                </div>
                <p className="text-2xl font-bold truncate" style={{ color: 'var(--hub-text)' }}>
                  {stat.value}
                </p>
                <DeltaBadge deltaPct={stat.deltaPct} />
                {stat.sparkline && revenueSparkPoints.length > 1 && (
                  <div style={{ height: 44, marginTop: 12 }}>
                    <Bar data={sparklineData} options={sparklineOptions} />
                  </div>
                )}
              </motion.div>
            );
          })}
        </motion.div>
      )}

      {!isLoading && isEmpty && (
        <EmptyState
          title={t('insights.dashboard.emptyTitle')}
          description={t('insights.dashboard.emptyDescription')}
        />
      )}
    </motion.div>
  );
};

export default Dashboard;
