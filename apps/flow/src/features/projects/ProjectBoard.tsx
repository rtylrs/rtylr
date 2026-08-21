import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { HelmetContainer } from '@voxire/components';
import { useClickOutside } from '@voxire/utils';
import { toast } from 'sonner';
import { useParams } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ChevronDown, Plus, Settings2, Trash2 } from 'lucide-react';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useTasks } from '../tasks/useTasks';
import {
  useProjectColumns,
  useCreateColumn,
  useUpdateColumn,
  useDeleteColumn,
  useMoveTaskToColumn,
} from './useColumns';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import type { BoardGroupKey } from '../../shared/lib/boardUtils';
import { useProjectRealtime } from '../../shared/lib/useProjectRealtime';
import TaskCard from '../../shared/components/TaskCard';
import ConfirmDialog from '../../shared/components/ConfirmDialog';
import ProjectViewTabs from './ProjectViewTabs';
import { useCurrentProject } from './useCurrentProject';
import { useCanEditTask, useCanDeleteResource } from '../../shared/lib/currentUser';
import BoardSettingsPanel from './components/BoardSettingsPanel';
import type { ApiProjectColumn } from '../../shared/services/flowApi';
import type { Task } from '../../shared/stores/flowStore';

// ── Group-by (swimlanes) ────────────────────────────────────────────────
// "None" = plain kanban columns. Any other value partitions each column's
// tasks into horizontal lanes keyed by that attribute (assignee, priority,
// etc.). The choice persists per-project in localStorage so switching
// projects doesn't reset it.

/** Re-exported name kept local for readability; the type lives in boardUtils. */
type GroupKey = BoardGroupKey;

const GROUP_OPTIONS: { key: GroupKey; label: string }[] = [
  { key: 'none', label: 'None' },
  { key: 'assignee', label: 'Assignee' },
  { key: 'priority', label: 'Priority' },
  { key: 'due', label: 'Due window' },
];

const PRIORITY_ORDER: Task['priority'][] = ['urgent', 'high', 'medium', 'low', 'no_priority'];

const PRIORITY_LABEL: Record<Task['priority'], string> = {
  urgent: 'Urgent',
  high: 'High',
  medium: 'Medium',
  low: 'Low',
  no_priority: 'No priority',
};

interface Lane {
  key: string;
  label: string;
  tasks: Task[];
}

/**
 * Partition a column's tasks into lanes. Every column renders the same
 * set of lane keys in the same order so rows line up horizontally across
 * columns - this is the "swimlane" effect.
 */
function buildLanes(
  allLaneKeys: { key: string; label: string }[],
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
  return allLaneKeys.map(({ key, label }) => ({
    key,
    label,
    tasks: byKey.get(key) ?? [],
  }));
}

