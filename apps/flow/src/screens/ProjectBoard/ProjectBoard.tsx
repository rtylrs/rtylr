import React, { useState, useEffect } from 'react';
import { Link, useParams } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  Plus,
  Calendar,
  MessageSquare,
  Paperclip,
  MoreHorizontal,
  List,
  LayoutGrid,
} from 'lucide-react';
import { useFlowStore, Task } from '../../stores/flowStore';
import { cn, formatDate } from '../../lib/utils';
import { flowApi } from '../../services/flowApi';

const ACCENT = '#64C5B9';

const PRIORITY_COLORS: Record<string, string> = {
  urgent: '#ef4444',
  high: '#f97316',
  medium: '#3b82f6',
  low: '#6b7280',
};

const COLUMNS: { id: Task['status']; label: string; color: string; bgColor: string }[] = [
  { id: 'todo',        label: 'Todo',       color: '#6b7280', bgColor: 'rgba(107,114,128,0.1)' },
  { id: 'in_progress', label: 'In Progress',color: '#3b82f6', bgColor: 'rgba(59,130,246,0.1)'  },
  { id: 'in_review',   label: 'In Review',  color: '#f59e0b', bgColor: 'rgba(245,158,11,0.1)'  },
  { id: 'done',        label: 'Done',       color: '#22c55e', bgColor: 'rgba(34,197,94,0.1)'   },
];

const TaskCard: React.FC<{ task: Task }> = ({ task }) => {
  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -8 }}
      transition={{ duration: 0.18 }}
      className="rounded-lg p-3.5 cursor-pointer group"
      style={{
        background: 'var(--flow-card)',
        border: '1px solid var(--flow-border)',
        boxShadow: '0 1px 3px rgba(0,0,0,0.2)',
      }}
      whileHover={{ y: -1, boxShadow: '0 4px 12px rgba(0,0,0,0.3)' }}
    >
      {/* Title */}
      <div className="flex items-start gap-2 mb-3">
        <span
          className="mt-1 h-2 w-2 rounded-full shrink-0"
          style={{ background: PRIORITY_COLORS[task.priority] }}
        />
        <p
          className="text-[13px] font-medium leading-snug flex-1"
          style={{ color: 'var(--flow-text)' }}
        >
          {task.title}
        </p>
        <button
          className="opacity-0 group-hover:opacity-100 transition-opacity h-5 w-5 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
          style={{ color: 'var(--flow-text-dim)' }}
        >
          <MoreHorizontal size={13} />
        </button>
      </div>

      {/* Footer */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {/* Assignee avatar */}
          <span
            className="h-5 w-5 rounded-full flex items-center justify-center text-[9px] font-bold shrink-0"
            style={{ background: `${ACCENT}22`, color: ACCENT }}
          >
            {task.assigneeInitials}
          </span>

          {task.subtaskCount > 0 && (
            <span
              className="flex items-center gap-1 text-[11px]"
              style={{ color: 'var(--flow-text-dim)' }}
            >
              <MessageSquare size={11} />
              {task.subtaskCount}
            </span>
          )}
        </div>

        <span
          className="flex items-center gap-1 text-[11px]"
          style={{ color: 'var(--flow-text-dim)' }}
        >
          <Calendar size={11} />
          {formatDate(task.dueDate)}
        </span>
      </div>
    </motion.div>
  );
};

const KanbanColumn: React.FC<{
  column: typeof COLUMNS[number];
  tasks: Task[];
}> = ({ column, tasks }) => {
  return (
    <div className="flex flex-col min-w-[260px] w-[260px]">
      {/* Column header */}
      <div
        className="flex items-center gap-2 px-3 py-2.5 rounded-lg mb-3"
        style={{ background: column.bgColor }}
      >
        <span
          className="h-2 w-2 rounded-full"
          style={{ background: column.color }}
        />
        <span
          className="text-[12px] font-semibold flex-1"
          style={{ color: column.color }}
        >
          {column.label}
        </span>
        <span
          className="text-[11px] px-1.5 py-0.5 rounded-full font-medium"
          style={{
            background: `${column.color}22`,
            color: column.color,
          }}
        >
          {tasks.length}
        </span>
      </div>

      {/* Tasks */}
      <div className="flex flex-col gap-2 flex-1 min-h-[200px]">
        {tasks.map((task) => (
          <TaskCard key={task.id} task={task} />
        ))}
      </div>

      {/* Add task */}
      <button
        className="mt-3 flex items-center gap-2 px-3 py-2 rounded-lg text-[12px] w-full transition-colors hover:bg-[var(--flow-hover)]"
        style={{ color: 'var(--flow-text-dim)' }}
      >
        <Plus size={13} />
        Add Task
      </button>
    </div>
  );
};

const ProjectBoard: React.FC = () => {
  const { workspaceId = 'default', projectId = 'flow' } = useParams<{
    workspaceId: string;
    projectId: string;
  }>();

  const { projects, tasks } = useFlowStore();
  const [apiTasks, setApiTasks] = useState<Task[] | null>(null);

  useEffect(() => {
    flowApi.getTasks({ projectId }).then(setApiTasks).catch(() => {});
  }, [projectId]);

  const project = projects.find((p) => p.id === projectId);
  const projectTasks = apiTasks ?? tasks.filter((t) => t.projectId === projectId);

  if (!project) {
    return (
      <div className="p-6" style={{ color: 'var(--flow-text-muted)' }}>
        Project not found.
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      {/* Board Header */}
      <div
        className="flex items-center justify-between px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div className="flex items-center gap-3">
          <span
            className="h-3 w-3 rounded-full"
            style={{ background: project.color }}
          />
          <h1
            className="text-[15px] font-semibold"
            style={{ color: 'var(--flow-text)' }}
          >
            {project.name}
          </h1>
        </div>

        {/* View Switcher */}
        <div
          className="flex items-center rounded-lg p-0.5"
          style={{ background: 'var(--flow-surface)', border: '1px solid var(--flow-border)' }}
        >
          <button
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-[12px] font-medium transition-colors"
            style={{
              background: `${ACCENT}18`,
              color: ACCENT,
            }}
          >
            <LayoutGrid size={13} />
            Board
          </button>
          <Link
            to={`/w/${workspaceId}/p/${projectId}/list`}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-[12px] font-medium transition-colors hover:bg-[var(--flow-hover)] no-underline"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <List size={13} />
            List
          </Link>
        </div>
      </div>

      {/* Kanban Board */}
      <div className="flex-1 overflow-x-auto overflow-y-hidden">
        <div className="flex gap-5 p-6 h-full min-w-max">
          {COLUMNS.map((col) => (
            <KanbanColumn
              key={col.id}
              column={col}
              tasks={projectTasks.filter((t) => t.status === col.id)}
            />
          ))}
        </div>
      </div>
    </div>
  );
};

export default ProjectBoard;
