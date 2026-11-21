import React, { useMemo, useState } from 'react';
import { useOpenTask } from '../../shared/lib/useOpenTask';
import { motion, AnimatePresence } from 'motion/react';
import { X } from 'lucide-react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { useMediaQuery } from '@voxire/utils';
import { HelmetContainer } from '@voxire/components';
import { useTasks, useCreateTask } from '../tasks/useTasks';
import { useLabels } from '../tasks/useLabels';
import { useFlowStore } from '../../shared/stores/flowStore';
import PriorityIcon from '../../shared/components/PriorityIcon';
import type { ApiTask } from '../../shared/services/flowApi';
import { todayIso } from '../../shared/lib/dateUtils';
import { tasksByLocalDay, agendaDays, type MonthCell } from '../../shared/lib/monthGrid';
import { taskRange } from '../../shared/lib/calendarSpans';
import { useCanEditTaskIn } from '../../shared/lib/currentUser';
import { CalendarAgenda } from '../../shared/components/CalendarAgenda';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import { useCalendarView } from './useCalendarView';
import { useCalendarScheduling } from './useCalendarScheduling';
import { useCalendarPlanning } from './useCalendarPlanning';
import { PlanningGrid } from './PlanningGrid';
import { UnscheduledPanel, needsScheduling } from './UnscheduledPanel';
import { CalendarControls } from './CalendarControls';

const ACCENT_VAR = 'var(--accent)';

/**
 * The grid needs room; below that the month is a list.
 *
 * A min-width test rather than max-width, and it decides which layout is
 * *rendered* rather than which is hidden. Two reasons: a month grid plus a full
 * agenda is a lot of nodes to build and throw away on every render, and with
 * both mounted every task title exists twice in the DOM, which a test cannot
 * disambiguate and a screen reader only avoids because `display: none` happens
 * to prune it.
 */
const WIDE = '(min-width: 768px)';

/** Room for the Unscheduled panel beside the grid; below it the panel starts closed. */
const LARGE = '(min-width: 1024px)';

interface CreateModal {
  date: string;
}

