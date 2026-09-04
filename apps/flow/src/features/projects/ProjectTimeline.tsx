import React, { useMemo } from 'react';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { CalendarClock } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import { EmptyState } from '@voxire/ui';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useTasks } from '../tasks/useTasks';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import ProjectViewTabs from './ProjectViewTabs';
import { useCurrentProject } from './useCurrentProject';

/**
 * ProjectTimeline - a lightweight Gantt-style view. Each task is drawn as
 * a bar spanning `start_date` → `due_date`. Tasks without dates are stacked
 * at the top in an "Unscheduled" row so they're visible and draggable in.
 *
 * Full drag-to-reschedule is a follow-up (needs a task-date update endpoint
 * that the frontend doesn't yet call). For now this is a read-only canvas.
 */
const ProjectTimeline: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { workspaceId = PERSONAL_WORKSPACE_ID, projectId = '' } = useParams({ strict: false });

  useProjectRealtime(projectId);

  const { project, isLoading: projectLoading } = useCurrentProject(workspaceId, projectId);
  const { data: apiTasks = [] } = useTasks({ projectId, activeOnly: true, compact: true });

  // Derive a 30-day horizon anchored on the earliest task start (or today)
  // so the bars render at reasonable zoom.
  const { days, anchor } = useMemo(() => {
    const today = new Date();
    const starts = apiTasks
      .map((t) => (t.dueDate ? new Date(t.dueDate) : null))
      .filter((d): d is Date => !!d && !Number.isNaN(d.getTime()));
    const earliest =
      starts.length > 0 ? new Date(Math.min(...starts.map((d) => d.getTime()))) : today;
    const anchor = earliest < today ? earliest : today;
    const list: Date[] = [];
    for (let i = 0; i < 30; i++) {
      const d = new Date(anchor);
      d.setDate(anchor.getDate() + i);
      list.push(d);
    }
    return { days: list, anchor };
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

  const scheduled = apiTasks.filter((t) => !!t.dueDate);
  const unscheduled = apiTasks.filter((t) => !t.dueDate);

  const colWidth = 40;
  const dayMs = 1000 * 60 * 60 * 24;

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={t('flow.projectTimeline.pageTitle')} />
      <ProjectViewTabs
        project={project}
        projectId={projectId}
        taskCount={apiTasks.length}
        onNewTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
      />

      <div className="min-h-0 flex-1 overflow-auto">
        <div className="inline-block min-w-full">
          {/* Date header */}
          <div
            className="sticky top-0 z-10 flex"
            style={{
              background: 'var(--flow-surface)',
              borderBottom: '1px solid var(--flow-border)',
            }}
          >
            <div
              className="shrink-0 px-3 py-2 text-meta font-semibold"
              style={{
                width: 240,
                color: 'var(--flow-text-dim)',
                borderRight: '1px solid var(--flow-border)',
              }}
            >
              Task
            </div>
            {days.map((d, i) => (
              <div
                key={i}
                className="shrink-0 flex flex-col items-center justify-center py-1.5"
                style={{
                  width: colWidth,
                  color:
                    d.toDateString() === new Date().toDateString()
                      ? 'var(--accent)'
                      : 'var(--flow-text-dim)',
                  borderRight: '1px solid var(--flow-border)',
                }}
              >
                <span className="text-meta">
                  {d.toLocaleDateString(undefined, { weekday: 'short' })[0]}
                </span>
                <span className="text-meta font-semibold">{d.getDate()}</span>
              </div>
            ))}
          </div>

          {/* Unscheduled row */}
          {unscheduled.length > 0 && (
            <div
              className="flex items-stretch"
              style={{ borderBottom: '1px solid var(--flow-border)' }}
            >
              <div
                className="shrink-0 px-3 py-1.5 text-meta"
                style={{
                  width: 240,
                  color: 'var(--flow-text-dim)',
                  borderRight: '1px solid var(--flow-border)',
                }}
              >
                Unscheduled ({unscheduled.length})
              </div>
              <div className="flex-1 px-2 py-1.5 flex flex-wrap gap-1">
                {unscheduled.slice(0, 10).map((t) => (
                  <span
                    key={t.id}
                    className="text-meta px-1.5 py-0.5 rounded-sm"
                    style={{
                      background: 'var(--flow-hover)',
                      color: 'var(--flow-text-muted)',
                    }}
                  >
                    {t.title}
                  </span>
                ))}
                {unscheduled.length > 10 && (
                  <span
                    className="text-meta px-1.5 py-0.5 rounded-sm"
                    style={{ color: 'var(--flow-text-dim)' }}
                  >
                    +{unscheduled.length - 10} more
                  </span>
                )}
              </div>
            </div>
          )}

          {/* Scheduled rows */}
          {scheduled.length === 0 ? (
            <EmptyState
              icon={CalendarClock}
              title={t('flow.projectTimeline.emptyTitle')}
              description={t('flow.projectTimeline.emptyDescription')}
              action={{
                label: t('flow.projectTimeline.goToBoardCta'),
                onClick: () =>
                  navigate({
                    to: '/p/$projectId/board',
                    params: { projectId },
                  }),
              }}
            />
          ) : (
            scheduled.map((t) => {
              const dueDate = new Date(t.dueDate!);
              // ApiTask's contract doesn't yet surface start_date, so treat
              // every scheduled task as a single-day bar for now. When the
              // proto adds start_date (and the backend sends it) we can
              // swap in `t.startDate ? new Date(t.startDate) : dueDate`.
              const startDate = dueDate;
              const offsetDays = Math.max(
                0,
                Math.round((startDate.getTime() - anchor.getTime()) / dayMs)
              );
              const spanDays = Math.max(
                1,
                Math.round((dueDate.getTime() - startDate.getTime()) / dayMs) + 1
              );
              const isOverdue = dueDate < new Date() && t.status !== 'done';
              const barColor = isOverdue ? 'var(--flow-warn-fg)' : project.color;

              return (
                <div
                  key={t.id}
                  className="flex items-stretch hover:bg-[var(--flow-hover)]"
                  style={{ borderBottom: '1px solid var(--flow-border)' }}
                >
                  <div
                    className="shrink-0 px-3 py-1.5 text-meta truncate"
                    style={{
                      width: 240,
                      color: 'var(--flow-text)',
                      borderRight: '1px solid var(--flow-border)',
                    }}
                  >
                    {t.title}
                  </div>
                  <div className="relative" style={{ width: colWidth * days.length, height: 28 }}>
                    <div
                      className="absolute top-1.5 h-5 rounded-sm text-meta flex items-center px-2 text-white overflow-hidden"
                      style={{
                        left: offsetDays * colWidth + 4,
                        width: Math.max(spanDays * colWidth - 8, 24),
                        background: barColor,
                      }}
                      title={`${t.title} · ${dueDate.toDateString()}`}
                    >
                      {t.title}
                    </div>
                  </div>
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
};

export default ProjectTimeline;
