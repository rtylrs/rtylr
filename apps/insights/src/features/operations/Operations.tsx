import React from 'react';
import { motion } from 'motion/react';
import { Briefcase, Users, Clock, Calendar } from 'lucide-react';
import {
  useOperationsSummary,
  useStaffUtilization,
  useShiftCoverage,
  useTimeOffTrends,
  useDeptHeadcount,
  useHrDashboard,
} from './useOperations';
import { formatNumber, formatPercent } from '../../shared/lib/format';
import { CardSkeleton, SkeletonRow } from '../../shared/components/Skeleton';
import ErrorState from '../../shared/components/ErrorState';
import EmptyState from '../../shared/components/EmptyState';

const ACCENT = '#6366f1';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const itemVariants = {
  hidden: { opacity: 0, y: 12 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22, ease: 'easeOut' } },
};

const Operations: React.FC = () => {
  const { data: summary, isLoading: summaryLoading, isError, refetch } = useOperationsSummary();
  const { data: utilization, isLoading: utilLoading } = useStaffUtilization();
  const { data: shifts, isLoading: shiftsLoading } = useShiftCoverage();
  const { data: timeOff, isLoading: toLoading } = useTimeOffTrends();
  const { data: headcount, isLoading: hcLoading } = useDeptHeadcount();
  const { data: hrData, isLoading: hrLoading } = useHrDashboard();

  if (isError) {
    return <ErrorState message="Failed to load operations data." onRetry={refetch} />;
  }

  const isLoading = summaryLoading || hrLoading;

  const totalStaff = summary?.totalStaff ?? hrData?.totalEmployees ?? 0;
  const avgUtilization = utilization?.average ?? summary?.avgUtilization ?? 0;
  const shiftsCovered = shifts?.coveredPercent ?? summary?.shiftCoverage ?? 0;
  const pendingTimeOff = timeOff?.pending ?? summary?.pendingTimeOff ?? 0;

  const deptList: any[] = headcount?.data ?? headcount ?? [];
  const utilizationList: any[] = utilization?.data ?? utilization?.staff ?? [];
  const timeOffList: any[] = timeOff?.data ?? timeOff?.recent ?? [];

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="show"
      className="space-y-6 max-w-7xl"
    >
      <motion.div variants={itemVariants}>
        <h2 className="text-xl font-semibold" style={{ color: 'var(--hub-text)' }}>
          Operations
        </h2>
        <p className="text-sm mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
          Staff utilization, shifts, and department metrics
        </p>
      </motion.div>

      {isLoading ? (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <CardSkeleton key={i} />
          ))}
        </div>
      ) : (
        <motion.div variants={containerVariants} className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { label: 'Total Staff', value: formatNumber(totalStaff), icon: Users, color: ACCENT },
            {
              label: 'Avg Utilization',
              value: `${avgUtilization.toFixed(1)}%`,
              icon: Briefcase,
              color: '#10b981',
            },
            {
              label: 'Shift Coverage',
              value: `${shiftsCovered.toFixed(1)}%`,
              icon: Clock,
              color: '#f59e0b',
            },
            {
              label: 'Pending Time Off',
              value: formatNumber(pendingTimeOff),
              icon: Calendar,
              color: '#f87171',
            },
          ].map((stat) => {
            const Icon = stat.icon;
            return (
              <motion.div
                key={stat.label}
                variants={itemVariants}
                className="rounded-xl p-5"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
              >
                <div className="flex items-center justify-between mb-3">
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text-muted)' }}>
                    {stat.label}
                  </span>
                  <span
                    className="flex items-center justify-center h-7 w-7 rounded-lg"
                    style={{ background: `${stat.color}18` }}
                  >
                    <Icon size={14} style={{ color: stat.color }} />
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
        {/* Department Headcount */}
        <motion.div
          variants={itemVariants}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
            Department Headcount
          </h3>
          {hcLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <SkeletonRow key={i} cols={2} />
              ))}
            </div>
          ) : deptList.length === 0 ? (
            <EmptyState title="No departments" description="Department data will appear here." />
          ) : (
            <div className="space-y-1">
              {deptList.map((d: any, i: number) => (
                <div
                  key={d.id ?? i}
                  className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                  style={{ background: 'var(--hub-hover)' }}
                >
                  <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                    {d.name ?? d.department ?? 'Unknown'}
                  </span>
                  <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                    {formatNumber(d.count ?? d.headcount ?? 0)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </motion.div>

        {/* Staff Utilization */}
        <motion.div
          variants={itemVariants}
          className="rounded-xl p-5"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
        >
          <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
            Staff Utilization
          </h3>
          {utilLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <SkeletonRow key={i} cols={2} />
              ))}
            </div>
          ) : utilizationList.length === 0 ? (
            <EmptyState
              title="No utilization data"
              description="Staff utilization will appear here."
            />
          ) : (
            <div className="space-y-1">
              {utilizationList.map((s: any, i: number) => (
                <div
                  key={s.id ?? i}
                  className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                  style={{ background: 'var(--hub-hover)' }}
                >
                  <span
                    className="text-xs font-medium truncate"
                    style={{ color: 'var(--hub-text)' }}
                  >
                    {s.name ?? 'Unknown'}
                  </span>
                  <span
                    className="text-xs font-semibold"
                    style={{ color: (s.utilization ?? 0) >= 80 ? '#10b981' : '#f59e0b' }}
                  >
                    {(s.utilization ?? 0).toFixed(1)}%
                  </span>
                </div>
              ))}
            </div>
          )}
        </motion.div>
      </div>

      {/* Time Off Trends */}
      <motion.div
        variants={itemVariants}
        className="rounded-xl p-5"
        style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
      >
        <h3 className="text-sm font-semibold mb-4" style={{ color: 'var(--hub-text)' }}>
          Time Off Trends
        </h3>
        {toLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 3 }).map((_, i) => (
              <SkeletonRow key={i} cols={3} />
            ))}
          </div>
        ) : timeOffList.length === 0 ? (
          <EmptyState title="No time off data" description="Time off trends will appear here." />
        ) : (
          <div className="space-y-1">
            {timeOffList.map((t: any, i: number) => (
              <div
                key={t.id ?? i}
                className="flex items-center justify-between px-3 py-2.5 rounded-lg"
                style={{ background: 'var(--hub-hover)' }}
              >
                <span className="text-xs font-medium" style={{ color: 'var(--hub-text)' }}>
                  {t.month ?? t.period ?? `Period ${i + 1}`}
                </span>
                <span className="text-xs font-semibold" style={{ color: 'var(--hub-text)' }}>
                  {formatNumber(t.days ?? t.count ?? 0)} days
                </span>
              </div>
            ))}
          </div>
        )}
      </motion.div>
    </motion.div>
  );
};

export default Operations;
