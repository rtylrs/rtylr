import { EmptyState } from '../../../../shared/components/dash/EmptyState';
import React, { useMemo } from 'react';
import { useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ListTodo } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';

import { STATUS_COLORS, useFlowStore } from '../../shared/stores/flowStore';
import { useTasks } from '../tasks/useTasks';
import { STATUS_LABEL_KEYS } from '../../shared/lib/constantsFlow';
import { isOverdue, todayIso } from '../../shared/lib/dateUtils';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import ProjectViewTabs from './ProjectViewTabs';
import { useCurrentProject } from './useCurrentProject';
import { useProjectLabourCost } from '../../shared/hooks/queries/useProjectLabourCost';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import ProjectDashboardPanel from '../../../insights/shared/components/ProjectDashboardPanel';
import { useProjectStatuses } from './useStatuses';
import { useCycles } from '../cycles/useCycles';
import { computeProjectSummary } from './projectSummary';
import {
  FilteredListLink,
  HoursVsEstimateCard,
  NeedsAttentionCard,
  PriorityBreakdownCard,
  RecentChangeTiles,
} from './components/ProjectSummaryCards';

/**
 * ProjectDashboard - one-glance health view. The top is the Insights project
 * dashboard (T-567): progress, overdue tasks, hours logged and, for people who
 * may view finance invoices, the project's invoices, computed by the server.
 * Below it, computed client-side from the task list (T-591): what was
 * created, completed and updated in the last 7 days and what is due in the
 * next 7, a Needs attention list, estimated against logged hours, the status
 * and priority breakdowns and the workload per person. Every number opens the
 * List view filtered to the tasks behind it.
 *
 * Someone who cannot open the Insights dashboard (no Business Overview access,
 * or a private project they reach through their workspace rather than as a
 * project member) gets the four client-side tiles
 * (Total · Done · In progress · Overdue) instead of the server panel.
 */
