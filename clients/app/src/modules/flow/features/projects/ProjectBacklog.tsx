import React, { useMemo, useRef, useState } from 'react';
import { Link, useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import { ChevronDown, ChevronRight, GripVertical, Plus, Timer, Users } from 'lucide-react';
import { useFlowStore, type Task } from '../../shared/stores/flowStore';
import { useTasks } from '../tasks/useTasks';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import { useCanCreateTask, useCanEditTask } from '../../shared/lib/currentUser';
import { formatShortDate, getInitials } from '../../shared/lib/utils';
import { withAlpha } from '../../shared/lib/colorUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import TaskRow from '../../shared/components/TaskRow';
import WorkloadBar from '../../shared/components/WorkloadBar';
import CycleBadge from '../../shared/components/CycleBadge';
import { addDays, todayIso } from '../../shared/lib/dateUtils';
import { endDateForLength } from '../cycles/cycleModel';
import FilterBar from '../../shared/filters/FilterBar';
import type { ApiCycle, ApiStatus } from '../../shared/services/flowApi';
import {
  useCompleteCycle,
  useCreateCycle,
  useCycleCapacity,
  useCycles,
  useMoveCycleTask,
  useStartCycle,
} from '../cycles/useCycles';
import ProjectViewTabs from './ProjectViewTabs';
import { useCurrentProject } from './useCurrentProject';
import { useProjectStatuses } from './useStatuses';
import { useProjectTaskFilters } from './useProjectTaskFilters';
import { columnIdAtPoint, exceedsDragThreshold, type ColumnRect } from './boardDrag';
import { insertionIndex, isSamePlace, type CardSlot, type DropSlot } from './boardOrder';
import {
  BACKLOG_SECTION_ID,
  buildBacklogSections,
  capacityTone,
  hoursLabel,
  planMove,
  taskBucket,
  type BacklogSection,
} from './backlog/backlogModel';
import { CompleteCycleDialog, StartCycleDialog } from './backlog/CycleDialogs';

/**
 * Backlog: plan, start and close a project's cycles from one screen (T-570).
 *
 * The active cycle comes first, then upcoming cycles, then unplanned work.
 * Tasks are dragged between and within sections with the board's pointer drag
 * helpers; each drop is one moveCycleTask call, applied optimistically and
 * synced to other viewers through the project's realtime channel.
 */

// Section keys in the DOM. The unplanned section has no cycle id.
const sectionKey = (id: string) => id || 'backlog';
const sectionIdFromKey = (key: string) => (key === 'backlog' ? BACKLOG_SECTION_ID : key);

interface DragSession {
  pointerId: number;
  taskId: string;
  source: DropSlot;
  startX: number;
  startY: number;
  active: boolean;
  slot: DropSlot | null;
  cleanup: () => void;
}

function measureSections(root: HTMLElement | null): ColumnRect[] {
  if (!root) return [];
  return [...root.querySelectorAll<HTMLElement>('[data-backlog-section]')].map((el) => {
    const r = el.getBoundingClientRect();
    return {
      id: el.dataset.backlogSection ?? '',
      left: r.left,
      right: r.right,
      top: r.top,
      bottom: r.bottom,
    };
  });
}

function measureRows(root: HTMLElement | null, key: string): CardSlot[] {
  const section = root?.querySelector<HTMLElement>(`[data-backlog-section="${key}"]`);
  if (!section) return [];
  return [...section.querySelectorAll<HTMLElement>('[data-backlog-row]')].map((el) => {
    const r = el.getBoundingClientRect();
    return { id: el.dataset.backlogRow ?? '', mid: r.top + r.height / 2 };
  });
}

const ProjectBacklog: React.FC = () => {
  const { t } = useTranslation();
  const { projectId = '' } = useParams({ strict: false });
  useProjectRealtime(projectId);

  const {
    project,
    isLoading: projectLoading,
    loadFailed: projectFailed,
    refetch: refetchProject,
  } = useCurrentProject(projectId);
  const tasksQuery = useTasks({ projectId, activeOnly: true, compact: true });
  const cyclesQuery = useCycles(projectId);
  const { data: rawStatuses } = useProjectStatuses(projectId);
  const canEdit = useCanEditTask(projectId);
  const canCreate = useCanCreateTask(projectId);

  const apiTasks = useMemo(() => tasksQuery.data ?? [], [tasksQuery.data]);
  const mappedTasks = useMemo(() => apiTasks.map(apiTaskToTask), [apiTasks]);
  const filters = useProjectTaskFilters(projectId, mappedTasks);
  const { matchingIds } = filters;
  const visibleTasks = useMemo(
    () => (matchingIds === null ? mappedTasks : mappedTasks.filter((x) => matchingIds.has(x.id))),
    [mappedTasks, matchingIds]
  );
  const statusesById = useMemo(
    () => new Map<string, ApiStatus>((rawStatuses ?? []).map((s) => [s.id, s])),
    [rawStatuses]
  );
  const cycles = useMemo(() => cyclesQuery.data ?? [], [cyclesQuery.data]);
  const sections = useMemo(
    () => buildBacklogSections(cycles, visibleTasks, statusesById),
    [cycles, visibleTasks, statusesById]
  );
  const activeCycle = cycles.find((c) => c.status === 'active');
  const upcoming = sections.filter((s) => s.cycle?.status === 'upcoming');

  const moveTask = useMoveCycleTask();
  const startCycle = useStartCycle();
  const completeCycle = useCompleteCycle();
  const createCycle = useCreateCycle();
  const [starting, setStarting] = useState<ApiCycle | null>(null);
  const [completing, setCompleting] = useState<ApiCycle | null>(null);

  // ── Drag ────────────────────────────────────────────────────────────────
  const rootRef = useRef<HTMLDivElement>(null);
  const dragRef = useRef<DragSession | null>(null);
  const justDraggedRef = useRef(false);
  const [dragTaskId, setDragTaskId] = useState<string | null>(null);
  const [dropSlot, setDropSlot] = useState<DropSlot | null>(null);

  const commit = (taskId: string, targetId: string, index: number) => {
    const section = sections.find((s) => s.id === targetId);
    if (!section) return;
    const plan = planMove(section, taskId, index);
    moveTask.mutate(plan);
  };

  const slotAt = (taskId: string, x: number, y: number): DropSlot | null => {
    const key = columnIdAtPoint(measureSections(rootRef.current), x, y);
    if (key === null) return null;
    return { columnId: key, index: insertionIndex(measureRows(rootRef.current, key), taskId, y) };
  };

  const endDrag = () => {
    dragRef.current?.cleanup();
    dragRef.current = null;
    setDragTaskId(null);
    setDropSlot(null);
  };

  const startDrag = (e: React.PointerEvent, task: Task, section: BacklogSection) => {
    if (!canEdit || dragRef.current || e.button !== 0) return;
    const key = sectionKey(section.id);
    const sourceIndex = section.tasks.findIndex((x) => x.id === task.id);
    const session: DragSession = {
      pointerId: e.pointerId,
      taskId: task.id,
      source: { columnId: key, index: Math.max(0, sourceIndex) },
      startX: e.clientX,
      startY: e.clientY,
      active: false,
      slot: null,
      cleanup: () => undefined,
    };
    const previousSelect = document.body.style.userSelect;
    const onMove = (ev: PointerEvent) => {
      if (ev.pointerId !== session.pointerId) return;
      if (!session.active) {
        if (!exceedsDragThreshold(ev.clientX - session.startX, ev.clientY - session.startY)) {
          return;
        }
        session.active = true;
        document.body.style.userSelect = 'none';
        setDragTaskId(session.taskId);
      }
      if (ev.cancelable) ev.preventDefault();
      session.slot = slotAt(session.taskId, ev.clientX, ev.clientY);
      setDropSlot(session.slot);
    };
    const onUp = (ev: PointerEvent) => {
      if (ev.pointerId !== session.pointerId) return;
      const { active, slot } = session;
      endDrag();
      if (!active) return;
      justDraggedRef.current = true;
      window.setTimeout(() => {
        justDraggedRef.current = false;
      }, 0);
      if (slot === null || isSamePlace(session.source, slot)) return;
      commit(session.taskId, sectionIdFromKey(slot.columnId), slot.index);
    };
    const onCancel = (ev: PointerEvent) => {
      if (ev.pointerId === session.pointerId) endDrag();
    };
    const onKey = (ev: KeyboardEvent) => {
      if (ev.key === 'Escape') endDrag();
    };
    window.addEventListener('pointermove', onMove, { passive: false });
    window.addEventListener('pointerup', onUp, true);
    window.addEventListener('pointercancel', onCancel, true);
    window.addEventListener('keydown', onKey);
    session.cleanup = () => {
      window.removeEventListener('pointermove', onMove);
      window.removeEventListener('pointerup', onUp, true);
      window.removeEventListener('pointercancel', onCancel, true);
      window.removeEventListener('keydown', onKey);
      document.body.style.userSelect = previousSelect;
    };
    dragRef.current = session;
  };

  // Keyboard reordering: Alt + Up / Down moves the focused task one place.
  const onRowKeyDown = (e: React.KeyboardEvent, task: Task, section: BacklogSection) => {
    if (!canEdit || !e.altKey || (e.key !== 'ArrowUp' && e.key !== 'ArrowDown')) return;
    e.preventDefault();
    const index = section.tasks.findIndex((x) => x.id === task.id);
    const target = e.key === 'ArrowUp' ? index - 1 : index + 1;
    if (index < 0 || target < 0 || target >= section.tasks.length) return;
    commit(task.id, section.id, target);
  };

  // ── Loading and failure ────────────────────────────────────────────────
  const tasksFailed = hasQueryFailed(tasksQuery);
  const cyclesFailed = hasQueryFailed(cyclesQuery);
  if (projectLoading && !project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.project.loading')}
      </div>
    );
  }
  if (projectFailed || tasksFailed || cyclesFailed) {
    return (
      <ListLoadFailed
        onRetry={() => {
          void refetchProject();
          void tasksQuery.refetch();
          void cyclesQuery.refetch();
        }}
      />
    );
  }
  if (!project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.project.notFound')}
      </div>
    );
  }
  const loading = (tasksQuery.isPending || cyclesQuery.isPending) && !tasksFailed;

  const createUpcomingCycle = () => {
    // Starts the day after the last open cycle ends, or today, for two weeks.
    const lastEnd = cycles
      .filter((c) => c.status !== 'completed' && c.endDate)
      .map((c) => c.endDate)
      .sort()
      .pop();
    const startDate = lastEnd ? addDays(lastEnd, 1) : todayIso();
    createCycle.mutate(
      {
        projectId,
        name: t('flow.backlog.newCycleName', { number: cycles.length + 1 }),
        startDate,
        endDate: endDateForLength(startDate, 2),
      },
      { onSuccess: () => toast.success(t('flowToasts.cycle.created')) }
    );
  };

  const completingSection = sections.find((s) => s.id === completing?.id);
  const unfinished = (completingSection?.tasks ?? []).filter((task) => {
    const bucket = taskBucket(task, statusesById);
    return bucket !== 'done' && bucket !== 'cancelled';
  });
  const nextForCompleting = upcoming.find((s) => s.id !== completing?.id)?.cycle;

  return (
    <div className="flex h-full flex-col">
      <HelmetContainer title={t('flow.backlog.pageTitle')} />
      <ProjectViewTabs
        project={project}
        projectId={projectId}
        taskCount={apiTasks.length}
        onNewTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
        rightSlot={
          canEdit ? (
            <button
              type="button"
              onClick={createUpcomingCycle}
              disabled={createCycle.isPending}
              className="flow-focus flex h-7 items-center gap-1.5 rounded-sm px-2.5 text-flow-meta font-medium hover:bg-[var(--flow-hover)] disabled:opacity-50"
              style={{ border: '1px solid var(--flow-border)', color: 'var(--flow-text)' }}
            >
              <Timer size={12} /> {t('flow.backlog.createCycle')}
            </button>
          ) : undefined
        }
      />
      <FilterBar
        state={filters.state}
        currentUserId={filters.currentUserId}
        fields={filters.fields}
        people={filters.people}
        visible={matchingIds?.size ?? mappedTasks.length}
        total={mappedTasks.length}
        allowDim={false}
      />

      {loading ? (
        <div className="p-6" role="status" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.project.loading')}
        </div>
      ) : (
        <div ref={rootRef} className="flex-1 space-y-4 overflow-y-auto px-3 py-4 sm:px-5">
          {sections.map((section) => (
            <BacklogSectionView
              key={sectionKey(section.id)}
              section={section}
              canEdit={canEdit}
              canCreate={canCreate}
              hasActiveCycle={activeCycle !== undefined}
              allTasks={mappedTasks}
              dragTaskId={dragTaskId}
              dropIndex={
                dropSlot && dropSlot.columnId === sectionKey(section.id) ? dropSlot.index : null
              }
              onRowPointerDown={(e, task) => startDrag(e, task, section)}
              onRowKeyDown={(e, task) => onRowKeyDown(e, task, section)}
              suppressClick={() => justDraggedRef.current}
              onStart={() => setStarting(section.cycle ?? null)}
              onComplete={() => setCompleting(section.cycle ?? null)}
              onCreate={() =>
                useFlowStore.getState().setQuickCreateOpen(true, undefined, section.id || undefined)
              }
            />
          ))}
        </div>
      )}

      <StartCycleDialog
        cycle={starting}
        activeCycle={activeCycle}
        pending={startCycle.isPending}
        onOpenChange={(open) => !open && setStarting(null)}
        onStart={({ lengthWeeks, goal }) => {
          if (!starting) return;
          startCycle.mutate(
            { id: starting.id, lengthWeeks, goal },
            { onSuccess: () => setStarting(null) }
          );
        }}
      />
      <CompleteCycleDialog
        cycle={completing}
        unfinished={unfinished}
        nextCycle={nextForCompleting}
        pending={completeCycle.isPending}
        onOpenChange={(open) => !open && setCompleting(null)}
        onComplete={(decisions) => {
          if (!completing) return;
          completeCycle.mutate(
            { id: completing.id, decisions },
            { onSuccess: () => setCompleting(null) }
          );
        }}
      />
    </div>
  );
};

