import React, { useState, useRef, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Plus, Trash2 } from 'lucide-react';
import { useFlowStore } from '../../shared/stores/flowStore';
import { useUpdateTaskStatus, useTasks } from '../tasks/useTasks';
import {
  useProjectColumns,
  useCreateColumn,
  useUpdateColumn,
  useDeleteColumn,
  useMoveTaskToColumn,
} from './useColumns';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';
import { apiTaskToTask } from '../../shared/lib/taskMapper';
import TaskCard from '../../shared/components/TaskCard';
import ProjectViewTabs from './ProjectViewTabs';
import type { ApiProjectColumn } from '../../shared/services/flowApi';
import type { Task } from '../../shared/stores/flowStore';

const ACCENT = '#64C5B9';

const ProjectBoard: React.FC = () => {
  const { workspaceId = PERSONAL_WORKSPACE_ID, projectId = '' } = useParams<{
    workspaceId: string;
    projectId: string;
  }>();

  const { projects, selectedTaskIds, toggleSelectTask, activeWorkspaceId } = useFlowStore();
  const moveTask = useMoveTaskToColumn();

  const { data: apiTasks = [] } = useTasks({ projectId });
  const { data: columns = [], isLoading: columnsLoading } = useProjectColumns(projectId);
  const createColumn = useCreateColumn(projectId);
  const updateColumn = useUpdateColumn(projectId);
  const deleteColumn = useDeleteColumn(projectId);

  const [dragOver, setDragOver] = useState<string | null>(null); // column id
  const [addingColumn, setAddingColumn] = useState(false);
  const [newColumnName, setNewColumnName] = useState('');

  const project = projects.find((p) => p.id === projectId);

  if (!project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        Project not found.
      </div>
    );
  }

  // Tasks mapped to Task shape for TaskCard
  const tasks: Task[] = apiTasks.map(apiTaskToTask);

  // Bucket tasks by column. Tasks with no sectionId go into first column.
  function tasksForColumn(col: ApiProjectColumn, colIndex: number): Task[] {
    return tasks.filter((t) => {
      if (t.sectionId === col.id) return true;
      // Unsectioned tasks go into first column
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
    const taskId = e.dataTransfer.getData('text/plain');
    if (!taskId) return;
    moveTask.mutate({ taskId, columnId: colId });
    setDragOver(null);
  };

  const handleAddColumn = () => {
    const name = newColumnName.trim();
    if (!name) return;
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
      <ProjectViewTabs
        project={project}
        workspaceId={workspaceId}
        projectId={projectId}
        taskCount={tasks.length}
        onNewTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
      />

      {/* Tightened kanban canvas — the previous p-6 / gap-4 combo wasted
          too much room on laptop screens. These values stay reasonable
          down to narrow viewports because the column snaps to 248px. */}
      <div className="min-h-0 flex-1 overflow-auto">
        <div className="flex h-full min-w-max gap-2 px-3 py-3">
          {columns.map((col, i) => (
            <KanbanColumn
              key={col.id}
              col={col}
              colTasks={tasksForColumn(col, i)}
              isOver={dragOver === col.id}
              selectedTaskIds={selectedTaskIds}
              onDragOver={() => setDragOver(col.id)}
              onDragLeave={() => setDragOver(null)}
              onDrop={(e) => handleDrop(e, col.id)}
              onDragStart={handleDragStart}
              onSelect={toggleSelectTask}
              onAddTask={() => useFlowStore.getState().setQuickCreateOpen(true)}
              onRename={(name) =>
                updateColumn.mutate({
                  id: col.id,
                  name,
                  color: col.color,
                  sortOrder: col.sortOrder,
                })
              }
              onDelete={() => deleteColumn.mutate(col.id)}
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
                  placeholder="Column name"
                  className="w-full bg-transparent outline-none text-[12px] font-medium mb-1.5"
                  style={{ color: 'var(--flow-text)' }}
                />
                <div className="flex items-center gap-1.5">
                  <button
                    onClick={handleAddColumn}
                    disabled={!newColumnName.trim() || createColumn.isPending}
                    className="h-6 px-2.5 rounded text-[11px] font-semibold disabled:opacity-40"
                    style={{ background: ACCENT, color: '#ffffff' }}
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
    </div>
  );
};

// ── Column component ────────────────────────────────────────────────────────

const KanbanColumn: React.FC<{
  col: ApiProjectColumn;
  colTasks: Task[];
  isOver: boolean;
  selectedTaskIds: string[];
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
  colTasks,
  isOver,
  selectedTaskIds,
  onDragOver,
  onDragLeave,
  onDrop,
  onDragStart,
  onSelect,
  onAddTask,
  onRename,
  onDelete,
}) => {
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
      {/* Column header — tighter. 24px tall, no wasteful vertical padding. */}
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
            className="flex-1 text-left text-[12px] font-semibold truncate hover:opacity-70 transition-opacity"
            style={{ color: 'var(--flow-text)' }}
          >
            {col.name}
          </button>
        )}

        <span
          className="text-[10px] px-1.5 py-0.5 rounded-full font-medium shrink-0"
          style={{ background: `${col.color}22`, color: col.color }}
        >
          {colTasks.length}
        </span>

        {/* Actions (visible on header hover) */}
        <div className="flex items-center gap-0.5 opacity-0 group-hover/header:opacity-100 transition-opacity shrink-0">
          <button
            onClick={onAddTask}
            className="h-5 w-5 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-dim)' }}
            title="Add task"
          >
            <Plus size={11} />
          </button>
          <button
            onClick={onDelete}
            className="h-5 w-5 rounded flex items-center justify-center hover:bg-[var(--flow-hover)] hover:text-red-400"
            style={{ color: 'var(--flow-text-dim)' }}
            title="Delete column"
          >
            <Trash2 size={11} />
          </button>
        </div>
      </div>

      {/* Cards */}
      <div className="flex flex-col gap-1.5 flex-1 min-h-[80px] overflow-y-auto">
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

        {colTasks.length === 0 && (
          <button
            onClick={onAddTask}
            className="flex-1 flex items-center justify-center rounded text-[11px] min-h-[44px] hover:bg-[var(--flow-hover)] transition-colors"
            style={{
              color: 'var(--flow-text-dim)',
              border: '1px dashed var(--flow-border)',
            }}
          >
            <Plus size={12} className="mr-1" /> Add task
          </button>
        )}
      </div>
    </div>
  );
};

export default ProjectBoard;
