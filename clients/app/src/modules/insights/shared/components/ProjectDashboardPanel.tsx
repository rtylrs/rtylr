import React from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import {
  AlertTriangle,
  ArrowDownRight,
  ArrowUpRight,
  CheckCircle2,
  Clock,
  FileText,
  Hourglass,
  Timer,
  Wallet,
  type LucideIcon,
} from 'lucide-react';
import { requestErrorStatus, useProjectDashboard } from '../hooks/queries/useInsightsDashboard';
import { useInsightsSettings } from '../hooks/queries/useInsightsSettings';
import type { ProjectOverdueTask } from '../hooks/queries/types';
import { formatMoney, type Currency } from '../utils/money';
import { formatHours, formatNumber, formatPercent } from '../lib/format';
import { CardSkeleton } from './Skeleton';
import ErrorState from './ErrorState';

interface ProjectDashboardPanelProps {
  projectId: string;
  // Shown instead of the panel when the viewer cannot open this project's
  // dashboard: no Business Overview access (403), or a private project they
  // are not a member of (404). Without it the panel explains which.
  unavailableFallback?: React.ReactNode;
}

interface StatCardProps {
  label: string;
  value: string;
  icon: LucideIcon;
  detail?: string;
  deltaPct?: number;
  children?: React.ReactNode;
}

const DeltaLine = ({ deltaPct }: { deltaPct: number }): React.JSX.Element => {
  const { t } = useTranslation();
  const up = deltaPct > 0;
  const color =
    deltaPct === 0 ? 'var(--hub-text-dim)' : up ? 'var(--delta-up)' : 'var(--delta-down)';
  return (
    <div className="flex items-center gap-1 mt-1">
      {deltaPct > 0 && <ArrowUpRight size={12} style={{ color }} aria-hidden />}
      {deltaPct < 0 && <ArrowDownRight size={12} style={{ color }} aria-hidden />}
      <span className="text-xs" style={{ color }}>
        {formatPercent(deltaPct)} {t('insights.dashboard.vsPrior')}
      </span>
    </div>
  );
};

const StatCard = ({
  label,
  value,
  icon: Icon,
  detail,
  deltaPct,
  children,
}: StatCardProps): React.JSX.Element => (
  <div
    className="rounded-xl p-5 min-w-0"
    style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
  >
    <div className="flex items-center justify-between gap-2 mb-3">
      <span className="text-xs font-medium truncate" style={{ color: 'var(--hub-text-muted)' }}>
        {label}
      </span>
      <span
        className="flex items-center justify-center h-7 w-7 shrink-0 rounded-lg"
        style={{ background: 'var(--accent-muted)' }}
      >
        <Icon size={14} style={{ color: 'var(--accent-on-muted)' }} aria-hidden />
      </span>
    </div>
    <p className="text-2xl font-bold truncate" style={{ color: 'var(--hub-text)' }}>
      {value}
    </p>
    {detail && (
      <p className="text-xs mt-1" style={{ color: 'var(--hub-text-muted)' }}>
        {detail}
      </p>
    )}
    {deltaPct !== undefined && <DeltaLine deltaPct={deltaPct} />}
    {children}
  </div>
);

const Notice = ({ title, description }: { title: string; description?: string }) => (
  <div
    role="status"
    className="rounded-xl p-6 text-center"
    style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
  >
    <p className="text-sm font-semibold" style={{ color: 'var(--hub-text)' }}>
      {title}
    </p>
    {description && (
      <p className="text-sm mt-1" style={{ color: 'var(--hub-text-muted)' }}>
        {description}
      </p>
    )}
  </div>
);

// A due date is a calendar day: format it at local midnight so it never
// shifts a day, with Latin digits like the rest of the dashboard.
export const formatDueDate = (dueDate: string, language: string): string => {
  const date = new Date(`${dueDate}T00:00:00`);
  if (Number.isNaN(date.getTime())) return dueDate;
  const options: Intl.DateTimeFormatOptions = { year: 'numeric', month: 'short', day: 'numeric' };
  try {
    return date.toLocaleDateString(`${language}-u-nu-latn`, options);
  } catch {
    return date.toLocaleDateString('en-US', options);
  }
};

