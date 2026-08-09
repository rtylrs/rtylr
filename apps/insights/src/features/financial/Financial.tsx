import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Clock } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  useCashFlow,
  useExpensesByCategory,
  useInvoiceAging,
} from '../../shared/hooks/queries/useInsightsFinancial';
import { formatMoney } from '../../shared/utils/money';
import { formatNumber } from '../../shared/lib/format';
import { SkeletonRow } from '../../shared/components/Skeleton';
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

const Financial: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: cfRes, isLoading: cfLoading, isError, refetch } = useCashFlow();
  const { data: expRes, isLoading: expLoading } = useExpensesByCategory();
  const { data: agingRes, isLoading: agingLoading } = useInvoiceAging();

  if (isError) {
    return <ErrorState message={t('insights.errors.generic')} onRetry={refetch} />;
  }

  const cashFlow = cfRes?.series ?? [];
  const expenses = expRes?.categories ?? [];
  const buckets = agingRes?.buckets ?? [];

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <HelmetContainer title="Financial · rtylr" />
      <motion.div variants={itemVariants as any}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.financial.title')}
        </h2>
      </motion.div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <motion.div
          variants={itemVariants as any}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
            {t('insights.financial.expenses')}
          </h3>
          {expLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <SkeletonRow key={i} cols={2} />
              ))}
            </div>
          ) : expenses.length === 0 ? (
            <EmptyState title="" description={t('insights.financial.empty')} />
          ) : (
            <div className="space-y-1">
              {expenses.map((e) => (
                <div
                  key={e.categoryId}
                  className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                  style={{ background: 'var(--hub-hover)' }}
                >
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                    {e.categoryName}
                  </span>
                  <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                    {formatMoney(e.amountMinorUnits)}
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
          <div className="flex items-center gap-2 mb-4">
            <Clock size={14} style={{ color: 'var(--accent)' }} />
            <h3 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              {t('insights.financial.aging')}
            </h3>
          </div>
          {agingLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <SkeletonRow key={i} cols={2} />
              ))}
            </div>
          ) : buckets.length === 0 ? (
            <EmptyState title="" description={t('insights.financial.empty')} />
          ) : (
            <div className="space-y-1">
              {buckets.map((b) => (
                <div
                  key={b.bucket}
                  className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                  style={{ background: 'var(--hub-hover)' }}
                >
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                    {b.bucket}
                  </span>
                  <div className="flex items-center gap-3">
                    <span className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                      {formatNumber(b.invoiceCount)} invoices
                    </span>
                    <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                      {formatMoney(b.amountMinorUnits)}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </motion.div>
      </div>

      <motion.div
        variants={itemVariants as any}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.financial.cashFlow')}
        </h3>
        {cfLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : cashFlow.length === 0 ? (
          <EmptyState title="" description={t('insights.financial.empty')} />
        ) : (
          <div className="space-y-1">
            {cashFlow.map((c) => {
              const net = c.inflowMinorUnits - c.outflowMinorUnits;
              return (
                <div
                  key={c.date}
                  className="grid grid-cols-4 items-center px-3 py-2.5 rounded-lg"
                  style={{ background: 'var(--hub-hover)' }}
                >
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                    {c.date}
                  </span>
                  <span className="text-xs text-center" style={{ color: 'var(--delta-up)' }}>
                    {formatMoney(c.inflowMinorUnits)}
                  </span>
                  <span className="text-xs text-center" style={{ color: 'var(--delta-down)' }}>
                    {formatMoney(c.outflowMinorUnits)}
                  </span>
                  <span
                    className="text-xs font-semibold text-right"
                    style={{ color: 'var(--hub-text)' }}
                  >
                    {formatMoney(net)}
                  </span>
                </div>
              );
            })}
          </div>
        )}
      </motion.div>
    </motion.div>
  );
};

export default Financial;
