import React, { useMemo } from 'react';
import { HelmetContainer } from '@voxire/components';
import { Users, UserPlus, UserMinus, Star, DollarSign } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Bar } from 'react-chartjs-2';
import type { ChartOptions } from 'chart.js';
import { EmptyState } from '@voxire/ui';
import { refreshEmptyStateAction } from '../../shared/lib/refreshAction';
import {
  useCustomerKPIs,
  useTopCustomers,
  useCustomerChurn,
  useCustomerLTV,
} from '../../shared/hooks/queries/useInsightsCustomers';
import { useInsightsSettings } from '../../shared/hooks/queries/useInsightsSettings';
import { formatMoney, type Currency } from '../../shared/utils/money';
import { formatNumber } from '../../shared/lib/format';
import { CardSkeleton, SkeletonBlock, SkeletonRow } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';

const Customers: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: kpis, isLoading: kpisLoading, isError, refetch } = useCustomerKPIs();
  const {
    data: topRes,
    isLoading: topLoading,
    refetch: refetchTop,
  } = useTopCustomers({ limit: 10 });
  const { data: churnRes, isLoading: churnLoading, refetch: refetchChurn } = useCustomerChurn();
  const { data: ltvRes, isLoading: ltvLoading, refetch: refetchLtv } = useCustomerLTV();
  const { data: settings } = useInsightsSettings();
  const currency = (settings?.defaultCurrency as Currency) || 'USD';

  const topCustomersBarOptions: ChartOptions<'bar'> = useMemo(
    () => ({
      responsive: true,
      maintainAspectRatio: false,
      animation: false,
      indexAxis: 'y',
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
          grid: { color: 'var(--hub-border)' },
          ticks: {
            color: 'var(--hub-text-muted)',
            font: { size: 11 },
            callback: (value): string => formatMoney((value as number) * 100, currency),
          },
          border: { display: false },
        },
        y: {
          grid: { display: false },
          ticks: { color: 'var(--hub-text-muted)', font: { size: 11 } },
          border: { display: false },
        },
      },
    }),
    [currency]
  );

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

  const topCustomersBarData = {
    labels: topList.map((c) => c.name || c.email),
    datasets: [
      {
        data: topList.map((c) => c.lifetimeSpendMinorUnits / 100),
        backgroundColor: 'var(--accent-muted)',
        hoverBackgroundColor: 'var(--accent)',
        borderColor: 'var(--accent)',
        borderWidth: 1,
        borderRadius: 4,
      },
    ],
  };

  return (
    <div className="space-y-6 max-w-7xl">
      <HelmetContainer title={t('insights.customers.metaTitle')} />
      <div>
        <h1 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.customers.title')}
        </h1>
      </div>

      {kpisLoading ? (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            {
              id: 'total',
              label: t('insights.customers.kpis.total'),
              value: formatNumber(total),
              icon: Users,
            },
            {
              id: 'newThisMonth',

              label: t('insights.customers.kpis.newThisMonth'),
              value: formatNumber(newThisMonth),
              icon: UserPlus,
            },
            {
              id: 'active',

              label: t('insights.customers.kpis.active'),
              value: formatNumber(active),
              icon: Star,
            },
            {
              id: 'churned',

              label: t('insights.customers.kpis.churned'),
              value: formatNumber(churned),
              icon: UserMinus,
            },
          ].map((stat) => {
            const Icon = stat.icon;
            return (
              <div
                key={stat.id}
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
                    <Icon size={14} style={{ color: 'var(--accent-on-muted)' }} />
                  </span>
                </div>
                <p className="text-2xl font-bold" style={{ color: 'var(--hub-text)' }}>
                  {stat.value}
                </p>
              </div>
            );
          })}
        </div>
      )}

      <div
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h2 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.customers.top')}
        </h2>
        {topLoading ? (
          <SkeletonBlock className="h-64 w-full" />
        ) : topList.length === 0 ? (
          <EmptyState
            icon={Users}
            title={t('insights.customers.emptyTop.title')}
            description={t('insights.customers.emptyTop.description')}
            action={refreshEmptyStateAction(t('common.refresh'), () => refetchTop())}
          />
        ) : (
          <div
            style={{ height: Math.max(180, topList.length * 36) }}
            aria-label={t('insights.charts.topCustomers')}
          >
            <Bar data={topCustomersBarData} options={topCustomersBarOptions} />
          </div>
        )}
      </div>

      <div
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h2 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.customers.ltv')}
        </h2>
        {ltvLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <SkeletonRow key={i} cols={2} />
            ))}
          </div>
        ) : ltvBuckets.length === 0 ? (
          <EmptyState
            icon={DollarSign}
            title={t('insights.customers.emptyLtv.title')}
            description={t('insights.customers.emptyLtv.description')}
            action={refreshEmptyStateAction(t('common.refresh'), () => refetchLtv())}
          />
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
                  {t('insights.customers.customers', {
                    count: b.customerCount,
                    formatted: formatNumber(b.customerCount),
                  })}
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
        <h2 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          {t('insights.customers.churn')}
        </h2>
        {churnLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <SkeletonRow key={i} cols={2} />
            ))}
          </div>
        ) : churnWindows.length === 0 ? (
          <EmptyState
            icon={UserMinus}
            title={t('insights.customers.emptyChurn.title')}
            description={t('insights.customers.emptyChurn.description')}
            action={refreshEmptyStateAction(t('common.refresh'), () => refetchChurn())}
          />
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
                    ({formatNumber(w.churnedCount)} {t('insights.customers.churnedCount')})
                  </span>
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default Customers;