// The assignee after an overdue task's due date. The name is isolated with
// <bdi> so a Latin name reads correctly in Arabic and the reverse. An assigned
// task whose name is unknown shows nothing rather than claiming it is
// unassigned.
const OverdueTaskAssignee = ({ task }: { task: ProjectOverdueTask }): React.JSX.Element | null => {
  const { t } = useTranslation();
  const label =
    task.assigneeId === ''
      ? t('insights.projectDashboard.unassigned')
      : task.assigneeName !== ''
        ? task.assigneeName
        : null;
  if (label === null) return null;
  return (
    <>
      <span aria-hidden="true" style={{ color: 'var(--hub-text-dim)' }}>
        ·
      </span>
      <bdi className="truncate" style={{ color: 'var(--hub-text-muted)' }}>
        {label}
      </bdi>
    </>
  );
};

const OverdueTaskList = ({
  projectId,
  tasks,
  total,
}: {
  projectId: string;
  tasks: ProjectOverdueTask[];
  total: number;
}): React.JSX.Element => {
  const { t, i18n } = useTranslation();
  const more = Math.max(0, total - tasks.length);
  return (
    <section
      aria-labelledby="project-dashboard-overdue-heading"
      className="rounded-xl p-5 space-y-3"
      style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
    >
      <h2
        id="project-dashboard-overdue-heading"
        className="text-sm font-semibold"
        style={{ color: 'var(--hub-text)' }}
      >
        {t('insights.projectDashboard.overdueListTitle')}
      </h2>
      {tasks.length === 0 ? (
        <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
          {t('insights.projectDashboard.nothingOverdue')}
        </p>
      ) : (
        <ul className="divide-y" style={{ borderColor: 'var(--hub-border)' }}>
          {tasks.map((task) => (
            <li key={task.id} className="py-2 first:pt-0 last:pb-0">
              <Link
                to="/flow/p/$projectId/list"
                params={{ projectId }}
                search={{ openTask: task.id }}
                className="flex items-center justify-between gap-3 text-sm hover:underline"
              >
                <span className="truncate text-start" style={{ color: 'var(--hub-text)' }}>
                  {task.title}
                </span>
                <span className="flex shrink-0 items-center gap-1 text-xs min-w-0 max-w-[55%]">
                  <span className="shrink-0" style={{ color: 'var(--delta-down)' }}>
                    {t('insights.projectDashboard.dueOn', {
                      date: formatDueDate(task.dueDate, i18n.language),
                    })}
                  </span>
                  <OverdueTaskAssignee task={task} />
                </span>
              </Link>
            </li>
          ))}
        </ul>
      )}
      {more > 0 && (
        <p className="text-xs" style={{ color: 'var(--hub-text-muted)' }}>
          {t('insights.projectDashboard.moreOverdue', {
            count: more,
            formatted: formatNumber(more),
          })}
        </p>
      )}
    </section>
  );
};

/**
 * One Work project's dashboard: progress, overdue work, hours logged and,
 * when the viewer may see finance invoices, the invoices tagged with the
 * project. Shared by the Insights project page and the Work project
 * Dashboard view.
 */
