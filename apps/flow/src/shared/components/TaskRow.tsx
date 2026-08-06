import React, { useEffect, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { ExternalLink, Trash2, Copy } from 'lucide-react';
import { Task, useFlowStore } from '../stores/flowStore';
import { formatDate, cn } from '../lib/utils';
import PriorityIcon from './PriorityIcon';
import StatusIcon from './StatusIcon';
import AssigneeAvatar from './AssigneeAvatar';
import { Check } from 'lucide-react';
import { useDeleteTask } from '../../features/tasks/useTasks';
import { useCanDeleteResource } from '../lib/currentUser';
import { useTranslation } from 'react-i18next';

interface TaskRowProps {
  task: Task;
  showProject?: boolean;
  selected?: boolean;
  onSelect?: (e: React.MouseEvent) => void;
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
  const { t } = useTranslation();
  const navigate = useNavigate();
  const deleteTask = useDeleteTask();
  const canRemove = useCanDeleteResource();
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
        style={{
          ...itemStyle,
          color: canRemove ? '#ef4444' : 'var(--flow-text-dim)',
          cursor: canRemove ? 'pointer' : 'not-allowed',
        }}
        disabled={!canRemove}
        title={!canRemove ? t('common.deleteRestricted') : undefined}
        onMouseEnter={(e) => {
          if (canRemove) e.currentTarget.style.background = 'rgba(239,68,68,0.08)';
        }}
        onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
        onClick={() => {
          if (!canRemove) return;
          deleteTask.mutate(menu.taskId, {
            onSuccess: () => toast.success('Task deleted'),
            onError: () => toast.error('Failed to delete task'),
          });
          onClose();
        }}
      >
        <Trash2
          size={12}
          style={{ color: canRemove ? '#ef4444' : 'var(--flow-text-dim)', flexShrink: 0 }}
        />
        Delete
      </button>
    </div>,
    document.body
  );
};

export const TaskRow: React.FC<TaskRowProps> = ({ task, showProject, selected, onSelect }) => {
  const navigate = useNavigate();
  const { labels, projects, toggleSelectTask } = useFlowStore();
  const taskLabels = labels.filter((l) => task.labelIds?.includes(l.id));
  const project = projects.find((p) => p.id === task.projectId);
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);

  const overdue = task.dueDate && new Date(task.dueDate) < new Date() && task.status !== 'done';

  const handleCheckbox = (e: React.MouseEvent) => {
    e.stopPropagation();
    toggleSelectTask(task.id);
    onSelect?.(e);
  };

  const handleContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();
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
      <div
        onClick={(e) => {
          if (e.metaKey || e.ctrlKey || e.shiftKey) {
            toggleSelectTask(task.id);
            onSelect?.(e);
          } else {
            navigate({ to: `/task/${task.id}` as string });
          }
        }}
        onContextMenu={handleContextMenu}
        className={cn(
          'group flex items-center gap-3 px-4 py-2 cursor-pointer transition-colors',
          'hover:bg-[var(--flow-hover)]'
        )}
        style={{
          borderBottom: '1px solid var(--flow-border)',
          background: selected ? 'var(--accent-muted)' : undefined,
        }}
      >
        {/* Checkbox, always visible when selected, appears on row hover otherwise */}
        <button
          onClick={handleCheckbox}
          className={cn(
            'shrink-0 h-4 w-4 rounded border flex items-center justify-center transition-all',
            selected ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'
          )}
          style={{
            borderColor: selected ? 'var(--accent)' : 'var(--flow-border)',
            background: selected ? 'var(--accent)' : 'transparent',
          }}
          aria-label={selected ? 'Deselect task' : 'Select task'}
        >
          {selected && <Check size={9} color="#0a0a0a" strokeWidth={3} />}
        </button>
        <PriorityIcon priority={task.priority} />
        <span
          className="text-[10px] font-mono w-16 shrink-0"
          style={{ color: 'var(--flow-text-dim)' }}
        >
          {task.identifier}
        </span>
        <StatusIcon status={task.status} />
        <div className="flex-1 min-w-0">
          <p className="text-[13px] font-medium truncate" style={{ color: 'var(--flow-text)' }}>
            {task.title}
          </p>
          {task.waitingOn && (
            <p className="text-[10px] mt-0.5 truncate" style={{ color: 'var(--flow-text-dim)' }}>
              Waiting on {task.waitingOn}
            </p>
          )}
        </div>
        {showProject && project && (
          <span
            className="text-[11px] px-2 py-0.5 rounded shrink-0 font-medium"
            style={{
              background: `${project.color}22`,
              color: project.color,
            }}
          >
            {project.name}
          </span>
        )}
        <div className="flex items-center gap-1 shrink-0 max-w-[200px] overflow-hidden">
          {taskLabels.slice(0, 3).map((l) => (
            <span
              key={l.id}
              className="inline-flex items-center gap-0.5 rounded-full"
              style={{
                background: `${l.color}1a`,
                color: l.color,
                border: `1px solid ${l.color}33`,
                padding: '1px 5px',
                fontSize: 10,
              }}
            >
              <span className="h-1.5 w-1.5 rounded-full shrink-0" style={{ background: l.color }} />
              {l.name}
            </span>
          ))}
          {taskLabels.length > 3 && (
            <span
              className="text-[10px] px-1 rounded-full shrink-0"
              style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-dim)' }}
            >
              +{taskLabels.length - 3}
            </span>
          )}
        </div>
        {task.dueDate && (
          <span
            className="text-[11px] shrink-0 w-16 text-right"
            style={{ color: overdue ? '#ef4444' : 'var(--flow-text-dim)' }}
          >
            {formatDate(task.dueDate)}
          </span>
        )}
        <AssigneeAvatar name={task.assignee} initials={task.assigneeInitials} size={20} />
      </div>

      {contextMenu && <TaskContextMenu menu={contextMenu} onClose={() => setContextMenu(null)} />}
    </>
  );
};

export default TaskRow;
