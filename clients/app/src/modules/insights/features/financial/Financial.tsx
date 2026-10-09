import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { Clock, Receipt, ArrowLeftRight } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { EmptyState } from '@voxire/ui';
import { refreshEmptyStateAction } from '../../shared/lib/refreshAction';
import {
  useCashFlow,
  useExpensesByCategory,
  useInvoiceAging,
} from '../../shared/hooks/queries/useInsightsFinancial';
import { useInsightsSettings } from '../../shared/hooks/queries/useInsightsSettings';
import { formatMoney, type Currency } from '../../shared/utils/money';
import { formatNumber } from '../../shared/lib/format';
import { SkeletonRow } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';

const Financial: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: cfRes, isLoading: cfLoading, isError, refetch } = useCashFlow();
  const { data: expRes, isLoading: expLoading, refetch: refetchExpenses } = useExpensesByCategory();
  const { data: agingRes, isLoading: agingLoading, refetch: refetchAging } = useInvoiceAging();
  const { data: settings } = useInsightsSettings();
  const currency = (settings?.defaultCurrency as Currency) || 'USD';

  if (isError) {
    return <ErrorState message={t('insights.errors.generic')} onRetry={refetch} />;
  }

  const cashFlow = cfRes?.series ?? [];
  const expenses = expRes?.categories ?? [];
  const buckets = agingRes?.buckets ?? [];

  return (
    <div className="space-y-6 max-w-7xl">
      <HelmetContainer title={t('insights.financial.metaTitle')} />
      <div>
        <h1 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.financial.title')}
        </h1>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <h2 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
            {t('insights.financial.expenses')}
          </h2>
          {expLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <SkeletonRow key={i} cols={2} />
              ))}
            </div>
          ) : expenses.length === 0 ? (
            <EmptyState
              icon={Receipt}
              title={t('insights.financial.emptyExpenses.title')}
              description={t('insights.financial.emptyExpenses.description')}
              action={refreshEmptyStateAction(t('common.refresh'), () => refetchExpenses())}
              size="compact"
            />
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
                    {formatMoney(e.amountMinorUnits, currency)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        <div
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <div className="flex items-center gap-2 mb-4">
            <Clock size={14} style={{ color: 'var(--accent)' }} />
            <h2 className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
              {t('insights.financial.aging')}
            </h2>
          </div>
          {agingLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <SkeletonRow key={i} cols={2} />
              ))}
            </div>
          ) : buckets.length === 0 ? (
            <EmptyState
              icon={Clock}
              title={t('insights.financial.emptyAging.title')}
              description={t('insights.financial.emptyAging.description')}
              action={refreshEmptyStateAction(t('common.refresh'), () => refetchAging())}
              size="compact"
            />
          ) : (
            <div className="space-y-1">
              {buckets.map((b) => (
                <div
                  key={b.bucket}
                  className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                  style={{ background: 'var(--hub-hover)' }}
                >
                  <div>
                    <p className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                      {b.bucket}
                    </p>
                    <p className="text-[11px]" style={{ color: 'var(--hub-text-dim)' }}>
                      {t('insights.financial.invoices', {
                        count: b.invoiceCount,
                        formatted: formatNumber(b.invoiceCount),
                      })}
                    </p>
                  </div>
                  <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                    {formatMoney(b.amountMinorUnits, currency)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <div
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h2 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.financial.cashFlow')}
        </h2>
        {cfLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <SkeletonRow key={i} cols={4} />
            ))}
          </div>
        ) : cashFlow.length === 0 ? (
          <EmptyState
            icon={ArrowLeftRight}
            title={t('insights.financial.emptyCashFlow.title')}
            description={t('insights.financial.emptyCashFlow.description')}
            action={refreshEmptyStateAction(t('common.refresh'), () => refetch())}
          />
        ) : (
          <div className="space-y-1">
            {cashFlow.map((c) => {
              const net = c.inflowMinorUnits - c.outflowMinorUnits;
              return (
                <div
                  key={c.date}
                  className="flex items-center justify-between px-3 py-2.5 rounded-lg text-xs"
                  style={{ background: 'var(--hub-hover)' }}
                >
                  <span className="font-medium" style={{ color: 'var(--hub-text)' }}>
                    {c.date}
                  </span>
                  <div className="flex items-center gap-6">
                    <span style={{ color: 'var(--delta-up)' }}>
                      +{formatMoney(c.inflowMinorUnits, currency)}
                    </span>
                    <span style={{ color: 'var(--delta-down)' }}>
                      -{formatMoney(c.outflowMinorUnits, currency)}
                    </span>
                    <span
                      className="font-semibold w-24 text-end"
                      style={{ color: net >= 0 ? 'var(--delta-up)' : 'var(--delta-down)' }}
                    >
                      {formatMoney(net, currency)}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};

export default Financial;