function laneKeyFor(t: Task, groupBy: GroupKey): string {
  if (groupBy === 'assignee') return t.assignee || 'Unassigned';
  if (groupBy === 'priority') return t.priority;
  if (groupBy === 'due') {
    if (!t.dueDate) return 'No due date';
    const d = new Date(t.dueDate);
    const now = new Date();
    const days = Math.round((d.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
    if (days < 0) return 'Overdue';
    if (days === 0) return 'Today';
    if (days <= 7) return 'This week';
    if (days <= 30) return 'This month';
    return 'Later';
  }
  return '_all';
}

/**
 * Derive the deterministic ordered lane list across ALL tasks so every
 * column shows the same swimlanes in the same order. Without this, empty
 * columns would skip lanes and rows would misalign.
 */
function buildLaneIndex(tasks: Task[], groupBy: GroupKey): { key: string; label: string }[] {
  if (groupBy === 'none') return [{ key: '_all', label: '' }];

  if (groupBy === 'priority') {
    return PRIORITY_ORDER.map((p) => ({ key: p, label: PRIORITY_LABEL[p] }));
  }

  if (groupBy === 'due') {
    return [
      { key: 'Overdue', label: 'Overdue' },
      { key: 'Today', label: 'Today' },
      { key: 'This week', label: 'This week' },
      { key: 'This month', label: 'This month' },
      { key: 'Later', label: 'Later' },
      { key: 'No due date', label: 'No due date' },
    ];
  }

  // assignee - build from data, keep stable order by first-seen.
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
  const { workspaceId = PERSONAL_WORKSPACE_ID, projectId = '' } = useParams({ strict: false });

  const { selectedTaskIds, toggleSelectTask, activeWorkspaceId } = useFlowStore();
  const { project, isLoading: projectLoading } = useCurrentProject(workspaceId, projectId);
  const moveTask = useMoveTaskToColumn();
  const canEdit = useCanEditTask(projectId);
  // "Manage project" for board setup. A viewer still opens the panel and sees
  // why it is read-only, rather than the board's behaviour being unexplainable.
  const canManageBoard = useCanDeleteResource(projectId);

  useProjectRealtime(projectId);

  const { data: apiTasks = [] } = useTasks({ projectId, activeOnly: true, compact: true });
  const { data: columns = [], isLoading: columnsLoading } = useProjectColumns(projectId);
  const createColumn = useCreateColumn(projectId);
  const updateColumn = useUpdateColumn(projectId);
  const deleteColumn = useDeleteColumn(projectId);

  const [dragOver, setDragOver] = useState<string | null>(null); // column id
  const [addingColumn, setAddingColumn] = useState(false);
  const [newColumnName, setNewColumnName] = useState('');
  const [groupBy, setGroupByState] = useState<GroupKey>('none');
  const [pendingDeleteColumn, setPendingDeleteColumn] = useState<ApiProjectColumn | null>(null);
  const [boardSettingsOpen, setBoardSettingsOpen] = useState(false);

  // Hydrate group-by from localStorage once the project id is known.
  useEffect(() => {
    if (projectId) setGroupByState(loadGroupBy(projectId));
  }, [projectId]);

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

  // Precompute the ordered lane index across all tasks so every column
  // renders the same rows in the same order - otherwise empty columns
  // would silently skip lanes and the swimlane grid misaligns.
  const laneIndex = useMemo(() => buildLaneIndex(tasks, groupBy), [tasks, groupBy]);

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

  // Bucket tasks by column. Tasks with no sectionId go into first column.
  function tasksForColumn(col: ApiProjectColumn, colIndex: number): Task[] {
    // ApiProjectColumn's TypeScript annotation claims `id: string`, but
    // the wire sends a JSON number. Coerce on both sides so
    // `"5" === 5` (which would be false) doesn't make moved tasks
    // vanish from the board.
    const colKey = String(col.id);
    return tasks.filter((t) => {
      if (t.sectionId && String(t.sectionId) === colKey) return true;
      // Unsectioned tasks go into first column.
      if (colIndex === 0 && !t.sectionId) return true;
      return false;
    });
  }

  const handleDragStart = (e: React.DragEvent, taskId: string) => {
    e.dataTransfer.setData('text/plain', taskId);
    e.dataTransfer.effectAllowed = 'move';
  };

  const handleDrop = (e: React.DragEvent, colId: string) => {
    e.preventDefault();
    if (!canEdit) return;
    const taskId = e.dataTransfer.getData('text/plain');
    if (!taskId) return;
    moveTask.mutate({ taskId, columnId: colId });
    setDragOver(null);
  };

  const handleAddColumn = () => {
    const name = newColumnName.trim();
    if (!name) {
      toast.error(t('toasts.column.nameRequired'));
      return;
    }
    createColumn.mutate({
      workspaceId: activeWorkspaceId || project.workspaceId,
      name,
      color: '#6b7280',
      sortOrder: columns.length,
    });
    setNewColumnName('');
    setAddingColumn(false);
  };

  if (columnsLoading) {
    return (
      <div className="flex flex-col h-full">
        <ProjectViewTabs
          project={project}
          workspaceId={workspaceId}
          projectId={projectId}
          taskCount={tasks.length}
          onNewTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
        />
        <div className="flex-1 flex items-center justify-center">
          <p className="text-[12px]" style={{ color: 'var(--flow-text-dim)' }}>
            Loading board…
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
        workspaceId={workspaceId}
        projectId={projectId}
        taskCount={tasks.length}
        onNewTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
        rightSlot={
          <div className="flex items-center gap-1.5">
            <GroupByPicker groupBy={groupBy} onChange={setGroupBy} />
            <button
              type="button"
              onClick={() => setBoardSettingsOpen(true)}
              aria-label={t('flow.board.openSettings')}
              className="flow-focus flex h-11 w-11 items-center justify-center rounded-md md:h-8 md:w-8"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <Settings2 size={14} aria-hidden="true" />
            </button>
          </div>
        }
      />

      <BoardSettingsPanel
        open={boardSettingsOpen}
        onClose={() => setBoardSettingsOpen(false)}
        projectId={projectId}
        workspaceId={activeWorkspaceId || project.workspaceId}
        columns={columns}
        canManage={canManageBoard}
        groupBy={groupBy}
        onGroupByChange={setGroupBy}
      />

      {/* Tightened kanban canvas - the previous p-6 / gap-4 combo wasted
          too much room on laptop screens. These values stay reasonable
          down to narrow viewports because the column snaps to 248px. */}
      <div className="min-h-0 flex-1 overflow-auto">
        <div className="flex h-full min-w-max gap-2 px-3 py-3">
          {columns.map((col, i) => (
            <KanbanColumn
              key={String(col.id)}
              col={col}
              lanes={buildLanes(laneIndex, tasksForColumn(col, i), groupBy)}
              groupBy={groupBy}
              isOver={dragOver === String(col.id)}
              selectedTaskIds={selectedTaskIds}
              canEdit={canEdit}
              onDragOver={() => setDragOver(String(col.id))}
              onDragLeave={() => setDragOver(null)}
              onDrop={(e) => handleDrop(e, String(col.id))}
              onDragStart={handleDragStart}
              onSelect={toggleSelectTask}
              onAddTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
              onRename={(name) =>
                updateColumn.mutate({
                  id: String(col.id),
                  name,
                  color: col.color,
                  sortOrder: col.sortOrder,
                })
              }
              onDelete={() => setPendingDeleteColumn(col)}
            />
          ))}

          {/* Add column */}
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
                  className="w-full bg-transparent outline-none text-[12px] font-medium mb-1.5"
                  style={{ color: 'var(--flow-text)' }}
                />
                <div className="flex items-center gap-1.5">
                  <button
                    onClick={handleAddColumn}
                    disabled={!newColumnName.trim() || createColumn.isPending}
                    className="h-6 px-2.5 rounded text-[11px] font-semibold disabled:opacity-40"
                    style={{ background: 'var(--accent)', color: '#ffffff' }}
                  >
                    Add
                  </button>
                  <button
                    onClick={() => setAddingColumn(false)}
                    className="h-6 px-2 rounded text-[11px] hover:bg-[var(--flow-hover)]"
                    style={{ color: 'var(--flow-text-muted)' }}
                  >
                    Cancel
                  </button>
                </div>
              </div>
            ) : (
              <button
                onClick={() => setAddingColumn(true)}
                className="w-full flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-[11px] transition-colors hover:bg-[var(--flow-hover)]"
                style={{
                  color: 'var(--flow-text-dim)',
                  border: '1px dashed var(--flow-border)',
                }}
              >
                <Plus size={12} />
                Add column
              </button>
            )}
          </div>
        </div>
      </div>
      <ConfirmDialog
        open={pendingDeleteColumn !== null}
        onOpenChange={(o) => {
          if (!o) setPendingDeleteColumn(null);
        }}
        title={t('flow.board.deleteColumnLabel')}
        description={
          <>
            Delete the <strong>{pendingDeleteColumn?.name}</strong> column? Tasks in this column
            will need to be moved manually first or they will become orphaned.
          </>
        }
        confirmLabel="Delete column"
        pending={deleteColumn.isPending}
        onConfirm={() => {
          if (!pendingDeleteColumn) return;
          deleteColumn.mutate(String(pendingDeleteColumn.id), {
            onSettled: () => setPendingDeleteColumn(null),
          });
        }}
      />
    </div>
  );
};

// ── Column component ────────────────────────────────────────────────────────

const KanbanColumn: React.FC<{
  col: ApiProjectColumn;
  lanes: Lane[];
  groupBy: GroupKey;
  isOver: boolean;
  selectedTaskIds: string[];
  canEdit: boolean;
  onDragOver: () => void;
  onDragLeave: () => void;
  onDrop: (e: React.DragEvent) => void;
  onDragStart: (e: React.DragEvent, id: string) => void;
  onSelect: (id: string) => void;
  onAddTask: () => void;
  onRename: (name: string) => void;
  onDelete: () => void;
}> = ({
  col,
  lanes,
  groupBy,
  isOver,
  selectedTaskIds,
  canEdit,
  onDragOver,
  onDragLeave,
  onDrop,
  onDragStart,
  onSelect,
  onAddTask,
  onRename,
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

  // Total count for the column header pill (stays accurate regardless
  // of group-by - it's the sum of tasks across all lanes).
  const colTaskCount = lanes.reduce((n, l) => n + l.tasks.length, 0);

  return (
    <div
      onDragOver={(e) => {
        e.preventDefault();
        onDragOver();
      }}
      onDragLeave={onDragLeave}
      onDrop={onDrop}
      className="flex w-[min(248px,calc(100vw-1.5rem))] min-w-[min(248px,calc(100vw-1.5rem))] flex-col rounded-md p-1.5 sm:w-[248px] sm:min-w-[248px]"
      style={{
        background: isOver ? `${col.color}12` : 'transparent',
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
            className="flex-1 bg-transparent outline-none text-[12px] font-semibold min-w-0"
            style={{ color: 'var(--flow-text)' }}
          />
        ) : (
          <button
            onClick={() => setEditing(true)}
            className="flex-1 text-start text-[12px] font-semibold truncate hover:opacity-70 transition-opacity"
            style={{ color: 'var(--flow-text)' }}
          >
            {col.name}
          </button>
        )}

        <span
          className="text-[10px] px-1.5 py-0.5 rounded-full font-medium shrink-0"
          style={{ background: `${col.color}22`, color: col.color }}
        >
          {colTaskCount}
        </span>

        {/* Actions (visible on header hover) */}
        <div className="flex items-center gap-0.5 opacity-0 group-hover/header:opacity-100 transition-opacity shrink-0">
          <button
            onClick={onAddTask}
            className="h-5 w-5 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-dim)' }}
            title={t('flow.board.addTaskLabel')}
            aria-label={t('flow.board.addTaskLabel')}
          >
            <Plus size={11} />
          </button>
          <button type="button"
            onClick={onDelete}
            className="h-5 w-5 rounded flex items-center justify-center hover:bg-[var(--flow-hover)] hover:text-red-400"
            style={{ color: 'var(--flow-text-dim)' }}
            title={t('flow.board.deleteColumnLabel')}
            aria-label={t('flow.board.deleteColumnLabel')}
          >
            <Trash2 size={11} />
          </button>
        </div>
      </div>

      {/* Cards */}
      <div className="flex flex-col gap-1.5 flex-1 min-h-[80px] overflow-y-auto">
        {groupBy === 'none' ? (
          <>
            {lanes[0]?.tasks.map((task) => (
              <TaskCard
                key={task.id}
                task={task}
                draggable={canEdit}
                onDragStart={(e) => onDragStart(e, task.id)}
                selected={selectedTaskIds.includes(task.id)}
                onSelect={() => onSelect(task.id)}
              />
            ))}
            {colTaskCount === 0 && (
              <button type="button"
                onClick={onAddTask}
                className="flex-1 flex items-center justify-center rounded text-[11px] min-h-[44px] hover:bg-[var(--flow-hover)] transition-colors"
                style={{
                  color: 'var(--flow-text-dim)',
                  border: '1px dashed var(--flow-border)',
                }}
              >
                <Plus size={12} className="me-1" /> Add task
              </button>
            )}
          </>
        ) : (
          // Swimlanes: render each lane with a mini-header + its tasks.
          // Lanes are rendered in the same order across every column so
          // rows line up horizontally.
          lanes.map((lane) => (
            <div key={lane.key} className="flex flex-col gap-1">
              <div
                className="flex items-center gap-1.5 px-1 pt-1.5 pb-0.5 text-[10px] font-semibold uppercase tracking-wider"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <span className="truncate">{lane.label}</span>
                <span style={{ color: 'var(--flow-border)' }}>·</span>
                <span className="tabular-nums">{lane.tasks.length}</span>
              </div>
              <div className="flex flex-col gap-1.5">
                {lane.tasks.map((task) => (
                  <TaskCard
                    key={task.id}
                    task={task}
                    draggable={canEdit}
                    onDragStart={(e) => onDragStart(e, task.id)}
                    selected={selectedTaskIds.includes(task.id)}
                    onSelect={() => onSelect(task.id)}
                  />
                ))}
                {lane.tasks.length === 0 && (
                  <div
                    className="rounded text-[10px] min-h-[28px] flex items-center justify-center"
                    style={{
                      color: 'var(--flow-text-dim)',
                      border: '1px dashed var(--flow-border)',
                    }}
                  >
                    &mdash;
                  </div>
                )}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

// ── Group-by picker ─────────────────────────────────────────────────────────

const GroupByPicker: React.FC<{
  groupBy: GroupKey;
  onChange: (k: GroupKey) => void;
}> = ({ groupBy, onChange }) => {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  const close = useCallback((): void => setOpen(false), []);
  useClickOutside(ref, close);

  const current = GROUP_OPTIONS.find((o) => o.key === groupBy)!;

  return (
    <div ref={ref} className="relative">
      <button type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-1.5 h-7 px-2 rounded-md text-[11px] font-medium transition-colors hover:bg-[var(--flow-hover)]"
        style={{
          background: 'var(--flow-surface)',
          border: '1px solid var(--flow-border)',
          color: 'var(--flow-text-muted)',
        }}
      >
        <span style={{ color: 'var(--flow-text-dim)' }}>Group:</span>
        <span style={{ color: 'var(--flow-text)' }}>{current.label}</span>
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
          {GROUP_OPTIONS.map((o) => (
            <button type="button"
              key={o.key}
              onClick={() => {
                onChange(o.key);
                setOpen(false);
              }}
              className="w-full text-start px-2.5 py-1.5 text-[11px] transition-colors hover:bg-[var(--flow-hover)]"
              style={{
                color: o.key === groupBy ? 'var(--flow-text)' : 'var(--flow-text-muted)',
                fontWeight: o.key === groupBy ? 600 : 400,
              }}
            >
              {o.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default ProjectBoard;