const Calendar: React.FC = () => {
  const { t, i18n } = useTranslation();
  const openTask = useOpenTask();
  const isWide = useMediaQuery(WIDE);
  const isLarge = useMediaQuery(LARGE);
  const tasksQuery = useTasks({});
  const { data: apiTasks = [] } = tasksQuery;
  const tasks = apiTasks as ApiTask[];
  const tasksFailed = hasQueryFailed(tasksQuery);
  const tasksLoading = tasksQuery.isPending && !tasksFailed;
  const projects = useFlowStore((state) => state.projects);
  const { data: labels = [] } = useLabels();
  const createTask = useCreateTask();
  const [createModal, setCreateModal] = useState<CreateModal | null>(null);
  const [newTitle, setNewTitle] = useState('');
  const [newProjectId, setNewProjectId] = useState('');
  const [newLabelId, setNewLabelId] = useState('');
  // null until the user chooses, so the default can follow the width.
  const [panelChoice, setPanelChoice] = useState<boolean | null>(null);
  const panelOpen = panelChoice ?? isLarge;

  React.useEffect(() => {
    if (!newLabelId && labels.length > 0) setNewLabelId(String(labels[0].id));
  }, [labels, newLabelId]);

  const calendar = useCalendarView(i18n.language);
  const { view, cells, from, to } = calendar;

  const scheduling = useCalendarScheduling(tasks);
  const shownTasks = scheduling.tasks;

  // One lookup for the whole grid rather than a find() per pill per cell.
  const projectColorById = useMemo(
    () => new Map(projects.map((p) => [String(p.id), p.color])),
    [projects]
  );

  // Every project with a task here decides that task's drag; the ones with work
  // on screen come first for cycles, which are fetched per project and capped.
  const { taskProjectIds, cycleProjectIds } = useMemo(() => {
    const onScreen = new Set<string>();
    const all = new Set<string>();
    for (const task of shownTasks) {
      const id = String(task.projectId ?? '');
      if (!id) continue;
      all.add(id);
      const range = taskRange(task);
      if (range && range.start <= to && range.end >= from) onScreen.add(id);
    }
    const ordered = [...onScreen];
    for (const project of projects) {
      const id = String(project.id);
      if (!onScreen.has(id)) ordered.push(id);
    }
    return { taskProjectIds: [...all], cycleProjectIds: ordered };
  }, [shownTasks, projects, from, to]);

  const canEditIn = useCanEditTaskIn(taskProjectIds);
  const canEdit = (task: ApiTask): boolean => canEditIn(String(task.projectId));
  const planning = useCalendarPlanning({
    cycleProjectIds,
    milestoneProjectId: undefined,
    from,
    to,
  });

  const tasksByDate = useMemo(() => tasksByLocalDay(shownTasks), [shownTasks]);
  const agenda = useMemo(() => agendaDays(cells, tasksByDate), [cells, tasksByDate]);
  const unscheduledCount = useMemo(() => tasks.filter(needsScheduling).length, [tasks]);

  const today = todayIso();
  const taskTone = (task: ApiTask): string | undefined =>
    projectColorById.get(String(task.projectId));

  const openCreate = (cell: MonthCell) => {
    if (!cell.inMonth) return;
    setCreateModal({ date: cell.iso });
    setNewTitle('');
    setNewProjectId(projects[0]?.id ?? '');
    setNewLabelId(labels[0] ? String(labels[0].id) : '');
  };

  const submitCreate = () => {
    if (createTask.isPending) return;
    if (!newTitle.trim()) {
      toast.error(t('flowErrors.generic'));
      return;
    }
    const projectId = newProjectId || projects[0]?.id;
    // The guard used to also require `activeWorkspaceId`, which the removed
    // workspace layer leaves permanently empty, so Create could never succeed
    // and always blamed the project, which the modal had pre-filled. It also
    // required a label, so a business with no labels could not create a task
    // here at all; a label is optional on the create contract.
    if (!projectId) {
      toast.error(t('flowToasts.task.selectProjectFirst'));
      return;
    }
    createTask.mutate(
      {
        projectId,
        title: newTitle.trim(),
        dueDate: createModal!.date,
        ...(newLabelId ? { labels: [newLabelId] } : {}),
      },
      {
        onSuccess: () => {
          toast.success(t('flowToasts.task.created'));
          setCreateModal(null);
        },
        onError: () => toast.error(t('flowToasts.task.createFailed')),
      }
    );
  };

  const panel = (layout: 'side' | 'stacked') => (
    <UnscheduledPanel
      tasks={shownTasks}
      tone={taskTone}
      canEdit={canEdit}
      onOpen={(task) => openTask(String(task.id))}
      onClose={() => setPanelChoice(false)}
      scheduling={scheduling}
      layout={layout}
    />
  );

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={`${t('flow.calendar.heading')} · Flow`} />
      <div
        className="flex flex-wrap items-center justify-between gap-3 px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-flow-title font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.calendar.heading')}
          </h1>
          <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flow.calendar.subtitle')}
          </p>
        </div>
        <CalendarControls
          view={view}
          onView={calendar.setView}
          label={calendar.label}
          onStep={calendar.step}
          onToday={calendar.goToday}
          unscheduledCount={unscheduledCount}
          unscheduledOpen={panelOpen}
          onToggleUnscheduled={() => setPanelChoice(!panelOpen)}
        />
      </div>

      {/* At 390px a seven-column month gives each day about 55px, which truncates
          a title to two or three characters and leaves the pills too small to
          hit (rtylr#1215, S-10). The navigation above is shared; only the body
          swaps. */}
      {tasksFailed && (
        <div className="flex-1 overflow-auto p-4">
          <ListLoadFailed onRetry={() => void tasksQuery.refetch()} />
        </div>
      )}
      {tasksLoading && (
        <div className="flex-1 p-6 text-sm" style={{ color: 'var(--flow-text-dim)' }}>
          {t('flowCommon.loading')}
        </div>
      )}

      {!tasksFailed && !tasksLoading && !isWide && (
        <div className="flex-1 overflow-auto">
          {panelOpen && panel('stacked')}
          <CalendarAgenda
            entries={agenda}
            todayIso={today}
            taskId={(task) => String(task.id)}
            taskTitle={(task) => task.title}
            taskTone={taskTone}
            onOpen={(task) => openTask(String(task.id))}
            leading={(task) => <PriorityIcon priority={task.priority} size={11} />}
          />
        </div>
      )}

      {!tasksFailed && !tasksLoading && isWide && (
        <div className="flex flex-1 min-h-0">
          <div className="flex flex-1 min-w-0 flex-col p-4">
            <PlanningGrid
              cells={cells}
              view={view}
              tasks={shownTasks}
              cycles={planning.cycles}
              milestones={planning.milestones}
              leave={planning.leave}
              today={today}
              tone={taskTone}
              cycleTone={(cycle) => projectColorById.get(cycle.projectId)}
              leading={(task) => <PriorityIcon priority={task.priority} size={9} />}
              canEdit={canEdit}
              onOpen={(task) => openTask(String(task.id))}
              onCreate={openCreate}
              scheduling={scheduling}
            />
          </div>
          {panelOpen && panel('side')}
        </div>
      )}

      {/* Create task modal */}
      <AnimatePresence>
        {createModal && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-40"
              style={{ background: 'rgba(0,0,0,0.4)' }}
              onClick={() => setCreateModal(null)}
            />
            <motion.div
              initial={{ opacity: 0, scale: 0.96, y: 8 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.96 }}
              transition={{ duration: 0.15 }}
              className="fixed z-50 top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 rounded-xl shadow-2xl p-5 w-[360px]"
              style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
            >
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-flow-body font-semibold" style={{ color: 'var(--flow-text)' }}>
                  {t('flow.calendar.newTaskFor')}{' '}
                  <span style={{ color: ACCENT_VAR }}>
                    {new Date(createModal.date + 'T00:00:00').toLocaleDateString(i18n.language, {
                      month: 'short',
                      day: 'numeric',
                    })}
                  </span>
                </h2>
                <button
                  type="button"
                  onClick={() => setCreateModal(null)}
                  aria-label={t('flowCommon.close')}
                  className="h-6 w-6 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
                  style={{ color: 'var(--flow-text-dim)' }}
                >
                  <X size={12} />
                </button>
              </div>

              <input
                autoFocus
                value={newTitle}
                onChange={(e) => setNewTitle(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && submitCreate()}
                placeholder={t('flow.calendar.titlePlaceholder')}
                className="w-full rounded-lg px-3 py-2 text-flow-body mb-3"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border-control)',
                  color: 'var(--flow-text)',
                }}
              />

              {projects.length > 0 && (
                <select
                  value={newProjectId}
                  onChange={(e) => setNewProjectId(e.target.value)}
                  className="w-full rounded-lg px-3 py-2 text-flow-meta mb-4"
                  style={{
                    background: 'var(--flow-input-bg)',
                    border: '1px solid var(--flow-border-control)',
                    color: 'var(--flow-text)',
                  }}
                >
                  {projects.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))}
                </select>
              )}

              <select
                aria-label={t('flowCommon.settings.labels')}
                value={newLabelId}
                onChange={(e) => setNewLabelId(e.target.value)}
                disabled={labels.length === 0}
                className="w-full rounded-lg px-3 py-2 text-flow-meta mb-4 disabled:opacity-50"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border-control)',
                  color: 'var(--flow-text)',
                }}
              >
                {labels.length === 0 ? (
                  <option value="">{t('flowCommon.noData')}</option>
                ) : (
                  labels.map((label) => (
                    <option key={label.id} value={label.id}>
                      {label.name}
                    </option>
                  ))
                )}
              </select>

              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setCreateModal(null)}
                  className="px-3 py-1.5 rounded-sm text-flow-meta"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {t('flow.calendar.cancel')}
                </button>
                <button
                  type="button"
                  onClick={submitCreate}
                  disabled={!newTitle.trim() || createTask.isPending}
                  className="px-3 py-1.5 rounded-sm text-flow-meta font-semibold disabled:opacity-40"
                  style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
                >
                  {createTask.isPending ? t('flow.calendar.creating') : t('flow.calendar.create')}
                </button>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  );
};

export default Calendar;
