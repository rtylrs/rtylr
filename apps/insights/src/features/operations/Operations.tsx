import React from 'react';
import { HelmetContainer } from '@voxire/components';
import { motion } from 'motion/react';
import { Users, Calendar, Clock, Activity } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  useOperationsKPIs,
  useHeadcount,
  useShiftMetrics,
  useTimeOff,
  useUtilization,
} from '../../shared/hooks/queries/useInsightsOperations';
import { formatNumber, formatPercent } from '../../shared/lib/format';
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

const Operations: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: kpis, isLoading: kpisLoading, isError, refetch } = useOperationsKPIs();
  const { data: headcount, isLoading: headcountLoading } = useHeadcount();
  const { data: shifts, isLoading: shiftsLoading } = useShiftMetrics();
  const { data: timeOff, isLoading: timeOffLoading } = useTimeOff();
  const { data: utilization, isLoading: utilLoading } = useUtilization();

  if (isError) {
    return <ErrorState message={t('insights.errors.generic')} onRetry={refetch} />;
  }

  const stats = [
    {
      label: t('insights.operations.kpis.activeStaff'),
      value: formatNumber(kpis?.activeStaff ?? 0),
      icon: Users,
    },
    {
      label: t('insights.operations.kpis.shiftsThisWeek'),
      value: formatNumber(kpis?.shiftsThisWeek ?? 0),
      icon: Calendar,
    },
    {
      label: t('insights.operations.kpis.pendingTimeOff'),
      value: formatNumber(kpis?.pendingTimeOff ?? 0),
      icon: Clock,
    },
    {
      label: t('insights.operations.kpis.openLeaveRequests'),
      value: formatNumber(kpis?.openLeaveRequests ?? 0),
      icon: Activity,
    },
  ];

  return (
    <motion.div
      variants={containerVariants as any}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <HelmetContainer title="Operations · rtylr" />
      <motion.div variants={itemVariants as any}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          {t('insights.operations.title')}
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
          {stats.map((stat) => {
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

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <motion.div
          variants={itemVariants as any}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
            {t('insights.operations.headcount')}
          </h3>
          {headcountLoading ? (
            <SkeletonRow cols={2} />
          ) : !headcount ? (
            <EmptyState title="" description={t('insights.operations.empty')} />
          ) : (
            <div className="space-y-3">
              <div className="flex items-baseline gap-3">
                <span className="text-3xl font-bold" style={{ color: 'var(--hub-text)' }}>
                  {formatNumber(headcount.activeStaff)}
                </span>
                <span
                  className="text-xs"
                  style={{
                    color: headcount.momDelta >= 0 ? 'var(--delta-up)' : 'var(--delta-down)',
                  }}
                >
                  {formatPercent(headcount.momDeltaPct)} MoM
                </span>
              </div>
            </div>
          )}
        </motion.div>

        <motion.div
          variants={itemVariants as any}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
            {t('insights.operations.utilization')}
          </h3>
          {utilLoading ? (
            <SkeletonRow cols={2} />
          ) : !utilization ? (
            <EmptyState title="" description={t('insights.operations.empty')} />
          ) : (
            <div className="space-y-2">
              <p className="text-3xl font-bold" style={{ color: 'var(--hub-text)' }}>
                {utilization.utilizationPct.toFixed(1)}%
              </p>
              <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
                {utilization.scheduledHours.toFixed(1)} / {utilization.availableHours.toFixed(1)}{' '}
                hours
              </p>
            </div>
          )}
        </motion.div>

        <motion.div
          variants={itemVariants as any}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
            {t('insights.operations.shifts')}
          </h3>
          {shiftsLoading ? (
            <SkeletonRow cols={2} />
          ) : !shifts ? (
            <EmptyState title="" description={t('insights.operations.empty')} />
          ) : (
            <div className="space-y-2 text-xs" style={{ color: 'var(--hub-text)' }}>
              <div className="flex justify-between">
                <span style={{ color: 'var(--hub-text-muted)' }}>Scheduled</span>
                <span className="font-semibold">{formatNumber(shifts.scheduled)}</span>
              </div>
              <div className="flex justify-between">
                <span style={{ color: 'var(--hub-text-muted)' }}>Completed</span>
                <span className="font-semibold">{formatNumber(shifts.completed)}</span>
              </div>
              <div className="flex justify-between">
                <span style={{ color: 'var(--hub-text-muted)' }}>No-shows</span>
                <span className="font-semibold">{formatNumber(shifts.noShows)}</span>
              </div>
              <div className="flex justify-between">
                <span style={{ color: 'var(--hub-text-muted)' }}>Completion</span>
                <span className="font-semibold">{shifts.completionRatePct.toFixed(1)}%</span>
              </div>
            </div>
          )}
        </motion.div>

        <motion.div
          variants={itemVariants as any}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
            {t('insights.operations.timeOff')}
          </h3>
          {timeOffLoading ? (
            <SkeletonRow cols={2} />
          ) : !timeOff ? (
            <EmptyState title="" description={t('insights.operations.empty')} />
          ) : (
            <div className="space-y-2 text-xs" style={{ color: 'var(--hub-text)' }}>
              <div className="flex justify-between">
                <span style={{ color: 'var(--hub-text-muted)' }}>Pending</span>
                <span className="font-semibold">{formatNumber(timeOff.pending)}</span>
              </div>
              <div className="flex justify-between">
                <span style={{ color: 'var(--hub-text-muted)' }}>Approved</span>
                <span className="font-semibold">{formatNumber(timeOff.approved)}</span>
              </div>
            </div>
          )}
        </motion.div>
      </div>
    </motion.div>
  );
};

export default Operations;