const ProjectDashboard: React.FC = () => {
  const { t, i18n } = useTranslation();
  const { projectId = '' } = useParams({ strict: false });

  useProjectRealtime(projectId);

  const {
    project,
    isLoading: projectLoading,
    loadFailed: projectLoadFailed,
    refetch: refetchProjects,
  } = useCurrentProject(projectId);
  // HR->FLOW-01 (#837): live labour cost from HR's port, admin-gated server-side.
  const { data: labourCost } = useProjectLabourCost(projectId);
  const tasksQuery = useTasks({ projectId, compact: true });
  const { data: apiTasks = [] } = tasksQuery;
  const tasksLoadFailed = hasQueryFailed(tasksQuery);
  const tasksLoading = tasksQuery.isPending && !tasksLoadFailed;
  const { data: statuses = [] } = useProjectStatuses(projectId);
  const { data: cycles = [] } = useCycles(projectId);

  // The board and list drop archived tasks, so the summary does too: its
  // counts must equal the rows its links open.
  const summary = useMemo(() => {
    const tasks = apiTasks.map(apiTaskToTask).filter((task) => !task.archived);
    const loggedHours = new Map(apiTasks.map((t) => [String(t.id), Number(t.actualHours ?? 0)]));
    return computeProjectSummary({ tasks, loggedHours, statuses, cycles, today: todayIso() });
  }, [apiTasks, statuses, cycles]);

  const stats = useMemo(() => {
    const total = apiTasks.length;
    const done = apiTasks.filter((t) => t.status === 'done').length;
    const inProgress = apiTasks.filter((t) => t.status === 'in_progress').length;
    // A due date is a calendar day, not an instant. `new Date(dueDate) < now`
    // parses a bare "YYYY-MM-DD" as UTC midnight, so every task due today
    // counted as overdue from 00:00 UTC onward (rtylr#1077) and contradicted
    // the task row, My tasks and the project calendar. `isOverdue` is the one
    // shared definition of "finished means done or cancelled".
    const overdue = apiTasks.filter((t) => isOverdue(t.dueDate, t.status)).length;

    const statusCounts: Record<string, number> = {
      todo: 0,
      in_progress: 0,
      in_review: 0,
      done: 0,
      cancelled: 0,
      backlog: 0,
    };
    apiTasks.forEach((t) => {
      statusCounts[t.status] = (statusCounts[t.status] ?? 0) + 1;
    });

    return { total, done, inProgress, overdue, statusCounts };
  }, [apiTasks]);

  if (projectLoading && !project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        Loading project…
      </div>
    );
  }

  // A failed request is not a missing project, and the sentence has to differ:
  // this branch used to render the hardcoded English "Project not found." for
  // both, so a 500 told the user their project had gone.
  if (projectLoadFailed || tasksLoadFailed) {
    return (
      <ListLoadFailed
        onRetry={() => {
          void refetchProjects();
          void tasksQuery.refetch();
        }}
      />
    );
  }

  if (tasksLoading) {
    return (
      <div role="status" className="p-6 text-sm" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flowCommon.loading')}
      </div>
    );
  }

  if (!project) {
    return (
      <div className="p-6 text-flow-body" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.projectNotFound')}
      </div>
    );
  }

  const progress = stats.total ? Math.round((stats.done / stats.total) * 100) : 0;

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={t('flow.projectDashboard.pageTitle')} />
      <ProjectViewTabs
        project={project}
        projectId={projectId}
        taskCount={apiTasks.length}
        onNewTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
      />

      <div className="flex-1 overflow-auto p-5 space-y-5 max-w-6xl w-full mx-auto">
        <ProjectDashboardPanel
          projectId={projectId}
          unavailableFallback={
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              <StatTile
                label={t('flow.projectDashboard.totalTasks')}
                value={stats.total}
                tint={project.color}
              />
              <StatTile
                label={t('flow.projectDashboard.done')}
                value={stats.done}
                tint="var(--flow-status-approved)"
                hint={t('flow.projectDashboard.percentComplete', { percent: progress })}
              />
              <StatTile
                label={t('flow.projectDashboard.inProgress')}
                value={stats.inProgress}
                tint={STATUS_COLORS.in_progress}
              />
              <StatTile
                label={t('flow.projectDashboard.overdue')}
                value={stats.overdue}
                tint="var(--flow-danger-fg)"
              />
            </div>
          }
        />

        {labourCost && Number(labourCost.costMinorUnits ?? 0) > 0 && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <StatTile
              label={t('flow.projectDashboard.labourCost')}
              value={Number(labourCost.costMinorUnits ?? 0) / 100}
              tint={STATUS_COLORS.in_progress}
              hint={
                Number(labourCost.uncostedMinutes ?? 0) > 0
                  ? t('flow.projectDashboard.partialCost')
                  : undefined
              }
            />
          </div>
        )}

        <RecentChangeTiles projectId={projectId} recent={summary.recent} />

        <div className="grid gap-5 md:grid-cols-2">
          <NeedsAttentionCard projectId={projectId} attention={summary.attention} />
          {summary.hours.estimated > 0 && <HoursVsEstimateCard hours={summary.hours} />}
        </div>

        {/* Status breakdown bar */}
        <Card title={t('flow.projectDashboard.statusBreakdown')}>
          {stats.total === 0 ? (
            <EmptyState
              icon={ListTodo}
              title={t('flow.projectDashboard.noTasksYet')}
              description={t('flow.projectDashboard.noTasksYetDescription')}
              action={{
                label: t('flow.projectDashboard.addTaskCta'),
                onClick: () => useFlowStore.getState().setQuickCreateOpen(true),
              }}
              size="compact"
            />
          ) : (
            <>
              <div
                className="flex h-6 rounded-md overflow-hidden"
                style={{ background: 'var(--flow-hover)' }}
              >
                {(
                  ['done', 'in_progress', 'in_review', 'todo', 'cancelled', 'backlog'] as const
                ).map((status) => {
                  const n = stats.statusCounts[status] ?? 0;
                  if (!n) return null;
                  return (
                    <div
                      key={status}
                      style={{
                        width: `${(n / stats.total) * 100}%`,
                        background: STATUS_COLORS[status],
                      }}
                      title={`${status}: ${n}`}
                    />
                  );
                })}
              </div>
              <div className="flex flex-wrap gap-3 mt-3 text-flow-meta">
                {(
                  ['done', 'in_progress', 'in_review', 'todo', 'cancelled', 'backlog'] as const
                ).map((status) => {
                  const content = (
                    <>
                      <span
                        className="h-2 w-2 rounded-full"
                        style={{ background: STATUS_COLORS[status] }}
                      />
                      {t(STATUS_LABEL_KEYS[status])}
                      <span style={{ color: 'var(--flow-text-dim)' }}>
                        ({stats.statusCounts[status] ?? 0})
                      </span>
                    </>
                  );
                  const filters = summary.statusFilters[status];
                  // A legend entry opens the list on the project statuses
                  // that map to it, when the project has any.
                  return filters && stats.statusCounts[status] ? (
                    <FilteredListLink
                      key={status}
                      projectId={projectId}
                      filters={filters}
                      className="flex items-center gap-1.5 rounded-sm px-1.5 py-0.5 transition-colors hover:bg-[var(--flow-hover)]"
                      style={{ color: 'var(--flow-text-muted)' }}
                    >
                      {content}
                    </FilteredListLink>
                  ) : (
                    <span
                      key={status}
                      className="flex items-center gap-1.5 px-1.5 py-0.5"
                      style={{ color: 'var(--flow-text-muted)' }}
                    >
                      {content}
                    </span>
                  );
                })}
              </div>
            </>
          )}
        </Card>

        <PriorityBreakdownCard projectId={projectId} priorities={summary.priorities} />

        {/* Workload per person: share of the project's tasks, open tasks and
            the hours estimated on them. A row opens that person's open tasks. */}
        <Card title={t('flow.projectDashboard.byAssignee')}>
          {summary.workload.length === 0 ? (
            <EmptyState
              icon={ListTodo}
              title={t('flow.projectDashboard.noTasksYet')}
              action={{
                label: t('flow.projectDashboard.addTaskCta'),
                onClick: () => useFlowStore.getState().setQuickCreateOpen(true),
              }}
              size="compact"
            />
          ) : (
            <div className="space-y-2">
              {summary.workload.slice(0, 8).map((a) => {
                const label = a.name || t('flow.projectDashboard.unassigned');
                const detail = t('flow.projectDashboard.summary.workloadDetail', {
                  open: a.open,
                  hours: new Intl.NumberFormat(i18n.language, {
                    maximumFractionDigits: 1,
                  }).format(a.openEstimateHours),
                });
                const row = (
                  <>
                    <span
                      className="text-flow-meta w-[128px] shrink-0 truncate"
                      style={{ color: 'var(--flow-text-muted)' }}
                    >
                      {label}
                    </span>
                    <div
                      className="flex-1 h-1.5 rounded-full overflow-hidden"
                      style={{ background: 'var(--flow-hover)' }}
                    >
                      <div
                        className="h-full rounded-full"
                        style={{
                          width: `${(a.total / stats.total) * 100}%`,
                          background: project.color,
                        }}
                      />
                    </div>
                    <span
                      className="hidden sm:inline text-flow-meta tabular-nums shrink-0"
                      style={{ color: 'var(--flow-text-muted)' }}
                      data-workload-detail={a.key}
                    >
                      {detail}
                    </span>
                    <span
                      className="text-flow-meta font-mono tabular-nums w-8 shrink-0 text-end"
                      style={{ color: 'var(--flow-text-dim)' }}
                    >
                      {a.total}
                    </span>
                  </>
                );
                return a.filters ? (
                  <FilteredListLink
                    key={a.key}
                    projectId={projectId}
                    filters={a.filters}
                    aria-label={`${label}: ${detail}`}
                    className="flex items-center gap-3 rounded-sm px-1 py-0.5 transition-colors hover:bg-[var(--flow-hover)]"
                  >
                    {row}
                  </FilteredListLink>
                ) : (
                  <div key={a.key} className="flex items-center gap-3 px-1 py-0.5">
                    {row}
                  </div>
                );
              })}
            </div>
          )}
        </Card>
      </div>
    </div>
  );
};

