import React from 'react';
import { motion } from 'motion/react';
import { Users, UserPlus, UserMinus, Star } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  useCustomerKPIs,
  useTopCustomers,
  useCustomerChurn,
  useCustomerLTV,
} from '../../shared/hooks/queries/useInsightsCustomers';
import { formatMoney } from '../../shared/utils/money';
import { formatNumber } from '../../shared/lib/format';
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

const Customers: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: kpis, isLoading: kpisLoading, isError, refetch } = useCustomerKPIs();
  const { data: topRes, isLoading: topLoading } = useTopCustomers({ limit: 10 });
  const { data: churnRes, isLoading: churnLoading } = useCustomerChurn();
  const { data: ltvRes, isLoading: ltvLoading } = useCustomerLTV();

  if (isError) {
    return <ErrorState message={t('insights.errors.generic')} onRetry={refetch} />;
  }

  const total = kpis?.total ?? 0;
  const newThisMonth = kpis?.newThisMonth ?? 0;
  const active = kpis?.active ?? 0;
  const churned = kpis?.churned ?? 0;
  const topList = topRes?.customers ?? [];
  const churnWindows = churnRes?.windows ?? [];
  const ltvBuckets = ltvRes?.buckets ?? [];

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <motion.div variants={itemVariants as any}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.customers.title')}
        </h2>
      </motion.div>

      {kpisLoading ? (
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
          {[
            { label: t('insights.customers.kpis.total'), value: formatNumber(total), icon: Users },
            {
              label: t('insights.customers.kpis.newThisMonth'),
              value: formatNumber(newThisMonth),
              icon: UserPlus,
            },
            {
              label: t('insights.customers.kpis.active'),
              value: formatNumber(active),
              icon: Star,
            },
            {
              label: t('insights.customers.kpis.churned'),
              value: formatNumber(churned),
              icon: UserMinus,
            },
          ].map((stat) => {
            const Icon = stat.icon;
            return (
              <motion.div
                key={stat.label}
                variants={itemVariants as any}
                className="rounded-xl p-5"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
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
                <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>
                  {stat.value}
                </p>
              </motion.div>
            );
          })}
        </motion.div>
      )}

      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.customers.top')}
        </h3>
        {topLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : topList.length === 0 ? (
          <EmptyState title="" description={t('insights.customers.empty')} />
        ) : (
          <div className="space-y-1">
            {topList.map((c, i) => (
              <div
                key={c.customerId}
                className="grid grid-cols-3 items-center px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <div className="flex items-center gap-2">
                  <span className="text-[10px] w-5" style={{ color: 'var(--hub-text-dim)' }}>
                    #{i + 1}
                  </span>
                  <span
                    className="text-xs font-medium truncate"
                    style={{ color: 'var(--hub-text)' }}
                  >
                    {c.name || c.email}
                  </span>
                </div>
                <span className="text-xs text-center" style={{ color: 'var(--hub-text-muted)' }}>
                  {formatNumber(c.orderCount)} orders
                </span>
                <span
                  className="text-xs font-semibold text-right"
                  style={{ color: 'var(--hub-text)' }}
                >
                  {formatMoney(c.lifetimeSpendMinorUnits)}
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
          {t('insights.customers.ltv')}
        </h3>
        {ltvLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <SkeletonRow key={i} cols={2} />
            ))}
          </div>
        ) : ltvBuckets.length === 0 ? (
          <EmptyState title="" description={t('insights.customers.empty')} />
        ) : (
          <div className="space-y-1">
            {ltvBuckets.map((b) => (
              <div
                key={b.label}
                className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                  {b.label}
                </span>
                <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                  {formatNumber(b.customerCount)} customers
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
          {t('insights.customers.churn')}
        </h3>
        {churnLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <SkeletonRow key={i} cols={2} />
            ))}
          </div>
        ) : churnWindows.length === 0 ? (
          <EmptyState title="" description={t('insights.customers.empty')} />
        ) : (
          <div className="space-y-1">
            {churnWindows.map((w) => (
              <div
                key={w.windowDays}
                className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                  {w.windowDays}d
                </span>
                <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                  {w.churnPct.toFixed(1)}%{' '}
                  <span style={{ color: 'var(--hub-text-muted)' }}>
                    ({formatNumber(w.churnedCount)} churned)
                  </span>
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