const ProjectDashboardPanel: React.FC<ProjectDashboardPanelProps> = ({
  projectId,
  unavailableFallback,
}) => {
  const { t, i18n } = useTranslation();
  const { data, isLoading, isError, error, refetch } = useProjectDashboard(projectId);
  const { data: settings } = useInsightsSettings();
  const currency = (settings?.defaultCurrency as Currency) || 'USD';

  if (isError) {
    const status = requestErrorStatus(error);
    if ((status === 403 || status === 404) && unavailableFallback !== undefined) {
      return <>{unavailableFallback}</>;
    }
    if (status === 404) {
      return (
        <Notice
          title={t('insights.projectDashboard.notFoundTitle')}
          description={t('insights.projectDashboard.notFoundDescription')}
        />
      );
    }
    if (status === 403) {
      return <Notice title={t('insights.projectDashboard.noAccess')} />;
    }
    return <ErrorState message={t('insights.errors.generic')} onRetry={refetch} />;
  }

  if (isLoading || !data) {
    return (
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4" aria-busy="true">
        {Array.from({ length: 4 }).map((_, i) => (
          <CardSkeleton key={i} />
        ))}
      </div>
    );
  }

  const pct = Math.min(100, Math.max(0, Math.round(data.progressPct)));
  const invoices = data.invoices;

  return (
    <div className="space-y-6">
      <section aria-label={t('insights.projectDashboard.workTitle')}>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <StatCard
            label={t('insights.projectDashboard.progress')}
            value={t('insights.projectDashboard.progressValue', { percent: pct })}
            icon={CheckCircle2}
            detail={t('insights.projectDashboard.tasksDone', {
              done: formatNumber(data.completedTasks),
              total: formatNumber(data.totalTasks),
            })}
          >
            <div
              role="progressbar"
              aria-label={t('insights.projectDashboard.progress')}
              aria-valuemin={0}
              aria-valuemax={100}
              aria-valuenow={pct}
              className="h-1.5 mt-3 rounded-full overflow-hidden"
              style={{ background: 'var(--hub-border)' }}
            >
              <div
                className="h-full rounded-full"
                style={{ width: `${pct}%`, background: 'var(--accent)' }}
              />
            </div>
          </StatCard>
          <StatCard
            label={t('insights.projectDashboard.overdueTasks')}
            value={formatNumber(data.overdueTaskCount)}
            icon={AlertTriangle}
          />
          <StatCard
            label={t('insights.projectDashboard.hoursThisMonth')}
            value={t('insights.projectDashboard.hoursValue', {
              hours: formatHours(data.hoursThisMonth.value, i18n.language),
            })}
            icon={Timer}
            deltaPct={data.hoursThisMonth.deltaPct}
          />
          <StatCard
            label={t('insights.projectDashboard.hoursAllTime')}
            value={t('insights.projectDashboard.hoursValue', {
              hours: formatHours(data.allTimeMinutes, i18n.language),
            })}
            icon={Hourglass}
          />
        </div>
      </section>

      {invoices && (
        <section aria-labelledby="project-dashboard-invoices-heading" className="space-y-3">
          <h2
            id="project-dashboard-invoices-heading"
            className="text-sm font-semibold"
            style={{ color: 'var(--hub-text)' }}
          >
            {t('insights.projectDashboard.invoicesTitle')}
          </h2>
          {invoices.invoiceCount === 0 ? (
            <p className="text-sm" style={{ color: 'var(--hub-text-muted)' }}>
              {t('insights.projectDashboard.noInvoices')}
            </p>
          ) : (
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
              <StatCard
                label={t('insights.projectDashboard.invoiced')}
                value={formatMoney(invoices.invoiced, currency)}
                icon={FileText}
                detail={t('insights.projectDashboard.invoiceCount', {
                  count: invoices.invoiceCount,
                  formatted: formatNumber(invoices.invoiceCount),
                })}
              />
              <StatCard
                label={t('insights.projectDashboard.paid')}
                value={formatMoney(invoices.paid, currency)}
                icon={Wallet}
              />
              <StatCard
                label={t('insights.projectDashboard.outstanding')}
                value={formatMoney(invoices.outstanding, currency)}
                icon={Clock}
              />
              <StatCard
                label={t('insights.projectDashboard.overdueInvoices')}
                value={formatMoney(invoices.overdue, currency)}
                icon={AlertTriangle}
                detail={t('insights.projectDashboard.overdueInvoiceCount', {
                  count: invoices.overdueCount,
                  formatted: formatNumber(invoices.overdueCount),
                })}
              />
            </div>
          )}
        </section>
      )}

      <OverdueTaskList
        projectId={data.project.id || projectId}
        tasks={data.overdueTasks}
        total={data.overdueTaskCount}
      />
    </div>
  );
};

export default ProjectDashboardPanel;
