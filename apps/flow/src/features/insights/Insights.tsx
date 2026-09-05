import React from 'react';
import { motion } from 'motion/react';
import { BarChart3, TrendingUp, Clock, CheckCircle2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useInsights } from './useInsights';
import { Skeleton, CardSkeleton } from '../../shared/components/Skeleton';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';

const Insights: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const insightsQuery = useInsights();
  const { data, isLoading } = insightsQuery;

  // A failed request leaves `data` undefined while isLoading is false, and the
  // guard below treats that as loading, so the skeletons never went away.
  if (hasQueryFailed(insightsQuery)) {
    return <ListLoadFailed onRetry={() => void insightsQuery.refetch()} />;
  }

  if (isLoading || !data) {
    return (
      <div className="flex flex-col h-full overflow-auto">
        <div
          className="px-6 py-4 shrink-0"
          style={{ borderBottom: '1px solid var(--flow-border)' }}
        >
          <Skeleton width={120} height={16} className="mb-1" />
          <Skeleton width={200} height={11} />
        </div>
        <div className="p-6 max-w-6xl space-y-6">
          <div className="grid grid-cols-4 gap-4">
            {[1, 2, 3, 4].map((i) => (
              <CardSkeleton key={i} />
            ))}
          </div>
          <CardSkeleton />
          <div className="grid grid-cols-2 gap-4">
            <CardSkeleton />
            <CardSkeleton />
          </div>
          <CardSkeleton />
        </div>
      </div>
    );
  }

  const maxThroughput = Math.max(...(data.throughput?.map((t) => t.completed) ?? [0]));

  return (
    <div className="flex flex-col h-full overflow-auto">
      <HelmetContainer title={`${t('common.insights.heading')} · Flow`} />
      <div className="px-6 py-4 shrink-0" style={{ borderBottom: '1px solid var(--flow-border)' }}>
        <h1 className="text-title font-semibold" style={{ color: 'var(--flow-text)' }}>
          {t('common.insights.heading')}
        </h1>
        <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
          {t('common.insights.subtitle')}
        </p>
      </div>

      <div className="p-6 max-w-6xl space-y-6">
        {/* KPI row */}
        <div className="grid grid-cols-4 gap-4">
          <Kpi
            icon={CheckCircle2}
            label={t('common.insights.kpi.completed')}
            value={String(data.doneTasks)}
            hint={t('common.insights.kpi.completedHint', { total: data.totalTasks })}
            color="var(--flow-status-approved)"
          />
          <Kpi
            icon={TrendingUp}
            label={t('common.insights.kpi.velocity')}
            value={`${Math.round(
              data.velocity.reduce((s, v) => s + v.completed, 0) / Math.max(1, data.velocity.length)
            )}`}
            hint={t('common.insights.kpi.velocityHint')}
            color="var(--flow-info-fg)"
          />
          <Kpi
            icon={Clock}
            label={t('common.insights.kpi.cycleTime')}
            value={`${data.cycleTime}d`}
            hint={t('common.insights.kpi.cycleTimeHint')}
            color="#f59e0b"
          />
          <Kpi
            icon={BarChart3}
            label={t('common.insights.kpi.throughput')}
            value={`${data.throughput.reduce((s, t) => s + t.completed, 0)}`}
            hint={t('common.insights.kpi.throughputHint')}
            color="#a855f7"
          />
        </div>

        {/* Burndown */}
        <ChartCard title={t('common.insights.burndown')}>
          {data.burndown.length === 0 ? (
            <EmptyChart onGoToCycles={() => navigate({ to: '/cycles' })} />
          ) : (
            <div className="relative h-[160px]">
              <svg width="100%" height="100%" viewBox="0 0 400 160" preserveAspectRatio="none">
                <defs>
                  <linearGradient id="bgrad" x1="0" x2="0" y1="0" y2="1">
                    <stop offset="0" stopColor="var(--accent)" stopOpacity="0.3" />
                    <stop offset="1" stopColor="var(--accent)" stopOpacity="0" />
                  </linearGradient>
                </defs>
                {(() => {
                  const max = Math.max(...data.burndown.map((b) => b.ideal));
                  const pts = data.burndown.map((b, i) => {
                    const x = (i / (data.burndown.length - 1)) * 400;
                    const y = 160 - (b.actual / max) * 150;
                    return `${x},${y}`;
                  });
                  const idealPts = data.burndown.map((b, i) => {
                    const x = (i / (data.burndown.length - 1)) * 400;
                    const y = 160 - (b.ideal / max) * 150;
                    return `${x},${y}`;
                  });
                  return (
                    <>
                      <polyline
                        points={idealPts.join(' ')}
                        fill="none"
                        stroke="rgba(255,255,255,0.15)"
                        strokeWidth="1.5"
                        strokeDasharray="4 4"
                      />
                      <polygon points={`0,160 ${pts.join(' ')} 400,160`} fill="url(#bgrad)" />
                      <polyline
                        points={pts.join(' ')}
                        fill="none"
                        stroke="var(--accent)"
                        strokeWidth="2"
                      />
                    </>
                  );
                })()}
              </svg>
            </div>
          )}
        </ChartCard>

        <div className="grid grid-cols-2 gap-4">
          <ChartCard title={t('common.insights.throughput')}>
            <div className="flex items-end gap-2 h-[160px]">
              {data.throughput.map((t) => (
                <div key={t.week} className="flex-1 flex flex-col items-center gap-1">
                  <motion.div
                    initial={{ height: 0 }}
                    animate={{ height: `${(t.completed / maxThroughput) * 100}%` }}
                    transition={{ duration: 0.4 }}
                    className="w-full rounded-t-sm"
                    style={{ background: '#4583C4' }}
                  />
                  <span className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
                    {t.week}
                  </span>
                </div>
              ))}
            </div>
          </ChartCard>

          <ChartCard title={t('common.insights.velocity')}>
            <div className="space-y-2">
              {data.velocity.map((v) => (
                <div key={v.cycle}>
                  <div className="flex items-center justify-between text-meta mb-1">
                    <span style={{ color: 'var(--flow-text-muted)' }}>{v.cycle}</span>
                    <span style={{ color: 'var(--flow-text)' }}>
                      {v.completed} / {v.total}
                    </span>
                  </div>
                  <div
                    className="h-2 rounded-full overflow-hidden"
                    style={{ background: 'var(--flow-hover)' }}
                  >
                    <motion.div
                      initial={{ width: 0 }}
                      animate={{ width: `${v.completionRate}%` }}
                      transition={{ duration: 0.5 }}
                      className="h-full"
                      style={{ background: 'var(--accent)' }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </ChartCard>
        </div>

        <ChartCard title={t('common.insights.timeByPerson')}>
          <div className="space-y-2">
            {data.timeByMember.map((m) => (
              <div key={m.id} className="flex items-center gap-3">
                <span className="text-meta w-[96px] shrink-0" style={{ color: 'var(--flow-text)' }}>
                  {m.name}
                </span>
                <div
                  className="flex-1 h-2 rounded-full overflow-hidden"
                  style={{ background: 'var(--flow-hover)' }}
                >
                  <div
                    className="h-full"
                    style={{
                      width: `${Math.min(100, (m.hours / 40) * 100)}%`,
                      background: '#a855f7',
                    }}
                  />
                </div>
                <span
                  className="text-meta w-12 text-right shrink-0"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {m.hours}h
                </span>
              </div>
            ))}
          </div>
        </ChartCard>
      </div>
    </div>
  );
};

const Kpi: React.FC<{
  icon: React.ComponentType<{ size?: number }>;
  label: string;
  value: string;
  hint: string;
  color: string;
}> = ({ icon: Icon, label, value, hint, color }) => (
  <div
    className="rounded-xl p-4"
    style={{
      background: 'var(--flow-card)',
      border: '1px solid var(--flow-border)',
    }}
  >
    <div className="flex items-center gap-1.5 mb-2">
      <Icon size={11} />
      <span
        className="text-meta uppercase tracking-wider font-semibold"
        style={{ color: 'var(--flow-text-dim)' }}
      >
        {label}
      </span>
    </div>
    <p className="text-screen font-bold" style={{ color }}>
      {value}
    </p>
    <p className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
      {hint}
    </p>
  </div>
);

const ChartCard: React.FC<{ title: string; children: React.ReactNode }> = ({ title, children }) => (
  <div
    className="rounded-xl p-5"
    style={{
      background: 'var(--flow-card)',
      border: '1px solid var(--flow-border)',
    }}
  >
    <h3
      className="text-meta font-semibold uppercase tracking-wider mb-3"
      style={{ color: 'var(--flow-text-muted)' }}
    >
      {title}
    </h3>
    {children}
  </div>
);

const EmptyChart: React.FC<{ onGoToCycles: () => void }> = ({ onGoToCycles }) => {
  const { t } = useTranslation();
  return (
    <div className="flex h-[160px] items-center justify-center">
      <EmptyState
        icon={BarChart3}
        title={t('common.insights.noData')}
        action={{
          label: t('common.insights.noDataCta'),
          onClick: onGoToCycles,
        }}
        size="compact"
      />
    </div>
  );
};

export default Insights;
