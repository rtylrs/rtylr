import { hasQueryFailed } from '../../shared/lib/queryFailureUtils';
import React, { useCallback, useEffect, useId, useMemo, useRef, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useClickOutside } from '@voxire/utils';
import { createPortal } from 'react-dom';
import { motion } from 'motion/react';
import { toast } from 'sonner';
import { useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import {
  ArrowLeft,
  ArrowRight,
  ChevronDown,
  GripVertical,
  ListChecks,
  MoreHorizontal,
  Pencil,
  Plus,
  Settings2,
  Trash2,
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@voxire/ui';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useTasks } from '../tasks/useTasks';
import { useMoveTaskToColumn } from './useColumns';
import {
  useCreateStatus,
  useProjectStatuses,
  useReorderStatuses,
  useUpdateStatus,
} from './useStatuses';
import {
  displayStatusName,
  STATUS_COLORS,
  closedRecently,
  insertionPlan,
  isClosedBucket,
  isOnlyDoneStatus,
  movedStatusIds,
  sortStatuses,
  statusIdOf,
} from './statusUtils';
import {
  DUE_LANES,
  PRIORITY_LABEL_KEY,
  PRIORITY_ORDER,
  fieldGroupOptions,
  fieldLanes,
  groupableFields,
  laneKeyFor,
} from './taskGroups';
import { useProjectFields } from './useFields';
import { fieldIdFromKey } from '../../shared/lib/customFields';
import type { ApiField } from '../../shared/services/flowApi';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import type { BoardGroupKey } from '../../shared/lib/boardUtils';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import TaskCard from '../../shared/components/TaskCard';
import ProjectViewTabs from './ProjectViewTabs';
import MoveToColumnMenu, { type MoveTarget } from './components/MoveToColumnMenu';
import {
  columnIdAtPoint,
  exceedsDragThreshold,
  resolveDropTarget,
  type ColumnRect,
} from './boardDrag';
import {
  insertionIndex,
  isSamePlace,
  orderBoardTasks,
  orderedIdsAfterDrop,
  type CardSlot,
  type DropSlot,
} from './boardOrder';
import { useCurrentProject } from './useCurrentProject';
import {
  useCanEditTask,
  useCanCreateTask,
  useCanDeleteResource,
} from '../../shared/lib/currentUser';
import BoardSettingsPanel from './components/BoardSettingsPanel';
import StatusSettingsPanel from './components/StatusSettingsPanel';
import DeleteStatusDialog from './components/DeleteStatusDialog';
import NewStatusDialog from './components/NewStatusDialog';
import type { ApiStatus } from '../../shared/services/flowApi';
import type { Task } from '../../shared/stores/flowStore';
import { withAlpha, onTint } from '../../shared/lib/colorUtils';
import { ListLoadFailed } from '../../shared/components/ListLoadFailed';
import FilterBar from '../../shared/filters/FilterBar';
import { useProjectTaskFilters } from './useProjectTaskFilters';
import { projectFiltersForView, useSaveFiltersAsView } from '../views/useSaveFiltersAsView';

/** What the board renders while a card is being dragged. */
interface DragView {
  taskId: string;
  /** Where the card started: its column and its place among that column's cards. */
  source: DropSlot;
  /** Where it would land now, or null while the pointer is outside every column. */
  slot: DropSlot | null;
  /** The card's height, so the placeholder holds exactly the space it will fill. */
  height: number;
}

/** The floating copy of the card that follows the pointer. */
interface GhostInfo {
  task: Task;
  width: number;
}

/** Everything one drag gesture needs, captured when the pointer goes down. */
interface DragSession {
  pointerId: number;
  taskId: string;
  task: Task;
  source: DropSlot;
  /** Cards can only be reordered when the board is not split into swimlanes. */
  reorder: boolean;
  startX: number;
  startY: number;
  x: number;
  y: number;
  /** Where inside the card the pointer grabbed it, so the copy does not jump. */
  offsetX: number;
  offsetY: number;
  width: number;
  height: number;
  /** Each column's card ids in display order, as they were when the drag began. */
  columnIds: Map<string, string[]>;
  /** Each column's card centres, measured against the list's own top edge. */
  slots: Map<string, CardSlot[]>;
  slot: DropSlot | null;
  active: boolean;
  cleanup: () => void;
}

/**
 * The card being dragged stays mounted but leaves the layout. A touch pointer
 * keeps targeting the node it went down on, so that node must not be unmounted
 * or re-parented while the drag is in progress.
 */
const HIDDEN_CARD_STYLE: React.CSSProperties = {
  position: 'absolute',
  visibility: 'hidden',
  pointerEvents: 'none',
  width: '100%',
};

const sameSlot = (a: DropSlot | null, b: DropSlot | null): boolean =>
  a === b || (a !== null && b !== null && a.columnId === b.columnId && a.index === b.index);

/** Card centres of one column list, relative to the top of its scrollable content. */
const measureSlots = (list: HTMLElement | null | undefined): CardSlot[] => {
  if (!list) return [];
  const origin = list.getBoundingClientRect().top - list.scrollTop;
  return [...list.querySelectorAll<HTMLElement>('[data-board-card]')].map((el) => {
    const r = el.getBoundingClientRect();
    return { id: el.dataset.boardCard ?? '', mid: r.top - origin + r.height / 2 };
  });
};

// ── Group-by (swimlanes) ────────────────────────────────────────────────
// "None" = plain kanban columns. Any other value partitions each column's
// tasks into horizontal lanes keyed by that attribute (assignee, priority,
// etc.). The choice persists per-project in localStorage so switching
// projects doesn't reset it.

/** Re-exported name kept local for readability; the type lives in boardUtils. */
type GroupKey = BoardGroupKey;

// labelKey, not label: module scope, so it cannot call `t` itself. A custom
// field option carries its own name as `label` instead.
type GroupOption = { key: GroupKey; labelKey?: string; label?: string };
const GROUP_OPTIONS: GroupOption[] = [
  { key: 'none', labelKey: 'flow.board.groupBy.none' },
  { key: 'assignee', labelKey: 'flow.board.groupBy.assignee' },
  { key: 'priority', labelKey: 'flow.board.groupBy.priority' },
  { key: 'due', labelKey: 'flow.board.groupBy.due' },
];

interface Lane {
  key: string;
  /** A person's name, for assignee lanes. Never translated. */
  label?: string;
  /** A translation key, for priority and due-window lanes. */
  labelKey?: string;
  tasks: Task[];
}

/**
 * Partition a column's tasks into lanes. Every column renders the same
 * set of lane keys in the same order so rows line up horizontally across
 * columns - this is the "swimlane" effect.
 */
function buildLanes(
  allLaneKeys: { key: string; label?: string; labelKey?: string }[],
  columnTasks: Task[],
  groupBy: GroupKey
): Lane[] {
  if (groupBy === 'none') {
    return [{ key: '_all', label: '', tasks: columnTasks }];
  }
  const byKey = new Map<string, Task[]>();
  for (const t of columnTasks) {
    const k = laneKeyFor(t, groupBy);
    const bucket = byKey.get(k) ?? [];
    bucket.push(t);
    byKey.set(k, bucket);
  }
  return allLaneKeys.map(({ key, label, labelKey }) => ({
    key,
    label,
    labelKey,
    tasks: byKey.get(key) ?? [],
  }));
}

/**
 * Derive the deterministic ordered lane list across ALL tasks so every
 * column shows the same swimlanes in the same order. Without this, empty
 * columns would skip lanes and rows would misalign.
 */
function buildLaneIndex(
  tasks: Task[],
  groupBy: GroupKey,
  fields: readonly ApiField[]
): { key: string; label?: string; labelKey?: string }[] {
  if (groupBy === 'none') return [{ key: '_all', label: '' }];

  const fieldId = fieldIdFromKey(groupBy);
  if (fieldId !== null) {
    const field = fields.find((f) => f.id === fieldId);
    return field ? fieldLanes(field) : [{ key: '_all', label: '' }];
  }

  if (groupBy === 'priority') {
    return PRIORITY_ORDER.map((p) => ({ key: p, labelKey: PRIORITY_LABEL_KEY[p] }));
  }

  if (groupBy === 'due') return DUE_LANES;

  // assignee - build from data, keep stable order by first-seen. These labels
  // are people's names, so they carry `label` and are never translated.
  const seen = new Map<string, string>();
  for (const t of tasks) {
    const k = laneKeyFor(t, groupBy);
    if (!seen.has(k)) seen.set(k, k);
  }
  return Array.from(seen.values()).map((k) => ({ key: k, label: k }));
}

function loadGroupBy(projectId: string): GroupKey {
  try {
    const raw = localStorage.getItem(`flow:board:groupBy:${projectId}`);
    if (raw === 'none' || raw === 'assignee' || raw === 'priority' || raw === 'due') {
      return raw;
    }
    if (raw && /^field:\d+$/.test(raw)) return raw as GroupKey;
  } catch {
    // localStorage unavailable (SSR / private mode) - fall through.
  }
  return 'none';
}

function saveGroupBy(projectId: string, key: GroupKey): void {
  try {
    localStorage.setItem(`flow:board:groupBy:${projectId}`, key);
  } catch {
    // ignore
  }
}

const ProjectBoard: React.FC = () => {
  const { t } = useTranslation();
  const { projectId = '' } = useParams({ strict: false });

  const selectedTaskIds = useFlowStore((state) => state.selectedTaskIds);

  const toggleSelectTask = useFlowStore((state) => state.toggleSelectTask);
  const {
    project,
    isLoading: projectLoading,
    loadFailed: projectMetadataFailed,
    refetch: refetchProjectMetadata,
  } = useCurrentProject(projectId);
  const moveTask = useMoveTaskToColumn();
  const canEdit = useCanEditTask(projectId);
  const canCreate = useCanCreateTask(projectId);
  // "Manage project" for board setup. A viewer still opens the panel and sees
  // why it is read-only, rather than the board's behaviour being unexplainable.
  const canManageBoard = useCanDeleteResource(projectId);

  useProjectRealtime(projectId);

  const tasksQuery = useTasks({ projectId, activeOnly: true, compact: true });
  const { data: apiTasks = [] } = tasksQuery;
  const projectLoadFailed = projectMetadataFailed || hasQueryFailed(tasksQuery);
  const refetchProjects = () => {
    void refetchProjectMetadata();
    void tasksQuery.refetch();
  };
  // The project's statuses are its columns; this is the board's one source.
  const columnsQuery = useProjectStatuses(projectId);
  const columns = useMemo(() => sortStatuses(columnsQuery.data ?? []), [columnsQuery.data]);
  const createStatus = useCreateStatus(projectId);
  const updateStatus = useUpdateStatus(projectId);
  const reorderStatuses = useReorderStatuses(projectId);

  // The live drag preview. `source` is where the card started; `slot` is where
  // the pointer says it would land now, or null while it is outside every
  // column (the placeholder then sits back at the source).
  const [dragView, setDragView] = useState<DragView | null>(null);
  const [ghost, setGhost] = useState<GhostInfo | null>(null);
  const draggingTaskId = dragView?.taskId ?? null;
  const dragOver = dragView?.slot?.columnId ?? null; // column id
  const [addingColumn, setAddingColumn] = useState(false);
  const [newColumnName, setNewColumnName] = useState('');
  const [storedGroupBy, setGroupByState] = useState<GroupKey>('none');
  const [pendingDeleteColumn, setPendingDeleteColumn] = useState<ApiStatus | null>(null);
  // Where a status named in the dialog is inserted; null while it is closed.
  const [insertAt, setInsertAt] = useState<number | null>(null);
  // Done and cancelled columns whose older cards the user asked to see.
  const [showOlder, setShowOlder] = useState<ReadonlySet<string>>(new Set());
  const [boardSettingsOpen, setBoardSettingsOpen] = useState(false);
  const [statusSettingsOpen, setStatusSettingsOpen] = useState(false);

  // Hydrate group-by from localStorage once the project id is known.
  useEffect(() => {
    if (projectId) setGroupByState(loadGroupBy(projectId));
  }, [projectId]);

  // Grouping by a custom field needs that field to still be a select of this
  // project; a deleted one falls back to no grouping rather than hiding cards.
  const fieldsQuery = useProjectFields(projectId ?? '');
  const groupFields = useMemo(() => groupableFields(fieldsQuery.data ?? []), [fieldsQuery.data]);
  const groupOptions = useMemo(
    () => [...GROUP_OPTIONS, ...fieldGroupOptions(groupFields)],
    [groupFields]
  );
  const groupFieldId = fieldIdFromKey(storedGroupBy);
  const groupBy: GroupKey =
    groupFieldId !== null && !groupFields.some((f) => f.id === groupFieldId)
      ? 'none'
      : storedGroupBy;

  const setGroupBy = (k: GroupKey): void => {
    setGroupByState(k);
    if (projectId) saveGroupBy(projectId, k);
  };

  // Tasks mapped to Task shape for TaskCard. Computed BEFORE any early
  // return so the useMemo below always runs in the same hook slot on
  // every render (React #310 otherwise - conditional hook order).
  // `activeOnly` is a query param the backend doesn't currently act on,
  // so archived tasks still come back over the wire - filter client-side
  // to match Triage/FocusToday/FlowLayout.
  const tasks: Task[] = useMemo(
    () => apiTasks.map(apiTaskToTask).filter((t) => !t.archived),
    [apiTasks]
  );

  // ── Filters ─────────────────────────────────────────────────────────────
  // Shared by every view of this project, carried in the URL and remembered
  // per project. Hide is the default; "Dim instead of hide" keeps every card.
  const {
    state: filterState,
    fields: filterFields,
    people,
    currentUserId,
    matchingIds,
    shownTasks: boardTasks,
    dimmedIds,
  } = useProjectTaskFilters(projectId, tasks);
  const { filters, dim } = filterState;
  const saveAsView = useSaveFiltersAsView();
  const hideUnmatched = matchingIds !== null && !dim;

  // Precompute the ordered lane index across all tasks so every column
  // renders the same rows in the same order - otherwise empty columns
  // would silently skip lanes and the swimlane grid misaligns.
  const laneIndex = useMemo(
    () => buildLaneIndex(boardTasks, groupBy, groupFields),
    [boardTasks, groupBy, groupFields]
  );

  // Every column is a move destination for the non-drag path (B2).
  const moveTargets: MoveTarget[] = useMemo(
    () => columns.map((c) => ({ id: String(c.id), name: c.name })),
    [columns]
  );

  const columnRefs = useRef(new Map<string, HTMLDivElement | null>());
  const listRefs = useRef(new Map<string, HTMLDivElement | null>());
  const ghostRef = useRef<HTMLDivElement | null>(null);
  const justDraggedRef = useRef(false);
  const dragRef = useRef<DragSession | null>(null);

  if (projectLoading && !project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        {t('flow.project.loading')}
      </div>
    );
  }

  // A failed project list leaves `project` undefined exactly as a missing one
  // does, so a 500 used to tell the user their project had gone.
  if (projectLoadFailed || hasQueryFailed(columnsQuery)) {
    return (
      <ListLoadFailed
        onRetry={() => {
          refetchProjects();
          void columnsQuery.refetch();
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

  const knownStatusIds = new Set(columns.map((c) => c.id));
  // Cards keep a hand-arranged order. With some hidden, a drop would be placed
  // against cards the user cannot see, so a filtered board (like a swimlane
  // one) only moves cards between columns.
  const canReorder = groupBy === 'none' && !hideUnmatched;
  // A status filter narrows the columns themselves; an empty column for every
  // unchosen status would only push the chosen ones off screen.
  const shownColumns = columns
    .map((col, index) => ({ col, index }))
    .filter(({ col }) => filters.statuses.length === 0 || filters.statuses.includes(col.id));
  const now = Date.now();

  /** Every card on a column, in board order. A task whose status is not a column here sits in the first. */
  function allTasksForColumn(col: ApiStatus, colIndex: number): Task[] {
    return orderBoardTasks(
      boardTasks.filter((t) => {
        const own = statusIdOf(t);
        if (own === col.id) return true;
        return colIndex === 0 && (!own || !knownStatusIds.has(own));
      })
    );
  }

  /** The cards a closed column hides until "Show older" is chosen. */
  function olderTasksForColumn(col: ApiStatus, colIndex: number): Task[] {
    if (!isClosedBucket(col.bucket) || showOlder.has(col.id)) return [];
    return allTasksForColumn(col, colIndex).filter((t) => !closedRecently(t, col.bucket, now));
  }

  function tasksForColumn(col: ApiStatus, colIndex: number): Task[] {
    const hidden = new Set(olderTasksForColumn(col, colIndex).map((t) => t.id));
    return allTasksForColumn(col, colIndex).filter((t) => !hidden.has(t.id));
  }

  // Pointer-based card drag (B1). HTML5 drag-and-drop never fired on touch, so
  // the board was mouse-only; pointer events cover mouse, touch and pen in one
  // path.
  //
  // Mouse and pen can start a drag anywhere on a card; the card suppresses its
  // open click only after the movement threshold is crossed. Touch starts on
  // the grip so the board remains scrollable elsewhere.
  const measureColumns = (): ColumnRect[] => {
    const out: ColumnRect[] = [];
    for (const [id, el] of columnRefs.current) {
      if (!el) continue;
      const r = el.getBoundingClientRect();
      out.push({ id, left: r.left, right: r.right, top: r.top, bottom: r.bottom });
    }
    return out;
  };

  const positionGhost = (): void => {
    const drag = dragRef.current;
    const el = ghostRef.current;
    if (!drag || !el) return;
    el.style.transform = `translate3d(${drag.x - drag.offsetX}px, ${drag.y - drag.offsetY}px, 0) rotate(1.5deg) scale(1.02)`;
  };

  /** Which column, and which place in it, the pointer is over right now. */
  const slotAt = (drag: DragSession, x: number, y: number): DropSlot | null => {
    const columnId = columnIdAtPoint(measureColumns(), x, y);
    if (columnId === null) return null;
    const list = listRefs.current.get(columnId);
    const slots = drag.slots.get(columnId) ?? [];
    if (!list) return { columnId, index: slots.length };
    const y0 = list.getBoundingClientRect().top - list.scrollTop;
    return { columnId, index: insertionIndex(slots, drag.taskId, y - y0) };
  };

  const endSession = (): void => {
    const drag = dragRef.current;
    if (!drag) return;
    drag.cleanup();
    dragRef.current = null;
    setDragView(null);
    setGhost(null);
  };

  const commitDrop = (drag: DragSession, slot: DropSlot | null): void => {
    // Released outside every column: the drag is cancelled.
    if (slot === null) return;
    if (!drag.reorder) {
      // Swimlanes: a card changes column but has no single place to take.
      const target = resolveDropTarget(slot.columnId, drag.source.columnId);
      if (target !== null) moveTask.mutate({ taskId: drag.taskId, columnId: target });
      return;
    }
    if (isSamePlace(drag.source, slot)) return;
    // Hidden older cards keep their place after the visible ones.
    const colIndex = columns.findIndex((c) => c.id === slot.columnId);
    const hiddenIds =
      colIndex < 0
        ? []
        : olderTasksForColumn(columns[colIndex], colIndex)
            .map((t) => t.id)
            .filter((id) => id !== drag.taskId);
    moveTask.mutate({
      taskId: drag.taskId,
      columnId: slot.columnId,
      orderedTaskIds: [
        ...orderedIdsAfterDrop(drag.columnIds.get(slot.columnId) ?? [], drag.taskId, slot.index),
        ...hiddenIds,
      ],
    });
  };

  // Live card drag. Pointer events cover mouse, touch and pen in one path.
  //
  // The card under the pointer never moves in the DOM: a touch pointer keeps
  // targeting the node it went down on, so re-parenting it mid-drag would end the
  // gesture. It stays where it is, hidden; a placeholder shows where it would land and the
  // cards around it slide aside; a floating copy follows the pointer. The
  // listeners sit on `window` for the same reason, because the pointer is off
  // the card almost at once.
  const handleDragPointerDown = (
    e: React.PointerEvent,
    taskId: string,
    sourceColumnId: string
  ): void => {
    if (!canEdit || dragRef.current) return;
    const task = tasks.find((candidate) => candidate.id === taskId);
    if (!task) return;
    e.stopPropagation();

    const card = e.currentTarget as HTMLElement;
    const cardRect = card.getBoundingClientRect();
    const columnIds = new Map<string, string[]>();
    const slots = new Map<string, CardSlot[]>();
    columns.forEach((col, i) => {
      const key = String(col.id);
      columnIds.set(
        key,
        tasksForColumn(col, i).map((candidate) => candidate.id)
      );
      slots.set(key, measureSlots(listRefs.current.get(key)));
    });
    const sourceIndex = (columnIds.get(sourceColumnId) ?? []).indexOf(taskId);

    const drag: DragSession = {
      pointerId: e.pointerId,
      taskId,
      task,
      source: { columnId: sourceColumnId, index: Math.max(0, sourceIndex) },
      reorder: canReorder,
      startX: e.clientX,
      startY: e.clientY,
      x: e.clientX,
      y: e.clientY,
      offsetX: e.clientX - cardRect.left,
      offsetY: e.clientY - cardRect.top,
      width: cardRect.width,
      height: cardRect.height,
      columnIds,
      slots,
      slot: null,
      active: false,
      cleanup: () => undefined,
    };
    drag.slot = drag.source;

    const previousCursor = document.body.style.cursor;
    const previousSelect = document.body.style.userSelect;

    const onMove = (ev: PointerEvent): void => {
      if (ev.pointerId !== drag.pointerId) return;
      drag.x = ev.clientX;
      drag.y = ev.clientY;
      if (!drag.active) {
        if (!exceedsDragThreshold(ev.clientX - drag.startX, ev.clientY - drag.startY)) return;
        drag.active = true;
        document.body.style.cursor = 'grabbing';
        document.body.style.userSelect = 'none';
        setGhost({ task: drag.task, width: drag.width });
        setDragView({
          taskId: drag.taskId,
          source: drag.source,
          slot: drag.source,
          height: drag.height,
        });
      }
      if (ev.cancelable) ev.preventDefault();
      positionGhost();
      const slot = slotAt(drag, ev.clientX, ev.clientY);
      if (!sameSlot(slot, drag.slot)) {
        drag.slot = slot;
        setDragView((view) => (view ? { ...view, slot } : view));
      }
    };
    const onUp = (ev: PointerEvent): void => {
      if (ev.pointerId !== drag.pointerId) return;
      const { active, slot } = drag;
      endSession();
      if (!active) return;
      // The click that follows a drag must not open the task.
      justDraggedRef.current = true;
      window.setTimeout(() => {
        justDraggedRef.current = false;
      }, 0);
      commitDrop(drag, slot);
    };
    const onCancel = (ev: PointerEvent): void => {
      if (ev.pointerId === drag.pointerId) endSession();
    };
    const onKey = (ev: KeyboardEvent): void => {
      if (ev.key === 'Escape') endSession();
    };

    window.addEventListener('pointermove', onMove, { passive: false });
    window.addEventListener('pointerup', onUp, true);
    window.addEventListener('pointercancel', onCancel, true);
    window.addEventListener('keydown', onKey);
    drag.cleanup = () => {
      window.removeEventListener('pointermove', onMove);
      window.removeEventListener('pointerup', onUp, true);
      window.removeEventListener('pointercancel', onCancel, true);
      window.removeEventListener('keydown', onKey);
      document.body.style.cursor = previousCursor;
      document.body.style.userSelect = previousSelect;
    };
    dragRef.current = drag;
  };

  // The card's own pointer-up runs after the capture-phase listener above, so by
  // then this already says whether the gesture was a drag.
  const endCardPointer = (): boolean => justDraggedRef.current;

  /** The placeholder this column should show, if the dragged card would land in it. */
  const placeholderFor = (columnId: string): { index: number; height: number } | null => {
    if (!dragView) return null;
    const at = dragView.slot ?? dragView.source;
    return at.columnId === columnId ? { index: at.index, height: dragView.height } : null;
  };

  const handleAddColumn = () => {
    const name = newColumnName.trim();
    if (!name) {
      toast.error(t('flowToasts.column.nameRequired'));
      return;
    }
    createStatus.mutate({
      name,
      bucket: 'active',
      color: STATUS_COLORS[4],
      ...insertionPlan(columns, columns.length),
    });
    setNewColumnName('');
    setAddingColumn(false);
  };

  const insertStatus = (name: string): void => {
    if (insertAt === null) return;
    createStatus.mutate(
      { name, bucket: 'active', color: STATUS_COLORS[4], ...insertionPlan(columns, insertAt) },
      { onSuccess: () => setInsertAt(null) }
    );
  };

  const writeStatus = (col: ApiStatus, patch: Partial<Pick<ApiStatus, 'name' | 'color'>>) =>
    updateStatus.mutate({
      id: col.id,
      name: patch.name ?? col.name,
      bucket: col.bucket,
      color: patch.color ?? col.color,
      sortOrder: col.sortOrder,
      isDefault: col.isDefault,
    });

  if (
    (columnsQuery.isPending && !hasQueryFailed(columnsQuery)) ||
    (tasksQuery.isPending && !hasQueryFailed(tasksQuery))
  ) {
    return (
      <div className="flex flex-col h-full">
        <ProjectViewTabs
          project={project}
          projectId={projectId}
          taskCount={tasks.length}
          onNewTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
        />
        <div className="flex-1 flex items-center justify-center">
          <p className="text-flow-meta" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flow.project.loading')}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={t('flow.board.pageTitle')} />
      <ProjectViewTabs
        project={project}
        projectId={projectId}
        taskCount={tasks.length}
        onNewTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
        rightSlot={
          <div className="flex items-center gap-1.5">
            <GroupByPicker groupBy={groupBy} options={groupOptions} onChange={setGroupBy} />
            {canManageBoard && (
              <button
                type="button"
                onClick={() => setStatusSettingsOpen(true)}
                aria-label={t('flow.status.openSettings')}
                className="flow-focus flex h-[44px] w-[44px] items-center justify-center rounded-md md:h-8 md:w-8"
                style={{ color: 'var(--flow-text-muted)' }}
              >
                <ListChecks size={14} aria-hidden="true" />
              </button>
            )}
            <button
              type="button"
              onClick={() => setBoardSettingsOpen(true)}
              aria-label={t('flow.board.openSettings')}
              className="flow-focus flex h-[44px] w-[44px] items-center justify-center rounded-md md:h-8 md:w-8"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <Settings2 size={14} aria-hidden="true" />
            </button>
          </div>
        }
      />

      <FilterBar
        state={filterState}
        currentUserId={currentUserId}
        fields={filterFields}
        people={people}
        visible={matchingIds?.size ?? tasks.length}
        total={tasks.length}
        savingView={saveAsView.pending}
        onSaveAsView={() =>
          saveAsView.save(
            t('flow.filters.savedViewName', { project: project.name }),
            projectFiltersForView(filters, projectId, columns)
          )
        }
      />

      <BoardSettingsPanel
        open={boardSettingsOpen}
        onClose={() => setBoardSettingsOpen(false)}
        columns={columns}
        canManage={canManageBoard}
        groupBy={groupBy}
        fieldGroupOptions={fieldGroupOptions(groupFields)}
        onGroupByChange={setGroupBy}
        onManageStatuses={() => {
          setBoardSettingsOpen(false);
          setStatusSettingsOpen(true);
        }}
      />

      <StatusSettingsPanel
        open={statusSettingsOpen}
        onClose={() => setStatusSettingsOpen(false)}
        projectId={projectId}
        canManage={canManageBoard}
      />

      {/* Tightened kanban canvas - the previous p-6 / gap-4 combo wasted
          too much room on laptop screens. These values stay reasonable
          down to narrow viewports because the column snaps to 248px. */}
      <div className="min-h-0 flex-1 overflow-auto">
        <div className="flex h-full min-w-max gap-2 px-3 py-3">
          {shownColumns.map(({ col, index: i }) => (
            <KanbanColumn
              key={String(col.id)}
              col={col}
              lanes={buildLanes(laneIndex, tasksForColumn(col, i), groupBy)}
              dimmedIds={dimmedIds}
              groupBy={groupBy}
              isOver={dragOver === String(col.id)}
              selectedTaskIds={selectedTaskIds}
              canEdit={canEdit}
              canCreate={canCreate}
              canManage={canManageBoard}
              draggingTaskId={draggingTaskId}
              moveTargets={moveTargets}
              registerRef={(el) => columnRefs.current.set(String(col.id), el)}
              registerList={(el) => listRefs.current.set(String(col.id), el)}
              reorder={canReorder}
              placeholder={placeholderFor(String(col.id))}
              onMoveTask={(taskId, columnId) => moveTask.mutate({ taskId, columnId })}
              onDragPointerDown={handleDragPointerDown}
              onDragPointerEnd={endCardPointer}
              onSelect={toggleSelectTask}
              onAddTask={() => useFlowStore.getState().setQuickCreateOpen(true, String(col.id))}
              hiddenCount={olderTasksForColumn(col, i).length}
              onShowOlder={() => setShowOlder((prev) => new Set(prev).add(col.id))}
              onRename={(name) => writeStatus(col, { name })}
              onRecolor={(color) => writeStatus(col, { color })}
              onInsert={(side) => setInsertAt(side === 'before' ? i : i + 1)}
              canMoveBefore={i > 0}
              canMoveAfter={i < columns.length - 1}
              onMove={(side) => {
                const ids = movedStatusIds(columns, i, side === 'before' ? i - 1 : i + 1);
                if (ids) reorderStatuses.mutate(ids);
              }}
              deleteBlocked={isOnlyDoneStatus(col, columns)}
              onDelete={() => setPendingDeleteColumn(col)}
            />
          ))}

          {/* Add column */}
          {canManageBoard && (
            <div className="w-[248px] shrink-0">
              {addingColumn ? (
                <div
                  className="rounded-md p-2"
                  style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
                >
                  <input
                    autoFocus
                    value={newColumnName}
                    onChange={(e) => setNewColumnName(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') handleAddColumn();
                      if (e.key === 'Escape') setAddingColumn(false);
                    }}
                    placeholder={t('flow.board.columnName')}
                    className="w-full bg-transparent text-flow-meta font-medium mb-1.5"
                    style={{ color: 'var(--flow-text)' }}
                  />
                  <div className="flex items-center gap-1.5">
                    <button
                      type="button"
                      onClick={handleAddColumn}
                      disabled={!newColumnName.trim() || createStatus.isPending}
                      className="h-6 px-2.5 rounded-sm text-flow-meta font-semibold disabled:opacity-40"
                      style={{
                        background: 'var(--flow-action-bg)',
                        color: 'var(--flow-on-accent)',
                      }}
                    >
                      {t('flow.taskDetail.add')}
                    </button>
                    <button
                      type="button"
                      onClick={() => setAddingColumn(false)}
                      className="h-6 px-2 rounded-sm text-flow-meta hover:bg-[var(--flow-hover)]"
                      style={{ color: 'var(--flow-text-muted)' }}
                    >
                      {t('flowCommon.cancel')}
                    </button>
                  </div>
                </div>
              ) : (
                <button
                  type="button"
                  onClick={() => setAddingColumn(true)}
                  className="w-full flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-flow-meta transition-colors hover:bg-[var(--flow-hover)]"
                  style={{
                    color: 'var(--flow-text-dim)',
                    border: '1px dashed var(--flow-border)',
                  }}
                >
                  <Plus size={12} />
                  {t('flow.board.addColumn')}
                </button>
              )}
            </div>
          )}
        </div>
      </div>
      <DeleteStatusDialog
        projectId={projectId}
        status={pendingDeleteColumn}
        statuses={columns}
        onClose={() => setPendingDeleteColumn(null)}
      />
      <NewStatusDialog
        open={insertAt !== null}
        onOpenChange={(open) => !open && setInsertAt(null)}
        onCreate={insertStatus}
        pending={createStatus.isPending}
      />
      {ghost &&
        createPortal(
          // The copy that follows the pointer. It is positioned straight on the
          // element by the move handler, so a drag costs no React render per move.
          <div
            ref={(el) => {
              ghostRef.current = el;
              if (el) positionGhost();
            }}
            aria-hidden="true"
            style={{
              position: 'fixed',
              top: 0,
              left: 0,
              width: ghost.width,
              zIndex: 1000,
              pointerEvents: 'none',
              willChange: 'transform',
              filter: 'drop-shadow(0 12px 24px rgba(0, 0, 0, 0.28))',
            }}
          >
            <TaskCard task={ghost.task} />
          </div>,
          document.body
        )}
    </div>
  );
};

// ── Column component ────────────────────────────────────────────────────────

type Side = 'before' | 'after';

const KanbanColumn: React.FC<{
  col: ApiStatus;
  lanes: Lane[];
  /** Cards that do not match the filters, faded in "Dim instead of hide". */
  dimmedIds?: ReadonlySet<string> | null;
  groupBy: GroupKey;
  isOver: boolean;
  selectedTaskIds: string[];
  canEdit: boolean;
  canCreate: boolean;
  canManage: boolean;
  draggingTaskId: string | null;
  moveTargets: readonly MoveTarget[];
  registerRef: (el: HTMLDivElement | null) => void;
  registerList: (el: HTMLDivElement | null) => void;
  /** Whether cards can be placed (false in swimlane view, where a drag only changes column). */
  reorder: boolean;
  /** Where the dragged card would land in this column, when it would. */
  placeholder: { index: number; height: number } | null;
  onMoveTask: (taskId: string, columnId: string) => void;
  onDragPointerDown: (e: React.PointerEvent, taskId: string, sourceColumnId: string) => void;
  onDragPointerEnd: (e: React.PointerEvent) => boolean;
  onSelect: (id: string) => void;
  onAddTask: () => void;
  /** Older cards a closed column is hiding. */
  hiddenCount: number;
  onShowOlder: () => void;
  onRename: (name: string) => void;
  onRecolor: (color: string) => void;
  onInsert: (side: Side) => void;
  canMoveBefore: boolean;
  canMoveAfter: boolean;
  onMove: (side: Side) => void;
  /** The project's only done status, which cannot be deleted. */
  deleteBlocked: boolean;
  onDelete: () => void;
}> = ({
  col,
  lanes,
  dimmedIds = null,
  groupBy,
  isOver,
  selectedTaskIds,
  canEdit,
  canCreate,
  canManage,
  draggingTaskId,
  moveTargets,
  registerRef,
  registerList,
  reorder,
  placeholder,
  onMoveTask,
  onDragPointerDown,
  onDragPointerEnd,
  onSelect,
  onAddTask,
  hiddenCount,
  onShowOlder,
  onRename,
  onRecolor,
  onInsert,
  canMoveBefore,
  canMoveAfter,
  onMove,
  deleteBlocked,
  onDelete,
}) => {
  const { t } = useTranslation();
  const [editing, setEditing] = useState(false);
  const [editName, setEditName] = useState(col.name);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (editing) inputRef.current?.focus();
  }, [editing]);

  const commitRename = () => {
    const name = editName.trim();
    if (name && name !== col.name) onRename(name);
    setEditing(false);
  };

  // Both card branches below render the same affordances; keeping them in one
  // place stops the swimlane view from quietly losing the move menu.
  const renderCardControls = (taskId: string, title: string): React.ReactNode => (
    <>
      {canEdit && (
        <button
          type="button"
          aria-hidden
          tabIndex={-1}
          data-flow-drag-handle
          onClick={(e) => e.stopPropagation()}
          className="flex h-8 w-8 cursor-grab items-center justify-center rounded-sm hover:bg-[var(--flow-hover)] active:cursor-grabbing"
          style={{ color: 'var(--flow-text-dim)', touchAction: 'none' }}
        >
          <GripVertical size={12} />
        </button>
      )}
      <MoveToColumnMenu
        taskTitle={title}
        targets={moveTargets}
        currentColumnId={String(col.id)}
        onMove={(columnId) => onMoveTask(taskId, columnId)}
        disabled={!canEdit}
      />
    </>
  );

  // The cards of one lane. While a card is being dragged it stays mounted but
  // leaves the layout, and a placeholder holds the place it would take; the
  // cards around the placeholder slide aside through TaskCard's layout animation.
  const renderCards = (laneTasks: Task[]): React.ReactNode[] => {
    const out: React.ReactNode[] = [];
    let visible = 0;
    const slotAt = (): React.ReactNode =>
      placeholder && (
        <DropPlaceholder key="drag-placeholder" height={placeholder.height} color={col.color} />
      );
    for (const task of laneTasks) {
      const lifted = reorder && draggingTaskId === task.id;
      if (!lifted) {
        if (reorder && placeholder && placeholder.index === visible) out.push(slotAt());
        visible += 1;
      }
      out.push(
        <div
          key={task.id}
          data-board-card={task.id}
          data-filter-dimmed={dimmedIds?.has(task.id) ? '' : undefined}
          style={
            lifted
              ? HIDDEN_CARD_STYLE
              : dimmedIds?.has(task.id)
                ? { opacity: 0.35, transition: 'opacity 0.15s' }
                : undefined
          }
        >
          <TaskCard
            task={task}
            controls={renderCardControls(task.id, task.title)}
            dimmed={!reorder && draggingTaskId === task.id}
            selected={selectedTaskIds.includes(task.id)}
            onSelect={() => onSelect(task.id)}
            onDragPointerDown={
              canEdit ? (e, id) => onDragPointerDown(e, id, String(col.id)) : undefined
            }
            onDragPointerEnd={onDragPointerEnd}
          />
        </div>
      );
    }
    if (reorder && placeholder && placeholder.index >= visible) out.push(slotAt());
    return out;
  };

  // Total count for the column header pill (stays accurate regardless
  // of group-by - it's the sum of tasks across all lanes).
  const colTaskCount = lanes.reduce((n, l) => n + l.tasks.length, 0);
  // While dimming, the pill reads "matching/all" so the column still says how
  // much of it the filter picked out.
  const colMatchCount = dimmedIds
    ? lanes.reduce((n, l) => n + l.tasks.filter((task) => !dimmedIds.has(task.id)).length, 0)
    : null;

  return (
    <div
      ref={registerRef}
      data-board-column={String(col.id)}
      // 296px, inside the 280-320 the board rule asks for. It was 248, which
      // squeezed a two-line title to the point that the card carried less than
      // the row beside it (rtylr#1078, S-08). The viewport clamp stays, so a
      // 390px phone still gets one full-width column rather than a horizontal
      // scrollbar on the page itself.
      className="flex w-[min(296px,calc(100vw-1.5rem))] min-w-[min(296px,calc(100vw-1.5rem))] flex-col rounded-md p-1.5 sm:w-[296px] sm:min-w-[296px]"
      style={{
        background: isOver ? withAlpha(col.color, 5) : 'transparent',
        border: isOver ? `1px dashed ${col.color}` : '1px solid transparent',
        transition: 'background 0.15s, border 0.15s',
      }}
    >
      {/* Column header - tighter. 24px tall, no wasteful vertical padding. */}
      <div className="flex items-center gap-1.5 px-1.5 h-7 mb-1.5 group/header">
        <span className="h-2 w-2 rounded-full shrink-0" style={{ background: col.color }} />

        {editing ? (
          <input
            ref={inputRef}
            value={editName}
            onChange={(e) => setEditName(e.target.value)}
            onBlur={commitRename}
            onKeyDown={(e) => {
              if (e.key === 'Enter') commitRename();
              if (e.key === 'Escape') {
                setEditName(col.name);
                setEditing(false);
              }
            }}
            className="flex-1 bg-transparent text-flow-meta font-semibold min-w-0"
            style={{ color: 'var(--flow-text)' }}
          />
        ) : (
          <button
            type="button"
            disabled={!canManage}
            onClick={() => setEditing(true)}
            className="flex-1 text-start text-flow-meta font-semibold truncate hover:opacity-70 transition-opacity"
            style={{ color: 'var(--flow-text)' }}
          >
            {displayStatusName(col.name, t)}
          </button>
        )}

        <span
          className="text-flow-meta px-1.5 py-0.5 rounded-full font-medium shrink-0"
          style={{ background: withAlpha(col.color, 13), color: onTint(col.color) }}
        >
          {colMatchCount === null ? colTaskCount : `${colMatchCount}/${colTaskCount}`}
        </span>

        {canManage && (
          <ColumnMenu
            col={col}
            onRename={() => {
              setEditName(col.name);
              setEditing(true);
            }}
            onRecolor={onRecolor}
            onInsert={onInsert}
            canMoveBefore={canMoveBefore}
            canMoveAfter={canMoveAfter}
            onMove={onMove}
            deleteBlocked={deleteBlocked}
            onDelete={onDelete}
          />
        )}
      </div>

      {/* Add task, directly under the header and on every column.
          It used to appear only in empty columns, centred vertically, so the
          two columns that had cards offered no way to add one; and the
          alternative in the header was hover-only at 20px, which a touch screen
          cannot reach at all (rtylr#1078, Asana rule 1.2). 32px on a pointer,
          44px where a finger has to hit it. */}
      {canCreate && (
        <button
          type="button"
          onClick={onAddTask}
          aria-label={t('flow.board.addTaskLabel')}
          className="flow-focus mb-1.5 flex min-h-11 w-full items-center gap-1.5 rounded-md px-1.5 text-flow-meta font-medium transition-colors hover:bg-[var(--flow-hover)] md:min-h-8"
          style={{ color: 'var(--flow-text-muted)' }}
        >
          <Plus size={12} />
          {t('flow.board.addTask')}
        </button>
      )}

      {/* Cards */}
      <div
        ref={registerList}
        data-board-list
        className="relative flex flex-col gap-1.5 flex-1 min-h-[80px] overflow-y-auto"
      >
        {groupBy === 'none' ? (
          <>
            {renderCards(lanes[0]?.tasks ?? [])}
            {colTaskCount === 0 && !placeholder && hiddenCount === 0 && (
              // A sentence, not a second add control: the row above the cards is
              // already the way in, and this said "Add task" in hardcoded
              // English on a screen that ships in Arabic and French.
              <p
                className="flex flex-1 items-center justify-center rounded-md px-3 py-4 text-center text-flow-meta"
                style={{
                  color: 'var(--flow-text-muted)',
                  border: '1px dashed var(--flow-border)',
                }}
              >
                {t('flow.board.emptyColumn')}
              </p>
            )}
          </>
        ) : (
          // Swimlanes: render each lane with a mini-header + its tasks.
          // Lanes are rendered in the same order across every column so
          // rows line up horizontally.
          lanes.map((lane) => (
            <div key={lane.key} className="flex flex-col gap-1">
              <div
                className="flex items-center gap-1.5 px-1 pt-1.5 pb-0.5 text-flow-meta font-semibold uppercase tracking-wider"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <span className="truncate">{lane.labelKey ? t(lane.labelKey) : lane.label}</span>
                <span style={{ color: 'var(--flow-border)' }}>·</span>
                <span className="tabular-nums">{lane.tasks.length}</span>
              </div>
              <div className="flex flex-col gap-1.5">
                {lane.tasks.map((task) => (
                  <TaskCard
                    key={task.id}
                    task={task}
                    controls={renderCardControls(task.id, task.title)}
                    dimmed={draggingTaskId === task.id}
                    selected={selectedTaskIds.includes(task.id)}
                    onSelect={() => onSelect(task.id)}
                    onDragPointerDown={
                      canEdit ? (e, id) => onDragPointerDown(e, id, String(col.id)) : undefined
                    }
                    onDragPointerEnd={onDragPointerEnd}
                  />
                ))}
                {lane.tasks.length === 0 && (
                  <p
                    className="flex min-h-[28px] items-center justify-center rounded-md text-flow-meta"
                    style={{
                      color: 'var(--flow-text-muted)',
                      border: '1px dashed var(--flow-border)',
                    }}
                  >
                    {t('flow.board.emptyLane')}
                  </p>
                )}
              </div>
            </div>
          ))
        )}
        {hiddenCount > 0 && (
          <button
            type="button"
            onClick={onShowOlder}
            className="flow-focus mt-1 min-h-11 rounded-md px-1.5 text-start text-flow-meta font-medium underline-offset-2 hover:underline md:min-h-8"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.board.showOlder', { count: hiddenCount })}
          </button>
        )}
      </div>
    </div>
  );
};

// ── Column menu ─────────────────────────────────────────────────────────────

/**
 * Every change to a column, behind one labelled button. Delete sits last and
 * opens a dialog, so it can no longer fire from a stray hover click.
 */
const ColumnMenu: React.FC<{
  col: ApiStatus;
  onRename: () => void;
  onRecolor: (color: string) => void;
  onInsert: (side: Side) => void;
  canMoveBefore: boolean;
  canMoveAfter: boolean;
  onMove: (side: Side) => void;
  deleteBlocked: boolean;
  onDelete: () => void;
}> = ({
  col,
  onRename,
  onRecolor,
  onInsert,
  canMoveBefore,
  canMoveAfter,
  onMove,
  deleteBlocked,
  onDelete,
}) => {
  const { t } = useTranslation();
  const blockedReasonId = useId();
  // Rename, insert and delete move focus (an input, a dialog), so they run
  // after the menu has closed instead of racing its focus return.
  const afterClose = useRef<(() => void) | null>(null);
  const deferred = (run: () => void) => () => {
    afterClose.current = run;
  };
  const itemClass = 'text-flow-meta';
  const itemStyle = { color: 'var(--flow-text)' };
  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        aria-label={t('flow.board.columnMenu', { name: displayStatusName(col.name, t) })}
        className="flow-focus flex h-11 w-11 shrink-0 items-center justify-center rounded-sm text-[var(--flow-text-dim)] hover:bg-[var(--flow-hover)] md:h-6 md:w-6"
      >
        <MoreHorizontal size={13} aria-hidden="true" />
      </DropdownMenuTrigger>
      <DropdownMenuContent
        onCloseAutoFocus={(event) => {
          const run = afterClose.current;
          if (!run) return;
          afterClose.current = null;
          event.preventDefault();
          run();
        }}
        align="end"
        className="min-w-[200px]"
        style={{
          background: 'var(--flow-card)',
          border: '1px solid var(--flow-border)',
          boxShadow: 'var(--flow-shadow)',
        }}
      >
        <DropdownMenuItem className={itemClass} style={itemStyle} onSelect={deferred(onRename)}>
          <Pencil size={13} aria-hidden="true" />
          {t('flow.board.renameStatus')}
        </DropdownMenuItem>
        <DropdownMenuLabel className="text-flow-meta" style={{ color: 'var(--flow-text-muted)' }}>
          {t('flow.board.statusColor')}
        </DropdownMenuLabel>
        <DropdownMenuGroup className="flex flex-wrap gap-1 px-2 pb-1.5">
          {STATUS_COLORS.map((c) => (
            <DropdownMenuItem
              key={c}
              onSelect={() => onRecolor(c)}
              aria-label={t('flow.status.chooseColorLabel', { color: c })}
              aria-current={col.color === c ? 'true' : undefined}
              className="h-6 w-6 justify-center rounded-full p-0"
              style={{
                background: c,
                outline: col.color === c ? `2px solid ${c}` : 'none',
                outlineOffset: 2,
              }}
            />
          ))}
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className={itemClass}
          style={itemStyle}
          onSelect={deferred(() => onInsert('before'))}
        >
          <Plus size={13} aria-hidden="true" />
          {t('flow.board.addStatusBefore')}
        </DropdownMenuItem>
        <DropdownMenuItem
          className={itemClass}
          style={itemStyle}
          onSelect={deferred(() => onInsert('after'))}
        >
          <Plus size={13} aria-hidden="true" />
          {t('flow.board.addStatusAfter')}
        </DropdownMenuItem>
        <DropdownMenuItem
          className={itemClass}
          style={itemStyle}
          disabled={!canMoveBefore}
          onSelect={() => onMove('before')}
        >
          <ArrowLeft size={13} aria-hidden="true" className="rtl:rotate-180" />
          {t('flow.board.moveStatusBefore')}
        </DropdownMenuItem>
        <DropdownMenuItem
          className={itemClass}
          style={itemStyle}
          disabled={!canMoveAfter}
          onSelect={() => onMove('after')}
        >
          <ArrowRight size={13} aria-hidden="true" className="rtl:rotate-180" />
          {t('flow.board.moveStatusAfter')}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className={itemClass}
          style={{ color: 'var(--flow-danger-fg)' }}
          disabled={deleteBlocked}
          aria-describedby={deleteBlocked ? blockedReasonId : undefined}
          onSelect={deferred(onDelete)}
        >
          <Trash2 size={13} aria-hidden="true" />
          {t('flow.board.deleteStatus')}
        </DropdownMenuItem>
        {deleteBlocked && (
          <p
            id={blockedReasonId}
            className="max-w-[220px] px-2 pb-1.5 text-flow-meta"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.status.lastDoneProtected')}
          </p>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
};

// ── Drag placeholder ────────────────────────────────────────────────────────

/** The empty slot a dragged card would drop into, sized to the card. */
const DropPlaceholder: React.FC<{ height: number; color: string }> = ({ height, color }) => (
  <motion.div
    layout
    initial={{ opacity: 0 }}
    animate={{ opacity: 1 }}
    transition={{ duration: 0.12 }}
    aria-hidden="true"
    data-board-placeholder
    className="shrink-0 rounded-lg"
    style={{
      height,
      border: `1.5px dashed ${color}`,
      background: withAlpha(color, 8),
    }}
  />
);

// ── Group-by picker ─────────────────────────────────────────────────────────

const GroupByPicker: React.FC<{
  groupBy: GroupKey;
  options: GroupOption[];
  onChange: (k: GroupKey) => void;
}> = ({ groupBy, options, onChange }) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  const close = useCallback((): void => setOpen(false), []);
  useClickOutside(ref, close);

  const current = options.find((o) => o.key === groupBy) ?? GROUP_OPTIONS[0]!;
  const optionLabel = (o: GroupOption): string => (o.labelKey ? t(o.labelKey) : (o.label ?? ''));

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-1.5 h-7 px-2 rounded-md text-flow-meta font-medium transition-colors hover:bg-[var(--flow-hover)]"
        style={{
          background: 'var(--flow-surface)',
          border: '1px solid var(--flow-border)',
          color: 'var(--flow-text-muted)',
        }}
      >
        <span style={{ color: 'var(--flow-text-dim)' }}>{t('flow.board.groupByLabel')}</span>
        <span style={{ color: 'var(--flow-text)' }}>{optionLabel(current)}</span>
        <ChevronDown size={11} />
      </button>
      {open && (
        <div
          className="absolute end-0 top-full mt-1 z-20 min-w-[140px] rounded-md py-1 shadow-md"
          style={{
            background: 'var(--flow-card)',
            border: '1px solid var(--flow-border)',
          }}
        >
          {options.map((o) => (
            <button
              type="button"
              key={o.key}
              onClick={() => {
                onChange(o.key);
                setOpen(false);
              }}
              className="w-full text-start px-2.5 py-1.5 text-flow-meta transition-colors hover:bg-[var(--flow-hover)]"
              style={{
                color: o.key === groupBy ? 'var(--flow-text)' : 'var(--flow-text-muted)',
                fontWeight: o.key === groupBy ? 600 : 400,
              }}
            >
              {optionLabel(o)}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default ProjectBoard;
