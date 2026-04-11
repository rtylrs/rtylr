import React, { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  Plus,
  List,
  LayoutGrid,
  Activity,
  Settings as SettingsIcon,
} from 'lucide-react';
import { useFlowStore, Task } from '../../stores/flowStore';
import { PERSONAL_WORKSPACE_ID } from '../../lib/personalWorkspace';
import TaskCard from '../../components/TaskCard';

const ACCENT = '#64C5B9';

const COLUMNS: { id: Task['status']; label: string; color: string; bgColor: string }[] = [
  { id: 'backlog', label: 'Backlog', color: '#6b7280', bgColor: 'rgba(107,114,128,0.08)' },
  { id: 'todo', label: 'Todo', color: '#9ca3af', bgColor: 'rgba(156,163,175,0.08)' },
  { id: 'in_progress', label: 'In Progress', color: '#3b82f6', bgColor: 'rgba(59,130,246,0.08)' },
  { id: 'in_review', label: 'In Review', color: '#f59e0b', bgColor: 'rgba(245,158,11,0.08)' },
  { id: 'done', label: 'Done', color: '#22c55e', bgColor: 'rgba(34,197,94,0.08)' },
];

const ProjectBoard: React.FC = () => {
  const { workspaceId = PERSONAL_WORKSPACE_ID, projectId = 'p-web' } = useParams<{
    workspaceId: string;
    projectId: string;
  }>();

  const { projects, tasks, updateTask, setQuickCreateOpen, selectedTaskIds, toggleSelectTask } =
    useFlowStore();
  const [dragOver, setDragOver] = useState<Task['status'] | null>(null);

  const project = projects.find((p) => p.id === projectId);
  const projectTasks = tasks.filter((t) => t.projectId === projectId && !t.archived);

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

  const handleDrop = (e: React.DragEvent, status: Task['status']) => {
    e.preventDefault();
    const id = e.dataTransfer.getData('text/plain');
    if (!id) return;
    const t = projectTasks.find((x) => x.id === id);
    if (t && t.status !== status) {
      updateTask(id, { status });
    }
    setDragOver(null);
  };

  return (
    <div className="flex flex-col h-full">
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
            style={{
              background: 'var(--flow-hover)',
              color: 'var(--flow-text-dim)',
            }}
          >
            {projectTasks.length}
          </span>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <button
            onClick={() => setQuickCreateOpen(true)}
            className="flex items-center gap-1.5 h-7 px-2.5 rounded text-[11px] font-semibold"
            style={{ background: ACCENT, color: '#0a0a0a' }}
          >
            <Plus size={11} /> New task
          </button>
          <div
            className="flex items-center rounded-lg p-0.5"
            style={{
              background: 'var(--flow-surface)',
              border: '1px solid var(--flow-border)',
            }}
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

      <div className="min-h-0 flex-1 overflow-x-auto overflow-y-hidden">
        <div className="flex h-full min-w-max gap-3 p-4 sm:gap-4 sm:p-6">
          {COLUMNS.map((col) => {
            const colTasks = projectTasks.filter((t) => t.status === col.id);
            return (
              <div
                key={col.id}
                onDragOver={(e) => {
                  e.preventDefault();
                  setDragOver(col.id);
                }}
                onDragLeave={() => setDragOver(null)}
                onDrop={(e) => handleDrop(e, col.id)}
                className="flex w-[min(260px,calc(100vw-2rem))] min-w-[min(260px,calc(100vw-2rem))] flex-col rounded-lg p-2 sm:w-[260px] sm:min-w-[260px]"
                style={{
                  background:
                    dragOver === col.id ? col.bgColor : 'transparent',
                  border:
                    dragOver === col.id
                      ? `1px dashed ${col.color}`
                      : '1px solid transparent',
                }}
              >
                <div className="flex items-center gap-2 px-2 py-2 mb-2">
                  <span
                    className="h-2 w-2 rounded-full"
                    style={{ background: col.color }}
                  />
                  <span
                    className="text-[12px] font-semibold flex-1"
                    style={{ color: col.color }}
                  >
                    {col.label}
                  </span>
                  <span
                    className="text-[10px] px-1.5 py-0.5 rounded-full font-medium"
                    style={{
                      background: `${col.color}22`,
                      color: col.color,
                    }}
                  >
                    {colTasks.length}
                  </span>
                  <button
                    onClick={() => setQuickCreateOpen(true)}
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
                      onDragStart={(e) => handleDragStart(e, task.id)}
                      selected={selectedTaskIds.includes(task.id)}
                      onSelect={() => toggleSelectTask(task.id)}
                    />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default ProjectBoard;
