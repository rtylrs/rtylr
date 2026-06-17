import React, { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { useNavigate } from '@tanstack/react-router';
import { motion } from 'motion/react';
import {
  Calendar,
  MessageSquare,
  Paperclip,
  GitBranch,
  ExternalLink,
  Trash2,
  Copy,
} from 'lucide-react';
import { toast } from 'sonner';
import { Task, useFlowStore } from '../stores/flowStore';
import { formatDate, cn } from '../lib/utils';
import PriorityIcon from './PriorityIcon';
import AssigneeAvatar from './AssigneeAvatar';
import LabelChip from './LabelChip';
import { useDeleteTask } from '../../features/tasks/useTasks';

interface TaskCardProps {
  task: Task;
  compact?: boolean;
  showProject?: boolean;
  draggable?: boolean;
  onDragStart?: (e: React.DragEvent) => void;
  selected?: boolean;
  onSelect?: () => void;
}

interface ContextMenuState {
  x: number;
  y: number;
  taskId: string;
}

const TaskContextMenu: React.FC<{
  menu: ContextMenuState;
  onClose: () => void;
}> = ({ menu, onClose }) => {
  const navigate = useNavigate();
  const deleteTask = useDeleteTask();
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const close = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose();
    };
    const closeKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('mousedown', close);
    document.addEventListener('keydown', closeKey);
    return () => {
      document.removeEventListener('mousedown', close);
      document.removeEventListener('keydown', closeKey);
    };
  }, [onClose]);

  const menuStyle: React.CSSProperties = {
    position: 'fixed',
    left: menu.x,
    top: menu.y,
    zIndex: 9999,
    minWidth: 160,
    background: 'var(--flow-card)',
    border: '1px solid var(--flow-border)',
    borderRadius: 8,
    boxShadow: '0 8px 24px rgba(0,0,0,0.35)',
    padding: '4px 0',
  };

  const itemStyle: React.CSSProperties = {
    display: 'flex',
    alignItems: 'center',
    gap: 8,
    width: '100%',
    padding: '6px 12px',
    fontSize: 12,
    cursor: 'pointer',
    background: 'transparent',
    border: 'none',
    color: 'var(--flow-text)',
    textAlign: 'left',
  };

  return createPortal(
    <div ref={ref} style={menuStyle} role="menu">
      <button
        style={itemStyle}
        onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--flow-hover)')}
        onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
        onClick={() => {
          navigate({ to: `/task/${menu.taskId}` as string });
          onClose();
        }}
      >
        <ExternalLink size={12} style={{ color: 'var(--flow-text-muted)', flexShrink: 0 }} />
        Open task
      </button>
      <button
        style={itemStyle}
        onMouseEnter={(e) => (e.currentTarget.style.background = 'var(--flow-hover)')}
        onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
        onClick={() => {
          void navigator.clipboard.writeText(`${window.location.origin}/task/${menu.taskId}`);
          toast.success('Link copied');
          onClose();
        }}
      >
        <Copy size={12} style={{ color: 'var(--flow-text-muted)', flexShrink: 0 }} />
        Copy link
      </button>
      <div style={{ height: 1, background: 'var(--flow-border)', margin: '4px 0' }} />
      <button
        style={{ ...itemStyle, color: '#ef4444' }}
        onMouseEnter={(e) => (e.currentTarget.style.background = 'rgba(239,68,68,0.08)')}
        onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
        onClick={() => {
          deleteTask.mutate(menu.taskId, {
            onSuccess: () => toast.success('Task deleted'),
            onError: () => toast.error('Failed to delete task'),
          });
          onClose();
        }}
      >
        <Trash2 size={12} style={{ color: '#ef4444', flexShrink: 0 }} />
        Delete
      </button>
    </div>,
    document.body
  );
};

