import React, { useMemo } from 'react';
import { useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ListTodo } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useTasks } from '../tasks/useTasks';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import ProjectViewTabs from './ProjectViewTabs';
import { useCurrentProject } from './useCurrentProject';
import { useProjectLabourCost } from '../../shared/hooks/queries/useProjectLabourCost';

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
  const { workspaceId = PERSONAL_WORKSPACE_ID, projectId = '' } = useParams({ strict: false });

  useProjectRealtime(projectId);

  const { project, isLoading: projectLoading } = useCurrentProject(workspaceId, projectId);
  // HR->FLOW-01 (#837): live labour cost from HR's port, admin-gated server-side.
  const { data: labourCost } = useProjectLabourCost(projectId);
  const { data: apiTasks = [] } = useTasks({ projectId, compact: true });

  const stats = useMemo(() => {
    const now = new Date();
    const total = apiTasks.length;
    const done = apiTasks.filter((t) => t.status === 'done').length;
    const inProgress = apiTasks.filter((t) => t.status === 'in_progress').length;
    const overdue = apiTasks.filter(
      (t) =>
        t.dueDate && new Date(t.dueDate) < now && t.status !== 'done' && t.status !== 'cancelled'
    ).length;

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

  if (!project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        Project not found.
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
          <StatTile label="Total tasks" value={stats.total} tint={project.color} />
          <StatTile
            label="Done"
            value={stats.done}
            tint="var(--flow-success, #10b981)"
            hint={`${progress}% complete`}
          />
          <StatTile label="In progress" value={stats.inProgress} tint="var(--flow-info, #3b82f6)" />
          <StatTile label="Overdue" value={stats.overdue} tint="var(--flow-danger, #ef4444)" />
          {labourCost && Number(labourCost.costMinorUnits ?? 0) > 0 && (
            <StatTile
              label="Labour cost"
              value={Number(labourCost.costMinorUnits ?? 0) / 100}
              tint="var(--flow-info, #3b82f6)"
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
                  [
                    ['done', 'var(--flow-success, #10b981)'],
                    ['in_progress', 'var(--flow-info, #3b82f6)'],
                    ['in_review', '#a855f7'],
                    ['todo', 'var(--flow-text-dim)'],
                    ['cancelled', 'var(--flow-danger, #ef4444)'],
                    ['backlog', 'var(--flow-border)'],
                  ] as const
                ).map(([status, color]) => {
                  const n = stats.statusCounts[status] ?? 0;
                  if (!n) return null;
                  return (
                    <div
                      key={status}
                      style={{
                        width: `${(n / stats.total) * 100}%`,
                        background: color,
                      }}
                      title={`${status}: ${n}`}
                    />
                  );
                })}
              </div>
              <div className="flex flex-wrap gap-3 mt-3 text-meta">
                {(
                  [
                    ['done', 'var(--flow-success, #10b981)', 'Done'],
                    ['in_progress', 'var(--flow-info, #3b82f6)', 'In progress'],
                    ['in_review', '#a855f7', 'In review'],
                    ['todo', 'var(--flow-text-dim)', 'To do'],
                    ['cancelled', 'var(--flow-danger, #ef4444)', 'Cancelled'],
                    ['backlog', 'var(--flow-border)', 'Backlog'],
                  ] as const
                ).map(([status, color, label]) => (
                  <span
                    key={status}
                    className="flex items-center gap-1.5"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    <span className="h-2 w-2 rounded-full" style={{ background: color }} />
                    {label}
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
                    className="text-meta w-[128px] truncate"
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
                    className="text-meta font-mono tabular-nums w-8 text-right"
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
      className="text-meta font-semibold uppercase tracking-wider mb-1"
      style={{ color: 'var(--flow-text-dim)' }}
    >
      {label}
    </div>
    <div className="flex items-baseline gap-2">
      <span className="text-screen font-bold tabular-nums" style={{ color: 'var(--flow-text)' }}>
        {value}
      </span>
      {hint && (
        <span className="text-meta" style={{ color: tint }}>
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
      className="text-meta font-semibold uppercase tracking-wider mb-3"
      style={{ color: 'var(--flow-text-dim)' }}
    >
      {title}
    </h3>
    {children}
  </section>
);

export default ProjectDashboard;