interface SectionViewProps {
  section: BacklogSection;
  canEdit: boolean;
  canCreate: boolean;
  hasActiveCycle: boolean;
  allTasks: Task[];
  dragTaskId: string | null;
  dropIndex: number | null;
  onRowPointerDown: (e: React.PointerEvent, task: Task) => void;
  onRowKeyDown: (e: React.KeyboardEvent, task: Task) => void;
  suppressClick: () => boolean;
  onStart: () => void;
  onComplete: () => void;
  onCreate: () => void;
}

const BacklogSectionView: React.FC<SectionViewProps> = ({
  section,
  canEdit,
  canCreate,
  hasActiveCycle,
  allTasks,
  dragTaskId,
  dropIndex,
  onRowPointerDown,
  onRowKeyDown,
  suppressClick,
  onStart,
  onComplete,
  onCreate,
}) => {
  const { t, i18n } = useTranslation();
  const { cycle, totals, tasks } = section;
  const [collapsed, setCollapsed] = useState(false);
  const [showCapacity, setShowCapacity] = useState(false);
  const capacityQuery = useCycleCapacity(cycle?.id);
  const capacity = capacityQuery.data;
  const available = (capacity?.members ?? []).reduce(
    (sum, m) => sum + (m.availableMinutes ?? 0),
    0
  );
  const committed = totals.estimateMinutes;
  const tone = cycle ? capacityTone(committed, available) : 'ok';
  const toneColor =
    tone === 'over'
      ? 'var(--flow-danger-fg)'
      : tone === 'warn'
        ? 'var(--flow-warn-fg)'
        : 'var(--flow-text-muted)';
  const titleId = `backlog-section-${sectionKey(section.id)}`;
  const titleById = new Map(allTasks.map((task) => [task.id, task.title]));

  const placeholder = (
    <div
      aria-hidden
      className="mx-2 h-0.5 rounded-full"
      style={{ background: 'var(--accent)' }}
      data-testid="backlog-drop-indicator"
    />
  );

  return (
    <section
      aria-labelledby={titleId}
      data-backlog-section={sectionKey(section.id)}
      className="rounded-lg"
      style={{
        background: 'var(--flow-card)',
        border: `1px solid ${dropIndex !== null ? 'var(--accent)' : 'var(--flow-border)'}`,
      }}
    >
      <header className="flex flex-wrap items-start gap-x-3 gap-y-1.5 px-3 py-2.5">
        <button
          type="button"
          onClick={() => setCollapsed((c) => !c)}
          aria-expanded={!collapsed}
          aria-label={collapsed ? t('flow.backlog.expand') : t('flow.backlog.collapse')}
          className="flow-focus mt-0.5 rounded-sm hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-dim)' }}
        >
          {collapsed ? (
            <ChevronRight size={14} className="rtl:rotate-180" />
          ) : (
            <ChevronDown size={14} />
          )}
        </button>
        <div className="min-w-0 flex-1 space-y-0.5">
          <div className="flex flex-wrap items-center gap-2">
            <h2
              id={titleId}
              className="text-flow-body font-semibold"
              style={{ color: 'var(--flow-text)' }}
            >
              {cycle ? (
                <Link
                  to="/flow/cycles/$cycleId"
                  params={{ cycleId: cycle.id }}
                  className="flow-focus hover:underline"
                >
                  {cycle.name}
                </Link>
              ) : (
                t('flow.backlog.unplanned')
              )}
            </h2>
            {cycle && <CycleBadge cycle={cycle} compact showState />}
            <span className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
              {t('flow.backlog.taskCount', { count: tasks.length })}
            </span>
          </div>
          {cycle && (cycle.startDate || cycle.endDate) && (
            <p className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {formatShortDate(cycle.startDate, i18n.language)} –{' '}
              {formatShortDate(cycle.endDate, i18n.language)}
            </p>
          )}
          {cycle?.goal && (
            <p dir="auto" className="text-flow-meta" style={{ color: 'var(--flow-text)' }}>
              <span className="font-medium">{t('flow.backlog.goal')}: </span>
              {cycle.goal}
            </p>
          )}
          {(cycle?.carriedOver ?? []).map((carry) => (
            <p
              key={carry.fromCycleId}
              className="text-flow-meta"
              style={{ color: 'var(--flow-text-muted)' }}
              title={(carry.taskIds ?? []).map((id) => titleById.get(id) ?? id).join(', ')}
            >
              {t('flow.backlog.carriedOver', {
                count: carry.taskIds?.length ?? 0,
                name: carry.fromCycleName ?? '',
              })}
            </p>
          ))}
          <p
            className="flex flex-wrap gap-x-3 text-flow-meta"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <span>{t('flow.backlog.totals.todo', { count: totals.todo })}</span>
            <span>{t('flow.backlog.totals.active', { count: totals.active })}</span>
            <span>{t('flow.backlog.totals.done', { count: totals.done })}</span>
            {cycle ? (
              <span style={{ color: toneColor }} data-capacity-tone={tone}>
                {t('flow.backlog.committedOfAvailable', {
                  committed: hoursLabel(committed),
                  available: hoursLabel(available),
                })}
              </span>
            ) : (
              <span>{t('flow.backlog.estimate', { hours: hoursLabel(committed) })}</span>
            )}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-1.5">
          {cycle && (
            <button
              type="button"
              onClick={() => setShowCapacity((v) => !v)}
              aria-expanded={showCapacity}
              className="flow-focus flex h-7 items-center gap-1 rounded-sm px-2 text-flow-meta hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <Users size={12} /> {t('flow.backlog.capacity')}
            </button>
          )}
          {cycle?.status === 'upcoming' && canEdit && (
            <button
              type="button"
              onClick={onStart}
              disabled={hasActiveCycle}
              title={hasActiveCycle ? t('flow.backlog.startBlocked') : undefined}
              className="flow-focus h-7 rounded-sm px-2.5 text-flow-meta font-semibold disabled:opacity-50"
              style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
            >
              {t('flow.backlog.startCycle')}
            </button>
          )}
          {cycle?.status === 'active' && canEdit && (
            <button
              type="button"
              onClick={onComplete}
              className="flow-focus h-7 rounded-sm px-2.5 text-flow-meta font-semibold"
              style={{ background: 'var(--flow-action-bg)', color: 'var(--flow-on-accent)' }}
            >
              {t('flow.backlog.completeCycle')}
            </button>
          )}
        </div>
      </header>

      {cycle?.status === 'upcoming' && hasActiveCycle && canEdit && (
        <p className="px-3 pb-2 text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
          {t('flow.backlog.startBlocked')}
        </p>
      )}

      {showCapacity && cycle && (
        <div className="px-4 pb-2" style={{ borderTop: '1px solid var(--flow-border)' }}>
          {capacityQuery.isPending ? (
            <p className="py-2 text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
              {t('flow.project.loading')}
            </p>
          ) : (
            <>
              {capacity?.leaveApplied === false && (
                <p className="pt-2 text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
                  {t('flow.backlog.leaveUnavailable')}
                </p>
              )}
              {(capacity?.members ?? []).length === 0 && (
                <p className="py-2 text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
                  {t('flow.backlog.noMembers')}
                </p>
              )}
              {(capacity?.members ?? []).map((m) => {
                const memberTone = capacityTone(m.committedMinutes ?? 0, m.availableMinutes ?? 0);
                const avail = m.availableMinutes ?? 0;
                return (
                  <div key={m.userId}>
                    <WorkloadBar
                      name={m.name || t('flow.backlog.unknownMember')}
                      initials={getInitials(m.name ?? '') || '?'}
                      load={avail > 0 ? (m.committedMinutes ?? 0) / avail : 0}
                      capacity={Math.round((avail / 60) * 10) / 10}
                      assigned={Math.round(((m.committedMinutes ?? 0) / 60) * 10) / 10}
                      tone={memberTone}
                    />
                    {(m.leaveDays ?? 0) > 0 && (
                      <p
                        className="-mt-1 pb-1 ps-10 text-flow-meta"
                        style={{ color: 'var(--flow-text-dim)' }}
                      >
                        {t('flow.backlog.leaveDays', { count: m.leaveDays ?? 0 })}
                      </p>
                    )}
                  </div>
                );
              })}
              {(capacity?.unassignedCount ?? 0) > 0 && (
                <p className="pb-1 text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
                  {t('flow.backlog.unassigned', {
                    count: capacity?.unassignedCount ?? 0,
                    hours: hoursLabel(capacity?.unassignedMinutes ?? 0),
                  })}
                </p>
              )}
            </>
          )}
        </div>
      )}

      {!collapsed && (
        <div className="pb-1.5" style={{ borderTop: '1px solid var(--flow-border)' }}>
          {tasks.length === 0 && (
            <p
              className="mx-2 my-1.5 rounded-md px-3 py-3 text-center text-flow-meta"
              style={{
                color: 'var(--flow-text-dim)',
                border: `1px dashed ${withAlpha('var(--flow-text-dim)', 40)}`,
              }}
            >
              {cycle ? t('flow.backlog.emptyCycle') : t('flow.backlog.emptyBacklog')}
            </p>
          )}
          {tasks.map((task, i) => {
            const othersBefore = tasks.slice(0, i).filter((x) => x.id !== dragTaskId).length;
            return (
              <React.Fragment key={task.id}>
                {dropIndex !== null &&
                  task.id !== dragTaskId &&
                  dropIndex === othersBefore &&
                  placeholder}
                <div
                  data-backlog-row={task.id}
                  className="group flex items-center"
                  style={{ opacity: dragTaskId === task.id ? 0.4 : 1, touchAction: 'pan-y' }}
                  onPointerDown={(e) => {
                    // Mouse and pen drag from anywhere on the row; touch only
                    // from the grip, so a swipe still scrolls the page.
                    if (e.pointerType !== 'touch') onRowPointerDown(e, task);
                  }}
                  onClickCapture={(e) => {
                    if (suppressClick()) {
                      e.stopPropagation();
                      e.preventDefault();
                    }
                  }}
                  onKeyDown={(e) => onRowKeyDown(e, task)}
                >
                  {canEdit && (
                    <span
                      aria-hidden
                      className="flex h-8 w-5 shrink-0 cursor-grab items-center justify-center opacity-40 group-hover:opacity-100"
                      style={{ color: 'var(--flow-text-dim)', touchAction: 'none' }}
                      onPointerDown={(e) => {
                        if (e.pointerType === 'touch') {
                          e.stopPropagation();
                          onRowPointerDown(e, task);
                        }
                      }}
                    >
                      <GripVertical size={12} />
                    </span>
                  )}
                  <div className="min-w-0 flex-1">
                    <TaskRow task={task} />
                  </div>
                </div>
              </React.Fragment>
            );
          })}
          {dropIndex !== null &&
            dropIndex >= tasks.filter((x) => x.id !== dragTaskId).length &&
            placeholder}
          {canCreate && (
            <button
              type="button"
              onClick={onCreate}
              className="flow-focus mx-2 mt-1 flex items-center gap-1.5 rounded-sm px-2 py-1 text-flow-meta hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <Plus size={12} /> {t('flow.backlog.create')}
            </button>
          )}
        </div>
      )}
    </section>
  );
};

export default ProjectBacklog;