export const TaskCard: React.FC<TaskCardProps> = ({
  task,
  compact,
  showProject,
  draggable,
  onDragStart,
  selected,
  onSelect,
}) => {
  const navigate = useNavigate();
  const { labels, comments, attachments, subtasks, projects } = useFlowStore();
  const taskLabels = labels.filter((l) => task.labelIds?.includes(l.id));
  const project = projects.find((p) => p.id === task.projectId);
  const commentCount = comments.filter((c) => c.taskId === task.id).length;
  const attachmentCount = attachments.filter((a) => a.taskId === task.id).length;
  const taskSubtasks = subtasks.filter((s) => s.taskId === task.id);
  const doneSubtasks = taskSubtasks.filter((s) => s.completed).length;
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);

  const overdue = task.dueDate && new Date(task.dueDate) < new Date() && task.status !== 'done';

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const vw = window.innerWidth;
    const vh = window.innerHeight;
    const menuW = 160;
    const menuH = 120;
    const x = e.clientX + menuW > vw ? e.clientX - menuW : e.clientX;
    const y = e.clientY + menuH > vh ? e.clientY - menuH : e.clientY;
    setContextMenu({ x, y, taskId: task.id });
  };

  return (
    <>
      <motion.div
        layout
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -8 }}
        transition={{ duration: 0.18 }}
        draggable={draggable}
        onDragStart={
          onDragStart as unknown as React.ComponentProps<typeof motion.div>['onDragStart']
        }
        onClick={(e) => {
          if (e.metaKey || e.ctrlKey) {
            onSelect?.();
          } else {
            navigate({ to: `/task/${task.id}` as string });
          }
        }}
        onContextMenu={handleContextMenu}
        className={cn('rounded-lg cursor-pointer group relative', compact ? 'p-2.5' : 'p-3.5')}
        style={{
          background: 'var(--flow-card)',
          border: selected ? '1px solid #64C5B9' : '1px solid var(--flow-border)',
          boxShadow: selected ? '0 0 0 3px var(--accent-muted)' : '0 1px 3px rgba(0,0,0,0.15)',
        }}
        whileHover={{ y: -1 }}
      >
        {/* Header row */}
        <div className="flex items-start justify-between gap-2 mb-2">
          <div className="flex min-w-0 items-center gap-2">
            <PriorityIcon priority={task.priority} />
            <span
              className="text-[10px] font-mono shrink-0"
              style={{ color: 'var(--flow-text-dim)' }}
            >
              {task.identifier}
            </span>
          </div>
          {showProject && (project || task.projectName) && (
            <span
              className="min-w-0 max-w-[8rem] truncate rounded px-1.5 py-0.5 text-[10px] font-semibold"
              style={{
                background: `${project?.color ?? '#64C5B9'}22`,
                color: project?.color ?? 'var(--accent)',
              }}
              title={project?.name ?? task.projectName}
            >
              {project?.name ?? task.projectName}
            </span>
          )}
        </div>

        {/* Title */}
        <p
          className={cn('font-medium leading-snug mb-2', compact ? 'text-[12px]' : 'text-[13px]')}
          style={{ color: 'var(--flow-text)' }}
        >
          {task.title}
        </p>

        {/* Labels */}
        {taskLabels.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-2">
            {taskLabels.slice(0, 3).map((l) => (
              <LabelChip key={l.id} label={l} small />
            ))}
          </div>
        )}

        {/* Footer */}
        <div className="flex items-center justify-between mt-1">
          <div className="flex items-center gap-2">
            <AssigneeAvatar name={task.assignee} initials={task.assigneeInitials} size={18} />
            {taskSubtasks.length > 0 && (
              <span
                className="flex items-center gap-0.5 text-[10px]"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <GitBranch size={10} />
                {doneSubtasks}/{taskSubtasks.length}
              </span>
            )}
            {commentCount > 0 && (
              <span
                className="flex items-center gap-0.5 text-[10px]"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <MessageSquare size={10} />
                {commentCount}
              </span>
            )}
            {attachmentCount > 0 && (
              <span
                className="flex items-center gap-0.5 text-[10px]"
                style={{ color: 'var(--flow-text-dim)' }}
              >
                <Paperclip size={10} />
                {attachmentCount}
              </span>
            )}
          </div>

          {task.dueDate && (
            <span
              className="flex items-center gap-1 text-[10px]"
              style={{ color: overdue ? '#ef4444' : 'var(--flow-text-dim)' }}
            >
              <Calendar size={10} />
              {formatDate(task.dueDate)}
            </span>
          )}
        </div>
      </motion.div>

      {contextMenu && <TaskContextMenu menu={contextMenu} onClose={() => setContextMenu(null)} />}
    </>
  );
};

export default TaskCard;
