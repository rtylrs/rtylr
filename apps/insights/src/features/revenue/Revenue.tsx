import React from 'react';
import { motion } from 'motion/react';
import { TrendingUp, ArrowUpRight, ArrowDownRight } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  useRevenueSeries,
  useRevenueByApp,
  useRevenueComparison,
} from '../../shared/hooks/queries/useRevenue';
import { formatMoney } from '../../shared/utils/money';
import { formatPercent } from '../../shared/lib/format';
import { CardSkeleton, SkeletonRow } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';
import EmptyState from '../../shared/components/EmptyState';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Revenue: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: series, isLoading, isError, refetch } = useRevenueSeries();
  const { data: byApp, isLoading: byAppLoading } = useRevenueByApp();
  const { data: comparison, isLoading: compLoading } = useRevenueComparison();

  if (isError) {
    return <ErrorState message={t('insights.errors.generic')} onRetry={refetch} />;
  }

  const seriesRows = series?.series ?? [];
  const appRows = byApp?.byApp ?? [];
  const difference = (comparison?.thisPeriod ?? 0) - (comparison?.lastPeriod ?? 0);
  const deltaPct = comparison?.deltaPct ?? 0;
  const deltaUp = deltaPct >= 0;

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <motion.div variants={itemVariants as any} className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
            {t('insights.revenue.title')}
          </h2>
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
        <motion.div
          variants={containerVariants as any}
          className="grid grid-cols-1 lg:grid-cols-3 gap-4"
        >
          <motion.div
            variants={itemVariants as any}
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <p className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
              {t('insights.revenue.currentPeriod')}
            </p>
            <p className="text-2xl font-bold mt-1" style={{ color: 'var(--hub-text)' }}>
              {formatMoney(comparison?.thisPeriod ?? 0)}
            </p>
            <div className="flex items-center gap-1 mt-1">
              {deltaUp ? (
                <ArrowUpRight size={12} style={{ color: 'var(--delta-up)' }} />
              ) : (
                <ArrowDownRight size={12} style={{ color: 'var(--delta-down)' }} />
              )}
              <span
                className="text-xs"
                style={{ color: deltaUp ? 'var(--delta-up)' : 'var(--delta-down)' }}
              >
                {formatPercent(deltaPct)} vs previous
              </span>
            </div>
          </motion.div>
          <motion.div
            variants={itemVariants as any}
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <p className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
              {t('insights.revenue.previousPeriod')}
            </p>
            <p className="text-2xl font-bold mt-1" style={{ color: 'var(--hub-text)' }}>
              {formatMoney(comparison?.lastPeriod ?? 0)}
            </p>
          </motion.div>
          <motion.div
            variants={itemVariants as any}
            className="rounded-xl p-5"
            style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
          >
            <p className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
              {t('insights.revenue.difference')}
            </p>
            <p className="text-2xl font-bold mt-1" style={{ color: 'var(--hub-text)' }}>
              {formatMoney(difference)}
            </p>
          </motion.div>
        </motion.div>
      )}

      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <div className="flex items-center gap-2 mb-4">
          <TrendingUp size={14} style={{ color: 'var(--accent)' }} />
          <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
            {t('insights.revenue.trend')}
          </h3>
        </div>
        {isLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : seriesRows.length === 0 ? (
          <EmptyState title="" description={t('insights.revenue.emptyTrend')} />
        ) : (
          <div className="space-y-1">
            {seriesRows.map((point) => (
              <div
                key={point.date}
                className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                  {point.date}
                </span>
                <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                  {formatMoney(point.amountMinorUnits)}
                </span>
              </div>
            ))}
          </div>
        )}
      </motion.div>

      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.revenue.byApp')}
        </h3>
        {byAppLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 3 }).map((_, i) => (
              <SkeletonRow key={i} cols={2} />
            ))}
          </div>
        ) : appRows.length === 0 ? (
          <EmptyState title="" description={t('insights.revenue.emptyByApp')} />
        ) : (
          <div className="space-y-1">
            {appRows.map((row) => (
              <div
                key={row.app}
                className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                  {row.app}
                </span>
                <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                  {formatMoney(row.amountMinorUnits)}{' '}
                  <span style={{ color: 'var(--hub-text-muted)' }}>({row.pct.toFixed(1)}%)</span>
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
