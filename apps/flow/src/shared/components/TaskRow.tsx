import React, { useState } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { Task, useFlowStore } from '../stores/flowStore';
import { useLabels } from '../../features/tasks/useLabels';
import { labelsForTask } from '../lib/taskLabelUtils';
import { formatDate, cn } from '../lib/utils';
import PriorityIcon from './PriorityIcon';
import StatusIcon from './StatusIcon';
import AssigneeAvatar from './AssigneeAvatar';
import { Check } from 'lucide-react';
import { TaskContextMenu, ContextMenuState } from './TaskContextMenu';
import { withAlpha } from '../lib/colorUtils';

interface TaskRowProps {
  task: Task;
  showProject?: boolean;
  selected?: boolean;
  onSelect?: (e: React.MouseEvent | React.KeyboardEvent) => void;
}

export const TaskRow: React.FC<TaskRowProps> = ({ task, showProject, selected, onSelect }) => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { projects, toggleSelectTask } = useFlowStore();
  // See TaskCard: the store's `labels` is never populated, so this read from
  // the query is what makes the chips appear.
  const { data: allLabels = [] } = useLabels();
  const taskLabels = labelsForTask(allLabels, task);
  const project = projects.find((p) => p.id === task.projectId);
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);

  const overdue = task.dueDate && new Date(task.dueDate) < new Date() && task.status !== 'done';

  // Shared by mouse click and Enter/Space keydown so the row opens the task
  // (or toggles selection with a modifier held) the same way for both input
  // methods - see the "Select task" checkbox below for the matching guard
  // that keeps its own activation from bubbling into this handler.
  const activateRow = (e: React.MouseEvent | React.KeyboardEvent) => {
    if (e.metaKey || e.ctrlKey || e.shiftKey) {
      toggleSelectTask(task.id);
      onSelect?.(e);
    } else {
      navigate({ to: `/task/${task.id}` as string });
    }
  };

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
    setContextMenu({ x, y, taskId: task.id, projectId: task.projectId, taskTitle: task.title });
  };

  return (
    <>
      <div
        onClick={activateRow}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            activateRow(e);
          }
        }}
        onContextMenu={handleContextMenu}
        role="button"
        tabIndex={0}
        aria-label={t('common.taskRow.openTask', {
          identifier: task.identifier,
          title: task.title,
        })}
        className={cn(
          'flow-focus group flex items-center gap-3 px-4 py-2 cursor-pointer transition-colors',
          'hover:bg-[var(--flow-hover)]'
        )}
        style={{
          borderBottom: '1px solid var(--flow-border)',
          background: selected ? 'var(--accent-muted)' : undefined,
        }}
      >
        {/* Checkbox, always visible when selected, appears on row hover otherwise.
            Its own click/keydown stop propagation so activating it never also
            triggers the row's activateRow (open task), and vice versa. */}
        <button
          type="button"
          onClick={handleCheckbox}
          onKeyDown={(e) => e.stopPropagation()}
          className={cn(
            'flow-focus shrink-0 h-4 w-4 rounded-sm border flex items-center justify-center transition-all',
            selected ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'
          )}
          style={{
            borderColor: selected ? 'var(--accent)' : 'var(--flow-border)',
            background: selected ? 'var(--accent)' : 'transparent',
          }}
          aria-label={selected ? t('common.taskRow.deselectTask') : t('common.taskRow.selectTask')}
        >
          {selected && <Check size={9} color="#0a0a0a" strokeWidth={3} />}
        </button>
        <PriorityIcon priority={task.priority} />
        <span
          className="text-meta font-mono w-[64px] shrink-0"
          style={{ color: 'var(--flow-text-dim)' }}
        >
          {task.identifier}
        </span>
        <StatusIcon status={task.status} />
        <div className="flex-1 min-w-0">
          <p className="text-body font-medium truncate" style={{ color: 'var(--flow-text)' }}>
            {task.title}
          </p>
          {task.waitingOn && (
            <p className="text-meta mt-0.5 truncate" style={{ color: 'var(--flow-text-dim)' }}>
              Waiting on {task.waitingOn}
            </p>
          )}
        </div>
        {showProject && project && (
          <span
            className="text-meta px-2 py-0.5 rounded-sm shrink-0 font-medium"
            style={{
              background: withAlpha(project.color, 13),
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
                background: withAlpha(l.color, 10),
                color: l.color,
                border: `1px solid ${withAlpha(l.color, 20)}`,
                padding: '1px 5px',
                fontSize: 'var(--text-meta)',
              }}
            >
              <span className="h-1.5 w-1.5 rounded-full shrink-0" style={{ background: l.color }} />
              {l.name}
            </span>
          ))}
          {taskLabels.length > 3 && (
            <span
              className="text-meta px-1 rounded-full shrink-0"
              style={{ background: 'var(--flow-hover)', color: 'var(--flow-text-dim)' }}
            >
              +{taskLabels.length - 3}
            </span>
          )}
        </div>
        {task.dueDate && (
          <span
            className="text-meta shrink-0 w-[64px] text-right"
            style={{ color: overdue ? 'var(--flow-danger-fg)' : 'var(--flow-text-dim)' }}
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
