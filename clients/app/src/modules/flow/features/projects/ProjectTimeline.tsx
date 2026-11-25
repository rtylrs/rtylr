import { EmptyState } from '../../../../shared/components/dash/EmptyState';
import React, { useMemo, useState } from 'react';
import { useNavigate, useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { CalendarClock } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';

import { useFlowStore } from '../../shared/stores/flowStore';
import { useTasks } from '../tasks/useTasks';
import { useMilestones } from '../milestones/useMilestones';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import { useOpenTask } from '../../shared/lib/useOpenTask';
import { isOverdue, todayIso } from '../../shared/lib/dateUtils';
import { formatShortDate } from '../../shared/lib/utils';
import FlowSelect from '../../shared/components/FlowSelect';
import ProjectViewTabs from './ProjectViewTabs';
import { useCurrentProject } from './useCurrentProject';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import {
  buildBars,
  daysBetween,
  dependencyArrows,
  flattenRows,
  groupBars,
  timelineWindow,
  type TimelineGroupBy,
  type TimelineTask,
} from './timeline/timelineModel';

const LABEL_WIDTH = 260;
const ROW_HEIGHT = 32;
const ZOOM_WIDTH = { days: 36, weeks: 14 } as const;
type Zoom = keyof typeof ZOOM_WIDTH;

const DONE_STATUSES = new Set(['done', 'cancelled']);

/**
 * Project Timeline (T-342): one bar per task from its start date to its due
 * date, grouped by assignee or milestone, with an arrow from each blocking
 * task to the task it blocks. Tasks with neither date are listed above the
 * grid. Dates are edited in the task panel or the List; the grid is read-only.
 */
const ProjectTimeline: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const openTask = useOpenTask();
  const { projectId = '' } = useParams({ strict: false });
  const [groupBy, setGroupBy] = useState<TimelineGroupBy>('none');
  const [zoom, setZoom] = useState<Zoom>('days');

  useProjectRealtime(projectId);

  const {
    project,
    isLoading: projectLoading,
    loadFailed: projectLoadFailed,
    refetch: refetchProjects,
  } = useCurrentProject(projectId);
  const tasksQuery = useTasks({ projectId, activeOnly: true, compact: true });
  const { data: apiTasks = [] } = tasksQuery;
  const { data: milestones = [] } = useMilestones(projectId);
  const tasksLoadFailed = hasQueryFailed(tasksQuery);
  const tasksLoading = tasksQuery.isPending && !tasksLoadFailed;
  const today = todayIso();

  const tasks = useMemo(
    (): TimelineTask[] =>
      apiTasks
        .filter((task) => !task.archived && !task.parentTaskId)
        .map((task) => ({
          id: String(task.id),
          title: task.title,
          identifier: task.identifier,
          startDate: task.startDate || undefined,
          dueDate: task.dueDate || undefined,
          done: DONE_STATUSES.has(task.status),
          assigneeId: task.assigneeId ? String(task.assigneeId) : undefined,
          assigneeName: task.assigneeName || undefined,
          milestoneId: task.milestoneId ? String(task.milestoneId) : undefined,
          dependencies: (task.dependencies ?? []).map(String),
        })),
    [apiTasks]
  );
  const days = useMemo(() => timelineWindow(tasks, today), [tasks, today]);
  const groupLabels = useMemo(() => {
    const labels = new Map<string, string>();
    if (groupBy === 'assignee') {
      for (const task of tasks) {
        if (task.assigneeId) labels.set(task.assigneeId, task.assigneeName ?? task.assigneeId);
      }
    } else if (groupBy === 'milestone') {
      for (const m of milestones) labels.set(String(m.id), m.name);
    }
    return labels;
  }, [groupBy, tasks, milestones]);
  const rows = useMemo(() => {
    const bars = buildBars(tasks, days[0]);
    const noneLabel =
      groupBy === 'milestone'
        ? t('flow.projectTimeline.noMilestone')
        : t('flow.projectTimeline.unassigned');
    return flattenRows(groupBars(bars, groupBy, groupLabels, noneLabel), groupBy !== 'none');
  }, [tasks, days, groupBy, groupLabels, t]);
  const arrows = useMemo(() => dependencyArrows(rows), [rows]);
  const unscheduled = tasks.filter((task) => !task.startDate && !task.dueDate);

  if (projectLoading && !project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flowCommon.loading')}
      </div>
    );
  }

  // A failed request is not a missing project, and the sentence has to differ.
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

  const colWidth = ZOOM_WIDTH[zoom];
  const gridWidth = colWidth * days.length;
  const todayIndex = daysBetween(days[0], today);
  const milestoneMarks = milestones
    .map((m) => ({ id: String(m.id), name: m.name, day: m.targetDate?.slice(0, 10) }))
    .filter((m): m is { id: string; name: string; day: string } => Boolean(m.day))
    .map((m) => ({ ...m, index: daysBetween(days[0], m.day) }))
    .filter((m) => m.index >= 0 && m.index < days.length);

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={t('flow.projectTimeline.pageTitle')} />
      <ProjectViewTabs
        project={project}
        projectId={projectId}
        taskCount={apiTasks.length}
        onNewTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
        rightSlot={
          <>
            <div className="w-[180px]">
              <FlowSelect
                label={t('flow.projectTimeline.groupBy')}
                value={groupBy}
                options={(['none', 'assignee', 'milestone'] as const).map((value) => ({
                  value,
                  label: t(`flow.projectTimeline.group.${value}`),
                }))}
                onChange={(value) => setGroupBy(value as TimelineGroupBy)}
              />
            </div>
            <div
              className="inline-flex rounded-md p-0.5"
              role="group"
              aria-label={t('flow.projectTimeline.zoom')}
              style={{ border: '1px solid var(--flow-border-control)' }}
            >
              {(['days', 'weeks'] as const).map((value) => (
                <button
                  key={value}
                  type="button"
                  aria-pressed={zoom === value}
                  onClick={() => setZoom(value)}
                  className="px-2 py-0.5 rounded-sm text-flow-meta"
                  style={{
                    background: zoom === value ? 'var(--accent-muted)' : 'transparent',
                    color: 'var(--flow-text)',
                  }}
                >
                  {t(`flow.projectTimeline.zoomLevel.${value}`)}
                </button>
              ))}
            </div>
          </>
        }
      />

      {unscheduled.length > 0 && (
        <div
          className="flex flex-wrap items-center gap-1 px-4 py-2"
          style={{ borderBottom: '1px solid var(--flow-border)' }}
        >
          <span className="text-flow-meta me-1" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flow.projectTimeline.unscheduled', { count: unscheduled.length })}
          </span>
          {unscheduled.slice(0, 12).map((task) => (
            <button
              key={task.id}
              type="button"
              onClick={() => openTask(task.id)}
              className="text-flow-meta px-1.5 py-0.5 rounded-sm hover:underline"
              style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-muted)' }}
            >
              {task.title}
            </button>
          ))}
          {unscheduled.length > 12 && (
            <span className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
              {t('flow.projectTimeline.more', { count: unscheduled.length - 12 })}
            </span>
          )}
        </div>
      )}

      {/* Time runs left to right in Arabic too, and the scroller has to be
          left to right as well or it opens scrolled to the far end. The labels
          inside still follow the reading direction. */}
      <div className="min-h-0 flex-1 overflow-auto" style={{ direction: 'ltr' }}>
        {rows.length === 0 ? (
          <EmptyState
            icon={CalendarClock}
            title={t('flow.projectTimeline.emptyTitle')}
            description={t('flow.projectTimeline.emptyDescription')}
            action={{
              label: t('flow.projectTimeline.goToBoardCta'),
              onClick: () => navigate({ to: '/flow/p/$projectId/list', params: { projectId } }),
            }}
          />
        ) : (
          <div className="inline-block min-w-full">
            <div
              className="sticky top-0 z-10 flex"
              style={{
                background: 'var(--flow-surface)',
                borderBottom: '1px solid var(--flow-border)',
              }}
            >
              <div
                className="sticky left-0 z-10 shrink-0 px-3 py-2 text-flow-meta font-semibold"
                dir="auto"
                style={{
                  width: LABEL_WIDTH,
                  background: 'var(--flow-surface)',
                  color: 'var(--flow-text-dim)',
                  borderRight: '1px solid var(--flow-border)',
                }}
              >
                {t('flow.projectTimeline.taskColumn')}
              </div>
              <div className="relative" style={{ width: gridWidth, height: 40 }}>
                {days.map((day, i) => {
                  const date = new Date(`${day}T00:00:00`);
                  const label = zoom === 'days' ? true : date.getDay() === 1 || i === 0;
                  if (!label) return null;
                  return (
                    <div
                      key={day}
                      className="absolute top-0 flex flex-col items-center justify-center text-flow-meta"
                      style={{
                        left: i * colWidth,
                        width: zoom === 'days' ? colWidth : colWidth * 7,
                        height: 40,
                        alignItems: zoom === 'days' ? 'center' : 'flex-start',
                        color: i === todayIndex ? 'var(--accent)' : 'var(--flow-text-dim)',
                      }}
                    >
                      {zoom === 'days' ? (
                        <>
                          <span>
                            {date.toLocaleDateString(i18n.language, { weekday: 'narrow' })}
                          </span>
                          <span className="font-semibold">{date.getDate()}</span>
                        </>
                      ) : (
                        <span className="font-semibold ps-1">
                          {formatShortDate(day, i18n.language)}
                        </span>
                      )}
                    </div>
                  );
                })}
                {milestoneMarks.map((m) => (
                  <span
                    key={m.id}
                    className="absolute bottom-0.5 h-2.5 w-2.5 rotate-45"
                    style={{
                      left: m.index * colWidth + colWidth / 2 - 5,
                      background: 'var(--flow-warn-fg)',
                    }}
                    title={`${m.name} · ${formatShortDate(m.day, i18n.language)}`}
                    aria-label={m.name}
                    role="img"
                  />
                ))}
              </div>
            </div>

            <div className="relative">
              {rows.map((row) =>
                row.kind === 'group' ? (
                  <div
                    key={`g-${row.key}`}
                    className="flex items-center"
                    style={{
                      height: ROW_HEIGHT,
                      background: 'var(--flow-hover)',
                      borderBottom: '1px solid var(--flow-border)',
                    }}
                  >
                    <div
                      className="sticky left-0 px-3 text-flow-meta font-semibold truncate"
                      dir="auto"
                      style={{ width: LABEL_WIDTH, color: 'var(--flow-text)' }}
                    >
                      {row.label} <span style={{ color: 'var(--flow-text-dim)' }}>{row.count}</span>
                    </div>
                  </div>
                ) : (
                  <div
                    key={row.bar.task.id}
                    className="flex items-stretch hover:bg-[var(--flow-hover)]"
                    style={{ height: ROW_HEIGHT, borderBottom: '1px solid var(--flow-border)' }}
                  >
                    <button
                      type="button"
                      onClick={() => openTask(row.bar.task.id)}
                      className="sticky left-0 z-[1] shrink-0 px-3 text-start text-flow-meta truncate"
                      dir="auto"
                      style={{
                        width: LABEL_WIDTH,
                        background: 'var(--flow-surface)',
                        color: row.bar.task.done ? 'var(--flow-text-dim)' : 'var(--flow-text)',
                        borderRight: '1px solid var(--flow-border)',
                      }}
                    >
                      {row.bar.task.identifier && (
                        <span className="me-1.5" style={{ color: 'var(--flow-text-dim)' }}>
                          {row.bar.task.identifier}
                        </span>
                      )}
                      {row.bar.task.title}
                    </button>
                    <div className="relative" style={{ width: gridWidth }}>
                      <button
                        type="button"
                        onClick={() => openTask(row.bar.task.id)}
                        className="absolute top-1.5 rounded-sm text-flow-meta flex items-center px-2 overflow-hidden whitespace-nowrap"
                        style={{
                          left: row.bar.offset * colWidth + 2,
                          width: Math.max(row.bar.span * colWidth - 4, 12),
                          height: ROW_HEIGHT - 12,
                          background: row.bar.task.done
                            ? 'var(--flow-border-control)'
                            : isOverdue(row.bar.end) && !row.bar.task.done
                              ? 'var(--flow-warn-fg)'
                              : project.color,
                          color: 'var(--flow-on-accent)',
                        }}
                        title={`${row.bar.task.title} · ${formatShortDate(row.bar.start, i18n.language)} → ${formatShortDate(row.bar.end, i18n.language)}`}
                        aria-label={t('flow.projectTimeline.barLabel', {
                          title: row.bar.task.title,
                          start: formatShortDate(row.bar.start, i18n.language),
                          end: formatShortDate(row.bar.end, i18n.language),
                        })}
                      >
                        <span className="truncate" dir="auto">
                          {zoom === 'days' || row.bar.span > 5 ? row.bar.task.title : ''}
                        </span>
                      </button>
                    </div>
                  </div>
                )
              )}

              {/* Today, then the dependency arrows, over the grid. */}
              <svg
                className="pointer-events-none absolute top-0"
                style={{ left: LABEL_WIDTH }}
                width={gridWidth}
                height={rows.length * ROW_HEIGHT}
                aria-hidden="true"
              >
                <defs>
                  <marker
                    id="timeline-arrow"
                    viewBox="0 0 8 8"
                    refX="7"
                    refY="4"
                    markerWidth="7"
                    markerHeight="7"
                    orient="auto-start-reverse"
                  >
                    <path d="M0,0 L8,4 L0,8 z" fill="var(--flow-text-muted)" />
                  </marker>
                </defs>
                {todayIndex >= 0 && todayIndex < days.length && (
                  <line
                    x1={todayIndex * colWidth + colWidth / 2}
                    x2={todayIndex * colWidth + colWidth / 2}
                    y1={0}
                    y2={rows.length * ROW_HEIGHT}
                    stroke="var(--accent)"
                    strokeDasharray="3 3"
                  />
                )}
                {arrows.map((arrow) => {
                  const x1 = arrow.fromDay * colWidth - 2;
                  const y1 = arrow.fromRow * ROW_HEIGHT + ROW_HEIGHT / 2;
                  const x2 = arrow.toDay * colWidth + 2;
                  const y2 = arrow.toRow * ROW_HEIGHT + ROW_HEIGHT / 2;
                  const elbow = Math.max(x1 + 8, Math.min(x2 - 8, x1 + 8));
                  const d =
                    x2 - 8 >= x1
                      ? `M${x1},${y1} H${elbow} V${y2} H${x2}`
                      : `M${x1},${y1} h8 V${(y1 + y2) / 2} H${x2 - 8} V${y2} H${x2}`;
                  return (
                    <path
                      key={`${arrow.fromTaskId}-${arrow.toTaskId}`}
                      d={d}
                      fill="none"
                      stroke="var(--flow-text-muted)"
                      strokeWidth={1.5}
                      markerEnd="url(#timeline-arrow)"
                    />
                  );
                })}
              </svg>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ProjectTimeline;