const StatTile: React.FC<{
  label: string;
  value: number;
  tint: string;
  hint?: string;
}> = ({ label, value, tint, hint }) => (
  <div
    className="rounded-md p-3"
    style={{
      background: 'var(--flow-card)',
      border: '1px solid var(--flow-border)',
    }}
  >
    <div
      className="text-flow-meta font-semibold uppercase tracking-wider mb-1"
      style={{ color: 'var(--flow-text-dim)' }}
    >
      {label}
    </div>
    <div className="flex items-baseline gap-2">
      <span
        className="text-flow-screen font-bold tabular-nums"
        style={{ color: 'var(--flow-text)' }}
      >
        {value}
      </span>
      {hint && (
        <span className="text-flow-meta" style={{ color: tint }}>
          {hint}
        </span>
      )}
    </div>
  </div>
);

const Card: React.FC<{ title: string; children: React.ReactNode }> = ({ title, children }) => (
  <section
    className="rounded-md p-4"
    style={{
      background: 'var(--flow-card)',
      border: '1px solid var(--flow-border)',
    }}
  >
    <h3
      className="text-flow-meta font-semibold uppercase tracking-wider mb-3"
      style={{ color: 'var(--flow-text-dim)' }}
    >
      {title}
    </h3>
    {children}
  </section>
);

export default ProjectDashboard;
