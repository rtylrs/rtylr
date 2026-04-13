import React, { useState, useMemo } from 'react';
import { Link, useParams } from 'react-router-dom';
import { Plus, List, LayoutGrid, Activity, Settings as SettingsIcon, Rows3 } from 'lucide-react';
import { useFlowStore, Task } from '../../shared/stores/flowStore';
import { useUpdateTaskStatus, useCreateTask } from '../tasks/useTasks';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';
import TaskCard from '../../shared/components/TaskCard';

const ACCENT = '#64C5B9';

const COLUMNS: { id: Task['status']; label: string; color: string; bgColor: string }[] = [
  { id: 'backlog', label: 'Backlog', color: '#6b7280', bgColor: 'rgba(107,114,128,0.08)' },
  { id: 'todo', label: 'Todo', color: '#9ca3af', bgColor: 'rgba(156,163,175,0.08)' },
  { id: 'in_progress', label: 'In Progress', color: '#3b82f6', bgColor: 'rgba(59,130,246,0.08)' },
  { id: 'in_review', label: 'In Review', color: '#f59e0b', bgColor: 'rgba(245,158,11,0.08)' },
  { id: 'done', label: 'Done', color: '#22c55e', bgColor: 'rgba(34,197,94,0.08)' },
];

const PRIORITY_COLORS: Record<string, string> = {
  urgent: '#ef4444',
  high: '#f97316',
  medium: '#eab308',
  low: '#22c55e',
  none: '#6b7280',
};

type GroupBy = 'none' | 'assignee' | 'priority';

function groupTasks(
  tasks: Task[],
  groupBy: GroupBy
): { key: string; label: string; tasks: Task[] }[] {
  if (groupBy === 'none') return [{ key: '__all__', label: '', tasks }];

  if (groupBy === 'priority') {
    const order = ['urgent', 'high', 'medium', 'low', 'none'];
    const map: Record<string, Task[]> = {};
    for (const t of tasks) {
      const k = t.priority || 'none';
      if (!map[k]) map[k] = [];
      map[k].push(t);
    }
    return order
      .filter((k) => map[k]?.length)
      .map((k) => ({
        key: k,
        label: k.charAt(0).toUpperCase() + k.slice(1),
        tasks: map[k],
      }));
  }

  if (groupBy === 'assignee') {
    const map: Record<string, { key: string; label: string; tasks: Task[] }> = {};
    for (const t of tasks) {
      const k = t.assignee || 'Unassigned';
      if (!map[k]) map[k] = { key: k, label: k, tasks: [] };
      map[k].tasks.push(t);
    }
    return Object.values(map);
  }

  return [{ key: '__all__', label: '', tasks }];
}

