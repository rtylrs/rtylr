import { EmptyState } from '../../../../shared/components/dash/EmptyState';
import React, { useMemo } from 'react';
import { useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ListTodo } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';

import { STATUS_COLORS, useFlowStore } from '../../shared/stores/flowStore';
import { useTasks } from '../tasks/useTasks';
import { STATUS_LABEL_KEYS } from '../../shared/lib/constantsFlow';
import { isOverdue } from '../../shared/lib/dateUtils';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import ProjectViewTabs from './ProjectViewTabs';
import { useCurrentProject } from './useCurrentProject';
import { useProjectLabourCost } from '../../shared/hooks/queries/useProjectLabourCost';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';

/**
 * ProjectDashboard - one-glance health view. Four tiles across the top:
 *   Total · Done · In progress · Overdue
 * Then a status breakdown bar and a grouped-by-assignee distribution card.
 *
 * Everything is computed client-side from the task list so the page renders
 * instantly; once the backend ships a `/flow/projects/{id}/stats` endpoint
 * we can swap to the server-computed values for correctness at scale.
 */
const ProjectDashboard: React.FC = () => {
  const { t } = useTranslation();
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
  const { data: apiTasks = [] } = useTasks({ projectId, compact: true });

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

    const assigneeTally = new Map<string, { name: string; count: number }>();
    apiTasks.forEach((t) => {
      const name = t.assignee || 'Unassigned';
      const prev = assigneeTally.get(name) ?? { name, count: 0 };
      prev.count += 1;
      assigneeTally.set(name, prev);
    });
    const byAssignee = Array.from(assigneeTally.values()).sort((a, b) => b.count - a.count);

    return { total, done, inProgress, overdue, statusCounts, byAssignee };
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
  if (projectLoadFailed) {
    return <ListLoadFailed onRetry={() => void refetchProjects()} />;
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

      <div className="flex-1 overflow-auto p-5 space-y-5 max-w-5xl w-full mx-auto">
        {/* Stat tiles */}
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
          {labourCost && Number(labourCost.costMinorUnits ?? 0) > 0 && (
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
          )}
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
                ).map((status) => (
                  <span
                    key={status}
                    className="flex items-center gap-1.5"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    <span
                      className="h-2 w-2 rounded-full"
                      style={{ background: STATUS_COLORS[status] }}
                    />
                    {t(STATUS_LABEL_KEYS[status])}
                    <span style={{ color: 'var(--flow-text-dim)' }}>
                      ({stats.statusCounts[status] ?? 0})
                    </span>
                  </span>
                ))}
              </div>
            </>
          )}
        </Card>

        {/* Assignee distribution */}
        <Card title={t('flow.projectDashboard.byAssignee')}>
          {stats.byAssignee.length === 0 ? (
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
              {stats.byAssignee.slice(0, 8).map((a) => (
                <div key={a.name} className="flex items-center gap-3">
                  <span
                    className="text-flow-meta w-[128px] truncate"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    {a.name}
                  </span>
                  <div
                    className="flex-1 h-1.5 rounded-full overflow-hidden"
                    style={{ background: 'var(--flow-hover)' }}
                  >
                    <div
                      className="h-full rounded-full"
                      style={{
                        width: `${(a.count / stats.total) * 100}%`,
                        background: project.color,
                      }}
                    />
                  </div>
                  <span
                    className="text-flow-meta font-mono tabular-nums w-8 text-right"
                    style={{ color: 'var(--flow-text-dim)' }}
                  >
                    {a.count}
                  </span>
                </div>
              ))}
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
