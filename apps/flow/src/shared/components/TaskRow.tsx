import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Task, useFlowStore } from '../stores/flowStore';
import { formatDate, cn } from '../lib/utils';
import PriorityIcon from './PriorityIcon';
import StatusIcon from './StatusIcon';
import AssigneeAvatar from './AssigneeAvatar';
import LabelChip from './LabelChip';

interface TaskRowProps {
  task: Task;
  showProject?: boolean;
  selected?: boolean;
  onSelect?: (e: React.MouseEvent) => void;
}

export const TaskRow: React.FC<TaskRowProps> = ({
  task,
  showProject,
  selected,
  onSelect,
}) => {
  const navigate = useNavigate();
  const { labels, projects } = useFlowStore();
  const taskLabels = labels.filter((l) => task.labels?.includes(l.id));
  const project = projects.find((p) => p.id === task.projectId);

  const overdue = task.dueDate && new Date(task.dueDate) < new Date() && task.status !== 'done';

  return (
    <div
      onClick={(e) => {
        if (e.metaKey || e.ctrlKey || e.shiftKey) {
          onSelect?.(e);
        } else {
          navigate(`/task/${task.id}`);
        }
      }}
      className={cn(
        'group flex items-center gap-3 px-4 py-2 cursor-pointer transition-colors',
        'hover:bg-[var(--flow-hover)]'
      )}
      style={{
        borderBottom: '1px solid var(--flow-border)',
        background: selected ? 'rgba(100,197,185,0.08)' : undefined,
      }}
    >
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
      <div className="flex items-center gap-1 shrink-0 max-w-[180px] overflow-hidden">
        {taskLabels.slice(0, 2).map((l) => (
          <LabelChip key={l.id} label={l} small />
        ))}
      </div>
      {task.dueDate && (
        <span
          className="text-[11px] shrink-0 w-16 text-right"
          style={{ color: overdue ? '#ef4444' : 'var(--flow-text-dim)' }}
        >
          {formatDate(task.dueDate)}
        </span>
      )}
      <AssigneeAvatar
        name={task.assignee}
        initials={task.assigneeInitials}
        size={20}
      />
    </div>
  );
};

export default TaskRow;