const ProjectBoard: React.FC = () => {
  const { workspaceId = PERSONAL_WORKSPACE_ID, projectId = 'p-web' } = useParams<{
    workspaceId: string;
    projectId: string;
  }>();

  const { projects, tasks, setQuickCreateOpen, selectedTaskIds, toggleSelectTask } = useFlowStore();
  const updateTaskStatus = useUpdateTaskStatus();
  const [dragOver, setDragOver] = useState<{ colId: Task['status']; groupKey: string } | null>(
    null
  );
  const [groupBy, setGroupBy] = useState<GroupBy>('none');
  const [showGroupMenu, setShowGroupMenu] = useState(false);

  const project = projects.find((p) => p.id === projectId);
  const projectTasks = tasks.filter((t) => t.projectId === projectId && !t.archived);
  const groups = useMemo(() => groupTasks(projectTasks, groupBy), [projectTasks, groupBy]);

  if (!project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        Project not found.
      </div>
    );
  }

  const handleDragStart = (e: React.DragEvent, id: string) => {
    e.dataTransfer.setData('text/plain', id);
    e.dataTransfer.effectAllowed = 'move';
  };

  const handleDrop = (e: React.DragEvent, status: Task['status'], _groupKey: string) => {
    e.preventDefault();
    const id = e.dataTransfer.getData('text/plain');
    if (!id) return;
    const t = projectTasks.find((x) => x.id === id);
    if (t && t.status !== status) {
      updateTaskStatus.mutate({ id, status: status as any });
    }
    setDragOver(null);
  };

  const groupLabel = groupBy === 'none' ? 'Group by' : `Grouped by ${groupBy}`;

  return (
    <div className="flex flex-col h-full">
      {/* ── Header ── */}
      <div
        className="flex flex-col gap-3 px-4 py-3 shrink-0 sm:flex-row sm:items-center sm:justify-between sm:px-6 sm:py-4"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div className="flex min-w-0 items-center gap-3">
          <span className="h-3 w-3 rounded-full" style={{ background: project.color }} />
          <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            {project.name}
          </h1>
          <span
            className="text-[11px] px-2 py-0.5 rounded-full font-semibold"
            style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-dim)' }}
          >
            {projectTasks.length}
          </span>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {/* Group by picker */}
          <div className="relative">
            <button
              onClick={() => setShowGroupMenu((v) => !v)}
              className="flex items-center gap-1.5 h-7 px-2.5 rounded text-[11px]"
              style={{
                background: groupBy !== 'none' ? `${ACCENT}18` : 'var(--flow-hover)',
                color: groupBy !== 'none' ? ACCENT : 'var(--flow-text-muted)',
                border: '1px solid var(--flow-border)',
              }}
            >
              <Rows3 size={10} />
              {groupLabel}
            </button>
            {showGroupMenu && (
              <div
                className="absolute right-0 top-8 z-20 rounded-lg py-1 shadow-xl min-w-[140px]"
                style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
              >
                {(['none', 'assignee', 'priority'] as GroupBy[]).map((g) => (
                  <button
                    key={g}
                    onClick={() => {
                      setGroupBy(g);
                      setShowGroupMenu(false);
                    }}
                    className="w-full text-left px-3 py-1.5 text-[11px] hover:bg-[var(--flow-hover)]"
                    style={{ color: groupBy === g ? ACCENT : 'var(--flow-text)' }}
                  >
                    {g === 'none' ? 'No grouping' : `By ${g}`}
                  </button>
                ))}
              </div>
            )}
          </div>

          <button
            onClick={() => setQuickCreateOpen(true)}
            className="flex items-center gap-1.5 h-7 px-2.5 rounded text-[11px] font-semibold"
            style={{ background: ACCENT, color: '#0a0a0a' }}
          >
            <Plus size={11} /> New task
          </button>

          <div
            className="flex items-center rounded-lg p-0.5"
            style={{ background: 'var(--flow-surface)', border: '1px solid var(--flow-border)' }}
          >
            <Link
              to={`/w/${workspaceId}/p/${projectId}/overview`}
              className="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[11px] hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <Activity size={11} /> Overview
            </Link>
            <button
              className="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[11px] font-medium"
              style={{ background: `${ACCENT}18`, color: ACCENT }}
            >
              <LayoutGrid size={11} /> Board
            </button>
            <Link
              to={`/w/${workspaceId}/p/${projectId}/list`}
              className="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[11px] hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <List size={11} /> List
            </Link>
            <Link
              to={`/w/${workspaceId}/p/${projectId}/settings`}
              className="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[11px] hover:bg-[var(--flow-hover)]"
              style={{ color: 'var(--flow-text-muted)' }}
            >
              <SettingsIcon size={11} />
            </Link>
          </div>
        </div>
      </div>

      {/* ── Board body ── */}
      <div className="min-h-0 flex-1 overflow-auto">
        {groupBy === 'none' ? (
          // Flat kanban (original layout)
          <div className="flex h-full min-w-max gap-3 p-4 sm:gap-4 sm:p-6">
            {COLUMNS.map((col) => {
              const colTasks = projectTasks.filter((t) => t.status === col.id);
              const isOver = dragOver?.colId === col.id && dragOver?.groupKey === '__all__';
              return (
                <KanbanColumn
                  key={col.id}
                  col={col}
                  colTasks={colTasks}
                  isOver={isOver}
                  onDragOver={() => setDragOver({ colId: col.id, groupKey: '__all__' })}
                  onDragLeave={() => setDragOver(null)}
                  onDrop={(e) => handleDrop(e, col.id, '__all__')}
                  onAdd={() => setQuickCreateOpen(true)}
                  selectedTaskIds={selectedTaskIds}
                  onSelect={toggleSelectTask}
                  onDragStart={handleDragStart}
                />
              );
            })}
          </div>
        ) : (
          // Swimlane layout
          <div className="min-w-max p-4 sm:p-6 space-y-6">
            {/* Column headers (sticky-ish) */}
            <div className="flex gap-3 sm:gap-4">
              <div className="w-[140px] shrink-0" /> {/* swimlane label gutter */}
              {COLUMNS.map((col) => (
                <div key={col.id} className="w-[220px] shrink-0 flex items-center gap-2 px-2">
                  <span className="h-2 w-2 rounded-full" style={{ background: col.color }} />
                  <span className="text-[11px] font-semibold" style={{ color: col.color }}>
                    {col.label}
                  </span>
                </div>
              ))}
            </div>

            {groups.map((group) => (
              <div key={group.key}>
                {/* Swimlane label row */}
                <div
                  className="flex items-center gap-2 px-3 py-2 rounded-lg mb-2"
                  style={{ background: 'var(--flow-hover)' }}
                >
                  {groupBy === 'priority' && (
                    <span
                      className="h-2 w-2 rounded-full"
                      style={{ background: PRIORITY_COLORS[group.key] ?? '#6b7280' }}
                    />
                  )}
                  <span className="text-[12px] font-semibold" style={{ color: 'var(--flow-text)' }}>
                    {group.label}
                  </span>
                  <span
                    className="text-[10px] px-1.5 py-0.5 rounded-full"
                    style={{ background: 'rgba(0,0,0,0.08)', color: 'var(--flow-text-dim)' }}
                  >
                    {group.tasks.length}
                  </span>
                </div>

                <div className="flex gap-3 sm:gap-4">
                  {COLUMNS.map((col) => {
                    const colTasks = group.tasks.filter((t) => t.status === col.id);
                    const isOver = dragOver?.colId === col.id && dragOver?.groupKey === group.key;
                    return (
                      <div
                        key={col.id}
                        onDragOver={(e) => {
                          e.preventDefault();
                          setDragOver({ colId: col.id, groupKey: group.key });
                        }}
                        onDragLeave={() => setDragOver(null)}
                        onDrop={(e) => handleDrop(e, col.id, group.key)}
                        className="flex w-[220px] shrink-0 flex-col rounded-lg p-2 min-h-[120px]"
                        style={{
                          background: isOver ? col.bgColor : 'var(--flow-card)',
                          border: isOver
                            ? `1px dashed ${col.color}`
                            : '1px solid var(--flow-border)',
                        }}
                      >
                        <div className="flex flex-col gap-2">
                          {colTasks.map((task) => (
                            <TaskCard
                              key={task.id}
                              task={task}
                              draggable
                              onDragStart={(e) => handleDragStart(e, task.id)}
                              selected={selectedTaskIds.includes(task.id)}
                              onSelect={() => toggleSelectTask(task.id)}
                            />
                          ))}
                          {colTasks.length === 0 && (
                            <div
                              className="text-center py-4 text-[10px]"
                              style={{ color: 'var(--flow-text-dim)' }}
                            >
                              Drop here
                            </div>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

// Extracted KanbanColumn for flat mode
const KanbanColumn: React.FC<{
  col: { id: Task['status']; label: string; color: string; bgColor: string };
  colTasks: Task[];
  isOver: boolean;
  onDragOver: () => void;
  onDragLeave: () => void;
  onDrop: (e: React.DragEvent) => void;
  onAdd: () => void;
  selectedTaskIds: string[];
  onSelect: (id: string) => void;
  onDragStart: (e: React.DragEvent, id: string) => void;
}> = ({
  col,
  colTasks,
  isOver,
  onDragOver,
  onDragLeave,
  onDrop,
  onAdd,
  selectedTaskIds,
  onSelect,
  onDragStart,
}) => (
  <div
    onDragOver={(e) => {
      e.preventDefault();
      onDragOver();
    }}
    onDragLeave={onDragLeave}
    onDrop={onDrop}
    className="flex w-[min(260px,calc(100vw-2rem))] min-w-[min(260px,calc(100vw-2rem))] flex-col rounded-lg p-2 sm:w-[260px] sm:min-w-[260px]"
    style={{
      background: isOver ? col.bgColor : 'transparent',
      border: isOver ? `1px dashed ${col.color}` : '1px solid transparent',
    }}
  >
    <div className="flex items-center gap-2 px-2 py-2 mb-2">
      <span className="h-2 w-2 rounded-full" style={{ background: col.color }} />
      <span className="text-[12px] font-semibold flex-1" style={{ color: col.color }}>
        {col.label}
      </span>
      <span
        className="text-[10px] px-1.5 py-0.5 rounded-full font-medium"
        style={{ background: `${col.color}22`, color: col.color }}
      >
        {colTasks.length}
      </span>
      <button
        onClick={onAdd}
        className="h-4 w-4 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
        style={{ color: 'var(--flow-text-dim)' }}
      >
        <Plus size={11} />
      </button>
    </div>
    <div className="flex flex-col gap-2 flex-1 min-h-[200px] overflow-y-auto">
      {colTasks.map((task) => (
        <TaskCard
          key={task.id}
          task={task}
          draggable
          onDragStart={(e) => onDragStart(e, task.id)}
          selected={selectedTaskIds.includes(task.id)}
          onSelect={() => onSelect(task.id)}
        />
      ))}
    </div>
  </div>
);

export default ProjectBoard;
