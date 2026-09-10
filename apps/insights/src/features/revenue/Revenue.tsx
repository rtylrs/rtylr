import React, { useMemo } from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion, type Variants } from 'motion/react';
import { TrendingUp, ArrowUpRight, ArrowDownRight, PieChart } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Line, Bar } from 'react-chartjs-2';
import type { ChartOptions } from 'chart.js';
import { EmptyState } from '@voxire/ui';
import {
  useRevenueSeries,
  useRevenueByApp,
  useRevenueComparison,
} from '../../shared/hooks/queries/useRevenue';
import { useInsightsSettings } from '../../shared/hooks/queries/useInsightsSettings';
import { formatMoney, type Currency } from '../../shared/utils/money';
import { formatPercent } from '../../shared/lib/format';
import { CardSkeleton, SkeletonBlock } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';

const containerVariants: Variants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Revenue: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: series, isLoading, isError, refetch } = useRevenueSeries();
  const { data: byApp, isLoading: byAppLoading, refetch: refetchByApp } = useRevenueByApp();
  const { data: comparison, isLoading: compLoading } = useRevenueComparison();
  const { data: settings } = useInsightsSettings();
  const currency = (settings?.defaultCurrency as Currency) || 'USD';

  const areaOptions: ChartOptions<'line'> = useMemo(
    () => ({
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { display: false },
        tooltip: {
          backgroundColor: 'var(--hub-card)',
          titleColor: 'var(--hub-text-muted)',
          bodyColor: 'var(--hub-text)',
          borderColor: 'var(--hub-border)',
          borderWidth: 1,
          padding: 10,
          callbacks: {
            label: (ctx): string => formatMoney((ctx.raw as number) * 100, currency),
          },
        },
      },
      scales: {
        x: {
          grid: { display: false },
          ticks: {
            color: 'var(--hub-text-muted)',
            font: { size: 11 },
            maxTicksLimit: 6,
          },
          border: { display: false },
        },
        y: {
          grid: { color: 'var(--hub-border)' },
          ticks: {
            color: 'var(--hub-text-muted)',
            font: { size: 11 },
            callback: (value): string => formatMoney((value as number) * 100, currency),
          },
          border: { display: false },
        },
      },
    }),
    [currency]
  );

  const barOptions: ChartOptions<'bar'> = useMemo(
    () => ({
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { display: false },
        tooltip: {
          backgroundColor: 'var(--hub-card)',
          titleColor: 'var(--hub-text-muted)',
          bodyColor: 'var(--hub-text)',
          borderColor: 'var(--hub-border)',
          borderWidth: 1,
          padding: 10,
          callbacks: {
            label: (ctx): string => formatMoney((ctx.raw as number) * 100, currency),
          },
        },
      },
      scales: {
        x: {
          grid: { display: false },
          ticks: { color: 'var(--hub-text-muted)', font: { size: 11 } },
          border: { display: false },
        },
        y: {
          grid: { color: 'var(--hub-border)' },
          ticks: {
            color: 'var(--hub-text-muted)',
            font: { size: 11 },
            callback: (value): string => formatMoney((value as number) * 100, currency),
          },
          border: { display: false },
        },
      },
    }),
    [currency]
  );

  if (isError) {
    return <ErrorState message={t('insights.errors.generic')} onRetry={refetch} />;
  }

  const seriesRows = series?.series ?? [];
  const areaData = {
    labels: seriesRows.map((p) => p.date),
    datasets: [
      {
        data: seriesRows.map((p) => p.amountMinorUnits / 100),
        borderColor: 'var(--accent)',
        backgroundColor: 'rgba(99, 102, 241, 0.08)',
        fill: true,
        tension: 0.35,
        borderWidth: 2,
        pointRadius: 0,
        pointHoverRadius: 4,
      },
    ],
  };

  const appRows = byApp?.byApp ?? [];
  const barData = {
    labels: appRows.map((a) => a.app),
    datasets: [
      {
        data: appRows.map((a) => a.amountMinorUnits / 100),
        backgroundColor: 'var(--accent-muted)',
        hoverBackgroundColor: 'var(--accent-muted)',
        borderColor: 'var(--accent)',
        borderWidth: 1,
        borderRadius: 4,
      },
    ],
  };

  const deltaPct = comparison?.deltaPct ?? 0;
  const deltaUp = deltaPct > 0;
  const difference = (comparison?.thisPeriod ?? 0) - (comparison?.lastPeriod ?? 0);

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <HelmetContainer title="Revenue · rtylr" />
      <motion.div variants={itemVariants} className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
            {t('insights.revenue.title')}
          </h1>
          <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            {t('insights.revenue.subtitle')}
          </p>
        </div>
      </motion.div>

      {compLoading ? (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      ) : (
        <motion.div variants={containerVariants} className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          <motion.div
            variants={itemVariants}
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <p className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
              {t('insights.revenue.currentPeriod')}
            </p>
            <p className="text-2xl font-bold mt-1" style={{ color: 'var(--hub-text)' }}>
              {formatMoney(comparison?.thisPeriod ?? 0, currency)}
            </p>
            <div className="flex items-center gap-1 mt-1">
              {deltaPct === 0 ? (
                <span className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
                  {formatPercent(0)} {t('insights.revenue.vsPrevious')}
                </span>
              ) : (
                <>
                  {deltaUp ? (
                    <ArrowUpRight size={12} style={{ color: 'var(--delta-up)' }} />
                  ) : (
                    <ArrowDownRight size={12} style={{ color: 'var(--delta-down)' }} />
                  )}
                  <span
                    className="text-xs"
                    style={{ color: deltaUp ? 'var(--delta-up)' : 'var(--delta-down)' }}
                  >
                    {formatPercent(deltaPct)} {t('insights.revenue.vsPrevious')}
                  </span>
                </>
              )}
            </div>
          </motion.div>
          <motion.div
            variants={itemVariants}
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <p className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
              {t('insights.revenue.previousPeriod')}
            </p>
            <p className="text-2xl font-bold mt-1" style={{ color: 'var(--hub-text)' }}>
              {formatMoney(comparison?.lastPeriod ?? 0, currency)}
            </p>
          </motion.div>
          <motion.div
            variants={itemVariants}
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <p className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
              {t('insights.revenue.difference')}
            </p>
            <p className="text-2xl font-bold mt-1" style={{ color: 'var(--hub-text)' }}>
              {formatMoney(difference, currency)}
            </p>
          </motion.div>
        </motion.div>
      )}

      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center gap-2 mb-4">
          <TrendingUp size={14} style={{ color: 'var(--accent)' }} />
          <h2 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
            {t('insights.revenue.trend')}
          </h2>
        </div>
        {isLoading ? (
          <SkeletonBlock className="h-56 w-full" />
        ) : seriesRows.length === 0 ? (
          <EmptyState
            icon={TrendingUp}
            title={t('insights.revenue.emptyTrend.title')}
            description={t('insights.revenue.emptyTrend.description')}
            action={{ label: t('common.refresh'), onClick: () => refetch() }}
          />
        ) : (
          <div style={{ height: 220 }} aria-label="Revenue trend chart">
            <Line data={areaData} options={areaOptions} />
          </div>
        )}
      </motion.div>

      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h2 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.revenue.byApp')}
        </h2>
        {byAppLoading ? (
          <SkeletonBlock className="h-44 w-full" />
        ) : appRows.length === 0 ? (
          <EmptyState
            icon={PieChart}
            title={t('insights.revenue.emptyByApp.title')}
            description={t('insights.revenue.emptyByApp.description')}
            action={{ label: t('common.refresh'), onClick: () => refetchByApp() }}
          />
        ) : (
          <div style={{ height: 180 }} aria-label="Revenue by app chart">
            <Bar data={barData} options={barOptions} />
          </div>
        )}
      </motion.div>
    </motion.div>
  );
};

export default Revenue;
