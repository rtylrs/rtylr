import React from 'react';
import { HelmetContainer } from '@voxire/components';
import {
  DollarSign,
  ShoppingCart,
  Users,
  Percent,
  ArrowUpRight,
  ArrowDownRight,
  BarChart3,
  CheckCircle2,
  AlertTriangle,
  FileText,
  Clock,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Bar } from 'react-chartjs-2';
import type { ChartOptions } from 'chart.js';
import { EmptyState } from '@voxire/ui';
import { refreshEmptyStateAction } from '../../shared/lib/refreshAction';
import {
  useInsightsDashboard,
  useWorkSummary,
} from '../../shared/hooks/queries/useInsightsDashboard';
import { useInvoiceAging } from '../../shared/hooks/queries/useInsightsFinancial';
import { useRevenueSeries } from '../../shared/hooks/queries/useRevenue';
import { useInsightsSettings } from '../../shared/hooks/queries/useInsightsSettings';
import { formatMoney, type Currency } from '../../shared/utils/money';
import { formatNumber, formatPercent } from '../../shared/lib/format';
import { CardSkeleton } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';
import CaretAskBar from '../../shared/components/CaretAskBar';

export const formatToday = (language: string): string =>
  new Date().toLocaleDateString(language, {
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
  const { t } = useTranslation();
  if (deltaPct === 0) {
    return (
      <div className="flex items-center gap-1 mt-1">
        <span className="text-xs" style={{ color: 'var(--hub-text-dim)' }}>
          {formatPercent(0)} {t('insights.dashboard.vsPrior')}
        </span>
      </div>
    );
  }
  const up = deltaPct > 0;
  return (
    <div className="flex items-center gap-1 mt-1">
      {up ? (
        <ArrowUpRight size={12} style={{ color: 'var(--delta-up)' }} />
      ) : (
        <ArrowDownRight size={12} style={{ color: 'var(--delta-down)' }} />
      )}
      <span className="text-xs" style={{ color: up ? 'var(--delta-up)' : 'var(--delta-down)' }}>
        {formatPercent(deltaPct)} {t('insights.dashboard.vsPrior')}
      </span>
    </div>
  );
};

const Dashboard: React.FC = (): React.JSX.Element => {
  const { t, i18n } = useTranslation();
  const { data, isLoading, isError, refetch } = useInsightsDashboard();
  const { data: seriesData } = useRevenueSeries();
  const { data: settings } = useInsightsSettings();
  const { data: work } = useWorkSummary();
  const { data: agingData } = useInvoiceAging();
  const currency = (settings?.defaultCurrency as Currency) || 'USD';

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
      id: 'revenue',

      label: t('insights.dashboard.cards.revenue'),
      value: data ? formatMoney(data.revenue.value, currency) : '-',
      deltaPct: data?.revenue.deltaPct ?? 0,
      icon: DollarSign,
      sparkline: true,
    },
    {
      id: 'customers',

      label: t('insights.dashboard.cards.customers'),
      value: data ? formatNumber(data.customers.value) : '-',
      deltaPct: data?.customers.deltaPct ?? 0,
      icon: Users,
      sparkline: false,
    },
    {
      id: 'orders',

      label: t('insights.dashboard.cards.orders'),
      value: data ? formatNumber(data.orders.value) : '-',
      deltaPct: data?.orders.deltaPct ?? 0,
      icon: ShoppingCart,
      sparkline: false,
    },
    {
      id: 'margin',

      label: t('insights.dashboard.cards.margin'),
      value: data ? `${data.margin.value.toFixed(1)}%` : '-',
      deltaPct: data?.margin.deltaPct ?? 0,
      icon: Percent,
      sparkline: false,
    },
  ];

  const agingBuckets = agingData?.buckets ?? [];
  const unpaidMinor = agingBuckets.reduce((sum, b) => sum + b.amountMinorUnits, 0);
  const overdueMinor = agingBuckets
    .filter((b) => b.bucket !== 'current')
    .reduce((sum, b) => sum + b.amountMinorUnits, 0);

  const serviceStats = [
    ...(work && work.totalTasks > 0
      ? [
          {
            id: 'tasks-done',
            label: t('insights.dashboard.service.tasksDone'),
            value: t('insights.dashboard.service.tasksDoneValue', {
              done: formatNumber(work.completedTasks),
              total: formatNumber(work.totalTasks),
            }),
            icon: CheckCircle2,
          },
          {
            id: 'tasks-overdue',
            label: t('insights.dashboard.service.overdueTasks'),
            value: formatNumber(work.overdueTasks),
            icon: AlertTriangle,
          },
        ]
      : []),
    ...(unpaidMinor > 0
      ? [
          {
            id: 'invoices-unpaid',
            label: t('insights.dashboard.service.unpaidInvoices'),
            value: formatMoney(unpaidMinor, currency),
            icon: FileText,
          },
          {
            id: 'invoices-overdue',
            label: t('insights.dashboard.service.overdueInvoices'),
            value: formatMoney(overdueMinor, currency),
            icon: Clock,
          },
        ]
      : []),
  ];
  const projects = (work?.projects ?? []).slice(0, 6);

  // Work and unpaid invoices count as activity too: a service business with
  // projects and invoices but no POS orders is not an empty business.
  const isEmpty =
    !!data &&
    data.revenue.value === 0 &&
    data.customers.value === 0 &&
    data.orders.value === 0 &&
    (work?.totalTasks ?? 0) === 0 &&
    unpaidMinor === 0;

  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      <HelmetContainer title={t('insights.dashboard.metaTitle')} />
      <div>
        <CaretAskBar />
      </div>

      <div>
        <h1 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.dashboard.title')}
        </h1>
        <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
          {formatToday(i18n.language)}
        </p>
      </div>

      {isLoading ? (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {stats.map((stat) => {
            const Icon = stat.icon;
            return (
              <div
                key={stat.id}
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
                    <Icon size={14} style={{ color: 'var(--accent-on-muted)' }} />
                  </span>
                </div>
                <p className="text-2xl font-bold truncate" style={{ color: 'var(--hub-text)' }}>
                  {stat.value}
                </p>
                <DeltaBadge deltaPct={stat.deltaPct} />
                {stat.sparkline && revenueSparkPoints.length > 1 && (
                  <div
                    style={{ height: 44, marginTop: 12 }}
                    aria-label={t('insights.charts.revenueSparkline')}
                  >
                    <Bar data={sparklineData} options={sparklineOptions} />
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {serviceStats.length > 0 && (
        <section aria-labelledby="insights-service-heading" className="space-y-3">
          <h2
            id="insights-service-heading"
            className="text-sm font-semibold"
            style={{ color: 'var(--hub-text)' }}
          >
            {t('insights.dashboard.service.title')}
          </h2>
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            {serviceStats.map((stat) => {
              const Icon = stat.icon;
              return (
                <div
                  key={stat.id}
                  className="rounded-xl p-5"
                  style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
                >
                  <div className="flex items-center justify-between mb-3">
                    <span
                      className="text-xs font-medium"
                      style={{ color: 'var(--hub-text-muted)' }}
                    >
                      {stat.label}
                    </span>
                    <span
                      className="flex items-center justify-center h-7 w-7 rounded-lg"
                      style={{ background: 'var(--accent-muted)' }}
                    >
                      <Icon size={14} style={{ color: 'var(--accent-on-muted)' }} />
                    </span>
                  </div>
                  <p className="text-2xl font-bold truncate" style={{ color: 'var(--hub-text)' }}>
                    {stat.value}
                  </p>
                </div>
              );
            })}
          </div>
        </section>
      )}

      {projects.length > 0 && (
        <section
          aria-labelledby="insights-projects-heading"
          className="rounded-xl p-5 space-y-3"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <h2
            id="insights-projects-heading"
            className="text-sm font-semibold"
            style={{ color: 'var(--hub-text)' }}
          >
            {t('insights.dashboard.service.projectsTitle')}
          </h2>
          <ul className="space-y-3">
            {projects.map((project) => {
              const pct = Math.min(100, Math.max(0, Math.round(project.progressPct)));
              return (
                <li key={project.id} className="space-y-1">
                  <div className="flex items-center justify-between gap-3 text-sm">
                    <span className="truncate" style={{ color: 'var(--hub-text)' }}>
                      {project.name}
                    </span>
                    <span className="shrink-0 text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                      {t('insights.dashboard.service.tasksDoneValue', {
                        done: formatNumber(project.completedTasks),
                        total: formatNumber(project.totalTasks),
                      })}
                    </span>
                  </div>
                  <div
                    role="progressbar"
                    aria-label={project.name}
                    aria-valuemin={0}
                    aria-valuemax={100}
                    aria-valuenow={pct}
                    className="h-1.5 rounded-full overflow-hidden"
                    style={{ background: 'var(--hub-border)' }}
                  >
                    <div
                      className="h-full rounded-full"
                      style={{ width: `${pct}%`, background: 'var(--accent)' }}
                    />
                  </div>
                </li>
              );
            })}
          </ul>
        </section>
      )}

      {!isLoading && isEmpty && (
        <EmptyState
          icon={BarChart3}
          title={t('insights.dashboard.emptyTitle')}
          description={t('insights.dashboard.emptyDescription')}
          action={refreshEmptyStateAction(t('common.refresh'), () => refetch())}
        />
      )}
    </div>
  );
};

export default Dashboard;
